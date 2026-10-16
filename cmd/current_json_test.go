package cmd

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/kostikovk/fgm/internal/report"
	"github.com/kostikovk/fgm/internal/resolve"
	"github.com/kostikovk/fgm/internal/testutil"
)

var updateGolden = flag.Bool("update", false, "rewrite cmd/testdata golden files")

func newReportingApp() *app.App {
	resolver := currenttoolchain.New(currenttoolchain.Config{GoResolver: resolve.New(nil)})
	return &app.App{
		Resolver:        resolver,
		CurrentReporter: report.New(report.Config{Resolver: resolver}),
	}
}

func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

func assertGolden(t *testing.T, name string, root string, got string) {
	t.Helper()
	got = strings.ReplaceAll(got, root, "$ROOT")
	path := filepath.Join("testdata", "current", name)
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}
	if got != string(want) {
		t.Errorf("output mismatch for %s\n--- got\n%s\n--- want\n%s", name, got, want)
	}
}

func TestCurrentCommandMachineReadable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		golden   string
		files    map[string]string
		dir      string
		args     []string
		wantCode int
	}{
		{
			name:   "should print one JSON object when a single module is selected",
			golden: "single.json",
			files: map[string]string{
				"go.mod":    "module m\n\ngo 1.26.0\n",
				"Makefile":  "GOLANGCI_LINT_VERSION ?= v2.5.0\n",
				".fgm.toml": "[toolchain]\ngo = \"1.26.8\"\n",
			},
			args: []string{"current", "--json"},
		},
		{
			name:   "should print one entry per go.work use entry when a workspace is listed",
			golden: "workspace.json",
			files: map[string]string{
				"go.work":         "go 1.26.0\n\nuse (\n\t./svc/b\n\t./svc/a\n)\n",
				"svc/a/go.mod":    "module a\n\ngo 1.26.0\n",
				"svc/b/go.mod":    "module b\n\ngo 1.25.0\n",
				"svc/a/.fgm.toml": "[toolchain]\ngolangci_lint = \"v2.9.0\"\n",
				"unused/go.mod":   "module unused\n\ngo 1.24.0\n",
			},
			args: []string{"current", "--all", "--json"},
		},
		{
			name:   "should list nested modules and skip vendored ones when no workspace exists",
			golden: "nested.txt",
			files: map[string]string{
				"go.mod":            "module root\n\ngo 1.26.0\n",
				"tools/go.mod":      "module tools\n\ngo 1.25.0\n\ntoolchain go1.25.4\n",
				"vendor/x/go.mod":   "module x\n\ngo 1.20\n",
				"testdata/y/go.mod": "module y\n\ngo 1.20\n",
			},
			args: []string{"current", "--all"},
		},
		{
			name:   "should mark the failing module and exit 1 when one module cannot resolve",
			golden: "failing.json",
			files: map[string]string{
				"a/go.mod": "module a\n\ngo 1.26.0\n",
				"b/go.mod": "module b\n",
			},
			args:     []string{"current", "--all", "--json"},
			wantCode: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, tc.files)
			args := append(tc.args, "--chdir", filepath.Join(root, tc.dir))

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(newReportingApp()), args...)

			// Assert
			if code := ExitCode(err); code != tc.wantCode {
				t.Fatalf("exit code = %d (err %v); want %d", code, err, tc.wantCode)
			}
			assertGolden(t, tc.golden, root, stdout)
		})
	}
}

func TestCurrentCommandInvalidInvocation(t *testing.T) {
	t.Parallel()

	t.Run("RunE", func(t *testing.T) {
		t.Parallel()

		t.Run("should exit 2 when current receives a positional argument", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := NewRootCmd(newReportingApp())

			// Act
			_, _, err := testutil.ExecuteCommand(t, root, "current", "extra")

			// Assert
			if code := ExitCode(err); code != 2 {
				t.Errorf("exit code = %d; want 2", code)
			}
		})

		t.Run("should exit 2 when current receives an unknown flag", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := NewRootCmd(newReportingApp())

			// Act
			_, _, err := testutil.ExecuteCommand(t, root, "current", "--yaml")

			// Assert
			if code := ExitCode(err); code != 2 {
				t.Errorf("exit code = %d; want 2", code)
			}
		})
	})
}
