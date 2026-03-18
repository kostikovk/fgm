package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/resolve"
	"github.com/kostikovk/fgm/internal/testutil"
)

type stubGoImporter struct {
	importAutoFn func(ctx context.Context) ([]app.ImportedGo, error)
}

func (s stubGoImporter) ImportAuto(ctx context.Context) ([]app.ImportedGo, error) {
	return s.importAutoFn(ctx)
}

type stubLintImporter struct {
	importAutoFn func(ctx context.Context) ([]app.ImportedLint, error)
}

func (s stubLintImporter) ImportAuto(ctx context.Context) ([]app.ImportedLint, error) {
	return s.importAutoFn(ctx)
}

func TestImportAuto_RejectsNilImporters(t *testing.T) {
	t.Parallel()

	application := &app.App{GoImporter: nil, LintImporter: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "import", "auto")
	if err == nil {
		t.Fatal("expected an error when both importers are nil")
	}
	if !strings.Contains(err.Error(), "importers are not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestImportAuto_GoImporterErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoImporter: stubGoImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
				return nil, fmt.Errorf("go import boom")
			},
		},
		LintImporter: stubLintImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedLint, error) {
				return nil, nil
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "import", "auto")
	if err == nil {
		t.Fatal("expected an error when GoImporter fails")
	}
	if !strings.Contains(err.Error(), "go import boom") {
		t.Fatalf("err = %q, want go import boom", err)
	}
}

func TestImportAuto_LintImporterErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoImporter: stubGoImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
				return nil, nil
			},
		},
		LintImporter: stubLintImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedLint, error) {
				return nil, fmt.Errorf("lint import boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "import", "auto")
	if err == nil {
		t.Fatal("expected an error when LintImporter fails")
	}
	if !strings.Contains(err.Error(), "lint import boom") {
		t.Fatalf("err = %q, want lint import boom", err)
	}
}

func TestImportAuto_ReportsNothingImported(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoImporter: stubGoImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
				return nil, nil
			},
		},
		LintImporter: stubLintImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedLint, error) {
				return nil, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "import", "auto")
	if err != nil {
		t.Fatalf("execute import auto: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "No Go or golangci-lint installations were imported") {
		t.Fatalf("stdout = %q, want nothing imported message", stdout)
	}
}

func TestImportAuto_WorksWithOnlyGoImporter(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoImporter: stubGoImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
				return []app.ImportedGo{{Version: "1.25.7", Path: "/usr/local/go"}}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "import", "auto")
	if err != nil {
		t.Fatalf("execute import auto: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "Imported Go 1.25.7") {
		t.Fatalf("stdout = %q, want Go import line", stdout)
	}
}

func TestImportAuto_WorksWithOnlyLintImporter(t *testing.T) {
	t.Parallel()

	application := &app.App{
		LintImporter: stubLintImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedLint, error) {
				return []app.ImportedLint{{Version: "v2.11.2", Path: "/usr/local/bin/golangci-lint"}}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "import", "auto")
	if err != nil {
		t.Fatalf("execute import auto: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "Imported golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want lint import line", stdout)
	}
}

func TestImportAuto_ReportsImportedVersions(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		GoImporter: stubGoImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
				return []app.ImportedGo{
					{Version: "1.26.1", Path: "/usr/local/go"},
					{Version: "1.25.7", Path: "/opt/homebrew/Cellar/go/1.25.7/libexec"},
				}, nil
			},
		},
		LintImporter: stubLintImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedLint, error) {
				return []app.ImportedLint{
					{Version: "v2.11.2", Path: "/opt/homebrew/bin/golangci-lint"},
				}, nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "import", "auto")
	if err != nil {
		t.Fatalf("execute import auto: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Imported Go 1.26.1 from /usr/local/go") {
		t.Fatalf("stdout = %q, want it to contain first import line", stdout)
	}
	if !strings.Contains(stdout, "Imported Go 1.25.7 from /opt/homebrew/Cellar/go/1.25.7/libexec") {
		t.Fatalf("stdout = %q, want it to contain second import line", stdout)
	}
	if !strings.Contains(stdout, "Imported golangci-lint v2.11.2 from /opt/homebrew/bin/golangci-lint") {
		t.Fatalf("stdout = %q, want it to contain lint import line", stdout)
	}
}
