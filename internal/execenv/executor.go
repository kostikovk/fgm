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

// LintBinaryLocator returns the executable path for a golangci-lint version.
type LintBinaryLocator interface {
	LintBinaryPath(ctx context.Context, version string) (string, error)
}

// Executor runs commands with the selected Go version prepended to PATH.
type Executor struct {
	resolver    SelectionResolver
	goLocator   GoBinaryLocator
	lintLocator LintBinaryLocator
	pathEnv     string
}

// Config configures an Executor.
type Config struct {
	Resolver    SelectionResolver
	GoLocator   GoBinaryLocator
	LintLocator LintBinaryLocator
	PathEnv     string
}

// New constructs an Executor.
func New(config Config) *Executor {
	return &Executor{
		resolver:    config.Resolver,
		goLocator:   config.GoLocator,
		lintLocator: config.LintLocator,
		pathEnv:     config.PathEnv,
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

	goBinaryPath, err := e.goLocator.GoBinaryPath(ctx, selection.GoVersion)
	if err != nil {
		return err
	}

	pathDirs := []string{filepath.Dir(goBinaryPath)}
	if selection.LintVersion != "" && e.lintLocator != nil {
		lintBinaryPath, err := e.lintLocator.LintBinaryPath(ctx, selection.LintVersion)
		if err != nil {
			return err
		}
		pathDirs = append(pathDirs, filepath.Dir(lintBinaryPath))
	}

	env := withPrependedPath(os.Environ(), pathDirs, e.pathEnv)
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

func withPrependedPath(env []string, toolDirs []string, fallbackPath string) []string {
	pathValue := fallbackPath
	for _, entry := range env {
		if len(entry) > 5 && entry[:5] == "PATH=" {
			pathValue = entry[5:]
			break
		}
	}

	prefix := make([]string, 0, len(toolDirs))
	for _, dir := range toolDirs {
		if dir == "" {
			continue
		}
		prefix = append(prefix, dir)
	}

	newPath := strings.Join(prefix, string(os.PathListSeparator))
	if pathValue != "" {
		if newPath != "" {
			newPath = newPath + string(os.PathListSeparator) + pathValue
		} else {
			newPath = pathValue
		}
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
