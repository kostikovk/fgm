package versionutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrNotFound indicates a file was not found walking upward from a directory.
var ErrNotFound = errors.New("not found")

// FindNearestFile walks upward from dir looking for a file named name.
// Returns ErrNotFound (wrapped) when the file is not found.
func FindNearestFile(dir string, name string) (string, error) {
	current := dir
	for {
		candidate := filepath.Join(current, name)
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("stat %s: %w", candidate, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("%s not found from %s upward: %w", name, dir, ErrNotFound)
}
