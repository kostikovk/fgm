package lintdetect

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kostikovk/fgm/internal/workspace"
)

// writeSyntheticRepo lays out modules × packages directories and workflow files, with one
// golangci-lint-action step that has no working-directory.
func writeSyntheticRepo(tb testing.TB, modules int, packages int, workflows int) string {
	tb.Helper()
	root := tb.TempDir()
	mustWrite := func(path string, content string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			tb.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			tb.Fatalf("write: %v", err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		tb.Fatalf("mkdir .git: %v", err)
	}
	for m := range modules {
		dir := filepath.Join(root, "svc", fmt.Sprintf("m%02d", m))
		mustWrite(filepath.Join(dir, "go.mod"), "module m\n\ngo 1.26.0\n")
		for p := range packages {
			if err := os.MkdirAll(filepath.Join(dir, "pkg", fmt.Sprintf("p%02d", p)), 0o755); err != nil {
				tb.Fatalf("mkdir: %v", err)
			}
		}
	}
	for w := range workflows {
		body := "jobs:\n  build:\n    steps:\n      - run: go test ./...\n"
		if w == 0 {
			body += "      - uses: golangci/golangci-lint-action@v8\n        with:\n          version: v2.5.0\n"
		}
		mustWrite(filepath.Join(root, ".github", "workflows", fmt.Sprintf("w%03d.yml", w)), body)
	}
	return root
}

func BenchmarkDetectInMultiModuleRepo(b *testing.B) {
	root := writeSyntheticRepo(b, 50, 100, 200)
	module := filepath.Join(root, "svc", "m07")

	b.ResetTimer()
	for b.Loop() {
		if _, err := Detect(module); err != nil {
			b.Fatalf("Detect: %v", err)
		}
	}
}

// Not parallel: it reads the process-wide walk counter.
func TestDetectWorkflowScope(t *testing.T) {
	t.Run("Detect", func(t *testing.T) {
		t.Run("should apply a root workflow pin to the root module without walking the tree", func(t *testing.T) {
			// Arrange
			root := writeSyntheticRepo(t, 3, 2, 2)
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module root\n\ngo 1.26.0\n"), 0o644); err != nil {
				t.Fatalf("write root go.mod: %v", err)
			}
			walksBefore := workspace.WalkCount()

			// Act
			rootPins, rootErr := Detect(root)
			modulePins, moduleErr := Detect(filepath.Join(root, "svc", "m01"))

			// Assert
			if rootErr != nil || moduleErr != nil {
				t.Fatalf("Detect errors: %v, %v", rootErr, moduleErr)
			}
			if len(rootPins.Pins) != 1 || len(modulePins.Pins) != 0 {
				t.Errorf("root pins %v, module pins %v; want the workflow pin on the root module only", rootPins.Pins, modulePins.Pins)
			}
			if workspace.WalkCount() != walksBefore {
				t.Errorf("Detect walked the repository; want no walk when the root has a go.mod")
			}
		})

		t.Run("should ignore a root workflow pin in a repository with several modules and no root go.mod", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeSyntheticRepo(t, 3, 2, 2)

			// Act
			result, err := Detect(filepath.Join(root, "svc", "m01"))

			// Assert
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			if len(result.Pins) != 0 {
				t.Errorf("pins = %v; want none", result.Pins)
			}
		})
	})
}
