package golangcilint

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type goPatchManifest struct {
	LatestGoVersions map[string]string `json:"latest_go_versions"`
}

// LatestGoPatch returns the newest Go patch release recorded in the embedded catalog for
// goVersion's minor, without network I/O.
func (p *Provider) LatestGoPatch(goVersion string) (string, bool, error) {
	minor, err := parseGoMinor(goVersion)
	if err != nil {
		return "", false, err
	}
	var manifest goPatchManifest
	if err := json.Unmarshal(p.compatibilityData, &manifest); err != nil {
		return "", false, fmt.Errorf("decode golangci-lint compatibility manifest: %w", err)
	}
	version, ok := manifest.LatestGoVersions[strconv.Itoa(minor)]
	return version, ok, nil
}

// LintCompatibility reports whether lintVersion can run on goVersion according to the embedded
// catalog. known is false when the catalog has no entry for lintVersion.
func (p *Provider) LintCompatibility(goVersion string, lintVersion string) (compatible bool, known bool, err error) {
	minor, err := parseGoMinor(goVersion)
	if err != nil {
		return false, false, err
	}
	manifest, err := p.loadCompatibilityManifest()
	if err != nil {
		return false, false, err
	}
	entry, ok := manifest.Versions["v"+strings.TrimPrefix(strings.TrimSpace(lintVersion), "v")]
	if !ok {
		return false, false, nil
	}
	return minor >= entry.MinGoMinor && minor <= entry.MaxGoMinor, true, nil
}
