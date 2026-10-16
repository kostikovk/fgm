package lintdetect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTree(t *testing.T, files map[string]string) string {
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

func pinSummary(root string, pins []Pin) string {
	parts := make([]string, 0, len(pins))
	for _, pin := range pins {
		rel, _ := filepath.Rel(root, pin.File)
		parts = append(parts, fmt.Sprintf("%s@%s:%03d", pin.Version, filepath.ToSlash(rel), pin.Line))
	}
	return strings.Join(parts, ",")
}

const workflowAt = `name: ci
jobs:
  lint:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: %s
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v8
        with:
          version: v2.5.0
`

func TestDetect(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		files map[string]string
		dir   string
		want  string
	}{
		{
			name:  "should detect an asdf pin when .tool-versions lists golangci-lint",
			files: map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", ".tool-versions": "golang 1.26.0\ngolangci-lint 2.5.0\n"},
			want:  "v2.5.0@.tool-versions:002",
		},
		{
			name:  "should detect a mise pin when mise.toml lists golangci-lint under tools",
			files: map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", "mise.toml": "[env]\nX = \"1\"\n\n[tools]\ngo = \"1.26\"\ngolangci-lint = \"2.5.0\"\n"},
			want:  "v2.5.0@mise.toml:006",
		},
		{
			name:  "should detect a go.mod tool pin when the tool directive names golangci-lint",
			files: map[string]string{"go.mod": "module m\n\ngo 1.26.0\n\ntool github.com/golangci/golangci-lint/v2/cmd/golangci-lint\n\nrequire (\n\tgithub.com/golangci/golangci-lint/v2 v2.5.0 // indirect\n\tgolang.org/x/mod v0.20.0\n)\n"},
			want:  "v2.5.0@go.mod:008",
		},
		{
			name:  "should detect a workflow pin when the repository has a single module",
			files: map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", ".github/workflows/ci.yml": strings.Replace(workflowAt, "%s", ".", 1)},
			want:  "v2.5.0@.github/workflows/ci.yml:012",
		},
		{
			name: "should detect a workflow pin when the working directory equals the module path",
			files: map[string]string{
				"svc/a/go.mod":             "module a\n\ngo 1.26.0\n",
				"svc/b/go.mod":             "module b\n\ngo 1.26.0\n",
				".github/workflows/ci.yml": strings.Replace(workflowAt, "%s", "svc/a", 1),
			},
			dir:  "svc/a",
			want: "v2.5.0@.github/workflows/ci.yml:012",
		},
		{
			name: "should ignore a workflow pin when the working directory names another module",
			files: map[string]string{
				"svc/a/go.mod":             "module a\n\ngo 1.26.0\n",
				"svc/b/go.mod":             "module b\n\ngo 1.26.0\n",
				".github/workflows/ci.yml": strings.Replace(workflowAt, "%s", "./svc/b", 1),
			},
			dir:  "svc/a",
			want: "",
		},
		{
			name:  "should detect a Makefile version variable when it is assigned with ?=",
			files: map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", "Makefile": "BIN := bin\nGOLANGCI_LINT_VERSION ?= v2.5.0\n"},
			want:  "v2.5.0@Makefile:002",
		},
		{
			name:  "should detect a go install pin when a make include installs golangci-lint at a version",
			files: map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", "tools.mk": "lint-install:\n\tgo install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0\n"},
			want:  "v2.5.0@tools.mk:002",
		},
		{
			name:  "should detect an install.sh pin when the line mentions golangci-lint",
			files: map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", "Makefile": "lint:\n\tcurl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b bin v2.5.0\n"},
			want:  "v2.5.0@Makefile:002",
		},
		{
			name:  "should detect a root Makefile pin when the module lives in a subdirectory",
			files: map[string]string{"api/go.mod": "module m\n\ngo 1.26.0\n", "Makefile": "GOLANGCI_LINT_VERSION := 2.5.0\n"},
			dir:   "api",
			want:  "v2.5.0@Makefile:001",
		},
		{
			name: "should report every pin when sources disagree",
			files: map[string]string{
				"go.mod":         "module m\n\ngo 1.26.0\n",
				".tool-versions": "golangci-lint 2.6.0\n",
				"Makefile":       "GOLANGCI_LINT_VERSION = v2.5.0\n",
			},
			want: "v2.6.0@.tool-versions:001,v2.5.0@Makefile:001",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeTree(t, tc.files)

			// Act
			result, err := Detect(filepath.Join(root, tc.dir))

			// Assert
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			if got := pinSummary(root, result.Pins); got != tc.want {
				t.Errorf("pins = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestResult(t *testing.T) {
	t.Parallel()

	t.Run("Agreed", func(t *testing.T) {
		t.Parallel()

		t.Run("should return the first pin when all pins share one version", func(t *testing.T) {
			t.Parallel()

			// Arrange
			result := Result{Pins: []Pin{{Version: "v2.5.0", File: "a", Line: 1}, {Version: "v2.5.0", File: "b", Line: 2}}}

			// Act
			pin, ok := result.Agreed()

			// Assert
			if !ok || pin.File != "a" {
				t.Errorf("Agreed = %+v, %v; want the first pin", pin, ok)
			}
		})

		t.Run("should report no agreement when pins name different versions", func(t *testing.T) {
			t.Parallel()

			// Arrange
			result := Result{Pins: []Pin{{Version: "v2.5.0"}, {Version: "v2.6.0"}}}

			// Act
			_, ok := result.Agreed()

			// Assert
			if ok || !result.Conflict() {
				t.Errorf("Agreed ok = %v, Conflict = %v; want a conflict", ok, result.Conflict())
			}
		})
	})
}

func TestDetectSkipsBrokenWorkflow(t *testing.T) {
	t.Parallel()

	t.Run("Detect", func(t *testing.T) {
		t.Parallel()

		t.Run("should keep detecting other pins when a workflow file is not valid YAML", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeTree(t, map[string]string{
				"go.mod":                        "module m\n\ngo 1.26.0\n",
				".github/workflows/broken.yaml": "foo: [unclosed\n",
				"Makefile":                      "GOLANGCI_LINT_VERSION ?= v2.5.0\n",
			})

			// Act
			result, err := Detect(root)

			// Assert
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			if got := pinSummary(root, result.Pins); got != "v2.5.0@Makefile:001" {
				t.Errorf("pins = %q; want the Makefile pin", got)
			}
		})

		t.Run("should ignore commented-out pins when Makefile and mise comments mention versions", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeTree(t, map[string]string{
				"go.mod":    "module m\n\ngo 1.26.0\n",
				"Makefile":  "# go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.0\nlint:\n\tgolangci-lint run\n",
				"mise.toml": "[tools]\ngolangci-lint = \"latest\" # was 1.64.8\n",
			})

			// Act
			result, err := Detect(root)

			// Assert
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			if len(result.Pins) != 0 {
				t.Errorf("pins = %q; want none", pinSummary(root, result.Pins))
			}
		})
	})
}
