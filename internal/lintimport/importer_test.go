package lintimport

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type stubRegistry struct {
	registered map[string]string
}

func (s *stubRegistry) RegisterExistingLintInstallation(version string, binaryPath string) (string, error) {
	s.registered[version] = binaryPath
	return filepath.Join("/tmp/fgm/golangci-lint", version), nil
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
