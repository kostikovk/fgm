package repocheck

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/workspace"
)

// pathToken matches a shell or Make path word. Make and shell expansions may contain spaces, so
// $(...) and ${...} groups count as part of the word.
const pathToken = `(?:\$\$?\([^)]*\)|\$\$?\{[^}]*\}|[^\s;&|'"(){}])+`

var (
	// toolPathPattern finds a path ending in golangci-lint or go, e.g.
	// "$(shell go env GOPATH)/bin/golangci-lint".
	toolPathPattern = regexp.MustCompile(`(` + pathToken + `)/(golangci-lint|go)(?:\.exe)?(?:$|[\s;&|'")])`)
	// probeBefore and probeAfter recognise invocations that only test, locate, or ask the version
	// of a binary; they do not run it.
	probeBefore = regexp.MustCompile(`(?:\btest\s+-[xfe]|\[\s+-[xfe]|\bwhich|\bcommand\s+-v|\btype)\s*$`)
	probeAfter  = regexp.MustCompile(`^\s+(?:--?version|version)\b`)
	// installShInto and goInstallInto find installs of an exact version into a directory.
	installShInto = regexp.MustCompile(`install\.sh[^\n]*?-b\s+(` + pathToken + `)\s+v?([0-9]+\.[0-9]+\.[0-9]+)`)
	goInstallInto = regexp.MustCompile(`GOBIN=(` + pathToken + `)\s+go\s+install\s+\S*@v?([0-9]+\.[0-9]+\.[0-9]+)`)
)

// gobinPrefixes name directories that are GOBIN itself, so the tool sits right below them.
var gobinPrefixes = []string{"$(GOBIN)", "${GOBIN}", "$GOBIN", "$$GOBIN", "$${GOBIN}"}

// bypassHit is a C7 finding plus the expanded path it was raised for.
type bypassHit struct {
	finding app.CheckFinding
	path    string
}

// C7: a build script runs golangci-lint or go by a path that bypasses PATH, so fgm exec cannot
// route it to the selected version.
func checkToolBypass(current app.CurrentReport, module moduleFacts) ([]app.CheckFinding, error) {
	if module.root == "" {
		return nil, nil
	}
	var findings []app.CheckFinding
	for _, dir := range workspace.Ancestors(module.root, module.repoRoot) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", dir, err)
		}
		var makefiles []string
		for _, entry := range entries {
			if entry.Type().IsRegular() && isMakefile(entry.Name()) {
				makefiles = append(makefiles, filepath.Join(dir, entry.Name()))
			}
		}
		model := parseMakefiles(dir, makefiles)
		for _, target := range model.order {
			for _, line := range model.rules[target].recipe {
				hits := bypassFindings(current, line.file, line.number, line.text, model.vars)
				findings = append(findings, model.applyInstallGuards(target, hits)...)
			}
		}
	}

	steps, err := moduleWorkflowSteps(module)
	if err != nil {
		return nil, err
	}
	for _, step := range steps {
		for index, line := range strings.Split(step.run, "\n") {
			for _, hit := range bypassFindings(current, step.file, step.runFirstLine+index, line, nil) {
				findings = append(findings, hit.finding)
			}
		}
	}
	return findings, nil
}

// bypassFindings reports at most one hit per tool on a line, skipping lines that only install a
// tool and invocations that only probe it.
func bypassFindings(current app.CurrentReport, file string, lineNumber int, line string, vars map[string]string) []bypassHit {
	line = strings.SplitN(line, "#", 2)[0]
	if strings.Contains(line, "install.sh") || strings.Contains(line, "go install") || strings.Contains(line, "curl ") || strings.Contains(line, "wget ") {
		return nil
	}
	expanded := expandMake(line, vars, nil)
	original := toolPathPattern.FindAllStringSubmatch(line, -1)
	reported := map[string]bool{}
	var hits []bypassHit
	for index, match := range toolPathPattern.FindAllStringSubmatchIndex(expanded, -1) {
		prefix, tool := expanded[match[2]:match[3]], expanded[match[4]:match[5]]
		if reported[tool] || !bypassesPath(prefix) {
			continue
		}
		if probeBefore.MatchString(expanded[:match[2]]) || probeAfter.MatchString(expanded[match[5]:]) {
			continue
		}
		reported[tool] = true

		toolPath := prefix + "/" + tool
		invoked := toolPath
		if expanded != line && len(original) > index && original[index][2] == tool {
			if written := original[index][1] + "/" + tool; written != toolPath {
				invoked = fmt.Sprintf("%s (expands to %s)", written, toolPath)
			}
		}
		message := fmt.Sprintf("%s is invoked as %s, which bypasses fgm exec", tool, invoked)
		switch {
		case tool == "golangci-lint" && current.GolangCILint != nil:
			message += fmt.Sprintf(" (selected: %s)", current.GolangCILint.Version)
		case tool == "go" && current.Go != nil:
			message += fmt.Sprintf(" (selected: %s)", current.Go.Version)
		}
		hits = append(hits, bypassHit{
			path: toolPath,
			finding: app.CheckFinding{
				ID: "C7", Severity: "WARN", File: file, Line: lineNumber,
				Message: message,
				Hint:    fmt.Sprintf("call %s via PATH, or run fgm exec -- %s ...", tool, tool),
			},
		})
	}
	return hits
}

// bypassesPath reports whether prefix is a shared tool directory (GOPATH/bin, GOBIN, ~/go/bin,
// or any absolute .../bin) rather than PATH lookup or a project-local bin directory.
func bypassesPath(prefix string) bool {
	if slices.Contains(gobinPrefixes, prefix) {
		return true
	}
	dir, ok := strings.CutSuffix(prefix, "/bin")
	if !ok {
		return false
	}
	switch {
	case strings.Contains(dir, "GOPATH"):
		return true
	case dir == "~/go" || strings.HasSuffix(dir, "HOME/go") || strings.HasSuffix(dir, "HOME}/go") || strings.HasSuffix(dir, "HOME)/go"):
		return true
	case strings.HasPrefix(dir, "/"):
		return true
	}
	return false
}

func isMakefile(name string) bool {
	return name == "Makefile" || name == "makefile" || name == "GNUmakefile" || strings.HasSuffix(name, ".mk")
}

var (
	makeAssignment = regexp.MustCompile(`^(?:export\s+|override\s+)?([A-Za-z_][A-Za-z0-9_.-]*)\s*(?:\?=|::=|:=|=)\s*(.*)$`)
	makeInclude    = regexp.MustCompile(`^-?s?include\s+(.+)$`)
	makeRuleLine   = regexp.MustCompile(`^([^:=\s][^:=]*?)\s*::?(?:\s+(.*))?$`)
	makeReference  = regexp.MustCompile(`\$[({]([A-Za-z_][A-Za-z0-9_.-]*)[)}]`)
	goEnvShell     = regexp.MustCompile(`\$\(shell\s+go\s+env\s+(GOPATH|GOBIN)\s*\)`)
)

// maxMakeExpansionDepth bounds recursive variable expansion.
const maxMakeExpansionDepth = 4

type recipeLine struct {
	file   string
	number int
	text   string // command without the tab and recipe prefixes (@ silent, - ignore errors, + always run)
}

type makeRule struct {
	prereqs []string
	recipe  []recipeLine
}

// makeModel is what C7 can see of a directory's makefiles without running Make: simple variable
// assignments, and rules with their prerequisites and recipe lines.
type makeModel struct {
	vars  map[string]string
	rules map[string]*makeRule
	order []string
}

// parseMakefiles reads the makefiles in dir and the files they include by a literal path.
// Functions, conditionals, += and target-specific variables are ignored.
func parseMakefiles(dir string, makefiles []string) makeModel {
	model := makeModel{vars: map[string]string{}, rules: map[string]*makeRule{}}
	seen := map[string]bool{}
	var read func(file string)
	read = func(file string) {
		if seen[file] {
			return
		}
		seen[file] = true
		content, err := os.ReadFile(file)
		if err != nil {
			return
		}
		var current []string
		continuing := false
		for index, raw := range strings.Split(string(content), "\n") {
			if continuing || (strings.HasPrefix(raw, "\t") && current != nil) {
				text := strings.TrimLeft(raw, "\t ")
				if !continuing {
					text = strings.TrimLeft(text, "@-+ \t")
				}
				for _, target := range current {
					rule := model.rules[target]
					rule.recipe = append(rule.recipe, recipeLine{file: file, number: index + 1, text: text})
				}
				continuing = strings.HasSuffix(strings.TrimRight(raw, " \t"), "\\")
				continue
			}
			line := strings.TrimSpace(strings.SplitN(raw, "#", 2)[0])
			switch {
			case line == "":
			case makeInclude.MatchString(line):
				for name := range strings.FieldsSeq(makeInclude.FindStringSubmatch(line)[1]) {
					if !strings.Contains(name, "$") {
						read(filepath.Join(dir, name))
					}
				}
				current = nil
			case makeAssignment.MatchString(line):
				match := makeAssignment.FindStringSubmatch(line)
				model.vars[match[1]] = strings.TrimSpace(match[2])
				current = nil
			case makeRuleLine.MatchString(line):
				match := makeRuleLine.FindStringSubmatch(line)
				prereqs, _, _ := strings.Cut(match[2], ";")
				current = strings.Fields(match[1])
				for _, target := range current {
					rule, ok := model.rules[target]
					if !ok {
						rule = &makeRule{}
						model.rules[target] = rule
						model.order = append(model.order, target)
					}
					for prereq := range strings.FieldsSeq(prereqs) {
						if prereq != "|" {
							rule.prereqs = append(rule.prereqs, prereq)
						}
					}
				}
			default:
				current = nil
			}
		}
	}
	for _, file := range makefiles {
		read(file)
	}
	return model
}

// applyInstallGuards drops a hit when target, or a prerequisite it reaches, reinstalls an exact
// version into the hit's directory whenever the binary there reports another version: the run then
// always uses that version, which is consistent (or, when it differs from the selection, a pin
// disagreement that C3 reports). An install into that directory without a version check keeps
// the hit and says so.
func (m makeModel) applyInstallGuards(target string, hits []bypassHit) []app.CheckFinding {
	findings := make([]app.CheckFinding, 0, len(hits))
	if len(hits) == 0 {
		return findings
	}
	script := m.reachableRecipe(target)
	for _, hit := range hits {
		dir := path.Dir(hit.path)
		version, installed := installedVersion(script, dir)
		if !installed {
			findings = append(findings, hit.finding)
			continue
		}
		checked := (strings.Contains(script, hit.path+" version") || strings.Contains(script, hit.path+" --version")) &&
			strings.Contains(script, "grep") && strings.Contains(script, version)
		if checked {
			continue
		}
		hit.finding.Message += " (the install target does not check the installed version)"
		findings = append(findings, hit.finding)
	}
	return findings
}

// reachableRecipe returns the expanded recipes of target and every prerequisite it reaches.
func (m makeModel) reachableRecipe(target string) string {
	visited := map[string]bool{}
	var lines []string
	var visit func(name string)
	visit = func(name string) {
		rule, ok := m.rules[name]
		if !ok || visited[name] {
			return
		}
		visited[name] = true
		for _, line := range rule.recipe {
			lines = append(lines, expandMake(line.text, m.vars, nil))
		}
		for _, prereq := range rule.prereqs {
			visit(prereq)
		}
	}
	visit(target)
	return strings.Join(lines, "\n")
}

func installedVersion(script string, dir string) (string, bool) {
	for _, pattern := range []*regexp.Regexp{installShInto, goInstallInto} {
		for _, match := range pattern.FindAllStringSubmatch(script, -1) {
			if strings.TrimRight(match[1], "/;") == dir {
				return match[2], true
			}
		}
	}
	return "", false
}

// expandMake replaces $(VAR) and ${VAR} with visible assignments, recursively up to
// maxMakeExpansionDepth. References in a cycle, or without an assignment (environment or
// command line), stay as written, so they are never reported. $(shell go env GOBIN) becomes the
// GOBIN token C7 already knows.
func expandMake(text string, vars map[string]string, stack []string) string {
	text = goEnvShell.ReplaceAllStringFunc(text, func(call string) string {
		if strings.Contains(call, "GOBIN") {
			return "$(GOBIN)"
		}
		return call
	})
	if len(vars) == 0 || len(stack) >= maxMakeExpansionDepth {
		return text
	}
	return makeReference.ReplaceAllStringFunc(text, func(reference string) string {
		name := makeReference.FindStringSubmatch(reference)[1]
		value, ok := vars[name]
		if !ok || slices.Contains(stack, name) {
			return reference
		}
		expanded := expandMake(value, vars, append(stack, name))
		if makeReference.MatchString(expanded) && strings.Contains(expanded, reference) {
			return reference
		}
		return expanded
	})
}
