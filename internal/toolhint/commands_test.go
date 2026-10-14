package toolhint

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCommandHints(t *testing.T) {
	t.Parallel()

	t.Run("should format install go command when Go version is provided", func(t *testing.T) {
		t.Parallel()

		// Arrange
		version := "1.27.2"

		// Act
		got := InstallGoCommand(version)

		// Assert
		want := "fgm install go 1.27.2"
		if got != want {
			t.Fatalf("InstallGoCommand(%q) = %q, want %q", version, got, want)
		}
	})

	t.Run("should format install golangci-lint command when lint version is provided", func(t *testing.T) {
		t.Parallel()

		// Arrange
		version := "v2.13.0"

		// Act
		got := InstallLintCommand(version)

		// Assert
		want := "fgm install golangci-lint v2.13.0"
		if got != want {
			t.Fatalf("InstallLintCommand(%q) = %q, want %q", version, got, want)
		}
	})

	t.Run("should return non-interactive exec command template when ExecCommand is called", func(t *testing.T) {
		t.Parallel()

		// Arrange / Act
		got := ExecCommand()

		// Assert
		want := "fgm exec -- <command>"
		if got != want {
			t.Fatalf("ExecCommand() = %q, want %q", got, want)
		}
	})

	t.Run("should return Homebrew cask upgrade command when executable path is under Homebrew prefix", func(t *testing.T) {
		t.Parallel()

		// Arrange
		paths := []string{
			"/opt/homebrew/bin/fgm",
			"/opt/homebrew/Caskroom/fgm/0.5.0/fgm",
			"/usr/local/Caskroom/fgm/0.5.0/fgm",
			"/home/linuxbrew/.linuxbrew/bin/fgm",
		}

		for _, path := range paths {
			// Act
			got := UpgradeFGMCommand(path)

			// Assert
			want := "brew update && brew upgrade --cask fgm"
			if got != want {
				t.Fatalf("UpgradeFGMCommand(%q) = %q, want %q", path, got, want)
			}
		}
	})

	t.Run("should return install script upgrade command when executable path is outside Homebrew", func(t *testing.T) {
		t.Parallel()

		// Arrange
		paths := []string{
			"/Users/dev/.local/bin/fgm",
			"/usr/bin/fgm",
			"",
		}

		for _, path := range paths {
			// Act
			got := UpgradeFGMCommand(path)

			// Assert
			want := "curl -fsSL https://raw.githubusercontent.com/kostikovk/fgm/main/install.sh | sh"
			if got != want {
				t.Fatalf("UpgradeFGMCommand(%q) = %q, want %q", path, got, want)
			}
		}
	})

	t.Run("should resolve symlink pointing into Caskroom and return Homebrew cask upgrade command", func(t *testing.T) {
		t.Parallel()

		if runtime.GOOS == "windows" {
			t.Skip("symlink test is unix-only")
		}

		// Arrange
		root := t.TempDir()
		caskBinDir := filepath.Join(root, "Caskroom", "fgm", "0.5.0")
		if err := os.MkdirAll(caskBinDir, 0o755); err != nil {
			t.Fatalf("mkdir caskBinDir: %v", err)
		}
		realBin := filepath.Join(caskBinDir, "fgm")
		if err := os.WriteFile(realBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatalf("write realBin: %v", err)
		}

		binDir := filepath.Join(root, "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatalf("mkdir binDir: %v", err)
		}
		symlinkBin := filepath.Join(binDir, "fgm")
		if err := os.Symlink(realBin, symlinkBin); err != nil {
			t.Fatalf("symlink: %v", err)
		}

		// Act
		got := UpgradeFGMCommand(symlinkBin)

		// Assert
		want := "brew update && brew upgrade --cask fgm"
		if got != want {
			t.Fatalf("UpgradeFGMCommand(%q) = %q, want %q", symlinkBin, got, want)
		}
	})
}
