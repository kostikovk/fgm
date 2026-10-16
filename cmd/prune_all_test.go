package cmd

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/testutil"
)

// newMultiModulePruneApp selects Go and lint per module directory name and records deletions.
func newMultiModulePruneApp(deletedGo *[]string, deletedLint *[]string) *app.App {
	selections := map[string]app.Selection{
		"api":    {GoVersion: "1.26.9", LintVersion: "v2.9.0"},
		"worker": {GoVersion: "1.25.14", LintVersion: "v2.5.0"},
	}
	return &app.App{
		Resolver: stubSelectionResolver{currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
			if selection, ok := selections[filepath.Base(workDir)]; ok {
				return selection, nil
			}
			return app.Selection{GoVersion: "1.27.2", GoSource: "global"}, nil
		}},
		GoStore: stubGoStore{
			shimDirFn:         func() string { return "/tmp/fgm/shims" },
			globalGoVersionFn: func(ctx context.Context) (string, bool, error) { return "1.27.2", true, nil },
			listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.27.2", "1.26.9", "1.25.14", "1.24.1"}, nil
			},
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/" + version + "/bin/go", nil
			},
			deleteGoVersionFn: func(ctx context.Context, version string) (string, error) {
				*deletedGo = append(*deletedGo, version)
				return "/tmp/fgm/go/" + version, nil
			},
		},
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) { return []string{"v2.9.0", "v2.5.0", "v1.64.8"}, nil },
			deleteLintVersionFn: func(ctx context.Context, version string) (string, error) {
				*deletedLint = append(*deletedLint, version)
				return "/tmp/fgm/golangci-lint/" + version, nil
			},
		},
	}
}

var multiModuleFiles = map[string]string{
	"svc/api/go.mod":    "module api\n\ngo 1.26.0\n",
	"svc/worker/go.mod": "module worker\n\ngo 1.25.0\n",
}

func TestPruneCommandMultiModule(t *testing.T) {
	t.Parallel()

	t.Run("RunE", func(t *testing.T) {
		t.Parallel()

		t.Run("should keep every version a module selects when --all is set", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, multiModuleFiles)
			var deletedGo, deletedLint []string

			// Act
			_, _, err := testutil.ExecuteCommand(t, NewRootCmd(newMultiModulePruneApp(&deletedGo, &deletedLint)), "prune", "--all", "--chdir", root)

			// Assert
			if err != nil {
				t.Fatalf("prune --all: %v", err)
			}
			if !slices.Equal(deletedGo, []string{"1.24.1"}) || !slices.Equal(deletedLint, []string{"v1.64.8"}) {
				t.Errorf("deleted go %v lint %v; want only 1.24.1 and v1.64.8", deletedGo, deletedLint)
			}
		})

		t.Run("should name a selecting module for each kept version when --all runs dry", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, multiModuleFiles)
			var deletedGo, deletedLint []string

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(newMultiModulePruneApp(&deletedGo, &deletedLint)), "prune", "--all", "--dry-run", "--chdir", root)

			// Assert
			if err != nil {
				t.Fatalf("prune --all --dry-run: %v", err)
			}
			for _, want := range []string{
				"Keep Go 1.27.2 (global default)",
				"Keep Go 1.26.9 (selected by svc/api)",
				"Keep golangci-lint v2.5.0 (selected by svc/worker)",
				"Would remove Go 1.24.1",
			} {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout = %q; want it to contain %q", stdout, want)
				}
			}
			if len(deletedGo)+len(deletedLint) != 0 {
				t.Errorf("dry run deleted go %v lint %v", deletedGo, deletedLint)
			}
		})

		t.Run("should remove nothing and warn when the directory has no Go metadata and --all is absent", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, multiModuleFiles)
			var deletedGo, deletedLint []string

			// Act
			_, stderr, err := testutil.ExecuteCommand(t, NewRootCmd(newMultiModulePruneApp(&deletedGo, &deletedLint)), "prune", "--chdir", root)

			// Assert
			if err == nil || !strings.Contains(err.Error(), "--all") {
				t.Errorf("err = %v; want a refusal that suggests --all", err)
			}
			if !strings.Contains(stderr, "2 Go modules") {
				t.Errorf("stderr = %q; want a warning counting the modules", stderr)
			}
			if len(deletedGo)+len(deletedLint) != 0 {
				t.Errorf("deleted go %v lint %v; want nothing", deletedGo, deletedLint)
			}
		})

		t.Run("should warn about sibling modules when pruning from inside one module", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, multiModuleFiles)
			var deletedGo, deletedLint []string

			// Act
			_, stderr, err := testutil.ExecuteCommand(t, NewRootCmd(newMultiModulePruneApp(&deletedGo, &deletedLint)), "prune", "--dry-run", "--chdir", filepath.Join(root, "svc", "api"))

			// Assert
			if err != nil {
				t.Fatalf("prune --dry-run: %v", err)
			}
			if !strings.Contains(stderr, "1 other Go module") || !strings.Contains(stderr, "--all") {
				t.Errorf("stderr = %q; want a sibling-module warning suggesting --all", stderr)
			}
		})
	})
}
