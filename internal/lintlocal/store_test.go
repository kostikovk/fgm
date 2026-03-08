package lintlocal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreListLocalLintVersions_ReturnsInstalledVersions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	versionDir := filepath.Join(root, "golangci-lint", "v2.11.2")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "golangci-lint"), []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	olderDir := filepath.Join(root, "golangci-lint", "v2.11.1")
	if err := os.MkdirAll(olderDir, 0o755); err != nil {
		t.Fatalf("MkdirAll older: %v", err)
	}
	if err := os.WriteFile(filepath.Join(olderDir, "golangci-lint"), []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile older: %v", err)
	}

	store := New(root)
	versions, err := store.ListLocalLintVersions(context.Background())
	if err != nil {
		t.Fatalf("ListLocalLintVersions: %v", err)
	}

	if len(versions) != 2 {
		t.Fatalf("len(versions) = %d, want %d", len(versions), 2)
	}
	if versions[0] != "v2.11.2" {
		t.Fatalf("versions[0] = %q, want %q", versions[0], "v2.11.2")
	}
	if versions[1] != "v2.11.1" {
		t.Fatalf("versions[1] = %q, want %q", versions[1], "v2.11.1")
	}
}
