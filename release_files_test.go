package main

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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
			"postflight_steps do",
			"com.apple.quarantine",
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

func TestSetupFGMActionContract(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("setup-fgm/action.yml")
	if err != nil {
		t.Fatalf("read setup-fgm/action.yml: %v", err)
	}
	var action struct {
		Inputs map[string]struct {
			Required bool   `yaml:"required"`
			Default  string `yaml:"default"`
		} `yaml:"inputs"`
		Outputs map[string]any `yaml:"outputs"`
		Runs    struct {
			Using string `yaml:"using"`
		} `yaml:"runs"`
	}
	if err := yaml.Unmarshal(content, &action); err != nil {
		t.Fatalf("parse action.yml: %v", err)
	}

	t.Run("action.yml", func(t *testing.T) {
		t.Parallel()

		t.Run("should require an exact fgm version input when the action is used", func(t *testing.T) {
			t.Parallel()

			// Arrange
			version, ok := action.Inputs["version"]

			// Act
			required := ok && version.Required

			// Assert
			if !required {
				t.Error("inputs.version is not required")
			}
		})

		t.Run("should default working-directory, install and cache inputs when they are omitted", func(t *testing.T) {
			t.Parallel()

			// Arrange
			want := map[string]string{"working-directory": ".", "install": "true", "cache": "true"}

			// Act
			got := map[string]string{}
			for name := range want {
				got[name] = action.Inputs[name].Default
			}

			// Assert
			for name, value := range want {
				if got[name] != value {
					t.Errorf("inputs.%s default = %q; want %q", name, got[name], value)
				}
			}
		})

		t.Run("should expose the resolved versions and sources as outputs when the action finishes", func(t *testing.T) {
			t.Parallel()

			// Arrange
			want := []string{"go-version", "go-source", "golangci-lint-version", "golangci-lint-source"}

			// Act
			missing := []string{}
			for _, name := range want {
				if _, ok := action.Outputs[name]; !ok {
					missing = append(missing, name)
				}
			}

			// Assert
			if len(missing) > 0 || action.Runs.Using != "composite" {
				t.Errorf("missing outputs %v, runs.using = %q", missing, action.Runs.Using)
			}
		})

		t.Run("should pin install.sh by tag and keep GOTOOLCHAIN local when steps run", func(t *testing.T) {
			t.Parallel()

			// Arrange
			text := string(content)

			// Act
			usesLocalInstaller := strings.Contains(text, `sh "$GITHUB_ACTION_PATH/../install.sh"`) && strings.Contains(text, "FGM_VERSION:")

			// Assert
			if !usesLocalInstaller || !strings.Contains(text, `echo "GOTOOLCHAIN=local" >> "$GITHUB_ENV"`) {
				t.Error("action.yml must install via the bundled install.sh with FGM_VERSION and export GOTOOLCHAIN=local")
			}
		})
	})
}
