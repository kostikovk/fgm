package shim

import (
	"context"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
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
