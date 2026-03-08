package cmd

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/execenv"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

func TestExecCommand_RunsCommandWithSelectedGoOnPATH(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	workDir := t.TempDir()
	goBinDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(goBinDir, 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}

	goBinary := filepath.Join(goBinDir, "go")
	script := "#!/bin/sh\nprintf 'go version go1.25.7 darwin/arm64\\n'\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	store := stubGoStore{
		globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
			return "1.25.7", true, nil
		},
		goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			return goBinary, nil
		},
	}

	executor := execenv.New(resolve.New(store), store, "")
	application := app.New(app.Config{
		Resolver: resolve.New(store),
		GoStore:  store,
		Executor: executor,
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "exec", "--chdir", workDir, "--", "go", "version")
	if err != nil {
		t.Fatalf("execute exec command: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "go version go1.25.7") {
		t.Fatalf("stdout = %q, want it to contain fake go output", stdout)
	}
}
