package fgmconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNearest_LoadsNearestRepoConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	content := "[toolchain]\ngolangci_lint = \"v2.11.2\"\n"
	if err := os.WriteFile(filepath.Join(root, fileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	result, found, err := LoadNearest(workDir)
	if err != nil {
		t.Fatalf("LoadNearest: %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if result.File.Toolchain.GolangCILint != "v2.11.2" {
		t.Fatalf("GolangCILint = %q, want %q", result.File.Toolchain.GolangCILint, "v2.11.2")
	}
}

func TestLoadNearest_ReturnsNotFoundWhenNoConfigExists(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	_, found, err := LoadNearest(workDir)
	if err != nil {
		t.Fatalf("LoadNearest: %v", err)
	}
	if found {
		t.Fatal("found = true, want false")
	}
}

func TestSaveNearest_CreatesConfigAtNearestGoModRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	path, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.10.1"},
	})
	if err != nil {
		t.Fatalf("SaveNearest: %v", err)
	}

	wantPath := filepath.Join(root, fileName)
	if path != wantPath {
		t.Fatalf("path = %q, want %q", path, wantPath)
	}

	result, found, err := LoadNearest(workDir)
	if err != nil {
		t.Fatalf("LoadNearest: %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if result.File.Toolchain.GolangCILint != "v2.10.1" {
		t.Fatalf("GolangCILint = %q, want %q", result.File.Toolchain.GolangCILint, "v2.10.1")
	}
}

func TestSaveNearest_UpdatesExistingNearestConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, fileName),
		[]byte("[toolchain]\ngolangci_lint = \"v2.11.2\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write config: %v", err)
	}

	path, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "auto"},
	})
	if err != nil {
		t.Fatalf("SaveNearest: %v", err)
	}

	if path != filepath.Join(root, fileName) {
		t.Fatalf("path = %q, want existing config path", path)
	}

	result, found, err := LoadNearest(workDir)
	if err != nil {
		t.Fatalf("LoadNearest: %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if result.File.Toolchain.GolangCILint != "auto" {
		t.Fatalf("GolangCILint = %q, want %q", result.File.Toolchain.GolangCILint, "auto")
	}
}
