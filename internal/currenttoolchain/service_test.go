package currenttoolchain

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
)

type stubGoResolver struct {
	currentFn func(ctx context.Context, workDir string) (app.Selection, error)
}

func (s stubGoResolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	return s.currentFn(ctx, workDir)
}

type stubLintStore struct {
	listLocalLintVersionsFn func(ctx context.Context) ([]string, error)
}

func (s stubLintStore) ListLocalLintVersions(ctx context.Context) ([]string, error) {
	return s.listLocalLintVersionsFn(ctx)
}

func (s stubLintStore) DeleteLintVersion(ctx context.Context, version string) (string, error) {
	return "", nil
}

type stubLintRemoteProvider struct {
	listRemoteLintVersionsFn func(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

func (s stubLintRemoteProvider) ListRemoteLintVersions(
	ctx context.Context,
	goVersion string,
) ([]app.LintVersion, error) {
	return s.listRemoteLintVersionsFn(ctx, goVersion)
}

func TestServiceCurrent_PrefersCompatibleInstalledLintVersion(t *testing.T) {
	t.Parallel()

	service := New(Config{
		GoResolver: stubGoResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.0", GoSource: "go.mod"}, nil
			},
		},
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"v2.11.1"}, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{
					{Version: "v2.11.2", Recommended: true},
					{Version: "v2.11.1"},
				}, nil
			},
		},
	})

	selection, err := service.Current(context.Background(), ".")
	if err != nil {
		t.Fatalf("Current: %v", err)
	}

	if selection.LintVersion != "v2.11.1" {
		t.Fatalf("LintVersion = %q, want %q", selection.LintVersion, "v2.11.1")
	}
	if selection.LintSource != "local" {
		t.Fatalf("LintSource = %q, want %q", selection.LintSource, "local")
	}
}

func TestServiceCurrent_ReturnsGoOnlyWhenNoCompatibleLintKnown(t *testing.T) {
	t.Parallel()

	service := New(Config{
		GoResolver: stubGoResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.0", GoSource: "go.mod"}, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return nil, nil
			},
		},
	})

	selection, err := service.Current(context.Background(), ".")
	if err != nil {
		t.Fatalf("Current: %v", err)
	}

	if selection.GoVersion != "1.25.0" {
		t.Fatalf("GoVersion = %q, want %q", selection.GoVersion, "1.25.0")
	}
	if selection.LintVersion != "" {
		t.Fatalf("LintVersion = %q, want empty", selection.LintVersion)
	}
}

func TestServiceCurrent_PrefersPinnedLintVersionFromRepoConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		GoResolver: stubGoResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.0", GoSource: "go.mod"}, nil
			},
		},
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"v2.11.1"}, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{
					{Version: "v2.11.2", Recommended: true},
					{Version: "v2.11.1"},
				}, nil
			},
		},
	})

	selection, err := service.Current(context.Background(), workDir)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}

	if selection.LintVersion != "v2.10.1" {
		t.Fatalf("LintVersion = %q, want %q", selection.LintVersion, "v2.10.1")
	}
	if selection.LintSource != "config" {
		t.Fatalf("LintSource = %q, want %q", selection.LintSource, "config")
	}
}

func TestServiceCurrent_TreatsAutoPinnedLintAsFallbackToAutoResolution(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"auto\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	service := New(Config{
		GoResolver: stubGoResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.0", GoSource: "go.mod"}, nil
			},
		},
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"v2.11.1"}, nil
			},
		},
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{
					{Version: "v2.11.2", Recommended: true},
					{Version: "v2.11.1"},
				}, nil
			},
		},
	})

	selection, err := service.Current(context.Background(), workDir)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}

	if selection.LintVersion != "v2.11.1" {
		t.Fatalf("LintVersion = %q, want %q", selection.LintVersion, "v2.11.1")
	}
	if selection.LintSource != "local" {
		t.Fatalf("LintSource = %q, want %q", selection.LintSource, "local")
	}
}
