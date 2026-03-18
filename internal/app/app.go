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

// BuildInfo describes the current FGM build metadata.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// GoImporter imports existing Go installations into FGM.
type GoImporter interface {
	ImportAuto(ctx context.Context) ([]ImportedGo, error)
}

// LintImporter imports existing golangci-lint installations into FGM.
type LintImporter interface {
	ImportAuto(ctx context.Context) ([]ImportedLint, error)
}

// GoUpgradeResult describes a Go upgrade action.
type GoUpgradeResult struct {
	Version     string
	Path        string
	LintVersion string
	DryRun      bool
}

// GoUpgradeOptions configures a Go upgrade operation.
type GoUpgradeOptions struct {
	WorkDir  string
	Version  string
	DryRun   bool
	WithLint bool
}

// GoUpgrader upgrades Go at global or project scope.
type GoUpgrader interface {
	UpgradeGlobal(ctx context.Context, options GoUpgradeOptions) (GoUpgradeResult, error)
	UpgradeProject(ctx context.Context, options GoUpgradeOptions) (GoUpgradeResult, error)
}

// LintUpgradeResult describes a golangci-lint upgrade action.
type LintUpgradeResult struct {
	Version   string
	GoVersion string
	DryRun    bool
}

// LintUpgradeOptions configures a golangci-lint upgrade operation.
type LintUpgradeOptions struct {
	WorkDir string
	Version string
	DryRun  bool
}

// LintUpgrader upgrades golangci-lint to a compatible version.
type LintUpgrader interface {
	Upgrade(ctx context.Context, options LintUpgradeOptions) (LintUpgradeResult, error)
}

// Doctor reports environment and configuration diagnostics for FGM.
type Doctor interface {
	Diagnose(ctx context.Context, workDir string) ([]DoctorFinding, error)
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

// LintConfigGenerator generates golangci-lint configuration files.
type LintConfigGenerator interface {
	Generate(ctx context.Context, opts LintConfigOptions) ([]byte, error)
}

// LintConfigOptions configures lint config generation.
type LintConfigOptions struct {
	WorkDir     string
	Preset      string // "minimal", "standard", "strict"
	WithImports bool
	Force       bool
}

// LintDoctor audits an existing golangci-lint configuration.
type LintDoctor interface {
	Diagnose(ctx context.Context, workDir string) ([]LintFinding, error)
}

// LintFinding describes a single lint config diagnostic.
type LintFinding struct {
	Severity string // "ERROR", "WARN", "INFO", "OK"
	Message  string
}

// DoctorFinding describes a single environment diagnostic.
type DoctorFinding struct {
	Severity string // "OK", "WARN"
	Message  string
	FixKind  string // "", "shell_profile"
}

// ProfileInstaller appends shell integration to the user's shell profile.
type ProfileInstaller interface {
	InstallProfile(shell string) (profilePath string, modified bool, err error)
}

// App holds the services used by Cobra commands.
type App struct {
	Resolver            Resolver
	GoStore             GoLocalVersionStore
	LintStore           LintLocalVersionStore
	GoRemoteProvider    GoRemoteVersionProvider
	LintRemoteProvider  LintRemoteVersionProvider
	GoInstaller         GoInstaller
	LintInstaller       LintInstaller
	GoImporter          GoImporter
	LintImporter        LintImporter
	GoUpgrader          GoUpgrader
	LintUpgrader        LintUpgrader
	BuildInfo           BuildInfo
	Doctor              Doctor
	Executor            Executor
	EnvRenderer         EnvRenderer
	LintConfigGenerator LintConfigGenerator
	LintDoctor          LintDoctor
	ProfileInstaller    ProfileInstaller
}
