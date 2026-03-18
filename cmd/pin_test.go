package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/testutil"
)

func TestPinLintCommand_RejectsEmptyVersion(t *testing.T) {
	t.Parallel()

	root := NewRootCmd(&app.App{})
	_, _, err := testutil.ExecuteCommand(t, root, "pin", "golangci-lint", "  ")
	if err == nil {
		t.Fatal("expected an error when version is empty/whitespace")
	}
	if !strings.Contains(err.Error(), "version must not be empty") {
		t.Fatalf("err = %q, want version must not be empty", err)
	}
}

func TestPinLintCommand_SaveNearestErrorIsPropagated(t *testing.T) {
	t.Parallel()

	// Use a path under /dev/null so MkdirAll will fail.
	nonExistent := "/dev/null/impossible"
	root := NewRootCmd(&app.App{})
	_, _, err := testutil.ExecuteCommand(t, root, "pin", "golangci-lint", "v2.11.2", "--chdir", nonExistent)
	if err == nil {
		t.Fatal("expected an error when SaveNearest fails")
	}
	// The exact error message depends on OS, but it should not be nil.
}

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

	root := NewRootCmd(&app.App{})
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

	root := NewRootCmd(&app.App{})
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
