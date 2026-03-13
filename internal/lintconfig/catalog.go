package lintconfig

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//go:embed linters.json
var lintersJSON []byte

// Linter describes a known golangci-lint linter or formatter.
type Linter struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	IsFormatter bool   `json:"is_formatter"`
	MinGoMinor  int    `json:"min_go_minor,omitempty"`
}

// Conflict describes two linters that should not be enabled together.
type Conflict struct {
	A       string `json:"a"`
	B       string `json:"b"`
	Message string `json:"message"`
}

// Catalog holds the embedded linter metadata.
type Catalog struct {
	Linters   []Linter   `json:"linters"`
	Conflicts []Conflict `json:"conflicts"`
	byName    map[string]Linter
}

// LoadCatalog parses the embedded linters.json.
func LoadCatalog() (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(lintersJSON, &c); err != nil {
		return nil, fmt.Errorf("parse linters.json: %w", err)
	}
	c.byName = make(map[string]Linter, len(c.Linters))
	for _, l := range c.Linters {
		c.byName[l.Name] = l
	}
	return &c, nil
}

// Lookup returns a linter by name and whether it exists.
func (c *Catalog) Lookup(name string) (Linter, bool) {
	l, ok := c.byName[name]
	return l, ok
}

// Available returns whether the linter is usable with the given Go minor version.
func (l Linter) Available(goMinor int) bool {
	return l.MinGoMinor == 0 || goMinor >= l.MinGoMinor
}

// ConflictsWith returns the conflict message if linters a and b conflict, or empty string.
func (c *Catalog) ConflictsWith(a, b string) string {
	for _, conf := range c.Conflicts {
		if (conf.A == a && conf.B == b) || (conf.A == b && conf.B == a) {
			return conf.Message
		}
	}
	return ""
}

// ParseGoMinor extracts the minor version number from a Go version string.
// e.g. "1.22.5" → 22, "1.21" → 21.
func ParseGoMinor(goVersion string) (int, error) {
	parts := strings.SplitN(strings.TrimPrefix(goVersion, "go"), ".", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid go version: %s", goVersion)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid go minor version: %s", goVersion)
	}
	return minor, nil
}
