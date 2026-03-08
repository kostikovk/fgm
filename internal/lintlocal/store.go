package lintlocal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// Store manages installed golangci-lint versions in the FGM data root.
type Store struct {
	root string
}

// New constructs a Store using an FGM data root.
func New(root string) *Store {
	return &Store{root: root}
}

// ListLocalLintVersions returns installed golangci-lint versions from the FGM-managed store.
func (s *Store) ListLocalLintVersions(ctx context.Context) ([]string, error) {
	_ = ctx

	entries, err := os.ReadDir(filepath.Join(s.root, "golangci-lint"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read managed golangci-lint versions: %w", err)
	}

	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		binaryPath := filepath.Join(s.root, "golangci-lint", entry.Name(), lintBinaryName())
		if _, err := os.Stat(binaryPath); err == nil {
			versions = append(versions, entry.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions, nil
}

// DeleteLintVersion removes an FGM-managed golangci-lint version from the local store.
func (s *Store) DeleteLintVersion(ctx context.Context, version string) (string, error) {
	_ = ctx

	installDir := filepath.Join(s.root, "golangci-lint", version)
	info, err := os.Stat(installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("golangci-lint version %s is not managed by FGM", version)
		}
		return "", fmt.Errorf("stat managed golangci-lint version: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("managed golangci-lint path for %s is invalid", version)
	}
	if err := os.RemoveAll(installDir); err != nil {
		return "", fmt.Errorf("remove managed golangci-lint version: %w", err)
	}

	return installDir, nil
}

// LintBinaryPath returns the executable path for the requested golangci-lint version.
func (s *Store) LintBinaryPath(ctx context.Context, version string) (string, error) {
	_ = ctx

	binaryPath := filepath.Join(s.root, "golangci-lint", version, lintBinaryName())
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	return "", fmt.Errorf("golangci-lint version %s is not installed", version)
}

// RegisterExistingLintInstallation registers an existing golangci-lint binary into the FGM-managed store.
func (s *Store) RegisterExistingLintInstallation(version string, binaryPath string) (string, error) {
	installDir := filepath.Join(s.root, "golangci-lint", version)
	if _, err := os.Stat(filepath.Join(installDir, lintBinaryName())); err == nil {
		return installDir, nil
	}

	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create managed golangci-lint root: %w", err)
	}
	targetPath := filepath.Join(installDir, lintBinaryName())
	if err := os.Symlink(binaryPath, targetPath); err != nil {
		return "", fmt.Errorf("symlink existing golangci-lint installation: %w", err)
	}

	return installDir, nil
}

func lintBinaryName() string {
	if runtime.GOOS == "windows" {
		return "golangci-lint.exe"
	}
	return "golangci-lint"
}
