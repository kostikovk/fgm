package goupgrade

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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

	result, err := service.UpgradeGlobal(context.Background())
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

	result, err := service.UpgradeProject(context.Background(), workDir)
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

	if _, err := service.UpgradeProject(context.Background(), workDir); err != nil {
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
