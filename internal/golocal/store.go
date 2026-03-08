package golocal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const globalGoVersionStateFile = "global-go-version"

// Store manages Go toolchains known to FGM and the active global Go version.
type Store struct {
	root    string
	pathEnv string
}

// New constructs a Store using an FGM data root and PATH value.
func New(root string, pathEnv string) *Store {
	return &Store{root: root, pathEnv: pathEnv}
}

// DefaultRoot returns the default FGM data directory.
func DefaultRoot() (string, error) {
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "fgm"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, ".local", "share", "fgm"), nil
}

// ListLocalGoVersions returns Go versions from the FGM-managed store plus the current system Go on PATH.
func (s *Store) ListLocalGoVersions(ctx context.Context) ([]string, error) {
	versions := make(map[string]struct{})

	managedVersions, err := s.listManagedGoVersions()
	if err != nil {
		return nil, err
	}
	for _, version := range managedVersions {
		versions[version] = struct{}{}
	}

	systemVersion, _, err := s.systemGo(ctx)
	if err != nil {
		return nil, err
	}
	if systemVersion != "" {
		versions[systemVersion] = struct{}{}
	}

	return sortVersions(versions), nil
}

// HasGoVersion reports whether the requested version is available either in the FGM-managed store or as the current system Go.
func (s *Store) HasGoVersion(ctx context.Context, version string) (bool, error) {
	if _, err := os.Stat(s.managedGoBinaryPath(version)); err == nil {
		return true, nil
	}

	systemVersion, _, err := s.systemGo(ctx)
	if err != nil {
		return false, err
	}

	return systemVersion == version, nil
}

// GlobalGoVersion returns the globally selected Go version, if set.
func (s *Store) GlobalGoVersion(ctx context.Context) (string, bool, error) {
	_ = ctx

	content, err := os.ReadFile(s.globalStatePath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("read global Go version: %w", err)
	}

	version := strings.TrimSpace(string(content))
	if version == "" {
		return "", false, nil
	}

	return version, true, nil
}

// SetGlobalGoVersion persists the globally selected Go version.
func (s *Store) SetGlobalGoVersion(ctx context.Context, version string) error {
	_ = ctx

	if err := os.MkdirAll(filepath.Dir(s.globalStatePath()), 0o755); err != nil {
		return fmt.Errorf("create FGM state directory: %w", err)
	}
	if err := os.WriteFile(s.globalStatePath(), []byte(version+"\n"), 0o644); err != nil {
		return fmt.Errorf("write global Go version: %w", err)
	}
	return nil
}

// DeleteGoVersion removes an FGM-managed Go version from the local store.
func (s *Store) DeleteGoVersion(ctx context.Context, version string) (string, error) {
	globalVersion, ok, err := s.GlobalGoVersion(ctx)
	if err != nil {
		return "", err
	}
	if ok && globalVersion == version {
		return "", fmt.Errorf("go version %s is the current global version; switch away before removing it", version)
	}

	installDir := filepath.Join(s.root, "go", version)
	info, err := os.Lstat(installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("go version %s is not managed by FGM", version)
		}
		return "", fmt.Errorf("stat managed Go version: %w", err)
	}

	if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		return "", fmt.Errorf("managed Go path for %s is invalid", version)
	}
	if err := os.RemoveAll(installDir); err != nil {
		return "", fmt.Errorf("remove managed Go version: %w", err)
	}

	return installDir, nil
}

// GoBinaryPath returns the executable path for the requested Go version.
func (s *Store) GoBinaryPath(ctx context.Context, version string) (string, error) {
	managedBinary := s.managedGoBinaryPath(version)
	if _, err := os.Stat(managedBinary); err == nil {
		return managedBinary, nil
	}

	systemVersion, systemBinary, err := s.systemGo(ctx)
	if err != nil {
		return "", err
	}
	if systemVersion == version && systemBinary != "" {
		return systemBinary, nil
	}

	return "", fmt.Errorf("go version %s is not installed", version)
}

// EnsureShims writes the shell shim scripts used to dispatch go through FGM.
func (s *Store) EnsureShims() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("shim generation is not implemented on windows")
	}

	if err := os.MkdirAll(s.ShimDir(), 0o755); err != nil {
		return fmt.Errorf("create shim directory: %w", err)
	}

	shimContent := "#!/bin/sh\nexec fgm __shim go \"$@\"\n"
	if err := os.WriteFile(filepath.Join(s.ShimDir(), "go"), []byte(shimContent), 0o755); err != nil {
		return fmt.Errorf("write go shim: %w", err)
	}

	return nil
}

// ShimDir returns the directory that should be added to PATH for FGM shims.
func (s *Store) ShimDir() string {
	return filepath.Join(s.root, "shims")
}

func (s *Store) listManagedGoVersions() ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.root, "go"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read managed Go versions: %w", err)
	}

	versions := make(map[string]struct{})
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		version := entry.Name()
		if _, err := os.Stat(s.managedGoBinaryPath(version)); err == nil {
			versions[version] = struct{}{}
		}
	}

	return sortVersions(versions), nil
}

func (s *Store) systemGo(ctx context.Context) (string, string, error) {
	goBinary, ok := findGoBinary(s.pathEnv)
	if !ok {
		return "", "", nil
	}

	cmd := exec.CommandContext(ctx, goBinary, "version")
	if s.pathEnv != "" {
		cmd.Env = append(cmd.Environ(), "PATH="+s.pathEnv)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("run go version: %w", err)
	}

	version, ok := parseGoVersionOutput(string(output))
	if !ok {
		return "", "", fmt.Errorf("parse go version output: %q", strings.TrimSpace(string(output)))
	}

	return version, goBinary, nil
}

func (s *Store) globalStatePath() string {
	return filepath.Join(s.root, "state", globalGoVersionStateFile)
}

func (s *Store) managedGoBinaryPath(version string) string {
	binaryName := "go"
	if runtime.GOOS == "windows" {
		binaryName = "go.exe"
	}
	return filepath.Join(s.root, "go", version, "bin", binaryName)
}

// RegisterExistingGoInstallation registers an existing GOROOT into the FGM-managed store.
func (s *Store) RegisterExistingGoInstallation(version string, goroot string) (string, error) {
	installDir := filepath.Join(s.root, "go", version)
	if _, err := os.Stat(installDir); err == nil {
		return installDir, nil
	}

	if err := os.MkdirAll(filepath.Dir(installDir), 0o755); err != nil {
		return "", fmt.Errorf("create managed Go root: %w", err)
	}
	if err := os.Symlink(goroot, installDir); err != nil {
		return "", fmt.Errorf("symlink existing Go installation: %w", err)
	}

	return installDir, nil
}

func findGoBinary(pathEnv string) (string, bool) {
	binaryName := "go"
	if runtime.GOOS == "windows" {
		binaryName = "go.exe"
	}

	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}

		candidate := filepath.Join(dir, binaryName)
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Mode()&0o111 == 0 && runtime.GOOS != "windows" {
			continue
		}
		return candidate, true
	}

	return "", false
}

func parseGoVersionOutput(output string) (string, bool) {
	fields := strings.Fields(output)
	if len(fields) < 3 {
		return "", false
	}
	if fields[0] != "go" || fields[1] != "version" {
		return "", false
	}

	version := strings.TrimPrefix(fields[2], "go")
	if version == "" {
		return "", false
	}

	return version, true
}

func sortVersions(set map[string]struct{}) []string {
	versions := make([]string, 0, len(set))
	for version := range set {
		versions = append(versions, version)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions
}
