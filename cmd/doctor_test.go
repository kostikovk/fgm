package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/resolve"
	"github.com/kostikovk/fgm/internal/testutil"
)

type stubDoctor struct {
	diagnoseFn func(ctx context.Context, workDir string) ([]string, error)
}

func (s stubDoctor) Diagnose(ctx context.Context, workDir string) ([]string, error) {
	return s.diagnoseFn(ctx, workDir)
}

func TestDoctorCommand_RejectsNilDoctor(t *testing.T) {
	t.Parallel()

	application := &app.App{Doctor: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "doctor")
	if err == nil {
		t.Fatal("expected an error when Doctor is nil")
	}
	if !strings.Contains(err.Error(), "doctor service is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestDoctorCommand_DiagnoseErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		Doctor: stubDoctor{
			diagnoseFn: func(ctx context.Context, workDir string) ([]string, error) {
				return nil, fmt.Errorf("diagnose boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "doctor")
	if err == nil {
		t.Fatal("expected an error when Diagnose fails")
	}
	if !strings.Contains(err.Error(), "diagnose boom") {
		t.Fatalf("err = %q, want diagnose boom", err)
	}
}

func TestDoctorCommand_PrintsDiagnostics(t *testing.T) {
	t.Parallel()

	application := &app.App{
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
	}

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
