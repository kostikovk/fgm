package main

import (
	"os"
	"strings"
	"testing"
)

func TestGoReleaserConfigExistsAndContainsReleaseMetadata(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(".goreleaser.yaml")
	if err != nil {
		t.Fatalf("read .goreleaser.yaml: %v", err)
	}

	joined := string(content)
	for _, want := range []string{
		"project_name: fgm",
		"builds:",
		"archives:",
		"checksum:",
		"ldflags:",
		"brews:",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf(".goreleaser.yaml = %q, want it to contain %q", joined, want)
		}
	}
}

func TestReleaseWorkflowExistsAndRunsGoReleaser(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(".github/workflows/release.yml")
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	joined := string(content)
	for _, want := range []string{
		"on:",
		"tags:",
		"goreleaser/goreleaser-action",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("release workflow = %q, want it to contain %q", joined, want)
		}
	}
}

func TestInstallScriptExistsAndHandlesPlatformSelection(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("install.sh")
	if err != nil {
		t.Fatalf("read install.sh: %v", err)
	}

	joined := string(content)
	for _, want := range []string{
		"uname -s",
		"uname -m",
		"curl",
		"tar -xzf",
		"checksums.txt",
		"GITHUB_REPOSITORY",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("install.sh = %q, want it to contain %q", joined, want)
		}
	}
}

func TestMakefileEmbedsBuildMetadata(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}

	joined := string(content)
	for _, want := range []string{
		"LDFLAGS :=",
		"-X main.buildVersion=$(VERSION)",
		"-X main.buildCommit=$(COMMIT)",
		"-X main.buildDate=$(DATE)",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("Makefile = %q, want it to contain %q", joined, want)
		}
	}
}
