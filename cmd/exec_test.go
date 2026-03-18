package cmd

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/kostikovk/fgm/internal/execenv"
	"github.com/kostikovk/fgm/internal/resolve"
	"github.com/kostikovk/fgm/internal/testutil"
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

	executor := execenv.New(execenv.Config{
		Resolver:  resolve.New(store),
		GoLocator: store,
		PathEnv:   "",
	})
	application := &app.App{
		Resolver: resolve.New(store),
		GoStore:  store,
		Executor: executor,
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "exec", "--chdir", workDir, "--", "go", "version")
	if err != nil {
		t.Fatalf("execute exec command: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "go version go1.25.7") {
		t.Fatalf("stdout = %q, want it to contain fake go output", stdout)
	}
}

func TestExecCommand_RunsCommandWithSelectedLintOnPATH(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	workDir := t.TempDir()
	goBinDir := filepath.Join(t.TempDir(), "go-bin")
	lintBinDir := filepath.Join(t.TempDir(), "lint-bin")
	if err := os.MkdirAll(goBinDir, 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}
	if err := os.MkdirAll(lintBinDir, 0o755); err != nil {
		t.Fatalf("mkdir lint bin dir: %v", err)
	}

	goBinary := filepath.Join(goBinDir, "go")
	if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go version go1.25.7 darwin/arm64\\n'\n"), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}
	lintBinary := filepath.Join(lintBinDir, "golangci-lint")
	if err := os.WriteFile(lintBinary, []byte("#!/bin/sh\nprintf 'golangci-lint has version v2.11.2\\n'\n"), 0o755); err != nil {
		t.Fatalf("write fake lint binary: %v", err)
	}

	goStore := stubGoStore{
		globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
			return "1.25.7", true, nil
		},
		goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			return goBinary, nil
		},
	}
	lintStore := stubLintStore{
		listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
			return []string{"v2.11.2"}, nil
		},
		lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			return lintBinary, nil
		},
	}
	lintProvider := stubLintRemoteProvider{
		listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
			return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
		},
	}
	locator := execLocatorStub{
		goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			return goBinary, nil
		},
		lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			return lintBinary, nil
		},
	}

	resolver := currenttoolchain.New(currenttoolchain.Config{
		GoResolver:         resolve.New(goStore),
		LintStore:          lintStore,
		LintRemoteProvider: lintProvider,
	})
	executor := execenv.New(execenv.Config{
		Resolver:    resolver,
		GoLocator:   locator,
		LintLocator: locator,
		PathEnv:     "",
	})
	application := &app.App{
		Resolver:           resolver,
		GoStore:            goStore,
		LintStore:          lintStore,
		LintRemoteProvider: lintProvider,
		Executor:           executor,
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "exec", "--chdir", workDir, "--", "golangci-lint", "version")
	if err != nil {
		t.Fatalf("execute exec command: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "golangci-lint has version v2.11.2") {
		t.Fatalf("stdout = %q, want it to contain fake lint output", stdout)
	}
}

func TestExecCommand_PrefersPinnedLintVersionFromRepoConfig(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	workDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(workDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	goBinDir := filepath.Join(t.TempDir(), "go-bin")
	lintBinDir := filepath.Join(t.TempDir(), "lint-bin")
	if err := os.MkdirAll(goBinDir, 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}
	if err := os.MkdirAll(lintBinDir, 0o755); err != nil {
		t.Fatalf("mkdir lint bin dir: %v", err)
	}

	goBinary := filepath.Join(goBinDir, "go")
	if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go version go1.25.0 darwin/arm64\\n'\n"), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}
	lintBinary := filepath.Join(lintBinDir, "golangci-lint")
	if err := os.WriteFile(lintBinary, []byte("#!/bin/sh\nprintf 'golangci-lint has version v2.10.1\\n'\n"), 0o755); err != nil {
		t.Fatalf("write fake lint binary: %v", err)
	}

	locator := execLocatorStub{
		goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			return goBinary, nil
		},
		lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			if version != "v2.10.1" {
				t.Fatalf("lint version = %q, want %q", version, "v2.10.1")
			}
			return lintBinary, nil
		},
	}

	resolver := currenttoolchain.New(currenttoolchain.Config{
		GoResolver: resolve.New(nil),
	})
	executor := execenv.New(execenv.Config{
		Resolver:    resolver,
		GoLocator:   locator,
		LintLocator: locator,
		PathEnv:     "",
	})
	application := &app.App{
		Resolver: resolver,
		Executor: executor,
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "exec", "--chdir", workDir, "--", "golangci-lint", "version")
	if err != nil {
		t.Fatalf("execute exec command: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "golangci-lint has version v2.10.1") {
		t.Fatalf("stdout = %q, want it to contain pinned lint output", stdout)
	}
}

func TestExecCommand_RejectsNoCommand(t *testing.T) {
	t.Parallel()

	application := &app.App{}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "exec", "--")
	if err == nil {
		t.Fatal("expected an error when no command is provided")
	}
	if !strings.Contains(err.Error(), "no command provided") {
		t.Fatalf("err = %q, want no command error", err)
	}
}

func TestExecCommand_RejectsNilExecutor(t *testing.T) {
	t.Parallel()

	application := &app.App{Executor: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "exec", "--", "go", "version")
	if err == nil {
		t.Fatal("expected an error when Executor is nil")
	}
	if !strings.Contains(err.Error(), "executor is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

type execLocatorStub struct {
	goBinaryPathFn   func(ctx context.Context, version string) (string, error)
	lintBinaryPathFn func(ctx context.Context, version string) (string, error)
}

func (s execLocatorStub) GoBinaryPath(ctx context.Context, version string) (string, error) {
	return s.goBinaryPathFn(ctx, version)
}

func (s execLocatorStub) LintBinaryPath(ctx context.Context, version string) (string, error) {
	return s.lintBinaryPathFn(ctx, version)
}
