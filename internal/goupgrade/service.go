package goupgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/koskosovu4/fgm/internal/app"
)

// RemoteVersionProvider lists remotely available Go versions.
type RemoteVersionProvider interface {
	ListRemoteGoVersions(ctx context.Context) ([]string, error)
}

// Installer installs Go versions into the FGM store.
type Installer interface {
	InstallGoVersion(ctx context.Context, version string) (string, error)
}

// GlobalStore updates global Go selection state.
type GlobalStore interface {
	SetGlobalGoVersion(ctx context.Context, version string) error
	EnsureShims() error
}

// Config configures a Service.
type Config struct {
	RemoteProvider RemoteVersionProvider
	Installer      Installer
	GlobalStore    GlobalStore
}

// Service upgrades Go at global or project scope.
type Service struct {
	remoteProvider RemoteVersionProvider
	installer      Installer
	globalStore    GlobalStore
}

// New constructs a Service.
func New(config Config) *Service {
	return &Service{
		remoteProvider: config.RemoteProvider,
		installer:      config.Installer,
		globalStore:    config.GlobalStore,
	}
}

// UpgradeGlobal installs the latest remote Go version and selects it globally.
func (s *Service) UpgradeGlobal(ctx context.Context) (app.GoUpgradeResult, error) {
	version, err := s.latestVersion(ctx)
	if err != nil {
		return app.GoUpgradeResult{}, err
	}

	if _, err := s.installer.InstallGoVersion(ctx, version); err != nil {
		return app.GoUpgradeResult{}, err
	}
	if err := s.globalStore.SetGlobalGoVersion(ctx, version); err != nil {
		return app.GoUpgradeResult{}, err
	}
	if err := s.globalStore.EnsureShims(); err != nil {
		return app.GoUpgradeResult{}, err
	}

	return app.GoUpgradeResult{Version: version, Path: "global"}, nil
}

// UpgradeProject updates the nearest project Go metadata file to the latest remote version.
func (s *Service) UpgradeProject(ctx context.Context, workDir string) (app.GoUpgradeResult, error) {
	version, err := s.latestVersion(ctx)
	if err != nil {
		return app.GoUpgradeResult{}, err
	}

	if _, err := s.installer.InstallGoVersion(ctx, version); err != nil {
		return app.GoUpgradeResult{}, err
	}

	path, err := findProjectMetadataFile(workDir)
	if err != nil {
		return app.GoUpgradeResult{}, err
	}
	if err := rewriteVersionMetadata(path, version); err != nil {
		return app.GoUpgradeResult{}, err
	}

	return app.GoUpgradeResult{Version: version, Path: path}, nil
}

func (s *Service) latestVersion(ctx context.Context) (string, error) {
	versions, err := s.remoteProvider.ListRemoteGoVersions(ctx)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no remote Go versions are available")
	}

	latest := versions[0]
	for _, version := range versions[1:] {
		if compareVersions(version, latest) > 0 {
			latest = version
		}
	}

	return latest, nil
}

func findProjectMetadataFile(workDir string) (string, error) {
	if path, err := findNearestFile(workDir, "go.work"); err == nil {
		return path, nil
	}

	return findNearestFile(workDir, "go.mod")
}

func findNearestFile(dir string, name string) (string, error) {
	current := dir
	for {
		candidate := filepath.Join(current, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("%s not found from %s upward", name, dir)
}

func rewriteVersionMetadata(path string, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "toolchain" {
			return fmt.Errorf("toolchain directive is empty in %s", path)
		}
		if _, ok := strings.CutPrefix(trimmed, "toolchain "); ok {
			lines[idx] = replaceDirectiveLine(line, "toolchain go"+version)
			return writeLines(path, lines)
		}
	}

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if _, ok := strings.CutPrefix(trimmed, "go "); ok {
			lines[idx] = replaceDirectiveLine(line, "go "+version)
			return writeLines(path, lines)
		}
	}

	return fmt.Errorf("no go version directive found in %s", path)
}

func replaceDirectiveLine(line string, replacement string) string {
	indentation := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	return indentation + replacement
}

func writeLines(path string, lines []string) error {
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func compareVersions(left string, right string) int {
	leftParts := parseVersionParts(left)
	rightParts := parseVersionParts(right)

	for idx := 0; idx < len(leftParts) && idx < len(rightParts); idx++ {
		if leftParts[idx] > rightParts[idx] {
			return 1
		}
		if leftParts[idx] < rightParts[idx] {
			return -1
		}
	}

	if len(leftParts) > len(rightParts) {
		return 1
	}
	if len(leftParts) < len(rightParts) {
		return -1
	}
	return 0
}

func parseVersionParts(version string) []int {
	fields := strings.Split(version, ".")
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.Atoi(field)
		if err != nil {
			return parts
		}
		parts = append(parts, value)
	}
	return parts
}
