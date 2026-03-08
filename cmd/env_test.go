package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

type stubEnvRenderer struct {
	renderFn func(ctx context.Context, shell string) ([]string, error)
}

func (s stubEnvRenderer) Render(ctx context.Context, shell string) ([]string, error) {
	return s.renderFn(ctx, shell)
}

func TestEnvCommand_PrintsShellSetup(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		EnvRenderer: stubEnvRenderer{
			renderFn: func(ctx context.Context, shell string) ([]string, error) {
				if shell != "zsh" {
					t.Fatalf("shell = %q, want %q", shell, "zsh")
				}
				return []string{`export PATH="/tmp/fgm/shims":$PATH`}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "env", "--shell", "zsh")
	if err != nil {
		t.Fatalf("execute env: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, `export PATH="/tmp/fgm/shims":$PATH`) {
		t.Fatalf("stdout = %q, want export line", stdout)
	}
}
