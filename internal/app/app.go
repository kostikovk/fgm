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
	LintBinaryPath(ctx context.Context, version string) (string, error)
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

// FGMUpdateStatus describes whether a newer FGM CLI version is available.
type FGMUpdateStatus struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	UpgradeCommand  string
}

// FGMUpdateChecker checks or reads cached FGM release availability.
type FGMUpdateChecker interface {
	CheckLatest(ctx context.Context) (FGMUpdateStatus, bool, error)
	CachedNotice(ctx context.Context) (FGMUpdateStatus, bool)
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

// ExecOptions configures how an Executor builds the child process environment.
type ExecOptions struct {
	// GoToolchain is "local" (used when empty) or "inherit"; see internal/goenv.
	GoToolchain string
	// AllowUnselectedLint lets golangci-lint resolve from the inherited PATH when the selected version is missing.
	AllowUnselectedLint bool
}

// Executor runs commands with a selected Go toolchain on PATH.
type Executor interface {
	Exec(ctx context.Context, workDir string, args []string, options ExecOptions, stdin io.Reader, stdout io.Writer, stderr io.Writer) error
}

// EnvRenderer prints shell-specific environment setup for FGM.
type EnvRenderer interface {
	Render(ctx context.Context, shell string) ([]string, error)
}

// Selection describes the currently selected toolchain.
type Selection struct {
	GoVersion string
	// GoSource is the go.work/go.mod path, "config", "latest-patch", or "global".
	GoSource string
	// GoSourceFile is the file the Go version came from; empty for "global".
	GoSourceFile string
	// GoBase is the go directive a "latest-patch" selection was raised from.
	GoBase      string
	LintVersion string
	// LintSource is "config", "detected", "local", or "remote".
	LintSource string
	// LintSourceFile and LintSourceLine locate a "config" or "detected" lint pin.
	LintSourceFile string
	LintSourceLine int
	// Warnings describe selections that may be stale or ambiguous.
	Warnings []string
}

// CurrentReportSchema is the version of the `fgm current --json` document. Changes within a
// schema version are additive only.
const CurrentReportSchema = 1

// CurrentReport is the machine-readable selection for one directory (`fgm current --json`).
type CurrentReport struct {
	Schema       int         `json:"schema"`
	Dir          string      `json:"dir"`
	ModuleRoot   string      `json:"module_root"`
	Go           *GoReport   `json:"go"`
	GolangCILint *LintReport `json:"golangci_lint"`
	Warnings     []string    `json:"warnings"`
	Error        string      `json:"error,omitempty"`
}

// GoReport describes the selected Go toolchain in a CurrentReport.
type GoReport struct {
	Version string `json:"version"`
	// Source is "go.work", "go.mod", "config", "latest-patch", or "global".
	Source     string `json:"source"`
	SourceFile string `json:"source_file,omitempty"`
	Base       string `json:"base,omitempty"`
	Installed  bool   `json:"installed"`
	Path       string `json:"path,omitempty"`
}

// LintReport describes the selected golangci-lint in a CurrentReport.
type LintReport struct {
	Version string `json:"version"`
	// Source is "config", "detected", "local", or "remote".
	Source     string `json:"source"`
	SourceFile string `json:"source_file,omitempty"`
	SourceLine int    `json:"source_line,omitempty"`
	Installed  bool   `json:"installed"`
	Path       string `json:"path,omitempty"`
	// Compatible is null when the compatibility catalog does not know the version.
	Compatible *bool `json:"compatible"`
}

// CurrentReporter builds CurrentReports for one directory or every module below it.
type CurrentReporter interface {
	Report(ctx context.Context, dir string) (CurrentReport, error)
	ReportAll(ctx context.Context, dir string) ([]CurrentReport, error)
}

// CheckFinding is one repository consistency problem reported by `fgm check`.
type CheckFinding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"` // "ERROR", "WARN", "INFO"
	// File is relative to the report's dir.
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

// CheckReport is a CurrentReport plus the consistency findings for that module.
type CheckReport struct {
	CurrentReport
	Findings []CheckFinding `json:"findings"`
}

// RepoChecker runs offline repository consistency checks.
type RepoChecker interface {
	Check(ctx context.Context, dir string) (CheckReport, error)
	CheckAll(ctx context.Context, dir string) ([]CheckReport, error)
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
	GoResolver          Resolver
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
	FGMUpdateChecker    FGMUpdateChecker
	CurrentReporter     CurrentReporter
	RepoChecker         RepoChecker
}
