package repocheck

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
		for _, entry := range entries {
			if !entry.Type().IsRegular() || !isMakefile(entry.Name()) {
				continue
			}
			name := entry.Name()
			path := filepath.Join(dir, name)
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", path, err)
			}
			for index, line := range strings.Split(string(content), "\n") {
				if !strings.HasPrefix(line, "\t") {
					continue
				}
				findings = append(findings, bypassFindings(current, path, index+1, line)...)
			}
		}
	}

	steps, err := moduleWorkflowSteps(module)
	if err != nil {
		return nil, err
	}
	for _, step := range steps {
		for index, line := range strings.Split(step.run, "\n") {
			findings = append(findings, bypassFindings(current, step.file, step.runFirstLine+index, line)...)
		}
	}
	return findings, nil
}

func bypassFindings(current app.CurrentReport, file string, lineNumber int, line string) []app.CheckFinding {
	line = strings.SplitN(line, "#", 2)[0]
	// Lines that only download or build the tool do not run it.
	if strings.Contains(line, "install.sh") || strings.Contains(line, "go install") || strings.Contains(line, "curl ") || strings.Contains(line, "wget ") {
		return nil
	}
	var findings []app.CheckFinding
	for _, match := range toolPathPattern.FindAllStringSubmatch(line, -1) {
		prefix, tool := match[1], match[2]
		if !bypassesPath(prefix) {
			continue
		}
		selected := ""
		switch {
		case tool == "golangci-lint" && current.GolangCILint != nil:
			selected = current.GolangCILint.Version
		case tool == "go" && current.Go != nil:
			selected = current.Go.Version
		}
		message := fmt.Sprintf("%s is invoked as %s/%s, which bypasses fgm exec", tool, prefix, tool)
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
	for _, gobin := range gobinPrefixes {
		if prefix == gobin {
			return true
		}
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
