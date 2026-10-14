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
		if modulePath, ok := readModulePathFile(modPath); ok {
			return modulePath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func readModulePathFile(modPath string) (string, bool) {
	f, err := os.Open(modPath)
	if err != nil {
		return "", false
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), true
		}
	}
	return "", true
}
