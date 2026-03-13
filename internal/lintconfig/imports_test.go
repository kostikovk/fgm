package lintconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindModulePath_Found(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	goMod := "module github.com/example/myproject\n\ngo 1.26.1\n"
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	got := FindModulePath(workDir)
	if got != "github.com/example/myproject" {
		t.Fatalf("FindModulePath = %q, want %q", got, "github.com/example/myproject")
	}
}

func TestFindModulePath_WalksUp(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	subDir := filepath.Join(root, "cmd", "server")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	goMod := "module github.com/example/nested\n\ngo 1.26.1\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	got := FindModulePath(subDir)
	if got != "github.com/example/nested" {
		t.Fatalf("FindModulePath = %q, want %q", got, "github.com/example/nested")
	}
}

func TestFindModulePath_NotFound(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	got := FindModulePath(workDir)
	if got != "" {
		t.Fatalf("FindModulePath = %q, want empty", got)
	}
}
