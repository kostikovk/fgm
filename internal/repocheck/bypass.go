package repocheck

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/workspace"
)

// toolPathPattern finds a path token ending in golangci-lint or go, e.g.
// "$(shell go env GOPATH)/bin/golangci-lint". Make and shell expansions may contain spaces, so
// the prefix allows $(...) and ${...} groups.
var toolPathPattern = regexp.MustCompile(`((?:\$\$?\([^)]*\)|\$\$?\{[^}]*\}|[^\s;&|'"(){}])+)/(golangci-lint|go)(?:\.exe)?(?:$|[\s;&|'")])`)

// gobinPrefixes name directories that are GOBIN itself, so the tool sits right below them.
var gobinPrefixes = []string{"$(GOBIN)", "${GOBIN}", "$GOBIN", "$$GOBIN", "$${GOBIN}"}

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
		vars := makeVariables(dir, makefiles)
		for _, path := range makefiles {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", path, err)
			}
			for index, line := range strings.Split(string(content), "\n") {
				if !strings.HasPrefix(line, "\t") {
					continue
				}
				// Recipe prefixes (@ silent, - ignore errors, + always run) are not part of the command.
				recipe := strings.TrimLeft(strings.TrimPrefix(line, "\t"), "@-+ \t")
				findings = append(findings, bypassFindings(current, path, index+1, recipe, vars)...)
			}
		}
	}

	steps, err := moduleWorkflowSteps(module)
	if err != nil {
		return nil, err
	}
	for _, step := range steps {
		for index, line := range strings.Split(step.run, "\n") {
			findings = append(findings, bypassFindings(current, step.file, step.runFirstLine+index, line, nil)...)
		}
	}
	return findings, nil
}

func bypassFindings(current app.CurrentReport, file string, lineNumber int, line string, vars map[string]string) []app.CheckFinding {
	line = strings.SplitN(line, "#", 2)[0]
	// Lines that only download or build the tool do not run it.
	if strings.Contains(line, "install.sh") || strings.Contains(line, "go install") || strings.Contains(line, "curl ") || strings.Contains(line, "wget ") {
		return nil
	}
	expanded := expandMake(line, vars, nil)
	original := toolPathPattern.FindAllStringSubmatch(line, -1)
	var findings []app.CheckFinding
	for index, match := range toolPathPattern.FindAllStringSubmatch(expanded, -1) {
		prefix, tool := match[1], match[2]
		if !bypassesPath(prefix) {
			continue
		}
		invoked := fmt.Sprintf("%s/%s", prefix, tool)
		if expanded != line && len(original) > index && original[index][2] == tool {
			if written := fmt.Sprintf("%s/%s", original[index][1], tool); written != invoked {
				invoked = fmt.Sprintf("%s (expands to %s)", written, invoked)
			}
		}
		selected := ""
		switch {
		case tool == "golangci-lint" && current.GolangCILint != nil:
			selected = current.GolangCILint.Version
		case tool == "go" && current.Go != nil:
			selected = current.Go.Version
		}
		message := fmt.Sprintf("%s is invoked as %s, which bypasses fgm exec", tool, invoked)
		if selected != "" {
			message += fmt.Sprintf(" (selected: %s)", selected)
		}
		findings = append(findings, app.CheckFinding{
			ID: "C7", Severity: "WARN", File: file, Line: lineNumber,
			Message: message,
			Hint:    fmt.Sprintf("call %s via PATH, or run fgm exec -- %s ...", tool, tool),
		})
	}
	return findings
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
	makeReference  = regexp.MustCompile(`\$[({]([A-Za-z_][A-Za-z0-9_.-]*)[)}]`)
	goEnvShell     = regexp.MustCompile(`\$\(shell\s+go\s+env\s+(GOPATH|GOBIN)\s*\)`)
)

// maxMakeExpansionDepth bounds recursive variable expansion.
const maxMakeExpansionDepth = 4

// makeVariables collects simple assignments from the makefiles in dir and from files they include
// by a literal path. Make is not evaluated: functions, conditionals and += are ignored.
func makeVariables(dir string, makefiles []string) map[string]string {
	vars := map[string]string{}
	seen := map[string]bool{}
	var read func(path string)
	read = func(path string) {
		if seen[path] {
			return
		}
		seen[path] = true
		content, err := os.ReadFile(path)
		if err != nil {
			return
		}
		for line := range strings.SplitSeq(string(content), "\n") {
			if strings.HasPrefix(line, "\t") {
				continue
			}
			line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if match := makeInclude.FindStringSubmatch(line); match != nil {
				for name := range strings.FieldsSeq(match[1]) {
					if !strings.Contains(name, "$") {
						read(filepath.Join(dir, name))
					}
				}
				continue
			}
			if match := makeAssignment.FindStringSubmatch(line); match != nil {
				vars[match[1]] = strings.TrimSpace(match[2])
			}
		}
	}
	for _, path := range makefiles {
		read(path)
	}
	return vars
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
