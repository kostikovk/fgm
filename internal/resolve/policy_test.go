package resolve

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

type stubPatchCatalog map[string]string

func (s stubPatchCatalog) LatestGoPatch(goVersion string) (string, bool, error) {
	version, ok := s[minorKey(goVersion)]
	return version, ok, nil
}

type stubVersionList struct {
	versions []string
	err      error
}

func (s stubVersionList) ListLocalGoVersions(ctx context.Context) ([]string, error) {
	return s.versions, s.err
}

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}

func TestResolverGoPolicy(t *testing.T) {
	t.Parallel()

	catalog := stubPatchCatalog{"1.19": "1.19.13", "1.26": "1.26.9"}

	cases := []struct {
		name       string
		files      map[string]string
		dir        string
		config     Config
		wantGo     string
		wantSource string
		wantFile   string
		wantBase   string
		wantWarn   bool
		wantErr    bool
	}{
		{
			name:       "should select the explicit version when .fgm.toml pins go",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0\n\ntoolchain go1.26.3\n", ".fgm.toml": "[toolchain]\ngo = \"1.26.8\"\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.26.8",
			wantSource: "config",
			wantFile:   ".fgm.toml",
		},
		{
			name:    "should return an error when the explicit version is lower than the go directive",
			files:   map[string]string{"go.mod": "module m\n\ngo 1.26.5\n", ".fgm.toml": "[toolchain]\ngo = \"1.26.1\"\n"},
			config:  Config{PatchCatalog: catalog},
			wantErr: true,
		},
		{
			name:       "should select the newest catalog patch when the policy is latest-patch without a toolchain directive",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.26.9",
			wantSource: "latest-patch",
			wantFile:   "go.mod",
			wantBase:   "1.26.0",
		},
		{
			name:       "should select the newest catalog patch when latest-patch is newer than the toolchain directive",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0\n\ntoolchain go1.26.3\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.26.9",
			wantSource: "latest-patch",
			wantFile:   "go.mod",
			wantBase:   "1.26.0",
		},
		{
			name:       "should select the toolchain directive when it is newer than the latest catalog patch",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0\n\ntoolchain go1.27.1\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.27.1",
			wantSource: "go.mod",
			wantFile:   "go.mod",
		},
		{
			name:       "should select the newest 1.19 patch when a pre-1.21 module uses latest-patch",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.19\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.19.13",
			wantSource: "latest-patch",
			wantFile:   "go.mod",
			wantBase:   "1.19",
		},
		{
			name: "should apply latest-patch to the go.work directive when a workspace wraps a nested module",
			files: map[string]string{
				"go.work":    "go 1.26.0\n\nuse ./svc\n",
				"svc/go.mod": "module m\n\ngo 1.26.4\n",
				".fgm.toml":  "[toolchain]\ngo = \"latest-patch\"\n",
			},
			dir:        "svc",
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.26.9",
			wantSource: "latest-patch",
			wantFile:   "go.work",
			wantBase:   "1.26.0",
		},
		{
			name:       "should use the go directive with a warning when no patch data exists",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.28.0\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.28.0",
			wantSource: "go.mod",
			wantFile:   "go.mod",
			wantWarn:   true,
		},
		{
			name:       "should ignore a trailing comment when the go directive has one",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0 // minimum\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.26.0",
			wantSource: "go.mod",
			wantFile:   "go.mod",
		},
		{
			name:       "should ignore a trailing comment when the toolchain directive has one",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0\n\ntoolchain go1.26.3 // pinned for CI\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.26.3",
			wantSource: "go.mod",
			wantFile:   "go.mod",
		},
		{
			name:       "should select a release candidate toolchain when it is newer than the latest catalog patch",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.26.0\n\ntoolchain go1.27rc1\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"},
			config:     Config{PatchCatalog: catalog},
			wantGo:     "1.27rc1",
			wantSource: "go.mod",
			wantFile:   "go.mod",
		},
		{
			name:    "should return an error when the explicit version is lower than a release candidate go directive",
			files:   map[string]string{"go.mod": "module m\n\ngo 1.27rc1\n", ".fgm.toml": "[toolchain]\ngo = \"1.26.3\"\n"},
			config:  Config{PatchCatalog: catalog},
			wantErr: true,
		},
		{
			name:       "should fall back to the newest installed patch of the minor when the go directive is a release candidate",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.27rc1\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"},
			config:     Config{PatchCatalog: catalog, LocalVersions: stubVersionList{versions: []string{"1.27.1", "1.26.9"}}},
			wantGo:     "1.27.1",
			wantSource: "latest-patch",
			wantFile:   "go.mod",
			wantBase:   "1.27rc1",
			wantWarn:   true,
		},
		{
			name:       "should fall back to the newest installed patch with a warning when the catalog lacks the minor",
			files:      map[string]string{"go.mod": "module m\n\ngo 1.28.0\n", ".fgm.toml": "[toolchain]\ngo = \"latest-patch\"\n"},
			config:     Config{PatchCatalog: catalog, LocalVersions: stubVersionList{versions: []string{"1.28.1", "1.27.9"}}},
			wantGo:     "1.28.1",
			wantSource: "latest-patch",
			wantFile:   "go.mod",
			wantBase:   "1.28.0",
			wantWarn:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := t.TempDir()
			writeFiles(t, root, tc.files)
			resolver := NewWithConfig(tc.config)

			// Act
			selection, err := resolver.Current(context.Background(), filepath.Join(root, tc.dir))

			// Assert
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Current returned %+v; want an error", selection)
				}
				return
			}
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if selection.GoVersion != tc.wantGo {
				t.Errorf("GoVersion = %q; want %q", selection.GoVersion, tc.wantGo)
			}
			gotSource := selection.GoSource
			if filepath.IsAbs(gotSource) {
				gotSource = filepath.Base(gotSource)
			}
			if gotSource != tc.wantSource {
				t.Errorf("GoSource = %q; want %q", selection.GoSource, tc.wantSource)
			}
			if filepath.Base(selection.GoSourceFile) != tc.wantFile {
				t.Errorf("GoSourceFile = %q; want base %q", selection.GoSourceFile, tc.wantFile)
			}
			if selection.GoBase != tc.wantBase {
				t.Errorf("GoBase = %q; want %q", selection.GoBase, tc.wantBase)
			}
			if got := len(selection.Warnings) > 0; got != tc.wantWarn {
				t.Errorf("Warnings = %v; want present=%v", selection.Warnings, tc.wantWarn)
			}
		})
	}
}

func TestLatestPatchCandidates(t *testing.T) {
	t.Parallel()

	t.Run("newestPatch", func(t *testing.T) {
		t.Parallel()

		t.Run("should ignore other minors and prereleases when picking the newest patch", func(t *testing.T) {
			t.Parallel()

			// Arrange
			versions := []string{"1.26.3", "1.27.0", "1.26.10rc1", "1.26.9"}

			// Act
			got, ok := newestPatch(versions, "1.26.0")

			// Assert
			if !ok || got != "1.26.9" {
				t.Errorf("newestPatch = %q, %v; want 1.26.9", got, ok)
			}
			if !slices.Contains(versions, "1.27.0") {
				t.Fatal("input mutated")
			}
		})
	})
}
