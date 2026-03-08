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
