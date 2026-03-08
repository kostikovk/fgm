package goimport

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

func (s *stubRegistry) RegisterExistingGoInstallation(version string, goroot string) (string, error) {
	s.registered[version] = goroot
	return filepath.Join("/tmp/fgm/go", version), nil
}

func TestImporterImportAuto_RegistersDetectedGoInstallations(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script detection test is unix-only")
	}

	goRootA := createExistingGoInstallation(t, "1.25.7")
	goRootB := createExistingGoInstallation(t, "1.26.1")
	registry := &stubRegistry{registered: make(map[string]string)}

	importer := New([]string{goRootA, goRootB}, registry)
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
