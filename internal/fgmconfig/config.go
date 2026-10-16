package fgmconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

const fileName = ".fgm.toml"

// LatestPatchPolicy selects the newest released patch of the go directive's minor version.
const LatestPatchPolicy = "latest-patch"

var goVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+(\.[0-9]+)?((rc|beta)[0-9]+)?$`)

// GoPolicy is the repository Go version policy from .fgm.toml.
type GoPolicy struct {
	// Version is an explicit Go version without the "go" prefix; empty for latest-patch.
	Version     string
	LatestPatch bool
	Path        string
}

// ParseGoPolicyValue validates a toolchain.go value and returns the normalized form
// ("latest-patch" or a version without the "go" prefix).
func ParseGoPolicyValue(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == LatestPatchPolicy {
		return trimmed, nil
	}
	version := strings.TrimPrefix(trimmed, "go")
	if !goVersionPattern.MatchString(version) {
		return "", fmt.Errorf("invalid Go version policy %q (want a version such as 1.26.8 or %q)", value, LatestPatchPolicy)
	}
	return version, nil
}

// ResolveGoPolicy returns the Go version policy from the nearest .fgm.toml.
// Returns (GoPolicy{}, false, nil) when no config exists or the go key is empty.
func ResolveGoPolicy(workDir string) (GoPolicy, bool, error) {
	config, found, err := LoadNearest(workDir)
	if err != nil || !found {
		return GoPolicy{}, false, err
	}
	if strings.TrimSpace(config.File.Toolchain.Go) == "" {
		return GoPolicy{}, false, nil
	}
	value, err := ParseGoPolicyValue(config.File.Toolchain.Go)
	if err != nil {
		return GoPolicy{}, false, fmt.Errorf("%s: %w", config.Path, err)
	}
	if value == LatestPatchPolicy {
		return GoPolicy{LatestPatch: true, Path: config.Path}, true, nil
	}
	return GoPolicy{Version: value, Path: config.Path}, true, nil
}

// ResolvePinnedLint returns the pinned golangci-lint version from the nearest .fgm.toml.
// Returns ("", false, nil) when no config exists, the field is empty, or set to "auto".
func ResolvePinnedLint(workDir string) (string, bool, error) {
	config, found, err := LoadNearest(workDir)
	if err != nil || !found {
		return "", false, err
	}

	version := strings.TrimSpace(config.File.Toolchain.GolangCILint)
	if version == "" || version == "auto" {
		return "", false, nil
	}

	return version, true, nil
}

// PinnedLintPath returns the path of the .fgm.toml that pins golangci-lint, if any.
func PinnedLintPath(workDir string) (string, bool, error) {
	config, found, err := LoadNearest(workDir)
	if err != nil || !found {
		return "", false, err
	}
	version := strings.TrimSpace(config.File.Toolchain.GolangCILint)
	if version == "" || version == "auto" {
		return "", false, nil
	}
	return config.Path, true, nil
}

// LintDetectionEnabled reports whether lint pins may be detected from other repository files.
func LintDetectionEnabled(workDir string) (bool, error) {
	config, found, err := LoadNearest(workDir)
	if err != nil || !found {
		return err == nil, err
	}
	detect := config.File.Toolchain.GolangCILintDetect
	return detect == nil || *detect, nil
}

// ToolchainConfig describes FGM-specific toolchain overrides.
type ToolchainConfig struct {
	Go           string `toml:"go,omitempty"`
	GolangCILint string `toml:"golangci_lint,omitempty"`
	// GolangCILintDetect disables reading lint pins from other repository files when false.
	GolangCILintDetect *bool `toml:"golangci_lint_detect,omitempty"`
}

// File describes the supported FGM repo config.
type File struct {
	Toolchain ToolchainConfig `toml:"toolchain"`
	Check     CheckConfig     `toml:"check,omitempty"`
}

// CheckConfig configures `fgm check`.
type CheckConfig struct {
	// Skip lists check IDs (for example "C5") that fgm check does not run.
	Skip []string `toml:"skip,omitempty"`
}

// Result wraps a loaded repo config and its source path.
type Result struct {
	Path string
	File File
}

// SaveNearest writes .fgm.toml in the nearest config or Go workspace root.
func SaveNearest(workDir string, file File) (string, error) {
	path, err := resolveWritePath(workDir)
	if err != nil {
		return "", err
	}

	content, err := toml.Marshal(file)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}

	return path, nil
}

// UpdateNearest applies update to the .fgm.toml that SaveNearest would write and saves it,
// preserving keys that update does not change.
func UpdateNearest(workDir string, update func(*File)) (string, error) {
	path, err := resolveWritePath(workDir)
	if err != nil {
		return "", err
	}
	var file File
	if content, err := os.ReadFile(path); err == nil {
		if err := toml.Unmarshal(content, &file); err != nil {
			return "", fmt.Errorf("parse %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	update(&file)
	return SaveNearest(workDir, file)
}

// LoadNearest loads the nearest .fgm.toml by walking upward from workDir.
func LoadNearest(workDir string) (Result, bool, error) {
	path, found, err := findNearest(workDir)
	if err != nil || !found {
		return Result{}, found, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Result{}, false, fmt.Errorf("read %s: %w", path, err)
	}

	var file File
	if err := toml.Unmarshal(content, &file); err != nil {
		return Result{}, false, fmt.Errorf("parse %s: %w", path, err)
	}

	return Result{
		Path: path,
		File: file,
	}, true, nil
}

func findNearest(workDir string) (string, bool, error) {
	current := workDir
	for {
		candidate := filepath.Join(current, fileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true, nil
		} else if !os.IsNotExist(err) {
			return "", false, fmt.Errorf("stat %s: %w", candidate, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}

func resolveWritePath(workDir string) (string, error) {
	if path, found, err := findNearest(workDir); err != nil {
		return "", err
	} else if found {
		return path, nil
	}

	for _, name := range []string{"go.work", "go.mod"} {
		if path, found, err := findNearestNamedFile(workDir, name); err != nil {
			return "", err
		} else if found {
			return filepath.Join(filepath.Dir(path), fileName), nil
		}
	}

	return filepath.Join(workDir, fileName), nil
}

func findNearestNamedFile(workDir string, name string) (string, bool, error) {
	current := workDir
	for {
		candidate := filepath.Join(current, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true, nil
		} else if !os.IsNotExist(err) {
			return "", false, fmt.Errorf("stat %s: %w", candidate, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}
