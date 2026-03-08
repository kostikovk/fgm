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
