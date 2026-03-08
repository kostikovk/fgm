package goupgrade

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
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
