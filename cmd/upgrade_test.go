package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/goupgrade"
	"github.com/koskosovu4/fgm/internal/testutil"
)

type stubGoUpgrader struct {
	upgradeGlobalFn  func(ctx context.Context) (app.GoUpgradeResult, error)
	upgradeProjectFn func(ctx context.Context, workDir string) (app.GoUpgradeResult, error)
}

func (s stubGoUpgrader) UpgradeGlobal(ctx context.Context) (app.GoUpgradeResult, error) {
	return s.upgradeGlobalFn(ctx)
}

func (s stubGoUpgrader) UpgradeProject(ctx context.Context, workDir string) (app.GoUpgradeResult, error) {
	return s.upgradeProjectFn(ctx, workDir)
}

type stubRemoteGoProvider struct {
	listRemoteGoVersionsFn func(ctx context.Context) ([]string, error)
}

func (s stubRemoteGoProvider) ListRemoteGoVersions(ctx context.Context) ([]string, error) {
	return s.listRemoteGoVersionsFn(ctx)
}

func TestUpgradeGoCommand_UpgradesGlobalVersion(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context) (app.GoUpgradeResult, error) {
				return app.GoUpgradeResult{Version: "1.26.1", Path: "global"}, nil
			},
			upgradeProjectFn: func(ctx context.Context, workDir string) (app.GoUpgradeResult, error) {
				t.Fatal("UpgradeProject should not be called")
				return app.GoUpgradeResult{}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "upgrade", "go", "--global")
	if err != nil {
		t.Fatalf("execute upgrade go --global: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Upgraded global Go to 1.26.1") {
		t.Fatalf("stdout = %q, want global upgrade line", stdout)
	}
}

func TestUpgradeGoCommand_UpgradesProjectVersion(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := app.New(app.Config{
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
	})

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
