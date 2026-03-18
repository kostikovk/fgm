package goupgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
)

type stubRemoteProvider struct {
	listRemoteGoVersionsFn func(ctx context.Context) ([]string, error)
}

func (s stubRemoteProvider) ListRemoteGoVersions(ctx context.Context) ([]string, error) {
	return s.listRemoteGoVersionsFn(ctx)
}

type stubInstaller struct {
	installGoVersionFn func(ctx context.Context, version string) (string, error)
}

func (s stubInstaller) InstallGoVersion(ctx context.Context, version string) (string, error) {
	return s.installGoVersionFn(ctx, version)
}

type stubLintInstaller struct {
	installLintVersionFn func(ctx context.Context, version string) (string, error)
}

func (s stubLintInstaller) InstallLintVersion(ctx context.Context, version string) (string, error) {
	return s.installLintVersionFn(ctx, version)
}

type stubLintRemoteProvider struct {
	listRemoteLintVersionsFn func(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

func (s stubLintRemoteProvider) ListRemoteLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
	return s.listRemoteLintVersionsFn(ctx, goVersion)
}

type stubGlobalStore struct {
	setGlobalGoVersionFn func(ctx context.Context, version string) error
	ensureShimsFn        func() error
}

func (s stubGlobalStore) SetGlobalGoVersion(ctx context.Context, version string) error {
	return s.setGlobalGoVersionFn(ctx, version)
}

func (s stubGlobalStore) EnsureShims() error {
	return s.ensureShimsFn()
}

func TestServiceUpgradeGlobal_InstallsLatestVersionAndSetsGlobalState(t *testing.T) {
	t.Parallel()

	var installed string
	var selected string
	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.25.8", "1.26.1", "1.24.9"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				installed = version
				return "/tmp/fgm/go/" + version, nil
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error {
				selected = version
				return nil
			},
			ensureShimsFn: func() error { return nil },
		},
	})

	result, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{})
	if err != nil {
		t.Fatalf("UpgradeGlobal: %v", err)
	}

	if installed != "1.26.1" {
		t.Fatalf("installed = %q, want %q", installed, "1.26.1")
	}
	if selected != "1.26.1" {
		t.Fatalf("selected = %q, want %q", selected, "1.26.1")
	}
	if result.Version != "1.26.1" {
		t.Fatalf("result.Version = %q, want %q", result.Version, "1.26.1")
	}
}

func TestServiceUpgradeProject_UpdatesNearestGoWorkToolchain(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	goWorkPath := filepath.Join(root, "go.work")
	if err := os.WriteFile(goWorkPath, []byte("go 1.25.0\n\ntoolchain go1.25.1\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1", "1.25.8"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
	})

	result, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("UpgradeProject: %v", err)
	}

	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		t.Fatalf("read go.work: %v", err)
	}
	if string(content) != "go 1.25.0\n\ntoolchain go1.26.1\n" {
		t.Fatalf("go.work = %q, want updated toolchain", string(content))
	}
	if result.Path != goWorkPath {
		t.Fatalf("result.Path = %q, want %q", result.Path, goWorkPath)
	}
}

func TestServiceUpgradeProject_UpdatesGoDirectiveWhenNoToolchainExists(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1", "1.25.8"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
	})

	if _, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{WorkDir: workDir}); err != nil {
		t.Fatalf("UpgradeProject: %v", err)
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if string(content) != "module example.com/demo\n\ngo 1.26.1\n" {
		t.Fatalf("go.mod = %q, want updated go directive", string(content))
	}
}

func TestServiceUpgradeGlobal_DryRunDoesNotInstallOrSetGlobalState(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				t.Fatal("InstallGoVersion should not be called during dry-run")
				return "", nil
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error {
				t.Fatal("SetGlobalGoVersion should not be called during dry-run")
				return nil
			},
			ensureShimsFn: func() error {
				t.Fatal("EnsureShims should not be called during dry-run")
				return nil
			},
		},
	})

	result, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{DryRun: true})
	if err != nil {
		t.Fatalf("UpgradeGlobal: %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if result.Version != "1.26.1" {
		t.Fatalf("Version = %q, want %q", result.Version, "1.26.1")
	}
}

func TestServiceUpgradeGlobal_WithLintInstallsRecommendedLint(t *testing.T) {
	t.Parallel()

	var installedLint string
	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				if goVersion != "1.26.1" {
					t.Fatalf("goVersion = %q, want %q", goVersion, "1.26.1")
				}
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				installedLint = version
				return "/tmp/fgm/golangci-lint/" + version, nil
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	result, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{WithLint: true})
	if err != nil {
		t.Fatalf("UpgradeGlobal: %v", err)
	}
	if installedLint != "v2.11.2" {
		t.Fatalf("installedLint = %q, want %q", installedLint, "v2.11.2")
	}
	if result.LintVersion != "v2.11.2" {
		t.Fatalf("LintVersion = %q, want %q", result.LintVersion, "v2.11.2")
	}
}

func TestServiceUpgradeProject_DryRunDoesNotRewriteMetadata(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	original := "module example.com/demo\n\ngo 1.25.0\n"
	if err := os.WriteFile(goModPath, []byte(original), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				t.Fatal("InstallGoVersion should not be called during dry-run")
				return "", nil
			},
		},
	})

	result, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{
		WorkDir: workDir,
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("UpgradeProject: %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if string(content) != original {
		t.Fatalf("go.mod = %q, want original content", string(content))
	}
}

func TestServiceUpgradeProject_UsesExplicitVersionOverride(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	var installed string
	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				installed = version
				return "/tmp/fgm/go/" + version, nil
			},
		},
	})

	if _, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{
		WorkDir: workDir,
		Version: "1.25.8",
	}); err != nil {
		t.Fatalf("UpgradeProject: %v", err)
	}
	if installed != "1.25.8" {
		t.Fatalf("installed = %q, want %q", installed, "1.25.8")
	}
}

func TestServiceUpgradeProject_WithPinnedLintInstallsPinnedVersion(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(workDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	var installedLint string
	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				installedLint = version
				return "/tmp/fgm/golangci-lint/" + version, nil
			},
		},
	})

	result, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{
		WorkDir:  workDir,
		WithLint: true,
	})
	if err != nil {
		t.Fatalf("UpgradeProject: %v", err)
	}
	if installedLint != "v2.10.1" {
		t.Fatalf("installedLint = %q, want %q", installedLint, "v2.10.1")
	}
	if result.LintVersion != "v2.10.1" {
		t.Fatalf("LintVersion = %q, want %q", result.LintVersion, "v2.10.1")
	}
}

func TestServiceUpgradeGlobal_ErrorWhenInstallFails(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("install failed")
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "install failed") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "install failed")
	}
}

func TestServiceUpgradeGlobal_ErrorWhenSetGlobalFails(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error {
				return fmt.Errorf("set global failed")
			},
			ensureShimsFn: func() error { return nil },
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "set global failed") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "set global failed")
	}
}

func TestServiceUpgradeGlobal_ErrorWhenEnsureShimsFails(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn: func() error {
				return fmt.Errorf("ensure shims failed")
			},
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "ensure shims failed") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "ensure shims failed")
	}
}

func TestServiceUpgradeGlobal_DryRunWithLint(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				t.Fatal("InstallGoVersion should not be called during dry-run")
				return "", nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				t.Fatal("InstallLintVersion should not be called during dry-run")
				return "", nil
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error {
				t.Fatal("SetGlobalGoVersion should not be called during dry-run")
				return nil
			},
			ensureShimsFn: func() error {
				t.Fatal("EnsureShims should not be called during dry-run")
				return nil
			},
		},
	})

	result, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{
		DryRun:   true,
		WithLint: true,
	})
	if err != nil {
		t.Fatalf("UpgradeGlobal: %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if result.Version != "1.26.1" {
		t.Fatalf("Version = %q, want %q", result.Version, "1.26.1")
	}
	if result.LintVersion != "v2.11.2" {
		t.Fatalf("LintVersion = %q, want %q", result.LintVersion, "v2.11.2")
	}
}

func TestServiceUpgradeProject_ErrorWhenNoGoModFound(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
	})

	_, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{WorkDir: workDir})
	if err == nil {
		t.Fatal("expected error when no go.mod or go.work exists, got nil")
	}
}

func TestServiceUpgradeProject_WithLint(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	var installedLint string
	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				installedLint = version
				return "/tmp/fgm/golangci-lint/" + version, nil
			},
		},
	})

	result, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{
		WorkDir:  workDir,
		WithLint: true,
	})
	if err != nil {
		t.Fatalf("UpgradeProject: %v", err)
	}
	if installedLint != "v2.11.2" {
		t.Fatalf("installedLint = %q, want %q", installedLint, "v2.11.2")
	}
	if result.LintVersion != "v2.11.2" {
		t.Fatalf("LintVersion = %q, want %q", result.LintVersion, "v2.11.2")
	}
}

func TestServiceUpgradeProject_DryRunWithLint(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	original := "module example.com/demo\n\ngo 1.25.0\n"
	if err := os.WriteFile(goModPath, []byte(original), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				t.Fatal("InstallGoVersion should not be called during dry-run")
				return "", nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				t.Fatal("InstallLintVersion should not be called during dry-run")
				return "", nil
			},
		},
	})

	result, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{
		WorkDir:  workDir,
		DryRun:   true,
		WithLint: true,
	})
	if err != nil {
		t.Fatalf("UpgradeProject: %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if result.LintVersion != "v2.11.2" {
		t.Fatalf("LintVersion = %q, want %q", result.LintVersion, "v2.11.2")
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if string(content) != original {
		t.Fatalf("go.mod = %q, want original content", string(content))
	}
}

func TestInstallLint_ErrorWhenInstallerNil(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		// LintInstaller intentionally nil
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{WithLint: true})
	if err == nil {
		t.Fatal("expected error when LintInstaller is nil, got nil")
	}
	if !strings.Contains(err.Error(), "golangci-lint installer is not configured") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "golangci-lint installer is not configured")
	}
}

func TestInstallLint_ErrorWhenInstallerFails(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("lint install failed")
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{WithLint: true})
	if err == nil {
		t.Fatal("expected error when lint install fails, got nil")
	}
	if !strings.Contains(err.Error(), "lint install failed") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "lint install failed")
	}
}

func TestResolveLintVersion_EmptyWhenNoProvider(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		// LintRemoteProvider intentionally nil
		// LintInstaller intentionally nil
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	result, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{WithLint: true})
	if err != nil {
		t.Fatalf("UpgradeGlobal: %v", err)
	}
	if result.LintVersion != "" {
		t.Fatalf("LintVersion = %q, want empty string", result.LintVersion)
	}
}

func TestResolveLintVersion_EmptyWhenNoVersions(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{}, nil
			},
		},
		// LintInstaller intentionally nil — resolveLintVersion returns "" so installLint returns early
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	result, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{WithLint: true})
	if err != nil {
		t.Fatalf("UpgradeGlobal: %v", err)
	}
	if result.LintVersion != "" {
		t.Fatalf("LintVersion = %q, want empty string", result.LintVersion)
	}
}

func TestRewriteVersionMetadata_EmptyToolchainDirective(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module foo\n\ntoolchain\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
	})

	_, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{WorkDir: workDir})
	if err == nil {
		t.Fatal("expected error for bare toolchain directive, got nil")
	}
	if !strings.Contains(err.Error(), "toolchain directive is empty") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "toolchain directive is empty")
	}
}

func TestRewriteVersionMetadata_NoGoDirective(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module foo\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
	})

	_, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{WorkDir: workDir})
	if err == nil {
		t.Fatal("expected error for missing go directive, got nil")
	}
	if !strings.Contains(err.Error(), "no go version directive found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "no go version directive found")
	}
}

func TestUpgradeGlobal_DryRunWithLint_ResolveLintVersionError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	// Create a malformed .fgm.toml to make pinnedlint.ResolvePinned error.
	if err := os.WriteFile(filepath.Join(workDir, ".fgm.toml"), []byte("[[[invalid toml"), 0o644); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{
		WorkDir:  workDir,
		DryRun:   true,
		WithLint: true,
	})
	if err == nil {
		t.Fatal("expected error when resolveLintVersion fails in dry-run global, got nil")
	}
}

func TestUpgradeProject_TargetVersionError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return nil, fmt.Errorf("remote list boom")
			},
		},
	})

	_, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{WorkDir: workDir})
	if err == nil {
		t.Fatal("expected error when targetVersion fails, got nil")
	}
	if !strings.Contains(err.Error(), "remote list boom") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "remote list boom")
	}
}

func TestUpgradeProject_DryRunWithLint_ResolveLintVersionError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	// Create a malformed .fgm.toml so pinnedlint.ResolvePinned returns an error.
	if err := os.WriteFile(filepath.Join(workDir, ".fgm.toml"), []byte("[[[invalid toml"), 0o644); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
	})

	_, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{
		WorkDir:  workDir,
		DryRun:   true,
		WithLint: true,
	})
	if err == nil {
		t.Fatal("expected error when resolveLintVersion fails in dry-run project, got nil")
	}
}

func TestUpgradeProject_InstallGoVersionError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("install go boom")
			},
		},
	})

	_, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{WorkDir: workDir})
	if err == nil {
		t.Fatal("expected error when InstallGoVersion fails, got nil")
	}
	if !strings.Contains(err.Error(), "install go boom") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "install go boom")
	}
}

func TestUpgradeProject_WithLint_InstallLintError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2"}}, nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("lint install project boom")
			},
		},
	})

	_, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{
		WorkDir:  workDir,
		WithLint: true,
	})
	if err == nil {
		t.Fatal("expected error when installLint fails in project upgrade, got nil")
	}
	if !strings.Contains(err.Error(), "lint install project boom") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "lint install project boom")
	}
}

func TestResolveLintVersion_PinnedLintError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	// Malformed .fgm.toml causes pinnedlint.ResolvePinned to error.
	if err := os.WriteFile(filepath.Join(workDir, ".fgm.toml"), []byte("[[[bad toml"), 0o644); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{
		WorkDir:  workDir,
		WithLint: true,
	})
	if err == nil {
		t.Fatal("expected error when pinnedlint.ResolvePinned fails, got nil")
	}
}

func TestResolveLintVersion_ListRemoteLintVersionsError(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return nil, fmt.Errorf("lint remote boom")
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{WithLint: true})
	if err == nil {
		t.Fatal("expected error when ListRemoteLintVersions fails, got nil")
	}
	if !strings.Contains(err.Error(), "lint remote boom") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "lint remote boom")
	}
}

func TestLatestVersion_ListRemoteGoVersionsError(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return nil, fmt.Errorf("remote go boom")
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error when ListRemoteGoVersions fails, got nil")
	}
	if !strings.Contains(err.Error(), "remote go boom") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "remote go boom")
	}
}

func TestFindProjectMetadataFile_GoWorkStatError(t *testing.T) {
	t.Parallel()

	// Create a directory named "go.work" so that FindNearestFile
	// finds it via os.Stat (err == nil) but then rejects it because
	// info.IsDir() is true, which means it keeps walking upward.
	// Instead, we need an os.Stat error that is NOT os.IsNotExist.
	// We can achieve this by creating a symlink that points to itself
	// (broken symlink loop) named "go.work" which causes os.Stat to
	// return a "too many levels of symbolic links" error.
	root := t.TempDir()
	workDir := filepath.Join(root, "sub")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create a symlink loop: go.work -> go.work (self-referencing).
	if err := os.Symlink(filepath.Join(root, "go.work"), filepath.Join(root, "go.work")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.26.1"}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
	})

	_, err := service.UpgradeProject(context.Background(), app.GoUpgradeOptions{WorkDir: workDir})
	if err == nil {
		t.Fatal("expected error when go.work stat fails with non-NotExist error, got nil")
	}
}

func TestRewriteVersionMetadata_StatError(t *testing.T) {
	t.Parallel()

	// Pass a path that does not exist to rewriteVersionMetadata.
	err := rewriteVersionMetadata("/nonexistent/path/to/go.mod", "1.26.1")
	if err == nil {
		t.Fatal("expected error when os.Stat fails, got nil")
	}
	if !strings.Contains(err.Error(), "stat") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "stat")
	}
}

func TestRewriteVersionMetadata_ReadFileError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goModPath := filepath.Join(workDir, "go.mod")
	// Write go.mod then make it unreadable. os.Stat will succeed but os.ReadFile will fail.
	if err := os.WriteFile(goModPath, []byte("module x\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.Chmod(goModPath, 0o000); err != nil {
		t.Fatalf("chmod go.mod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(goModPath, 0o644) })

	err := rewriteVersionMetadata(goModPath, "1.26.1")
	if err == nil {
		t.Fatal("expected error when os.ReadFile fails, got nil")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "read")
	}
}

func TestWriteLines_WriteFileError(t *testing.T) {
	t.Parallel()

	// Attempt to write to a path inside a non-existent directory.
	err := writeLines("/nonexistent/dir/go.mod", []string{"module x", "", "go 1.26.1", ""}, 0o644)
	if err == nil {
		t.Fatal("expected error when os.WriteFile fails, got nil")
	}
	if !strings.Contains(err.Error(), "write") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "write")
	}
}

func TestLatestVersion_ErrorWhenNoVersions(t *testing.T) {
	t.Parallel()

	service := New(Config{
		RemoteProvider: stubRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{}, nil
			},
		},
		Installer: stubInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version, nil
			},
		},
		GlobalStore: stubGlobalStore{
			setGlobalGoVersionFn: func(ctx context.Context, version string) error { return nil },
			ensureShimsFn:        func() error { return nil },
		},
	})

	_, err := service.UpgradeGlobal(context.Background(), app.GoUpgradeOptions{})
	if err == nil {
		t.Fatal("expected error when no remote versions available, got nil")
	}
	if !strings.Contains(err.Error(), "no remote Go versions are available") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "no remote Go versions are available")
	}
}
