package goimport

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/goenv"
)

// Registry registers existing Go installations in the FGM-managed store.
type Registry interface {
	RegisterExistingGoInstallation(version string, goroot string) (string, error)
}

// Importer discovers existing Go installations and registers them into FGM.
type Importer struct {
	config Config
}

// Config configures an Importer.
type Config struct {
	// Candidates are system GOROOT directories whose version comes from `go version`.
	Candidates []string
	Registry   Registry
	// ModCache returns GOMODCACHE, where Go's own toolchain switching caches toolchains;
	// nil or "" skips that source.
	ModCache func(ctx context.Context) string
	// SDKDir is where golang.org/dl installs SDKs (usually ~/sdk); "" skips that source.
	SDKDir string
	// GOOS and GOARCH select module-cache toolchains; runtime values are used when empty.
	GOOS   string
	GOARCH string
}

// New constructs an Importer.
func New(config Config) *Importer {
	if config.GOOS == "" {
		config.GOOS = runtime.GOOS
	}
	if config.GOARCH == "" {
		config.GOARCH = runtime.GOARCH
	}
	return &Importer{config: config}
}

// ImportAuto discovers Go installations from system paths, the module cache, and
// golang.org/dl SDKs, and registers them. Entries with SkipReason were found but not imported.
func (i *Importer) ImportAuto(ctx context.Context) ([]app.ImportedGo, error) {
	seenVersions := make(map[string]string)
	var results []app.ImportedGo

	for _, candidate := range i.config.Candidates {
		goBinary := filepath.Join(candidate, "bin", goBinaryName())
		if _, err := os.Stat(goBinary); err != nil {
			continue
		}

		version, err := detectVersion(ctx, goBinary)
		if err != nil {
			return nil, err
		}
		if _, seen := seenVersions[version]; seen {
			continue
		}

		if _, err := i.config.Registry.RegisterExistingGoInstallation(version, candidate); err != nil {
			return nil, err
		}

		seenVersions[version] = candidate
		results = append(results, app.ImportedGo{
			Version: version,
			Path:    candidate,
		})
	}

	for _, sdk := range i.toolchainCandidates(ctx) {
		if reason := sdk.problem(); reason != "" {
			sdk.SkipReason = reason
		} else if previous, seen := seenVersions[sdk.Version]; seen {
			sdk.SkipReason = "already imported from " + previous
		} else {
			if _, err := i.config.Registry.RegisterExistingGoInstallation(sdk.Version, sdk.Path); err != nil {
				return nil, err
			}
			seenVersions[sdk.Version] = sdk.Path
		}
		results = append(results, sdk.ImportedGo)
	}

	return results, nil
}

type sdkCandidate struct {
	app.ImportedGo
}

// problem explains why a toolchain directory cannot be imported, or returns "".
// The VERSION file is trusted instead of running `go version`, so importing stays read-only.
func (c sdkCandidate) problem() string {
	if info, err := os.Stat(filepath.Join(c.Path, "bin", goBinaryName())); err != nil || info.IsDir() {
		return "bin/" + goBinaryName() + " is missing"
	}
	content, err := os.ReadFile(filepath.Join(c.Path, "VERSION"))
	if err != nil {
		return "VERSION file is missing"
	}
	declared := strings.TrimSpace(strings.SplitN(string(content), "\n", 2)[0])
	if declared != "go"+c.Version {
		return fmt.Sprintf("VERSION says %s, directory says go%s", declared, c.Version)
	}
	return ""
}

const moduleCacheToolchainPrefix = "toolchain@v0.0.1-go"

// toolchainCandidates lists module-cache toolchains for this platform, then golang.org/dl SDKs,
// each in lexical order.
func (i *Importer) toolchainCandidates(ctx context.Context) []sdkCandidate {
	var candidates []sdkCandidate
	if i.config.ModCache != nil {
		if modCache := i.config.ModCache(ctx); modCache != "" {
			platform := "." + i.config.GOOS + "-" + i.config.GOARCH
			entries, _ := os.ReadDir(filepath.Join(modCache, "golang.org"))
			for _, entry := range entries {
				version, ok := strings.CutPrefix(entry.Name(), moduleCacheToolchainPrefix)
				if !ok || !entry.IsDir() {
					continue
				}
				version, ok = strings.CutSuffix(version, platform)
				if !ok {
					continue
				}
				candidates = append(candidates, sdkCandidate{app.ImportedGo{
					Version: version,
					Path:    filepath.Join(modCache, "golang.org", entry.Name()),
					Source:  "module cache",
				}})
			}
		}
	}
	if i.config.SDKDir != "" {
		entries, _ := os.ReadDir(i.config.SDKDir)
		for _, entry := range entries {
			version, ok := strings.CutPrefix(entry.Name(), "go")
			if !ok || !entry.IsDir() || version == "" || version[0] < '0' || version[0] > '9' {
				continue
			}
			candidates = append(candidates, sdkCandidate{app.ImportedGo{
				Version: version,
				Path:    filepath.Join(i.config.SDKDir, entry.Name()),
				Source:  "golang.org/dl",
			}})
		}
	}
	return candidates
}

// ModCacheLocator returns a lazy GOMODCACHE lookup: `go env GOMODCACHE` of the go on PATH,
// else $GOMODCACHE, else the first $GOPATH entry plus /pkg/mod, else ~/go/pkg/mod.
func ModCacheLocator(getenv func(string) string, pathEnv string, home string) func(ctx context.Context) string {
	return func(ctx context.Context) string {
		if goroot := gorootFromPathGo(pathEnv); goroot != "" {
			cmd := exec.CommandContext(ctx, filepath.Join(goroot, "bin", goBinaryName()), "env", "GOMODCACHE")
			cmd.Env, _ = goenv.Apply(cmd.Environ(), goenv.ModeLocal)
			if output, err := cmd.Output(); err == nil {
				if value := strings.TrimSpace(string(output)); value != "" {
					return value
				}
			}
		}
		if value := getenv("GOMODCACHE"); value != "" {
			return value
		}
		if gopath := filepath.SplitList(getenv("GOPATH")); len(gopath) > 0 && gopath[0] != "" {
			return filepath.Join(gopath[0], "pkg", "mod")
		}
		if home == "" {
			return ""
		}
		return filepath.Join(home, "go", "pkg", "mod")
	}
}

// DefaultCandidates returns common locations that may already contain Go installations.
func DefaultCandidates(pathEnv string) []string {
	seen := make(map[string]struct{})
	var ordered []string
	addCandidate := func(path string) {
		if path == "" {
			return
		}
		if _, exists := seen[path]; exists {
			return
		}
		seen[path] = struct{}{}
		ordered = append(ordered, path)
	}

	addCandidate(gorootFromPathGo(pathEnv))

	for _, path := range []string{
		"/usr/local/go",
		"/opt/homebrew/opt/go/libexec",
		"/usr/local/opt/go/libexec",
	} {
		addCandidate(path)
	}

	for _, pattern := range []string{
		"/opt/homebrew/Cellar/go/*/libexec",
		"/usr/local/Cellar/go/*/libexec",
	} {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			addCandidate(match)
		}
	}

	return ordered
}

func detectVersion(ctx context.Context, goBinary string) (string, error) {
	cmd := exec.CommandContext(ctx, goBinary, "version")
	// An inherited GOTOOLCHAIN or GOROOT would report (or break on) another toolchain.
	cmd.Env, _ = goenv.Apply(cmd.Environ(), goenv.ModeLocal)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run %s version: %w", goBinary, err)
	}

	fields := strings.Fields(string(output))
	if len(fields) < 3 {
		return "", fmt.Errorf("parse go version output: %q", strings.TrimSpace(string(output)))
	}

	version := strings.TrimPrefix(fields[2], "go")
	if version == "" {
		return "", fmt.Errorf("parse go version output: %q", strings.TrimSpace(string(output)))
	}

	return version, nil
}

func gorootFromPathGo(pathEnv string) string {
	binaryName := goBinaryName()
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, binaryName)
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Size() > 0 && info.Size() < 256 {
			if data, readErr := os.ReadFile(candidate); readErr == nil && strings.Contains(string(data), "fgm __shim") {
				continue
			}
		}
		return filepath.Dir(filepath.Dir(candidate))
	}

	return ""
}

func goBinaryName() string {
	if runtime.GOOS == "windows" {
		return "go.exe"
	}
	return "go"
}
