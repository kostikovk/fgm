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

func TestPinGoCommand(t *testing.T) {
	t.Parallel()

	readConfig := func(t *testing.T, dir string) string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(dir, ".fgm.toml"))
		if err != nil {
			t.Fatalf("read .fgm.toml: %v", err)
		}
		return string(content)
	}

	t.Run("RunE", func(t *testing.T) {
		t.Parallel()

		t.Run("should keep the golangci-lint pin when go latest-patch is pinned", func(t *testing.T) {
			t.Parallel()

			// Arrange
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".fgm.toml"), []byte("[toolchain]\ngolangci_lint = \"v2.9.0\"\n"), 0o644); err != nil {
				t.Fatalf("write .fgm.toml: %v", err)
			}

			// Act
			_, _, err := testutil.ExecuteCommand(t, NewRootCmd(&app.App{}), "pin", "go", "latest-patch", "--chdir", dir)

			// Assert
			if err != nil {
				t.Fatalf("pin go: %v", err)
			}
			content := readConfig(t, dir)
			if !strings.Contains(content, "go = 'latest-patch'") || !strings.Contains(content, "golangci_lint = 'v2.9.0'") {
				t.Errorf(".fgm.toml = %q; want both keys", content)
			}
		})

		t.Run("should store the version without the go prefix when an explicit version is pinned", func(t *testing.T) {
			t.Parallel()

			// Arrange
			dir := t.TempDir()

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(&app.App{}), "pin", "go", "go1.26.8", "--chdir", dir)

			// Assert
			if err != nil {
				t.Fatalf("pin go: %v", err)
			}
			if content := readConfig(t, dir); !strings.Contains(content, "go = '1.26.8'") {
				t.Errorf(".fgm.toml = %q; want go = '1.26.8'", content)
			}
			if !strings.Contains(stdout, "Pinned Go 1.26.8") {
				t.Errorf("stdout = %q; want a pinned message", stdout)
			}
		})

		t.Run("should return an error when the policy is neither a version nor latest-patch", func(t *testing.T) {
			t.Parallel()

			// Arrange
			dir := t.TempDir()

			// Act
			_, _, err := testutil.ExecuteCommand(t, NewRootCmd(&app.App{}), "pin", "go", "newest", "--chdir", dir)

			// Assert
			if err == nil {
				t.Fatal("pin go returned no error; want an invalid policy error")
			}
		})
	})
}

func TestPinLintCommandKeepsGoPolicy(t *testing.T) {
	t.Parallel()

	t.Run("RunE", func(t *testing.T) {
		t.Parallel()

		t.Run("should keep the go policy when golangci-lint is pinned", func(t *testing.T) {
			t.Parallel()

			// Arrange
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".fgm.toml"), []byte("[toolchain]\ngo = \"latest-patch\"\n"), 0o644); err != nil {
				t.Fatalf("write .fgm.toml: %v", err)
			}

			// Act
			_, _, err := testutil.ExecuteCommand(t, NewRootCmd(&app.App{}), "pin", "golangci-lint", "v2.9.0", "--chdir", dir)

			// Assert
			if err != nil {
				t.Fatalf("pin golangci-lint: %v", err)
			}
			content, err := os.ReadFile(filepath.Join(dir, ".fgm.toml"))
			if err != nil {
				t.Fatalf("read .fgm.toml: %v", err)
			}
			if !strings.Contains(string(content), "go = 'latest-patch'") {
				t.Errorf(".fgm.toml = %q; want the go policy kept", content)
			}
		})
	})
}
