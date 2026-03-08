package fgmconfig

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

const fileName = ".fgm.toml"

// ToolchainConfig describes FGM-specific toolchain overrides.
type ToolchainConfig struct {
	GolangCILint string `toml:"golangci_lint"`
}

// File describes the supported FGM repo config.
type File struct {
	Toolchain ToolchainConfig `toml:"toolchain"`
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
