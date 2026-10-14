package shim

import (
	"context"
	"fmt"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
)

type stubSelectionResolver struct {
	currentFn func(ctx context.Context, workDir string) (app.Selection, error)
}

func (s stubSelectionResolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	return s.currentFn(ctx, workDir)
}

type stubBinaryLocator struct {
	goBinaryPathFn func(ctx context.Context, version string) (string, error)
}

func (s stubBinaryLocator) GoBinaryPath(ctx context.Context, version string) (string, error) {
	return s.goBinaryPathFn(ctx, version)
}

func TestResolveGoBinary_UsesSelectedVersion(t *testing.T) {
	t.Parallel()

	resolver := New(stubSelectionResolver{
		currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
			return app.Selection{GoVersion: "1.25.7", GoSource: "global"}, nil
		},
	}, stubBinaryLocator{
		goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			if version != "1.25.7" {
				t.Fatalf("version = %q, want %q", version, "1.25.7")
			}
			return "/tmp/fgm/go/1.25.7/bin/go", nil
		},
	})

	path, err := resolver.ResolveGoBinary(context.Background(), "/tmp/project")
	if err != nil {
		t.Fatalf("ResolveGoBinary: %v", err)
	}

	if path != "/tmp/fgm/go/1.25.7/bin/go" {
		t.Fatalf("path = %q, want %q", path, "/tmp/fgm/go/1.25.7/bin/go")
	}
}

func TestResolveGoBinary_ErrorWhenSelectionFails(t *testing.T) {
	t.Parallel()

	resolver := New(stubSelectionResolver{
		currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
			return app.Selection{}, fmt.Errorf("selection broken")
		},
	}, stubBinaryLocator{
		goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			t.Fatalf("locator should not be called when selection fails")
			return "", nil
		},
	})

	_, err := resolver.ResolveGoBinary(context.Background(), "/tmp/project")
	if err == nil {
		t.Fatalf("expected error from selection resolver, got nil")
	}
}

func TestResolveGoBinary_ErrorWhenLocatorFails(t *testing.T) {
	t.Parallel()

	resolver := New(stubSelectionResolver{
		currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
			return app.Selection{GoVersion: "1.25.7", GoSource: "global"}, nil
		},
	}, stubBinaryLocator{
		goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
			return "", fmt.Errorf("locator broken")
		},
	})

	_, err := resolver.ResolveGoBinary(context.Background(), "/tmp/project")
	if err == nil {
		t.Fatalf("expected error from locator, got nil")
	}
}

type stubLintLocator struct {
	lintBinaryPathFn func(ctx context.Context, version string) (string, error)
}

func (s stubLintLocator) LintBinaryPath(ctx context.Context, version string) (string, error) {
	return s.lintBinaryPathFn(ctx, version)
}

func TestResolveLint_UsesSelectedLintAndGoBinaryDir(t *testing.T) {
	t.Parallel()

	t.Run("should return lint binary path and selected Go bin directory when both toolchains are resolved", func(t *testing.T) {
		t.Parallel()

		// Arrange
		resolver := NewWithLint(
			stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						GoSource:    "go.mod",
						LintVersion: "v2.14.0",
						LintSource:  "local",
					}, nil
				},
			},
			stubBinaryLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/go/1.27.2/bin/go", nil
				},
			},
			stubLintLocator{
				lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					if version != "v2.14.0" {
						t.Fatalf("version = %q, want v2.14.0", version)
					}
					return "/tmp/fgm/golangci-lint/v2.14.0/golangci-lint", nil
				},
			},
		)

		// Act
		execInfo, err := resolver.ResolveLint(context.Background(), "/tmp/project")

		// Assert
		if err != nil {
			t.Fatalf("ResolveLint: %v", err)
		}
		if execInfo.BinaryPath != "/tmp/fgm/golangci-lint/v2.14.0/golangci-lint" {
			t.Fatalf("BinaryPath = %q, want %q", execInfo.BinaryPath, "/tmp/fgm/golangci-lint/v2.14.0/golangci-lint")
		}
		if execInfo.GoBinaryDir != "/tmp/fgm/go/1.27.2/bin" {
			t.Fatalf("GoBinaryDir = %q, want %q", execInfo.GoBinaryDir, "/tmp/fgm/go/1.27.2/bin")
		}
	})
}

func TestResolveLint_ErrorWhenNoCompatibleLintVersion(t *testing.T) {
	t.Parallel()

	t.Run("should return an error when selection has no compatible golangci-lint version", func(t *testing.T) {
		t.Parallel()

		// Arrange
		resolver := NewWithLint(
			stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2"}, nil
				},
			},
			stubBinaryLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/go/1.27.2/bin/go", nil
				},
			},
			stubLintLocator{
				lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "", nil
				},
			},
		)

		// Act
		_, err := resolver.ResolveLint(context.Background(), "/tmp/project")

		// Assert
		if err == nil {
			t.Fatal("expected error when LintVersion is empty, got nil")
		}
	})
}

func TestResolveLint_ErrorWhenLintLocatorFails(t *testing.T) {
	t.Parallel()

	t.Run("should return an error when lint binary locator fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		resolver := NewWithLint(
			stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2", LintVersion: "v2.14.0"}, nil
				},
			},
			stubBinaryLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/go/1.27.2/bin/go", nil
				},
			},
			stubLintLocator{
				lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "", fmt.Errorf("lint not installed")
				},
			},
		)

		// Act
		_, err := resolver.ResolveLint(context.Background(), "/tmp/project")

		// Assert
		if err == nil {
			t.Fatal("expected error when lint locator fails, got nil")
		}
	})
}

func TestResolveLint_ErrorWhenGoLocatorFails(t *testing.T) {
	t.Parallel()

	t.Run("should return an error when selected Go version is not installed", func(t *testing.T) {
		t.Parallel()

		// Arrange
		resolver := NewWithLint(
			stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2", LintVersion: "v2.14.0"}, nil
				},
			},
			stubBinaryLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "", fmt.Errorf("go version %s is not installed", version)
				},
			},
			stubLintLocator{
				lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/golangci-lint/v2.14.0/golangci-lint", nil
				},
			},
		)

		// Act
		_, err := resolver.ResolveLint(context.Background(), "/tmp/project")

		// Assert
		if err == nil {
			t.Fatal("expected error when GoBinaryPath fails, got nil")
		}
	})
}
