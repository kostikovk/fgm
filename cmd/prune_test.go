package cmd

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/testutil"
)

func TestPruneCommand_RemovesUnusedGoAndLintVersionsKeepingGlobalAndCurrent(t *testing.T) {
	t.Parallel()

	t.Run("should remove unused Go and golangci-lint versions while preserving global Go, current project Go, and current project linter when prune runs", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var deletedGo []string
		var deletedLint []string

		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						GoSource:    "go.mod",
						LintVersion: "v2.14.0",
						LintSource:  "local",
					}, nil
				},
			},
			GoStore: stubGoStore{
				shimDirFn: func() string { return "/tmp/fgm/shims" },
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
					return "1.26.4", true, nil
				},
				listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"1.27.2", "1.26.4", "1.25.7", "1.24.1"}, nil
				},
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/go/" + version + "/bin/go", nil
				},
				deleteGoVersionFn: func(ctx context.Context, version string) (string, error) {
					deletedGo = append(deletedGo, version)
					return "/tmp/fgm/go/" + version, nil
				},
			},
			LintStore: stubLintStore{
				listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"v2.14.0", "v2.11.2", "v1.64.8"}, nil
				},
				deleteLintVersionFn: func(ctx context.Context, version string) (string, error) {
					deletedLint = append(deletedLint, version)
					return "/tmp/fgm/golangci-lint/" + version, nil
				},
			},
		}

		// Act
		stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "prune")

		// Assert
		if err != nil {
			t.Fatalf("prune: %v", err)
		}
		if !slices.Equal(deletedGo, []string{"1.25.7", "1.24.1"}) {
			t.Fatalf("deletedGo = %v, want [1.25.7 1.24.1]", deletedGo)
		}
		if !slices.Equal(deletedLint, []string{"v2.11.2", "v1.64.8"}) {
			t.Fatalf("deletedLint = %v, want [v2.11.2 v1.64.8]", deletedLint)
		}
		for _, want := range []string{
			"Removed Go 1.25.7 from /tmp/fgm/go/1.25.7",
			"Removed Go 1.24.1 from /tmp/fgm/go/1.24.1",
			"Removed golangci-lint v2.11.2 from /tmp/fgm/golangci-lint/v2.11.2",
			"Removed golangci-lint v1.64.8 from /tmp/fgm/golangci-lint/v1.64.8",
		} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
			}
		}
	})
}

func TestPruneCommand_DryRunPreviewsWithoutDeleting(t *testing.T) {
	t.Parallel()

	t.Run("should print would-remove lines without calling DeleteGoVersion or DeleteLintVersion when --dry-run is passed", func(t *testing.T) {
		t.Parallel()

		// Arrange
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						LintVersion: "v2.14.0",
					}, nil
				},
			},
			GoStore: stubGoStore{
				shimDirFn: func() string { return "/tmp/fgm/shims" },
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
					return "1.27.2", true, nil
				},
				listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"1.27.2", "1.25.7"}, nil
				},
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/go/" + version + "/bin/go", nil
				},
				deleteGoVersionFn: func(ctx context.Context, version string) (string, error) {
					t.Fatal("DeleteGoVersion should not be called during --dry-run")
					return "", nil
				},
			},
			LintStore: stubLintStore{
				listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"v2.14.0", "v2.10.1"}, nil
				},
				deleteLintVersionFn: func(ctx context.Context, version string) (string, error) {
					t.Fatal("DeleteLintVersion should not be called during --dry-run")
					return "", nil
				},
			},
		}

		// Act
		stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "prune", "--dry-run")

		// Assert
		if err != nil {
			t.Fatalf("prune --dry-run: %v", err)
		}
		for _, want := range []string{
			"Would remove Go 1.25.7",
			"Would remove golangci-lint v2.10.1",
		} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
			}
		}
	})
}

func TestPruneCommand_SkipsUnmanagedSystemGoOutsideStoreRoot(t *testing.T) {
	t.Parallel()

	t.Run("should skip unmanaged system Go installation outside FGM store root when pruning", func(t *testing.T) {
		t.Parallel()

		// Arrange
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2"}, nil
				},
			},
			GoStore: stubGoStore{
				shimDirFn: func() string { return "/tmp/fgm/shims" },
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
					return "1.27.2", true, nil
				},
				listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"1.27.2", "1.23.6"}, nil
				},
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					if version == "1.23.6" {
						return "/usr/local/go/bin/go", nil
					}
					return "/tmp/fgm/go/" + version + "/bin/go", nil
				},
				deleteGoVersionFn: func(ctx context.Context, version string) (string, error) {
					t.Fatalf("DeleteGoVersion should not be called for unmanaged system Go %s", version)
					return "", nil
				},
			},
		}

		// Act
		stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "prune")

		// Assert
		if err != nil {
			t.Fatalf("prune: %v", err)
		}
		if !strings.Contains(stdout, "No unused Go or golangci-lint versions to prune.") {
			t.Fatalf("stdout = %q, want empty notice", stdout)
		}
	})
}

func TestPruneCommand_PreservesLintCompatibleWithGlobalGo(t *testing.T) {
	t.Parallel()

	t.Run("should preserve installed golangci-lint compatible with global Go when current repo uses a newer Go and linter", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var deletedLint []string
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						GoSource:    "go.mod",
						LintVersion: "v2.14.0",
						LintSource:  "local",
					}, nil
				},
			},
			GoStore: stubGoStore{
				shimDirFn: func() string { return "/tmp/fgm/shims" },
				globalGoVersionFn: func(ctx context.Context) (string, bool, error) {
					return "1.25.7", true, nil
				},
				listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"1.27.2", "1.25.7"}, nil
				},
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/go/" + version + "/bin/go", nil
				},
			},
			LintStore: stubLintStore{
				listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"v2.14.0", "v2.11.2", "v1.64.8"}, nil
				},
				deleteLintVersionFn: func(ctx context.Context, version string) (string, error) {
					deletedLint = append(deletedLint, version)
					return "/tmp/fgm/golangci-lint/" + version, nil
				},
			},
			LintRemoteProvider: stubLintRemoteProvider{
				listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
					if goVersion == "1.25.7" {
						return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
					}
					return nil, nil
				},
			},
		}

		// Act
		_, _, err := testutil.ExecuteCommand(t, NewRootCmd(application), "prune")

		// Assert
		if err != nil {
			t.Fatalf("prune: %v", err)
		}
		if !slices.Equal(deletedLint, []string{"v1.64.8"}) {
			t.Fatalf("deletedLint = %v, want only [v1.64.8] (preserving v2.14.0 for repo and v2.11.2 for global Go 1.25.7)", deletedLint)
		}
	})
}
