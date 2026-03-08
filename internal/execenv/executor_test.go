package execenv

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
)

type stubResolver struct {
	currentFn func(ctx context.Context, workDir string) (app.Selection, error)
}

func (s stubResolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	return s.currentFn(ctx, workDir)
}

type stubLocator struct {
	goBinaryPathFn   func(ctx context.Context, version string) (string, error)
	lintBinaryPathFn func(ctx context.Context, version string) (string, error)
}

func (s stubLocator) GoBinaryPath(ctx context.Context, version string) (string, error) {
	return s.goBinaryPathFn(ctx, version)
}

func (s stubLocator) LintBinaryPath(ctx context.Context, version string) (string, error) {
	if s.lintBinaryPathFn == nil {
		return "", nil
	}
	return s.lintBinaryPathFn(ctx, version)
}

func TestExecutorExec_PrependsSelectedGoBinaryDirToPATH(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	workDir := t.TempDir()
	goBinDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(goBinDir, 0o755); err != nil {
		t.Fatalf("mkdir go bin dir: %v", err)
	}

	goBinary := filepath.Join(goBinDir, "go")
	script := "#!/bin/sh\nprintf 'go version go1.25.7 darwin/arm64\\n'\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go binary: %v", err)
	}

	executor := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				if gotWorkDir != workDir {
					t.Fatalf("workDir = %q, want %q", gotWorkDir, workDir)
				}
				return app.Selection{GoVersion: "1.25.7", GoSource: "global"}, nil
			},
		},
		GoLocator: stubLocator{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				if version != "1.25.7" {
					t.Fatalf("version = %q, want %q", version, "1.25.7")
				}
				return goBinary, nil
			},
		},
		PathEnv: "",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := executor.Exec(context.Background(), workDir, []string{"go", "version"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("Exec: %v\nstderr:\n%s", err, stderr.String())
	}

	if !strings.Contains(stdout.String(), "go version go1.25.7") {
		t.Fatalf("stdout = %q, want it to contain fake go output", stdout.String())
	}
}

func TestExecutorExec_PrependsSelectedLintBinaryDirToPATH(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	workDir := t.TempDir()
	lintBinDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(lintBinDir, 0o755); err != nil {
		t.Fatalf("mkdir lint bin dir: %v", err)
	}

	lintBinary := filepath.Join(lintBinDir, "golangci-lint")
	script := "#!/bin/sh\nprintf 'golangci-lint has version v2.11.2\\n'\n"
	if err := os.WriteFile(lintBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake lint binary: %v", err)
	}

	executor := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				if gotWorkDir != workDir {
					t.Fatalf("workDir = %q, want %q", gotWorkDir, workDir)
				}
				return app.Selection{
					GoVersion:   "1.25.7",
					GoSource:    "global",
					LintVersion: "v2.11.2",
					LintSource:  "local",
				}, nil
			},
		},
		GoLocator: stubLocator{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return filepath.Join(workDir, "unused-go"), nil
			},
		},
		LintLocator: stubLocator{
			lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				if version != "v2.11.2" {
					t.Fatalf("version = %q, want %q", version, "v2.11.2")
				}
				return lintBinary, nil
			},
		},
		PathEnv: "",
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := executor.Exec(
		context.Background(),
		workDir,
		[]string{"golangci-lint", "version"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if err != nil {
		t.Fatalf("Exec: %v\nstderr:\n%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "golangci-lint has version v2.11.2") {
		t.Fatalf("stdout = %q, want it to contain fake lint output", stdout.String())
	}
}

func TestExecutorExec_ReturnsErrorWhenArgsEmpty(t *testing.T) {
	t.Parallel()

	executor := New(Config{})
	err := executor.Exec(context.Background(), t.TempDir(), nil, nil, nil, nil)
	if err == nil {
		t.Fatal("expected an error when args is empty")
	}
	if !strings.Contains(err.Error(), "no command provided") {
		t.Fatalf("err = %q, want it to contain %q", err, "no command provided")
	}
}

func TestExecutorExec_ReturnsErrorWhenResolverFails(t *testing.T) {
	t.Parallel()

	executor := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{}, fmt.Errorf("resolver boom")
			},
		},
	})

	var stdout, stderr bytes.Buffer
	err := executor.Exec(context.Background(), t.TempDir(), []string{"go", "version"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected an error when resolver fails")
	}
	if !strings.Contains(err.Error(), "resolver boom") {
		t.Fatalf("err = %q, want it to contain %q", err, "resolver boom")
	}
}

func TestExecutorExec_ReturnsErrorWhenGoLocatorFails(t *testing.T) {
	t.Parallel()

	executor := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7", GoSource: "global"}, nil
			},
		},
		GoLocator: stubLocator{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("go locator boom")
			},
		},
	})

	var stdout, stderr bytes.Buffer
	err := executor.Exec(context.Background(), t.TempDir(), []string{"go", "version"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected an error when go locator fails")
	}
	if !strings.Contains(err.Error(), "go locator boom") {
		t.Fatalf("err = %q, want it to contain %q", err, "go locator boom")
	}
}

func TestExecutorExec_ReturnsErrorWhenLintLocatorFails(t *testing.T) {
	t.Parallel()

	executor := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion:   "1.25.7",
					GoSource:    "global",
					LintVersion: "v2.11.2",
					LintSource:  "local",
				}, nil
			},
		},
		GoLocator: stubLocator{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return filepath.Join(t.TempDir(), "go"), nil
			},
		},
		LintLocator: stubLocator{
			lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("lint locator boom")
			},
		},
	})

	var stdout, stderr bytes.Buffer
	err := executor.Exec(context.Background(), t.TempDir(), []string{"go", "version"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected an error when lint locator fails")
	}
	if !strings.Contains(err.Error(), "lint locator boom") {
		t.Fatalf("err = %q, want it to contain %q", err, "lint locator boom")
	}
}

func TestResolveCommandPath_ReturnsAbsolutePathAsIs(t *testing.T) {
	t.Parallel()

	absPath := filepath.Join(string(filepath.Separator), "usr", "local", "bin", "go")
	got, err := resolveCommandPath(absPath, nil)
	if err != nil {
		t.Fatalf("resolveCommandPath: %v", err)
	}
	if got != absPath {
		t.Fatalf("resolveCommandPath = %q, want %q", got, absPath)
	}
}

func TestResolveCommandPath_ReturnsErrorWhenCommandNotFound(t *testing.T) {
	t.Parallel()

	env := []string{"PATH=" + t.TempDir()}
	_, err := resolveCommandPath("nonexistent-binary-abc123", env)
	if err == nil {
		t.Fatal("expected an error when command not found on PATH")
	}
	if !strings.Contains(err.Error(), "resolve command") {
		t.Fatalf("err = %q, want it to contain %q", err, "resolve command")
	}
}

func TestResolveCommandPath_SkipsNonExecutableFiles(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("executable permission test is unix-only")
	}

	dir := t.TempDir()

	// Create a file that exists but is not executable.
	nonExec := filepath.Join(dir, "mytool")
	if err := os.WriteFile(nonExec, []byte("not executable"), 0o644); err != nil {
		t.Fatalf("write non-executable file: %v", err)
	}

	env := []string{"PATH=" + dir}
	_, err := resolveCommandPath("mytool", env)
	if err == nil {
		t.Fatal("expected an error when file is not executable")
	}
	if !strings.Contains(err.Error(), "resolve command") {
		t.Fatalf("err = %q, want it to contain %q", err, "resolve command")
	}
}

func TestResolveCommandPath_FindsExecutableOnPATH(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("executable permission test is unix-only")
	}

	dir := t.TempDir()
	execFile := filepath.Join(dir, "mytool")
	if err := os.WriteFile(execFile, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write executable file: %v", err)
	}

	env := []string{"PATH=" + dir}
	got, err := resolveCommandPath("mytool", env)
	if err != nil {
		t.Fatalf("resolveCommandPath: %v", err)
	}
	if got != execFile {
		t.Fatalf("resolveCommandPath = %q, want %q", got, execFile)
	}
}

func TestWithPrependedPath_HandlesEmptyToolDirs(t *testing.T) {
	t.Parallel()

	env := []string{"PATH=/usr/local/bin:/usr/bin"}
	result := withPrependedPath(env, []string{"", ""}, "/fallback")

	found := false
	for _, entry := range result {
		if len(entry) > 5 && entry[:5] == "PATH=" {
			found = true
			// With all empty tool dirs, PATH should remain unchanged.
			if !strings.Contains(entry, "/usr/local/bin") {
				t.Fatalf("PATH entry = %q, want it to contain original path", entry)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected PATH entry in result")
	}
}

func TestWithPrependedPath_UsesFallbackPathWhenNoPATHInEnv(t *testing.T) {
	t.Parallel()

	env := []string{"HOME=/home/user"}
	toolDirs := []string{"/my/go/bin"}
	result := withPrependedPath(env, toolDirs, "/fallback/path")

	found := false
	for _, entry := range result {
		if len(entry) > 5 && entry[:5] == "PATH=" {
			found = true
			pathValue := entry[5:]
			if !strings.Contains(pathValue, "/my/go/bin") {
				t.Fatalf("PATH = %q, want it to contain /my/go/bin", pathValue)
			}
			if !strings.Contains(pathValue, "/fallback/path") {
				t.Fatalf("PATH = %q, want it to contain /fallback/path", pathValue)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected PATH entry in result")
	}
}

func TestWithPrependedPath_EmptyToolDirsAndEmptyPath(t *testing.T) {
	t.Parallel()

	env := []string{"HOME=/home/user"}
	result := withPrependedPath(env, nil, "")

	for _, entry := range result {
		if len(entry) >= 5 && entry[:5] == "PATH=" {
			pathValue := entry[5:]
			if pathValue != "" {
				t.Fatalf("PATH = %q, want empty", pathValue)
			}
			return
		}
	}
	t.Fatal("expected PATH entry in result")
}

func TestExecutorExec_ReturnsErrorWhenCommandNotFoundOnPATH(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only test")
	}

	workDir := t.TempDir()
	goBinDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(goBinDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	goBinary := filepath.Join(goBinDir, "go")
	if err := os.WriteFile(goBinary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	executor := New(Config{
		Resolver: stubResolver{
			currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7", GoSource: "global"}, nil
			},
		},
		GoLocator: stubLocator{
			goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
				return goBinary, nil
			},
		},
		PathEnv: "",
	})

	var stdout, stderr bytes.Buffer
	err := executor.Exec(
		context.Background(),
		workDir,
		[]string{"nonexistent-tool-xyz"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if err == nil {
		t.Fatal("expected an error when command is not found on PATH")
	}
	if !strings.Contains(err.Error(), "resolve command") {
		t.Fatalf("err = %q, want resolve command error", err)
	}
}

func TestResolveCommandPath_SkipsEmptyDirsInPATH(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("unix-only")
	}

	dir := t.TempDir()
	execFile := filepath.Join(dir, "mytool")
	if err := os.WriteFile(execFile, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write executable file: %v", err)
	}

	// PATH with empty segments
	env := []string{"PATH=" + string(os.PathListSeparator) + string(os.PathListSeparator) + dir}
	got, err := resolveCommandPath("mytool", env)
	if err != nil {
		t.Fatalf("resolveCommandPath: %v", err)
	}
	if got != execFile {
		t.Fatalf("resolveCommandPath = %q, want %q", got, execFile)
	}
}

func TestResolveCommandPath_SkipsDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "mytool")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	env := []string{"PATH=" + dir}
	_, err := resolveCommandPath("mytool", env)
	if err == nil {
		t.Fatal("expected an error when candidate is a directory")
	}
	if !strings.Contains(err.Error(), "resolve command") {
		t.Fatalf("err = %q, want it to contain %q", err, "resolve command")
	}
}
