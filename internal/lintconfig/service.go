package lintconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
)

// Resolver provides the current toolchain selection for a workspace.
type Resolver interface {
	Current(ctx context.Context, workDir string) (app.Selection, error)
}

// Config configures a lint config Service.
type Config struct {
	Resolver Resolver
}

// Service implements LintConfigGenerator and LintDoctor.
type Service struct {
	resolver Resolver
	catalog  *Catalog
}

// New constructs a lint config Service.
func New(cfg Config) (*Service, error) {
	catalog, err := LoadCatalog()
	if err != nil {
		return nil, fmt.Errorf("load linter catalog: %w", err)
	}
	return &Service{
		resolver: cfg.Resolver,
		catalog:  catalog,
	}, nil
}

// Generate produces a golangci-lint v2 configuration file.
func (s *Service) Generate(ctx context.Context, opts app.LintConfigOptions) ([]byte, error) {
	selection, err := s.resolver.Current(ctx, opts.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("resolve toolchain: %w", err)
	}
	if selection.GoVersion == "" {
		return nil, fmt.Errorf("no Go version resolved for %s", opts.WorkDir)
	}
	if _, err := NormalizePreset(opts.Preset); err != nil {
		return nil, err
	}

	lintVersion := selection.LintVersion
	if lintVersion != "" && !strings.HasPrefix(lintVersion, "v2") {
		return nil, fmt.Errorf("golangci-lint %s is v1; fgm lint init requires v2.x", lintVersion)
	}
	if lintVersion == "" {
		lintVersion = "v2.0.0"
	}

	// Determine output path — next to go.mod or go.work.
	outputDir := opts.WorkDir
	if dir := findProjectRoot(opts.WorkDir); dir != "" {
		outputDir = dir
	}
	outputPath := filepath.Join(outputDir, ".golangci.yml")

	if !opts.Force {
		if _, err := os.Stat(outputPath); err == nil {
			return nil, fmt.Errorf(".golangci.yml already exists at %s; use --force to overwrite", outputDir)
		}
	}

	modulePath := FindModulePath(opts.WorkDir)

	data, err := Generate(s.catalog, GenerateOptions{
		GoVersion:   selection.GoVersion,
		LintVersion: lintVersion,
		Preset:      opts.Preset,
		WithImports: opts.WithImports,
		ModulePath:  modulePath,
	})
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	return data, nil
}

// Diagnose checks an existing golangci-lint config for issues.
func (s *Service) Diagnose(ctx context.Context, workDir string) ([]app.LintFinding, error) {
	selection, err := s.resolver.Current(ctx, workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve toolchain: %w", err)
	}

	findings, err := Diagnose(s.catalog, workDir, selection.GoVersion)
	if err != nil {
		return nil, err
	}

	result := make([]app.LintFinding, len(findings))
	for i, f := range findings {
		result[i] = app.LintFinding{
			Severity: f.Severity,
			Message:  f.Message,
		}
	}
	return result, nil
}

// findProjectRoot walks up from workDir looking for go.work or go.mod.
func findProjectRoot(workDir string) string {
	dir := workDir
	for {
		for _, name := range []string{"go.work", "go.mod"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
