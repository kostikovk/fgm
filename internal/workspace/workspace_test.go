package workspace

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestWorkspace(t *testing.T) {
	t.Parallel()

	makeTree := func(t *testing.T, files ...string) string {
		t.Helper()
		root := t.TempDir()
		for _, name := range files {
			path := filepath.Join(root, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(path, []byte("module m\n"), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}
		}
		return root
	}

	t.Run("ListModuleDirs", func(t *testing.T) {
		t.Parallel()

		t.Run("should skip vendor, testdata, node_modules and dot directories when listing modules", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := makeTree(t, "go.mod", "svc/b/go.mod", "svc/a/go.mod", "vendor/x/go.mod", "pkg/testdata/go.mod", "web/node_modules/y/go.mod", ".cache/z/go.mod")

			// Act
			dirs, err := ListModuleDirs(root)

			// Assert
			if err != nil {
				t.Fatalf("ListModuleDirs: %v", err)
			}
			want := []string{root, filepath.Join(root, "svc/a"), filepath.Join(root, "svc/b")}
			if !slices.Equal(dirs, want) {
				t.Errorf("dirs = %v; want %v", dirs, want)
			}
		})

		t.Run("should not follow a symlinked directory when it points outside the root", func(t *testing.T) {
			t.Parallel()

			// Arrange
			outside := makeTree(t, "go.mod")
			root := makeTree(t, "go.mod")
			if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
				t.Skipf("symlink unsupported: %v", err)
			}

			// Act
			dirs, err := ListModuleDirs(root)

			// Assert
			if err != nil {
				t.Fatalf("ListModuleDirs: %v", err)
			}
			if !slices.Equal(dirs, []string{root}) {
				t.Errorf("dirs = %v; want only the root module", dirs)
			}
		})
	})

	t.Run("RepoRoot", func(t *testing.T) {
		t.Parallel()

		t.Run("should return the nearest ancestor containing .git when one exists", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := makeTree(t, "svc/a/go.mod")
			if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
				t.Fatalf("mkdir .git: %v", err)
			}

			// Act
			got, ok := RepoRoot(filepath.Join(root, "svc/a"))

			// Assert
			if !ok || got != root {
				t.Errorf("RepoRoot = %q, %v; want %q", got, ok, root)
			}
		})
	})
}
