package golangcilint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const defaultGoReleasesURL = "https://go.dev"

const minimumTrackedGoMinor = 18

var issueTitlePattern = regexp.MustCompile(`^go1\.(\d+)\s+support$`)
var issueSincePattern = regexp.MustCompile(`(?i)since\s+\[?(v\d+\.\d+(?:\.\d+)?)`)

type issueSearchResult struct {
	Items []issueItem `json:"items"`
}

type issueItem struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type goReleaseFeedEntry struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

// Manifest is the generated compatibility catalog stored on disk.
type Manifest struct {
	GeneratedAt       string                        `json:"generated_at,omitempty"`
	LatestGoVersions  map[string]string             `json:"latest_go_versions,omitempty"`
	SupportThresholds map[string]string             `json:"support_thresholds,omitempty"`
	Versions          map[string]compatibilityEntry `json:"versions"`
}

// GeneratorConfig configures the compatibility manifest generator.
type GeneratorConfig struct {
	Client        *http.Client
	BaseURL       string
	GoReleasesURL string
	Now           func() time.Time
	GitHubToken   string
}

// Generator builds compatibility manifests from upstream release and issue data.
type Generator struct {
	client        *http.Client
	baseURL       string
	goReleasesURL string
	now           func() time.Time
	githubToken   string
}

type supportThreshold struct {
	GoMinor      int
	SinceVersion string
}

// NewGenerator constructs a compatibility manifest generator.
func NewGenerator(config GeneratorConfig) *Generator {
	client := config.Client
	if client == nil {
		client = http.DefaultClient
	}
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	goReleasesURL := config.GoReleasesURL
	if goReleasesURL == "" {
		goReleasesURL = defaultGoReleasesURL
	}
	nowFn := config.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	return &Generator{
		client:        client,
		baseURL:       strings.TrimRight(baseURL, "/"),
		goReleasesURL: strings.TrimRight(goReleasesURL, "/"),
		now:           nowFn,
		githubToken:   config.GitHubToken,
	}
}

// Generate builds a compatibility manifest from upstream releases and support issues.
func (g *Generator) Generate(ctx context.Context) (Manifest, error) {
	releases, err := g.fetchLintReleases(ctx)
	if err != nil {
		return Manifest{}, err
	}
	thresholds, err := g.fetchSupportThresholds(ctx)
	if err != nil {
		return Manifest{}, err
	}
	latestGoVersions, err := g.fetchLatestGoVersions(ctx)
	if err != nil {
		return Manifest{}, err
	}

	return Manifest{
		GeneratedAt:       g.now().UTC().Format(time.RFC3339),
		LatestGoVersions:  latestGoVersions,
		SupportThresholds: stringifyThresholds(thresholds),
		Versions:          buildManifestEntries(releases, thresholds),
	}, nil
}

func (g *Generator) fetchLintReleases(ctx context.Context) ([]release, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		g.baseURL+"/repos/golangci/golangci-lint/releases?per_page=100",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build golangci-lint releases request: %w", err)
	}
	if g.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+g.githubToken)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch golangci-lint releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch golangci-lint releases: unexpected status %s", resp.Status)
	}

	var releases []release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decode golangci-lint releases: %w", err)
	}

	return releases, nil
}

func (g *Generator) fetchSupportThresholds(ctx context.Context) ([]supportThreshold, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		g.baseURL+"/search/issues?q=repo:golangci/golangci-lint+is:issue+in:title+%22go1.%22+support&per_page=100",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build golangci-lint issues request: %w", err)
	}
	if g.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+g.githubToken)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch golangci-lint issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch golangci-lint issues: unexpected status %s", resp.Status)
	}

	var searchResult issueSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("decode golangci-lint issues: %w", err)
	}

	thresholds := make([]supportThreshold, 0, len(searchResult.Items))
	for _, item := range searchResult.Items {
		threshold, ok := parseSupportThreshold(item)
		if !ok {
			continue
		}
		thresholds = append(thresholds, threshold)
	}

	slices.SortFunc(thresholds, func(left supportThreshold, right supportThreshold) int {
		if left.GoMinor != right.GoMinor {
			return left.GoMinor - right.GoMinor
		}
		return compareSemver(left.SinceVersion, right.SinceVersion)
	})

	return thresholds, nil
}

func (g *Generator) fetchLatestGoVersions(ctx context.Context) (map[string]string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		g.goReleasesURL+"/dl/?mode=json&include=all",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build Go releases request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch Go releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch Go releases: unexpected status %s", resp.Status)
	}

	var entries []goReleaseFeedEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("decode Go releases: %w", err)
	}

	latest := make(map[string]string)
	for _, entry := range entries {
		if !entry.Stable {
			continue
		}

		minor, ok := goMinorString(entry.Version)
		if !ok {
			continue
		}
		minorValue, err := strconv.Atoi(minor)
		if err != nil || minorValue < minimumTrackedGoMinor {
			continue
		}
		version := strings.TrimPrefix(entry.Version, "go")
		if current, ok := latest[minor]; !ok || compareGoVersion(version, current) > 0 {
			latest[minor] = version
		}
	}

	return latest, nil
}

func parseSupportThreshold(item issueItem) (supportThreshold, bool) {
	titleMatch := issueTitlePattern.FindStringSubmatch(strings.ToLower(strings.TrimSpace(item.Title)))
	if len(titleMatch) != 2 {
		return supportThreshold{}, false
	}
	goMinor, err := strconv.Atoi(titleMatch[1])
	if err != nil {
		return supportThreshold{}, false
	}

	bodyMatches := issueSincePattern.FindAllStringSubmatch(item.Body, -1)
	if len(bodyMatches) == 0 {
		return supportThreshold{}, false
	}

	sinceVersion := ""
	for _, match := range bodyMatches {
		version := match[1]
		if sinceVersion == "" || compareSemver(version, sinceVersion) > 0 {
			sinceVersion = version
		}
	}
	if sinceVersion == "" {
		return supportThreshold{}, false
	}

	return supportThreshold{
		GoMinor:      goMinor,
		SinceVersion: sinceVersion,
	}, true
}

func buildManifestEntries(releases []release, thresholds []supportThreshold) map[string]compatibilityEntry {
	entries := make(map[string]compatibilityEntry)
	if len(thresholds) == 0 {
		return entries
	}

	fallbackMaxMinor := max(thresholds[0].GoMinor-1, minimumTrackedGoMinor)

	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		if !strings.HasPrefix(release.TagName, "v") {
			continue
		}

		maxMinor := fallbackMaxMinor
		for _, threshold := range thresholds {
			if compareSemver(release.TagName, threshold.SinceVersion) >= 0 && threshold.GoMinor > maxMinor {
				maxMinor = threshold.GoMinor
			}
		}

		entries[release.TagName] = compatibilityEntry{
			MinGoMinor: minimumTrackedGoMinor,
			MaxGoMinor: maxMinor,
		}
	}

	return entries
}

func stringifyThresholds(thresholds []supportThreshold) map[string]string {
	values := make(map[string]string, len(thresholds))
	for _, threshold := range thresholds {
		values[strconv.Itoa(threshold.GoMinor)] = threshold.SinceVersion
	}

	return values
}

func goMinorString(version string) (string, bool) {
	trimmed := strings.TrimPrefix(version, "go")
	parts := strings.Split(trimmed, ".")
	if len(parts) < 2 {
		return "", false
	}

	if _, err := strconv.Atoi(parts[1]); err != nil {
		return "", false
	}

	return parts[1], true
}

func compareGoVersion(left string, right string) int {
	lmajor, lminor, lpatch := parseSemverParts(left)
	rmajor, rminor, rpatch := parseSemverParts(right)

	switch {
	case lmajor != rmajor:
		return lmajor - rmajor
	case lminor != rminor:
		return lminor - rminor
	default:
		return lpatch - rpatch
	}
}

func compareSemver(left string, right string) int {
	lmaj, lmin, lpatch := parseSemverParts(left)
	rmaj, rmin, rpatch := parseSemverParts(right)

	switch {
	case lmaj != rmaj:
		return lmaj - rmaj
	case lmin != rmin:
		return lmin - rmin
	case lpatch != rpatch:
		return lpatch - rpatch
	default:
		return strings.Compare(left, right)
	}
}
