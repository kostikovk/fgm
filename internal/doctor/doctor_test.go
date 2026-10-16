package doctor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
)

type stubGoStore struct {
	globalGoVersionFn func(ctx context.Context) (string, bool, error)
	shimDirFn         func() string
	goBinaryPathFn    func(ctx context.Context, version string) (string, error)
}

func (s stubGoStore) GlobalGoVersion(ctx context.Context) (string, bool, error) {
	return s.globalGoVersionFn(ctx)
}

func (s stubGoStore) ShimDir() string {
	return s.shimDirFn()
}

func (s stubGoStore) GoBinaryPath(ctx context.Context, version string) (string, error) {
	if s.goBinaryPathFn == nil {
		return "", nil
	}
	return s.goBinaryPathFn(ctx, version)
}

type stubResolver struct {
	currentFn func(ctx context.Context, workDir string) (app.Selection, error)
}

func (s stubResolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	return s.currentFn(ctx, workDir)
}

type stubLintStore struct {
	lintBinaryPathFn func(ctx context.Context, version string) (string, error)
}

func (s stubLintStore) LintBinaryPath(ctx context.Context, version string) (string, error) {
	if s.lintBinaryPathFn == nil {
		return "", nil
	}
	return s.lintBinaryPathFn(ctx, version)
}

type stubLintRemoteProvider struct {
	listRemoteLintVersionsFn func(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

func (s stubLintRemoteProvider) ListRemoteLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
	return s.listRemoteLintVersionsFn(ctx, goVersion)
}

func TestDiagnose_GlobalGoVersionErrorIsPropagated(t *testing.T) {
	t.Parallel()

	service := New(Config{
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "", false, fmt.Errorf("global version boom")
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
		},
		PathEnv: "/usr/local/bin",
	})

	_, err := service.Diagnose(context.Background(), ".")
	if err == nil {
		t.Fatal("expected error from GlobalGoVersion, got nil")
	}
	if !strings.Contains(err.Error(), "global version boom") {
		t.Fatalf("err = %q, want global version boom", err)
	}
}

func TestDiagnose_ResolvePinnedErrorIsPropagated(t *testing.T) {
	t.Parallel()

	// Create a workDir with a malformed .fgm.toml so ResolvePinned returns an error.
	workDir := t.TempDir()
	// Write an invalid TOML so parsing fails.
	if err := os.WriteFile(
		filepath.Join(workDir, ".fgm.toml"),
		[]byte("this is not valid TOML [[["),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion:   "1.25.7",
					LintVersion: "v2.11.2",
				}, nil
			},
		},
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.25.7", true, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/1.25.7/bin/go", nil
			},
		},
		PathEnv: "/tmp/fgm/shims:/usr/local/bin",
	})

	_, err := service.Diagnose(context.Background(), workDir)
	if err == nil {
		t.Fatal("expected error from ResolvePinned due to malformed .fgm.toml, got nil")
	}
}

func TestDiagnose_LintRemoteProviderErrorIsPropagated(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(workDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion:   "1.25.7",
					LintVersion: "v2.10.1",
					LintSource:  "config",
				}, nil
			},
		},
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.25.7", true, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/1.25.7/bin/go", nil
			},
		},
		LintStore: stubLintStore{
			lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/golangci-lint/v2.10.1/golangci-lint", nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return nil, fmt.Errorf("lint remote boom")
			},
		},
		PathEnv: "/tmp/fgm/shims:/usr/local/bin",
	})

	_, err := service.Diagnose(context.Background(), workDir)
	if err == nil {
		t.Fatal("expected error from LintRemoteProvider, got nil")
	}
	if !strings.Contains(err.Error(), "lint remote boom") {
		t.Fatalf("err = %q, want lint remote boom", err)
	}
}

// findingsToString formats findings as "SEVERITY message" lines for test assertions.
func findingsToString(findings []app.DoctorFinding) string {
	parts := make([]string, len(findings))
	for i, f := range findings {
		parts[i] = f.Severity + " " + f.Message
	}
	return strings.Join(parts, "\n")
}

func TestDiagnose_NoCompatibleLintWhenNoPinnedAndNoLintVersion(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion:   "1.25.7",
					LintVersion: "",
				}, nil
			},
		},
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.25.7", true, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/1.25.7/bin/go", nil
			},
		},
		PathEnv: "/tmp/fgm/shims:/usr/local/bin",
	})

	findings, err := service.Diagnose(context.Background(), workDir)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := findingsToString(findings)
	if !strings.Contains(joined, "WARN no compatible golangci-lint version is known for Go 1.25.7") {
		t.Fatalf("joined = %q, want no-compatible-lint warning", joined)
	}
}

func TestDiagnose_ReportsGlobalVersionAndShimStatus(t *testing.T) {
	t.Parallel()

	service := New(Config{
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.25.7", true, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
		},
		PathEnv: "/tmp/fgm/shims:/usr/local/bin",
	})

	findings, err := service.Diagnose(context.Background(), ".")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := findingsToString(findings)
	if !strings.Contains(joined, "OK global Go version: 1.25.7") {
		t.Fatalf("joined = %q, want global version line", joined)
	}
	if !strings.Contains(joined, "OK shim dir is on PATH: /tmp/fgm/shims") {
		t.Fatalf("joined = %q, want shim path line", joined)
	}
}

func TestDiagnose_SetsFixKindWhenShimDirMissing(t *testing.T) {
	t.Parallel()

	service := New(Config{
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "", false, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
		},
		PathEnv: "/usr/local/bin",
	})

	findings, err := service.Diagnose(context.Background(), ".")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	var found bool
	for _, f := range findings {
		if strings.Contains(f.Message, "shim dir is not on PATH") {
			found = true
			if f.FixKind != "shell_profile" {
				t.Fatalf("FixKind = %q, want %q", f.FixKind, "shell_profile")
			}
		}
	}
	if !found {
		t.Fatal("expected shim-not-on-PATH finding")
	}
}

func TestDiagnose_ReportsSelectedToolchainAvailability(t *testing.T) {
	t.Parallel()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				if workDir != "/repo" {
					t.Fatalf("workDir = %q, want %q", workDir, "/repo")
				}
				return app.Selection{
					GoVersion:   "1.25.7",
					LintVersion: "v2.11.2",
				}, nil
			},
		},
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.25.7", true, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "", context.Canceled
			},
		},
		LintStore: stubLintStore{
			lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "", context.Canceled
			},
		},
		PathEnv: "/tmp/fgm/shims:/usr/local/bin",
	})

	findings, err := service.Diagnose(context.Background(), "/repo")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := findingsToString(findings)
	if !strings.Contains(joined, "WARN selected Go version is not installed: 1.25.7") {
		t.Fatalf("joined = %q, want missing Go line", joined)
	}
	if !strings.Contains(joined, `fgm install go 1.25.7`) {
		t.Fatalf("joined = %q, want missing Go install hint", joined)
	}
	if !strings.Contains(joined, "WARN selected golangci-lint version is not installed: v2.11.2") {
		t.Fatalf("joined = %q, want missing lint line", joined)
	}
	if !strings.Contains(joined, `fgm install golangci-lint v2.11.2`) {
		t.Fatalf("joined = %q, want missing lint install hint", joined)
	}
}

func TestDiagnose_ActionableHintsForGlobalAndShim(t *testing.T) {
	t.Parallel()

	t.Run("should clarify global Go is for outside Go projects and suggest fgm exec when shim dir is not on PATH", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service := New(Config{
			Resolver: stubResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion: "1.25.7",
						GoSource:  "/repo/go.mod",
					}, nil
				},
			},
			GoStore: stubGoStore{
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
					return "", false, nil
				},
				shimDirFn: func() string {
					return "/tmp/fgm/shims"
				},
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/go/1.25.7/bin/go", nil
				},
			},
			PathEnv: "/usr/local/bin",
		})

		// Act
		findings, err := service.Diagnose(context.Background(), "/repo")

		// Assert
		if err != nil {
			t.Fatalf("Diagnose: %v", err)
		}
		joined := findingsToString(findings)
		if !strings.Contains(joined, "outside Go projects") {
			t.Fatalf("joined = %q, want global Go warning to clarify it is needed outside Go projects", joined)
		}
		if !strings.Contains(joined, "fgm exec -- <command>") {
			t.Fatalf("joined = %q, want shim warning to include non-interactive fgm exec hint", joined)
		}
	})
}

func TestDiagnose_ReportsPinnedLintCompatibilityStatus(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(workDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion:   "1.26.1",
					LintVersion: "v2.10.1",
					LintSource:  "config",
				}, nil
			},
		},
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.26.1", true, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/1.26.1/bin/go", nil
			},
		},
		LintStore: stubLintStore{
			lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/golangci-lint/v2.10.1/golangci-lint", nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.10.1"}}, nil
			},
		},
		PathEnv: "/tmp/fgm/shims:/usr/local/bin",
	})

	findings, err := service.Diagnose(context.Background(), workDir)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := findingsToString(findings)
	if !strings.Contains(joined, "OK repo pins golangci-lint: v2.10.1") {
		t.Fatalf("joined = %q, want pinned lint line", joined)
	}
	if !strings.Contains(joined, "OK pinned golangci-lint version is compatible with Go 1.26.1: v2.10.1") {
		t.Fatalf("joined = %q, want compatible lint line", joined)
	}
}

func TestDiagnose_WarnsWhenPinnedLintIsMissingAndIncompatible(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(workDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion:   "1.26.1",
					LintVersion: "v2.10.1",
					LintSource:  "config",
				}, nil
			},
		},
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.26.1", true, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/1.26.1/bin/go", nil
			},
		},
		LintStore: stubLintStore{
			lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "", context.Canceled
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2"}}, nil
			},
		},
		PathEnv: "/tmp/fgm/shims:/usr/local/bin",
	})

	findings, err := service.Diagnose(context.Background(), workDir)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := findingsToString(findings)
	if !strings.Contains(joined, "WARN pinned golangci-lint version is not installed: v2.10.1") {
		t.Fatalf("joined = %q, want missing pinned lint line", joined)
	}
	if !strings.Contains(joined, "WARN pinned golangci-lint version is not in the compatible set for Go 1.26.1: v2.10.1") {
		t.Fatalf("joined = %q, want incompatible pinned lint line", joined)
	}
}

func TestDiagnose_IgnoresResolverErrorAndReturnsBaseDiagnostics(t *testing.T) {
	t.Parallel()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{}, fmt.Errorf("resolver boom")
			},
		},
		GoStore: stubGoStore{
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.25.7", true, nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
		},
		PathEnv: "/usr/local/bin",
	})

	findings, err := service.Diagnose(context.Background(), ".")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := findingsToString(findings)
	if !strings.Contains(joined, "OK global Go version: 1.25.7") {
		t.Fatalf("joined = %q, want global version line", joined)
	}
	if strings.Contains(joined, "resolver boom") {
		t.Fatalf("joined = %q, want resolver errors to be ignored", joined)
	}
}

type stubFGMUpdateChecker struct {
	checkLatestFn  func(ctx context.Context) (app.FGMUpdateStatus, bool, error)
	cachedNoticeFn func(ctx context.Context) (app.FGMUpdateStatus, bool)
}

func (s stubFGMUpdateChecker) CheckLatest(ctx context.Context) (app.FGMUpdateStatus, bool, error) {
	if s.checkLatestFn == nil {
		return app.FGMUpdateStatus{}, false, nil
	}
	return s.checkLatestFn(ctx)
}

func (s stubFGMUpdateChecker) CachedNotice(ctx context.Context) (app.FGMUpdateStatus, bool) {
	if s.cachedNoticeFn == nil {
		return app.FGMUpdateStatus{}, false
	}
	return s.cachedNoticeFn(ctx)
}

func TestDiagnose_ReportsFGMUpdateFinding(t *testing.T) {
	t.Parallel()

	t.Run("should emit WARN finding with upgrade command when a newer FGM version is available", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service := New(Config{
			GoStore: stubGoStore{
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
					return "1.27.2", true, nil
				},
				shimDirFn: func() string { return "/tmp/fgm/shims" },
			},
			FGMUpdateChecker: stubFGMUpdateChecker{
				checkLatestFn: func(ctx context.Context) (app.FGMUpdateStatus, bool, error) {
					return app.FGMUpdateStatus{
						CurrentVersion:  "0.5.0",
						LatestVersion:   "0.6.0",
						UpdateAvailable: true,
						UpgradeCommand:  "brew update && brew upgrade --cask fgm",
					}, true, nil
				},
			},
			PathEnv: "/tmp/fgm/shims",
		})

		// Act
		findings, err := service.Diagnose(context.Background(), ".")

		// Assert
		if err != nil {
			t.Fatalf("Diagnose: %v", err)
		}
		joined := findingsToString(findings)
		want := `WARN fgm update available: 0.5.0 -> 0.6.0 (run "brew update && brew upgrade --cask fgm")`
		if !strings.Contains(joined, want) {
			t.Fatalf("joined = %q, want it to contain %q", joined, want)
		}
	})

	t.Run("should emit OK finding when FGM is already up to date", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service := New(Config{
			GoStore: stubGoStore{
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
					return "1.27.2", true, nil
				},
				shimDirFn: func() string { return "/tmp/fgm/shims" },
			},
			FGMUpdateChecker: stubFGMUpdateChecker{
				checkLatestFn: func(ctx context.Context) (app.FGMUpdateStatus, bool, error) {
					return app.FGMUpdateStatus{
						CurrentVersion:  "0.5.0",
						LatestVersion:   "0.5.0",
						UpdateAvailable: false,
					}, true, nil
				},
			},
			PathEnv: "/tmp/fgm/shims",
		})

		// Act
		findings, err := service.Diagnose(context.Background(), ".")

		// Assert
		if err != nil {
			t.Fatalf("Diagnose: %v", err)
		}
		joined := findingsToString(findings)
		want := "OK fgm is up to date: 0.5.0"
		if !strings.Contains(joined, want) {
			t.Fatalf("joined = %q, want it to contain %q", joined, want)
		}
	})
}

func TestDiagnoseGoToolchainEnvironment(t *testing.T) {
	t.Parallel()

	newService := func(env map[string]string) *Service {
		return New(Config{
			GoStore: stubGoStore{
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) { return "1.26.0", true, nil },
				shimDirFn:         func() string { return "/tmp/fgm/shims" },
			},
			PathEnv: "/tmp/fgm/shims",
			Getenv:  func(key string) string { return env[key] },
		})
	}
	findWarning := func(findings []app.DoctorFinding, needle string) bool {
		for _, finding := range findings {
			if finding.Severity == "WARN" && strings.Contains(finding.Message, needle) {
				return true
			}
		}
		return false
	}

	t.Run("Diagnose", func(t *testing.T) {
		t.Parallel()

		t.Run("should warn when the shell sets GOTOOLCHAIN to a toolchain version", func(t *testing.T) {
			t.Parallel()

			// Arrange
			service := newService(map[string]string{"GOTOOLCHAIN": "go1.26.8"})

			// Act
			findings, err := service.Diagnose(context.Background(), t.TempDir())

			// Assert
			if err != nil {
				t.Fatalf("Diagnose: %v", err)
			}
			if !findWarning(findings, "GOTOOLCHAIN=go1.26.8") {
				t.Errorf("findings = %+v; want a GOTOOLCHAIN warning", findings)
			}
		})

		t.Run("should warn when the shell sets GOROOT", func(t *testing.T) {
			t.Parallel()

			// Arrange
			service := newService(map[string]string{"GOROOT": "/opt/go"})

			// Act
			findings, err := service.Diagnose(context.Background(), t.TempDir())

			// Assert
			if err != nil {
				t.Fatalf("Diagnose: %v", err)
			}
			if !findWarning(findings, "GOROOT=/opt/go") {
				t.Errorf("findings = %+v; want a GOROOT warning", findings)
			}
		})

		t.Run("should not warn about GOTOOLCHAIN when the shell sets it to auto", func(t *testing.T) {
			t.Parallel()

			// Arrange
			service := newService(map[string]string{"GOTOOLCHAIN": "auto"})

			// Act
			findings, err := service.Diagnose(context.Background(), t.TempDir())

			// Assert
			if err != nil {
				t.Fatalf("Diagnose: %v", err)
			}
			if findWarning(findings, "GOTOOLCHAIN") {
				t.Errorf("findings = %+v; want no GOTOOLCHAIN warning", findings)
			}
		})
	})
}

func TestDiagnoseSelectionWarnings(t *testing.T) {
	t.Parallel()

	t.Run("Diagnose", func(t *testing.T) {
		t.Parallel()

		t.Run("should report a selection warning as WARN when the selection may be stale", func(t *testing.T) {
			t.Parallel()

			// Arrange
			service := New(Config{
				Resolver: stubResolver{currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.28.1", Warnings: []string{"latest go 1.28 patch taken from installed versions"}}, nil
				}},
				GoStore: stubGoStore{
					globalGoVersionFn: func(ctx context.Context) (string, bool, error) { return "1.28.1", true, nil },
					shimDirFn:         func() string { return "/tmp/fgm/shims" },
				},
				PathEnv: "/tmp/fgm/shims",
				Getenv:  func(string) string { return "" },
			})

			// Act
			findings, err := service.Diagnose(context.Background(), t.TempDir())

			// Assert
			if err != nil {
				t.Fatalf("Diagnose: %v", err)
			}
			for _, finding := range findings {
				if finding.Severity == "WARN" && strings.Contains(finding.Message, "taken from installed versions") {
					return
				}
			}
			t.Errorf("findings = %+v; want the selection warning", findings)
		})
	})
}

type stubLintCatalog struct {
	compatible bool
	known      bool
}

func (s stubLintCatalog) LintCompatibility(goVersion string, lintVersion string) (bool, bool, error) {
	return s.compatible, s.known, nil
}

func TestDiagnoseDetectedLint(t *testing.T) {
	t.Parallel()

	newService := func(catalog stubLintCatalog) *Service {
		return New(Config{
			Resolver: stubResolver{currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion: "1.27.0", LintVersion: "v2.5.0", LintSource: "detected",
					LintSourceFile: filepath.Join(workDir, "Makefile"), LintSourceLine: 3,
				}, nil
			}},
			GoStore: stubGoStore{
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) { return "1.27.0", true, nil },
				shimDirFn:         func() string { return "/tmp/fgm/shims" },
			},
			LintCatalog: catalog,
			PathEnv:     "/tmp/fgm/shims",
			Getenv:      func(string) string { return "" },
		})
	}
	hasFinding := func(findings []app.DoctorFinding, severity string, needle string) bool {
		for _, finding := range findings {
			if finding.Severity == severity && strings.Contains(finding.Message, needle) {
				return true
			}
		}
		return false
	}

	t.Run("Diagnose", func(t *testing.T) {
		t.Parallel()

		t.Run("should warn with the pin location when the detected linter cannot run on the resolved Go", func(t *testing.T) {
			t.Parallel()

			// Arrange
			service := newService(stubLintCatalog{compatible: false, known: true})

			// Act
			findings, err := service.Diagnose(context.Background(), t.TempDir())

			// Assert
			if err != nil {
				t.Fatalf("Diagnose: %v", err)
			}
			if !hasFinding(findings, "WARN", "detected golangci-lint v2.5.0 (Makefile:3) cannot run on Go 1.27.0") {
				t.Errorf("findings = %+v; want an incompatibility warning", findings)
			}
		})

		t.Run("should report the detected pin as OK when the catalog marks it compatible", func(t *testing.T) {
			t.Parallel()

			// Arrange
			service := newService(stubLintCatalog{compatible: true, known: true})

			// Act
			findings, err := service.Diagnose(context.Background(), t.TempDir())

			// Assert
			if err != nil {
				t.Fatalf("Diagnose: %v", err)
			}
			if !hasFinding(findings, "OK", "repo pins golangci-lint v2.5.0 (detected: Makefile:3)") {
				t.Errorf("findings = %+v; want the detected pin reported", findings)
			}
		})
	})
}

type stubImports struct{ broken []app.ImportedGo }

func (s stubImports) BrokenGoImports(ctx context.Context) ([]app.ImportedGo, error) {
	return s.broken, nil
}

func TestDiagnoseBrokenImports(t *testing.T) {
	t.Parallel()

	t.Run("Diagnose", func(t *testing.T) {
		t.Parallel()

		t.Run("should warn with an install hint when an imported Go points at a removed directory", func(t *testing.T) {
			t.Parallel()

			// Arrange
			service := New(Config{
				GoStore: stubGoStore{
					globalGoVersionFn: func(ctx context.Context) (string, bool, error) { return "1.26.8", true, nil },
					shimDirFn:         func() string { return "/tmp/fgm/shims" },
				},
				Imports: stubImports{broken: []app.ImportedGo{{Version: "1.26.8", Path: "/cache/golang.org/toolchain@v0.0.1-go1.26.8.darwin-arm64"}}},
				PathEnv: "/tmp/fgm/shims",
				Getenv:  func(string) string { return "" },
			})

			// Act
			findings, err := service.Diagnose(context.Background(), t.TempDir())

			// Assert
			if err != nil {
				t.Fatalf("Diagnose: %v", err)
			}
			for _, finding := range findings {
				if finding.Severity == "WARN" && strings.Contains(finding.Message, "imported Go 1.26.8") && strings.Contains(finding.Message, "fgm install go 1.26.8") {
					return
				}
			}
			t.Errorf("findings = %+v; want a dangling import warning", findings)
		})
	})
}
