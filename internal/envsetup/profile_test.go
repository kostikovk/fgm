package envsetup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		shell   string
		goos    string
		want    string
		wantErr bool
	}{
		{"zsh", "darwin", ".zshrc", false},
		{"zsh", "linux", ".zshrc", false},
		{"bash", "darwin", ".bash_profile", false},
		{"bash", "linux", ".bashrc", false},
		{"fish", "linux", filepath.Join(".config", "fish", "config.fish"), false},
		{"powershell", "windows", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.shell+"_"+tt.goos, func(t *testing.T) {
			t.Parallel()
			got, err := ProfilePath(tt.shell, tt.goos, "/home/user")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := filepath.Join("/home/user", tt.want)
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestEvalLine(t *testing.T) {
	t.Parallel()

	if got := EvalLine("zsh"); got != `eval "$(fgm env)"` {
		t.Fatalf("zsh: got %q", got)
	}
	if got := EvalLine("bash"); got != `eval "$(fgm env)"` {
		t.Fatalf("bash: got %q", got)
	}
	if got := EvalLine("fish"); got != `fgm env | source` {
		t.Fatalf("fish: got %q", got)
	}
}

func TestInstallProfile_AppendsWhenMissing(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	existing := "# existing config\nexport FOO=bar\n"
	if err := os.WriteFile(filepath.Join(homeDir, ".zshrc"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	path, modified, err := InstallProfile("zsh", "darwin", homeDir)
	if err != nil {
		t.Fatalf("InstallProfile: %v", err)
	}
	if !modified {
		t.Fatal("expected modified=true")
	}
	if path != filepath.Join(homeDir, ".zshrc") {
		t.Fatalf("path = %q", path)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), `eval "$(fgm env)"`) {
		t.Fatalf("content = %q, want fgm env line", string(content))
	}
	if !strings.Contains(string(content), existing) {
		t.Fatalf("original content missing")
	}
}

func TestInstallProfile_SkipsWhenAlreadyPresent(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	existing := "# existing config\neval \"$(fgm env)\"\n"
	if err := os.WriteFile(filepath.Join(homeDir, ".zshrc"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	_, modified, err := InstallProfile("zsh", "darwin", homeDir)
	if err != nil {
		t.Fatalf("InstallProfile: %v", err)
	}
	if modified {
		t.Fatal("expected modified=false when line already present")
	}
}

func TestInstallProfile_CreatesFileWhenAbsent(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()

	path, modified, err := InstallProfile("zsh", "darwin", homeDir)
	if err != nil {
		t.Fatalf("InstallProfile: %v", err)
	}
	if !modified {
		t.Fatal("expected modified=true")
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), `eval "$(fgm env)"`) {
		t.Fatalf("content = %q, want fgm env line", string(content))
	}
}

func TestInstallProfile_CreatesFishDirectories(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()

	path, modified, err := InstallProfile("fish", "linux", homeDir)
	if err != nil {
		t.Fatalf("InstallProfile: %v", err)
	}
	if !modified {
		t.Fatal("expected modified=true")
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "fgm env | source") {
		t.Fatalf("content = %q, want fish eval line", string(content))
	}
}
