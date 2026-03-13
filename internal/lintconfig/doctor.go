package lintconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Finding describes a single lint config diagnostic.
type Finding struct {
	Severity string // "ERROR", "WARN", "INFO"
	Message  string
}

// String formats a finding for display.
func (f Finding) String() string {
	return f.Severity + " " + f.Message
}

// rawConfig is a loose representation of .golangci.yml for doctor parsing.
type rawConfig struct {
	Version    string        `yaml:"version"`
	Linters    rawLinters    `yaml:"linters"`
	Formatters rawFormatters `yaml:"formatters"`
}

type rawLinters struct {
	Enable  []string `yaml:"enable"`
	Disable []string `yaml:"disable"`
}

type rawFormatters struct {
	Enable []string `yaml:"enable"`
}

// Diagnose checks an existing .golangci.yml for issues.
func Diagnose(catalog *Catalog, workDir string, goVersion string) ([]Finding, error) {
	var findings []Finding

	configPath := findConfigFile(workDir)
	if configPath == "" {
		findings = append(findings, Finding{
			Severity: "WARN",
			Message:  "no .golangci.yml found; run 'fgm lint init' to generate one",
		})
		return findings, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg rawConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		findings = append(findings, Finding{
			Severity: "ERROR",
			Message:  fmt.Sprintf("failed to parse %s: %s", filepath.Base(configPath), err),
		})
		return findings, nil
	}

	// Check v2 format marker.
	if cfg.Version != "2" {
		findings = append(findings, Finding{
			Severity: "ERROR",
			Message:  "config is missing 'version: \"2\"'; golangci-lint v2 requires this field",
		})
	}

	goMinor := 0
	if goVersion != "" {
		if m, err := ParseGoMinor(goVersion); err == nil {
			goMinor = m
		}
	}

	// Collect all enabled linters and formatters.
	allEnabled := make(map[string]bool)
	for _, name := range cfg.Linters.Enable {
		allEnabled[name] = true
	}
	for _, name := range cfg.Formatters.Enable {
		allEnabled[name] = true
	}

	// Check each enabled linter.
	for name := range allEnabled {
		l, exists := catalog.Lookup(name)
		if !exists {
			findings = append(findings, Finding{
				Severity: "WARN",
				Message:  fmt.Sprintf("unknown linter %q; check for typos", name),
			})
			continue
		}
		if goMinor > 0 && !l.Available(goMinor) {
			findings = append(findings, Finding{
				Severity: "WARN",
				Message:  fmt.Sprintf("linter %q requires Go 1.%d+, project uses Go %s", name, l.MinGoMinor, goVersion),
			})
		}
	}

	// Check for conflicts.
	enabledList := make([]string, 0, len(allEnabled))
	for name := range allEnabled {
		enabledList = append(enabledList, name)
	}
	for i := 0; i < len(enabledList); i++ {
		for j := i + 1; j < len(enabledList); j++ {
			if msg := catalog.ConflictsWith(enabledList[i], enabledList[j]); msg != "" {
				findings = append(findings, Finding{
					Severity: "WARN",
					Message:  msg,
				})
			}
		}
	}

	// Suggest missing standard linters.
	if goMinor > 0 {
		standardLinters := PresetLinters(catalog, PresetStandard, goMinor)
		for _, name := range standardLinters {
			if !allEnabled[name] {
				findings = append(findings, Finding{
					Severity: "INFO",
					Message:  fmt.Sprintf("consider enabling %q (included in standard preset)", name),
				})
			}
		}
	}

	if len(findings) == 0 {
		findings = append(findings, Finding{
			Severity: "OK",
			Message:  fmt.Sprintf("golangci-lint config looks good (%s)", filepath.Base(configPath)),
		})
	}

	return findings, nil
}

func findConfigFile(workDir string) string {
	names := []string{".golangci.yml", ".golangci.yaml"}
	dir := workDir
	for {
		for _, name := range names {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
