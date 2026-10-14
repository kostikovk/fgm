package main

import (
	"os"
	"strings"
	"testing"
)

func TestGoReleaserConfigExistsAndContainsReleaseMetadata(t *testing.T) {
	t.Parallel()

	t.Run("should include build archive checksum changelog and homebrew metadata when reading .goreleaser.yaml", func(t *testing.T) {
		t.Parallel()

		// Arrange & Act
		content, err := os.ReadFile(".goreleaser.yaml")
		if err != nil {
			t.Fatalf("read .goreleaser.yaml: %v", err)
		}

		// Assert
		joined := string(content)
		for _, want := range []string{
			"project_name: fgm",
			"builds:",
			"archives:",
			"checksum:",
			"ldflags:",
			"changelog:",
			"Features",
			"Bug Fixes",
			"homebrew_casks:",
			"binaries:",
		} {
			if !strings.Contains(joined, want) {
				t.Fatalf(".goreleaser.yaml = %q, want it to contain %q", joined, want)
			}
		}
		if strings.Contains(joined, "\nbrews:") {
			t.Fatalf(".goreleaser.yaml still contains deprecated brews section")
		}
	})
}

func TestReleaseWorkflowExistsAndRunsGoReleaser(t *testing.T) {
	t.Parallel()

	t.Run("should auto-tag on main push and run GoReleaser with module proxy warm-up when reading release workflow", func(t *testing.T) {
		t.Parallel()

		// Arrange & Act
		content, err := os.ReadFile(".github/workflows/release.yml")
		if err != nil {
			t.Fatalf("read release workflow: %v", err)
		}

		// Assert
		joined := string(content)
		for _, want := range []string{
			"on:",
			"branches:",
			"- main",
			"tags:",
			"concurrency:",
			"go run ./scripts/next-version",
			"GORELEASER_CURRENT_TAG:",
			"goreleaser/goreleaser-action",
			"proxy.golang.org/github.com/kostikovk/fgm/@v/",
		} {
			if !strings.Contains(joined, want) {
				t.Fatalf("release workflow = %q, want it to contain %q", joined, want)
			}
		}
	})
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

	t.Run("should embed build ldflags and expose next-version target when reading Makefile", func(t *testing.T) {
		t.Parallel()

		// Arrange & Act
		content, err := os.ReadFile("Makefile")
		if err != nil {
			t.Fatalf("read Makefile: %v", err)
		}

		// Assert
		joined := string(content)
		for _, want := range []string{
			"LDFLAGS :=",
			"-X main.buildVersion=$(VERSION)",
			"-X main.buildCommit=$(COMMIT)",
			"-X main.buildDate=$(DATE)",
			"next-version:",
			"go run ./scripts/next-version",
		} {
			if !strings.Contains(joined, want) {
				t.Fatalf("Makefile = %q, want it to contain %q", joined, want)
			}
		}
	})
}
