package doctor

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoStore provides the local state doctor needs to inspect.
type GoStore interface {
	GlobalGoVersion(ctx context.Context) (string, bool, error)
	ShimDir() string
}

// Service reports diagnostics for FGM environment setup.
type Service struct {
	goStore GoStore
	pathEnv string
}

// New constructs a doctor Service.
func New(goStore GoStore, pathEnv string) *Service {
	return &Service{
		goStore: goStore,
		pathEnv: pathEnv,
	}
}

// Diagnose returns human-readable diagnostics for the current FGM setup.
func (s *Service) Diagnose(ctx context.Context) ([]string, error) {
	lines := make([]string, 0, 4)

	version, ok, err := s.goStore.GlobalGoVersion(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		lines = append(lines, "OK global Go version: "+version)
	} else {
		lines = append(lines, "WARN no global Go version is selected")
	}

	shimDir := s.goStore.ShimDir()
	if pathContainsDir(s.pathEnv, shimDir) {
		lines = append(lines, "OK shim dir is on PATH: "+shimDir)
	} else {
		lines = append(lines, "WARN shim dir is not on PATH: "+shimDir)
		lines = append(lines, `Run: eval "$(fgm env)"`)
	}

	if _, err := exec.LookPath("fgm"); err == nil {
		lines = append(lines, "OK fgm is available on PATH")
	} else {
		lines = append(lines, "WARN fgm is not available on PATH")
	}

	return lines, nil
}

func pathContainsDir(pathEnv string, dir string) bool {
	for _, entry := range filepath.SplitList(pathEnv) {
		if strings.TrimSpace(entry) == dir {
			return true
		}
	}
	return false
}
