package goimport

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
)

func writeSDK(t *testing.T, dir string, versionFile string, withBinary bool) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if withBinary {
		if err := os.WriteFile(filepath.Join(dir, "bin", goBinaryName()), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatalf("write go: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte(versionFile), 0o644); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}
}

func summarize(items []app.ImportedGo) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		state := "imported"
		if item.SkipReason != "" {
			state = "skipped"
		}
		parts = append(parts, item.Version+" "+state+" ("+item.Source+")")
	}
	return strings.Join(parts, "; ")
}

func TestImporterToolchainSources(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(t *testing.T, modCache string, sdkDir string)
		want  string
	}{
		{
			name: "should import a module-cache toolchain when bin/go and VERSION match its directory",
			setup: func(t *testing.T, modCache string, sdkDir string) {
				writeSDK(t, filepath.Join(modCache, "golang.org", "toolchain@v0.0.1-go1.26.8.linux-amd64"), "go1.26.8\ntime 2026-01-01\n", true)
			},
			want: "1.26.8 imported (module cache)",
		},
		{
			name: "should import a golang.org/dl SDK when bin/go and VERSION match its directory",
			setup: func(t *testing.T, modCache string, sdkDir string) {
				writeSDK(t, filepath.Join(sdkDir, "go1.25.3"), "go1.25.3\n", true)
			},
			want: "1.25.3 imported (golang.org/dl)",
		},
		{
			name: "should skip with a reason when VERSION does not match the directory name",
			setup: func(t *testing.T, modCache string, sdkDir string) {
				writeSDK(t, filepath.Join(sdkDir, "go1.25.3"), "go1.25.2\n", true)
			},
			want: "1.25.3 skipped (golang.org/dl)",
		},
		{
			name: "should skip with a reason when bin/go is missing",
			setup: func(t *testing.T, modCache string, sdkDir string) {
				writeSDK(t, filepath.Join(modCache, "golang.org", "toolchain@v0.0.1-go1.26.8.linux-amd64"), "go1.26.8\n", false)
			},
			want: "1.26.8 skipped (module cache)",
		},
		{
			name: "should ignore a module-cache toolchain when it targets another platform",
			setup: func(t *testing.T, modCache string, sdkDir string) {
				writeSDK(t, filepath.Join(modCache, "golang.org", "toolchain@v0.0.1-go1.26.8.darwin-arm64"), "go1.26.8\n", true)
			},
			want: "",
		},
		{
			name: "should skip the second copy when the same version exists in both sources",
			setup: func(t *testing.T, modCache string, sdkDir string) {
				writeSDK(t, filepath.Join(modCache, "golang.org", "toolchain@v0.0.1-go1.26.8.linux-amd64"), "go1.26.8\n", true)
				writeSDK(t, filepath.Join(sdkDir, "go1.26.8"), "go1.26.8\n", true)
			},
			want: "1.26.8 imported (module cache); 1.26.8 skipped (golang.org/dl)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			modCache, sdkDir := t.TempDir(), t.TempDir()
			tc.setup(t, modCache, sdkDir)
			registry := &stubRegistry{registered: make(map[string]string)}
			importer := New(Config{
				Registry: registry,
				ModCache: func(context.Context) string { return modCache },
				SDKDir:   sdkDir,
				GOOS:     "linux",
				GOARCH:   "amd64",
			})

			// Act
			got, err := importer.ImportAuto(context.Background())

			// Assert
			if err != nil {
				t.Fatalf("ImportAuto: %v", err)
			}
			if summarize(got) != tc.want {
				t.Errorf("result = %q; want %q", summarize(got), tc.want)
			}
		})
	}
}

func TestImporterToolchainSourcesIdempotent(t *testing.T) {
	t.Parallel()

	t.Run("ImportAuto", func(t *testing.T) {
		t.Parallel()

		t.Run("should register the same path once when imported twice", func(t *testing.T) {
			t.Parallel()

			// Arrange
			sdkDir := t.TempDir()
			writeSDK(t, filepath.Join(sdkDir, "go1.25.3"), "go1.25.3\n", true)
			registry := &stubRegistry{registered: make(map[string]string)}
			importer := New(Config{Registry: registry, SDKDir: sdkDir, GOOS: "linux", GOARCH: "amd64"})

			// Act
			_, firstErr := importer.ImportAuto(context.Background())
			_, secondErr := importer.ImportAuto(context.Background())

			// Assert
			if firstErr != nil || secondErr != nil {
				t.Fatalf("ImportAuto errors: %v, %v", firstErr, secondErr)
			}
			if len(registry.registered) != 1 || registry.registered["1.25.3"] != filepath.Join(sdkDir, "go1.25.3") {
				t.Errorf("registered = %v; want one entry for 1.25.3", registry.registered)
			}
		})
	})
}
