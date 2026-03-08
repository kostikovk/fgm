package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

type stubGoStore struct {
	listLocalGoVersionsFn func(ctx context.Context) ([]string, error)
	hasGoVersionFn        func(ctx context.Context, version string) (bool, error)
	globalGoVersionFn     func(ctx context.Context) (string, bool, error)
	setGlobalGoVersionFn  func(ctx context.Context, version string) error
	deleteGoVersionFn     func(ctx context.Context, version string) (string, error)
	goBinaryPathFn        func(ctx context.Context, version string) (string, error)
	ensureShimsFn         func() error
	shimDirFn             func() string
}

func (s stubGoStore) ListLocalGoVersions(ctx context.Context) ([]string, error) {
	return s.listLocalGoVersionsFn(ctx)
}

func (s stubGoStore) HasGoVersion(ctx context.Context, version string) (bool, error) {
	if s.hasGoVersionFn == nil {
		return false, nil
	}
	return s.hasGoVersionFn(ctx, version)
}

func (s stubGoStore) GlobalGoVersion(ctx context.Context) (string, bool, error) {
	if s.globalGoVersionFn == nil {
		return "", false, nil
	}
	return s.globalGoVersionFn(ctx)
}

func (s stubGoStore) SetGlobalGoVersion(ctx context.Context, version string) error {
	if s.setGlobalGoVersionFn == nil {
		return nil
	}
	return s.setGlobalGoVersionFn(ctx, version)
}

func (s stubGoStore) DeleteGoVersion(ctx context.Context, version string) (string, error) {
	if s.deleteGoVersionFn == nil {
		return "", nil
	}
	return s.deleteGoVersionFn(ctx, version)
}

func (s stubGoStore) GoBinaryPath(ctx context.Context, version string) (string, error) {
	if s.goBinaryPathFn == nil {
		return "", nil
	}
	return s.goBinaryPathFn(ctx, version)
}

func (s stubGoStore) EnsureShims() error {
	if s.ensureShimsFn == nil {
		return nil
	}
	return s.ensureShimsFn()
}

func (s stubGoStore) ShimDir() string {
	if s.shimDirFn == nil {
		return ""
	}
	return s.shimDirFn()
}

type stubGoRemoteProvider struct {
	listRemoteGoVersionsFn func(ctx context.Context) ([]string, error)
}

func (s stubGoRemoteProvider) ListRemoteGoVersions(ctx context.Context) ([]string, error) {
	return s.listRemoteGoVersionsFn(ctx)
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

type stubLintStore struct {
	listLocalLintVersionsFn func(ctx context.Context) ([]string, error)
	deleteLintVersionFn     func(ctx context.Context, version string) (string, error)
}

func (s stubLintStore) ListLocalLintVersions(ctx context.Context) ([]string, error) {
	return s.listLocalLintVersionsFn(ctx)
}

func (s stubLintStore) DeleteLintVersion(ctx context.Context, version string) (string, error) {
	if s.deleteLintVersionFn == nil {
		return "", nil
	}
	return s.deleteLintVersionFn(ctx, version)
}

func TestVersionsGoLocal_ShowsInstalledVersions(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.24.3", "1.25.1"}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "versions", "go", "--local")
	if err != nil {
		t.Fatalf("execute versions go --local: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "1.24.3") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "1.24.3")
	}
	if !strings.Contains(stdout, "1.25.1") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "1.25.1")
	}
}

func TestVersionsGoRemote_ShowsAvailableVersions(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		GoRemoteProvider: stubGoRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.25.2", "1.25.1"}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "versions", "go", "--remote")
	if err != nil {
		t.Fatalf("execute versions go --remote: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "1.25.2") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "1.25.2")
	}
	if !strings.Contains(stdout, "1.25.1") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "1.25.1")
	}
}

func TestVersionsGoRemote_MarksCurrentVersion(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.25.1\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		GoRemoteProvider: stubGoRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.25.2", "1.25.1", "1.24.9"}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "versions", "go", "--remote", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute versions go --remote: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "* 1.25.1") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "* 1.25.1")
	}
}

func TestVersionsGolangCILintRemote_ShowsCompatibleVersions(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(
				ctx context.Context,
				goVersion string,
			) ([]app.LintVersion, error) {
				if goVersion != "1.25.0" {
					t.Fatalf("goVersion = %q, want %q", goVersion, "1.25.0")
				}

				return []app.LintVersion{
					{Version: "v2.6.2", Recommended: true},
					{Version: "v2.6.1"},
				}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(
		t,
		root,
		"versions",
		"golangci-lint",
		"--remote",
		"--go",
		"1.25.0",
	)
	if err != nil {
		t.Fatalf("execute versions golangci-lint --remote: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "* v2.6.2") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "* v2.6.2")
	}
	if !strings.Contains(stdout, "v2.6.1") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "v2.6.1")
	}
}

func TestVersionsGolangCILintRemote_UsesResolvedRepoGoVersion(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.24.3\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(
				ctx context.Context,
				goVersion string,
			) ([]app.LintVersion, error) {
				if goVersion != "1.24.3" {
					t.Fatalf("goVersion = %q, want %q", goVersion, "1.24.3")
				}

				return []app.LintVersion{{Version: "v2.5.0", Recommended: true}}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(
		t,
		root,
		"versions",
		"golangci-lint",
		"--remote",
		"--chdir",
		tempDir,
	)
	if err != nil {
		t.Fatalf("execute versions golangci-lint --remote: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "* v2.5.0") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "* v2.5.0")
	}
}

func TestVersionsGolangCILintLocal_ShowsInstalledVersions(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"v2.11.2", "v2.11.1"}, nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "versions", "golangci-lint", "--local")
	if err != nil {
		t.Fatalf("execute versions golangci-lint --local: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "v2.11.2") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "v2.11.2")
	}
	if !strings.Contains(stdout, "v2.11.1") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "v2.11.1")
	}
}
