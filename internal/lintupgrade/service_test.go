package lintupgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
)

type stubResolver struct {
	currentFn func(ctx context.Context, workDir string) (app.Selection, error)
}

func (s stubResolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	return s.currentFn(ctx, workDir)
}

type stubRemoteProvider struct {
	listRemoteLintVersionsFn func(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

func (s stubRemoteProvider) ListRemoteLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
	return s.listRemoteLintVersionsFn(ctx, goVersion)
}

type stubInstaller struct {
	installLintVersionFn func(ctx context.Context, version string) (string, error)
}

func (s stubInstaller) InstallLintVersion(ctx context.Context, version string) (string, error) {
	return s.installLintVersionFn(ctx, version)
}

func TestUpgrade_InstallsRecommendedVersion(t *testing.T) {
	t.Parallel()

	var installed string
	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.1"}, nil
			},
		},
		RemoteProvider: stubRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				if goVersion != "1.26.1" {
					t.Fatalf("goVersion = %q, want %q", goVersion, "1.26.1")
				}
				return []app.LintVersion{
					{Version: "v2.11.2", Recommended: true},
					{Version: "v2.11.1"},
				}, nil
			},
		},
		Installer: stubInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				installed = version
				return "/tmp/fgm/golangci-lint/" + version, nil
			},
		},
	})

	result, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{})
	if err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if installed != "v2.11.2" {
		t.Fatalf("installed = %q, want %q", installed, "v2.11.2")
	}
	if result.Version != "v2.11.2" {
		t.Fatalf("Version = %q, want %q", result.Version, "v2.11.2")
	}
	if result.GoVersion != "1.26.1" {
		t.Fatalf("GoVersion = %q, want %q", result.GoVersion, "1.26.1")
	}
	if result.DryRun {
		t.Fatal("DryRun = true, want false")
	}
}

func TestUpgrade_DryRunDoesNotInstall(t *testing.T) {
	t.Parallel()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.1"}, nil
			},
		},
		RemoteProvider: stubRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		Installer: stubInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				t.Fatal("InstallLintVersion should not be called during dry-run")
				return "", nil
			},
		},
	})

	result, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if result.Version != "v2.11.2" {
		t.Fatalf("Version = %q, want %q", result.Version, "v2.11.2")
	}
}

func TestUpgrade_UsesExplicitVersion(t *testing.T) {
	t.Parallel()

	var installed string
	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				installed = version
				return "/tmp/fgm/golangci-lint/" + version, nil
			},
		},
	})

	result, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{Version: "v2.10.0"})
	if err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if installed != "v2.10.0" {
		t.Fatalf("installed = %q, want %q", installed, "v2.10.0")
	}
	if result.Version != "v2.10.0" {
		t.Fatalf("Version = %q, want %q", result.Version, "v2.10.0")
	}
}

func TestUpgrade_UsesPinnedVersion(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(workDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	var installed string
	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, wd string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.1"}, nil
			},
		},
		RemoteProvider: stubRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				t.Fatal("ListRemoteLintVersions should not be called when pinned")
				return nil, nil
			},
		},
		Installer: stubInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				installed = version
				return "/tmp/fgm/golangci-lint/" + version, nil
			},
		},
	})

	result, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if installed != "v2.10.1" {
		t.Fatalf("installed = %q, want %q", installed, "v2.10.1")
	}
	if result.Version != "v2.10.1" {
		t.Fatalf("Version = %q, want %q", result.Version, "v2.10.1")
	}
}

func TestUpgrade_ErrorWhenNoCompatibleVersions(t *testing.T) {
	t.Parallel()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.1"}, nil
			},
		},
		RemoteProvider: stubRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{}, nil
			},
		},
	})

	_, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error when no compatible versions, got nil")
	}
	if !strings.Contains(err.Error(), "no compatible golangci-lint versions") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "no compatible golangci-lint versions")
	}
}

func TestUpgrade_ErrorWhenResolverFails(t *testing.T) {
	t.Parallel()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{}, fmt.Errorf("resolver boom")
			},
		},
	})

	_, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "resolver boom") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "resolver boom")
	}
}

func TestUpgrade_ErrorWhenInstallFails(t *testing.T) {
	t.Parallel()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.1"}, nil
			},
		},
		RemoteProvider: stubRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		Installer: stubInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("install boom")
			},
		},
	})

	_, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "install boom") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "install boom")
	}
}

func TestUpgrade_ErrorWhenInstallerNil(t *testing.T) {
	t.Parallel()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.1"}, nil
			},
		},
		RemoteProvider: stubRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		// Installer intentionally nil
	})

	_, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error when installer is nil, got nil")
	}
	if !strings.Contains(err.Error(), "golangci-lint installer is not configured") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "golangci-lint installer is not configured")
	}
}

func TestUpgrade_ErrorWhenResolverNil(t *testing.T) {
	t.Parallel()

	service := New(Config{})

	_, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error when resolver is nil, got nil")
	}
	if !strings.Contains(err.Error(), "resolver is not configured") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "resolver is not configured")
	}
}

func TestUpgrade_ErrorWhenNoGoVersion(t *testing.T) {
	t.Parallel()

	service := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{}, nil
			},
		},
	})

	_, err := service.Upgrade(context.Background(), app.LintUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error when no Go version resolved, got nil")
	}
	if !strings.Contains(err.Error(), "no Go version resolved") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "no Go version resolved")
	}
}
