package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/kostikovk/fgm/internal/repocheck"
	"github.com/kostikovk/fgm/internal/report"
	"github.com/kostikovk/fgm/internal/resolve"
	"github.com/kostikovk/fgm/internal/testutil"
)

func newCheckingApp() *app.App {
	resolver := currenttoolchain.New(currenttoolchain.Config{GoResolver: resolve.New(nil)})
	return &app.App{RepoChecker: repocheck.New(repocheck.Config{Reporter: report.New(report.Config{Resolver: resolver})})}
}

const checkWorkflow = "jobs:\n  test:\n    steps:\n      - uses: actions/setup-go@v5\n        with:\n          go-version: \"1.27\"\n"

func TestCheckCommand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		golden     string
		files      map[string]string
		args       []string
		wantCode   int
		wantStdout string
	}{
		{
			name:   "should print the envelope with findings and exit 1 when an error finding exists",
			golden: "check-error.json",
			files: map[string]string{
				"go.mod":     "module m\n\ngo 1.26.0\n",
				"Dockerfile": "FROM golang:1.25\n",
			},
			args:     []string{"check", "--json"},
			wantCode: 1,
		},
		{
			name:       "should exit 0 when only warnings exist and fail-on is error",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", ".github/workflows/ci.yml": checkWorkflow},
			args:       []string{"check"},
			wantStdout: "C4b WARN .github/workflows/ci.yml:6",
		},
		{
			name:     "should exit 1 when a warning exists and fail-on is warn",
			files:    map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", ".github/workflows/ci.yml": checkWorkflow},
			args:     []string{"check", "--fail-on", "warn"},
			wantCode: 1,
		},
		{
			name:       "should report no findings and exit 0 when the repository is consistent",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", "Dockerfile": "FROM golang:1.26\n"},
			args:       []string{"check"},
			wantStdout: "OK no consistency findings",
		},
		{
			name:     "should exit 2 when fail-on names an unknown severity",
			files:    map[string]string{"go.mod": "module m\n\ngo 1.26.0\n"},
			args:     []string{"check", "--fail-on", "info"},
			wantCode: 2,
		},
		{
			name:     "should exit 2 when a module cannot be resolved",
			files:    map[string]string{"go.mod": "module m\n"},
			args:     []string{"check"},
			wantCode: 2,
		},
		{
			name:       "should prefix each module when all modules are checked",
			files:      map[string]string{"a/go.mod": "module a\n\ngo 1.26.0\n", "b/go.mod": "module b\n\ngo 1.26.0\n", "b/Dockerfile": "FROM golang:1.25\n"},
			args:       []string{"check", "--all"},
			wantCode:   1,
			wantStdout: "a: OK no consistency findings\nb: C5 ERROR b/Dockerfile:1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, tc.files)
			args := append(tc.args, "--chdir", root)

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(newCheckingApp()), args...)

			// Assert
			if code := ExitCode(err); code != tc.wantCode {
				t.Fatalf("exit code = %d (err %v); want %d\nstdout:\n%s", code, err, tc.wantCode, stdout)
			}
			if tc.golden != "" {
				assertGolden(t, tc.golden, root, stdout)
			}
			if tc.wantStdout != "" && !strings.Contains(stdout, tc.wantStdout) {
				t.Errorf("stdout = %q; want it to contain %q", stdout, tc.wantStdout)
			}
		})
	}
}

func TestCheckCommandSkip(t *testing.T) {
	t.Parallel()

	t.Run("RunE", func(t *testing.T) {
		t.Parallel()

		t.Run("should exit 0 when the only failing check is skipped in .fgm.toml", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, map[string]string{
				"go.mod":     "module m\n\ngo 1.26.0\n",
				"Dockerfile": "FROM golang:1.25\n",
				".fgm.toml":  "[check]\nskip = [\"C5\"]\n",
			})

			// Act
			_, _, err := testutil.ExecuteCommand(t, NewRootCmd(newCheckingApp()), "check", "--chdir", filepath.Join(root, "."))

			// Assert
			if code := ExitCode(err); code != 0 {
				t.Errorf("exit code = %d (err %v); want 0", code, err)
			}
		})
	})
}

func TestCheckCommandToolBypass(t *testing.T) {
	t.Parallel()

	files := map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", "Makefile": "lint:\n\t$(shell go env GOPATH)/bin/golangci-lint run ./...\n"}

	t.Run("RunE", func(t *testing.T) {
		t.Parallel()

		t.Run("should exit 0 when only a C7 warning exists and fail-on is error", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, files)

			// Act
			stdout, _, err := testutil.ExecuteCommand(t, NewRootCmd(newCheckingApp()), "check", "--chdir", root)

			// Assert
			if code := ExitCode(err); code != 0 || !strings.Contains(stdout, "C7 WARN Makefile:2") {
				t.Errorf("exit code = %d, stdout = %q; want 0 and a C7 warning", code, stdout)
			}
		})

		t.Run("should exit 1 when a C7 warning exists and fail-on is warn", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeFixture(t, files)

			// Act
			_, _, err := testutil.ExecuteCommand(t, NewRootCmd(newCheckingApp()), "check", "--fail-on", "warn", "--chdir", root)

			// Assert
			if code := ExitCode(err); code != 1 {
				t.Errorf("exit code = %d; want 1", code)
			}
		})
	})
}
