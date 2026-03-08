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

func TestStoreHasGoVersion_TrueForManagedVersion(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createManagedGoVersion(t, root, "1.25.7")

	store := New(root, "")
	ok, err := store.HasGoVersion(context.Background(), "1.25.7")
	if err != nil {
		t.Fatalf("HasGoVersion: %v", err)
	}
	if !ok {
		t.Fatal("expected HasGoVersion to return true for a managed version")
	}
}

func TestStoreHasGoVersion_FalseForMissingVersion(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "")
	ok, err := store.HasGoVersion(context.Background(), "1.99.0")
	if err != nil {
		t.Fatalf("HasGoVersion: %v", err)
	}
	if ok {
		t.Fatal("expected HasGoVersion to return false for a non-existent version")
	}
}

func TestStoreGlobalGoVersion_EmptyWhenNotSet(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "")
	version, ok, err := store.GlobalGoVersion(context.Background())
	if err != nil {
		t.Fatalf("GlobalGoVersion: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false when no global version is set, got version=%q", version)
	}
	if version != "" {
		t.Fatalf("expected empty version string, got %q", version)
	}
}

func TestStoreDeleteGoVersion_ErrorForNonExistentVersion(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "")
	_, err := store.DeleteGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatal("expected an error when deleting a non-existent version")
	}
	if !strings.Contains(err.Error(), "not managed by FGM") {
		t.Fatalf("err = %q, want it to contain %q", err, "not managed by FGM")
	}
}

func TestStoreRegisterExistingGoInstallation_CreatesSymlink(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink test is unix-only")
	}

	root := t.TempDir()
	goroot := t.TempDir()

	store := New(root, "")
	installDir, err := store.RegisterExistingGoInstallation("1.25.7", goroot)
	if err != nil {
		t.Fatalf("RegisterExistingGoInstallation: %v", err)
	}

	info, err := os.Lstat(installDir)
	if err != nil {
		t.Fatalf("Lstat registered install dir: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected a symlink at %q, got mode %v", installDir, info.Mode())
	}

	target, err := os.Readlink(installDir)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if target != goroot {
		t.Fatalf("symlink target = %q, want %q", target, goroot)
	}
}

func TestStoreRegisterExistingGoInstallation_IdempotentIfExists(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink test is unix-only")
	}

	root := t.TempDir()
	goroot := t.TempDir()

	store := New(root, "")
	first, err := store.RegisterExistingGoInstallation("1.25.7", goroot)
	if err != nil {
		t.Fatalf("first RegisterExistingGoInstallation: %v", err)
	}

	second, err := store.RegisterExistingGoInstallation("1.25.7", goroot)
	if err != nil {
		t.Fatalf("second RegisterExistingGoInstallation: %v", err)
	}

	if first != second {
		t.Fatalf("expected idempotent result, got %q and %q", first, second)
	}
}

func TestParseGoVersionOutput_ValidOutput(t *testing.T) {
	t.Parallel()

	version, ok := parseGoVersionOutput("go version go1.25.7 darwin/arm64\n")
	if !ok {
		t.Fatal("expected ok=true for valid go version output")
	}
	if version != "1.25.7" {
		t.Fatalf("version = %q, want %q", version, "1.25.7")
	}
}

func TestParseGoVersionOutput_InvalidOutput(t *testing.T) {
	t.Parallel()

	version, ok := parseGoVersionOutput("garbage")
	if ok {
		t.Fatalf("expected ok=false for invalid output, got version=%q", version)
	}
	if version != "" {
		t.Fatalf("expected empty version, got %q", version)
	}
}

func TestParseGoVersionOutput_MissingPrefix(t *testing.T) {
	t.Parallel()

	// The output lacks the leading "go" field, so fields[0] != "go".
	version, ok := parseGoVersionOutput("version go1.25.7 darwin/arm64")
	if ok {
		t.Fatalf("expected ok=false when output is missing the leading 'go' field, got version=%q", version)
	}
	if version != "" {
		t.Fatalf("expected empty version, got %q", version)
	}
}

func TestSortVersions_NumericOrder(t *testing.T) {
	t.Parallel()

	set := map[string]struct{}{
		"1.21": {},
		"1.9":  {},
		"1.3":  {},
	}

	sorted := sortVersions(set)

	if len(sorted) != 3 {
		t.Fatalf("len(sorted) = %d, want 3", len(sorted))
	}
	if sorted[0] != "1.21" || sorted[1] != "1.9" || sorted[2] != "1.3" {
		t.Fatalf("sorted = %v, want [1.21 1.9 1.3] (numeric descending, not lexicographic)", sorted)
	}
}

func TestDefaultRoot_UsesXdgDataHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/custom/xdg")
	root, err := DefaultRoot()
	if err != nil {
		t.Fatalf("DefaultRoot: %v", err)
	}
	if root != filepath.Join("/custom/xdg", "fgm") {
		t.Fatalf("root = %q, want %q", root, filepath.Join("/custom/xdg", "fgm"))
	}
}

func TestDefaultRoot_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	root, err := DefaultRoot()
	if err != nil {
		t.Fatalf("DefaultRoot: %v", err)
	}
	if !strings.Contains(root, filepath.Join(".local", "share", "fgm")) {
		t.Fatalf("root = %q, want it to contain %q", root, filepath.Join(".local", "share", "fgm"))
	}
}

func TestGoBinaryPath_ReturnsSystemBinary(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script test is unix-only")
	}

	root := t.TempDir()
	pathDir := t.TempDir()
	systemGoBinary := filepath.Join(pathDir, "go")
	script := "#!/bin/sh\nprintf 'go version go1.25.3 darwin/arm64\\n'\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	store := New(root, pathDir)
	binaryPath, err := store.GoBinaryPath(context.Background(), "1.25.3")
	if err != nil {
		t.Fatalf("GoBinaryPath: %v", err)
	}
	if binaryPath != systemGoBinary {
		t.Fatalf("binaryPath = %q, want %q", binaryPath, systemGoBinary)
	}
}

func TestGoBinaryPath_ErrorWhenSystemVersionMismatches(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script test is unix-only")
	}

	root := t.TempDir()
	pathDir := t.TempDir()
	systemGoBinary := filepath.Join(pathDir, "go")
	script := "#!/bin/sh\nprintf 'go version go1.25.3 darwin/arm64\\n'\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	store := New(root, pathDir)
	_, err := store.GoBinaryPath(context.Background(), "1.99.0")
	if err == nil {
		t.Fatal("expected an error when requested version differs from system version")
	}
	if !strings.Contains(err.Error(), "not installed") {
		t.Fatalf("err = %q, want it to contain %q", err, "not installed")
	}
}

func TestGoBinaryPath_ErrorWhenNotInstalled(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "")
	_, err := store.GoBinaryPath(context.Background(), "1.99.0")
	if err == nil {
		t.Fatal("expected an error when version is not installed")
	}
	if !strings.Contains(err.Error(), "not installed") {
		t.Fatalf("err = %q, want it to contain %q", err, "not installed")
	}
}

func TestDeleteGoVersion_ErrorForRegularFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	installDir := filepath.Join(root, "go")
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	filePath := filepath.Join(installDir, "1.25.7")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	store := New(root, "")
	_, err := store.DeleteGoVersion(context.Background(), "1.25.7")
	if err == nil {
		t.Fatal("expected an error when install path is a regular file")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("err = %q, want it to contain %q", err, "invalid")
	}
}

func TestListManagedGoVersions_SkipsNonDirectoryEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	goDir := filepath.Join(root, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create a regular file in the go directory (should be skipped).
	if err := os.WriteFile(filepath.Join(goDir, "somefile"), []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Create a valid managed version directory.
	createManagedGoVersion(t, root, "1.24.3")

	store := New(root, "")
	versions, err := store.ListLocalGoVersions(context.Background())
	if err != nil {
		t.Fatalf("ListLocalGoVersions: %v", err)
	}

	if len(versions) != 1 {
		t.Fatalf("len(versions) = %d, want 1", len(versions))
	}
	if versions[0] != "1.24.3" {
		t.Fatalf("versions[0] = %q, want %q", versions[0], "1.24.3")
	}
}

func TestParseGoVersionOutput_EmptyVersionAfterTrimPrefix(t *testing.T) {
	t.Parallel()

	// "go version go darwin/arm64" - after TrimPrefix("go") the version field is empty.
	version, ok := parseGoVersionOutput("go version go darwin/arm64")
	if ok {
		t.Fatalf("expected ok=false when version is empty after trim, got version=%q", version)
	}
	if version != "" {
		t.Fatalf("expected empty version, got %q", version)
	}
}

func TestFindGoBinary_SkipsEmptyPathEntries(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix PATH test")
	}

	dir := t.TempDir()
	goBinary := filepath.Join(dir, "go")
	if err := os.WriteFile(goBinary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	// PATH with leading empty entry.
	pathEnv := string(filepath.ListSeparator) + dir
	result, ok := findGoBinary(pathEnv)
	if !ok {
		t.Fatal("expected findGoBinary to find the binary despite empty PATH entries")
	}
	if result != goBinary {
		t.Fatalf("result = %q, want %q", result, goBinary)
	}
}

func TestFindGoBinary_SkipsNonExecutableFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix executable-bit test")
	}

	dir := t.TempDir()
	goBinary := filepath.Join(dir, "go")
	// Write a file with no execute permission.
	if err := os.WriteFile(goBinary, []byte("not executable"), 0o644); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	_, ok := findGoBinary(dir)
	if ok {
		t.Fatal("expected findGoBinary to skip a non-executable file")
	}
}

func TestFindGoBinary_SkipsDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create a directory named "go" instead of a file.
	goDir := filepath.Join(dir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, ok := findGoBinary(dir)
	if ok {
		t.Fatal("expected findGoBinary to skip a directory named 'go'")
	}
}

func TestFindGoBinary_ReturnsFalseForEmptyPath(t *testing.T) {
	t.Parallel()

	_, ok := findGoBinary("")
	if ok {
		t.Fatal("expected findGoBinary to return false for empty PATH")
	}
}

func TestSystemGo_ReturnsEmptyWhenNoBinaryOnPath(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "")
	version, binary, err := store.systemGo(context.Background())
	if err != nil {
		t.Fatalf("systemGo: %v", err)
	}
	if version != "" {
		t.Fatalf("version = %q, want empty", version)
	}
	if binary != "" {
		t.Fatalf("binary = %q, want empty", binary)
	}
}

// --- Additional tests for uncovered error paths ---

func TestListLocalGoVersions_ReadDirError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Create "go" as a regular file so ReadDir fails with a non-IsNotExist error.
	goPath := filepath.Join(root, "go")
	if err := os.WriteFile(goPath, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	store := New(root, "")
	_, err := store.ListLocalGoVersions(context.Background())
	if err == nil {
		t.Fatal("expected an error when ReadDir fails")
	}
	if !strings.Contains(err.Error(), "read managed Go versions") {
		t.Fatalf("err = %q, want it to contain %q", err, "read managed Go versions")
	}
}

func TestListLocalGoVersions_SystemGoError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script test is unix-only")
	}

	root := t.TempDir()
	pathDir := t.TempDir()

	// Create a go binary that exits with an error.
	systemGoBinary := filepath.Join(pathDir, "go")
	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	store := New(root, pathDir)
	_, err := store.ListLocalGoVersions(context.Background())
	if err == nil {
		t.Fatal("expected an error when systemGo fails")
	}
	if !strings.Contains(err.Error(), "run go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "run go version")
	}
}

func TestHasGoVersion_SystemGoError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script test is unix-only")
	}

	root := t.TempDir()
	pathDir := t.TempDir()

	// Create a go binary that exits with an error.
	systemGoBinary := filepath.Join(pathDir, "go")
	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	store := New(root, pathDir)
	_, err := store.HasGoVersion(context.Background(), "1.25.0")
	if err == nil {
		t.Fatal("expected an error when systemGo fails")
	}
	if !strings.Contains(err.Error(), "run go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "run go version")
	}
}

func TestGlobalGoVersion_NonNotExistReadError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Create the state directory path as a regular file so that reading
	// the global state file (a path beneath it) fails with a non-IsNotExist error.
	statePath := filepath.Join(root, "state")
	if err := os.WriteFile(statePath, []byte("block"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	store := New(root, "")
	_, _, err := store.GlobalGoVersion(context.Background())
	if err == nil {
		t.Fatal("expected an error when reading global Go version fails with non-NotExist error")
	}
	if !strings.Contains(err.Error(), "read global Go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "read global Go version")
	}
}

func TestGlobalGoVersion_EmptyFileContent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Manually create the state file with only whitespace.
	stateDir := filepath.Join(root, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "global-go-version"), []byte("   \n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	store := New(root, "")
	version, ok, err := store.GlobalGoVersion(context.Background())
	if err != nil {
		t.Fatalf("GlobalGoVersion: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false when file content is only whitespace")
	}
	if version != "" {
		t.Fatalf("expected empty version, got %q", version)
	}
}

func TestSetGlobalGoVersion_MkdirAllError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Block directory creation by making "state" a regular file.
	statePath := filepath.Join(root, "state")
	if err := os.WriteFile(statePath, []byte("block"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	store := New(root, "")
	err := store.SetGlobalGoVersion(context.Background(), "1.25.0")
	if err == nil {
		t.Fatal("expected an error when MkdirAll fails")
	}
	if !strings.Contains(err.Error(), "create FGM state directory") {
		t.Fatalf("err = %q, want it to contain %q", err, "create FGM state directory")
	}
}

func TestSetGlobalGoVersion_WriteFileError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Create the state directory as read-only so WriteFile fails.
	stateDir := filepath.Join(root, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Make the directory read-only so WriteFile fails.
	if err := os.Chmod(stateDir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so t.TempDir cleanup can succeed.
		_ = os.Chmod(stateDir, 0o755)
	})

	store := New(root, "")
	err := store.SetGlobalGoVersion(context.Background(), "1.25.0")
	if err == nil {
		t.Fatal("expected an error when WriteFile fails")
	}
	if !strings.Contains(err.Error(), "write global Go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "write global Go version")
	}
}

func TestDeleteGoVersion_GlobalGoVersionReadError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Make the state path a file so GlobalGoVersion returns a non-NotExist error.
	statePath := filepath.Join(root, "state")
	if err := os.WriteFile(statePath, []byte("block"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	store := New(root, "")
	_, err := store.DeleteGoVersion(context.Background(), "1.25.0")
	if err == nil {
		t.Fatal("expected an error when GlobalGoVersion fails")
	}
	if !strings.Contains(err.Error(), "read global Go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "read global Go version")
	}
}

func TestDeleteGoVersion_StatError_NotIsNotExist(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Create the "go" directory but make it unreadable so Lstat on a child fails
	// with a permission error (not IsNotExist).
	goDir := filepath.Join(root, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(goDir, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(goDir, 0o755)
	})

	store := New(root, "")
	_, err := store.DeleteGoVersion(context.Background(), "1.25.0")
	if err == nil {
		t.Fatal("expected an error when Lstat fails with permission error")
	}
	if !strings.Contains(err.Error(), "stat managed Go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "stat managed Go version")
	}
}

func TestDeleteGoVersion_RemoveAllError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createManagedGoVersion(t, root, "1.25.0")

	// Make the version directory non-removable by removing write permission on parent.
	goDir := filepath.Join(root, "go")
	if err := os.Chmod(goDir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(goDir, 0o755)
	})

	store := New(root, "")
	_, err := store.DeleteGoVersion(context.Background(), "1.25.0")
	if err == nil {
		t.Fatal("expected an error when RemoveAll fails")
	}
	if !strings.Contains(err.Error(), "remove managed Go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "remove managed Go version")
	}
}

func TestGoBinaryPath_SystemGoError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script test is unix-only")
	}

	root := t.TempDir()
	pathDir := t.TempDir()

	// Create a go binary that exits with an error.
	systemGoBinary := filepath.Join(pathDir, "go")
	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	store := New(root, pathDir)
	_, err := store.GoBinaryPath(context.Background(), "1.25.0")
	if err == nil {
		t.Fatal("expected an error when systemGo fails")
	}
	if !strings.Contains(err.Error(), "run go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "run go version")
	}
}

func TestEnsureShims_MkdirAllError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only test")
	}

	root := t.TempDir()

	// Block shim directory creation by placing a file at the shims path.
	shimsPath := filepath.Join(root, "shims")
	if err := os.WriteFile(shimsPath, []byte("block"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	store := New(root, "")
	err := store.EnsureShims()
	if err == nil {
		t.Fatal("expected an error when MkdirAll fails for shim directory")
	}
	if !strings.Contains(err.Error(), "create shim directory") {
		t.Fatalf("err = %q, want it to contain %q", err, "create shim directory")
	}
}

func TestEnsureShims_WriteFileError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only test")
	}

	root := t.TempDir()

	// Create the shims directory as read-only so WriteFile fails.
	shimsDir := filepath.Join(root, "shims")
	if err := os.MkdirAll(shimsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(shimsDir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(shimsDir, 0o755)
	})

	store := New(root, "")
	err := store.EnsureShims()
	if err == nil {
		t.Fatal("expected an error when WriteFile fails for shim script")
	}
	if !strings.Contains(err.Error(), "write go shim") {
		t.Fatalf("err = %q, want it to contain %q", err, "write go shim")
	}
}

func TestListManagedGoVersions_ReadDirNonNotExistError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Create "go" as a regular file so ReadDir returns a non-IsNotExist error.
	goPath := filepath.Join(root, "go")
	if err := os.WriteFile(goPath, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	store := New(root, "")
	_, err := store.listManagedGoVersions()
	if err == nil {
		t.Fatal("expected an error when ReadDir fails with non-NotExist error")
	}
	if !strings.Contains(err.Error(), "read managed Go versions") {
		t.Fatalf("err = %q, want it to contain %q", err, "read managed Go versions")
	}
}

func TestSystemGo_ExecError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script test is unix-only")
	}

	pathDir := t.TempDir()
	systemGoBinary := filepath.Join(pathDir, "go")
	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	store := New(t.TempDir(), pathDir)
	_, _, err := store.systemGo(context.Background())
	if err == nil {
		t.Fatal("expected an error when exec fails")
	}
	if !strings.Contains(err.Error(), "run go version") {
		t.Fatalf("err = %q, want it to contain %q", err, "run go version")
	}
}

func TestSystemGo_VersionParseMismatch(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script test is unix-only")
	}

	pathDir := t.TempDir()
	systemGoBinary := filepath.Join(pathDir, "go")
	// Output that doesn't match "go version goX.Y.Z ..." format.
	script := "#!/bin/sh\nprintf 'something unexpected\\n'\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	store := New(t.TempDir(), pathDir)
	_, _, err := store.systemGo(context.Background())
	if err == nil {
		t.Fatal("expected an error when version output cannot be parsed")
	}
	if !strings.Contains(err.Error(), "parse go version output") {
		t.Fatalf("err = %q, want it to contain %q", err, "parse go version output")
	}
}

func TestListLocalGoVersions_ManagedVersionWithoutBinary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Create a version directory without the go binary inside it.
	versionDir := filepath.Join(root, "go", "1.25.0")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	store := New(root, "")
	versions, err := store.ListLocalGoVersions(context.Background())
	if err != nil {
		t.Fatalf("ListLocalGoVersions: %v", err)
	}

	// The version directory without a binary should be skipped.
	if len(versions) != 0 {
		t.Fatalf("len(versions) = %d, want 0 (version without binary should be skipped)", len(versions))
	}
}

func TestListLocalGoVersions_EmptyWhenNoManagedOrSystemGo(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "")
	versions, err := store.ListLocalGoVersions(context.Background())
	if err != nil {
		t.Fatalf("ListLocalGoVersions: %v", err)
	}
	if len(versions) != 0 {
		t.Fatalf("len(versions) = %d, want 0", len(versions))
	}
}

func TestHasGoVersion_TrueForSystemGo(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script PATH test is unix-only")
	}

	pathDir := t.TempDir()
	systemGoBinary := filepath.Join(pathDir, "go")
	script := "#!/bin/sh\nprintf 'go version go1.25.3 darwin/arm64\\n'\n"
	if err := os.WriteFile(systemGoBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake system go binary: %v", err)
	}

	store := New(t.TempDir(), pathDir)
	ok, err := store.HasGoVersion(context.Background(), "1.25.3")
	if err != nil {
		t.Fatalf("HasGoVersion: %v", err)
	}
	if !ok {
		t.Fatal("expected HasGoVersion to detect system Go")
	}
}

func TestGlobalGoVersion_EmptyStateFile(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir(), "")
	if err := os.MkdirAll(filepath.Dir(store.globalStatePath()), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	if err := os.WriteFile(store.globalStatePath(), []byte("\n"), 0o644); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	version, ok, err := store.GlobalGoVersion(context.Background())
	if err != nil {
		t.Fatalf("GlobalGoVersion: %v", err)
	}
	if ok || version != "" {
		t.Fatalf("GlobalGoVersion = (%q, %v), want empty false", version, ok)
	}
}

func TestEnsureShims_MkdirError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell shim generation test is unix-only")
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "shims"), []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	store := New(root, "")
	err := store.EnsureShims()
	if err == nil {
		t.Fatal("expected EnsureShims mkdir error")
	}
	if !strings.Contains(err.Error(), "create shim directory") {
		t.Fatalf("err = %q, want shim directory error", err)
	}
}

func TestEnsureShims_WriteError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell shim generation test is unix-only")
	}

	root := t.TempDir()
	shimDir := filepath.Join(root, "shims")
	if err := os.MkdirAll(shimDir, 0o755); err != nil {
		t.Fatalf("mkdir shim dir: %v", err)
	}
	if err := os.Mkdir(filepath.Join(shimDir, "go"), 0o755); err != nil {
		t.Fatalf("mkdir blocking dir: %v", err)
	}

	store := New(root, "")
	err := store.EnsureShims()
	if err == nil {
		t.Fatal("expected EnsureShims write error")
	}
	if !strings.Contains(err.Error(), "write go shim") {
		t.Fatalf("err = %q, want write error", err)
	}
}

func TestRegisterExistingGoInstallation_MkdirError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go"), []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	store := New(root, "")
	_, err := store.RegisterExistingGoInstallation("1.25.7", t.TempDir())
	if err == nil {
		t.Fatal("expected register mkdir error")
	}
	if !strings.Contains(err.Error(), "create managed Go root") {
		t.Fatalf("err = %q, want mkdir error", err)
	}
}

func TestRegisterExistingGoInstallation_SymlinkError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink test is unix-only")
	}

	root := t.TempDir()
	goRootDir := filepath.Join(root, "go")
	if err := os.MkdirAll(goRootDir, 0o555); err != nil {
		t.Fatalf("mkdir go dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(goRootDir, 0o755)
	})

	store := New(root, "")
	_, err := store.RegisterExistingGoInstallation("1.25.7", t.TempDir())
	if err == nil {
		t.Fatal("expected register symlink error")
	}
	if !strings.Contains(err.Error(), "symlink existing Go installation") {
		t.Fatalf("err = %q, want symlink error", err)
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
