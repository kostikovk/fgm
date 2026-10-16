package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kostikovk/fgm/internal/golocal"
	"github.com/kostikovk/fgm/internal/testutil"
)

func TestGeneratedGoShim(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shims are generated on macOS and Linux only")
	}
	if testing.Short() {
		t.Skip("builds the fgm binary")
	}

	t.Run("go shim", func(t *testing.T) {
		t.Run("should pass flags verbatim and keep the tool's exit code when go test runs through the shim", func(t *testing.T) {
			// Arrange
			binDir := t.TempDir()
			build := exec.Command("go", "build", "-o", filepath.Join(binDir, "fgm"), ".")
			if output, err := build.CombinedOutput(); err != nil {
				t.Fatalf("build fgm: %v\n%s", err, output)
			}
			dataHome := t.TempDir()
			storeRoot := filepath.Join(dataHome, "fgm")
			fakeGo := filepath.Join(storeRoot, "go", "1.25.3", "bin", "go")
			if err := os.MkdirAll(filepath.Dir(fakeGo), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			script := "#!/bin/sh\nfor arg in \"$@\"; do printf '[%s]\\n' \"$arg\"; done\nexit 3\n"
			if err := testutil.WriteExecutable(fakeGo, []byte(script), 0o755); err != nil {
				t.Fatalf("write fake go: %v", err)
			}
			store := golocal.New(storeRoot, "")
			if err := store.EnsureShims(); err != nil {
				t.Fatalf("EnsureShims: %v", err)
			}
			module := t.TempDir()
			if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module m\n\ngo 1.25.3\n"), 0o644); err != nil {
				t.Fatalf("write go.mod: %v", err)
			}
			cmd := exec.Command(filepath.Join(store.ShimDir(), "go"), "test", "-run", "TestX", "-count=1", "./...")
			cmd.Dir = module
			cmd.Env = append(os.Environ(),
				"PATH="+binDir+string(os.PathListSeparator)+"/usr/bin"+string(os.PathListSeparator)+"/bin",
				"XDG_DATA_HOME="+dataHome,
				"FGM_NO_UPDATE_NOTIFIER=1",
			)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr

			// Act
			err := cmd.Run()

			// Assert
			exitErr, ok := errors.AsType[*exec.ExitError](err)
			if !ok || exitErr.ExitCode() != 3 {
				t.Fatalf("exit = %v; want exit status 3\nstderr:\n%s", err, stderr.String())
			}
			if want := "[test]\n[-run]\n[TestX]\n[-count=1]\n[./...]\n"; stdout.String() != want {
				t.Errorf("tool argv = %q; want %q", stdout.String(), want)
			}
		})
	})
}
