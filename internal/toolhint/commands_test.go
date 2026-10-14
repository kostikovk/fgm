package toolhint

import "testing"

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
}
