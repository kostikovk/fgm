package golangcilint

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/kostikovk/fgm/internal/app"
)

const defaultBaseURL = "https://api.github.com"

//go:embed compatibility.json
var defaultCompatibilityData []byte

type release struct {
	TagName    string  `json:"tag_name"`
	Body       string  `json:"body"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

// Archive describes a downloadable golangci-lint archive for a specific platform.
type Archive struct {
	Version  string
	Filename string
	URL      string
	SHA256   string
}

// Config configures a golangci-lint release provider.
type Config struct {
	Client            *http.Client
	BaseURL           string
	GOOS              string
	GOARCH            string
	CompatibilityData []byte
}

// Provider lists remote golangci-lint releases for a single platform.
type Provider struct {
	client            *http.Client
	baseURL           string
	goos              string
	goarch            string
	compatibilityData []byte

	releasesOnce      sync.Once
	cachedReleases    []release
	cachedReleasesErr error
}

type compatibilityManifest struct {
	Versions map[string]compatibilityEntry `json:"versions"`
}

type compatibilityEntry struct {
	MinGoMinor int `json:"min_go_minor"`
	MaxGoMinor int `json:"max_go_minor"`
}

// New constructs a remote golangci-lint provider.
func New(config Config) *Provider {
	if config.Client == nil {
		config.Client = http.DefaultClient
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		client:            config.Client,
		baseURL:           strings.TrimRight(config.BaseURL, "/"),
		goos:              config.GOOS,
		goarch:            config.GOARCH,
		compatibilityData: pickCompatibilityData(config.CompatibilityData),
	}
}

// ListRemoteLintVersions returns compatible golangci-lint versions for a target Go version.
func (p *Provider) ListRemoteLintVersions(
	ctx context.Context,
	goVersion string,
) ([]app.LintVersion, error) {
	targetMinor, err := parseGoMinor(goVersion)
	if err != nil {
		return nil, err
	}
	manifest, err := p.loadCompatibilityManifest()
	if err != nil {
		return nil, err
	}

	releases, err := p.fetchReleases(ctx)
	if err != nil {
		return nil, err
	}

	versions := make([]app.LintVersion, 0, len(releases))
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		if !supportsPlatform(release.Assets, p.goos, p.goarch) {
			continue
		}

		entry, ok := manifest.Versions[release.TagName]
		if !ok {
			continue
		}
		if targetMinor < entry.MinGoMinor || targetMinor > entry.MaxGoMinor {
			continue
		}

		versions = append(versions, app.LintVersion{Version: release.TagName})
	}

	slices.SortFunc(versions, func(a app.LintVersion, b app.LintVersion) int {
		return compareSemverDesc(a.Version, b.Version)
	})
	if len(versions) > 0 {
		versions[0].Recommended = true
	}

	return versions, nil
}

func (p *Provider) loadCompatibilityManifest() (compatibilityManifest, error) {
	var manifest compatibilityManifest
	if err := json.Unmarshal(p.compatibilityData, &manifest); err != nil {
		return compatibilityManifest{}, fmt.Errorf("decode golangci-lint compatibility manifest: %w", err)
	}
	if manifest.Versions == nil {
		manifest.Versions = map[string]compatibilityEntry{}
	}

	return manifest, nil
}

func (p *Provider) fetchReleases(ctx context.Context) ([]release, error) {
	p.releasesOnce.Do(func() {
		p.cachedReleases, p.cachedReleasesErr = p.fetchAllReleasePages(ctx)
	})
	return p.cachedReleases, p.cachedReleasesErr
}

var linkNextPattern = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func (p *Provider) fetchAllReleasePages(ctx context.Context) ([]release, error) {
	url := p.baseURL + "/repos/golangci/golangci-lint/releases?per_page=100"
	var allReleases []release

	for url != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("build golangci-lint releases request: %w", err)
		}

		resp, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch golangci-lint releases: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("fetch golangci-lint releases: unexpected status %s", resp.Status)
		}

		var page []release
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("decode golangci-lint releases: %w", err)
		}
		_ = resp.Body.Close()

		allReleases = append(allReleases, page...)

		url = parseNextLink(resp.Header.Get("Link"))
		if len(allReleases) > 500 {
			break
		}
	}

	return allReleases, nil
}

func parseNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	match := linkNextPattern.FindStringSubmatch(linkHeader)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

// FindArchive returns archive metadata for a golangci-lint version on the configured platform.
func (p *Provider) FindArchive(ctx context.Context, version string) (Archive, error) {
	releases, err := p.fetchReleases(ctx)
	if err != nil {
		return Archive{}, err
	}

	for _, release := range releases {
		if release.TagName != version {
			continue
		}

		for _, asset := range release.Assets {
			if !assetMatchesPlatform(asset.Name, p.goos, p.goarch) {
				continue
			}
			if !strings.HasSuffix(asset.Name, ".tar.gz") && !strings.HasSuffix(asset.Name, ".zip") {
				continue
			}

			return Archive{
				Version:  version,
				Filename: asset.Name,
				URL:      asset.BrowserDownloadURL,
				SHA256:   strings.TrimPrefix(asset.Digest, "sha256:"),
			}, nil
		}
	}

	return Archive{}, fmt.Errorf(
		"golangci-lint archive for version %s and platform %s/%s was not found",
		version,
		p.goos,
		p.goarch,
	)
}

func parseGoMinor(version string) (int, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(version), "go")
	parts := strings.Split(trimmed, ".")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid Go version %q", version)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid Go version %q", version)
	}

	return minor, nil
}

func supportsPlatform(assets []asset, goos string, goarch string) bool {
	for _, asset := range assets {
		if assetMatchesPlatform(asset.Name, goos, goarch) {
			return true
		}
	}

	return false
}

func assetMatchesPlatform(name string, goos string, goarch string) bool {
	needle := "-" + goos + "-" + goarch + "."
	return strings.Contains(name, needle)
}

func pickCompatibilityData(configData []byte) []byte {
	if len(configData) > 0 {
		return configData
	}

	return defaultCompatibilityData
}

func compareSemverDesc(left string, right string) int {
	return -compareSemver(left, right)
}

func parseSemverParts(version string) (int, int, int) {
	trimmed := strings.TrimPrefix(version, "v")
	parts := strings.Split(trimmed, ".")
	if len(parts) < 2 {
		return 0, 0, 0
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch := 0
	if len(parts) >= 3 {
		patch, _ = strconv.Atoi(parts[2])
	}

	return major, minor, patch
}
