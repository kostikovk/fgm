package pinnedlint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePinned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string // returns workDir
		wantVersion string
		wantFound   bool
		wantErr     bool
	}{
		{
			name: "no .fgm.toml found",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			wantVersion: "",
			wantFound:   false,
			wantErr:     false,
		},
		{
			name: "golangci_lint pinned to explicit version",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				content := "[toolchain]\ngolangci_lint = \"v2.10.1\"\n"
				if err := os.WriteFile(filepath.Join(dir, ".fgm.toml"), []byte(content), 0o644); err != nil {
					t.Fatalf("write .fgm.toml: %v", err)
				}
				return dir
			},
			wantVersion: "v2.10.1",
			wantFound:   true,
			wantErr:     false,
		},
		{
			name: "golangci_lint set to auto",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				content := "[toolchain]\ngolangci_lint = \"auto\"\n"
				if err := os.WriteFile(filepath.Join(dir, ".fgm.toml"), []byte(content), 0o644); err != nil {
					t.Fatalf("write .fgm.toml: %v", err)
				}
				return dir
			},
			wantVersion: "",
			wantFound:   false,
			wantErr:     false,
		},
		{
			name: "golangci_lint set to empty string",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				content := "[toolchain]\ngolangci_lint = \"\"\n"
				if err := os.WriteFile(filepath.Join(dir, ".fgm.toml"), []byte(content), 0o644); err != nil {
					t.Fatalf("write .fgm.toml: %v", err)
				}
				return dir
			},
			wantVersion: "",
			wantFound:   false,
			wantErr:     false,
		},
		{
			name: ".fgm.toml in parent dir",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				workDir := filepath.Join(root, "sub", "project")
				if err := os.MkdirAll(workDir, 0o755); err != nil {
					t.Fatalf("mkdir workdir: %v", err)
				}
				content := "[toolchain]\ngolangci_lint = \"v2.10.1\"\n"
				if err := os.WriteFile(filepath.Join(root, ".fgm.toml"), []byte(content), 0o644); err != nil {
					t.Fatalf("write .fgm.toml: %v", err)
				}
				return workDir
			},
			wantVersion: "v2.10.1",
			wantFound:   true,
			wantErr:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			workDir := tc.setup(t)

			gotVersion, gotFound, err := ResolvePinned(workDir)

			if (err != nil) != tc.wantErr {
				t.Fatalf("ResolvePinned() error = %v, wantErr %v", err, tc.wantErr)
			}
			if gotFound != tc.wantFound {
				t.Errorf("ResolvePinned() found = %v, want %v", gotFound, tc.wantFound)
			}
			if gotVersion != tc.wantVersion {
				t.Errorf("ResolvePinned() version = %q, want %q", gotVersion, tc.wantVersion)
			}
		})
	}
}
