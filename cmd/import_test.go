package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

type stubGoImporter struct {
	importAutoFn func(ctx context.Context) ([]app.ImportedGo, error)
}

func (s stubGoImporter) ImportAuto(ctx context.Context) ([]app.ImportedGo, error) {
	return s.importAutoFn(ctx)
}

func TestImportAuto_ReportsImportedVersions(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		GoImporter: stubGoImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
				return []app.ImportedGo{
					{Version: "1.26.1", Path: "/usr/local/go"},
					{Version: "1.25.7", Path: "/opt/homebrew/Cellar/go/1.25.7/libexec"},
				}, nil
			},
		},
	})

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
}
