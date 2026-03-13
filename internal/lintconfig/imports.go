package lintconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// FindModulePath reads the go.mod in the given directory (or walks up)
// and returns the module path, or empty string if not found.
func FindModulePath(workDir string) string {
	dir := workDir
	for {
		modPath := filepath.Join(dir, "go.mod")
		if f, err := os.Open(modPath); err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "module ") {
					return strings.TrimSpace(strings.TrimPrefix(line, "module"))
				}
			}
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
