// Package repocheck runs offline consistency checks between the Go and golangci-lint versions
// a repository declares in go.mod/go.work, .fgm.toml, CI workflows, Dockerfiles, and lint
// configuration. It reads repository files only: no network, installs, or writes.
package repocheck

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/fgmconfig"
	"github.com/kostikovk/fgm/internal/lintdetect"
	"github.com/kostikovk/fgm/internal/report"
	"github.com/kostikovk/fgm/internal/versionutil"
	"github.com/kostikovk/fgm/internal/workspace"
	"gopkg.in/yaml.v3"
)

// Reporter builds the selection envelope for a directory.
type Reporter interface {
	Report(ctx context.Context, dir string) (app.CurrentReport, error)
}

// Config configures a Service.
type Config struct {
	Reporter Reporter
}

// Service runs the checks.
type Service struct {
	reporter Reporter
}

// New constructs a Service.
func New(config Config) *Service {
	return &Service{reporter: config.Reporter}
}

var (
	minorPattern      = regexp.MustCompile(`([0-9]+)\.([0-9]+)`)
	dockerFromPattern = regexp.MustCompile(`(?i)^\s*FROM\s+(?:--\S+\s+)*(?:\S+/)?golang:(\S+)`)
	tagVersionPattern = regexp.MustCompile(`^([0-9]+\.[0-9]+(?:\.[0-9]+)?)(?:$|-|@)`)
)

// Check reports the findings for the module containing dir.
func (s *Service) Check(ctx context.Context, dir string) (app.CheckReport, error) {
	current, err := s.reporter.Report(ctx, dir)
	if err != nil {
		return app.CheckReport{}, err
	}
	result := app.CheckReport{CurrentReport: current, Findings: []app.CheckFinding{}}

	skip, err := skippedChecks(dir)
	if err != nil {
		return app.CheckReport{}, err
	}
	module, err := loadModule(dir)
	if err != nil {
		return app.CheckReport{}, err
	}

	var findings []app.CheckFinding
	for _, check := range []func() ([]app.CheckFinding, error){
		func() ([]app.CheckFinding, error) { return checkLintCompatibility(current, module), nil },
		func() ([]app.CheckFinding, error) { return checkLintConfigSchema(current, module) },
		func() ([]app.CheckFinding, error) { return checkLintPinsAgree(module) },
		func() ([]app.CheckFinding, error) { return checkSetupGo(current, module) },
		func() ([]app.CheckFinding, error) { return checkDockerfiles(module) },
		func() ([]app.CheckFinding, error) { return checkGoWork(module) },
		func() ([]app.CheckFinding, error) { return checkToolBypass(current, module) },
	} {
		found, err := check()
		if err != nil {
			return app.CheckReport{}, err
		}
		findings = append(findings, found...)
	}

	for _, finding := range findings {
		if skip[finding.ID] || (finding.ID == "C4b" && skip["C4"]) {
			continue
		}
		finding.File = relative(dir, finding.File)
		result.Findings = append(result.Findings, finding)
	}
	return result, nil
}

// CheckAll checks every module under dir (see report.ModuleDirs).
func (s *Service) CheckAll(ctx context.Context, dir string) ([]app.CheckReport, error) {
	modules, err := report.ModuleDirs(dir)
	if err != nil {
		return nil, err
	}
	results := make([]app.CheckReport, 0, len(modules))
	for _, module := range modules {
		result, err := s.Check(ctx, module)
		if err != nil {
			result = app.CheckReport{
				Schema: app.CurrentReportSchema, Dir: module, Warnings: []string{}, Error: err.Error(),
				Findings: []app.CheckFinding{},
			}
		}
		results = append(results, result)
	}
	return results, nil
}

type moduleFacts struct {
	dir         string
	root        string // module root; "" outside a module
	repoRoot    string
	goMod       string
	goDirective string
	goLine      int
}

func loadModule(dir string) (moduleFacts, error) {
	facts := moduleFacts{dir: dir}
	root, ok := workspace.ModuleRoot(dir)
	if !ok {
		facts.repoRoot = workspace.Boundary(dir)
		return facts, nil
	}
	facts.root = root
	facts.repoRoot = workspace.Boundary(root)
	facts.goMod = filepath.Join(root, "go.mod")
	version, line, _, err := workspace.GoDirective(facts.goMod)
	if err != nil {
		return moduleFacts{}, err
	}
	facts.goDirective, facts.goLine = version, line
	return facts, nil
}

func skippedChecks(dir string) (map[string]bool, error) {
	config, found, err := fgmconfig.LoadNearest(dir)
	if err != nil || !found {
		return map[string]bool{}, err
	}
	skip := map[string]bool{}
	for _, id := range config.File.Check.Skip {
		skip[strings.TrimSpace(id)] = true
	}
	return skip, nil
}

// C1: the selected or detected linter cannot run on the resolved Go.
func checkLintCompatibility(current app.CurrentReport, module moduleFacts) []app.CheckFinding {
	lint := current.GolangCILint
	if lint == nil || lint.Compatible == nil || *lint.Compatible || current.Go == nil {
		return nil
	}
	file, line := lintSourceLocation(current, module)
	return []app.CheckFinding{{
		ID:       "C1",
		Severity: "ERROR",
		File:     file,
		Line:     line,
		Message:  fmt.Sprintf("golangci-lint %s cannot run on Go %s (compatibility catalog)", lint.Version, current.Go.Version),
		Hint:     fmt.Sprintf("pin a compatible linter (fgm versions golangci-lint --remote --go %s) or use fgm pin golangci-lint auto", current.Go.Version),
	}}
}

// C2: lint config schema does not match the linter major version.
func checkLintConfigSchema(current app.CurrentReport, module moduleFacts) ([]app.CheckFinding, error) {
	lint := current.GolangCILint
	if lint == nil {
		return nil, nil
	}
	major, err := strconv.Atoi(strings.SplitN(strings.TrimPrefix(lint.Version, "v"), ".", 2)[0])
	if err != nil {
		return nil, nil
	}
	start := module.root
	if start == "" {
		start = module.dir
	}
	configPath := ""
	for _, dir := range workspace.Ancestors(start, module.repoRoot) {
		for _, name := range []string{".golangci.yml", ".golangci.yaml"} {
			if info, err := os.Stat(filepath.Join(dir, name)); err == nil && !info.IsDir() {
				configPath = filepath.Join(dir, name)
				break
			}
		}
		if configPath != "" {
			break
		}
	}
	if configPath == "" {
		return nil, nil
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", configPath, err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(content, &document); err != nil {
		return []app.CheckFinding{{ID: "C2", Severity: "ERROR", File: configPath, Line: 1, Message: "lint config is not valid YAML: " + err.Error()}}, nil
	}
	schema, line := "", 1
	if len(document.Content) > 0 && document.Content[0].Kind == yaml.MappingNode {
		mapping := document.Content[0]
		for index := 0; index+1 < len(mapping.Content); index += 2 {
			if mapping.Content[index].Value == "version" {
				schema, line = strings.TrimSpace(mapping.Content[index+1].Value), mapping.Content[index].Line
			}
		}
	}
	switch {
	case major >= 2 && schema != "2":
		return []app.CheckFinding{{
			ID: "C2", Severity: "ERROR", File: configPath, Line: line,
			Message: fmt.Sprintf("golangci-lint %s needs a v2 config (version: \"2\"), found %s", lint.Version, describeSchema(schema)),
			Hint:    "run golangci-lint migrate, or fgm lint init --force for a fresh v2 config",
		}}, nil
	case major == 1 && schema == "2":
		return []app.CheckFinding{{
			ID: "C2", Severity: "ERROR", File: configPath, Line: line,
			Message: fmt.Sprintf("golangci-lint %s cannot read a version: \"2\" config", lint.Version),
			Hint:    "pin a v2 linter (fgm pin golangci-lint auto) or use a v1 config",
		}}, nil
	}
	return nil, nil
}

func describeSchema(schema string) string {
	if schema == "" {
		return "no version key"
	}
	return fmt.Sprintf("version: %q", schema)
}

// C3: linter pin sources disagree.
func checkLintPinsAgree(module moduleFacts) ([]app.CheckFinding, error) {
	var pins []lintdetect.Pin
	if path, found, err := fgmconfig.PinnedLintPath(module.dir); err != nil {
		return nil, err
	} else if found {
		version, _, err := fgmconfig.ResolvePinnedLint(module.dir)
		if err != nil {
			return nil, err
		}
		pins = append(pins, lintdetect.Pin{Version: "v" + strings.TrimPrefix(version, "v"), File: path, Line: keyLine(path, "golangci_lint")})
	}
	enabled, err := fgmconfig.LintDetectionEnabled(module.dir)
	if err != nil {
		return nil, err
	}
	if enabled {
		detected, err := lintdetect.Detect(module.dir)
		if err != nil {
			return nil, err
		}
		pins = append(pins, detected.Pins...)
	}
	if !(lintdetect.Result{Pins: pins}).Conflict() {
		return nil, nil
	}
	candidates := make([]string, 0, len(pins))
	for _, pin := range pins {
		candidates = append(candidates, fmt.Sprintf("%s (%s:%d)", pin.Version, relative(module.dir, pin.File), pin.Line))
	}
	return []app.CheckFinding{{
		ID: "C3", Severity: "ERROR", File: pins[0].File, Line: pins[0].Line,
		Message: "golangci-lint pins disagree: " + strings.Join(candidates, ", "),
		Hint:    "make every source name the same version",
	}}, nil
}

// C4/C4b: actions/setup-go go-version versus the go directive and the resolved Go.
func checkSetupGo(current app.CurrentReport, module moduleFacts) ([]app.CheckFinding, error) {
	if module.root == "" {
		return nil, nil
	}
	steps, err := moduleWorkflowSteps(module)
	if err != nil {
		return nil, err
	}
	var findings []app.CheckFinding
	for _, step := range steps {
		if !strings.HasPrefix(step.uses, "actions/setup-go@") || step.goVersionFile != "" || step.goVersion == "" {
			continue
		}
		if strings.Contains(step.goVersion, "${{") {
			findings = append(findings, app.CheckFinding{
				ID: "C4", Severity: "INFO", File: step.file, Line: step.goVersionLine,
				Message: fmt.Sprintf("setup-go go-version %s is an expression; not checked", step.goVersion),
			})
			continue
		}
		workflowMinor, ok := minorOf(step.goVersion)
		if !ok {
			continue
		}
		directiveMinor, _ := minorOf(module.goDirective)
		switch {
		case versionutil.CompareVersions(workflowMinor, directiveMinor) < 0:
			findings = append(findings, app.CheckFinding{
				ID: "C4", Severity: "ERROR", File: step.file, Line: step.goVersionLine,
				Message: fmt.Sprintf("setup-go go-version %s is older than the go %s directive in %s", step.goVersion, module.goDirective, relative(module.dir, module.goMod)),
				Hint:    "use go-version-file: go.mod, or raise go-version",
			})
		case current.Go != nil:
			if resolvedMinor, ok := minorOf(current.Go.Version); ok && resolvedMinor != workflowMinor {
				findings = append(findings, app.CheckFinding{
					ID: "C4b", Severity: "WARN", File: step.file, Line: step.goVersionLine,
					Message: fmt.Sprintf("setup-go go-version %s differs from the resolved Go %s", step.goVersion, current.Go.Version),
					Hint:    "use go-version-file: go.mod, or align go-version with the resolved Go",
				})
			}
		}
	}
	return findings, nil
}

type workflowStep struct {
	file          string
	uses          string
	goVersion     string
	goVersionLine int
	goVersionFile string
	run           string
	runFirstLine  int
}

// moduleWorkflowSteps returns workflow steps that apply to the module, using the same
// working-directory rule as lint pin detection.
func moduleWorkflowSteps(module moduleFacts) ([]workflowStep, error) {
	files, err := lintdetect.WorkflowFiles(module.repoRoot)
	if err != nil {
		return nil, err
	}
	modulePath := lintdetect.RelativeSlash(module.repoRoot, module.root)
	singleModule := -1
	var steps []workflowStep
	for _, file := range files {
		parsed, err := lintdetect.ParseWorkflowSteps(file)
		if errors.Is(err, lintdetect.ErrInvalidWorkflow) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, step := range parsed {
			if step.WorkingDirectory != modulePath {
				if singleModule == -1 {
					modules, err := workspace.ListModuleDirs(module.repoRoot)
					if err != nil {
						return nil, err
					}
					singleModule = len(modules)
				}
				if singleModule != 1 {
					continue
				}
			}
			steps = append(steps, workflowStep{
				file: file, uses: step.Uses,
				goVersion: step.GoVersion, goVersionLine: step.GoVersionLine, goVersionFile: step.GoVersionFile,
				run: step.Run, runFirstLine: step.RunFirstLine,
			})
		}
	}
	return steps, nil
}

// C5: Dockerfile golang base image older than the go directive.
func checkDockerfiles(module moduleFacts) ([]app.CheckFinding, error) {
	if module.root == "" {
		return nil, nil
	}
	var findings []app.CheckFinding
	for _, dir := range workspace.Ancestors(module.root, module.repoRoot) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", dir, err)
		}
		for _, entry := range entries {
			if !entry.Type().IsRegular() || !isDockerfile(entry.Name()) {
				continue
			}
			found, err := checkDockerfile(filepath.Join(dir, entry.Name()), module)
			if err != nil {
				return nil, err
			}
			findings = append(findings, found...)
		}
	}
	return findings, nil
}

func isDockerfile(name string) bool {
	return name == "Dockerfile" || name == "Containerfile" || strings.HasPrefix(name, "Dockerfile.") || strings.HasSuffix(name, ".Dockerfile")
}

func checkDockerfile(path string, module moduleFacts) ([]app.CheckFinding, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	var findings []app.CheckFinding
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		match := dockerFromPattern.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		tag := match[1]
		if strings.Contains(tag, "$") {
			findings = append(findings, app.CheckFinding{
				ID: "C5", Severity: "INFO", File: path, Line: line,
				Message: fmt.Sprintf("golang:%s uses a build argument; not checked", tag),
			})
			continue
		}
		version := tagVersionPattern.FindStringSubmatch(tag)
		if version == nil {
			continue
		}
		if dockerTagOlder(version[1], module.goDirective) {
			findings = append(findings, app.CheckFinding{
				ID: "C5", Severity: "ERROR", File: path, Line: line,
				Message: fmt.Sprintf("golang:%s is older than the go %s directive in %s", tag, module.goDirective, relative(module.dir, module.goMod)),
				Hint:    "raise the golang image tag",
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return findings, nil
}

// dockerTagOlder compares a full tag (1.26.3) exactly and a floating tag (1.26) by minor,
// since the floating tag always carries the newest patch of that minor.
func dockerTagOlder(tag string, goDirective string) bool {
	if strings.Count(tag, ".") >= 2 {
		return versionutil.CompareVersions(tag, goDirective) < 0
	}
	directiveMinor, _ := minorOf(goDirective)
	return versionutil.CompareVersions(tag, directiveMinor) < 0
}

// C6: go.work go version lower than a used module's go directive.
func checkGoWork(module moduleFacts) ([]app.CheckFinding, error) {
	if module.root == "" {
		return nil, nil
	}
	workRoot, ok := workspace.WorkRoot(module.root)
	if !ok {
		return nil, nil
	}
	workPath := filepath.Join(workRoot, "go.work")
	uses, err := report.ModuleDirs(workRoot)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(uses, module.root) {
		return nil, nil
	}
	version, line, found, err := workspace.GoDirective(workPath)
	if err != nil || !found {
		return nil, err
	}
	if versionutil.CompareVersions(version, module.goDirective) >= 0 {
		return nil, nil
	}
	return []app.CheckFinding{{
		ID: "C6", Severity: "ERROR", File: workPath, Line: line,
		Message: fmt.Sprintf("go.work declares go %s, lower than the go %s directive in %s", version, module.goDirective, relative(module.dir, module.goMod)),
		Hint:    "raise the go directive in go.work (go work use -r . updates it)",
	}}, nil
}

func lintSourceLocation(current app.CurrentReport, module moduleFacts) (string, int) {
	lint := current.GolangCILint
	if lint.SourceFile == "" {
		return "", 0
	}
	path := filepath.Join(current.Dir, filepath.FromSlash(lint.SourceFile))
	if lint.Source == "config" {
		return path, keyLine(path, "golangci_lint")
	}
	return path, lint.SourceLine
}

// keyLine returns the 1-based line of the first "key =" assignment in a TOML file, or 0.
func keyLine(path string, key string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	for index, line := range strings.Split(string(content), "\n") {
		name, _, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(name) == key {
			return index + 1
		}
	}
	return 0
}

func minorOf(version string) (string, bool) {
	match := minorPattern.FindStringSubmatch(version)
	if match == nil {
		return "", false
	}
	return match[1] + "." + match[2], true
}

func relative(dir string, path string) string {
	if path == "" {
		return ""
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
