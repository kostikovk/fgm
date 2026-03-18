package goupgrade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/pinnedlint"
	"github.com/kostikovk/fgm/internal/versionutil"
)

// RemoteVersionProvider lists remotely available Go versions.
type RemoteVersionProvider interface {
	ListRemoteGoVersions(ctx context.Context) ([]string, error)
}

// Installer installs Go versions into the FGM store.
type Installer interface {
	InstallGoVersion(ctx context.Context, version string) (string, error)
}

// LintRemoteVersionProvider lists compatible golangci-lint versions for a Go version.
type LintRemoteVersionProvider interface {
	ListRemoteLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

// LintInstaller installs golangci-lint versions into the FGM store.
type LintInstaller interface {
	InstallLintVersion(ctx context.Context, version string) (string, error)
}

// GlobalStore updates global Go selection state.
type GlobalStore interface {
	SetGlobalGoVersion(ctx context.Context, version string) error
	EnsureShims() error
}

// Config configures a Service.
type Config struct {
	RemoteProvider     RemoteVersionProvider
	Installer          Installer
	LintRemoteProvider LintRemoteVersionProvider
	LintInstaller      LintInstaller
	GlobalStore        GlobalStore
}

// Service upgrades Go at global or project scope.
type Service struct {
	remoteProvider     RemoteVersionProvider
	installer          Installer
	lintRemoteProvider LintRemoteVersionProvider
	lintInstaller      LintInstaller
	globalStore        GlobalStore
}

// New constructs a Service.
func New(config Config) *Service {
	return &Service{
		remoteProvider:     config.RemoteProvider,
		installer:          config.Installer,
		lintRemoteProvider: config.LintRemoteProvider,
		lintInstaller:      config.LintInstaller,
		globalStore:        config.GlobalStore,
	}
}

// UpgradeGlobal installs the selected Go version and selects it globally.
func (s *Service) UpgradeGlobal(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
	version, err := s.targetVersion(ctx, options)
	if err != nil {
		return app.GoUpgradeResult{}, err
	}

	if options.DryRun {
		result := app.GoUpgradeResult{Version: version, Path: "global", DryRun: true}
		if options.WithLint {
			lintVersion, err := s.resolveLintVersion(ctx, version, options.WorkDir)
			if err != nil {
				return app.GoUpgradeResult{}, err
			}
			result.LintVersion = lintVersion
		}
		return result, nil
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

	result := app.GoUpgradeResult{Version: version, Path: "global"}
	if options.WithLint {
		lintVersion, err := s.installLint(ctx, version, options.WorkDir)
		if err != nil {
			return app.GoUpgradeResult{}, err
		}
		result.LintVersion = lintVersion
	}

	return result, nil
}

// UpgradeProject updates the nearest project Go metadata file to the selected version.
func (s *Service) UpgradeProject(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
	version, err := s.targetVersion(ctx, options)
	if err != nil {
		return app.GoUpgradeResult{}, err
	}

	path, err := findProjectMetadataFile(options.WorkDir)
	if err != nil {
		return app.GoUpgradeResult{}, err
	}
	if options.DryRun {
		result := app.GoUpgradeResult{Version: version, Path: path, DryRun: true}
		if options.WithLint {
			lintVersion, err := s.resolveLintVersion(ctx, version, options.WorkDir)
			if err != nil {
				return app.GoUpgradeResult{}, err
			}
			result.LintVersion = lintVersion
		}
		return result, nil
	}

	if _, err := s.installer.InstallGoVersion(ctx, version); err != nil {
		return app.GoUpgradeResult{}, err
	}
	if err := rewriteVersionMetadata(path, version); err != nil {
		return app.GoUpgradeResult{}, err
	}

	result := app.GoUpgradeResult{Version: version, Path: path}
	if options.WithLint {
		lintVersion, err := s.installLint(ctx, version, options.WorkDir)
		if err != nil {
			return app.GoUpgradeResult{}, err
		}
		result.LintVersion = lintVersion
	}

	return result, nil
}

func (s *Service) targetVersion(ctx context.Context, options app.GoUpgradeOptions) (string, error) {
	if options.Version != "" {
		return options.Version, nil
	}
	return s.latestVersion(ctx)
}

func (s *Service) installLint(ctx context.Context, goVersion string, workDir string) (string, error) {
	lintVersion, err := s.resolveLintVersion(ctx, goVersion, workDir)
	if err != nil || lintVersion == "" {
		return lintVersion, err
	}
	if s.lintInstaller == nil {
		return "", fmt.Errorf("golangci-lint installer is not configured")
	}
	if _, err := s.lintInstaller.InstallLintVersion(ctx, lintVersion); err != nil {
		return "", err
	}
	return lintVersion, nil
}

func (s *Service) resolveLintVersion(ctx context.Context, goVersion string, workDir string) (string, error) {
	if pinned, ok, err := pinnedlint.ResolvePinned(workDir); err != nil {
		return "", err
	} else if ok {
		return pinned, nil
	}
	if s.lintRemoteProvider == nil {
		return "", nil
	}
	versions, err := s.lintRemoteProvider.ListRemoteLintVersions(ctx, goVersion)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", nil
	}
	return versions[0].Version, nil
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
		if versionutil.CompareVersions(version, latest) > 0 {
			latest = version
		}
	}

	return latest, nil
}

func findProjectMetadataFile(workDir string) (string, error) {
	if path, err := versionutil.FindNearestFile(workDir, "go.work"); err == nil {
		return path, nil
	} else if !errors.Is(err, versionutil.ErrNotFound) {
		return "", err
	}

	return versionutil.FindNearestFile(workDir, "go.mod")
}

func rewriteVersionMetadata(path string, version string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	originalMode := info.Mode()

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
			return writeLines(path, lines, originalMode)
		}
	}

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if _, ok := strings.CutPrefix(trimmed, "go "); ok {
			lines[idx] = replaceDirectiveLine(line, "go "+version)
			return writeLines(path, lines, originalMode)
		}
	}

	return fmt.Errorf("no go version directive found in %s", path)
}

func replaceDirectiveLine(line string, replacement string) string {
	indentation := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	return indentation + replacement
}

func writeLines(path string, lines []string, mode os.FileMode) error {
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), mode); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
