// Package workspace locates repository and Go module boundaries on disk.
package workspace

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// RepoRoot returns the nearest ancestor of dir (inclusive) that contains a .git entry.
func RepoRoot(dir string) (string, bool) {
	current := filepath.Clean(dir)
	for {
		if _, err := os.Lstat(filepath.Join(current, ".git")); err == nil {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

// ModuleRoot returns the directory of the nearest go.mod at or above dir.
func ModuleRoot(dir string) (string, bool) {
	return nearestDirWith(dir, "go.mod")
}

// WorkRoot returns the directory of the nearest go.work at or above dir.
func WorkRoot(dir string) (string, bool) {
	return nearestDirWith(dir, "go.work")
}

// Boundary returns the outermost directory that repository-level files apply to for dir:
// the repository root, else the go.work root, else the module root, else dir itself.
func Boundary(dir string) string {
	if root, ok := RepoRoot(dir); ok {
		return root
	}
	if root, ok := WorkRoot(dir); ok {
		return root
	}
	if root, ok := ModuleRoot(dir); ok {
		return root
	}
	return filepath.Clean(dir)
}

// Ancestors returns from, its parents, and so on up to and including to.
// When to is not an ancestor of from, only from is returned.
func Ancestors(from string, to string) []string {
	from, to = filepath.Clean(from), filepath.Clean(to)
	dirs := []string{from}
	if !Contains(to, from) {
		return dirs
	}
	for current := from; current != to; {
		current = filepath.Dir(current)
		dirs = append(dirs, current)
	}
	return dirs
}

// Contains reports whether path is root or lies below it.
func Contains(root string, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// ListModuleDirs returns, in lexical order, every directory below root (inclusive) holding a
// go.mod, skipping vendor, testdata, node_modules and dot-directories. Symlinks are not followed.
func ListModuleDirs(root string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrPermission) && path != root {
				return filepath.SkipDir
			}
			return err
		}
		if entry.IsDir() {
			if path != root && skipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() == "go.mod" && entry.Type().IsRegular() {
			dirs = append(dirs, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(dirs)
	return dirs, nil
}

func skipDir(name string) bool {
	switch name {
	case "vendor", "testdata", "node_modules":
		return true
	}
	return strings.HasPrefix(name, ".")
}

func nearestDirWith(dir string, name string) (string, bool) {
	current := filepath.Clean(dir)
	for {
		if info, err := os.Stat(filepath.Join(current, name)); err == nil && !info.IsDir() {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}
