// Package lintdetect finds golangci-lint versions that a repository already pins for its CI,
// so fgm selects the linter CI runs instead of a catalog guess. Detection reads files only.
package lintdetect

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/kostikovk/fgm/internal/workspace"
	"gopkg.in/yaml.v3"
)

// Pin is one golangci-lint version found in a repository file.
type Pin struct {
	Version string // normalized with a "v" prefix
	File    string // absolute path
	Line    int
	Kind    string // "tool-versions", "mise", "go.mod", "workflow", "makefile"
}

// Result holds every pin that applies to a module, in detector order.
type Result struct {
	Pins []Pin
}

// Agreed returns the first pin when all pins name one version.
func (r Result) Agreed() (Pin, bool) {
	if len(r.Pins) == 0 || r.Conflict() {
		return Pin{}, false
	}
	return r.Pins[0], true
}

// Conflict reports whether pins name more than one version.
func (r Result) Conflict() bool {
	for _, pin := range r.Pins {
		if pin.Version != r.Pins[0].Version {
			return true
		}
	}
	return false
}

var (
	semverPattern       = regexp.MustCompile(`v?([0-9]+\.[0-9]+\.[0-9]+)`)
	makeVariablePattern = regexp.MustCompile(`^\s*(?:export\s+)?GOLANGCI_LINT_VERSION\s*(?:\?=|::=|:=|=)\s*v?([0-9]+\.[0-9]+\.[0-9]+)\s*(?:#.*)?$`)
	makeAtPattern       = regexp.MustCompile(`golangci-lint@v?([0-9]+\.[0-9]+\.[0-9]+)`)
	installShPattern    = regexp.MustCompile(`\sv([0-9]+\.[0-9]+\.[0-9]+)\b`)
	lintToolPattern     = regexp.MustCompile(`^github\.com/golangci/golangci-lint(/v[0-9]+)?/cmd/golangci-lint$`)
	exactVersionPattern = regexp.MustCompile(`^v?([0-9]+\.[0-9]+\.[0-9]+)$`)
)

// Detect returns golangci-lint pins that apply to the module containing workDir, searching from
// the module root upward to the repository root.
func Detect(workDir string) (Result, error) {
	moduleRoot, ok := workspace.ModuleRoot(workDir)
	if !ok {
		moduleRoot = filepath.Clean(workDir)
	}
	boundary := workspace.Boundary(moduleRoot)
	dirs := workspace.Ancestors(moduleRoot, boundary)

	var pins []Pin
	for _, detect := range []func() ([]Pin, error){
		func() ([]Pin, error) { return nearest(dirs, []string{".tool-versions"}, scanToolVersions) },
		func() ([]Pin, error) { return nearest(dirs, []string{"mise.toml", ".mise.toml"}, scanMise) },
		func() ([]Pin, error) { return scanFileIfExists(filepath.Join(moduleRoot, "go.mod"), scanGoModTool) },
		func() ([]Pin, error) { return detectWorkflows(boundary, moduleRoot) },
		func() ([]Pin, error) { return nearestMakefiles(dirs) },
	} {
		found, err := detect()
		if err != nil {
			return Result{}, err
		}
		pins = append(pins, found...)
	}
	return Result{Pins: pins}, nil
}

type lineScanner func(path string, lines []string) []Pin

// nearest returns pins from the first directory whose named files yield any.
func nearest(dirs []string, names []string, scan lineScanner) ([]Pin, error) {
	for _, dir := range dirs {
		var pins []Pin
		for _, name := range names {
			found, err := scanFileIfExists(filepath.Join(dir, name), scan)
			if err != nil {
				return nil, err
			}
			pins = append(pins, found...)
		}
		if len(pins) > 0 {
			return pins, nil
		}
	}
	return nil, nil
}

func nearestMakefiles(dirs []string) ([]Pin, error) {
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read %s: %w", dir, err)
		}
		present := map[string]bool{}
		var names []string
		for _, entry := range entries {
			if entry.Type().IsRegular() {
				present[entry.Name()] = true
				if strings.HasSuffix(entry.Name(), ".mk") {
					names = append(names, entry.Name())
				}
			}
		}
		// make reads only the first of these. Names come from the directory listing so a
		// case-insensitive filesystem does not report the same file twice.
		for _, candidate := range []string{"GNUmakefile", "makefile", "Makefile"} {
			if present[candidate] {
				names = append([]string{candidate}, names...)
				break
			}
		}
		var pins []Pin
		for _, name := range names {
			found, err := scanFileIfExists(filepath.Join(dir, name), scanMakefile)
			if err != nil {
				return nil, err
			}
			pins = append(pins, found...)
		}
		if len(pins) > 0 {
			return pins, nil
		}
	}
	return nil, nil
}

func scanFileIfExists(path string, scan lineScanner) ([]Pin, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return scan(path, lines), nil
}

func scanToolVersions(path string, lines []string) []Pin {
	var pins []Pin
	for index, line := range lines {
		fields := strings.Fields(strings.SplitN(line, "#", 2)[0])
		if len(fields) >= 2 && fields[0] == "golangci-lint" {
			if version, ok := exactVersion(fields[1]); ok {
				pins = append(pins, Pin{Version: version, File: path, Line: index + 1, Kind: "tool-versions"})
			}
		}
	}
	return pins
}

func scanMise(path string, lines []string) []Pin {
	var pins []Pin
	inTools := false
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inTools = trimmed == "[tools]"
			continue
		}
		if !inTools {
			continue
		}
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok || strings.Trim(strings.TrimSpace(key), `"'`) != "golangci-lint" {
			continue
		}
		if match := semverPattern.FindStringSubmatch(value); match != nil {
			pins = append(pins, Pin{Version: "v" + match[1], File: path, Line: index + 1, Kind: "mise"})
		}
	}
	return pins
}

func scanGoModTool(path string, lines []string) []Pin {
	toolModules := map[string]bool{}
	requires := map[string]Pin{}
	block := ""
	for index, line := range lines {
		trimmed := strings.TrimSpace(strings.SplitN(line, "//", 2)[0])
		switch {
		case trimmed == ")":
			block = ""
			continue
		case strings.HasSuffix(trimmed, "("):
			block = strings.TrimSpace(strings.TrimSuffix(trimmed, "("))
			continue
		}
		directive, rest := block, trimmed
		if block == "" {
			directive, rest, _ = strings.Cut(trimmed, " ")
		}
		fields := strings.Fields(rest)
		switch directive {
		case "tool":
			if len(fields) == 1 && lintToolPattern.MatchString(fields[0]) {
				toolModules[strings.TrimSuffix(fields[0], "/cmd/golangci-lint")] = true
			}
		case "require":
			if len(fields) >= 2 {
				if version, ok := exactVersion(fields[1]); ok {
					requires[fields[0]] = Pin{Version: version, File: path, Line: index + 1, Kind: "go.mod"}
				}
			}
		}
	}
	var pins []Pin
	for module := range toolModules {
		if pin, ok := requires[module]; ok {
			pins = append(pins, pin)
		}
	}
	return pins
}

func scanMakefile(path string, lines []string) []Pin {
	var pins []Pin
	for index, line := range lines {
		var match []string
		switch {
		case makeVariablePattern.MatchString(line):
			match = makeVariablePattern.FindStringSubmatch(line)
		case makeAtPattern.MatchString(line):
			match = makeAtPattern.FindStringSubmatch(line)
		case strings.Contains(line, "golangci-lint") && strings.Contains(line, "install.sh"):
			match = installShPattern.FindStringSubmatch(line)
		}
		if match != nil {
			pins = append(pins, Pin{Version: "v" + match[1], File: path, Line: index + 1, Kind: "makefile"})
		}
	}
	return pins
}

// detectWorkflows applies a golangci-lint-action pin to the module when the step's or job's
// working directory equals the module path, or when the repository has a single module.
// Reusable workflows from other repositories are not followed.
func detectWorkflows(repoRoot string, moduleRoot string) ([]Pin, error) {
	files, err := WorkflowFiles(repoRoot)
	if err != nil {
		return nil, err
	}
	modulePath := RelativeSlash(repoRoot, moduleRoot)

	singleModule := -1
	var pins []Pin
	for _, file := range files {
		steps, err := ParseWorkflowSteps(file)
		if err != nil {
			return nil, err
		}
		for _, step := range steps {
			if !strings.HasPrefix(step.Uses, "golangci/golangci-lint-action@") || step.Version == "" {
				continue
			}
			version, ok := exactVersion(step.Version)
			if !ok {
				continue
			}
			if step.WorkingDirectory != modulePath {
				if singleModule == -1 {
					modules, err := workspace.ListModuleDirs(repoRoot)
					if err != nil {
						return nil, err
					}
					singleModule = len(modules)
				}
				if singleModule != 1 {
					continue
				}
			}
			pins = append(pins, Pin{Version: version, File: file, Line: step.VersionLine, Kind: "workflow"})
		}
	}
	return pins, nil
}

// WorkflowStep is the subset of a GitHub Actions step that fgm inspects.
type WorkflowStep struct {
	Uses             string
	Version          string // with.version (golangci-lint-action)
	VersionLine      int
	GoVersion        string // with.go-version (setup-go)
	GoVersionLine    int
	GoVersionFile    string // with.go-version-file (setup-go)
	WorkingDirectory string // with.working-directory, else the job's defaults.run.working-directory, cleaned; "." when unset
	Line             int
}

// WorkflowFiles lists .github/workflows/*.yml and *.yaml under repoRoot in lexical order.
func WorkflowFiles(repoRoot string) ([]string, error) {
	var files []string
	for _, pattern := range []string{"*.yml", "*.yaml"} {
		matches, err := filepath.Glob(filepath.Join(repoRoot, ".github", "workflows", pattern))
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}
	slices.Sort(files)
	return files, nil
}

// ParseWorkflowSteps returns every step of every job in a workflow file.
func ParseWorkflowSteps(file string) ([]WorkflowStep, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(content, &document); err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	if len(document.Content) == 0 {
		return nil, nil
	}
	var steps []WorkflowStep
	jobs := mappingValue(document.Content[0], "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return nil, nil
	}
	for index := 1; index < len(jobs.Content); index += 2 {
		job := jobs.Content[index]
		jobDir := scalar(mappingValue(mappingValue(mappingValue(job, "defaults"), "run"), "working-directory"))
		stepList := mappingValue(job, "steps")
		if stepList == nil || stepList.Kind != yaml.SequenceNode {
			continue
		}
		for _, node := range stepList.Content {
			with := mappingValue(node, "with")
			step := WorkflowStep{
				Uses:          scalar(mappingValue(node, "uses")),
				Version:       scalar(mappingValue(with, "version")),
				GoVersion:     scalar(mappingValue(with, "go-version")),
				GoVersionFile: scalar(mappingValue(with, "go-version-file")),
				Line:          node.Line,
			}
			if version := mappingValue(with, "version"); version != nil {
				step.VersionLine = version.Line
			}
			if goVersion := mappingValue(with, "go-version"); goVersion != nil {
				step.GoVersionLine = goVersion.Line
			}
			dir := scalar(mappingValue(with, "working-directory"))
			if dir == "" {
				dir = jobDir
			}
			step.WorkingDirectory = cleanSlash(dir)
			steps = append(steps, step)
		}
	}
	return steps, nil
}

// RelativeSlash returns target relative to base with forward slashes ("." for base itself).
func RelativeSlash(base string, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return filepath.ToSlash(target)
	}
	return filepath.ToSlash(rel)
}

func cleanSlash(dir string) string {
	if dir == "" {
		return "."
	}
	return path.Clean(filepath.ToSlash(dir))
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1]
		}
	}
	return nil
}

func scalar(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return strings.TrimSpace(node.Value)
}

func exactVersion(value string) (string, bool) {
	match := exactVersionPattern.FindStringSubmatch(strings.Trim(strings.TrimSpace(value), `"'`))
	if match == nil {
		return "", false
	}
	return "v" + match[1], true
}
