package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

type stubDoctor struct {
	diagnoseFn func(ctx context.Context, workDir string) ([]string, error)
}

func (s stubDoctor) Diagnose(ctx context.Context, workDir string) ([]string, error) {
	return s.diagnoseFn(ctx, workDir)
}

func TestDoctorCommand_PrintsDiagnostics(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		Doctor: stubDoctor{
			diagnoseFn: func(ctx context.Context, workDir string) ([]string, error) {
				if workDir != "." {
					t.Fatalf("workDir = %q, want %q", workDir, ".")
				}
				return []string{
					"OK global Go version: 1.25.7",
					"WARN shim dir is not on PATH: /tmp/fgm/shims",
					`Run: eval "$(fgm env)"`,
				}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "doctor")
	if err != nil {
		t.Fatalf("execute doctor: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "OK global Go version: 1.25.7") {
		t.Fatalf("stdout = %q, want it to contain global version line", stdout)
	}
	if !strings.Contains(stdout, "WARN shim dir is not on PATH: /tmp/fgm/shims") {
		t.Fatalf("stdout = %q, want it to contain shim warning", stdout)
	}
	if !strings.Contains(stdout, `Run: eval "$(fgm env)"`) {
		t.Fatalf("stdout = %q, want it to contain env suggestion", stdout)
	}
}
