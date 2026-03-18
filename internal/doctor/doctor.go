package doctor

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/fgmconfig"
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

// Diagnose returns structured diagnostics for the current FGM setup.
func (s *Service) Diagnose(ctx context.Context, workDir string) ([]app.DoctorFinding, error) {
	findings := make([]app.DoctorFinding, 0, 8)

	version, ok, err := s.goStore.GlobalGoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "global Go version: " + version})
	} else {
		findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: "no global Go version is selected"})
	}

	shimDir := s.goStore.ShimDir()
	if pathContainsDir(s.pathEnv, shimDir) {
		findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "shim dir is on PATH: " + shimDir})
	} else {
		findings = append(findings, app.DoctorFinding{
			Severity: "WARN",
			Message:  "shim dir is not on PATH: " + shimDir,
			FixKind:  "shell_profile",
		})
	}

	if _, err := exec.LookPath("fgm"); err == nil {
		findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "fgm is available on PATH"})
	} else {
		findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: "fgm is not available on PATH"})
	}

	if s.resolver != nil {
		selection, err := s.resolver.Current(ctx, workDir)
		if err == nil {
			if _, err := s.goStore.GoBinaryPath(ctx, selection.GoVersion); err == nil {
				findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "selected Go version is installed: " + selection.GoVersion})
			} else {
				findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: "selected Go version is not installed: " + selection.GoVersion})
			}

			pinnedLintVersion, pinned, err := fgmconfig.ResolvePinnedLint(workDir)
			if err != nil {
				return nil, err
			}
			if pinned {
				findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "repo pins golangci-lint: " + pinnedLintVersion})
			}

			if selection.LintVersion != "" && s.lintStore != nil {
				if _, err := s.lintStore.LintBinaryPath(ctx, selection.LintVersion); err == nil {
					findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "selected golangci-lint version is installed: " + selection.LintVersion})
				} else {
					findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: "selected golangci-lint version is not installed: " + selection.LintVersion})
					if pinned && pinnedLintVersion == selection.LintVersion {
						findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: "pinned golangci-lint version is not installed: " + selection.LintVersion})
					}
				}
			}

			if pinned && s.lintRemoteProvider != nil {
				compatible, err := s.lintRemoteProvider.ListRemoteLintVersions(ctx, selection.GoVersion)
				if err != nil {
					return nil, err
				}
				if lintVersionCompatible(pinnedLintVersion, compatible) {
					findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "pinned golangci-lint version is compatible with Go " + selection.GoVersion + ": " + pinnedLintVersion})
				} else {
					findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: "pinned golangci-lint version is not in the compatible set for Go " + selection.GoVersion + ": " + pinnedLintVersion})
				}
			}

			if !pinned && selection.LintVersion == "" {
				findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: "no compatible golangci-lint version is known for Go " + selection.GoVersion})
			}
		}
	}

	return findings, nil
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
