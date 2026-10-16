package currenttoolchain

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
)

type stubCatalog struct{}

func (stubCatalog) ListCatalogLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
	return []app.LintVersion{{Version: "v2.9.0", Recommended: true}}, nil
}

func TestServiceDetectedLint(t *testing.T) {
	t.Parallel()

	newRepo := func(t *testing.T, files map[string]string) string {
		t.Helper()
		root := t.TempDir()
		files["go.mod"] = "module m\n\ngo 1.26.0\n"
		if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
			t.Fatalf("mkdir .git: %v", err)
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		return root
	}
	newService := func() *Service {
		return New(Config{
			GoResolver: stubGoResolver{currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.0", GoSource: "go.mod"}, nil
			}},
			LintCatalogProvider: stubCatalog{},
		})
	}

	t.Run("Current", func(t *testing.T) {
		t.Parallel()

		t.Run("should select the detected pin with its file and line when one source pins the linter", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := newRepo(t, map[string]string{"Makefile": "\nGOLANGCI_LINT_VERSION ?= v2.5.0\n"})

			// Act
			selection, err := newService().Current(context.Background(), root)

			// Assert
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if selection.LintVersion != "v2.5.0" || selection.LintSource != "detected" {
				t.Fatalf("lint = %s (%s); want v2.5.0 (detected)", selection.LintVersion, selection.LintSource)
			}
			if selection.LintSourceFile != filepath.Join(root, "Makefile") || selection.LintSourceLine != 2 {
				t.Errorf("source = %s:%d; want Makefile:2", selection.LintSourceFile, selection.LintSourceLine)
			}
		})

		t.Run("should fall back to auto with a warning listing candidates when detected pins disagree", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := newRepo(t, map[string]string{"Makefile": "GOLANGCI_LINT_VERSION = v2.5.0\n", ".tool-versions": "golangci-lint 2.6.0\n"})

			// Act
			selection, err := newService().Current(context.Background(), root)

			// Assert
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if selection.LintVersion != "v2.9.0" || selection.LintSource != "remote" {
				t.Errorf("lint = %s (%s); want the auto selection", selection.LintVersion, selection.LintSource)
			}
			warnings := strings.Join(selection.Warnings, "\n")
			if !strings.Contains(warnings, "v2.5.0 (Makefile:1)") || !strings.Contains(warnings, "v2.6.0 (.tool-versions:1)") {
				t.Errorf("warnings = %q; want both candidates with file:line", warnings)
			}
		})

		t.Run("should ignore detected pins when .fgm.toml disables detection", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := newRepo(t, map[string]string{
				"Makefile":  "GOLANGCI_LINT_VERSION = v2.5.0\n",
				".fgm.toml": "[toolchain]\ngolangci_lint_detect = false\n",
			})

			// Act
			selection, err := newService().Current(context.Background(), root)

			// Assert
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if selection.LintSource == "detected" {
				t.Errorf("lint = %s (%s); want detection disabled", selection.LintVersion, selection.LintSource)
			}
		})

		t.Run("should prefer the .fgm.toml version when a detected pin also exists", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := newRepo(t, map[string]string{
				"Makefile":  "GOLANGCI_LINT_VERSION = v2.5.0\n",
				".fgm.toml": "[toolchain]\ngolangci_lint = \"v2.7.0\"\n",
			})

			// Act
			selection, err := newService().Current(context.Background(), root)

			// Assert
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if selection.LintVersion != "v2.7.0" || selection.LintSource != "config" {
				t.Errorf("lint = %s (%s); want v2.7.0 (config)", selection.LintVersion, selection.LintSource)
			}
		})
	})
}
