package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/testutil"
)

func TestPinLintCommand_CreatesRepoConfigAtNearestGoModRoot(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	workDir := filepath.Join(rootDir, "services", "api")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	root := NewRootCmd(app.New(app.Config{}))
	stdout, stderr, err := testutil.ExecuteCommand(
		t,
		root,
		"pin",
		"golangci-lint",
		"v2.10.1",
		"--chdir",
		workDir,
	)
	if err != nil {
		t.Fatalf("execute pin: %v\nstderr:\n%s", err, stderr)
	}

	content, err := os.ReadFile(filepath.Join(rootDir, ".fgm.toml"))
	if err != nil {
		t.Fatalf("read .fgm.toml: %v", err)
	}

	if !strings.Contains(string(content), "golangci_lint") || !strings.Contains(string(content), "v2.10.1") {
		t.Fatalf(".fgm.toml = %q, want pinned lint", string(content))
	}
	if !strings.Contains(stdout, "Pinned golangci-lint v2.10.1") {
		t.Fatalf("stdout = %q, want success output", stdout)
	}
}

func TestPinLintCommand_UpdatesExistingRepoConfigToAuto(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	workDir := filepath.Join(rootDir, "services", "api")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workDir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(rootDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.11.2\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	root := NewRootCmd(app.New(app.Config{}))
	_, stderr, err := testutil.ExecuteCommand(
		t,
		root,
		"pin",
		"golangci-lint",
		"auto",
		"--chdir",
		workDir,
	)
	if err != nil {
		t.Fatalf("execute pin: %v\nstderr:\n%s", err, stderr)
	}

	content, err := os.ReadFile(filepath.Join(rootDir, ".fgm.toml"))
	if err != nil {
		t.Fatalf("read .fgm.toml: %v", err)
	}

	if !strings.Contains(string(content), "golangci_lint") || !strings.Contains(string(content), "auto") {
		t.Fatalf(".fgm.toml = %q, want auto lint pin", string(content))
	}
}
