package resolve

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/versionutil"
)

// GlobalVersionSource returns the globally selected Go version.
type GlobalVersionSource interface {
	GlobalGoVersion(ctx context.Context) (string, bool, error)
}

// Resolver locates Go toolchain metadata from workspace files.
type Resolver struct {
	global GlobalVersionSource
}

// New constructs a Resolver.
func New(global GlobalVersionSource) *Resolver {
	return &Resolver{global: global}
}

// Current resolves the currently selected Go version from the workspace.
func (r *Resolver) Current(ctx context.Context, workDir string) (app.Selection, error) {
	goWorkPath, err := versionutil.FindNearestFile(workDir, "go.work")
	if err != nil && !errors.Is(err, versionutil.ErrNotFound) {
		return app.Selection{}, err
	}
	if err == nil {
		goVersion, found, parseErr := parseVersionMetadata(goWorkPath)
		if parseErr != nil {
			return app.Selection{}, parseErr
		}
		if found {
			return app.Selection{
				GoVersion: goVersion,
				GoSource:  goWorkPath,
			}, nil
		}
	}

	goModPath, err := versionutil.FindNearestFile(workDir, "go.mod")
	if err != nil {
		if errors.Is(err, versionutil.ErrNotFound) {
			if r.global != nil {
				goVersion, found, globalErr := r.global.GlobalGoVersion(ctx)
				if globalErr != nil {
					return app.Selection{}, globalErr
				}
				if found {
					return app.Selection{
						GoVersion: goVersion,
						GoSource:  "global",
					}, nil
				}
			}
		}
		return app.Selection{}, err
	}

	goVersion, found, err := parseVersionMetadata(goModPath)
	if err != nil {
		return app.Selection{}, err
	}
	if !found {
		return app.Selection{}, fmt.Errorf("go directive not found in %s", goModPath)
	}

	return app.Selection{
		GoVersion: goVersion,
		GoSource:  goModPath,
	}, nil
}

func parseVersionMetadata(path string) (string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var goDirective string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "toolchain" {
			return "", false, fmt.Errorf("toolchain directive is empty in %s", path)
		}
		if toolchain, ok := strings.CutPrefix(line, "toolchain "); ok {
			toolchain = strings.TrimSpace(toolchain)
			toolchain = strings.TrimPrefix(toolchain, "go")
			if toolchain == "" {
				return "", false, fmt.Errorf("toolchain directive is empty in %s", path)
			}
			return toolchain, true, nil
		}
		if version, ok := strings.CutPrefix(line, "go "); ok {
			version = strings.TrimSpace(version)
			if version == "" {
				return "", false, fmt.Errorf("go directive is empty in %s", path)
			}
			goDirective = version
		}
	}

	if err := scanner.Err(); err != nil {
		return "", false, fmt.Errorf("scan %s: %w", path, err)
	}

	if goDirective != "" {
		return goDirective, true, nil
	}

	return "", false, nil
}
