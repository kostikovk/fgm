package currenttoolchain

import (
	"context"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/fgmconfig"
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
		return selection, nil
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
