package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/currenttoolchain"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
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
