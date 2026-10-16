package report

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
)

type stubResolver struct {
	currentFn func(ctx context.Context, dir string) (app.Selection, error)
}

func (s stubResolver) Current(ctx context.Context, dir string) (app.Selection, error) {
	return s.currentFn(ctx, dir)
}

type stubStores struct{ installed map[string]bool }

func (s stubStores) GoBinaryPath(ctx context.Context, version string) (string, error) {
	if s.installed["go"+version] {
		return "/store/go/" + version + "/bin/go", nil
	}
	return "", os.ErrNotExist
}

func (s stubStores) LintBinaryPath(ctx context.Context, version string) (string, error) {
	if s.installed[version] {
		return "/store/golangci-lint/" + version + "/golangci-lint", nil
	}
	return "", os.ErrNotExist
}

type stubCatalog struct{}

func (stubCatalog) LintCompatibility(goVersion string, lintVersion string) (bool, bool, error) {
	return lintVersion == "v2.9.0", lintVersion != "v9.9.9", nil
}

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	return root
}

func TestService(t *testing.T) {
	t.Parallel()

	t.Run("Report", func(t *testing.T) {
		t.Parallel()

		t.Run("should describe sources relative to the directory with install and compatibility state", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeTree(t, map[string]string{"go.mod": "module m\n\ngo 1.26.0\n"})
			dir := filepath.Join(root, "pkg")
			service := New(Config{
				Resolver: stubResolver{currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion: "1.26.9", GoSource: "latest-patch", GoSourceFile: filepath.Join(root, "go.mod"), GoBase: "1.26.0",
						LintVersion: "v2.9.0", LintSource: "detected", LintSourceFile: filepath.Join(root, "Makefile"), LintSourceLine: 31,
					}, nil
				}},
				GoStore:   stubStores{installed: map[string]bool{"go1.26.9": true}},
				LintStore: stubStores{},
				Catalog:   stubCatalog{},
			})

			// Act
			got, err := service.Report(context.Background(), dir)

			// Assert
			if err != nil {
				t.Fatalf("Report: %v", err)
			}
			if got.ModuleRoot != root || got.Dir != dir || got.Schema != 1 {
				t.Errorf("header = %+v; want dir and module root", got)
			}
			if *got.Go != (app.GoReport{Version: "1.26.9", Source: "latest-patch", SourceFile: "../go.mod", Base: "1.26.0", Installed: true, Path: "/store/go/1.26.9/bin/go"}) {
				t.Errorf("go = %+v", *got.Go)
			}
			lint := got.GolangCILint
			if lint.SourceFile != "../Makefile" || lint.SourceLine != 31 || lint.Installed || lint.Compatible == nil || !*lint.Compatible {
				t.Errorf("golangci_lint = %+v", *lint)
			}
			if got.Warnings == nil {
				t.Error("warnings = nil; want an empty list")
			}
		})
	})

	t.Run("ReportAll", func(t *testing.T) {
		t.Parallel()

		t.Run("should list go.work use entries in lexical order when a workspace exists", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeTree(t, map[string]string{
				"go.work":      "go 1.26.0\n\nuse (\n\t./svc/b\n\t./svc/a\n)\n",
				"svc/a/go.mod": "module a\n",
				"svc/b/go.mod": "module b\n",
				"other/go.mod": "module other\n",
			})
			service := New(Config{Resolver: stubResolver{currentFn: func(ctx context.Context, dir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.26.0", GoSource: filepath.Join(root, "go.work"), GoSourceFile: filepath.Join(root, "go.work")}, nil
			}}})

			// Act
			reports, err := service.ReportAll(context.Background(), root)

			// Assert
			if err != nil {
				t.Fatalf("ReportAll: %v", err)
			}
			if len(reports) != 2 || reports[0].Dir != filepath.Join(root, "svc/a") || reports[1].Dir != filepath.Join(root, "svc/b") {
				t.Errorf("dirs = %v; want svc/a, svc/b", dirsOf(reports))
			}
		})

		t.Run("should record the error on the failing module entry when one module cannot resolve", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeTree(t, map[string]string{"a/go.mod": "module a\n", "b/go.mod": "module b\n"})
			service := New(Config{Resolver: stubResolver{currentFn: func(ctx context.Context, dir string) (app.Selection, error) {
				if filepath.Base(dir) == "b" {
					return app.Selection{}, errors.New("go directive not found")
				}
				return app.Selection{GoVersion: "1.26.0", GoSource: "go.mod"}, nil
			}}})

			// Act
			reports, err := service.ReportAll(context.Background(), root)

			// Assert
			if err != nil {
				t.Fatalf("ReportAll: %v", err)
			}
			if len(reports) != 2 || reports[0].Error != "" || reports[1].Error != "go directive not found" || reports[1].Go != nil {
				t.Errorf("reports = %+v; want the error on b only", reports)
			}
		})
	})
}

func dirsOf(reports []app.CurrentReport) []string {
	dirs := make([]string, 0, len(reports))
	for _, report := range reports {
		dirs = append(dirs, report.Dir)
	}
	return dirs
}
