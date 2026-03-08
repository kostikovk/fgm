package resolve

import (
	"context"
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
