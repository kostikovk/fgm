package goreleases

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://go.dev"

type release struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []file `json:"files"`
}

type file struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Kind     string `json:"kind"`
	SHA256   string `json:"sha256"`
}

// Archive describes a downloadable Go release archive for a specific platform.
type Archive struct {
	Version  string
	Filename string
	URL      string
	SHA256   string
}

// Provider lists Go releases from go.dev filtered for a single platform.
type Provider struct {
	baseURL string
	client  *http.Client
	goos    string
	goarch  string
}

// Config configures a Provider.
type Config struct {
	Client  *http.Client
	BaseURL string
	GOOS    string
	GOARCH  string
}

// New constructs a remote Go release provider for a specific platform.
func New(config Config) *Provider {
	if config.Client == nil {
		config.Client = http.DefaultClient
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{
		baseURL: strings.TrimRight(config.BaseURL, "/"),
		client:  config.Client,
		goos:    config.GOOS,
		goarch:  config.GOARCH,
	}
}

// ListRemoteGoVersions returns stable Go versions available for the configured platform.
func (p *Provider) ListRemoteGoVersions(ctx context.Context) ([]string, error) {
	releases, err := p.fetchReleases(ctx)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, release := range releases {
		if !supportsPlatform(release.Files, p.goos, p.goarch) {
			continue
		}
		versions = append(versions, strings.TrimPrefix(release.Version, "go"))
	}

	return versions, nil
}

// FindArchive returns archive metadata for a Go version on the configured platform.
func (p *Provider) FindArchive(ctx context.Context, version string) (Archive, error) {
	releases, err := p.fetchReleases(ctx)
	if err != nil {
		return Archive{}, err
	}

	normalizedVersion := version
	if !strings.HasPrefix(normalizedVersion, "go") {
		normalizedVersion = "go" + normalizedVersion
	}

	for _, release := range releases {
		if release.Version != normalizedVersion {
			continue
		}

		for _, file := range release.Files {
			if file.OS != p.goos || file.Arch != p.goarch {
				continue
			}
			if file.Kind != "archive" {
				continue
			}

			return Archive{
				Version:  strings.TrimPrefix(release.Version, "go"),
				Filename: file.Filename,
				URL:      p.baseURL + "/dl/" + file.Filename,
				SHA256:   file.SHA256,
			}, nil
		}
	}

	return Archive{}, fmt.Errorf("Go archive for version %s and platform %s/%s was not found", version, p.goos, p.goarch)
}

func (p *Provider) fetchReleases(ctx context.Context) ([]release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/dl/?mode=json&include=all", nil)
	if err != nil {
		return nil, fmt.Errorf("build Go releases request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch Go releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch Go releases: unexpected status %s", resp.Status)
	}

	var releases []release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decode Go releases: %w", err)
	}

	return releases, nil
}

func supportsPlatform(files []file, goos string, goarch string) bool {
	for _, file := range files {
		if file.OS == goos && file.Arch == goarch {
			return true
		}
	}

	return false
}
