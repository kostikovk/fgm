package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/kostikovk/fgm/cmd"
	"github.com/kostikovk/fgm/internal/app"
)

type fakeRootCommand struct {
	executeContext func(context.Context) error
}

func (f fakeRootCommand) ExecuteContext(ctx context.Context) error {
	return f.executeContext(ctx)
}

func TestRunBuildsApplicationAndExecutesRootCommand(t *testing.T) {
	t.Parallel()

	var capturedApp *app.App
	var stderr bytes.Buffer

	err := run(
		context.Background(),
		&stderr,
		func(key string) string {
			switch key {
			case "PATH":
				return "/tmp/bin"
			case "SHELL":
				return "/bin/zsh"
			default:
				return ""
			}
		},
		func() (string, error) { return t.TempDir(), nil },
		func(application *app.App) rootCommand {
			capturedApp = application
			return fakeRootCommand{
				executeContext: func(ctx context.Context) error { return nil },
			}
		},
	)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if capturedApp == nil {
		t.Fatal("expected application to be passed to root command builder")
	}
	if capturedApp.GoStore == nil || capturedApp.LintStore == nil {
		t.Fatal("expected application stores to be initialized")
	}
	if capturedApp.BuildInfo.Version != "dev" || capturedApp.BuildInfo.Commit != "unknown" || capturedApp.BuildInfo.Date != "unknown" {
		t.Fatalf("BuildInfo = %+v, want default build metadata", capturedApp.BuildInfo)
	}
}

func TestMainWritesDefaultRootErrorAndExits(t *testing.T) {
	restore := stubMainGlobals(
		func() (string, error) { return "", errors.New("root failed") },
		func(_ *app.App) rootCommand {
			t.Fatal("newRootCmd should not be called when root resolution fails")
			return nil
		},
		func(string) string { return "" },
	)
	defer restore()

	var stderr bytes.Buffer
	stderrWriter = &stderr

	exitCode := runMainAndCaptureExit(t)
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
	if got := stderr.String(); got != "root failed\n" {
		t.Fatalf("stderr = %q, want %q", got, "root failed\n")
	}
}

func TestMainWritesExecuteErrorAndExits(t *testing.T) {
	restore := stubMainGlobals(
		func() (string, error) { return t.TempDir(), nil },
		func(_ *app.App) rootCommand {
			return fakeRootCommand{
				executeContext: func(context.Context) error { return errors.New("execute failed") },
			}
		},
		func(string) string { return "" },
	)
	defer restore()

	var stderr bytes.Buffer
	stderrWriter = &stderr

	exitCode := runMainAndCaptureExit(t)
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
	if got := stderr.String(); got != "execute failed\n" {
		t.Fatalf("stderr = %q, want %q", got, "execute failed\n")
	}
}

func TestMainDoesNotDuplicateAlreadyPrintedError(t *testing.T) {
	t.Run("should print error only once and exit with status 1 when RootRunner already rendered the error", func(t *testing.T) {
		// Arrange
		var stderr bytes.Buffer
		restore := stubMainGlobals(
			func() (string, error) { return t.TempDir(), nil },
			func(application *app.App) rootCommand {
				opts := cmd.UIOptions{
					Getenv:     func(string) string { return "" },
					IsTerminal: func(io.Writer) bool { return false },
				}
				root := cmd.NewRootCmdWithUI(application, opts)
				root.SetOut(&bytes.Buffer{})
				root.SetErr(&stderr)
				root.SetArgs([]string{"use", "go", "1.27.2"})
				return cmd.NewRootRunner(root, opts)
			},
			func(string) string { return "" },
		)
		defer restore()
		stderrWriter = &stderr

		// Act
		exitCode := runMainAndCaptureExit(t)

		// Assert
		if exitCode != 1 {
			t.Fatalf("exitCode = %d, want 1", exitCode)
		}
		if got := stderr.String(); got != "only --global is supported for now\n" {
			t.Fatalf("stderr = %q, want exact single error line", got)
		}
	})
}

func stubMainGlobals(
	resolveRoot func() (string, error),
	buildRoot func(*app.App) rootCommand,
	getenv func(string) string,
) func() {
	oldResolveRoot := defaultGoRoot
	oldBuildRoot := newRootCmd
	oldGetEnv := getEnv
	oldStderr := stderrWriter
	oldExit := exitMain

	defaultGoRoot = resolveRoot
	newRootCmd = buildRoot
	getEnv = getenv
	stderrWriter = nil
	exitMain = func(int) {}

	return func() {
		defaultGoRoot = oldResolveRoot
		newRootCmd = oldBuildRoot
		getEnv = oldGetEnv
		stderrWriter = oldStderr
		exitMain = oldExit
	}
}

func runMainAndCaptureExit(t *testing.T) (exitCode int) {
	t.Helper()

	exitMain = func(code int) {
		exitCode = code
		panic("exit")
	}

	defer func() {
		if recovered := recover(); recovered != "exit" {
			t.Fatalf("recovered = %v, want exit panic", recovered)
		}
	}()

	main()
	t.Fatal("main should have exited")
	return exitCode
}
