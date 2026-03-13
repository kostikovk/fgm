package cmd

import (
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/testutil"
)

func TestLintInitCommand_RejectsInvalidPreset(t *testing.T) {
	t.Parallel()

	root := NewRootCmd(&app.App{})
	_, stderr, err := testutil.ExecuteCommand(t, root, "lint", "init", "--preset", "nonexistent")
	if err == nil {
		t.Fatal("expected command to fail for invalid preset")
	}
	if !strings.Contains(stderr, "invalid preset") {
		t.Fatalf("stderr = %q, want invalid preset message", stderr)
	}
}
