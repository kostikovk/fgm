package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

func TestRemoveGoCommand_RejectsNilGoStore(t *testing.T) {
	t.Parallel()

	application := &app.App{GoStore: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "remove", "go", "1.25.7")
	if err == nil {
		t.Fatal("expected an error when GoStore is nil")
	}
	if !strings.Contains(err.Error(), "local Go version store is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestRemoveLintCommand_RejectsNilLintStore(t *testing.T) {
	t.Parallel()

	application := &app.App{LintStore: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "remove", "golangci-lint", "v2.11.2")
	if err == nil {
		t.Fatal("expected an error when LintStore is nil")
	}
	if !strings.Contains(err.Error(), "local golangci-lint version store is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestRemoveGoCommand_DeleteErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoStore: stubGoStore{
			deleteGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("delete go boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "remove", "go", "1.25.7")
	if err == nil {
		t.Fatal("expected an error when DeleteGoVersion fails")
	}
	if !strings.Contains(err.Error(), "delete go boom") {
		t.Fatalf("err = %q, want delete go boom", err)
	}
}

func TestRemoveLintCommand_DeleteErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
				return nil, nil
			},
			deleteLintVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("delete lint boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "remove", "golangci-lint", "v2.11.2")
	if err == nil {
		t.Fatal("expected an error when DeleteLintVersion fails")
	}
	if !strings.Contains(err.Error(), "delete lint boom") {
		t.Fatalf("err = %q, want delete lint boom", err)
	}
}

func TestRemoveGoCommand_RemovesManagedVersion(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			deleteGoVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "1.25.7" {
					t.Fatalf("version = %q, want %q", version, "1.25.7")
				}
				return "/tmp/fgm/go/1.25.7", nil
			},
		},
	}

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

	application := &app.App{
		Resolver: resolve.New(nil),
		LintStore: stubLintStore{
			deleteLintVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "v2.11.2" {
					t.Fatalf("version = %q, want %q", version, "v2.11.2")
				}
				return "/tmp/fgm/golangci-lint/v2.11.2", nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "remove", "golangci-lint", "v2.11.2")
	if err != nil {
		t.Fatalf("execute remove golangci-lint: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Removed golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Removed golangci-lint v2.11.2")
	}
}
