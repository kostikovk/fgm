package lintconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiagnose_NoConfigFile(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	workDir := t.TempDir()
	findings, err := Diagnose(catalog, workDir, "1.26.1")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != "WARN" {
		t.Fatalf("severity = %q, want WARN", findings[0].Severity)
	}
}

func TestDiagnose_IgnoresUnsupportedTomlConfig(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	workDir := t.TempDir()
	config := "version = \"2\"\n[linters]\nenable = [\"govet\"]\n"
	if err := os.WriteFile(filepath.Join(workDir, ".golangci.toml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	findings, err := Diagnose(catalog, workDir, "1.26.1")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != "WARN" {
		t.Fatalf("severity = %q, want WARN", findings[0].Severity)
	}
	if findings[0].Message != "no .golangci.yml found; run 'fgm lint init' to generate one" {
		t.Fatalf("message = %q", findings[0].Message)
	}
}

func TestDiagnose_MalformedYAML(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, ".golangci.yml"), []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	findings, err := Diagnose(catalog, workDir, "1.26.1")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	hasError := false
	for _, f := range findings {
		if f.Severity == "ERROR" {
			hasError = true
		}
	}
	if !hasError {
		t.Fatal("expected ERROR finding for malformed YAML")
	}
}

func TestDiagnose_MissingVersionField(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	workDir := t.TempDir()
	config := "linters:\n  enable:\n    - govet\n"
	if err := os.WriteFile(filepath.Join(workDir, ".golangci.yml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	findings, err := Diagnose(catalog, workDir, "1.26.1")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	hasVersionError := false
	for _, f := range findings {
		if f.Severity == "ERROR" && f.Message == "config is missing 'version: \"2\"'; golangci-lint v2 requires this field" {
			hasVersionError = true
		}
	}
	if !hasVersionError {
		t.Fatal("expected ERROR for missing version field")
	}
}

func TestDiagnose_UnknownLinter(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	workDir := t.TempDir()
	config := "version: \"2\"\nlinters:\n  enable:\n    - nonexistentlinter\n"
	if err := os.WriteFile(filepath.Join(workDir, ".golangci.yml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	findings, err := Diagnose(catalog, workDir, "1.26.1")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	hasUnknown := false
	for _, f := range findings {
		if f.Severity == "WARN" && f.Message == "unknown linter \"nonexistentlinter\"; check for typos" {
			hasUnknown = true
		}
	}
	if !hasUnknown {
		t.Fatal("expected WARN for unknown linter")
	}
}

func TestDiagnose_ConflictingLinters(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	workDir := t.TempDir()
	config := "version: \"2\"\nformatters:\n  enable:\n    - gofmt\n    - gofumpt\n"
	if err := os.WriteFile(filepath.Join(workDir, ".golangci.yml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	findings, err := Diagnose(catalog, workDir, "1.26.1")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	hasConflict := false
	for _, f := range findings {
		if f.Severity == "WARN" && f.Message == "gofumpt is a superset of gofmt; enable only one" {
			hasConflict = true
		}
	}
	if !hasConflict {
		t.Fatal("expected WARN for conflicting linters")
	}
}

func TestDiagnose_GoodConfig(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	// Generate a config with all standard linters and then diagnose it.
	workDir := t.TempDir()
	data, err := Generate(catalog, GenerateOptions{
		GoVersion:   "1.26.1",
		LintVersion: "v2.11.2",
		Preset:      PresetStandard,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, ".golangci.yml"), data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	findings, err := Diagnose(catalog, workDir, "1.26.1")
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}

	for _, f := range findings {
		if f.Severity == "ERROR" {
			t.Fatalf("unexpected ERROR: %s", f.Message)
		}
	}
}
