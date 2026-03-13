package lintconfig

import (
	"slices"
	"testing"
)

func TestPresetLinters_Standard(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	linters := PresetLinters(catalog, PresetStandard, 26)
	if len(linters) == 0 {
		t.Fatal("expected linters for standard preset, got none")
	}

	// govet must be in every preset.
	found := slices.Contains(linters, "govet")
	if !found {
		t.Fatal("govet not found in standard preset")
	}
}

func TestPresetLinters_FiltersGoVersion(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	// copyloopvar requires Go 1.22+.
	linters21 := PresetLinters(catalog, PresetStandard, 21)
	for _, l := range linters21 {
		if l == "copyloopvar" {
			t.Fatal("copyloopvar should be excluded for Go 1.21")
		}
	}

	linters22 := PresetLinters(catalog, PresetStandard, 22)
	found := slices.Contains(linters22, "copyloopvar")
	if !found {
		t.Fatal("copyloopvar should be included for Go 1.22")
	}
}

func TestNormalizePreset_DefaultsToStandard(t *testing.T) {
	t.Parallel()

	got, err := NormalizePreset("")
	if err != nil {
		t.Fatalf("NormalizePreset(\"\") error = %v", err)
	}
	if got != PresetStandard {
		t.Fatalf("NormalizePreset(\"\") = %q, want %q", got, PresetStandard)
	}
}

func TestNormalizePreset_RejectsInvalidPreset(t *testing.T) {
	t.Parallel()

	_, err := NormalizePreset("nonexistent")
	if err == nil {
		t.Fatal("expected invalid preset to return an error")
	}
}

func TestPresetLinters_MinimalIsSubsetOfStandard(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	minimal := PresetLinters(catalog, PresetMinimal, 26)
	standard := PresetLinters(catalog, PresetStandard, 26)

	standardSet := make(map[string]bool)
	for _, l := range standard {
		standardSet[l] = true
	}

	for _, l := range minimal {
		if !standardSet[l] {
			t.Fatalf("minimal linter %q not in standard preset", l)
		}
	}

	if len(minimal) >= len(standard) {
		t.Fatal("minimal should have fewer linters than standard")
	}
}

func TestPresetLinters_StrictHasMoreThanStandard(t *testing.T) {
	t.Parallel()

	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	standard := PresetLinters(catalog, PresetStandard, 26)
	strict := PresetLinters(catalog, PresetStrict, 26)

	if len(strict) <= len(standard) {
		t.Fatalf("strict (%d) should have more linters than standard (%d)", len(strict), len(standard))
	}
}
