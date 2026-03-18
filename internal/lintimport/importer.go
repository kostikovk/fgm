package lintimport

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
)

// Registry registers existing golangci-lint installations in the FGM store.
type Registry interface {
	RegisterExistingLintInstallation(version string, binaryPath string) (string, error)
}

// Importer discovers existing golangci-lint binaries and registers them into FGM.
type Importer struct {
	candidates []string
	registry   Registry
}

// Config configures an Importer.
type Config struct {
	Candidates []string
	Registry   Registry
}

// New constructs an Importer.
func New(config Config) *Importer {
	return &Importer{
		candidates: config.Candidates,
		registry:   config.Registry,
	}
}

// ImportAuto discovers existing golangci-lint binaries from known candidate paths.
func (i *Importer) ImportAuto(ctx context.Context) ([]app.ImportedLint, error) {
	seenVersions := make(map[string]struct{})
	var imported []app.ImportedLint

	for _, candidate := range i.candidates {
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}

		version, err := detectVersion(ctx, candidate)
		if err != nil {
			return nil, err
		}
		if _, seen := seenVersions[version]; seen {
			continue
		}

		if _, err := i.registry.RegisterExistingLintInstallation(version, candidate); err != nil {
			return nil, err
		}

		seenVersions[version] = struct{}{}
		imported = append(imported, app.ImportedLint{
			Version: version,
			Path:    candidate,
		})
	}

	return imported, nil
}

// DefaultCandidates returns common existing golangci-lint binary locations.
func DefaultCandidates(pathEnv string) []string {
	candidates := make(map[string]struct{})

	if binary := binaryFromPath(pathEnv); binary != "" {
		candidates[binary] = struct{}{}
	}

	for _, path := range []string{
		"/opt/homebrew/bin/" + binaryName(),
		"/usr/local/bin/" + binaryName(),
	} {
		candidates[path] = struct{}{}
	}

	for _, pattern := range []string{
		"/opt/homebrew/Cellar/golangci-lint/*/bin/" + binaryName(),
		"/usr/local/Cellar/golangci-lint/*/bin/" + binaryName(),
	} {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			candidates[match] = struct{}{}
		}
	}

	ordered := make([]string, 0, len(candidates))
	for candidate := range candidates {
		ordered = append(ordered, candidate)
	}

	return ordered
}

func detectVersion(ctx context.Context, binaryPath string) (string, error) {
	cmd := exec.CommandContext(ctx, binaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run %s version: %w", binaryPath, err)
	}

	fields := strings.Fields(string(output))
	for idx, field := range fields {
		if field == "version" && idx+1 < len(fields) {
			version := strings.TrimSpace(fields[idx+1])
			version = strings.TrimPrefix(version, "v")
			if version == "" {
				break
			}
			return "v" + version, nil
		}
	}

	return "", fmt.Errorf("parse golangci-lint version output: %q", strings.TrimSpace(string(output)))
}

func binaryFromPath(pathEnv string) string {
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}

		candidate := filepath.Join(dir, binaryName())
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Mode()&0o111 == 0 && runtime.GOOS != "windows" {
			continue
		}
		return candidate
	}

	return ""
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "golangci-lint.exe"
	}
	return "golangci-lint"
}
