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

	lines, err := service.Diagnose(context.Background(), workDir)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
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

	lines, err := service.Diagnose(context.Background(), ".")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "OK global Go version: 1.25.7") {
		t.Fatalf("joined = %q, want global version line", joined)
	}
	if !strings.Contains(joined, "OK shim dir is on PATH: /tmp/fgm/shims") {
		t.Fatalf("joined = %q, want shim path line", joined)
	}
}

func TestDiagnose_SuggestsEnvCommandWhenShimDirMissing(t *testing.T) {
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

	lines, err := service.Diagnose(context.Background(), ".")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, `Run: eval "$(fgm env)"`) {
		t.Fatalf("joined = %q, want env suggestion", joined)
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

	lines, err := service.Diagnose(context.Background(), "/repo")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "WARN selected Go version is not installed: 1.25.7") {
		t.Fatalf("joined = %q, want missing Go line", joined)
	}
	if !strings.Contains(joined, "WARN selected golangci-lint version is not installed: v2.11.2") {
		t.Fatalf("joined = %q, want missing lint line", joined)
	}
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

	lines, err := service.Diagnose(context.Background(), workDir)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
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

	lines, err := service.Diagnose(context.Background(), workDir)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
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

	lines, err := service.Diagnose(context.Background(), ".")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "OK global Go version: 1.25.7") {
		t.Fatalf("joined = %q, want global version line", joined)
	}
	if strings.Contains(joined, "resolver boom") {
		t.Fatalf("joined = %q, want resolver errors to be ignored", joined)
	}
}
