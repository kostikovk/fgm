package toolhint

import "fmt"

// InstallGoCommand returns the exact CLI command to install a specific Go version.
func InstallGoCommand(version string) string {
	return fmt.Sprintf("fgm install go %s", version)
}

// InstallLintCommand returns the exact CLI command to install a specific golangci-lint version.
func InstallLintCommand(version string) string {
	return fmt.Sprintf("fgm install golangci-lint %s", version)
}

// ExecCommand returns the CLI command template for non-interactive execution.
func ExecCommand() string {
	return "fgm exec -- <command>"
}
