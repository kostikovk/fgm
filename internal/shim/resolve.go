package shim

import (
	"context"

	"github.com/kostikovk/fgm/internal/app"
)

// SelectionResolver resolves the Go version for a working directory.
type SelectionResolver interface {
	Current(ctx context.Context, workDir string) (app.Selection, error)
}

// GoBinaryLocator returns the executable path for a Go version.
type GoBinaryLocator interface {
	GoBinaryPath(ctx context.Context, version string) (string, error)
}

// Resolver maps the selected Go version to a concrete executable path.
type Resolver struct {
	selection SelectionResolver
	locator   GoBinaryLocator
}

// New constructs a shim Resolver.
func New(selection SelectionResolver, locator GoBinaryLocator) *Resolver {
	return &Resolver{
		selection: selection,
		locator:   locator,
	}
}

// ResolveGoBinary returns the Go binary path for the current directory.
func (r *Resolver) ResolveGoBinary(ctx context.Context, workDir string) (string, error) {
	selection, err := r.selection.Current(ctx, workDir)
	if err != nil {
		return "", err
	}

	return r.locator.GoBinaryPath(ctx, selection.GoVersion)
}
