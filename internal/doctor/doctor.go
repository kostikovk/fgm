package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/kostikovk/fgm/internal/fgmconfig"
	"github.com/kostikovk/fgm/internal/goenv"
	"github.com/kostikovk/fgm/internal/toolhint"
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

// LintCatalog reports offline golangci-lint/Go compatibility from the embedded catalog.
type LintCatalog interface {
	LintCompatibility(goVersion string, lintVersion string) (compatible bool, known bool, err error)
}

// ImportLister reports imported Go versions whose source directory is gone.
type ImportLister interface {
	BrokenGoImports(ctx context.Context) ([]app.ImportedGo, error)
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
	LintCatalog        LintCatalog
	Imports            ImportLister
	FGMUpdateChecker   app.FGMUpdateChecker
	PathEnv            string
	// Getenv reads the caller's environment; os.Getenv is used when nil.
	Getenv func(string) string
}

// Service reports diagnostics for FGM environment setup.
type Service struct {
	resolver           Resolver
	goStore            GoStore
	lintStore          LintStore
	lintRemoteProvider LintRemoteProvider
	lintCatalog        LintCatalog
	imports            ImportLister
	fgmUpdateChecker   app.FGMUpdateChecker
	pathEnv            string
	getenv             func(string) string
}

// New constructs a doctor Service.
func New(config Config) *Service {
	getenv := config.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	return &Service{
		resolver:           config.Resolver,
		goStore:            config.GoStore,
		lintStore:          config.LintStore,
		lintRemoteProvider: config.LintRemoteProvider,
		lintCatalog:        config.LintCatalog,
		imports:            config.Imports,
		fgmUpdateChecker:   config.FGMUpdateChecker,
		pathEnv:            config.PathEnv,
		getenv:             getenv,
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
		findings = append(findings, app.DoctorFinding{
			Severity: "WARN",
			Message:  "no global Go version is selected (needed outside Go projects; run \"fgm use go <version> --global\")",
		})
	}

	shimDir := s.goStore.ShimDir()
	if pathContainsDir(s.pathEnv, shimDir) {
		findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "shim dir is on PATH: " + shimDir})
	} else {
		findings = append(findings, app.DoctorFinding{
			Severity: "WARN",
			Message:  fmt.Sprintf("shim dir is not on PATH: %s (for non-interactive runs, use %q)", shimDir, toolhint.ExecCommand()),
			FixKind:  "shell_profile",
		})
	}

	if value := s.getenv("GOTOOLCHAIN"); goenv.OverridesSelection(value) {
		findings = append(findings, app.DoctorFinding{
			Severity: "WARN",
			Message:  fmt.Sprintf("shell sets GOTOOLCHAIN=%s; shims and plain go calls may run a toolchain other than the selected one (fgm exec forces GOTOOLCHAIN=local)", value),
		})
	}
	if value := s.getenv("GOROOT"); value != "" {
		findings = append(findings, app.DoctorFinding{
			Severity: "WARN",
			Message:  fmt.Sprintf("shell sets GOROOT=%s; the selected go binary may use another installation's standard library (unset it; fgm exec drops it)", value),
		})
	}

	if s.imports != nil {
		broken, err := s.imports.BrokenGoImports(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range broken {
			findings = append(findings, app.DoctorFinding{
				Severity: "WARN",
				Message: fmt.Sprintf(
					"imported Go %s points to a removed directory %s (run %q, or \"fgm remove go %s\")",
					item.Version, item.Path, toolhint.InstallGoCommand(item.Version), item.Version,
				),
			})
		}
	}

	if _, err := exec.LookPath("fgm"); err == nil {
		findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "fgm is available on PATH"})
	} else {
		findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: "fgm is not available on PATH"})
	}

	if s.fgmUpdateChecker != nil {
		if status, ok, err := s.fgmUpdateChecker.CheckLatest(ctx); err == nil && ok {
			if status.UpdateAvailable {
				findings = append(findings, app.DoctorFinding{
					Severity: "WARN",
					Message:  fmt.Sprintf("fgm update available: %s -> %s (run %q)", status.CurrentVersion, status.LatestVersion, status.UpgradeCommand),
				})
			} else {
				findings = append(findings, app.DoctorFinding{
					Severity: "OK",
					Message:  "fgm is up to date: " + status.CurrentVersion,
				})
			}
		}
	}

	if s.resolver != nil {
		selection, err := s.resolver.Current(ctx, workDir)
		if err == nil {
			for _, warning := range selection.Warnings {
				findings = append(findings, app.DoctorFinding{Severity: "WARN", Message: warning})
			}
			if _, err := s.goStore.GoBinaryPath(ctx, selection.GoVersion); err == nil {
				findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "selected Go version is installed: " + selection.GoVersion})
			} else {
				findings = append(findings, app.DoctorFinding{
					Severity: "WARN",
					Message:  fmt.Sprintf("selected Go version is not installed: %s (run %q)", selection.GoVersion, toolhint.InstallGoCommand(selection.GoVersion)),
				})
			}

			pinnedLintVersion, pinned, err := fgmconfig.ResolvePinnedLint(workDir)
			if err != nil {
				return nil, err
			}
			if pinned {
				findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "repo pins golangci-lint: " + pinnedLintVersion})
			}
			if selection.LintSource == "detected" {
				findings = append(findings, s.detectedLintFindings(workDir, selection)...)
			}

			if selection.LintVersion != "" && s.lintStore != nil {
				if _, err := s.lintStore.LintBinaryPath(ctx, selection.LintVersion); err == nil {
					findings = append(findings, app.DoctorFinding{Severity: "OK", Message: "selected golangci-lint version is installed: " + selection.LintVersion})
				} else {
					findings = append(findings, app.DoctorFinding{
						Severity: "WARN",
						Message:  fmt.Sprintf("selected golangci-lint version is not installed: %s (run %q)", selection.LintVersion, toolhint.InstallLintCommand(selection.LintVersion)),
					})
					if pinned && pinnedLintVersion == selection.LintVersion {
						findings = append(findings, app.DoctorFinding{
							Severity: "WARN",
							Message:  fmt.Sprintf("pinned golangci-lint version is not installed: %s (run %q)", selection.LintVersion, toolhint.InstallLintCommand(selection.LintVersion)),
						})
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

func (s *Service) detectedLintFindings(workDir string, selection app.Selection) []app.DoctorFinding {
	location := fmt.Sprintf("%s:%d", currenttoolchain.DisplayPath(workDir, selection.LintSourceFile), selection.LintSourceLine)
	findings := []app.DoctorFinding{{
		Severity: "OK",
		Message:  fmt.Sprintf("repo pins golangci-lint %s (detected: %s)", selection.LintVersion, location),
	}}
	if s.lintCatalog == nil {
		return findings
	}
	compatible, known, err := s.lintCatalog.LintCompatibility(selection.GoVersion, selection.LintVersion)
	if err != nil || !known || compatible {
		return findings
	}
	return append(findings, app.DoctorFinding{
		Severity: "WARN",
		Message: fmt.Sprintf(
			"detected golangci-lint %s (%s) cannot run on Go %s according to the compatibility catalog",
			selection.LintVersion, location, selection.GoVersion,
		),
	})
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
