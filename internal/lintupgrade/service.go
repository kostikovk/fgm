package lintupgrade

import (
	"context"
	"fmt"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/pinnedlint"
)

// Resolver returns the selected toolchain for a workspace.
type Resolver interface {
	Current(ctx context.Context, workDir string) (app.Selection, error)
}

// RemoteVersionProvider lists compatible golangci-lint versions for a Go version.
type RemoteVersionProvider interface {
	ListRemoteLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

// Installer installs golangci-lint versions into the FGM store.
type Installer interface {
	InstallLintVersion(ctx context.Context, version string) (string, error)
}

// Config configures a Service.
type Config struct {
	Resolver       Resolver
	RemoteProvider RemoteVersionProvider
	Installer      Installer
}

// Service upgrades golangci-lint to a compatible version.
type Service struct {
	resolver       Resolver
	remoteProvider RemoteVersionProvider
	installer      Installer
}

// New constructs a Service.
func New(config Config) *Service {
	return &Service{
		resolver:       config.Resolver,
		remoteProvider: config.RemoteProvider,
		installer:      config.Installer,
	}
}

// Upgrade installs the selected golangci-lint version for the resolved Go toolchain.
func (s *Service) Upgrade(ctx context.Context, options app.LintUpgradeOptions) (app.LintUpgradeResult, error) {
	if s.resolver == nil {
		return app.LintUpgradeResult{}, fmt.Errorf("resolver is not configured")
	}

	selection, err := s.resolver.Current(ctx, options.WorkDir)
	if err != nil {
		return app.LintUpgradeResult{}, err
	}
	if selection.GoVersion == "" {
		return app.LintUpgradeResult{}, fmt.Errorf("no Go version resolved for the current directory")
	}

	version, err := s.targetVersion(ctx, options, selection.GoVersion)
	if err != nil {
		return app.LintUpgradeResult{}, err
	}

	if options.DryRun {
		return app.LintUpgradeResult{
			Version:   version,
			GoVersion: selection.GoVersion,
			DryRun:    true,
		}, nil
	}

	if s.installer == nil {
		return app.LintUpgradeResult{}, fmt.Errorf("golangci-lint installer is not configured")
	}
	if _, err := s.installer.InstallLintVersion(ctx, version); err != nil {
		return app.LintUpgradeResult{}, err
	}

	return app.LintUpgradeResult{
		Version:   version,
		GoVersion: selection.GoVersion,
	}, nil
}

func (s *Service) targetVersion(ctx context.Context, options app.LintUpgradeOptions, goVersion string) (string, error) {
	if options.Version != "" {
		return options.Version, nil
	}

	if pinned, ok, err := pinnedlint.ResolvePinned(options.WorkDir); err != nil {
		return "", err
	} else if ok {
		return pinned, nil
	}

	if s.remoteProvider == nil {
		return "", fmt.Errorf("remote golangci-lint provider is not configured")
	}
	versions, err := s.remoteProvider.ListRemoteLintVersions(ctx, goVersion)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no compatible golangci-lint versions found for Go %s", goVersion)
	}

	return versions[0].Version, nil
}
