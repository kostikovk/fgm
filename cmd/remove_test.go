package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

func TestRemoveGoCommand_RemovesManagedVersion(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			deleteGoVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "1.25.7" {
					t.Fatalf("version = %q, want %q", version, "1.25.7")
				}
				return "/tmp/fgm/go/1.25.7", nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "remove", "go", "1.25.7")
	if err != nil {
		t.Fatalf("execute remove go: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Removed Go 1.25.7") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Removed Go 1.25.7")
	}
}

func TestRemoveLintCommand_RemovesManagedVersion(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		LintStore: stubLintStore{
			deleteLintVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "v2.11.2" {
					t.Fatalf("version = %q, want %q", version, "v2.11.2")
				}
				return "/tmp/fgm/golangci-lint/v2.11.2", nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "remove", "golangci-lint", "v2.11.2")
	if err != nil {
		t.Fatalf("execute remove golangci-lint: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Removed golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Removed golangci-lint v2.11.2")
	}
}
