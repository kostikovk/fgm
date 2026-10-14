package shim

import (
	"context"
	"fmt"
	"path/filepath"

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

// LintBinaryLocator returns the executable path for a golangci-lint version.
type LintBinaryLocator interface {
	LintBinaryPath(ctx context.Context, version string) (string, error)
}

// LintExecution holds the resolved golangci-lint binary path and optional Go bin directory.
type LintExecution struct {
	BinaryPath  string
	GoBinaryDir string
}

// Resolver maps the selected Go or golangci-lint version to a concrete executable path.
type Resolver struct {
	selection   SelectionResolver
	locator     GoBinaryLocator
	lintLocator LintBinaryLocator
}

// New constructs a shim Resolver for Go binary resolution.
func New(selection SelectionResolver, locator GoBinaryLocator) *Resolver {
	return &Resolver{
		selection: selection,
		locator:   locator,
	}
}

// NewWithLint constructs a shim Resolver for Go and golangci-lint binary resolution.
func NewWithLint(selection SelectionResolver, locator GoBinaryLocator, lintLocator LintBinaryLocator) *Resolver {
	return &Resolver{
		selection:   selection,
		locator:     locator,
		lintLocator: lintLocator,
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

// ResolveLint returns the golangci-lint binary path and paired Go bin directory for the current directory.
func (r *Resolver) ResolveLint(ctx context.Context, workDir string) (LintExecution, error) {
	if r.lintLocator == nil {
		return LintExecution{}, fmt.Errorf("lint binary locator is not configured")
	}

	selection, err := r.selection.Current(ctx, workDir)
	if err != nil {
		return LintExecution{}, err
	}
	if selection.LintVersion == "" {
		return LintExecution{}, fmt.Errorf("no compatible golangci-lint version found for Go %s", selection.GoVersion)
	}

	lintBinaryPath, err := r.lintLocator.LintBinaryPath(ctx, selection.LintVersion)
	if err != nil {
		return LintExecution{}, err
	}

	var goBinDir string
	if r.locator != nil && selection.GoVersion != "" {
		goBinaryPath, goErr := r.locator.GoBinaryPath(ctx, selection.GoVersion)
		if goErr != nil {
			return LintExecution{}, goErr
		}
		if goBinaryPath != "" {
			goBinDir = filepath.Dir(goBinaryPath)
		}
	}

	return LintExecution{
		BinaryPath:  lintBinaryPath,
		GoBinaryDir: goBinDir,
	}, nil
}
