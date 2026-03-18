package goimport

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

// Registry registers existing Go installations in the FGM-managed store.
type Registry interface {
	RegisterExistingGoInstallation(version string, goroot string) (string, error)
}

// Importer discovers existing Go installations and registers them into FGM.
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

// ImportAuto discovers Go installations from known candidate paths and registers them.
func (i *Importer) ImportAuto(ctx context.Context) ([]app.ImportedGo, error) {
	seenVersions := make(map[string]struct{})
	var imported []app.ImportedGo

	for _, candidate := range i.candidates {
		goBinary := filepath.Join(candidate, "bin", goBinaryName())
		if _, err := os.Stat(goBinary); err != nil {
			continue
		}

		version, err := detectVersion(ctx, goBinary)
		if err != nil {
			return nil, err
		}
		if _, seen := seenVersions[version]; seen {
			continue
		}

		if _, err := i.registry.RegisterExistingGoInstallation(version, candidate); err != nil {
			return nil, err
		}

		seenVersions[version] = struct{}{}
		imported = append(imported, app.ImportedGo{
			Version: version,
			Path:    candidate,
		})
	}

	return imported, nil
}

// DefaultCandidates returns common locations that may already contain Go installations.
func DefaultCandidates(pathEnv string) []string {
	candidates := make(map[string]struct{})

	if goroot := gorootFromPathGo(pathEnv); goroot != "" {
		candidates[goroot] = struct{}{}
	}

	for _, path := range []string{
		"/usr/local/go",
		"/opt/homebrew/opt/go/libexec",
		"/usr/local/opt/go/libexec",
	} {
		candidates[path] = struct{}{}
	}

	for _, pattern := range []string{
		"/opt/homebrew/Cellar/go/*/libexec",
		"/usr/local/Cellar/go/*/libexec",
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

func detectVersion(ctx context.Context, goBinary string) (string, error) {
	cmd := exec.CommandContext(ctx, goBinary, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run %s version: %w", goBinary, err)
	}

	fields := strings.Fields(string(output))
	if len(fields) < 3 {
		return "", fmt.Errorf("parse go version output: %q", strings.TrimSpace(string(output)))
	}

	version := strings.TrimPrefix(fields[2], "go")
	if version == "" {
		return "", fmt.Errorf("parse go version output: %q", strings.TrimSpace(string(output)))
	}

	return version, nil
}

func gorootFromPathGo(pathEnv string) string {
	binaryName := goBinaryName()
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, binaryName)
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		return filepath.Dir(filepath.Dir(candidate))
	}

	return ""
}

func goBinaryName() string {
	if runtime.GOOS == "windows" {
		return "go.exe"
	}
	return "go"
}
