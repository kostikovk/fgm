package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/testutil"
)

func TestLanguageVersionCommands(t *testing.T) {
	t.Parallel()

	t.Run("install go", func(t *testing.T) {
		t.Parallel()

		t.Run("should install 1.26.0 and say so when asked for the language version 1.26", func(t *testing.T) {
			t.Parallel()

			// Arrange
			var installed string
			application := &app.App{GoInstaller: stubGoInstaller{installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				installed = version
				return "/tmp/fgm/go/" + version, nil
			}}}

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "install", "go", "1.26")

			// Assert
			if err != nil {
				t.Fatalf("install go: %v", err)
			}
			if installed != "1.26.0" || !strings.Contains(stdout, "Installed Go 1.26.0") || !strings.Contains(stdout, "1.26 is a language version") {
				t.Errorf("installed %q, stdout %q; want 1.26.0 with the mapping explained", installed, stdout)
			}
		})
	})

	t.Run("pin go", func(t *testing.T) {
		t.Parallel()

		t.Run("should write 1.26.0 when the language version 1.26 is pinned", func(t *testing.T) {
			t.Parallel()

			// Arrange
			dir := t.TempDir()

			// Act
			_, _, err := testutil.ExecuteCommand(t, NewRootCmd(&app.App{}), "pin", "go", "1.26", "--chdir", dir)

			// Assert
			if err != nil {
				t.Fatalf("pin go: %v", err)
			}
			content, err := os.ReadFile(filepath.Join(dir, ".fgm.toml"))
			if err != nil || !strings.Contains(string(content), "go = '1.26.0'") {
				t.Errorf(".fgm.toml = %q (err %v); want go = '1.26.0'", content, err)
			}
		})
	})

	t.Run("current", func(t *testing.T) {
		t.Parallel()

		t.Run("should show the directive next to the mapped release when go.mod declares a language version", func(t *testing.T) {
			t.Parallel()

			// Arrange
			application := &app.App{Resolver: stubSelectionResolver{currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.0", GoSource: "/repo/go.mod", GoSourceFile: "/repo/go.mod", GoBase: "1.26"}, nil
			}}}

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "current")

			// Assert
			if err != nil {
				t.Fatalf("current: %v", err)
			}
			if stdout != "go 1.26.0 (go.mod 1.26)\n" {
				t.Errorf("stdout = %q; want the directive in the label", stdout)
			}
		})
	})
}
