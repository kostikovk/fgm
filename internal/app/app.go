package app

import (
	"context"
	"io"
)

// Resolver returns the selected toolchain information for a workspace.
type Resolver interface {
	Current(ctx context.Context, workDir string) (Selection, error)
}

// GoLocalVersionStore lists installed Go toolchains for the active platform.
type GoLocalVersionStore interface {
	ListLocalGoVersions(ctx context.Context) ([]string, error)
	HasGoVersion(ctx context.Context, version string) (bool, error)
	GlobalGoVersion(ctx context.Context) (string, bool, error)
	SetGlobalGoVersion(ctx context.Context, version string) error
	DeleteGoVersion(ctx context.Context, version string) (string, error)
	GoBinaryPath(ctx context.Context, version string) (string, error)
	EnsureShims() error
	ShimDir() string
}

// GoRemoteVersionProvider lists remotely available Go toolchains.
type GoRemoteVersionProvider interface {
	ListRemoteGoVersions(ctx context.Context) ([]string, error)
}

// LintVersion describes a remotely available golangci-lint version.
type LintVersion struct {
	Version     string
	Recommended bool
}

// LintRemoteVersionProvider lists remotely available golangci-lint versions.
type LintRemoteVersionProvider interface {
	ListRemoteLintVersions(ctx context.Context, goVersion string) ([]LintVersion, error)
}

// LintLocalVersionStore lists installed golangci-lint versions.
type LintLocalVersionStore interface {
	ListLocalLintVersions(ctx context.Context) ([]string, error)
	DeleteLintVersion(ctx context.Context, version string) (string, error)
}

// GoInstaller installs Go toolchains into the local FGM-managed store.
type GoInstaller interface {
	InstallGoVersion(ctx context.Context, version string) (string, error)
}

// LintInstaller installs golangci-lint into the local FGM-managed store.
type LintInstaller interface {
	InstallLintVersion(ctx context.Context, version string) (string, error)
}

// ImportedGo describes a Go installation imported into FGM.
type ImportedGo struct {
	Version string
	Path    string
}

// ImportedLint describes a golangci-lint installation imported into FGM.
type ImportedLint struct {
	Version string
	Path    string
}

// GoImporter imports existing Go installations into FGM.
type GoImporter interface {
	ImportAuto(ctx context.Context) ([]ImportedGo, error)
}

// LintImporter imports existing golangci-lint installations into FGM.
type LintImporter interface {
	ImportAuto(ctx context.Context) ([]ImportedLint, error)
}

// Doctor reports environment and configuration diagnostics for FGM.
type Doctor interface {
	Diagnose(ctx context.Context, workDir string) ([]string, error)
}

// Executor runs commands with a selected Go toolchain on PATH.
type Executor interface {
	Exec(ctx context.Context, workDir string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error
}

// EnvRenderer prints shell-specific environment setup for FGM.
type EnvRenderer interface {
	Render(ctx context.Context, shell string) ([]string, error)
}

// Selection describes the currently selected toolchain.
type Selection struct {
	GoVersion   string
	GoSource    string
	LintVersion string
	LintSource  string
}

// App holds the services used by Cobra commands.
type App struct {
	Resolver           Resolver
	GoStore            GoLocalVersionStore
	LintStore          LintLocalVersionStore
	GoRemoteProvider   GoRemoteVersionProvider
	LintRemoteProvider LintRemoteVersionProvider
	GoInstaller        GoInstaller
	LintInstaller      LintInstaller
	GoImporter         GoImporter
	LintImporter       LintImporter
	Doctor             Doctor
	Executor           Executor
	EnvRenderer        EnvRenderer
}

// Config configures an App instance.
type Config struct {
	Resolver           Resolver
	GoStore            GoLocalVersionStore
	LintStore          LintLocalVersionStore
	GoRemoteProvider   GoRemoteVersionProvider
	LintRemoteProvider LintRemoteVersionProvider
	GoInstaller        GoInstaller
	LintInstaller      LintInstaller
	GoImporter         GoImporter
	LintImporter       LintImporter
	Doctor             Doctor
	Executor           Executor
	EnvRenderer        EnvRenderer
}

// New constructs the application service container.
func New(config Config) *App {
	return &App{
		Resolver:           config.Resolver,
		GoStore:            config.GoStore,
		LintStore:          config.LintStore,
		GoRemoteProvider:   config.GoRemoteProvider,
		LintRemoteProvider: config.LintRemoteProvider,
		GoInstaller:        config.GoInstaller,
		LintInstaller:      config.LintInstaller,
		GoImporter:         config.GoImporter,
		LintImporter:       config.LintImporter,
		Doctor:             config.Doctor,
		Executor:           config.Executor,
		EnvRenderer:        config.EnvRenderer,
	}
}
