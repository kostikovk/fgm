package resolve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stubGlobalVersionSource struct {
	globalGoVersionFn func(ctx context.Context) (string, bool, error)
}

func (s stubGlobalVersionSource) GlobalGoVersion(ctx context.Context) (string, bool, error) {
	return s.globalGoVersionFn(ctx)
}

func TestResolverCurrent_ResolvesGoVersionFromNearestGoMod(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	nestedDir := filepath.Join(projectDir, "nested", "deeper")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir nested dir: %v", err)
	}

	goMod := "module example.com/demo\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	selection, err := New(nil).Current(context.Background(), nestedDir)
	if err != nil {
		t.Fatalf("resolve current: %v", err)
	}

	if selection.GoVersion != "1.25.0" {
		t.Fatalf("GoVersion = %q, want %q", selection.GoVersion, "1.25.0")
	}
}

func TestResolverCurrent_PrefersToolchainDirective(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.24.0\n\ntoolchain go1.24.3\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	selection, err := New(nil).Current(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("resolve current: %v", err)
	}

	if selection.GoVersion != "1.24.3" {
		t.Fatalf("GoVersion = %q, want %q", selection.GoVersion, "1.24.3")
	}
}

func TestResolverCurrent_PrefersGoWorkOverNestedGoMod(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspace")
	moduleDir := filepath.Join(workspaceDir, "services", "api")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("mkdir module dir: %v", err)
	}

	goWork := "go 1.24.0\n\ntoolchain go1.25.1\n\nuse ./services/api\n"
	if err := os.WriteFile(filepath.Join(workspaceDir, "go.work"), []byte(goWork), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}

	goMod := "module example.com/api\n\ngo 1.23.0\n"
	if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	selection, err := New(nil).Current(context.Background(), moduleDir)
	if err != nil {
		t.Fatalf("resolve current: %v", err)
	}

	if selection.GoVersion != "1.25.1" {
		t.Fatalf("GoVersion = %q, want %q", selection.GoVersion, "1.25.1")
	}
}

func TestResolverCurrent_FallsBackToGoModWhenGoWorkHasNoVersionMetadata(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspace")
	moduleDir := filepath.Join(workspaceDir, "services", "api")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("mkdir module dir: %v", err)
	}

	goWork := "use ./services/api\n"
	if err := os.WriteFile(filepath.Join(workspaceDir, "go.work"), []byte(goWork), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}

	goMod := "module example.com/api\n\ngo 1.23.0\n"
	if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	selection, err := New(nil).Current(context.Background(), moduleDir)
	if err != nil {
		t.Fatalf("resolve current: %v", err)
	}

	if selection.GoVersion != "1.23.0" {
		t.Fatalf("GoVersion = %q, want %q", selection.GoVersion, "1.23.0")
	}
}

func TestResolverCurrent_ReturnsErrorWhenToolchainDirectiveIsEmpty(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.24.0\n\ntoolchain \n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	_, err := New(nil).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected an error when toolchain directive is empty")
	}

	if !strings.Contains(err.Error(), "toolchain directive is empty") {
		t.Fatalf("err = %q, want it to contain %q", err, "toolchain directive is empty")
	}
}

func TestResolverCurrent_ReturnsErrorWhenGoDirectiveMissing(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n"
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	_, err := New(nil).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected an error when go.mod has no go directive")
	}

	if !strings.Contains(err.Error(), "go directive not found") {
		t.Fatalf("err = %q, want it to contain %q", err, "go directive not found")
	}
}

func TestResolverCurrent_FallsBackToGlobalVersionOutsideRepos(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	selection, err := New(stubGlobalVersionSource{
		globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
			return "1.25.7", true, nil
		},
	}).Current(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("resolve current: %v", err)
	}

	if selection.GoVersion != "1.25.7" {
		t.Fatalf("GoVersion = %q, want %q", selection.GoVersion, "1.25.7")
	}
	if selection.GoSource != "global" {
		t.Fatalf("GoSource = %q, want %q", selection.GoSource, "global")
	}
}

func TestParseVersionMetadata_BareToolchainKeyword(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.24.0\n\ntoolchain\n"
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	_, err := New(nil).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected an error when bare toolchain keyword has no value")
	}
	if !strings.Contains(err.Error(), "toolchain directive is empty") {
		t.Fatalf("err = %q, want it to contain %q", err, "toolchain directive is empty")
	}
}

func TestParseVersionMetadata_EmptyGoDirective(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo \n"
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	_, err := New(nil).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected an error when go directive is empty")
	}
	if !strings.Contains(err.Error(), "go directive not found") {
		t.Fatalf("err = %q, want it to contain %q", err, "go directive not found")
	}
}

func TestParseVersionMetadata_ToolchainGoWithoutVersion(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.24.0\n\ntoolchain go\n"
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	_, err := New(nil).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected an error when toolchain go has no version after prefix")
	}
	if !strings.Contains(err.Error(), "toolchain directive is empty") {
		t.Fatalf("err = %q, want it to contain %q", err, "toolchain directive is empty")
	}
}

func TestResolverCurrent_ReturnsErrorWhenGlobalResolverFails(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	_, err := New(stubGlobalVersionSource{
		globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
			return "", false, fmt.Errorf("global boom")
		},
	}).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected an error when global version resolver fails")
	}
	if !strings.Contains(err.Error(), "global boom") {
		t.Fatalf("err = %q, want it to contain %q", err, "global boom")
	}
}

func TestResolverCurrent_ReturnsErrorWhenNoGoModAndNoGlobalVersion(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	_, err := New(stubGlobalVersionSource{
		globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
			return "", false, nil
		},
	}).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected an error when no go.mod and no global version set")
	}
	// The error should be the versionutil.ErrNotFound wrapped error from FindNearestFile.
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err = %q, want it to contain %q", err, "not found")
	}
}

func TestResolverCurrent_GoWorkFindNearestFileNonNotFoundError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workDir := filepath.Join(root, "sub")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create a symlink loop for go.work so FindNearestFile returns
	// a non-ErrNotFound error (e.g. "too many levels of symbolic links").
	goWorkLink := filepath.Join(root, "go.work")
	if err := os.Symlink(goWorkLink, goWorkLink); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err := New(nil).Current(context.Background(), workDir)
	if err == nil {
		t.Fatal("expected error when FindNearestFile go.work returns non-ErrNotFound, got nil")
	}
}

func TestResolverCurrent_GoWorkParseVersionMetadataError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	// Create a go.work with a bare "toolchain" keyword so parseVersionMetadata errors.
	if err := os.WriteFile(filepath.Join(tempDir, "go.work"), []byte("go 1.25.0\n\ntoolchain\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}

	_, err := New(nil).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected error when parseVersionMetadata fails on go.work, got nil")
	}
	if !strings.Contains(err.Error(), "toolchain directive is empty") {
		t.Fatalf("err = %q, want it to contain %q", err, "toolchain directive is empty")
	}
}

func TestParseVersionMetadata_OpenError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module x\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	// Make the file unreadable so os.Open fails.
	if err := os.Chmod(goModPath, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(goModPath, 0o644) })

	_, _, err := parseVersionMetadata(goModPath)
	if err == nil {
		t.Fatal("expected error when os.Open fails, got nil")
	}
	if !strings.Contains(err.Error(), "open") {
		t.Fatalf("err = %q, want it to contain %q", err, "open")
	}
}

func TestParseVersionMetadata_ScannerError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goModPath := filepath.Join(tempDir, "go.mod")
	// Create a file with a valid go directive followed by a line that exceeds
	// the scanner buffer (bufio.MaxScanTokenSize = 64KB), triggering scanner.Err().
	longLine := strings.Repeat("x", 70000) // exceeds 64KB default buffer
	content := "module x\n\ngo 1.25.0\n\n" + longLine + "\n"
	if err := os.WriteFile(goModPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	_, _, err := parseVersionMetadata(goModPath)
	if err == nil {
		t.Fatal("expected error when scanner.Err returns non-nil, got nil")
	}
	if !strings.Contains(err.Error(), "scan") {
		t.Fatalf("err = %q, want it to contain %q", err, "scan")
	}
}

func TestResolverCurrent_ReturnsErrorWhenNoGoModAndNilGlobal(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	_, err := New(nil).Current(context.Background(), tempDir)
	if err == nil {
		t.Fatal("expected an error when no go.mod and global resolver is nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err = %q, want it to contain %q", err, "not found")
	}
}
