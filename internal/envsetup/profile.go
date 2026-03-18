package envsetup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProfilePath returns the shell profile file path for the given shell.
func ProfilePath(shell string, goos string, homeDir string) (string, error) {
	switch shell {
	case "zsh":
		return filepath.Join(homeDir, ".zshrc"), nil
	case "bash":
		if goos == "darwin" {
			return filepath.Join(homeDir, ".bash_profile"), nil
		}
		return filepath.Join(homeDir, ".bashrc"), nil
	case "fish":
		return filepath.Join(homeDir, ".config", "fish", "config.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell %q for profile installation", shell)
	}
}

// EvalLine returns the eval snippet that should be appended to a shell profile.
func EvalLine(shell string) string {
	if shell == "fish" {
		return `fgm env | source`
	}
	return `eval "$(fgm env)"`
}

// InstallProfile appends the eval line to the shell profile if not already present.
// Returns the profile path and whether the file was modified.
func InstallProfile(shell string, goos string, homeDir string) (string, bool, error) {
	profilePath, err := ProfilePath(shell, goos, homeDir)
	if err != nil {
		return "", false, err
	}

	// Ensure parent directory exists (needed for fish).
	dir := filepath.Dir(profilePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return profilePath, false, fmt.Errorf("create directory %s: %w", dir, err)
	}

	evalLine := EvalLine(shell)

	// Read existing content.
	content, err := os.ReadFile(profilePath)
	if err != nil && !os.IsNotExist(err) {
		return profilePath, false, fmt.Errorf("read %s: %w", profilePath, err)
	}

	// Check if already present.
	if strings.Contains(string(content), "fgm env") {
		return profilePath, false, nil
	}

	// Append the eval line.
	f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return profilePath, false, fmt.Errorf("open %s: %w", profilePath, err)
	}
	defer f.Close()

	line := "\n# FGM shell integration\n" + evalLine + "\n"
	if _, err := f.WriteString(line); err != nil {
		return profilePath, false, fmt.Errorf("write %s: %w", profilePath, err)
	}

	return profilePath, true, nil
}
