package lintlocal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestStoreDeleteLintVersion_RemovesManagedVersion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	versionDir := filepath.Join(root, "golangci-lint", "v2.11.2")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "golangci-lint"), []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store := New(root)
	removedPath, err := store.DeleteLintVersion(context.Background(), "v2.11.2")
	if err != nil {
		t.Fatalf("DeleteLintVersion: %v", err)
	}

	if removedPath != versionDir {
		t.Fatalf("removedPath = %q, want %q", removedPath, versionDir)
	}
	if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, err=%v", versionDir, err)
	}
}

func TestStoreRegisterExistingLintInstallation_CreatesSymlinkedManagedVersion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	externalDir := t.TempDir()
	externalBinary := filepath.Join(externalDir, "golangci-lint")
	if err := os.WriteFile(externalBinary, []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile external: %v", err)
	}

	store := New(root)
	installPath, err := store.RegisterExistingLintInstallation("v2.11.2", externalBinary)
	if err != nil {
		t.Fatalf("RegisterExistingLintInstallation: %v", err)
	}

	wantPath := filepath.Join(root, "golangci-lint", "v2.11.2")
	if installPath != wantPath {
		t.Fatalf("installPath = %q, want %q", installPath, wantPath)
	}

	info, err := os.Lstat(filepath.Join(wantPath, "golangci-lint"))
	if err != nil {
		t.Fatalf("Lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("managed binary mode = %v, want symlink", info.Mode())
	}
}

func TestStoreLintBinaryPath_ReturnsBinaryPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	versionDir := filepath.Join(root, "golangci-lint", "v2.11.2")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	binaryPath := filepath.Join(versionDir, "golangci-lint")
	if err := os.WriteFile(binaryPath, []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store := New(root)
	got, err := store.LintBinaryPath(context.Background(), "v2.11.2")
	if err != nil {
		t.Fatalf("LintBinaryPath: %v", err)
	}
	if got != binaryPath {
		t.Fatalf("LintBinaryPath = %q, want %q", got, binaryPath)
	}
}

func TestStoreLintBinaryPath_ErrorForMissingVersion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store := New(root)

	_, err := store.LintBinaryPath(context.Background(), "v9.9.9")
	if err == nil {
		t.Fatalf("expected error for missing version, got nil")
	}
	if got := err.Error(); !strings.Contains(got, "not installed") {
		t.Fatalf("error = %q, want it to contain %q", got, "not installed")
	}
}

func TestStoreDeleteLintVersion_ErrorForNonExistentVersion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store := New(root)

	_, err := store.DeleteLintVersion(context.Background(), "v9.9.9")
	if err == nil {
		t.Fatalf("expected error for non-existent version, got nil")
	}
	if got := err.Error(); !strings.Contains(got, "not managed by FGM") {
		t.Fatalf("error = %q, want it to contain %q", got, "not managed by FGM")
	}
}

func TestStoreDeleteLintVersion_ErrorForRegularFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	lintDir := filepath.Join(root, "golangci-lint")
	if err := os.MkdirAll(lintDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Create a regular file where a version directory is expected.
	filePath := filepath.Join(lintDir, "v2.11.2")
	if err := os.WriteFile(filePath, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store := New(root)
	_, err := store.DeleteLintVersion(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatalf("expected error for file instead of dir, got nil")
	}
	if got := err.Error(); !strings.Contains(got, "invalid") {
		t.Fatalf("error = %q, want it to contain %q", got, "invalid")
	}
}

func TestStoreListLocalLintVersions_NilWhenRootMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Do not create the golangci-lint/ directory.
	store := New(root)
	versions, err := store.ListLocalLintVersions(context.Background())
	if err != nil {
		t.Fatalf("ListLocalLintVersions: %v", err)
	}
	if versions != nil {
		t.Fatalf("versions = %v, want nil", versions)
	}
}

func TestStoreListLocalLintVersions_SkipsNonDirectoryEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	lintDir := filepath.Join(root, "golangci-lint")
	if err := os.MkdirAll(lintDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Create a regular file inside the golangci-lint directory.
	if err := os.WriteFile(filepath.Join(lintDir, "stray-file"), []byte("junk"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store := New(root)
	versions, err := store.ListLocalLintVersions(context.Background())
	if err != nil {
		t.Fatalf("ListLocalLintVersions: %v", err)
	}
	if len(versions) != 0 {
		t.Fatalf("len(versions) = %d, want 0", len(versions))
	}
}

func TestStoreListLocalLintVersions_SkipsVersionsWithoutBinary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Create a version directory without a binary inside.
	versionDir := filepath.Join(root, "golangci-lint", "v2.11.2")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	store := New(root)
	versions, err := store.ListLocalLintVersions(context.Background())
	if err != nil {
		t.Fatalf("ListLocalLintVersions: %v", err)
	}
	if len(versions) != 0 {
		t.Fatalf("len(versions) = %d, want 0", len(versions))
	}
}

func TestStoreRegisterExistingLintInstallation_IdempotentIfExists(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	externalDir := t.TempDir()
	externalBinary := filepath.Join(externalDir, "golangci-lint")
	if err := os.WriteFile(externalBinary, []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile external: %v", err)
	}

	store := New(root)

	firstPath, err := store.RegisterExistingLintInstallation("v2.11.2", externalBinary)
	if err != nil {
		t.Fatalf("first RegisterExistingLintInstallation: %v", err)
	}

	secondPath, err := store.RegisterExistingLintInstallation("v2.11.2", externalBinary)
	if err != nil {
		t.Fatalf("second RegisterExistingLintInstallation: %v", err)
	}

	if firstPath != secondPath {
		t.Fatalf("idempotent paths differ: first=%q, second=%q", firstPath, secondPath)
	}
}

func TestStoreListLocalLintVersions_ReadDirErrorNotIsNotExist(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	lintDir := filepath.Join(root, "golangci-lint")
	// Create a regular file where a directory is expected so ReadDir returns
	// an error that is NOT os.IsNotExist.
	if err := os.WriteFile(lintDir, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store := New(root)
	_, err := store.ListLocalLintVersions(context.Background())
	if err == nil {
		t.Fatal("expected an error for ReadDir on a regular file")
	}
	if !strings.Contains(err.Error(), "read managed golangci-lint versions") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "read managed golangci-lint versions")
	}
}

func TestStoreDeleteLintVersion_StatErrorNotIsNotExist(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	lintDir := filepath.Join(root, "golangci-lint")
	if err := os.MkdirAll(lintDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Make the parent directory unreadable so os.Stat returns a permission error.
	if err := os.Chmod(lintDir, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(lintDir, 0o755)
	})

	store := New(root)
	_, err := store.DeleteLintVersion(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatal("expected an error for stat with permission denied")
	}
	if !strings.Contains(err.Error(), "stat managed golangci-lint version") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "stat managed golangci-lint version")
	}
}

func TestStoreDeleteLintVersion_RemoveAllError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	lintDir := filepath.Join(root, "golangci-lint")
	versionDir := filepath.Join(lintDir, "v2.11.2")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Put a file inside the version directory.
	if err := os.WriteFile(filepath.Join(versionDir, "golangci-lint"), []byte("bin"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Make the parent directory read-only to prevent removal of children.
	if err := os.Chmod(lintDir, 0o555); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(lintDir, 0o755)
	})

	store := New(root)
	_, err := store.DeleteLintVersion(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatal("expected an error for RemoveAll with read-only parent")
	}
	if !strings.Contains(err.Error(), "remove managed golangci-lint version") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "remove managed golangci-lint version")
	}
}

func TestStoreRegisterExisting_MkdirAllError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Create a regular file where the golangci-lint directory would go,
	// so MkdirAll fails.
	lintDir := filepath.Join(root, "golangci-lint")
	if err := os.WriteFile(lintDir, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store := New(root)
	_, err := store.RegisterExistingLintInstallation("v2.11.2", "/some/binary")
	if err == nil {
		t.Fatal("expected an error for MkdirAll when path is a file")
	}
	if !strings.Contains(err.Error(), "create managed golangci-lint root") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "create managed golangci-lint root")
	}
}

func TestStoreRegisterExisting_SymlinkError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	lintDir := filepath.Join(root, "golangci-lint")
	versionDir := filepath.Join(lintDir, "v2.11.2")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Make the version directory read-only so os.Symlink fails with permission denied.
	// The binary does not exist, so the os.Stat early-return won't fire.
	if err := os.Chmod(versionDir, 0o555); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(versionDir, 0o755)
	})

	store := New(root)
	_, err := store.RegisterExistingLintInstallation("v2.11.2", "/some/external/binary")
	if err == nil {
		t.Fatal("expected an error for Symlink when version directory is read-only")
	}
	if !strings.Contains(err.Error(), "symlink existing golangci-lint installation") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "symlink existing golangci-lint installation")
	}
}
