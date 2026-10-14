package execenv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/lintlocal"
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

func TestExecutorExec_LintContractScenarios(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script execution test is unix-only")
	}

	t.Run("should run Go tests when recommended lint is not installed", func(t *testing.T) {
		// Arrange
		workDir := t.TempDir()
		goBinDir := filepath.Join(t.TempDir(), "go-bin")
		if err := os.MkdirAll(goBinDir, 0o755); err != nil {
			t.Fatalf("mkdir go bin dir: %v", err)
		}
		goBinary := filepath.Join(goBinDir, "go")
		script := "#!/bin/sh\nprintf 'args=%s cwd=%s\\n' \"$*\" \"$(pwd)\"\n"
		if err := os.WriteFile(goBinary, []byte(script), 0o755); err != nil {
			t.Fatalf("write fake go binary: %v", err)
		}
		lintStore := lintlocal.New(t.TempDir())
		executor := New(Config{
			Resolver: stubResolver{
				currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.25.7",
						GoSource:    "go.mod",
						LintVersion: "v2.11.2",
						LintSource:  "remote",
					}, nil
				},
			},
			GoLocator: stubLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return goBinary, nil
				},
			},
			LintLocator: lintStore,
			PathEnv:     "/usr/bin:/bin",
		})
		var stdout, stderr bytes.Buffer

		// Act
		err := executor.Exec(
			context.Background(),
			workDir,
			[]string{"go", "test", "./..."},
			strings.NewReader(""),
			&stdout,
			&stderr,
		)

		// Assert
		if err != nil {
			t.Fatalf("Exec: %v\nstderr:\n%s", err, stderr.String())
		}
		if !strings.Contains(stdout.String(), "args=test ./...") {
			t.Fatalf("stdout = %q, want args=test ./...", stdout.String())
		}
		wantWorkDir, err := filepath.EvalSymlinks(workDir)
		if err != nil {
			wantWorkDir = workDir
		}
		if !strings.Contains(stdout.String(), wantWorkDir) && !strings.Contains(stdout.String(), workDir) {
			t.Fatalf("stdout = %q, want cwd %q", stdout.String(), workDir)
		}
	})

	t.Run("should run Go commands when pinned lint is not installed", func(t *testing.T) {
		// Arrange
		workDir := t.TempDir()
		goBinDir := filepath.Join(t.TempDir(), "go-bin")
		if err := os.MkdirAll(goBinDir, 0o755); err != nil {
			t.Fatalf("mkdir go bin dir: %v", err)
		}
		goBinary := filepath.Join(goBinDir, "go")
		if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go env ok\\n'\n"), 0o755); err != nil {
			t.Fatalf("write fake go binary: %v", err)
		}
		lintStore := lintlocal.New(t.TempDir())
		executor := New(Config{
			Resolver: stubResolver{
				currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.25.7",
						GoSource:    "go.mod",
						LintVersion: "v2.10.1",
						LintSource:  "config",
					}, nil
				},
			},
			GoLocator: stubLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return goBinary, nil
				},
			},
			LintLocator: lintStore,
			PathEnv:     "",
		})
		var stdout, stderr bytes.Buffer

		// Act
		err := executor.Exec(
			context.Background(),
			workDir,
			[]string{"go", "env"},
			strings.NewReader(""),
			&stdout,
			&stderr,
		)

		// Assert
		if err != nil {
			t.Fatalf("Exec: %v\nstderr:\n%s", err, stderr.String())
		}
		if !strings.Contains(stdout.String(), "go env ok") {
			t.Fatalf("stdout = %q, want 'go env ok'", stdout.String())
		}
	})

	t.Run("should prepend selected lint when it is installed", func(t *testing.T) {
		// Arrange
		workDir := t.TempDir()
		goBinDir := filepath.Join(t.TempDir(), "go-bin")
		managedLintDir := filepath.Join(t.TempDir(), "managed-lint")
		inheritedLintDir := filepath.Join(t.TempDir(), "inherited-lint")
		for _, dir := range []string{goBinDir, managedLintDir, inheritedLintDir} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", dir, err)
			}
		}

		goBinary := filepath.Join(goBinDir, "go")
		if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go\\n'\n"), 0o755); err != nil {
			t.Fatalf("write fake go: %v", err)
		}
		managedLint := filepath.Join(managedLintDir, "golangci-lint")
		if err := os.WriteFile(managedLint, []byte("#!/bin/sh\nprintf 'managed v2.11.2 path=%s\\n' \"$PATH\"\n"), 0o755); err != nil {
			t.Fatalf("write managed lint: %v", err)
		}
		inheritedLint := filepath.Join(inheritedLintDir, "golangci-lint")
		if err := os.WriteFile(inheritedLint, []byte("#!/bin/sh\nprintf 'inherited v1.64.0\\n'\n"), 0o755); err != nil {
			t.Fatalf("write inherited lint: %v", err)
		}
		t.Setenv("PATH", inheritedLintDir)

		executor := New(Config{
			Resolver: stubResolver{
				currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.25.7",
						GoSource:    "go.mod",
						LintVersion: "v2.11.2",
						LintSource:  "local",
					}, nil
				},
			},
			GoLocator: stubLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return goBinary, nil
				},
			},
			LintLocator: stubLocator{
				lintBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return managedLint, nil
				},
			},
			PathEnv: inheritedLintDir,
		})
		var stdout, stderr bytes.Buffer

		// Act
		err := executor.Exec(
			context.Background(),
			workDir,
			[]string{"golangci-lint", "version"},
			strings.NewReader(""),
			&stdout,
			&stderr,
		)

		// Assert
		if err != nil {
			t.Fatalf("Exec: %v\nstderr:\n%s", err, stderr.String())
		}
		out := strings.TrimSpace(stdout.String())
		if !strings.HasPrefix(out, "managed v2.11.2 ") {
			t.Fatalf("stdout = %q, want prefix %q", out, "managed v2.11.2 ")
		}
		wantPath := strings.Join([]string{goBinDir, managedLintDir, inheritedLintDir}, string(os.PathListSeparator))
		if !strings.Contains(out, "path="+wantPath) {
			t.Fatalf("stdout = %q, want PATH order %q", out, wantPath)
		}
	})

	t.Run("should use inherited lint when selected lint is absent", func(t *testing.T) {
		// Arrange
		workDir := t.TempDir()
		goBinDir := filepath.Join(t.TempDir(), "go-bin")
		inheritedLintDir := filepath.Join(t.TempDir(), "inherited-lint")
		for _, dir := range []string{goBinDir, inheritedLintDir} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", dir, err)
			}
		}

		goBinary := filepath.Join(goBinDir, "go")
		if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go\\n'\n"), 0o755); err != nil {
			t.Fatalf("write fake go: %v", err)
		}
		inheritedLint := filepath.Join(inheritedLintDir, "golangci-lint")
		if err := os.WriteFile(inheritedLint, []byte("#!/bin/sh\nprintf 'inherited v1.64.0\\n'\n"), 0o755); err != nil {
			t.Fatalf("write inherited lint: %v", err)
		}
		t.Setenv("PATH", inheritedLintDir)

		lintStore := lintlocal.New(t.TempDir())
		executor := New(Config{
			Resolver: stubResolver{
				currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.25.7",
						GoSource:    "go.mod",
						LintVersion: "v2.11.2",
						LintSource:  "remote",
					}, nil
				},
			},
			GoLocator: stubLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return goBinary, nil
				},
			},
			LintLocator: lintStore,
			PathEnv:     inheritedLintDir,
		})
		var stdout, stderr bytes.Buffer

		// Act
		err := executor.Exec(
			context.Background(),
			workDir,
			[]string{"golangci-lint", "version"},
			strings.NewReader(""),
			&stdout,
			&stderr,
		)

		// Assert
		if err != nil {
			t.Fatalf("Exec: %v\nstderr:\n%s", err, stderr.String())
		}
		if strings.TrimSpace(stdout.String()) != "inherited v1.64.0" {
			t.Fatalf("stdout = %q, want %q", stdout.String(), "inherited v1.64.0")
		}
	})

	t.Run("should report command lookup failure when lint is absent everywhere", func(t *testing.T) {
		// Arrange
		workDir := t.TempDir()
		goBinDir := filepath.Join(t.TempDir(), "go-bin")
		if err := os.MkdirAll(goBinDir, 0o755); err != nil {
			t.Fatalf("mkdir go bin dir: %v", err)
		}
		goBinary := filepath.Join(goBinDir, "go")
		if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go\\n'\n"), 0o755); err != nil {
			t.Fatalf("write fake go: %v", err)
		}
		t.Setenv("PATH", "")

		lintStore := lintlocal.New(t.TempDir())
		executor := New(Config{
			Resolver: stubResolver{
				currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.25.7",
						GoSource:    "go.mod",
						LintVersion: "v2.11.2",
						LintSource:  "remote",
					}, nil
				},
			},
			GoLocator: stubLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return goBinary, nil
				},
			},
			LintLocator: lintStore,
			PathEnv:     "",
		})
		var stdout, stderr bytes.Buffer

		// Act
		err := executor.Exec(
			context.Background(),
			workDir,
			[]string{"golangci-lint", "run"},
			strings.NewReader(""),
			&stdout,
			&stderr,
		)

		// Assert
		if err == nil {
			t.Fatal("expected command lookup error when golangci-lint is absent everywhere")
		}
		if !strings.Contains(err.Error(), `resolve command "golangci-lint" on PATH`) {
			t.Fatalf("err = %q, want command lookup failure", err.Error())
		}
	})

	t.Run("should preserve storage errors when lint lookup fails for another reason", func(t *testing.T) {
		// Arrange
		workDir := t.TempDir()
		goBinDir := filepath.Join(t.TempDir(), "go-bin")
		if err := os.MkdirAll(goBinDir, 0o755); err != nil {
			t.Fatalf("mkdir go bin dir: %v", err)
		}
		goBinary := filepath.Join(goBinDir, "go")
		if err := os.WriteFile(goBinary, []byte("#!/bin/sh\nprintf 'go\\n'\n"), 0o755); err != nil {
			t.Fatalf("write fake go: %v", err)
		}

		storeRoot := t.TempDir()
		if err := os.WriteFile(filepath.Join(storeRoot, "golangci-lint"), []byte("corrupted"), 0o644); err != nil {
			t.Fatalf("write corrupted lint root: %v", err)
		}
		lintStore := lintlocal.New(storeRoot)

		executor := New(Config{
			Resolver: stubResolver{
				currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.25.7",
						GoSource:    "go.mod",
						LintVersion: "v2.11.2",
						LintSource:  "remote",
					}, nil
				},
			},
			GoLocator: stubLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return goBinary, nil
				},
			},
			LintLocator: lintStore,
			PathEnv:     "",
		})
		var stdout, stderr bytes.Buffer

		// Act
		err := executor.Exec(
			context.Background(),
			workDir,
			[]string{"go", "version"},
			strings.NewReader(""),
			&stdout,
			&stderr,
		)

		// Assert
		if err == nil {
			t.Fatal("expected storage error when golangci-lint store root is a regular file")
		}
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("err = %v, want non-NotExist storage error", err)
		}
		if !errors.Is(err, syscall.ENOTDIR) {
			t.Fatalf("errors.Is(%v, syscall.ENOTDIR) = false, want true", err)
		}
	})

	t.Run("should report missing Go when selected Go is unavailable", func(t *testing.T) {
		// Arrange
		workDir := t.TempDir()
		executor := New(Config{
			Resolver: stubResolver{
				currentFn: func(ctx context.Context, gotWorkDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.25.7",
						GoSource:    "go.mod",
						LintVersion: "v2.11.2",
						LintSource:  "remote",
					}, nil
				},
			},
			GoLocator: stubLocator{
				goBinaryPathFn: func(ctx context.Context, version string) (string, error) {
					return "", fmt.Errorf("go version %s is not installed: %w", version, os.ErrNotExist)
				},
			},
			PathEnv: "",
		})
		var stdout, stderr bytes.Buffer

		// Act
		err := executor.Exec(
			context.Background(),
			workDir,
			[]string{"go", "version"},
			strings.NewReader(""),
			&stdout,
			&stderr,
		)

		// Assert
		if err == nil {
			t.Fatal("expected error when selected Go is unavailable")
		}
		if !strings.Contains(err.Error(), "go version 1.25.7 is not installed") {
			t.Fatalf("err = %q, want missing Go error", err.Error())
		}
	})
}
