// Package goenv controls the Go toolchain variables that fgm passes to child processes.
package goenv

import (
	"fmt"
	"strings"
)

// Mode selects how GOTOOLCHAIN and GOROOT reach a child process.
type Mode string

const (
	// ModeLocal forces GOTOOLCHAIN=local and drops GOROOT so the selected go binary runs itself.
	ModeLocal Mode = "local"
	// ModeInherit passes the caller's GOTOOLCHAIN and GOROOT through unchanged.
	ModeInherit Mode = "inherit"
	// ModeEnvVar names the environment variable that selects the mode when no flag is given.
	ModeEnvVar = "FGM_EXEC_GOTOOLCHAIN"
)

// ParseMode validates a mode value; an empty value selects ModeLocal.
func ParseMode(value string) (Mode, error) {
	switch Mode(strings.TrimSpace(value)) {
	case "", ModeLocal:
		return ModeLocal, nil
	case ModeInherit:
		return ModeInherit, nil
	default:
		return "", fmt.Errorf("invalid GOTOOLCHAIN mode %q (want %q or %q)", value, ModeLocal, ModeInherit)
	}
}

// Apply returns the child environment for mode and an optional one-line warning.
//
// The go binary on PATH already is the selected toolchain, so GOTOOLCHAIN=local is the only
// value that keeps "what fgm current reports" equal to "what runs"; GOROOT is dropped because
// the go binary derives it from its own location when unset.
func Apply(env []string, mode Mode) ([]string, string) {
	if mode == ModeInherit {
		value, _ := lookup(env, "GOTOOLCHAIN")
		switch value {
		case "", "local", "auto":
			return env, ""
		}
		return env, fmt.Sprintf("fgm: warning: inherited GOTOOLCHAIN=%s may run a Go toolchain other than the selected one", value)
	}

	result := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if strings.HasPrefix(entry, "GOTOOLCHAIN=") || strings.HasPrefix(entry, "GOROOT=") {
			continue
		}
		result = append(result, entry)
	}
	return append(result, "GOTOOLCHAIN=local"), ""
}

// OverridesSelection reports whether a GOTOOLCHAIN value names a specific toolchain,
// which makes a plain go call run something other than the fgm selection.
func OverridesSelection(value string) bool {
	switch strings.TrimSpace(value) {
	case "", "local", "auto", "path", "local+auto", "local+path":
		return false
	}
	return true
}

func lookup(env []string, key string) (string, bool) {
	prefix := key + "="
	for _, entry := range env {
		if value, ok := strings.CutPrefix(entry, prefix); ok {
			return value, true
		}
	}
	return "", false
}
