package doctor

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/pinnedlint"
)

// GoStore provides the local state doctor needs to inspect.
type GoStore interface {
	GlobalGoVersion(ctx context.Context) (string, bool, error)
	GoBinaryPath(ctx context.Context, version string) (string, error)
	ShimDir() string
}

// LintStore provides local lint binary lookup.
type LintStore interface {
	LintBinaryPath(ctx context.Context, version string) (string, error)
}

// LintRemoteProvider lists compatible remote golangci-lint versions for a Go version.
type LintRemoteProvider interface {
	ListRemoteLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

// Resolver provides the current toolchain selection for a workspace.
type Resolver interface {
	Current(ctx context.Context, workDir string) (app.Selection, error)
}

// Config configures a doctor Service.
type Config struct {
	Resolver           Resolver
	GoStore            GoStore
	LintStore          LintStore
	LintRemoteProvider LintRemoteProvider
	PathEnv            string
}

// Service reports diagnostics for FGM environment setup.
type Service struct {
	resolver           Resolver
	goStore            GoStore
	lintStore          LintStore
	lintRemoteProvider LintRemoteProvider
	pathEnv            string
}

// New constructs a doctor Service.
func New(config Config) *Service {
	return &Service{
		resolver:           config.Resolver,
		goStore:            config.GoStore,
		lintStore:          config.LintStore,
		lintRemoteProvider: config.LintRemoteProvider,
		pathEnv:            config.PathEnv,
	}
}

// Diagnose returns human-readable diagnostics for the current FGM setup.
func (s *Service) Diagnose(ctx context.Context, workDir string) ([]string, error) {
	lines := make([]string, 0, 8)

	version, ok, err := s.goStore.GlobalGoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		lines = append(lines, "OK global Go version: "+version)
	} else {
		lines = append(lines, "WARN no global Go version is selected")
	}

	shimDir := s.goStore.ShimDir()
	if pathContainsDir(s.pathEnv, shimDir) {
		lines = append(lines, "OK shim dir is on PATH: "+shimDir)
	} else {
		lines = append(lines, "WARN shim dir is not on PATH: "+shimDir)
		lines = append(lines, `Run: eval "$(fgm env)"`)
	}

	if _, err := exec.LookPath("fgm"); err == nil {
		lines = append(lines, "OK fgm is available on PATH")
	} else {
		lines = append(lines, "WARN fgm is not available on PATH")
	}

	if s.resolver != nil {
		selection, err := s.resolver.Current(ctx, workDir)
		if err == nil {
			if _, err := s.goStore.GoBinaryPath(ctx, selection.GoVersion); err == nil {
				lines = append(lines, "OK selected Go version is installed: "+selection.GoVersion)
			} else {
				lines = append(lines, "WARN selected Go version is not installed: "+selection.GoVersion)
			}

			pinnedLintVersion, pinned, err := pinnedlint.ResolvePinned(workDir)
			if err != nil {
				return nil, err
			}
			if pinned {
				lines = append(lines, "OK repo pins golangci-lint: "+pinnedLintVersion)
			}

			if selection.LintVersion != "" && s.lintStore != nil {
				if _, err := s.lintStore.LintBinaryPath(ctx, selection.LintVersion); err == nil {
					lines = append(lines, "OK selected golangci-lint version is installed: "+selection.LintVersion)
				} else {
					lines = append(lines, "WARN selected golangci-lint version is not installed: "+selection.LintVersion)
					if pinned && pinnedLintVersion == selection.LintVersion {
						lines = append(lines, "WARN pinned golangci-lint version is not installed: "+selection.LintVersion)
					}
				}
			}

			if pinned && s.lintRemoteProvider != nil {
				compatible, err := s.lintRemoteProvider.ListRemoteLintVersions(ctx, selection.GoVersion)
				if err != nil {
					return nil, err
				}
				if lintVersionCompatible(pinnedLintVersion, compatible) {
					lines = append(lines, "OK pinned golangci-lint version is compatible with Go "+selection.GoVersion+": "+pinnedLintVersion)
				} else {
					lines = append(lines, "WARN pinned golangci-lint version is not in the compatible set for Go "+selection.GoVersion+": "+pinnedLintVersion)
				}
			}

			if !pinned && selection.LintVersion == "" {
				lines = append(lines, "WARN no compatible golangci-lint version is known for Go "+selection.GoVersion)
			}
		}
	}

	return lines, nil
}

func pathContainsDir(pathEnv string, dir string) bool {
	for _, entry := range filepath.SplitList(pathEnv) {
		if strings.TrimSpace(entry) == dir {
			return true
		}
	}
	return false
}

func lintVersionCompatible(version string, compatible []app.LintVersion) bool {
	for _, candidate := range compatible {
		if candidate.Version == version {
			return true
		}
	}
	return false
}
