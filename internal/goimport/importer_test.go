package goimport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type stubRegistry struct {
	registered map[string]string
}

func (s *stubRegistry) RegisterExistingGoInstallation(version string, goroot string) (string, error) {
	s.registered[version] = goroot
	return filepath.Join("/tmp/fgm/go", version), nil
}

type failingRegistry struct {
	err error
}

func (f *failingRegistry) RegisterExistingGoInstallation(version string, goroot string) (string, error) {
	return "", f.err
}

func TestImporterImportAuto_RegistersDetectedGoInstallations(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	goRootA := createExistingGoInstallation(t, "1.25.7")
	goRootB := createExistingGoInstallation(t, "1.26.1")
	registry := &stubRegistry{registered: make(map[string]string)}

	importer := New(Config{
		Candidates: []string{goRootA, goRootB},
		Registry:   registry,
	})
	imported, err := importer.ImportAuto(context.Background())
	if err != nil {
		t.Fatalf("ImportAuto: %v", err)
	}

	if len(imported) != 2 {
		t.Fatalf("len(imported) = %d, want %d", len(imported), 2)
	}
	if registry.registered["1.25.7"] != goRootA {
		t.Fatalf("registered[1.25.7] = %q, want %q", registry.registered["1.25.7"], goRootA)
	}
	if registry.registered["1.26.1"] != goRootB {
		t.Fatalf("registered[1.26.1] = %q, want %q", registry.registered["1.26.1"], goRootB)
	}
}

func createExistingGoInstallation(t *testing.T, version string) string {
	t.Helper()

	goRoot := t.TempDir()
	goBinary := filepath.Join(goRoot, "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBinary), 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}
	script := "#!/bin/sh\nprintf 'go version go" + version + " darwin/arm64\\n'\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	return goRoot
}

func TestDefaultCandidates_IncludesStaticPaths(t *testing.T) {
	t.Parallel()

	candidates := DefaultCandidates("")

	staticPaths := []string{
		"/usr/local/go",
		"/opt/homebrew/opt/go/libexec",
		"/usr/local/opt/go/libexec",
	}
	candidateSet := make(map[string]struct{}, len(candidates))
	for _, c := range candidates {
		candidateSet[c] = struct{}{}
	}
	for _, want := range staticPaths {
		if _, ok := candidateSet[want]; !ok {
			t.Errorf("DefaultCandidates(\"\") missing static path %q", want)
		}
	}
}

func TestDefaultCandidates_IncludesGoOnPath(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	// Create a fake go binary inside a bin/ subdirectory so that
	// gorootFromPathGo can stat it and return the parent-of-parent as GOROOT.
	goRoot := t.TempDir()
	binDir := filepath.Join(goRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	goBinary := filepath.Join(binDir, "go")
	script := "#!/bin/sh\nprintf 'go version go1.99.0 linux/amd64\\n'\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	// pathEnv points at the bin/ dir so gorootFromPathGo will find the binary.
	candidates := DefaultCandidates(binDir)

	found := false
	for _, c := range candidates {
		if c == goRoot {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("DefaultCandidates(%q): GOROOT %q not in candidates %v", binDir, goRoot, candidates)
	}
}

func TestGorootFromPathGo_EmptyPath(t *testing.T) {
	t.Parallel()

	got := gorootFromPathGo("")
	if got != "" {
		t.Errorf("gorootFromPathGo(\"\") = %q, want \"\"", got)
	}
}

func TestGorootFromPathGo_FindsGoBinary(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	goRoot := t.TempDir()
	binDir := filepath.Join(goRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	goBinary := filepath.Join(binDir, "go")
	if err := os.WriteFile(goBinary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	got := gorootFromPathGo(binDir)
	if got != goRoot {
		t.Errorf("gorootFromPathGo(%q) = %q, want %q", binDir, got, goRoot)
	}
}

func TestGorootFromPathGo_SkipsEmptyDirSegments(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	goRoot := t.TempDir()
	binDir := filepath.Join(goRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	goBinary := filepath.Join(binDir, "go")
	if err := os.WriteFile(goBinary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	// PATH with an empty segment before the real dir.
	pathEnv := string(filepath.ListSeparator) + binDir
	got := gorootFromPathGo(pathEnv)
	if got != goRoot {
		t.Errorf("gorootFromPathGo(%q) = %q, want %q", pathEnv, got, goRoot)
	}
}

func TestDetectVersion_ParsesOutput(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	goRoot := createExistingGoInstallation(t, "1.21.3")
	goBinary := filepath.Join(goRoot, "bin", "go")

	version, err := detectVersion(context.Background(), goBinary)
	if err != nil {
		t.Fatalf("detectVersion: %v", err)
	}
	if version != "1.21.3" {
		t.Errorf("detectVersion = %q, want %q", version, "1.21.3")
	}
}

func TestDetectVersion_ErrorOnBadOutput(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	goRoot := t.TempDir()
	goBinary := filepath.Join(goRoot, "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBinary), 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}
	// Output has fewer than 3 fields, so parsing should fail.
	script := "#!/bin/sh\nprintf 'garbage\\n'\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	_, err := detectVersion(context.Background(), goBinary)
	if err == nil {
		t.Fatal("detectVersion: expected error for bad output, got nil")
	}
}

func TestDetectVersion_ErrorOnEmptyVersion(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	goRoot := t.TempDir()
	goBinary := filepath.Join(goRoot, "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBinary), 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}
	// Third field is "go" with no version suffix, so TrimPrefix yields "".
	script := "#!/bin/sh\nprintf 'go version go linux/amd64\\n'\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	_, err := detectVersion(context.Background(), goBinary)
	if err == nil {
		t.Fatal("detectVersion: expected error for empty version field, got nil")
	}
}

func TestImportAuto_SkipsDuplicateVersions(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	// Two candidates that both report the same version.
	goRootA := createExistingGoInstallation(t, "1.22.0")
	goRootB := createExistingGoInstallation(t, "1.22.0")
	registry := &stubRegistry{registered: make(map[string]string)}

	importer := New(Config{
		Candidates: []string{goRootA, goRootB},
		Registry:   registry,
	})
	imported, err := importer.ImportAuto(context.Background())
	if err != nil {
		t.Fatalf("ImportAuto: %v", err)
	}

	if len(imported) != 1 {
		t.Fatalf("len(imported) = %d, want 1 (duplicate should be skipped)", len(imported))
	}
	if registry.registered["1.22.0"] != goRootA {
		t.Errorf("registered[1.22.0] = %q, want %q", registry.registered["1.22.0"], goRootA)
	}
}

func TestImportAuto_SkipsMissingCandidates(t *testing.T) {
	t.Parallel()

	nonExistent := filepath.Join(t.TempDir(), "does-not-exist")
	registry := &stubRegistry{registered: make(map[string]string)}

	importer := New(Config{
		Candidates: []string{nonExistent},
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
		t.Errorf("registry should be empty, got %v", registry.registered)
	}
}

func TestGoBinaryName_ReturnsCorrectName(t *testing.T) {
	t.Parallel()

	got := goBinaryName()
	if runtime.GOOS == "windows" {
		if got != "go.exe" {
			t.Fatalf("goBinaryName() = %q, want %q", got, "go.exe")
		}
	} else {
		if got != "go" {
			t.Fatalf("goBinaryName() = %q, want %q", got, "go")
		}
	}
}

func TestImportAuto_RegistryErrorIsPropagated(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	goRoot := createExistingGoInstallation(t, "1.25.7")
	registry := &failingRegistry{err: fmt.Errorf("register boom")}

	importer := New(Config{
		Candidates: []string{goRoot},
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

	// Create a "go" binary that exits with an error.
	goRoot := t.TempDir()
	goBinary := filepath.Join(goRoot, "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBinary), 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}
	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	registry := &stubRegistry{registered: make(map[string]string)}
	importer := New(Config{
		Candidates: []string{goRoot},
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

func TestGorootFromPathGo_SkipsMissingBinary(t *testing.T) {
	t.Parallel()

	// Create a PATH directory that exists but contains no go binary.
	emptyDir := t.TempDir()

	got := gorootFromPathGo(emptyDir)
	if got != "" {
		t.Fatalf("gorootFromPathGo(%q) = %q, want empty string when dir has no go binary", emptyDir, got)
	}
}
