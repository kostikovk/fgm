package lintimport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

type stubRegistry struct {
	registered map[string]string
}

func (s *stubRegistry) RegisterExistingLintInstallation(version string, binaryPath string) (string, error) {
	s.registered[version] = binaryPath
	return filepath.Join("/tmp/fgm/golangci-lint", version), nil
}

type failingRegistry struct {
	err error
}

func (f *failingRegistry) RegisterExistingLintInstallation(version string, binaryPath string) (string, error) {
	return "", f.err
}

func TestImporterImportAuto_RegistersDetectedLintInstallations(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	lintA := createExistingLintBinary(t, "2.11.2")
	lintB := createExistingLintBinary(t, "v2.10.1")
	registry := &stubRegistry{registered: make(map[string]string)}

	importer := New(Config{
		Candidates: []string{lintA, lintB},
		Registry:   registry,
	})
	imported, err := importer.ImportAuto(context.Background())
	if err != nil {
		t.Fatalf("ImportAuto: %v", err)
	}

	if len(imported) != 2 {
		t.Fatalf("len(imported) = %d, want %d", len(imported), 2)
	}
	if registry.registered["v2.11.2"] != lintA {
		t.Fatalf("registered[v2.11.2] = %q, want %q", registry.registered["v2.11.2"], lintA)
	}
	if registry.registered["v2.10.1"] != lintB {
		t.Fatalf("registered[v2.10.1] = %q, want %q", registry.registered["v2.10.1"], lintB)
	}
}

func createExistingLintBinary(t *testing.T, version string) string {
	t.Helper()

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	script := "#!/bin/sh\nprintf 'golangci-lint has version " + version + " built with go1.25.0\\n'\n"
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake lint binary: %v", err)
	}

	return binaryPath
}

func TestDefaultCandidates_IncludesBinaryOnPath(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: executable bit check not applicable on windows")
	}

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	candidates := DefaultCandidates(binDir)

	found := slices.Contains(candidates, binaryPath)
	if !found {
		t.Fatalf("expected %q in candidates %v", binaryPath, candidates)
	}
}

func TestDefaultCandidates_IncludesStaticPaths(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: static paths are unix-specific")
	}

	candidates := DefaultCandidates("")

	wantPaths := []string{
		"/opt/homebrew/bin/golangci-lint",
		"/usr/local/bin/golangci-lint",
	}

	for _, want := range wantPaths {
		found := slices.Contains(candidates, want)
		if !found {
			t.Errorf("expected static path %q in candidates %v", want, candidates)
		}
	}
}

func TestBinaryFromPath_FindsBinary(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: executable bit check not applicable on windows")
	}

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	got := binaryFromPath(binDir)
	if got != binaryPath {
		t.Fatalf("binaryFromPath(%q) = %q, want %q", binDir, got, binaryPath)
	}
}

func TestBinaryFromPath_SkipsNonExecutable(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: executable bit check not applicable on windows")
	}

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("write non-executable file: %v", err)
	}

	got := binaryFromPath(binDir)
	if got != "" {
		t.Fatalf("binaryFromPath(%q) = %q, want empty string for non-executable file", binDir, got)
	}
}

func TestBinaryFromPath_EmptyPath(t *testing.T) {
	t.Parallel()

	got := binaryFromPath("")
	if got != "" {
		t.Fatalf("binaryFromPath(\"\") = %q, want empty string", got)
	}
}

func TestDetectVersion_ParsesOutput(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: shell script detection not applicable on windows")
	}

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	script := "#!/bin/sh\nprintf 'golangci-lint has version 1.57.2 built with go1.21.0 from abc123 on linux/amd64\\n'\n"
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	version, err := detectVersion(context.Background(), binaryPath)
	if err != nil {
		t.Fatalf("detectVersion: %v", err)
	}
	if version != "v1.57.2" {
		t.Fatalf("detectVersion = %q, want %q", version, "v1.57.2")
	}
}

func TestImportAuto_SkipsDuplicateVersions(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: shell script detection not applicable on windows")
	}

	// Both binaries report the same version.
	lintA := createExistingLintBinary(t, "2.5.0")
	lintB := createExistingLintBinary(t, "2.5.0")
	registry := &stubRegistry{registered: make(map[string]string)}

	importer := New(Config{
		Candidates: []string{lintA, lintB},
		Registry:   registry,
	})
	imported, err := importer.ImportAuto(context.Background())
	if err != nil {
		t.Fatalf("ImportAuto: %v", err)
	}

	if len(imported) != 1 {
		t.Fatalf("len(imported) = %d, want 1 (duplicate version should be skipped)", len(imported))
	}
	if imported[0].Version != "v2.5.0" {
		t.Fatalf("imported[0].Version = %q, want %q", imported[0].Version, "v2.5.0")
	}
	if len(registry.registered) != 1 {
		t.Fatalf("len(registry.registered) = %d, want 1", len(registry.registered))
	}
}

func TestImportAuto_SkipsMissingCandidates(t *testing.T) {
	t.Parallel()

	nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist", "golangci-lint")
	registry := &stubRegistry{registered: make(map[string]string)}

	importer := New(Config{
		Candidates: []string{nonExistentPath},
		Registry:   registry,
	})
	imported, err := importer.ImportAuto(context.Background())
	if err != nil {
		t.Fatalf("ImportAuto: %v", err)
	}

	if len(imported) != 0 {
		t.Fatalf("len(imported) = %d, want 0 (missing candidate should be skipped)", len(imported))
	}
	if len(registry.registered) != 0 {
		t.Fatalf("len(registry.registered) = %d, want 0", len(registry.registered))
	}
}

func TestImportAuto_SkipsDirectoryCandidates(t *testing.T) {
	t.Parallel()

	dirPath := t.TempDir()
	registry := &stubRegistry{registered: make(map[string]string)}

	importer := New(Config{
		Candidates: []string{dirPath},
		Registry:   registry,
	})
	imported, err := importer.ImportAuto(context.Background())
	if err != nil {
		t.Fatalf("ImportAuto: %v", err)
	}

	if len(imported) != 0 {
		t.Fatalf("len(imported) = %d, want 0 (directory candidate should be skipped)", len(imported))
	}
	if len(registry.registered) != 0 {
		t.Fatalf("len(registry.registered) = %d, want 0", len(registry.registered))
	}
}

func TestDetectVersion_ErrorOnOutputWithoutVersionKeyword(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: shell script detection not applicable on windows")
	}

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	// Output has no "version" keyword at all.
	script := "#!/bin/sh\nprintf 'golangci-lint 1.57.2 built with go1.21.0\\n'\n"
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	_, err := detectVersion(context.Background(), binaryPath)
	if err == nil {
		t.Fatal("detectVersion: expected error for output without 'version' keyword, got nil")
	}
}

func TestDetectVersion_ErrorOnEmptyVersionAfterKeyword(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: shell script detection not applicable on windows")
	}

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	// "version" is the last field, so idx+1 is out of bounds for fields.
	script := "#!/bin/sh\nprintf 'golangci-lint has version\\n'\n"
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	_, err := detectVersion(context.Background(), binaryPath)
	if err == nil {
		t.Fatal("detectVersion: expected error when 'version' is last field, got nil")
	}
}

func TestDetectVersion_ErrorOnVersionKeywordFollowedByV(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: shell script detection not applicable on windows")
	}

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	// "version" keyword followed by just "v" which after TrimPrefix("v") yields "".
	script := "#!/bin/sh\nprintf 'golangci-lint has version v\\n'\n"
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	_, err := detectVersion(context.Background(), binaryPath)
	if err == nil {
		t.Fatal("detectVersion: expected error when version string is just 'v', got nil")
	}
}

func TestImportAuto_RegistryErrorIsPropagated(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	lint := createExistingLintBinary(t, "2.11.2")
	registry := &failingRegistry{err: fmt.Errorf("register boom")}

	importer := New(Config{
		Candidates: []string{lint},
		Registry:   registry,
	})
	_, err := importer.ImportAuto(context.Background())
	if err == nil {
		t.Fatal("expected error from registry, got nil")
	}
	if !strings.Contains(err.Error(), "register boom") {
		t.Fatalf("err = %q, want register boom", err)
	}
}

func TestImportAuto_DetectVersionErrorIsPropagated(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	// Create a "golangci-lint" binary that exits with an error.
	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	registry := &stubRegistry{registered: make(map[string]string)}
	importer := New(Config{
		Candidates: []string{binaryPath},
		Registry:   registry,
	})
	_, err := importer.ImportAuto(context.Background())
	if err == nil {
		t.Fatal("expected error from detectVersion, got nil")
	}
	if !strings.Contains(err.Error(), "run") {
		t.Fatalf("err = %q, want run error", err)
	}
}

func TestBinaryFromPath_SkipsDirectories(t *testing.T) {
	t.Parallel()

	// Create a directory with the binary name.
	binDir := t.TempDir()
	dirPath := filepath.Join(binDir, "golangci-lint")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got := binaryFromPath(binDir)
	if got != "" {
		t.Fatalf("binaryFromPath(%q) = %q, want empty for directory candidate", binDir, got)
	}
}

func TestBinaryFromPath_SkipsEmptySegments(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only: executable bit check not applicable on windows")
	}

	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "golangci-lint")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	pathEnv := string(filepath.ListSeparator) + binDir
	got := binaryFromPath(pathEnv)
	if got != binaryPath {
		t.Fatalf("binaryFromPath(%q) = %q, want %q", pathEnv, got, binaryPath)
	}
}

func TestBinaryFromPath_SkipsMissingStat(t *testing.T) {
	t.Parallel()

	pathEnv := filepath.Join(t.TempDir(), "nonexistent-dir")
	got := binaryFromPath(pathEnv)
	if got != "" {
		t.Fatalf("binaryFromPath(%q) = %q, want empty for missing dir", pathEnv, got)
	}
}

func TestBinaryName_ReturnsCorrectName(t *testing.T) {
	t.Parallel()

	got := binaryName()
	if runtime.GOOS == "windows" {
		if got != "golangci-lint.exe" {
			t.Fatalf("binaryName() = %q, want %q", got, "golangci-lint.exe")
		}
	} else {
		if got != "golangci-lint" {
			t.Fatalf("binaryName() = %q, want %q", got, "golangci-lint")
		}
	}
}
