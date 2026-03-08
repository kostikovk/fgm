package execenv

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/koskosovu4/fgm/internal/app"
)

// SelectionResolver resolves the selected Go version for a working directory.
type SelectionResolver interface {
	Current(ctx context.Context, workDir string) (app.Selection, error)
}

// GoBinaryLocator returns the executable path for a Go version.
type GoBinaryLocator interface {
	GoBinaryPath(ctx context.Context, version string) (string, error)
}

// Executor runs commands with the selected Go version prepended to PATH.
type Executor struct {
	resolver SelectionResolver
	locator  GoBinaryLocator
	pathEnv  string
}

// New constructs an Executor.
func New(resolver SelectionResolver, locator GoBinaryLocator, pathEnv string) *Executor {
	return &Executor{
		resolver: resolver,
		locator:  locator,
		pathEnv:  pathEnv,
	}
}

// Exec runs a command with the selected Go binary directory prepended to PATH.
func (e *Executor) Exec(ctx context.Context, workDir string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	selection, err := e.resolver.Current(ctx, workDir)
	if err != nil {
		return err
	}

	goBinaryPath, err := e.locator.GoBinaryPath(ctx, selection.GoVersion)
	if err != nil {
		return err
	}

	env := withPrependedPath(os.Environ(), filepath.Dir(goBinaryPath), e.pathEnv)
	commandPath, err := resolveCommandPath(args[0], env)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, commandPath, args[1:]...)
	cmd.Dir = workDir
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = env

	return cmd.Run()
}

func withPrependedPath(env []string, goBinDir string, fallbackPath string) []string {
	pathValue := fallbackPath
	for _, entry := range env {
		if len(entry) > 5 && entry[:5] == "PATH=" {
			pathValue = entry[5:]
			break
		}
	}

	newPath := goBinDir
	if pathValue != "" {
		newPath = goBinDir + string(os.PathListSeparator) + pathValue
	}

	filtered := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if len(entry) > 5 && entry[:5] == "PATH=" {
			continue
		}
		filtered = append(filtered, entry)
	}

	return append(filtered, "PATH="+newPath)
}

func resolveCommandPath(command string, env []string) (string, error) {
	if strings.ContainsRune(command, filepath.Separator) {
		return command, nil
	}

	pathValue := ""
	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, "PATH="); ok {
			pathValue = after
			break
		}
	}

	for _, dir := range filepath.SplitList(pathValue) {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, command)
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Mode()&0o111 == 0 {
			continue
		}
		return candidate, nil
	}

	return "", fmt.Errorf("resolve command %q on PATH", command)
}
