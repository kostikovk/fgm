package lintconfig

import (
	"fmt"
	"strings"
)

// PresetMinimal is the name for the minimal linter preset.
const PresetMinimal = "minimal"

// PresetStandard is the name for the standard (default) linter preset.
const PresetStandard = "standard"

// PresetStrict is the name for the strict linter preset.
const PresetStrict = "strict"

// ValidPresets lists accepted preset names.
var ValidPresets = []string{PresetMinimal, PresetStandard, PresetStrict}

var validPresetSet = map[string]struct{}{
	PresetMinimal:  {},
	PresetStandard: {},
	PresetStrict:   {},
}

// presetLinters maps preset names to ordered lists of linter names.
// Each preset includes all linters from the previous tier.
var presetLinters = map[string][]string{
	PresetMinimal: {
		"govet",
		"errcheck",
		"staticcheck",
		"unused",
		"gosimple",
		"ineffassign",
		"typecheck",
	},
	PresetStandard: {
		// minimal
		"govet",
		"errcheck",
		"staticcheck",
		"unused",
		"gosimple",
		"ineffassign",
		"typecheck",
		// standard additions
		"gocritic",
		"revive",
		"misspell",
		"nolintlint",
		"exhaustive",
		"unconvert",
		"unparam",
		"bodyclose",
		"errname",
		"errorlint",
		"copyloopvar",
		"intrange",
	},
	PresetStrict: {
		// minimal
		"govet",
		"errcheck",
		"staticcheck",
		"unused",
		"gosimple",
		"ineffassign",
		"typecheck",
		// standard additions
		"gocritic",
		"revive",
		"misspell",
		"nolintlint",
		"exhaustive",
		"unconvert",
		"unparam",
		"bodyclose",
		"errname",
		"errorlint",
		"copyloopvar",
		"intrange",
		// strict additions
		"wrapcheck",
		"forcetypeassert",
		"goconst",
		"cyclop",
		"funlen",
		"nestif",
		"gocognit",
		"dupl",
		"perfsprint",
		"prealloc",
		"sloglint",
	},
}

// NormalizePreset returns the effective preset name or an error for invalid input.
func NormalizePreset(preset string) (string, error) {
	if preset == "" {
		return PresetStandard, nil
	}
	if _, ok := validPresetSet[preset]; !ok {
		return "", fmt.Errorf("invalid preset %q (must be one of: %s)", preset, strings.Join(ValidPresets, ", "))
	}
	return preset, nil
}

// PresetLinters returns the linter names for the given preset,
// filtered by Go minor version compatibility using the catalog.
func PresetLinters(catalog *Catalog, preset string, goMinor int) []string {
	names := presetLinters[preset]
	var result []string
	for _, name := range names {
		l, exists := catalog.Lookup(name)
		if !exists {
			continue
		}
		if l.Available(goMinor) {
			result = append(result, name)
		}
	}
	return result
}
