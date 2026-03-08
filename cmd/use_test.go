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

func TestUseGoGlobal_SetsGlobalVersionAndMentionsShims(t *testing.T) {
	t.Parallel()

	var setGlobalVersion string

	application := &app.App{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
				return version == "1.25.7", nil
			},
			setGlobalGoVersionFn: func(ctx context.Context, version string) error {
				setGlobalVersion = version
				return nil
			},
			ensureShimsFn: func() error {
				return nil
			},
			shimDirFn: func() string {
				return "/tmp/fgm/shims"
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "use", "go", "1.25.7", "--global")
	if err != nil {
		t.Fatalf("execute use go --global: %v\nstderr:\n%s", err, stderr)
	}

	if setGlobalVersion != "1.25.7" {
		t.Fatalf("setGlobalVersion = %q, want %q", setGlobalVersion, "1.25.7")
	}
	if !strings.Contains(stdout, "Selected Go 1.25.7 as the global default.") {
		t.Fatalf("stdout = %q, want it to contain global selection message", stdout)
	}
	if !strings.Contains(stdout, "/tmp/fgm/shims") {
		t.Fatalf("stdout = %q, want it to contain shim dir", stdout)
	}
}

func TestUseGo_RejectsWithoutGlobalFlag(t *testing.T) {
	t.Parallel()

	application := &app.App{}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "use", "go", "1.25.7")
	if err == nil {
		t.Fatal("expected an error when --global flag is not set")
	}
	if !strings.Contains(err.Error(), "only --global is supported") {
		t.Fatalf("err = %q, want only --global error", err)
	}
}

func TestUseGoGlobal_RejectsNilGoStore(t *testing.T) {
	t.Parallel()

	application := &app.App{GoStore: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "use", "go", "1.25.7", "--global")
	if err == nil {
		t.Fatal("expected an error when GoStore is nil")
	}
	if !strings.Contains(err.Error(), "local Go version store is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestUseGoGlobal_HasGoVersionErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
				return false, fmt.Errorf("has version boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "use", "go", "1.25.7", "--global")
	if err == nil {
		t.Fatal("expected an error when HasGoVersion fails")
	}
	if !strings.Contains(err.Error(), "has version boom") {
		t.Fatalf("err = %q, want has version boom", err)
	}
}

func TestUseGoGlobal_SetGlobalGoVersionErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
				return true, nil
			},
			setGlobalGoVersionFn: func(ctx context.Context, version string) error {
				return fmt.Errorf("set global boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "use", "go", "1.25.7", "--global")
	if err == nil {
		t.Fatal("expected an error when SetGlobalGoVersion fails")
	}
	if !strings.Contains(err.Error(), "set global boom") {
		t.Fatalf("err = %q, want set global boom", err)
	}
}

func TestUseGoGlobal_EnsureShimsErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
				return true, nil
			},
			setGlobalGoVersionFn: func(ctx context.Context, version string) error {
				return nil
			},
			ensureShimsFn: func() error {
				return fmt.Errorf("ensure shims boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "use", "go", "1.25.7", "--global")
	if err == nil {
		t.Fatal("expected an error when EnsureShims fails")
	}
	if !strings.Contains(err.Error(), "ensure shims boom") {
		t.Fatalf("err = %q, want ensure shims boom", err)
	}
}

func TestUseGoGlobal_RejectsMissingVersion(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		GoStore: stubGoStore{
			hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
				return false, nil
			},
		},
	}

	root := NewRootCmd(application)
	_, stderr, err := testutil.ExecuteCommand(t, root, "use", "go", "1.25.7", "--global")
	if err == nil {
		t.Fatal("expected an error when the requested Go version is not installed")
	}

	if !strings.Contains(stderr, "go version 1.25.7 is not installed") {
		t.Fatalf("stderr = %q, want it to contain missing-version error", stderr)
	}
}
