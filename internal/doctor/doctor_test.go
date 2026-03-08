package doctor

import (
	"context"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
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
