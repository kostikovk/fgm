package resolve

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/fgmconfig"
	"github.com/kostikovk/fgm/internal/versionutil"
)

// GlobalVersionSource returns the globally selected Go version.
type GlobalVersionSource interface {
	GlobalGoVersion(ctx context.Context) (string, bool, error)
}

// PatchCatalog returns the newest known patch release for a Go version's minor from offline data.
type PatchCatalog interface {
	LatestGoPatch(goVersion string) (string, bool, error)
}

// LocalVersionLister lists installed Go versions.
type LocalVersionLister interface {
	ListLocalGoVersions(ctx context.Context) ([]string, error)
}

// Config configures a Resolver. Only Global is needed without a latest-patch policy.
//
// Selection never uses the network: shims run it on every go call, and fgm current must
// report exactly what the shims and fgm exec run.
type Config struct {
	Global        GlobalVersionSource
	PatchCatalog  PatchCatalog
	LocalVersions LocalVersionLister
}

// Resolver locates Go toolchain metadata from workspace files.
type Resolver struct {
	config Config
}

// New constructs a Resolver.
func New(global GlobalVersionSource) *Resolver {
	return NewWithConfig(Config{Global: global})
}

// NewWithConfig constructs a Resolver with latest-patch data sources.
func NewWithConfig(config Config) *Resolver {
	return &Resolver{config: config}
}

type versionMetadata struct {
	path        string
	goDirective string
	toolchain   string
}

// Current resolves the currently selected Go version from the workspace.
//
// Order: .fgm.toml explicit go → toolchain directive (go.work, then go.mod) → go directive
// (raised to its newest patch under go = "latest-patch") → global default.
func (r *Resolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	policy, hasPolicy, err := fgmconfig.ResolveGoPolicy(workDir)
	if err != nil {
		return app.Selection{}, err
	}

	metadata, found, err := findVersionMetadata(workDir)
	if err != nil {
		return app.Selection{}, err
	}

	if hasPolicy && policy.Version != "" {
		if found && versionutil.CompareVersions(policy.Version, metadata.goDirective) < 0 {
			return app.Selection{}, fmt.Errorf(
				"%s pins go %s, lower than the go %s directive in %s",
				policy.Path, policy.Version, metadata.goDirective, metadata.path,
			)
		}
		return app.Selection{GoVersion: policy.Version, GoSource: "config", GoSourceFile: policy.Path}, nil
	}

	if found {
		if hasPolicy && policy.LatestPatch {
			return r.latestPatchSelection(ctx, metadata)
		}
		return directiveSelection(metadata), nil
	}

	if r.config.Global != nil {
		goVersion, ok, err := r.config.Global.GlobalGoVersion(ctx)
		if err != nil {
			return app.Selection{}, err
		}
		if ok {
			return app.Selection{GoVersion: goVersion, GoSource: "global"}, nil
		}
	}
	return app.Selection{}, fmt.Errorf(
		"go.mod not found from %s upward: %w (no go.work or go.mod found from %s and no global Go version is selected; pass --chdir <dir> or select a global version with \"fgm use go <version> --global\")",
		workDir,
		versionutil.ErrNotFound,
		workDir,
	)
}

// ValidateGoPin reports an error when version is lower than the go directive of the nearest
// go.work or go.mod, which Go itself would refuse to build with.
func ValidateGoPin(workDir string, version string) error {
	metadata, found, err := findVersionMetadata(workDir)
	if err != nil || !found {
		return err
	}
	if versionutil.CompareVersions(version, metadata.goDirective) < 0 {
		return fmt.Errorf("go %s is lower than the go %s directive in %s", version, metadata.goDirective, metadata.path)
	}
	return nil
}

func directiveSelection(metadata versionMetadata) app.Selection {
	version := metadata.goDirective
	if metadata.toolchain != "" {
		version = metadata.toolchain
	}
	return app.Selection{GoVersion: version, GoSource: metadata.path, GoSourceFile: metadata.path}
}

func (r *Resolver) latestPatchSelection(ctx context.Context, metadata versionMetadata) (app.Selection, error) {
	latest, warning, ok, err := r.latestPatch(ctx, metadata.goDirective)
	if err != nil {
		return app.Selection{}, err
	}
	if !ok {
		selection := directiveSelection(metadata)
		selection.Warnings = []string{fmt.Sprintf(
			"no release data for the latest go %s patch; using %s from %s",
			minorKey(metadata.goDirective), selection.GoVersion, filepath.Base(metadata.path),
		)}
		return selection, nil
	}
	if metadata.toolchain != "" && versionutil.CompareVersions(metadata.toolchain, latest) >= 0 {
		return directiveSelection(metadata), nil
	}

	selection := app.Selection{
		GoVersion:    latest,
		GoSource:     "latest-patch",
		GoSourceFile: metadata.path,
		GoBase:       metadata.goDirective,
	}
	if warning != "" {
		selection.Warnings = []string{warning}
	}
	return selection, nil
}

// latestPatch looks up the newest patch of goDirective's minor: embedded catalog, then installed
// versions (flagged as possibly stale). The result is never lower than goDirective itself.
func (r *Resolver) latestPatch(ctx context.Context, goDirective string) (string, string, bool, error) {
	if r.config.PatchCatalog != nil {
		version, ok, err := r.config.PatchCatalog.LatestGoPatch(goDirective)
		if err != nil {
			return "", "", false, err
		}
		if ok {
			return maxVersion(version, goDirective), "", true, nil
		}
	}
	if r.config.LocalVersions != nil {
		versions, err := r.config.LocalVersions.ListLocalGoVersions(ctx)
		if err != nil {
			return "", "", false, err
		}
		if version, ok := newestPatch(versions, goDirective); ok {
			warning := fmt.Sprintf(
				"latest go %s patch taken from installed versions (%s); the selection may be stale",
				minorKey(goDirective), version,
			)
			return maxVersion(version, goDirective), warning, true, nil
		}
	}
	return "", "", false, nil
}

func newestPatch(versions []string, goVersion string) (string, bool) {
	want := minorKey(goVersion)
	best := ""
	for _, version := range versions {
		version = strings.TrimPrefix(strings.TrimSpace(version), "go")
		if minorKey(version) != want || strings.Contains(version, "rc") || strings.Contains(version, "beta") {
			continue
		}
		if best == "" || versionutil.CompareVersions(version, best) > 0 {
			best = version
		}
	}
	return best, best != ""
}

func maxVersion(left string, right string) string {
	if versionutil.CompareVersions(left, right) >= 0 {
		return left
	}
	return right
}

// minorKey returns "major.minor" for a Go version such as "1.26.3" or "go1.19".
func minorKey(version string) string {
	parts := strings.SplitN(strings.TrimPrefix(version, "go"), ".", 3)
	if len(parts) < 2 {
		return parts[0]
	}
	return parts[0] + "." + parts[1]
}

func findVersionMetadata(workDir string) (versionMetadata, bool, error) {
	for _, name := range []string{"go.work", "go.mod"} {
		path, err := versionutil.FindNearestFile(workDir, name)
		if errors.Is(err, versionutil.ErrNotFound) {
			continue
		}
		if err != nil {
			return versionMetadata{}, false, err
		}
		metadata, err := parseVersionMetadata(path)
		if err != nil {
			return versionMetadata{}, false, err
		}
		if metadata.goDirective != "" || metadata.toolchain != "" {
			if metadata.goDirective == "" {
				metadata.goDirective = metadata.toolchain
			}
			return metadata, true, nil
		}
		if name == "go.mod" {
			return versionMetadata{}, false, fmt.Errorf("go directive not found in %s", path)
		}
	}
	return versionMetadata{}, false, nil
}

func parseVersionMetadata(path string) (versionMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return versionMetadata{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	metadata := versionMetadata{path: path}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "toolchain" {
			return versionMetadata{}, fmt.Errorf("toolchain directive is empty in %s", path)
		}
		if toolchain, ok := strings.CutPrefix(line, "toolchain "); ok {
			toolchain = strings.TrimPrefix(strings.TrimSpace(toolchain), "go")
			if toolchain == "" {
				return versionMetadata{}, fmt.Errorf("toolchain directive is empty in %s", path)
			}
			metadata.toolchain = toolchain
			continue
		}
		if version, ok := strings.CutPrefix(line, "go "); ok {
			version = strings.TrimSpace(version)
			if version == "" {
				return versionMetadata{}, fmt.Errorf("go directive is empty in %s", path)
			}
			metadata.goDirective = version
		}
	}
	if err := scanner.Err(); err != nil {
		return versionMetadata{}, fmt.Errorf("scan %s: %w", path, err)
	}
	return metadata, nil
}
