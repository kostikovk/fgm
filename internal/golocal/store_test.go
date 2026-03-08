package golocal

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStoreListLocalGoVersions_ReturnsManagedVersionsAndSystemGo(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script PATH test is unix-only")
	}

	root := t.TempDir()
	pathDir := t.TempDir()
	systemGoBinary := filepath.Join(pathDir, "go")
	script := "#!/bin/sh\nprintf 'go version go1.25.3 darwin/arm64\\n'\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	createManagedGoVersion(t, root, "1.24.3")
	createManagedGoVersion(t, root, "1.25.1")

	store := New(root, pathDir)
	versions, err := store.ListLocalGoVersions(context.Background())
	if err != nil {
		t.Fatalf("ListLocalGoVersions: %v", err)
	}

	if len(versions) != 3 {
		t.Fatalf("len(versions) = %d, want %d", len(versions), 3)
	}
	if versions[0] != "1.25.3" || versions[1] != "1.25.1" || versions[2] != "1.24.3" {
		t.Fatalf("versions = %v, want [1.25.3 1.25.1 1.24.3]", versions)
	}
}

func TestStoreSetAndGetGlobalGoVersion(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "")
	if err := store.SetGlobalGoVersion(context.Background(), "1.25.7"); err != nil {
		t.Fatalf("SetGlobalGoVersion: %v", err)
	}

	version, ok, err := store.GlobalGoVersion(context.Background())
	if err != nil {
		t.Fatalf("GlobalGoVersion: %v", err)
	}
	if !ok {
		t.Fatal("expected a persisted global version")
	}
	if version != "1.25.7" {
		t.Fatalf("version = %q, want %q", version, "1.25.7")
	}
}

func TestStoreGoBinaryPath_ReturnsManagedBinary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createManagedGoVersion(t, root, "1.25.7")

	store := New(root, "")
	binaryPath, err := store.GoBinaryPath(context.Background(), "1.25.7")
	if err != nil {
		t.Fatalf("GoBinaryPath: %v", err)
	}

	if !strings.Contains(binaryPath, filepath.Join("go", "1.25.7", "bin")) {
		t.Fatalf("binaryPath = %q, want it to contain managed path", binaryPath)
	}
}

func TestStoreEnsureShims_WritesGoShim(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell shim generation test is unix-only")
	}

	root := t.TempDir()
	store := New(root, "")
	if err := store.EnsureShims(); err != nil {
		t.Fatalf("EnsureShims: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(store.ShimDir(), "go"))
	if err != nil {
		t.Fatalf("read go shim: %v", err)
	}

	if !strings.Contains(string(content), "exec fgm __shim go") {
		t.Fatalf("shim content = %q, want it to contain %q", string(content), "exec fgm __shim go")
	}
}

func TestStoreDeleteGoVersion_RemovesManagedVersion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createManagedGoVersion(t, root, "1.25.7")

	store := New(root, "")
	removedPath, err := store.DeleteGoVersion(context.Background(), "1.25.7")
	if err != nil {
		t.Fatalf("DeleteGoVersion: %v", err)
	}

	if removedPath != filepath.Join(root, "go", "1.25.7") {
		t.Fatalf("removedPath = %q, want %q", removedPath, filepath.Join(root, "go", "1.25.7"))
	}
	if _, err := os.Stat(removedPath); !os.IsNotExist(err) {
		t.Fatalf("expected managed version to be removed, stat err = %v", err)
	}
}

func TestStoreDeleteGoVersion_RejectsGlobalVersion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createManagedGoVersion(t, root, "1.25.7")

	store := New(root, "")
	if err := store.SetGlobalGoVersion(context.Background(), "1.25.7"); err != nil {
		t.Fatalf("SetGlobalGoVersion: %v", err)
	}

	_, err := store.DeleteGoVersion(context.Background(), "1.25.7")
	if err == nil {
		t.Fatal("expected an error when deleting the active global version")
	}
	if !strings.Contains(err.Error(), "current global version") {
		t.Fatalf("err = %q, want it to contain %q", err, "current global version")
	}
}

func createManagedGoVersion(t *testing.T, root string, version string) {
	t.Helper()

	binaryName := "go"
	if runtime.GOOS == "windows" {
		binaryName = "go.exe"
	}

	binaryPath := filepath.Join(root, "go", version, "bin", binaryName)
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		t.Fatalf("mkdir managed go dir: %v", err)
	}

	contents := []byte("#!/bin/sh\n")
	perm := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		contents = []byte("")
		perm = 0o644
	}
	if err := os.WriteFile(binaryPath, contents, perm); err != nil {
		t.Fatalf("write managed go binary: %v", err)
	}
}
