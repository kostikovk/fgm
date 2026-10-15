package cmd

import (
	"context"
	"io"
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

func TestRootAndExecHelp_FlagAndDescriptionContracts(t *testing.T) {
	t.Parallel()

	t.Run("should omit config and verbose flags from root and subcommand help when help is requested", func(t *testing.T) {
		t.Parallel()

		// Arrange
		root := NewRootCmd(&app.App{})

		// Act
		rootHelp, _, rootErr := testutil.ExecuteCommand(t, root, "--help")
		subRoot := NewRootCmd(&app.App{})
		subHelp, _, subErr := testutil.ExecuteCommand(t, subRoot, "current", "--help")

		// Assert
		if rootErr != nil {
			t.Fatalf("root --help error: %v", rootErr)
		}
		if subErr != nil {
			t.Fatalf("current --help error: %v", subErr)
		}
		for _, flagName := range []string{"--config", "--verbose"} {
			if strings.Contains(rootHelp, flagName) {
				t.Errorf("root help = %q, want it not to contain %q", rootHelp, flagName)
			}
			if strings.Contains(subHelp, flagName) {
				t.Errorf("subcommand help = %q, want it not to contain %q", subHelp, flagName)
			}
		}
	})

	t.Run("should reject removed config and verbose flags when passed on CLI", func(t *testing.T) {
		t.Parallel()

		// Arrange
		cases := []struct {
			name string
			args []string
		}{
			{
				name: "should reject config flag when passed to root command",
				args: []string{"--config", "/tmp/config.toml", "version"},
			},
			{
				name: "should reject verbose flag when passed to subcommand",
				args: []string{"version", "--verbose"},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Arrange
				root := NewRootCmd(&app.App{})

				// Act
				_, _, err := testutil.ExecuteCommand(t, root, tc.args...)

				// Assert
				if err == nil {
					t.Fatalf("ExecuteCommand(%v) succeeded; want unknown flag error", tc.args)
				}
				if !strings.Contains(err.Error(), "unknown flag") {
					t.Fatalf("err = %q, want unknown flag error", err.Error())
				}
			})
		}
	})

	t.Run("should describe resolved Go and installed selected golangci-lint in exec help when help is requested", func(t *testing.T) {
		t.Parallel()

		// Arrange
		root := NewRootCmd(&app.App{})

		// Act
		stdout, _, err := testutil.ExecuteCommand(t, root, "exec", "--help")

		// Assert
		if err != nil {
			t.Fatalf("exec --help: %v", err)
		}
		if !strings.Contains(stdout, "golangci-lint") {
			t.Fatalf("exec --help = %q, want it to mention golangci-lint", stdout)
		}
		if !strings.Contains(stdout, "non-interactive") {
			t.Fatalf("exec --help = %q, want it to mention non-interactive sessions", stdout)
		}
	})

	t.Run("should include state and workflow guidance in help for env, install, import auto, pin, upgrade, and lint commands", func(t *testing.T) {
		t.Parallel()

		// Arrange
		cases := []struct {
			name string
			args []string
			want string
		}{
			{
				name: "should suggest fgm exec for non-interactive sessions when env help is requested",
				args: []string{"env", "--help"},
				want: "fgm exec",
			},
			{
				name: "should clarify global selection is unchanged when install help is requested",
				args: []string{"install", "--help"},
				want: "global",
			},
			{
				name: "should explain symlink registration when import auto help is requested",
				args: []string{"import", "auto", "--help"},
				want: "symlink",
			},
			{
				name: "should explain .fgm.toml policy recording when pin golangci-lint help is requested",
				args: []string{"pin", "golangci-lint", "--help"},
				want: ".fgm.toml",
			},
			{
				name: "should explain --to target freezing when upgrade go help is requested",
				args: []string{"upgrade", "go", "--help"},
				want: "--to",
			},
			{
				name: "should recommend auditing with lint doctor when lint init help is requested",
				args: []string{"lint", "init", "--help"},
				want: "lint doctor",
			},
			{
				name: "should explain ERROR output and zero exit status when lint doctor help is requested",
				args: []string{"lint", "doctor", "--help"},
				want: "ERROR",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Act
				stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(&app.App{}), tc.args...)

				// Assert
				if err != nil {
					t.Fatalf("ExecuteCommand(%v): %v", tc.args, err)
				}
				if !strings.Contains(stdout, tc.want) {
					t.Fatalf("help for %v = %q, want it to contain %q", tc.args, stdout, tc.want)
				}
			})
		}
	})
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

func TestExecCommand_InstallFlagInstallsMissingGoAndLintToStderrBeforeExec(t *testing.T) {
	t.Parallel()

	t.Run("should install missing Go and golangci-lint and print status to stderr while keeping stdout clean when --install is passed", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var installedGo string
		var installedLint string
		var execCalled bool

		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						GoSource:    "go.mod",
						LintVersion: "v2.14.0",
						LintSource:  "remote",
					}, nil
				},
			},
			GoStore: stubGoStore{
				hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
					return false, nil
				},
			},
			LintStore: stubLintStore{
				lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "", os.ErrNotExist
				},
			},
			GoInstaller: stubGoInstaller{
				installGoVersionFn: func(ctx context.Context, version string) (string, error) {
					installedGo = version
					return "/tmp/fgm/go/" + version, nil
				},
			},
			LintInstaller: stubLintInstaller{
				installLintVersionFn: func(ctx context.Context, version string) (string, error) {
					installedLint = version
					return "/tmp/fgm/golangci-lint/" + version, nil
				},
			},
			Executor: stubExecutor{
				execFn: func(ctx context.Context, workDir string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					execCalled = true
					_, err := io.WriteString(stdout, "child-stdout-only\n")
					return err
				},
			},
		}

		// Act
		stdout, stderr, err := testutil.ExecuteCommand(t, NewRootCmd(application), "exec", "--install", "--", "go", "version")

		// Assert
		if err != nil {
			t.Fatalf("exec --install: %v", err)
		}
		if installedGo != "1.27.2" {
			t.Fatalf("installedGo = %q, want 1.27.2", installedGo)
		}
		if installedLint != "v2.14.0" {
			t.Fatalf("installedLint = %q, want v2.14.0", installedLint)
		}
		if !execCalled {
			t.Fatal("expected Executor.Exec to be called after installing missing toolchains")
		}
		if stdout != "child-stdout-only\n" {
			t.Fatalf("stdout = %q, want exact child stdout without install logs", stdout)
		}
		for _, want := range []string{
			"Installed Go 1.27.2 to /tmp/fgm/go/1.27.2",
			"Installed golangci-lint v2.14.0 to /tmp/fgm/golangci-lint/v2.14.0",
		} {
			if !strings.Contains(stderr, want) {
				t.Fatalf("stderr = %q, want it to contain %q", stderr, want)
			}
		}
	})
}

func TestExecCommand_InstallFlagSkipsInstallWhenAlreadyInstalled(t *testing.T) {
	t.Parallel()

	t.Run("should skip installers when selected Go and golangci-lint are already installed", func(t *testing.T) {
		t.Parallel()

		// Arrange
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						LintVersion: "v2.14.0",
					}, nil
				},
			},
			GoStore: stubGoStore{
				hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
					return true, nil
				},
			},
			LintStore: stubLintStore{
				lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/golangci-lint/v2.14.0/golangci-lint", nil
				},
			},
			GoInstaller: stubGoInstaller{
				installGoVersionFn: func(ctx context.Context, version string) (string, error) {
					t.Fatal("InstallGoVersion should not be called when Go is already installed")
					return "", nil
				},
			},
			LintInstaller: stubLintInstaller{
				installLintVersionFn: func(ctx context.Context, version string) (string, error) {
					t.Fatal("InstallLintVersion should not be called when lint is already installed")
					return "", nil
				},
			},
			Executor: stubExecutor{
				execFn: func(ctx context.Context, workDir string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					_, err := io.WriteString(stdout, "ok\n")
					return err
				},
			},
		}

		// Act
		stdout, stderr, err := testutil.ExecuteCommand(t, NewRootCmd(application), "exec", "--install", "--", "go", "test")

		// Assert
		if err != nil {
			t.Fatalf("exec --install: %v", err)
		}
		if stdout != "ok\n" {
			t.Fatalf("stdout = %q, want ok", stdout)
		}
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty stderr when nothing was installed", stderr)
		}
	})

	t.Run("should pass --install after -- separator to child command without triggering fgm toolchain installer", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var gotArgs []string
		application := &app.App{
			GoInstaller: stubGoInstaller{
				installGoVersionFn: func(ctx context.Context, version string) (string, error) {
					t.Fatal("InstallGoVersion should not be called when --install is after --")
					return "", nil
				},
			},
			Executor: stubExecutor{
				execFn: func(ctx context.Context, workDir string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					gotArgs = append([]string(nil), args...)
					return nil
				},
			},
		}

		// Act
		_, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "exec", "--", "go", "test", "--install")

		// Assert
		if err != nil {
			t.Fatalf("exec -- go test --install: %v", err)
		}
		if strings.Join(gotArgs, " ") != "go test --install" {
			t.Fatalf("gotArgs = %v, want [go test --install]", gotArgs)
		}
	})

	t.Run("should pass child flags like -v after command name even when -- separator is omitted", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var gotArgs []string
		var installedGo string
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2"}, nil
				},
			},
			GoStore: stubGoStore{
				hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
					return false, nil
				},
			},
			GoInstaller: stubGoInstaller{
				installGoVersionFn: func(ctx context.Context, version string) (string, error) {
					installedGo = version
					return "/tmp/fgm/go/" + version, nil
				},
			},
			Executor: stubExecutor{
				execFn: func(ctx context.Context, workDir string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					gotArgs = append([]string(nil), args...)
					return nil
				},
			},
		}

		// Act
		_, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "exec", "--install", "go", "test", "-v", "./...")

		// Assert
		if err != nil {
			t.Fatalf("exec --install go test -v ./...: %v", err)
		}
		if installedGo != "1.27.2" {
			t.Fatalf("installedGo = %q, want 1.27.2", installedGo)
		}
		if strings.Join(gotArgs, " ") != "go test -v ./..." {
			t.Fatalf("gotArgs = %v, want [go test -v ./...]", gotArgs)
		}
	})
}
