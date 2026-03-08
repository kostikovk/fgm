package lintlocal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

		binaryPath := filepath.Join(s.root, "golangci-lint", entry.Name(), "golangci-lint")
		if _, err := os.Stat(binaryPath); err == nil {
			versions = append(versions, entry.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions, nil
}
