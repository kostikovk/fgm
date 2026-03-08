package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/goupgrade"
	"github.com/koskosovu4/fgm/internal/testutil"
)

type stubGoUpgrader struct {
	upgradeGlobalFn  func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error)
	upgradeProjectFn func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error)
}

func (s stubGoUpgrader) UpgradeGlobal(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
	return s.upgradeGlobalFn(ctx, options)
}

func (s stubGoUpgrader) UpgradeProject(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
	return s.upgradeProjectFn(ctx, options)
}

type stubRemoteGoProvider struct {
	listRemoteGoVersionsFn func(ctx context.Context) ([]string, error)
}

func (s stubRemoteGoProvider) ListRemoteGoVersions(ctx context.Context) ([]string, error) {
	return s.listRemoteGoVersionsFn(ctx)
}

func TestUpgradeGoCommand_UpgradesGlobalVersion(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				return app.GoUpgradeResult{Version: "1.26.1", Path: "global"}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeProject should not be called")
				return app.GoUpgradeResult{}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--global")
	if err != nil {
		t.Fatalf("execute upgrade go --global: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Upgraded global Go to 1.26.1") {
		t.Fatalf("stdout = %q, want global upgrade line", stdout)
	}
}

func TestUpgradeGoCommand_DryRunGlobalVersion(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				if !options.DryRun {
					t.Fatal("DryRun = false, want true")
				}
				return app.GoUpgradeResult{Version: "1.26.1", Path: "global", DryRun: true}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeProject should not be called")
				return app.GoUpgradeResult{}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--global", "--dry-run")
	if err != nil {
		t.Fatalf("execute upgrade go --global --dry-run: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Would upgrade global Go to 1.26.1") {
		t.Fatalf("stdout = %q, want dry-run global upgrade line", stdout)
	}
}

func TestUpgradeGoCommand_GlobalWithLintReportsLintInstall(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				if !options.WithLint {
					t.Fatal("WithLint = false, want true")
				}
				return app.GoUpgradeResult{
					Version:     "1.26.1",
					Path:        "global",
					LintVersion: "v2.11.2",
				}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeProject should not be called")
				return app.GoUpgradeResult{}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--global", "--with-lint")
	if err != nil {
		t.Fatalf("execute upgrade go --global --with-lint: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Installed golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want lint install line", stdout)
	}
}

func TestUpgradeGoCommand_UpgradesProjectVersion(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		GoUpgrader: goupgrade.New(goupgrade.Config{
			RemoteProvider: stubRemoteGoProvider{
				listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"1.26.1", "1.25.8"}, nil
				},
			},
			Installer: stubGoInstaller{
				installGoVersionFn: func(ctx context.Context, version string) (string, error) {
					if version != "1.26.1" {
						t.Fatalf("version = %q, want %q", version, "1.26.1")
					}
					return "/tmp/fgm/go/1.26.1", nil
				},
			},
		}),
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--project", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute upgrade go --project: %v\nstderr:\n%s", err, stderr)
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(content), "go 1.26.1") {
		t.Fatalf("go.mod = %q, want upgraded go directive", string(content))
	}
	if !strings.Contains(stdout, "Upgraded project Go to 1.26.1") {
		t.Fatalf("stdout = %q, want project upgrade line", stdout)
	}
}

func TestUpgradeGoCommand_DryRunProjectVersion(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goModPath := filepath.Join(tempDir, "go.mod")
	original := "module example.com/demo\n\ngo 1.25.0\n"
	if err := os.WriteFile(goModPath, []byte(original), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		GoUpgrader: goupgrade.New(goupgrade.Config{
			RemoteProvider: stubRemoteGoProvider{
				listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"1.26.1"}, nil
				},
			},
			Installer: stubGoInstaller{
				installGoVersionFn: func(ctx context.Context, version string) (string, error) {
					t.Fatal("InstallGoVersion should not be called during dry-run")
					return "", nil
				},
			},
		}),
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--project", "--dry-run", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute upgrade go --project --dry-run: %v\nstderr:\n%s", err, stderr)
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if string(content) != original {
		t.Fatalf("go.mod = %q, want original content", string(content))
	}
	if !strings.Contains(stdout, "Would upgrade project Go to 1.26.1") {
		t.Fatalf("stdout = %q, want dry-run project line", stdout)
	}
}

func TestUpgradeGoCommand_DryRunProjectWithLintReportsLintInstall(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeGlobal should not be called")
				return app.GoUpgradeResult{}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				if !options.WithLint || !options.DryRun {
					t.Fatalf("options = %+v, want WithLint and DryRun", options)
				}
				return app.GoUpgradeResult{
					Version:     "1.26.1",
					Path:        filepath.Join(tempDir, "go.mod"),
					LintVersion: "v2.11.2",
					DryRun:      true,
				}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--project", "--dry-run", "--with-lint", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute upgrade go --project --dry-run --with-lint: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Would install golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want lint dry-run line", stdout)
	}
}

func TestUpgradeGoCommand_GlobalUpgraderErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				return app.GoUpgradeResult{}, fmt.Errorf("global upgrade boom")
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeProject should not be called")
				return app.GoUpgradeResult{}, nil
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--global")
	if err == nil {
		t.Fatal("expected an error when UpgradeGlobal fails")
	}
	if !strings.Contains(err.Error(), "global upgrade boom") {
		t.Fatalf("err = %q, want global upgrade boom", err)
	}
}

func TestUpgradeGoCommand_ProjectUpgraderErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeGlobal should not be called")
				return app.GoUpgradeResult{}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				return app.GoUpgradeResult{}, fmt.Errorf("project upgrade boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--project")
	if err == nil {
		t.Fatal("expected an error when UpgradeProject fails")
	}
	if !strings.Contains(err.Error(), "project upgrade boom") {
		t.Fatalf("err = %q, want project upgrade boom", err)
	}
}

func TestUpgradeGoCommand_RejectsBothGlobalAndProject(t *testing.T) {
	t.Parallel()

	application := &app.App{}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--global", "--project")
	if err == nil {
		t.Fatal("expected an error when both --global and --project are set")
	}
	if !strings.Contains(err.Error(), "--global and --project are mutually exclusive") {
		t.Fatalf("err = %q, want mutually exclusive error", err)
	}
}

func TestUpgradeGoCommand_RejectsNeitherGlobalNorProject(t *testing.T) {
	t.Parallel()

	application := &app.App{}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "upgrade", "go")
	if err == nil {
		t.Fatal("expected an error when neither --global nor --project is set")
	}
	if !strings.Contains(err.Error(), "provide --global or --project") {
		t.Fatalf("err = %q, want provide flag error", err)
	}
}

func TestUpgradeGoCommand_RejectsNilGoUpgrader(t *testing.T) {
	t.Parallel()

	application := &app.App{GoUpgrader: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--global")
	if err == nil {
		t.Fatal("expected an error when GoUpgrader is nil")
	}
	if !strings.Contains(err.Error(), "go upgrader is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestUpgradeGoCommand_ProjectWithLintReportsLintInstall(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeGlobal should not be called")
				return app.GoUpgradeResult{}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				if !options.WithLint {
					t.Fatal("WithLint = false, want true")
				}
				return app.GoUpgradeResult{
					Version:     "1.26.1",
					Path:        "/tmp/go.mod",
					LintVersion: "v2.11.2",
				}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--project", "--with-lint")
	if err != nil {
		t.Fatalf("execute upgrade go --project --with-lint: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Upgraded project Go to 1.26.1") {
		t.Fatalf("stdout = %q, want project upgrade line", stdout)
	}
	if !strings.Contains(stdout, "Installed golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want lint install line", stdout)
	}
}

func TestUpgradeGoCommand_DryRunGlobalWithLintReportsLintInstall(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				if !options.DryRun || !options.WithLint {
					t.Fatalf("options = %+v, want DryRun and WithLint", options)
				}
				return app.GoUpgradeResult{
					Version:     "1.26.1",
					Path:        "global",
					LintVersion: "v2.11.2",
					DryRun:      true,
				}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeProject should not be called")
				return app.GoUpgradeResult{}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--global", "--dry-run", "--with-lint")
	if err != nil {
		t.Fatalf("execute upgrade go --global --dry-run --with-lint: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Would upgrade global Go to 1.26.1") {
		t.Fatalf("stdout = %q, want dry-run global upgrade line", stdout)
	}
	if !strings.Contains(stdout, "Would install golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want lint dry-run line", stdout)
	}
}

func TestUpgradeGoCommand_UsesExplicitVersionOverride(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := &app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeGlobal should not be called")
				return app.GoUpgradeResult{}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				if options.Version != "1.25.8" {
					t.Fatalf("Version = %q, want %q", options.Version, "1.25.8")
				}
				return app.GoUpgradeResult{Version: "1.25.8", Path: options.WorkDir}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--project", "--to", "1.25.8", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute upgrade go --project --to: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Upgraded project Go to 1.25.8") {
		t.Fatalf("stdout = %q, want explicit version line", stdout)
	}
}
