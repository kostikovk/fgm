package execenv

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
)

type stubResolver struct {
	currentFn func(ctx context.Context, workDir string) (app.Selection, error)
}

func (s stubResolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	return s.currentFn(ctx, workDir)
}

type stubLocator struct {
	goBinaryPathFn   func(ctx context.Context, version string) (string, error)
	lintBinaryPathFn func(ctx context.Context, version string) (string, error)
}

func (s stubLocator) GoBinaryPath(ctx context.Context, version string) (string, error) {
	return s.goBinaryPathFn(ctx, version)
}

func (s stubLocator) LintBinaryPath(ctx context.Context, version string) (string, error) {
	if s.lintBinaryPathFn == nil {
		return "", nil
	}
	return s.lintBinaryPathFn(ctx, version)
}

func TestExecutorExec_PrependsSelectedGoBinaryDirToPATH(t *testing.T) {
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

	executor := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				if gotWorkDir != workDir {
					t.Fatalf("workDir = %q, want %q", gotWorkDir, workDir)
				}
				return app.Selection{GoVersion: "1.25.7", GoSource: "global"}, nil
			},
		},
		GoLocator: stubLocator{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				if version != "1.25.7" {
					t.Fatalf("version = %q, want %q", version, "1.25.7")
				}
				return goBinary, nil
			},
		},
		PathEnv: "",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := executor.Exec(context.Background(), workDir, []string{"go", "version"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("Exec: %v\nstderr:\n%s", err, stderr.String())
	}

	if !strings.Contains(stdout.String(), "go version go1.25.7") {
		t.Fatalf("stdout = %q, want it to contain fake go output", stdout.String())
	}
}

func TestExecutorExec_PrependsSelectedLintBinaryDirToPATH(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	workDir := t.TempDir()
	lintBinDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(lintBinDir, 0o755); err != nil {
		t.Fatalf("mkdir lint bin dir: %v", err)
	}

	lintBinary := filepath.Join(lintBinDir, "golangci-lint")
	script := "#!/bin/sh\nprintf 'golangci-lint has version v2.11.2\\n'\n"
	if err := os.WriteFile(lintBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake lint binary: %v", err)
	}

	executor := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				if gotWorkDir != workDir {
					t.Fatalf("workDir = %q, want %q", gotWorkDir, workDir)
				}
				return app.Selection{
					GoVersion:   "1.25.7",
					GoSource:    "global",
					LintVersion: "v2.11.2",
					LintSource:  "local",
				}, nil
			},
		},
		GoLocator: stubLocator{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return filepath.Join(workDir, "unused-go"), nil
			},
		},
		LintLocator: stubLocator{
			lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				if version != "v2.11.2" {
					t.Fatalf("version = %q, want %q", version, "v2.11.2")
				}
				return lintBinary, nil
			},
		},
		PathEnv: "",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := executor.Exec(
		context.Background(),
		workDir,
		[]string{"golangci-lint", "version"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if err != nil {
		t.Fatalf("Exec: %v\nstderr:\n%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "golangci-lint has version v2.11.2") {
		t.Fatalf("stdout = %q, want it to contain fake lint output", stdout.String())
	}
}
