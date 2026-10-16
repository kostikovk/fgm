package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/testutil"
)

func TestShimCommand_RejectsNotConfigured(t *testing.T) {
	t.Parallel()

	application := &app.App{GoStore: nil, Resolver: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "__shim", "go")
	if err == nil {
		t.Fatal("expected an error when shim dependencies are not configured")
	}
	if !strings.Contains(err.Error(), "shim dependencies are not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestShimCommand_RejectsPartiallyConfiguredNilResolver(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoStore:  stubGoStore{},
		Resolver: nil,
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "__shim", "go")
	if err == nil {
		t.Fatal("expected an error when Resolver is nil")
	}
	if !strings.Contains(err.Error(), "shim dependencies are not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestShimCommand_RejectsUnsupportedTool(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoStore: stubGoStore{},
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7"}, nil
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "__shim", "rustc")
	if err == nil {
		t.Fatal("expected an error for unsupported shim tool")
	}
	if !strings.Contains(err.Error(), `unsupported shim tool "rustc"`) {
		t.Fatalf("err = %q, want unsupported tool error", err)
	}
}

func TestShimCommand_GoShimExecutesBinaryAndCapturesOutput(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	goBinDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(goBinDir, 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}
	goBinary := filepath.Join(goBinDir, "go")
	script := "#!/bin/sh\nprintf 'go version go1.25.7 darwin/arm64\\n'\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	application := &app.App{
		GoStore: stubGoStore{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return goBinary, nil
			},
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
				return "1.25.7", true, nil
			},
		},
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7"}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, _, err := testutil.ExecuteCommand(t, root, "__shim", "go", "version")
	if err != nil {
		t.Fatalf("shim go: %v", err)
	}
	if !strings.Contains(stdout, "go version go1.25.7") {
		t.Fatalf("stdout = %q, want it to contain go version output", stdout)
	}
}

func TestShimCommand_GoShimReturnsResolveError(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoStore: stubGoStore{},
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{}, fmt.Errorf("resolver boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "__shim", "go", "version")
	if err == nil {
		t.Fatal("expected an error when resolver fails")
	}
	if !strings.Contains(err.Error(), "resolver boom") {
		t.Fatalf("err = %q, want resolver boom", err)
	}
}

func TestShimCommand_GoShimReturnsErrorWhenBinaryNotFound(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoStore: stubGoStore{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/nonexistent/binary/go", nil
			},
		},
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7"}, nil
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "__shim", "go", "version")
	if err == nil {
		t.Fatal("expected an error when binary does not exist")
	}
}

func TestShimCommand_GoShimReturnsGoBinaryPathError(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoStore: stubGoStore{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("binary path boom")
			},
		},
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7"}, nil
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "__shim", "go", "version")
	if err == nil {
		t.Fatal("expected an error when GoBinaryPath fails")
	}
	if !strings.Contains(err.Error(), "binary path boom") {
		t.Fatalf("err = %q, want binary path boom", err)
	}
}

func TestShimCommand_PrefersGoResolverOverCompositeResolver(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	t.Run("should resolve Go via GoResolver without calling composite Resolver when GoResolver is configured", func(t *testing.T) {
		t.Parallel()

		// Arrange
		goBinDir := filepath.Join(t.TempDir(), "bin")
		if err := os.MkdirAll(goBinDir, 0o755); err != nil {
			t.Fatalf("mkdir go bin dir: %v", err)
		}
		goBinary := filepath.Join(goBinDir, "go")
		if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go1.27.2\\n'\n"), 0o755); err != nil {
			t.Fatalf("write fake go binary: %v", err)
		}

		application := &app.App{
			GoStore: stubGoStore{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return goBinary, nil
				},
			},
			GoResolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2", GoSource: "go.mod"}, nil
				},
			},
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					t.Fatal("composite Resolver should not be called by __shim go when GoResolver is set")
					return app.Selection{}, nil
				},
			},
		}

		// Act
		stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "__shim", "go", "version")

		// Assert
		if err != nil {
			t.Fatalf("shim go: %v", err)
		}
		if !strings.Contains(stdout, "go1.27.2") {
			t.Fatalf("stdout = %q, want go1.27.2", stdout)
		}
	})
}

func TestShimCommand_LintShimExecutesWithSelectedGoOnPath(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	t.Run("should execute resolved golangci-lint with selected Go bin directory prepended to PATH when __shim golangci-lint is called", func(t *testing.T) {
		t.Parallel()

		// Arrange
		goBinDir := filepath.Join(t.TempDir(), "go", "bin")
		if err := os.MkdirAll(goBinDir, 0o755); err != nil {
			t.Fatalf("mkdir go bin dir: %v", err)
		}
		goBinary := filepath.Join(goBinDir, "go")
		if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go version go1.27.2 darwin/arm64\\n'\n"), 0o755); err != nil {
			t.Fatalf("write fake go binary: %v", err)
		}

		lintDir := filepath.Join(t.TempDir(), "lint")
		if err := os.MkdirAll(lintDir, 0o755); err != nil {
			t.Fatalf("mkdir lint dir: %v", err)
		}
		lintBinary := filepath.Join(lintDir, "golangci-lint")
		lintScript := "#!/bin/sh\nprintf 'lint v2.14.0 using ' && go version\n"
		if err := os.WriteFile(lintBinary, []byte(lintScript), 0o755); err != nil {
			t.Fatalf("write fake lint binary: %v", err)
		}

		application := &app.App{
			GoStore: stubGoStore{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return goBinary, nil
				},
			},
			LintStore: stubLintStore{
				lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return lintBinary, nil
				},
			},
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						GoSource:    "go.mod",
						LintVersion: "v2.14.0",
						LintSource:  "local",
					}, nil
				},
			},
		}

		// Act
		stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "__shim", "golangci-lint", "version")

		// Assert
		if err != nil {
			t.Fatalf("shim golangci-lint: %v", err)
		}
		if !strings.Contains(stdout, "lint v2.14.0 using go version go1.27.2") {
			t.Fatalf("stdout = %q, want it to contain lint v2.14.0 using go version go1.27.2", stdout)
		}
	})
}

func TestShimCommand_LintShimRejectsNilLintStore(t *testing.T) {
	t.Parallel()

	t.Run("should return an error when LintStore is not configured for __shim golangci-lint", func(t *testing.T) {
		t.Parallel()

		// Arrange
		application := &app.App{
			GoStore:   stubGoStore{},
			LintStore: nil,
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2", LintVersion: "v2.14.0"}, nil
				},
			},
		}

		// Act
		_, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "__shim", "golangci-lint", "run")

		// Assert
		if err == nil {
			t.Fatal("expected an error when LintStore is nil")
		}
		if !strings.Contains(err.Error(), "shim dependencies are not configured") {
			t.Fatalf("err = %q, want not configured error", err)
		}
	})
}

func TestShimCommandGoToolchainEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	newEnvPrintingApp := func(t *testing.T) *app.App {
		t.Helper()
		goBinary := filepath.Join(t.TempDir(), "go")
		script := "#!/bin/sh\nprintf 'GOTOOLCHAIN=%s GOROOT=%s\\n' \"${GOTOOLCHAIN-unset}\" \"${GOROOT-unset}\"\n"
		if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
			t.Fatalf("write fake go binary: %v", err)
		}
		return &app.App{
			GoStore: stubGoStore{goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return goBinary, nil
			}},
			Resolver: stubSelectionResolver{currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.0"}, nil
			}},
		}
	}

	t.Run("RunE", func(t *testing.T) {
		t.Run("should run the go shim with GOTOOLCHAIN=local and no GOROOT when the shell sets both", func(t *testing.T) {
			// Arrange
			t.Setenv("GOTOOLCHAIN", "go1.26.8")
			t.Setenv("GOROOT", "/opt/go1.26.8")
			t.Setenv("FGM_EXEC_GOTOOLCHAIN", "")
			root := NewRootCmd(newEnvPrintingApp(t))

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, root, "__shim", "go")

			// Assert
			if err != nil {
				t.Fatalf("shim go: %v", err)
			}
			if stdout != "GOTOOLCHAIN=local GOROOT=unset\n" {
				t.Errorf("stdout = %q; want local toolchain without GOROOT", stdout)
			}
		})

		t.Run("should keep inherited GOTOOLCHAIN and GOROOT when FGM_EXEC_GOTOOLCHAIN is inherit", func(t *testing.T) {
			// Arrange
			t.Setenv("GOTOOLCHAIN", "go1.26.8")
			t.Setenv("GOROOT", "/opt/go1.26.8")
			t.Setenv("FGM_EXEC_GOTOOLCHAIN", "inherit")
			root := NewRootCmd(newEnvPrintingApp(t))

			// Act
			stdout, stderr, err := testutil.ExecuteCommand(t, root, "__shim", "go")

			// Assert
			if err != nil {
				t.Fatalf("shim go: %v", err)
			}
			if stdout != "GOTOOLCHAIN=go1.26.8 GOROOT=/opt/go1.26.8\n" {
				t.Errorf("stdout = %q; want inherited values", stdout)
			}
			if !strings.Contains(stderr, "GOTOOLCHAIN=go1.26.8") {
				t.Errorf("stderr = %q; want an inherit warning", stderr)
			}
		})
	})
}

func TestShimCommandForwardsArgumentsVerbatim(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	argvScript := "#!/bin/sh\nfor arg in \"$@\"; do printf '[%s]\\n' \"$arg\"; done\n"
	newApp := func(t *testing.T) *app.App {
		t.Helper()
		dir := t.TempDir()
		goBinary := filepath.Join(dir, "go")
		lintBinary := filepath.Join(dir, "golangci-lint")
		for _, path := range []string{goBinary, lintBinary} {
			if err := os.WriteFile(path, []byte(argvScript), 0o755); err != nil {
				t.Fatalf("write fake tool: %v", err)
			}
		}
		return &app.App{
			GoStore: stubGoStore{goBinaryPathFn: func(ctx context.Context, version string) (string, error) { return goBinary, nil }},
			Resolver: stubSelectionResolver{currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.0", LintVersion: "v2.9.0"}, nil
			}},
			LintStore: stubLintStore{lintBinaryPathFn: func(ctx context.Context, version string) (string, error) { return lintBinary, nil }},
		}
	}

	cases := []struct {
		name string
		args []string
	}{
		{name: "should forward a short flag with its value when go test filters tests", args: []string{"go", "test", "-run", "TestX", "./..."}},
		{name: "should forward a long flag with its value when golangci-lint limits new issues", args: []string{"golangci-lint", "run", "--new-from-rev", "HEAD~1", "./..."}},
		{name: "should forward --version to the tool when golangci-lint prints its version", args: []string{"golangci-lint", "--version"}},
		{name: "should forward -h to the tool when go prints help", args: []string{"go", "-h"}},
		{name: "should forward --help to the tool when golangci-lint prints help", args: []string{"golangci-lint", "--help"}},
		{name: "should forward a bare double dash and the arguments after it", args: []string{"go", "run", ".", "--", "-flag"}},
		{name: "should forward an empty argument and one containing spaces", args: []string{"go", "test", "", "two words"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			root := NewRootCmd(newApp(t))
			want := ""
			for _, arg := range tc.args[1:] {
				want += "[" + arg + "]\n"
			}

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, root, append([]string{"__shim"}, tc.args...)...)

			// Assert
			if err != nil {
				t.Fatalf("__shim: %v", err)
			}
			if stdout != want {
				t.Errorf("tool argv = %q; want %q", stdout, want)
			}
		})
	}
}
