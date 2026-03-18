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

type stubEnvRenderer struct {
	renderFn func(ctx context.Context, shell string) ([]string, error)
}

func (s stubEnvRenderer) Render(ctx context.Context, shell string) ([]string, error) {
	return s.renderFn(ctx, shell)
}

func TestEnvCommand_RejectsNilEnvRenderer(t *testing.T) {
	t.Parallel()

	application := &app.App{EnvRenderer: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "env")
	if err == nil {
		t.Fatal("expected an error when EnvRenderer is nil")
	}
	if !strings.Contains(err.Error(), "env renderer is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestEnvCommand_RenderErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		EnvRenderer: stubEnvRenderer{
			renderFn: func(ctx context.Context, shell string) ([]string, error) {
				return nil, fmt.Errorf("render boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "env")
	if err == nil {
		t.Fatal("expected an error when Render fails")
	}
	if !strings.Contains(err.Error(), "render boom") {
		t.Fatalf("err = %q, want render boom", err)
	}
}

func TestEnvCommand_PrintsShellSetup(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		EnvRenderer: stubEnvRenderer{
			renderFn: func(ctx context.Context, shell string) ([]string, error) {
				if shell != "zsh" {
					t.Fatalf("shell = %q, want %q", shell, "zsh")
				}
				return []string{`export PATH="/tmp/fgm/shims":$PATH`}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "env", "--shell", "zsh")
	if err != nil {
		t.Fatalf("execute env: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, `export PATH="/tmp/fgm/shims":$PATH`) {
		t.Fatalf("stdout = %q, want export line", stdout)
	}
}
