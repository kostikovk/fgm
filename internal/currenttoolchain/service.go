package currenttoolchain

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/fgmconfig"
	"github.com/kostikovk/fgm/internal/lintdetect"
)

// GoResolver resolves the current Go toolchain selection.
type GoResolver interface {
	Current(ctx context.Context, workDir string) (app.Selection, error)
}

// LintCatalogVersionProvider lists compatible golangci-lint versions from the embedded catalog.
type LintCatalogVersionProvider interface {
	ListCatalogLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

// Config configures a composite current toolchain resolver.
type Config struct {
	GoResolver          GoResolver
	LintStore           app.LintLocalVersionStore
	LintCatalogProvider LintCatalogVersionProvider
	LintRemoteProvider  app.LintRemoteVersionProvider
}

// Service resolves the current Go selection and enriches it with golangci-lint.
type Service struct {
	goResolver          GoResolver
	lintStore           app.LintLocalVersionStore
	lintCatalogProvider LintCatalogVersionProvider
	lintRemoteProvider  app.LintRemoteVersionProvider
}

// New constructs a composite current toolchain resolver.
func New(config Config) *Service {
	return &Service{
		goResolver:          config.GoResolver,
		lintStore:           config.LintStore,
		lintCatalogProvider: config.LintCatalogProvider,
		lintRemoteProvider:  config.LintRemoteProvider,
	}
}

// Current resolves the current Go version and the best known golangci-lint version.
//
// Lint precedence: .fgm.toml explicit version → a pin detected in other repository files
// (when all detected pins agree) → the newest catalog version compatible with the resolved Go.
func (s *Service) Current(ctx context.Context, workDir string) (app.Selection, error) {
	selection, err := s.goResolver.Current(ctx, workDir)
	if err != nil {
		return app.Selection{}, err
	}

	if lintVersion, ok, err := fgmconfig.ResolvePinnedLint(workDir); err != nil {
		return app.Selection{}, err
	} else if ok {
		selection.LintVersion = lintVersion
		selection.LintSource = "config"
		if path, found, err := fgmconfig.PinnedLintPath(workDir); err == nil && found {
			selection.LintSourceFile = path
		}
		return selection, nil
	}

	detectEnabled, err := fgmconfig.LintDetectionEnabled(workDir)
	if err != nil {
		return app.Selection{}, err
	}
	if detectEnabled {
		detected, err := lintdetect.Detect(workDir)
		if err != nil {
			return app.Selection{}, err
		}
		if pin, ok := detected.Agreed(); ok {
			selection.LintVersion = pin.Version
			selection.LintSource = "detected"
			selection.LintSourceFile = pin.File
			selection.LintSourceLine = pin.Line
			return selection, nil
		}
		if detected.Conflict() {
			selection.Warnings = append(selection.Warnings, ConflictWarning(workDir, detected.Pins))
		}
	}

	var compatibleVersions []app.LintVersion
	switch {
	case s.lintCatalogProvider != nil:
		compatibleVersions, err = s.lintCatalogProvider.ListCatalogLintVersions(ctx, selection.GoVersion)
		if err != nil {
			return app.Selection{}, err
		}
	case s.lintRemoteProvider != nil:
		compatibleVersions, err = s.lintRemoteProvider.ListRemoteLintVersions(ctx, selection.GoVersion)
		if err != nil {
			return app.Selection{}, err
		}
	default:
		return selection, nil
	}

	if len(compatibleVersions) == 0 {
		return selection, nil
	}

	if s.lintStore != nil {
		localVersions, err := s.lintStore.ListLocalLintVersions(ctx)
		if err != nil {
			return app.Selection{}, err
		}
		if version, ok := pickInstalledLintVersion(localVersions, compatibleVersions); ok {
			selection.LintVersion = version
			selection.LintSource = "local"
			return selection, nil
		}
	}

	selection.LintVersion = compatibleVersions[0].Version
	selection.LintSource = "remote"
	return selection, nil
}

func pickInstalledLintVersion(localVersions []string, remoteVersions []app.LintVersion) (string, bool) {
	known := make(map[string]struct{}, len(remoteVersions))
	for _, version := range remoteVersions {
		known[version.Version] = struct{}{}
	}

	for _, version := range localVersions {
		if _, ok := known[version]; ok {
			return version, true
		}
	}

	return "", false
}

// ConflictWarning describes disagreeing lint pins with paths relative to workDir.
func ConflictWarning(workDir string, pins []lintdetect.Pin) string {
	candidates := make([]string, 0, len(pins))
	for _, pin := range pins {
		candidates = append(candidates, fmt.Sprintf("%s (%s:%d)", pin.Version, DisplayPath(workDir, pin.File), pin.Line))
	}
	return "conflicting golangci-lint pins ignored: " + strings.Join(candidates, ", ")
}

// DisplayPath renders path relative to workDir when that is shorter to read.
func DisplayPath(workDir string, path string) string {
	rel, err := filepath.Rel(workDir, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
