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
	"github.com/kostikovk/fgm/internal/lintconfig"
	"github.com/kostikovk/fgm/internal/resolve"
	"github.com/kostikovk/fgm/internal/testutil"
)

func TestCurrentCommand_ResolvesGoVersionFromGoMod(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.24.0\n"
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		Resolver: resolve.New(nil),
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "current", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "go 1.24.0") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "go 1.24.0")
	}
}

func TestCurrentCommand_PrefersToolchainDirective(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.24.0\n\ntoolchain go1.24.3\n"
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		Resolver: resolve.New(nil),
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "current", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "go 1.24.3") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "go 1.24.3")
	}
}

func TestCurrentCommand_PrefersGoWorkOverNestedGoMod(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspace")
	moduleDir := filepath.Join(workspaceDir, "services", "api")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("mkdir module dir: %v", err)
	}

	goWork := "go 1.24.0\n\ntoolchain go1.25.1\n\nuse ./services/api\n"
	if err := os.WriteFile(filepath.Join(workspaceDir, "go.work"), []byte(goWork), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}

	goMod := "module example.com/api\n\ngo 1.23.0\n"
	if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		Resolver: resolve.New(nil),
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "current", "--chdir", moduleDir)
	if err != nil {
		t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "go 1.25.1") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "go 1.25.1")
	}
}

func TestCurrentCommand_FallsBackToGlobalVersionOutsideRepos(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store := stubGoStore{
		globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
			return "1.25.7", true, nil
		},
	}
	application := &app.App{
		Resolver: resolve.New(store),
		GoStore:  store,
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "current", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "go 1.25.7 (global)") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "go 1.25.7 (global)")
	}
}

func TestCurrentCommand_DisplaysCompatibleLintVersion(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		Resolver: currenttoolchain.New(currenttoolchain.Config{
			GoResolver: resolve.New(nil),
			LintRemoteProvider: stubLintRemoteProvider{
				listRemoteLintVersionsFn: func(
					ctx context.Context,
					goVersion string,
				) ([]app.LintVersion, error) {
					if goVersion != "1.25.0" {
						t.Fatalf("goVersion = %q, want %q", goVersion, "1.25.0")
					}
					return []app.LintVersion{
						{Version: "v2.11.2", Recommended: true},
						{Version: "v2.11.1"},
					}, nil
				},
			},
		}),
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "current", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "go 1.25.0") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "go 1.25.0")
	}
	if !strings.Contains(stdout, "golangci-lint v2.11.2 (remote)") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "golangci-lint v2.11.2 (remote)")
	}
}

func TestCurrentCommand_PrefersPinnedLintVersionFromRepoConfig(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(tempDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	application := &app.App{
		Resolver: currenttoolchain.New(currenttoolchain.Config{
			GoResolver: resolve.New(nil),
			LintRemoteProvider: stubLintRemoteProvider{
				listRemoteLintVersionsFn: func(
					ctx context.Context,
					goVersion string,
				) ([]app.LintVersion, error) {
					return []app.LintVersion{
						{Version: "v2.11.2", Recommended: true},
					}, nil
				},
			},
		}),
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "current", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "golangci-lint v2.10.1 (config)") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "golangci-lint v2.10.1 (config)")
	}
}

func TestCurrentCommand_ShowsGoSourceLabelsForWorkspaceFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.24.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		Resolver: resolve.New(nil),
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "current", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "go 1.24.0 (go.mod)") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "go 1.24.0 (go.mod)")
	}
}

func TestCurrentCommand_ShowsLocalLintSourceLabel(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion:   "1.25.7",
					GoSource:    "global",
					LintVersion: "v2.11.2",
					LintSource:  "local",
				}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "current")
	if err != nil {
		t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "golangci-lint v2.11.2 (local)") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "golangci-lint v2.11.2 (local)")
	}
}

func TestFormatCurrentLine_CoversSourceLabelBranches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{name: "empty", source: "", want: "go 1.25.7"},
		{name: "global", source: "global", want: "go 1.25.7 (global)"},
		{name: "config", source: "config", want: "go 1.25.7 (config)"},
		{name: "local", source: "local", want: "go 1.25.7 (local)"},
		{name: "remote", source: "remote", want: "go 1.25.7 (remote)"},
		{name: "path", source: "/tmp/project/go.work", want: "go 1.25.7 (go.work)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := formatCurrentLine("go", "1.25.7", tc.source)
			if got != tc.want {
				t.Fatalf("formatCurrentLine() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestWorkDirResolution_RelativeAndEnvScenarios(t *testing.T) {
	t.Run("should resolve parent metadata when invoked from a nested directory", func(t *testing.T) {
		// Arrange
		workspaceDir := t.TempDir()
		moduleDir := filepath.Join(workspaceDir, "services", "api")
		nestedDir := filepath.Join(moduleDir, "internal", "handler")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("mkdir nested dir: %v", err)
		}
		goWork := "go 1.24.0\n\ntoolchain go1.25.1\n\nuse ./services/api\n"
		if err := os.WriteFile(filepath.Join(workspaceDir, "go.work"), []byte(goWork), 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}
		goMod := "module example.com/api\n\ngo 1.23.0\n"
		if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(goMod), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}
		t.Chdir(nestedDir)

		application := &app.App{
			Resolver: resolve.New(nil),
		}
		root := NewRootCmd(application)

		// Act
		stdout, stderr, err := testutil.ExecuteCommand(t, root, "current")

		// Assert
		if err != nil {
			t.Fatalf("execute current: %v\nstderr:\n%s", err, stderr)
		}
		if !strings.Contains(stdout, "go 1.25.1 (go.work)") {
			t.Fatalf("stdout = %q, want %q", stdout, "go 1.25.1 (go.work)")
		}
	})

	t.Run("should find parent pin and lint config when chdir is relative", func(t *testing.T) {
		// Arrange
		repoDir := t.TempDir()
		cwdDir := filepath.Join(repoDir, "cmd")
		nestedDir := filepath.Join(cwdDir, "app", "sub")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("mkdir nested dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module example.com/repo\n\ngo 1.25.0\n"), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(repoDir, ".fgm.toml"), []byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"), 0o644); err != nil {
			t.Fatalf("write .fgm.toml: %v", err)
		}
		if err := os.WriteFile(filepath.Join(repoDir, ".golangci.yml"), []byte("version: \"2\"\nlinters:\n  enable:\n    - govet\n"), 0o644); err != nil {
			t.Fatalf("write .golangci.yml: %v", err)
		}
		t.Chdir(cwdDir)

		resolver := currenttoolchain.New(currenttoolchain.Config{
			GoResolver: resolve.New(nil),
		})
		lintSvc, err := lintconfig.New(lintconfig.Config{Resolver: resolver})
		if err != nil {
			t.Fatalf("lintconfig.New: %v", err)
		}
		application := &app.App{
			Resolver:            resolver,
			LintConfigGenerator: lintSvc,
			LintDoctor:          lintSvc,
		}

		// Act
		currentOut, currentErrOut, currentErr := testutil.ExecuteCommand(t, NewRootCmd(application), "current", "--chdir", "./app/sub")
		doctorOut, doctorErrOut, doctorErr := testutil.ExecuteCommand(t, NewRootCmd(application), "lint", "doctor", "--chdir", "./app/sub")
		pinOut, pinErrOut, pinErr := testutil.ExecuteCommand(t, NewRootCmd(application), "pin", "golangci-lint", "v2.11.2", "--chdir", "./app/sub")
		initOut, initErrOut, initErr := testutil.ExecuteCommand(t, NewRootCmd(application), "lint", "init", "--preset", "minimal", "--force", "--chdir", "./app/sub")

		// Assert
		if currentErr != nil {
			t.Fatalf("current --chdir ./app/sub: %v\nstderr:\n%s", currentErr, currentErrOut)
		}
		if !strings.Contains(currentOut, "go 1.25.0 (go.mod)") || !strings.Contains(currentOut, "golangci-lint v2.10.1 (config)") {
			t.Fatalf("currentOut = %q, want parent go.mod and pinned config", currentOut)
		}
		if doctorErr != nil {
			t.Fatalf("lint doctor --chdir ./app/sub: %v\nstderr:\n%s", doctorErr, doctorErrOut)
		}
		if strings.Contains(doctorOut, "no .golangci.yml found") {
			t.Fatalf("doctorOut = %q, expected parent .golangci.yml to be found", doctorOut)
		}
		if pinErr != nil {
			t.Fatalf("pin --chdir ./app/sub: %v\nstderr:\n%s", pinErr, pinErrOut)
		}
		updatedPin, err := os.ReadFile(filepath.Join(repoDir, ".fgm.toml"))
		if err != nil {
			t.Fatalf("read parent .fgm.toml: %v", err)
		}
		if !strings.Contains(string(updatedPin), "v2.11.2") {
			t.Fatalf("parent .fgm.toml = %q, want updated pin v2.11.2 (pinOut=%q)", string(updatedPin), pinOut)
		}
		if initErr != nil {
			t.Fatalf("lint init --chdir ./app/sub: %v\nstderr:\n%s", initErr, initErrOut)
		}
		if !strings.Contains(initOut, "Generated .golangci.yml (preset: minimal)") {
			t.Fatalf("initOut = %q, want minimal preset confirmation", initOut)
		}
		if _, err := os.Stat(filepath.Join(nestedDir, ".golangci.yml")); !os.IsNotExist(err) {
			t.Fatalf("expected .golangci.yml not to be created in nestedDir, err=%v", err)
		}
	})

	t.Run("should use FGM_CHDIR when chdir flag is omitted", func(t *testing.T) {
		// Arrange
		repoDir := t.TempDir()
		nestedDir := filepath.Join(repoDir, "services", "billing")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("mkdir nested dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module example.com/repo\n\ngo 1.26.1\n"), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}

		t.Chdir(repoDir)
		t.Setenv("FGM_CHDIR", "./services/billing")

		resolver := resolve.New(nil)
		application := &app.App{
			Resolver: resolver,
		}

		// Act
		currentOut, currentErrOut, currentErr := testutil.ExecuteCommand(t, NewRootCmd(application), "current")

		// Assert
		if currentErr != nil {
			t.Fatalf("current with FGM_CHDIR: %v\nstderr:\n%s", currentErr, currentErrOut)
		}
		if !strings.Contains(currentOut, "go 1.26.1 (go.mod)") {
			t.Fatalf("currentOut = %q, want go 1.26.1 (go.mod)", currentOut)
		}

		if runtime.GOOS == "windows" {
			return
		}

		// Arrange (Unix child-process cwd verification)
		goBinDir := filepath.Join(t.TempDir(), "go-bin")
		if err := os.MkdirAll(goBinDir, 0o755); err != nil {
			t.Fatalf("mkdir goBinDir: %v", err)
		}
		goBinary := filepath.Join(goBinDir, "go")
		if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'cwd=%s\\n' \"$(pwd)\"\n"), 0o755); err != nil {
			t.Fatalf("write fake go: %v", err)
		}
		goStore := stubGoStore{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return goBinary, nil
			},
		}
		application.GoStore = goStore
		application.Executor = execenv.New(execenv.Config{
			Resolver:  resolver,
			GoLocator: goStore,
		})

		// Act
		execOut, execErrOut, execErr := testutil.ExecuteCommand(t, NewRootCmd(application), "exec", "--", "go", "env")

		// Assert
		if execErr != nil {
			t.Fatalf("exec with FGM_CHDIR: %v\nstderr:\n%s", execErr, execErrOut)
		}
		wantExecDir, err := filepath.EvalSymlinks(nestedDir)
		if err != nil {
			wantExecDir = nestedDir
		}
		if !strings.Contains(execOut, wantExecDir) && !strings.Contains(execOut, nestedDir) {
			t.Fatalf("execOut = %q, want child cwd %q", execOut, wantExecDir)
		}
	})

	t.Run("should prefer explicit chdir when FGM_CHDIR is set", func(t *testing.T) {
		// Arrange
		envRepoDir := t.TempDir()
		flagRepoDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(envRepoDir, "go.mod"), []byte("module example.com/env\n\ngo 1.24.0\n"), 0o644); err != nil {
			t.Fatalf("write env go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(flagRepoDir, "go.mod"), []byte("module example.com/flag\n\ngo 1.26.1\n"), 0o644); err != nil {
			t.Fatalf("write flag go.mod: %v", err)
		}
		t.Setenv("FGM_CHDIR", envRepoDir)

		application := &app.App{
			Resolver: resolve.New(nil),
		}

		// Act
		stdout, stderr, err := testutil.ExecuteCommand(t, NewRootCmd(application), "current", "--chdir", flagRepoDir)

		// Assert
		if err != nil {
			t.Fatalf("current --chdir with FGM_CHDIR: %v\nstderr:\n%s", err, stderr)
		}
		if !strings.Contains(stdout, "go 1.26.1 (go.mod)") {
			t.Fatalf("stdout = %q, want go 1.26.1 (go.mod) from explicit --chdir", stdout)
		}
	})
}

func TestCurrentCommandGoPolicyLabels(t *testing.T) {
	t.Parallel()

	runCurrent := func(t *testing.T, selection app.Selection) (string, string) {
		t.Helper()
		application := &app.App{Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) { return selection, nil },
		}}
		stdout, stderr, err := testutil.ExecuteCommand(t, NewRootCmd(application), "current")
		if err != nil {
			t.Fatalf("current: %v", err)
		}
		return stdout, stderr
	}

	t.Run("RunE", func(t *testing.T) {
		t.Parallel()

		t.Run("should name the base go directive when the selection comes from latest-patch", func(t *testing.T) {
			t.Parallel()

			// Arrange
			selection := app.Selection{GoVersion: "1.26.9", GoSource: "latest-patch", GoSourceFile: "/repo/go.mod", GoBase: "1.26.0"}

			// Act
			stdout, _ := runCurrent(t, selection)

			// Assert
			if stdout != "go 1.26.9 (latest-patch of go.mod 1.26.0)\n" {
				t.Errorf("stdout = %q; want the latest-patch label", stdout)
			}
		})

		t.Run("should print selection warnings on stderr when the selection may be stale", func(t *testing.T) {
			t.Parallel()

			// Arrange
			selection := app.Selection{GoVersion: "1.28.1", GoSource: "latest-patch", GoSourceFile: "/repo/go.mod", GoBase: "1.28.0", Warnings: []string{"selection may be stale"}}

			// Act
			_, stderr := runCurrent(t, selection)

			// Assert
			if !strings.Contains(stderr, "selection may be stale") {
				t.Errorf("stderr = %q; want the warning", stderr)
			}
		})
	})
}

func TestCurrentCommandDetectedLintLabel(t *testing.T) {
	t.Parallel()

	t.Run("RunE", func(t *testing.T) {
		t.Parallel()

		t.Run("should name the file and line when the linter version was detected", func(t *testing.T) {
			t.Parallel()

			// Arrange
			workDir := t.TempDir()
			application := &app.App{Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, dir string) (app.Selection, error) {
					return app.Selection{
						GoVersion: "1.26.0", GoSource: filepath.Join(dir, "go.mod"),
						LintVersion: "v2.5.0", LintSource: "detected",
						LintSourceFile: filepath.Join(dir, "Makefile"), LintSourceLine: 31,
					}, nil
				},
			}}

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "current", "--chdir", workDir)

			// Assert
			if err != nil {
				t.Fatalf("current: %v", err)
			}
			if !strings.Contains(stdout, "golangci-lint v2.5.0 (detected: Makefile:31)") {
				t.Errorf("stdout = %q; want the detected label", stdout)
			}
		})
	})
}
