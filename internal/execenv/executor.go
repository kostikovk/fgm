package execenv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/goenv"
	"github.com/kostikovk/fgm/internal/toolhint"
)

const lintCommand = "golangci-lint"

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
	environ     func() []string
	shimDir     string
}

// Config configures an Executor.
type Config struct {
	Resolver    SelectionResolver
	GoLocator   GoBinaryLocator
	LintLocator LintBinaryLocator
	PathEnv     string
	// Environ returns the inherited environment; os.Environ is used when nil.
	Environ func() []string
	// ShimDir is the FGM shim directory; a golangci-lint shim found there reports a missing selection itself.
	ShimDir string
}

// New constructs an Executor.
func New(config Config) *Executor {
	return &Executor{
		resolver:    config.Resolver,
		goLocator:   config.GoLocator,
		lintLocator: config.LintLocator,
		pathEnv:     config.PathEnv,
		environ:     config.Environ,
		shimDir:     config.ShimDir,
	}
}

// Exec runs a command with the selected Go binary directory prepended to PATH.
func (e *Executor) Exec(ctx context.Context, workDir string, args []string, options app.ExecOptions, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}
	toolchainMode, err := goenv.ParseMode(options.GoToolchain)
	if err != nil {
		return err
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
	selectedLintMissing := false
	if selection.LintVersion != "" && e.lintLocator != nil {
		lintBinaryPath, err := e.lintLocator.LintBinaryPath(ctx, selection.LintVersion)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err == nil {
			pathDirs = append(pathDirs, filepath.Dir(lintBinaryPath))
		} else {
			selectedLintMissing = true
		}
	}
	if selectedLintMissing && !options.AllowUnselectedLint && commandName(args[0]) == lintCommand {
		return fmt.Errorf(
			"selected golangci-lint %s is not installed; refusing to run another version from PATH (run %q, pass --install, or --allow-unselected-lint)",
			selection.LintVersion,
			toolhint.InstallLintCommand(selection.LintVersion),
		)
	}

	inherited := os.Environ()
	if e.environ != nil {
		inherited = e.environ()
	}
	env, warning := goenv.Apply(withPrependedPath(inherited, pathDirs, e.pathEnv), toolchainMode)
	if warning != "" && stderr != nil {
		if _, err := fmt.Fprintln(stderr, warning); err != nil {
			return err
		}
	}
	if selectedLintMissing && !options.AllowUnselectedLint && stderr != nil {
		if found, err := resolveCommandPath(lintCommand, env); err == nil && filepath.Dir(found) != e.shimDir {
			if _, err := fmt.Fprintf(stderr, "fgm: warning: selected golangci-lint %s is not installed; golangci-lint on PATH resolves to %s\n", selection.LintVersion, found); err != nil {
				return err
			}
		}
	}
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

func commandName(command string) string {
	return strings.TrimSuffix(filepath.Base(command), ".exe")
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
