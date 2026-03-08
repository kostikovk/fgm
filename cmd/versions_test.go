package cmd

import (
	"context"
	"fmt"
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

func (s stubGoStore) LintBinaryPath(ctx context.Context, version string) (string, error) {
	return "", nil
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
	lintBinaryPathFn        func(ctx context.Context, version string) (string, error)
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

func (s stubLintStore) LintBinaryPath(ctx context.Context, version string) (string, error) {
	if s.lintBinaryPathFn == nil {
		return "", nil
	}
	return s.lintBinaryPathFn(ctx, version)
}

func TestVersionsGoLocal_ListErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoStore: stubGoStore{
			listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return nil, fmt.Errorf("list local go boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "go", "--local")
	if err == nil {
		t.Fatal("expected an error when ListLocalGoVersions fails")
	}
	if !strings.Contains(err.Error(), "list local go boom") {
		t.Fatalf("err = %q, want list local go boom", err)
	}
}

func TestVersionsGoRemote_ListErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoRemoteProvider: stubGoRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return nil, fmt.Errorf("list remote go boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "go", "--remote")
	if err == nil {
		t.Fatal("expected an error when ListRemoteGoVersions fails")
	}
	if !strings.Contains(err.Error(), "list remote go boom") {
		t.Fatalf("err = %q, want list remote go boom", err)
	}
}

func TestVersionsLintLocal_ListErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
				return nil, fmt.Errorf("list local lint boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "golangci-lint", "--local")
	if err == nil {
		t.Fatal("expected an error when ListLocalLintVersions fails")
	}
	if !strings.Contains(err.Error(), "list local lint boom") {
		t.Fatalf("err = %q, want list local lint boom", err)
	}
}

func TestVersionsLintRemote_ListErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return nil, fmt.Errorf("list remote lint boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "golangci-lint", "--remote", "--go", "1.25.0")
	if err == nil {
		t.Fatal("expected an error when ListRemoteLintVersions fails")
	}
	if !strings.Contains(err.Error(), "list remote lint boom") {
		t.Fatalf("err = %q, want list remote lint boom", err)
	}
}

func TestVersionsGoLocal_ShowsInstalledVersions(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.24.3", "1.25.1"}, nil
			},
		},
	}

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

	application := &app.App{
		Resolver: resolve.New(nil),
		GoRemoteProvider: stubGoRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.25.2", "1.25.1"}, nil
			},
		},
	}

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

	application := &app.App{
		Resolver: resolve.New(nil),
		GoRemoteProvider: stubGoRemoteProvider{
			listRemoteGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.25.2", "1.25.1", "1.24.9"}, nil
			},
		},
	}

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

	application := &app.App{
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
	}

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

	application := &app.App{
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
	}

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

func TestVersionsGo_RejectsBothLocalAndRemote(t *testing.T) {
	t.Parallel()

	application := &app.App{}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "go", "--local", "--remote")
	if err == nil {
		t.Fatal("expected an error when both --local and --remote are set")
	}
	if !strings.Contains(err.Error(), "--local and --remote are mutually exclusive") {
		t.Fatalf("err = %q, want mutually exclusive error", err)
	}
}

func TestVersionsGo_RejectsNeitherLocalNorRemote(t *testing.T) {
	t.Parallel()

	application := &app.App{}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "go")
	if err == nil {
		t.Fatal("expected an error when neither --local nor --remote is set")
	}
	if !strings.Contains(err.Error(), "provide --local or --remote") {
		t.Fatalf("err = %q, want provide flag error", err)
	}
}

func TestVersionsGoLocal_RejectsNilGoStore(t *testing.T) {
	t.Parallel()

	application := &app.App{GoStore: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "go", "--local")
	if err == nil {
		t.Fatal("expected an error when GoStore is nil")
	}
	if !strings.Contains(err.Error(), "local Go version store is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestVersionsGoRemote_RejectsNilGoRemoteProvider(t *testing.T) {
	t.Parallel()

	application := &app.App{GoRemoteProvider: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "go", "--remote")
	if err == nil {
		t.Fatal("expected an error when GoRemoteProvider is nil")
	}
	if !strings.Contains(err.Error(), "remote Go version provider is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestVersionsLint_RejectsBothLocalAndRemote(t *testing.T) {
	t.Parallel()

	application := &app.App{}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "golangci-lint", "--local", "--remote")
	if err == nil {
		t.Fatal("expected an error when both --local and --remote are set")
	}
	if !strings.Contains(err.Error(), "--local and --remote are mutually exclusive") {
		t.Fatalf("err = %q, want mutually exclusive error", err)
	}
}

func TestVersionsLint_RejectsNeitherLocalNorRemote(t *testing.T) {
	t.Parallel()

	application := &app.App{}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "golangci-lint")
	if err == nil {
		t.Fatal("expected an error when neither --local nor --remote is set")
	}
	if !strings.Contains(err.Error(), "provide --local or --remote") {
		t.Fatalf("err = %q, want provide flag error", err)
	}
}

func TestVersionsLintLocal_RejectsNilLintStore(t *testing.T) {
	t.Parallel()

	application := &app.App{LintStore: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "golangci-lint", "--local")
	if err == nil {
		t.Fatal("expected an error when LintStore is nil")
	}
	if !strings.Contains(err.Error(), "local golangci-lint version store is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestVersionsLintRemote_RejectsNilLintRemoteProvider(t *testing.T) {
	t.Parallel()

	application := &app.App{LintRemoteProvider: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "golangci-lint", "--remote", "--go", "1.25.0")
	if err == nil {
		t.Fatal("expected an error when LintRemoteProvider is nil")
	}
	if !strings.Contains(err.Error(), "remote golangci-lint version provider is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestVersionsLintRemote_RejectsMissingGoVersion(t *testing.T) {
	t.Parallel()

	application := &app.App{
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return nil, nil
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "versions", "golangci-lint", "--remote")
	if err == nil {
		t.Fatal("expected an error when no --go flag and no resolver")
	}
	if !strings.Contains(err.Error(), "provide --go or run inside a repo") {
		t.Fatalf("err = %q, want provide --go error", err)
	}
}

func TestVersionsGolangCILintLocal_ShowsInstalledVersions(t *testing.T) {
	t.Parallel()

	application := &app.App{
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"v2.11.2", "v2.11.1"}, nil
			},
		},
	}

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
