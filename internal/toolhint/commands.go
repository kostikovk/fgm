package toolhint

import (
	"fmt"
	"path/filepath"
	"strings"
)

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

// UpgradeFGMCommand returns the CLI command to upgrade FGM itself based on its executable path.
func UpgradeFGMCommand(execPath string) string {
	resolved := execPath
	if execPath != "" {
		if target, err := filepath.EvalSymlinks(execPath); err == nil && target != "" {
			resolved = target
		}
	}
	for _, candidate := range []string{execPath, resolved} {
		lower := strings.ToLower(candidate)
		if strings.Contains(lower, "/homebrew/") ||
			strings.Contains(lower, "/linuxbrew/") ||
			strings.Contains(lower, "/caskroom/") ||
			strings.Contains(lower, "/cellar/") {
			return "brew update && brew upgrade --cask fgm"
		}
	}
	return "curl -fsSL https://raw.githubusercontent.com/kostikovk/fgm/main/install.sh | sh"
}
