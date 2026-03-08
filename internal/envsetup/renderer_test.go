package envsetup

import (
	"context"
	"strings"
	"testing"
)

type stubGoStore struct {
	shimDirFn func() string
}

func (s stubGoStore) ShimDir() string {
	return s.shimDirFn()
}

func TestRender_AutoDetectsZsh(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return "/tmp/fgm/shims" },
		},
		ShellPath: "/bin/zsh",
		GOOS:      "darwin",
	})

	lines, err := renderer.Render(context.Background(), "")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, `export PATH="/tmp/fgm/shims":$PATH`) {
		t.Fatalf("joined = %q, want zsh export line", joined)
	}
}

func TestRender_UsesExplicitFishOverride(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return "/tmp/fgm/shims" },
		},
		ShellPath: "/bin/zsh",
		GOOS:      "darwin",
	})

	lines, err := renderer.Render(context.Background(), "fish")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, `set -gx PATH "/tmp/fgm/shims" $PATH`) {
		t.Fatalf("joined = %q, want fish line", joined)
	}
}

func TestRender_AutoDetectsBash(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return "/tmp/fgm/shims" },
		},
		ShellPath: "/bin/bash",
		GOOS:      "linux",
	})

	lines, err := renderer.Render(context.Background(), "")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, `export PATH="/tmp/fgm/shims":$PATH`) {
		t.Fatalf("joined = %q, want bash export line", joined)
	}
}

func TestRender_ErrorForUndetectedShell(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return "/tmp/fgm/shims" },
		},
		ShellPath: "",
		GOOS:      "linux",
	})

	_, err := renderer.Render(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "could not detect shell") {
		t.Fatalf("err = %q, want 'could not detect shell'", err)
	}
}

func TestRender_ErrorForUnsupportedShell(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return "/tmp/fgm/shims" },
		},
		ShellPath: "/bin/zsh",
		GOOS:      "darwin",
	})

	_, err := renderer.Render(context.Background(), "tcsh")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Fatalf("err = %q, want 'unsupported shell'", err)
	}
}

func TestRender_PowershellOnWindows(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return `C:\fgm\shims` },
		},
		ShellPath: `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
		GOOS:      "windows",
	})

	lines, err := renderer.Render(context.Background(), "")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "$env:PATH") {
		t.Fatalf("joined = %q, want powershell $env:PATH line", joined)
	}
}

func TestDetectShell_WindowsUnknownShell(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return `C:\fgm\shims` },
		},
		ShellPath: `C:\Windows\System32\cmd.exe`,
		GOOS:      "windows",
	})

	_, err := renderer.Render(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "could not detect shell") {
		t.Fatalf("err = %q, want 'could not detect shell'", err)
	}
}

func TestDetectShell_UnixUnknownShell(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return "/tmp/fgm/shims" },
		},
		ShellPath: "/bin/tcsh",
		GOOS:      "darwin",
	})

	_, err := renderer.Render(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "could not detect shell") {
		t.Fatalf("err = %q, want 'could not detect shell'", err)
	}
}

func TestNew_DefaultsGOOSToRuntime(t *testing.T) {
	t.Parallel()

	renderer := New(Config{
		GoStore: stubGoStore{
			shimDirFn: func() string { return "/tmp/fgm/shims" },
		},
		ShellPath: "/bin/bash",
	})

	lines, err := renderer.Render(context.Background(), "bash")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, `export PATH="/tmp/fgm/shims":$PATH`) {
		t.Fatalf("joined = %q, want bash export line", joined)
	}
}
