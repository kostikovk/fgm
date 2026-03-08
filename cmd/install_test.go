package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

type stubGoInstaller struct {
	installGoVersionFn func(ctx context.Context, version string) (string, error)
}

func (s stubGoInstaller) InstallGoVersion(ctx context.Context, version string) (string, error) {
	return s.installGoVersionFn(ctx, version)
}

func TestInstallGoCommand_InstallsVersion(t *testing.T) {
	t.Parallel()

	application := app.New(app.Config{
		Resolver: resolve.New(nil),
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "1.25.7" {
					t.Fatalf("version = %q, want %q", version, "1.25.7")
				}
				return "/tmp/fgm/go/1.25.7", nil
			},
		},
	})

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "install", "go", "1.25.7")
	if err != nil {
		t.Fatalf("execute install go: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Installed Go 1.25.7") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Installed Go 1.25.7")
	}
}
