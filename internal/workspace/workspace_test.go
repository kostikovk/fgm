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

func TestGoDirective(t *testing.T) {
	t.Parallel()

	t.Run("GoDirective", func(t *testing.T) {
		t.Parallel()

		t.Run("should return the version and line when the file declares a go directive", func(t *testing.T) {
			t.Parallel()

			// Arrange
			path := filepath.Join(t.TempDir(), "go.mod")
			if err := os.WriteFile(path, []byte("module m\n\ngo 1.26.0 // minimum\n"), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}

			// Act
			version, line, found, err := GoDirective(path)

			// Assert
			if err != nil || !found || version != "1.26.0" || line != 3 {
				t.Errorf("GoDirective = %q, %d, %v, %v; want 1.26.0 at line 3", version, line, found, err)
			}
		})
	})
}

func TestHasMultipleModules(t *testing.T) {
	t.Parallel()

	t.Run("HasMultipleModules", func(t *testing.T) {
		t.Parallel()

		t.Run("should report true without listing every module when two modules exist", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := t.TempDir()
			for _, name := range []string{"a/go.mod", "b/go.mod", "c/go.mod"} {
				if err := os.MkdirAll(filepath.Join(root, filepath.Dir(name)), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(root, name), []byte("module m\n"), 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
			}

			// Act
			first, err := HasMultipleModules(root)
			walksAfterFirst := WalkCount()
			second, secondErr := HasMultipleModules(root)

			// Assert
			if err != nil || secondErr != nil || !first || !second {
				t.Fatalf("HasMultipleModules = %v, %v (errors %v, %v); want true twice", first, second, err, secondErr)
			}
			if WalkCount() != walksAfterFirst {
				t.Errorf("second call walked the tree again")
			}
		})

		t.Run("should report false when the tree holds a single module", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module m\n"), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}

			// Act
			got, err := HasMultipleModules(root)

			// Assert
			if err != nil || got {
				t.Errorf("HasMultipleModules = %v, %v; want false", got, err)
			}
		})
	})
}
