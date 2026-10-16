// Package report builds the machine-readable `fgm current --json` documents.
package report

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/workspace"
)

// Resolver resolves the selection for a directory.
type Resolver interface {
	Current(ctx context.Context, workDir string) (app.Selection, error)
}

// GoLocator returns the installed binary of a Go version.
type GoLocator interface {
	GoBinaryPath(ctx context.Context, version string) (string, error)
}

// LintLocator returns the installed binary of a golangci-lint version.
type LintLocator interface {
	LintBinaryPath(ctx context.Context, version string) (string, error)
}

// Catalog reports offline golangci-lint/Go compatibility.
type Catalog interface {
	LintCompatibility(goVersion string, lintVersion string) (compatible bool, known bool, err error)
}

// Config configures a Service; stores and catalog are optional.
type Config struct {
	Resolver  Resolver
	GoStore   GoLocator
	LintStore LintLocator
	Catalog   Catalog
}

// Service builds CurrentReports.
type Service struct {
	config Config
}

// New constructs a Service.
func New(config Config) *Service {
	return &Service{config: config}
}

// Report resolves dir and describes the selection with paths relative to dir.
func (s *Service) Report(ctx context.Context, dir string) (app.CurrentReport, error) {
	report := app.CurrentReport{Schema: app.CurrentReportSchema, Dir: dir, Warnings: []string{}}
	if root, ok := workspace.ModuleRoot(dir); ok {
		report.ModuleRoot = root
	}

	selection, err := s.config.Resolver.Current(ctx, dir)
	if err != nil {
		return report, err
	}
	report.Warnings = append(report.Warnings, selection.Warnings...)

	goReport := &app.GoReport{
		Version:    selection.GoVersion,
		Source:     goSource(selection),
		SourceFile: relative(dir, selection.GoSourceFile),
		Base:       selection.GoBase,
	}
	if s.config.GoStore != nil {
		if path, err := s.config.GoStore.GoBinaryPath(ctx, selection.GoVersion); err == nil {
			goReport.Installed, goReport.Path = true, path
		}
	}
	report.Go = goReport

	if selection.LintVersion == "" {
		return report, nil
	}
	lintReport := &app.LintReport{
		Version:    selection.LintVersion,
		Source:     selection.LintSource,
		SourceFile: relative(dir, selection.LintSourceFile),
		SourceLine: selection.LintSourceLine,
	}
	if s.config.LintStore != nil {
		if path, err := s.config.LintStore.LintBinaryPath(ctx, selection.LintVersion); err == nil {
			lintReport.Installed, lintReport.Path = true, path
		}
	}
	if s.config.Catalog != nil {
		compatible, known, err := s.config.Catalog.LintCompatibility(selection.GoVersion, selection.LintVersion)
		if err != nil {
			return report, err
		}
		if known {
			lintReport.Compatible = &compatible
		}
	}
	report.GolangCILint = lintReport
	return report, nil
}

// ReportAll reports every module under dir: the use entries of the nearest go.work that lie
// under dir, otherwise every go.mod below dir. A module that fails to resolve carries Error.
func (s *Service) ReportAll(ctx context.Context, dir string) ([]app.CurrentReport, error) {
	modules, err := ModuleDirs(dir)
	if err != nil {
		return nil, err
	}
	reports := make([]app.CurrentReport, 0, len(modules))
	for _, module := range modules {
		report, err := s.Report(ctx, module)
		if err != nil {
			report = app.CurrentReport{Schema: app.CurrentReportSchema, Dir: module, ModuleRoot: report.ModuleRoot, Warnings: []string{}, Error: err.Error()}
		}
		reports = append(reports, report)
	}
	return reports, nil
}

// ModuleDirs enumerates the modules under dir in lexical order (see ReportAll). Entries that
// resolve outside dir, including through symlinks, are skipped.
func ModuleDirs(dir string) ([]string, error) {
	workRoot, ok := workspace.WorkRoot(dir)
	if !ok {
		return workspace.ListModuleDirs(dir)
	}
	uses, err := parseWorkUses(filepath.Join(workRoot, "go.work"))
	if err != nil {
		return nil, err
	}
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}
	var modules []string
	for _, use := range uses {
		module := use
		if !filepath.IsAbs(module) {
			module = filepath.Join(workRoot, module)
		}
		module = filepath.Clean(module)
		if !workspace.Contains(dir, module) {
			continue
		}
		realModule, err := filepath.EvalSymlinks(module)
		if err != nil || !workspace.Contains(realDir, realModule) {
			continue
		}
		modules = append(modules, module)
	}
	slices.Sort(modules)
	return slices.Compact(modules), nil
}

func parseWorkUses(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	var uses []string
	inBlock := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "//", 2)[0])
		switch {
		case inBlock && line == ")":
			inBlock = false
		case inBlock && line != "":
			uses = append(uses, strings.Trim(line, `"`))
		case line == "use (":
			inBlock = true
		case strings.HasPrefix(line, "use "):
			uses = append(uses, strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "use ")), `"`))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return uses, nil
}

func goSource(selection app.Selection) string {
	switch selection.GoSource {
	case "config", "latest-patch", "global", "":
		return selection.GoSource
	}
	return filepath.Base(selection.GoSource)
}

func relative(dir string, path string) string {
	if path == "" {
		return ""
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
