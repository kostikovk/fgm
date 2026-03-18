package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/resolve"
	"github.com/kostikovk/fgm/internal/testutil"
)

type stubDoctor struct {
	diagnoseFn func(ctx context.Context, workDir string) ([]app.DoctorFinding, error)
}

func (s stubDoctor) Diagnose(ctx context.Context, workDir string) ([]app.DoctorFinding, error) {
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
			diagnoseFn: func(ctx context.Context, workDir string) ([]app.DoctorFinding, error) {
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
			diagnoseFn: func(ctx context.Context, workDir string) ([]app.DoctorFinding, error) {
				if workDir != "." {
					t.Fatalf("workDir = %q, want %q", workDir, ".")
				}
				return []app.DoctorFinding{
					{Severity: "OK", Message: "global Go version: 1.25.7"},
					{Severity: "WARN", Message: "shim dir is not on PATH: /tmp/fgm/shims", FixKind: "shell_profile"},
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
	if !strings.Contains(stdout, "--fix") {
		t.Fatalf("stdout = %q, want it to contain --fix hint", stdout)
	}
}

func TestDoctorCommand_FixWithProfileInstaller(t *testing.T) {
	t.Parallel()

	var installCalled bool

	application := &app.App{
		Resolver: resolve.New(nil),
		Doctor: stubDoctor{
			diagnoseFn: func(ctx context.Context, workDir string) ([]app.DoctorFinding, error) {
				return []app.DoctorFinding{
					{Severity: "WARN", Message: "shim dir is not on PATH: /tmp/fgm/shims", FixKind: "shell_profile"},
				}, nil
			},
		},
		ProfileInstaller: &stubProfileInstaller{
			installProfileFn: func(shell string) (string, bool, error) {
				installCalled = true
				return "/home/user/.zshrc", true, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	// Use a bytes.Buffer as stdin so isTerminal returns false (auto-apply without prompt).
	root.SetIn(bytes.NewBufferString(""))
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"doctor", "--fix"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("execute doctor --fix: %v", err)
	}

	if !installCalled {
		t.Fatal("expected InstallProfile to be called")
	}
	if !strings.Contains(stdout.String(), "Added shell integration to /home/user/.zshrc") {
		t.Fatalf("stdout = %q, want added confirmation", stdout.String())
	}
}

type stubProfileInstaller struct {
	installProfileFn func(shell string) (string, bool, error)
}

func (s *stubProfileInstaller) InstallProfile(shell string) (string, bool, error) {
	return s.installProfileFn(shell)
}
