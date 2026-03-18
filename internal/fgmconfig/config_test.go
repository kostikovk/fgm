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

func TestSaveNearest_CreatesInWorkDirWhenNoModExists(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()

	path, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.12.0"},
	})
	if err != nil {
		t.Fatalf("SaveNearest: %v", err)
	}

	wantPath := filepath.Join(workDir, fileName)
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
	if result.File.Toolchain.GolangCILint != "v2.12.0" {
		t.Fatalf("GolangCILint = %q, want %q", result.File.Toolchain.GolangCILint, "v2.12.0")
	}
}

func TestSaveNearest_PrefersGoWork(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	subDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "go.work"), []byte("go 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	path, err := SaveNearest(subDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.13.0"},
	})
	if err != nil {
		t.Fatalf("SaveNearest: %v", err)
	}

	wantPath := filepath.Join(root, fileName)
	if path != wantPath {
		t.Fatalf("path = %q, want %q (next to go.work)", path, wantPath)
	}
}

func TestSaveNearest_ResolveWritePathError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "sub")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	// Create a symlink loop for go.work so findNearestNamedFile errors.
	goWorkLink := filepath.Join(root, "go.work")
	if err := os.Symlink(goWorkLink, goWorkLink); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.10.0"},
	})
	if err == nil {
		t.Fatal("expected error when resolveWritePath fails, got nil")
	}
}

func TestSaveNearest_MkdirAllError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// workDir does not exist; resolveWritePath will fall back to
	// filepath.Join(workDir, ".fgm.toml") since no .fgm.toml, go.work,
	// or go.mod is found walking upward. Then os.MkdirAll needs to
	// create the workDir directory, which fails because root is read-only.
	workDir := filepath.Join(root, "nonexistent")
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })

	_, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.10.0"},
	})
	if err == nil {
		t.Fatal("expected error when os.MkdirAll fails, got nil")
	}
}

func TestSaveNearest_WriteFileError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "project")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	// Make the target directory read-only so os.WriteFile fails.
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })

	_, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.10.0"},
	})
	if err == nil {
		t.Fatal("expected error when os.WriteFile fails, got nil")
	}
}

func TestLoadNearest_ReadFileError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	fgmPath := filepath.Join(workDir, fileName)
	if err := os.WriteFile(fgmPath, []byte("[toolchain]\ngolangci_lint = \"v2.11.2\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	// Make the file unreadable so os.ReadFile fails after findNearest succeeds.
	if err := os.Chmod(fgmPath, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(fgmPath, 0o644) })

	_, _, err := LoadNearest(workDir)
	if err == nil {
		t.Fatal("expected error when os.ReadFile fails, got nil")
	}
}

func TestLoadNearest_FindNearestStatError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "sub")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create a symlink loop for .fgm.toml so os.Stat returns a non-NotExist error.
	fgmLink := filepath.Join(root, fileName)
	if err := os.Symlink(fgmLink, fgmLink); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, _, err := LoadNearest(workDir)
	if err == nil {
		t.Fatal("expected error when findNearest hits stat error, got nil")
	}
}

func TestResolveWritePath_GoWorkStatError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "sub")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// No .fgm.toml exists, so findNearest returns found=false.
	// Create a symlink loop for go.work so findNearestNamedFile errors.
	goWorkLink := filepath.Join(root, "go.work")
	if err := os.Symlink(goWorkLink, goWorkLink); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.10.0"},
	})
	if err == nil {
		t.Fatal("expected error when go.work stat fails in resolveWritePath, got nil")
	}
}

func TestResolveWritePath_GoModStatError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "sub")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// No .fgm.toml exists, no go.work exists.
	// Create a symlink loop for go.mod so findNearestNamedFile errors.
	goModLink := filepath.Join(root, "go.mod")
	if err := os.Symlink(goModLink, goModLink); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.10.0"},
	})
	if err == nil {
		t.Fatal("expected error when go.mod stat fails in resolveWritePath, got nil")
	}
}

func TestFindNearestNamedFile_StatError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "sub")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create a symlink loop for a named file so os.Stat fails with non-NotExist error.
	targetLink := filepath.Join(root, "go.work")
	if err := os.Symlink(targetLink, targetLink); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	// findNearestNamedFile is private, but we can exercise it through resolveWritePath
	// via SaveNearest. With no .fgm.toml, resolveWritePath calls findNearestNamedFile
	// for "go.work" which encounters the symlink loop.
	_, err := SaveNearest(workDir, File{
		Toolchain: ToolchainConfig{GolangCILint: "v2.10.0"},
	})
	if err == nil {
		t.Fatal("expected error when findNearestNamedFile hits stat error, got nil")
	}
}

func TestResolvePinnedLint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		wantVersion string
		wantFound   bool
		wantErr     bool
	}{
		{
			name: "no .fgm.toml found",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
		},
		{
			name: "golangci_lint pinned to explicit version",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				content := "[toolchain]\ngolangci_lint = \"v2.10.1\"\n"
				if err := os.WriteFile(filepath.Join(dir, ".fgm.toml"), []byte(content), 0o644); err != nil {
					t.Fatalf("write .fgm.toml: %v", err)
				}
				return dir
			},
			wantVersion: "v2.10.1",
			wantFound:   true,
		},
		{
			name: "golangci_lint set to auto",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				content := "[toolchain]\ngolangci_lint = \"auto\"\n"
				if err := os.WriteFile(filepath.Join(dir, ".fgm.toml"), []byte(content), 0o644); err != nil {
					t.Fatalf("write .fgm.toml: %v", err)
				}
				return dir
			},
		},
		{
			name: "golangci_lint set to empty string",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				content := "[toolchain]\ngolangci_lint = \"\"\n"
				if err := os.WriteFile(filepath.Join(dir, ".fgm.toml"), []byte(content), 0o644); err != nil {
					t.Fatalf("write .fgm.toml: %v", err)
				}
				return dir
			},
		},
		{
			name: "fgm.toml in parent dir",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				workDir := filepath.Join(root, "sub", "project")
				if err := os.MkdirAll(workDir, 0o755); err != nil {
					t.Fatalf("mkdir workdir: %v", err)
				}
				content := "[toolchain]\ngolangci_lint = \"v2.10.1\"\n"
				if err := os.WriteFile(filepath.Join(root, ".fgm.toml"), []byte(content), 0o644); err != nil {
					t.Fatalf("write .fgm.toml: %v", err)
				}
				return workDir
			},
			wantVersion: "v2.10.1",
			wantFound:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			workDir := tc.setup(t)

			gotVersion, gotFound, err := ResolvePinnedLint(workDir)

			if (err != nil) != tc.wantErr {
				t.Fatalf("ResolvePinnedLint() error = %v, wantErr %v", err, tc.wantErr)
			}
			if gotFound != tc.wantFound {
				t.Errorf("ResolvePinnedLint() found = %v, want %v", gotFound, tc.wantFound)
			}
			if gotVersion != tc.wantVersion {
				t.Errorf("ResolvePinnedLint() version = %q, want %q", gotVersion, tc.wantVersion)
			}
		})
	}
}

func TestLoadNearest_ErrorForMalformedToml(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, fileName), []byte("this is not valid toml [[["), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, _, err := LoadNearest(workDir)
	if err == nil {
		t.Fatal("expected an error for malformed TOML")
	}
}
