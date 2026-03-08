package envsetup

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// GoStore provides the shim directory path.
type GoStore interface {
	ShimDir() string
}

// Config configures a Renderer.
type Config struct {
	GoStore   GoStore
	ShellPath string
	GOOS      string
}

// Renderer prints shell-specific environment setup snippets.
type Renderer struct {
	goStore   GoStore
	shellPath string
	goos      string
}

// New constructs a Renderer.
func New(config Config) *Renderer {
	goos := config.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}

	return &Renderer{
		goStore:   config.GoStore,
		shellPath: config.ShellPath,
		goos:      goos,
	}
}

// Render prints shell-specific environment setup lines.
func (r *Renderer) Render(ctx context.Context, shell string) ([]string, error) {
	_ = ctx

	detectedShell := shell
	if detectedShell == "" {
		detectedShell = detectShell(r.shellPath, r.goos)
	}
	if detectedShell == "" {
		return nil, fmt.Errorf("could not detect shell; use --shell")
	}

	shimDir := r.goStore.ShimDir()
	switch detectedShell {
	case "zsh", "bash":
		return []string{
			"# Add FGM shims before other Go binaries",
			fmt.Sprintf("export PATH=%q:$PATH", shimDir),
		}, nil
	case "fish":
		return []string{
			"# Add FGM shims before other Go binaries",
			fmt.Sprintf("set -gx PATH %q $PATH", shimDir),
		}, nil
	case "powershell":
		return []string{
			"# Add FGM shims before other Go binaries",
			fmt.Sprintf("$env:PATH = %q + ';' + $env:PATH", shimDir),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported shell %q", detectedShell)
	}
}

func detectShell(shellPath string, goos string) string {
	if goos == "windows" {
		lower := strings.ToLower(shellPath)
		if strings.Contains(lower, "powershell") || strings.Contains(lower, "pwsh") {
			return "powershell"
		}
		return ""
	}

	base := filepath.Base(shellPath)
	switch base {
	case "zsh", "bash", "fish":
		return base
	default:
		return ""
	}
}
