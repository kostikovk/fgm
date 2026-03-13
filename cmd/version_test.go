package cmd

import (
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/testutil"
)

func TestVersionCommand_PrintsBuildInfo(t *testing.T) {
	t.Parallel()

	root := NewRootCmd(&app.App{
		BuildInfo: app.BuildInfo{
			Version: "v0.1.0",
			Commit:  "abc1234",
			Date:    "2026-03-11T10:00:00Z",
		},
	})

	stdout, stderr, err := testutil.ExecuteCommand(t, root, "version")
	if err != nil {
		t.Fatalf("execute version: %v\nstderr:\n%s", err, stderr)
	}

	for _, want := range []string{
		"fgm v0.1.0",
		"commit abc1234",
		"date 2026-03-11T10:00:00Z",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestVersionCommand_UsesDefaultsForMissingBuildInfo(t *testing.T) {
	t.Parallel()

	root := NewRootCmd(&app.App{})
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "version")
	if err != nil {
		t.Fatalf("execute version: %v\nstderr:\n%s", err, stderr)
	}

	for _, want := range []string{
		"fgm dev",
		"commit unknown",
		"date unknown",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}
