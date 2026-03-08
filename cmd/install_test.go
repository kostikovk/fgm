package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/currenttoolchain"
	"github.com/koskosovu4/fgm/internal/resolve"
	"github.com/koskosovu4/fgm/internal/testutil"
)

type stubGoInstaller struct {
	installGoVersionFn func(ctx context.Context, version string) (string, error)
}

func (s stubGoInstaller) InstallGoVersion(ctx context.Context, version string) (string, error) {
	return s.installGoVersionFn(ctx, version)
}

type stubSelectionResolver struct {
	currentFn func(ctx context.Context, workDir string) (app.Selection, error)
}

func (s stubSelectionResolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	return s.currentFn(ctx, workDir)
}

type stubLintInstaller struct {
	installLintVersionFn func(ctx context.Context, version string) (string, error)
}

func (s stubLintInstaller) InstallLintVersion(ctx context.Context, version string) (string, error) {
	return s.installLintVersionFn(ctx, version)
}

func TestInstallCommand_RejectsNilResolver(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver:      nil,
		GoInstaller:   stubGoInstaller{installGoVersionFn: func(ctx context.Context, version string) (string, error) { return "", nil }},
		LintInstaller: stubLintInstaller{installLintVersionFn: func(ctx context.Context, version string) (string, error) { return "", nil }},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install")
	if err == nil {
		t.Fatal("expected an error when Resolver is nil")
	}
	if !strings.Contains(err.Error(), "resolver is not configured") {
		t.Fatalf("err = %q, want resolver not configured error", err)
	}
}

func TestInstallCommand_RejectsNilGoInstaller(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7"}, nil
			},
		},
		GoInstaller:   nil,
		LintInstaller: stubLintInstaller{installLintVersionFn: func(ctx context.Context, version string) (string, error) { return "", nil }},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install")
	if err == nil {
		t.Fatal("expected an error when GoInstaller is nil")
	}
	if !strings.Contains(err.Error(), "Go installer is not configured") {
		t.Fatalf("err = %q, want Go installer not configured error", err)
	}
}

func TestInstallCommand_RejectsNilLintInstaller(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7"}, nil
			},
		},
		GoInstaller:   stubGoInstaller{installGoVersionFn: func(ctx context.Context, version string) (string, error) { return "/tmp/go", nil }},
		LintInstaller: nil,
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install")
	if err == nil {
		t.Fatal("expected an error when LintInstaller is nil")
	}
	if !strings.Contains(err.Error(), "golangci-lint installer is not configured") {
		t.Fatalf("err = %q, want lint installer not configured error", err)
	}
}

func TestInstallCommand_SkipsLintWhenVersionEmpty(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7", LintVersion: ""}, nil
			},
		},
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/1.25.7", nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				t.Fatal("InstallLintVersion should not be called when LintVersion is empty")
				return "", nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "install")
	if err != nil {
		t.Fatalf("execute install: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "Installed Go 1.25.7") {
		t.Fatalf("stdout = %q, want it to contain Go install line", stdout)
	}
	if strings.Contains(stdout, "golangci-lint") {
		t.Fatalf("stdout = %q, want no lint install line when LintVersion is empty", stdout)
	}
}

func TestInstallGoCommand_RejectsNilGoInstaller(t *testing.T) {
	t.Parallel()

	application := &app.App{GoInstaller: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install", "go", "1.25.7")
	if err == nil {
		t.Fatal("expected an error when GoInstaller is nil")
	}
	if !strings.Contains(err.Error(), "Go installer is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestInstallLintCommand_RejectsNilLintInstaller(t *testing.T) {
	t.Parallel()

	application := &app.App{LintInstaller: nil}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install", "golangci-lint", "v2.11.2")
	if err == nil {
		t.Fatal("expected an error when LintInstaller is nil")
	}
	if !strings.Contains(err.Error(), "golangci-lint installer is not configured") {
		t.Fatalf("err = %q, want not configured error", err)
	}
}

func TestInstallGoCommand_InstallsVersion(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "1.25.7" {
					t.Fatalf("version = %q, want %q", version, "1.25.7")
				}
				return "/tmp/fgm/go/1.25.7", nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "install", "go", "1.25.7")
	if err != nil {
		t.Fatalf("execute install go: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Installed Go 1.25.7") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Installed Go 1.25.7")
	}
}

func TestInstallLintCommand_InstallsVersion(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: resolve.New(nil),
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "v2.11.2" {
					t.Fatalf("version = %q, want %q", version, "v2.11.2")
				}
				return "/tmp/fgm/golangci-lint/v2.11.2", nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "install", "golangci-lint", "v2.11.2")
	if err != nil {
		t.Fatalf("execute install golangci-lint: %v\nstderr:\n%s", err, stderr)
	}

	if !strings.Contains(stdout, "Installed golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Installed golangci-lint v2.11.2")
	}
}

func TestInstallCommand_InstallsResolvedGoAndLint(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{
					GoVersion:   "1.25.7",
					LintVersion: "v2.11.2",
				}, nil
			},
		},
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "1.25.7" {
					t.Fatalf("go version = %q, want %q", version, "1.25.7")
				}
				return "/tmp/fgm/go/1.25.7", nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "v2.11.2" {
					t.Fatalf("lint version = %q, want %q", version, "v2.11.2")
				}
				return "/tmp/fgm/golangci-lint/v2.11.2", nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "install")
	if err != nil {
		t.Fatalf("execute install: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "Installed Go 1.25.7") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Installed Go 1.25.7")
	}
	if !strings.Contains(stdout, "Installed golangci-lint v2.11.2") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Installed golangci-lint v2.11.2")
	}
}

func TestInstallCommand_ResolverErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{}, fmt.Errorf("resolver boom")
			},
		},
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", nil
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install")
	if err == nil {
		t.Fatal("expected an error when resolver fails")
	}
	if !strings.Contains(err.Error(), "resolver boom") {
		t.Fatalf("err = %q, want resolver boom", err)
	}
}

func TestInstallCommand_GoInstallerErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7"}, nil
			},
		},
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("go install boom")
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", nil
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install")
	if err == nil {
		t.Fatal("expected an error when GoInstaller fails")
	}
	if !strings.Contains(err.Error(), "go install boom") {
		t.Fatalf("err = %q, want go install boom", err)
	}
}

func TestInstallCommand_LintInstallerErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7", LintVersion: "v2.11.2"}, nil
			},
		},
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "/tmp/fgm/go/1.25.7", nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("lint install boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install")
	if err == nil {
		t.Fatal("expected an error when LintInstaller fails")
	}
	if !strings.Contains(err.Error(), "lint install boom") {
		t.Fatalf("err = %q, want lint install boom", err)
	}
}

func TestInstallGoCommand_InstallerErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("go install specific boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install", "go", "1.25.7")
	if err == nil {
		t.Fatal("expected an error when GoInstaller fails")
	}
	if !strings.Contains(err.Error(), "go install specific boom") {
		t.Fatalf("err = %q, want go install specific boom", err)
	}
}

func TestInstallLintCommand_InstallerErrorIsPropagated(t *testing.T) {
	t.Parallel()

	application := &app.App{
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				return "", fmt.Errorf("lint install specific boom")
			},
		},
	}

	root := NewRootCmd(application)
	_, _, err := testutil.ExecuteCommand(t, root, "install", "golangci-lint", "v2.11.2")
	if err == nil {
		t.Fatal("expected an error when LintInstaller fails")
	}
	if !strings.Contains(err.Error(), "lint install specific boom") {
		t.Fatalf("err = %q, want lint install specific boom", err)
	}
}

func TestInstallCommand_PrefersPinnedLintVersionFromRepoConfig(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	goMod := "module example.com/demo\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(tempDir, ".fgm.toml"),
		[]byte("[toolchain]\ngolangci_lint = \"v2.10.1\"\n"),
		0o644,
	); err != nil {
		t.Fatalf("write .fgm.toml: %v", err)
	}

	application := &app.App{
		Resolver: currenttoolchain.New(currenttoolchain.Config{
			GoResolver: resolve.New(nil),
		}),
		GoInstaller: stubGoInstaller{
			installGoVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "1.25.0" {
					t.Fatalf("go version = %q, want %q", version, "1.25.0")
				}
				return "/tmp/fgm/go/1.25.0", nil
			},
		},
		LintInstaller: stubLintInstaller{
			installLintVersionFn: func(ctx context.Context, version string) (string, error) {
				if version != "v2.10.1" {
					t.Fatalf("lint version = %q, want %q", version, "v2.10.1")
				}
				return "/tmp/fgm/golangci-lint/v2.10.1", nil
			},
		},
	}

	root := NewRootCmd(application)
	stdout, stderr, err := testutil.ExecuteCommand(t, root, "install", "--chdir", tempDir)
	if err != nil {
		t.Fatalf("execute install: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "Installed golangci-lint v2.10.1") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, "Installed golangci-lint v2.10.1")
	}
}
