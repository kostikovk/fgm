package repocheck

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/kostikovk/fgm/internal/report"
	"github.com/kostikovk/fgm/internal/resolve"
)

// catalog marks v1.x linters as unable to run on Go 1.26+ and knows nothing about v9.x.
type catalog struct{}

func (catalog) LintCompatibility(goVersion string, lintVersion string) (bool, bool, error) {
	switch {
	case strings.HasPrefix(lintVersion, "v9."):
		return false, false, nil
	case strings.HasPrefix(lintVersion, "v1."):
		return !strings.HasPrefix(goVersion, "1.26") && !strings.HasPrefix(goVersion, "1.27"), true, nil
	}
	return true, true, nil
}

func newService() *Service {
	resolver := currenttoolchain.New(currenttoolchain.Config{GoResolver: resolve.New(nil)})
	return New(Config{
		Reporter: report.New(report.Config{Resolver: resolver, Catalog: catalog{}}),
	})
}

func writeRepo(t *testing.T, files map[string]string) string {
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

func setupGoWorkflow(goVersion string) string {
	return fmt.Sprintf("jobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/setup-go@v5\n        with:\n          go-version: %s\n", goVersion)
}

const goMod126 = "module m\n\ngo 1.26.0\n"

func TestServiceCheck(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		files map[string]string
		dir   string
		want  string // "ID SEVERITY file:line", or "" for no findings
	}{
		{
			name:  "should report C1 when the pinned linter cannot run on the resolved Go",
			files: map[string]string{"go.mod": goMod126, ".fgm.toml": "[toolchain]\ngolangci_lint = \"v1.64.8\"\n"},
			want:  "C1 ERROR .fgm.toml:2",
		},
		{
			name:  "should not report C1 when the pinned linter supports the resolved Go",
			files: map[string]string{"go.mod": goMod126, ".fgm.toml": "[toolchain]\ngolangci_lint = \"v2.9.0\"\n", ".golangci.yml": "version: \"2\"\n"},
		},
		{
			name:  "should report C2 when a v2 linter reads a config without version 2",
			files: map[string]string{"go.mod": goMod126, ".fgm.toml": "[toolchain]\ngolangci_lint = \"v2.9.0\"\n", ".golangci.yml": "linters:\n  enable: [govet]\n"},
			want:  "C2 ERROR .golangci.yml:1",
		},
		{
			name:  "should report C2 when a v1 linter reads a version 2 config",
			files: map[string]string{"go.mod": "module m\n\ngo 1.24.0\n", ".fgm.toml": "[toolchain]\ngolangci_lint = \"v1.64.8\"\n", ".golangci.yaml": "run:\n  timeout: 5m\nversion: \"2\"\n"},
			want:  "C2 ERROR .golangci.yaml:3",
		},
		{
			name:  "should report C3 when .fgm.toml and a Makefile pin different linters",
			files: map[string]string{"go.mod": goMod126, ".fgm.toml": "[toolchain]\ngolangci_lint = \"v2.9.0\"\n", ".golangci.yml": "version: \"2\"\n", "Makefile": "GOLANGCI_LINT_VERSION ?= v2.5.0\n"},
			want:  "C3 ERROR .fgm.toml:2",
		},
		{
			name:  "should not report C3 when every pin names the same linter",
			files: map[string]string{"go.mod": goMod126, ".fgm.toml": "[toolchain]\ngolangci_lint = \"v2.9.0\"\n", ".golangci.yml": "version: \"2\"\n", "Makefile": "GOLANGCI_LINT_VERSION ?= v2.9.0\n"},
		},
		{
			name:  "should report C4 when setup-go uses a lower minor than the go directive",
			files: map[string]string{"go.mod": goMod126, ".github/workflows/ci.yml": setupGoWorkflow(`"1.25"`)},
			want:  "C4 ERROR .github/workflows/ci.yml:7",
		},
		{
			name:  "should report C4b when setup-go uses a different minor than the resolved Go",
			files: map[string]string{"go.mod": goMod126, ".github/workflows/ci.yml": setupGoWorkflow(`"1.27.x"`)},
			want:  "C4b WARN .github/workflows/ci.yml:7",
		},
		{
			name:  "should not report C4 when setup-go reads go-version-file",
			files: map[string]string{"go.mod": goMod126, ".github/workflows/ci.yml": "jobs:\n  t:\n    steps:\n      - uses: actions/setup-go@v5\n        with:\n          go-version-file: go.mod\n"},
		},
		{
			name:  "should report C5 when a Dockerfile builds with an older golang image",
			files: map[string]string{"go.mod": goMod126, "Dockerfile": "FROM golang:1.25-bookworm AS build\nRUN go build ./...\n"},
			want:  "C5 ERROR Dockerfile:1",
		},
		{
			name:  "should not report C5 when a floating tag covers the go directive minor",
			files: map[string]string{"go.mod": goMod126, "Dockerfile": "FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.26-alpine\n"},
		},
		{
			name:  "should report C5 as INFO when the golang tag comes from an ARG",
			files: map[string]string{"go.mod": goMod126, "Dockerfile": "ARG GO_VERSION=1.26\nFROM golang:${GO_VERSION}\n"},
			want:  "C5 INFO Dockerfile:2",
		},
		{
			name: "should report C6 when go.work declares a lower go than a used module",
			files: map[string]string{
				"go.work":    "go 1.25.0\n\nuse ./svc\n",
				"svc/go.mod": goMod126,
			},
			dir:  "svc",
			want: "C6 ERROR ../go.work:1",
		},
		{
			name: "should report C6 when go.work declares a lower go than a module's release candidate directive",
			files: map[string]string{
				"go.work":    "go 1.26.0\n\nuse ./svc\n",
				"svc/go.mod": "module m\n\ngo 1.27rc1\n",
			},
			dir:  "svc",
			want: "C6 ERROR ../go.work:1",
		},
		{
			name:  "should report C5 when a full golang tag is older than a release candidate go directive",
			files: map[string]string{"go.mod": "module m\n\ngo 1.27rc1\n", "Dockerfile": "FROM golang:1.26.3\n"},
			want:  "C5 ERROR Dockerfile:1",
		},
		{
			name:  "should skip a check when .fgm.toml lists it under check.skip",
			files: map[string]string{"go.mod": goMod126, "Dockerfile": "FROM golang:1.25\n", ".fgm.toml": "[check]\nskip = [\"C5\"]\n"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeRepo(t, tc.files)

			// Act
			result, err := newService().Check(context.Background(), filepath.Join(root, tc.dir))

			// Assert
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			var got []string
			for _, finding := range result.Findings {
				got = append(got, fmt.Sprintf("%s %s %s:%d", finding.ID, finding.Severity, finding.File, finding.Line))
			}
			if strings.Join(got, "; ") != tc.want {
				t.Errorf("findings = %q; want %q\n%+v", strings.Join(got, "; "), tc.want, result.Findings)
			}
		})
	}
}

func TestServiceCheckDigestPinnedImage(t *testing.T) {
	t.Parallel()

	t.Run("Check", func(t *testing.T) {
		t.Parallel()

		t.Run("should report C5 when an older golang image is pinned by digest", func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeRepo(t, map[string]string{"go.mod": goMod126, "Dockerfile": "FROM golang:1.24@sha256:0123456789abcdef AS build\n"})

			// Act
			result, err := newService().Check(context.Background(), root)

			// Assert
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			if len(result.Findings) != 1 || result.Findings[0].ID != "C5" {
				t.Errorf("findings = %+v; want one C5", result.Findings)
			}
		})
	})
}
