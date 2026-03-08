package golangcilint

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestProviderListRemoteLintVersions_FiltersByPlatformAndCompatibility(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/golangci/golangci-lint/releases" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/repos/golangci/golangci-lint/releases")
		}
		if got := r.URL.Query().Get("per_page"); got != "100" {
			t.Fatalf("per_page = %q, want %q", got, "100")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v2.6.2",
				"assets": [
					{"name": "golangci-lint-2.6.2-darwin-arm64.tar.gz"}
				]
			},
			{
				"tag_name": "v2.6.1",
				"assets": [
					{"name": "golangci-lint-2.6.1-darwin-arm64.tar.gz"}
				]
			},
			{
				"tag_name": "v2.6.0",
				"assets": [
					{"name": "golangci-lint-2.6.0-darwin-arm64.tar.gz"}
				]
			},
			{
				"tag_name": "v2.5.9",
				"assets": [
					{"name": "golangci-lint-2.5.9-linux-amd64.tar.gz"}
				]
			},
			{
				"tag_name": "v2.5.8",
				"prerelease": true,
				"assets": [
					{"name": "golangci-lint-2.5.8-darwin-arm64.tar.gz"}
				]
			},
			{
				"tag_name": "v2.5.7",
				"assets": [
					{"name": "golangci-lint-2.5.7-darwin-arm64.tar.gz"}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:  server.Client(),
		BaseURL: server.URL,
		GOOS:    "darwin",
		GOARCH:  "arm64",
		CompatibilityData: []byte(`{
			"versions": {
				"v2.6.2": {"min_go_minor": 18, "max_go_minor": 25},
				"v2.6.1": {"min_go_minor": 18, "max_go_minor": 25},
				"v2.6.0": {"min_go_minor": 18, "max_go_minor": 24}
			}
		}`),
	})

	versions, err := provider.ListRemoteLintVersions(context.Background(), "1.25.0")
	if err != nil {
		t.Fatalf("ListRemoteLintVersions: %v", err)
	}

	if len(versions) != 2 {
		t.Fatalf("len(versions) = %d, want %d", len(versions), 2)
	}
	if versions[0].Version != "v2.6.2" {
		t.Fatalf("versions[0].Version = %q, want %q", versions[0].Version, "v2.6.2")
	}
	if !versions[0].Recommended {
		t.Fatal("versions[0].Recommended = false, want true")
	}
	if versions[1].Version != "v2.6.1" {
		t.Fatalf("versions[1].Version = %q, want %q", versions[1].Version, "v2.6.1")
	}
}

func TestProviderListRemoteLintVersions_ExcludesVersionsMissingFromManifest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v2.6.2",
				"assets": [
					{"name": "golangci-lint-2.6.2-darwin-arm64.tar.gz"}
				]
			},
			{
				"tag_name": "v2.6.1",
				"assets": [
					{"name": "golangci-lint-2.6.1-darwin-arm64.tar.gz"}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:  server.Client(),
		BaseURL: server.URL,
		GOOS:    "darwin",
		GOARCH:  "arm64",
		CompatibilityData: []byte(`{
			"versions": {
				"v2.6.1": {"min_go_minor": 18, "max_go_minor": 25}
			}
		}`),
	})

	versions, err := provider.ListRemoteLintVersions(context.Background(), "1.25.0")
	if err != nil {
		t.Fatalf("ListRemoteLintVersions: %v", err)
	}

	if len(versions) != 1 {
		t.Fatalf("len(versions) = %d, want %d", len(versions), 1)
	}
	if versions[0].Version != "v2.6.1" {
		t.Fatalf("versions[0].Version = %q, want %q", versions[0].Version, "v2.6.1")
	}
}

func TestProviderListRemoteLintVersions_ReturnsErrorForInvalidGoVersion(t *testing.T) {
	t.Parallel()

	provider := New(Config{
		Client:            http.DefaultClient,
		GOOS:              "darwin",
		GOARCH:            "arm64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	if _, err := provider.ListRemoteLintVersions(context.Background(), "invalid"); err == nil {
		t.Fatal("expected an error for invalid Go version")
	}
}

func TestProviderListRemoteLintVersions_ReturnsErrorForInvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "darwin",
		GOARCH:            "arm64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	if _, err := provider.ListRemoteLintVersions(context.Background(), "1.25.0"); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestProviderListRemoteLintVersions_ReturnsErrorForInvalidManifest(t *testing.T) {
	t.Parallel()

	provider := New(Config{
		Client:            http.DefaultClient,
		GOOS:              "darwin",
		GOARCH:            "arm64",
		CompatibilityData: []byte(`{`),
	})

	if _, err := provider.ListRemoteLintVersions(context.Background(), "1.25.0"); err == nil {
		t.Fatal("expected an error for invalid compatibility manifest")
	}
}

func TestProviderFindArchive_ReturnsArchiveMetadataForPlatform(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v2.11.2",
				"assets": [
					{
						"name": "golangci-lint-2.11.2-darwin-arm64.tar.gz",
						"browser_download_url": "https://example.test/v2.11.2/darwin-arm64.tar.gz",
						"digest": "sha256:abc123"
					},
					{
						"name": "golangci-lint-2.11.2-linux-amd64.tar.gz",
						"browser_download_url": "https://example.test/v2.11.2/linux-amd64.tar.gz",
						"digest": "sha256:skipme"
					}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "darwin",
		GOARCH:            "arm64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	archive, err := provider.FindArchive(context.Background(), "v2.11.2")
	if err != nil {
		t.Fatalf("FindArchive: %v", err)
	}

	if archive.Filename != "golangci-lint-2.11.2-darwin-arm64.tar.gz" {
		t.Fatalf("archive.Filename = %q, want %q", archive.Filename, "golangci-lint-2.11.2-darwin-arm64.tar.gz")
	}
	if archive.URL != "https://example.test/v2.11.2/darwin-arm64.tar.gz" {
		t.Fatalf("archive.URL = %q, want %q", archive.URL, "https://example.test/v2.11.2/darwin-arm64.tar.gz")
	}
	if archive.SHA256 != "abc123" {
		t.Fatalf("archive.SHA256 = %q, want %q", archive.SHA256, "abc123")
	}
}

func TestParseNextLink_EmptyHeader(t *testing.T) {
	t.Parallel()

	got := parseNextLink("")
	if got != "" {
		t.Fatalf("parseNextLink(%q) = %q, want %q", "", got, "")
	}
}

func TestParseNextLink_NoNextRel(t *testing.T) {
	t.Parallel()

	header := `<https://api.github.com/repos?page=3>; rel="last"`
	got := parseNextLink(header)
	if got != "" {
		t.Fatalf("parseNextLink(%q) = %q, want %q", header, got, "")
	}
}

func TestParseNextLink_ValidNextLink(t *testing.T) {
	t.Parallel()

	header := `<https://api.github.com/repos?page=2>; rel="next", <https://api.github.com/repos?page=5>; rel="last"`
	got := parseNextLink(header)
	want := "https://api.github.com/repos?page=2"
	if got != want {
		t.Fatalf("parseNextLink(%q) = %q, want %q", header, got, want)
	}
}

func TestFetchAllReleasePages_HandlesPagination(t *testing.T) {
	t.Parallel()

	var reqCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := reqCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		switch page {
		case 1:
			w.Header().Set("Link", fmt.Sprintf(`<%s/repos/golangci/golangci-lint/releases?page=2>; rel="next"`, "http://"+r.Host))
			_, _ = w.Write([]byte(`[
				{"tag_name": "v2.1.0", "assets": [{"name": "golangci-lint-2.1.0-linux-amd64.tar.gz"}]}
			]`))
		case 2:
			_, _ = w.Write([]byte(`[
				{"tag_name": "v2.0.0", "assets": [{"name": "golangci-lint-2.0.0-linux-amd64.tar.gz"}]}
			]`))
		default:
			t.Errorf("unexpected request page %d", page)
		}
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	releases, err := provider.fetchAllReleasePages(context.Background())
	if err != nil {
		t.Fatalf("fetchAllReleasePages: %v", err)
	}

	if len(releases) != 2 {
		t.Fatalf("len(releases) = %d, want 2", len(releases))
	}
	if releases[0].TagName != "v2.1.0" {
		t.Fatalf("releases[0].TagName = %q, want %q", releases[0].TagName, "v2.1.0")
	}
	if releases[1].TagName != "v2.0.0" {
		t.Fatalf("releases[1].TagName = %q, want %q", releases[1].TagName, "v2.0.0")
	}
	if got := reqCount.Load(); got != 2 {
		t.Fatalf("request count = %d, want 2", got)
	}
}

func TestFetchAllReleasePages_ErrorForNon200(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.fetchAllReleasePages(context.Background())
	if err == nil {
		t.Fatal("expected an error for non-200 status")
	}
	if !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "unexpected status")
	}
}

func TestFindArchive_ErrorForMissingVersion(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v2.11.2",
				"assets": [
					{"name": "golangci-lint-2.11.2-darwin-arm64.tar.gz", "browser_download_url": "https://example.test/v2.11.2/darwin-arm64.tar.gz"}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "darwin",
		GOARCH:            "arm64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.FindArchive(context.Background(), "v9.9.9")
	if err == nil {
		t.Fatal("expected an error for missing version")
	}
	if !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "was not found")
	}
}

func TestFindArchive_SkipsNonArchiveAssets(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v2.11.2",
				"assets": [
					{"name": "golangci-lint-2.11.2-darwin-arm64.deb", "browser_download_url": "https://example.test/v2.11.2/darwin-arm64.deb"}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "darwin",
		GOARCH:            "arm64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.FindArchive(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatal("expected an error when only .deb assets exist")
	}
	if !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "was not found")
	}
}

func TestNew_NilClientDefaults(t *testing.T) {
	t.Parallel()

	provider := New(Config{
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	if provider.client != http.DefaultClient {
		t.Fatal("expected nil Client to default to http.DefaultClient")
	}
}

func TestNew_EmptyBaseURLDefaults(t *testing.T) {
	t.Parallel()

	provider := New(Config{
		Client:            http.DefaultClient,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	if provider.baseURL != defaultBaseURL {
		t.Fatalf("baseURL = %q, want %q", provider.baseURL, defaultBaseURL)
	}
}

func TestPickCompatibilityData_ReturnsDefaultWhenConfigEmpty(t *testing.T) {
	t.Parallel()

	got := pickCompatibilityData(nil)
	if len(got) == 0 {
		t.Fatal("expected non-empty default compatibility data")
	}
	// It should return the embedded default data.
	if &got[0] != &defaultCompatibilityData[0] {
		t.Fatal("expected pickCompatibilityData(nil) to return the embedded default")
	}
}

func TestParseSemverParts_TooFewParts(t *testing.T) {
	t.Parallel()

	major, minor, patch := parseSemverParts("v1")
	if major != 0 || minor != 0 || patch != 0 {
		t.Fatalf("parseSemverParts(%q) = (%d, %d, %d), want (0, 0, 0)", "v1", major, minor, patch)
	}
}

func TestLoadCompatibilityManifest_InvalidJSON(t *testing.T) {
	t.Parallel()

	provider := New(Config{
		Client:            http.DefaultClient,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`not valid json`),
	})

	_, err := provider.loadCompatibilityManifest()
	if err == nil {
		t.Fatal("expected an error for invalid JSON in compatibility manifest")
	}
	if !strings.Contains(err.Error(), "decode golangci-lint compatibility manifest") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "decode golangci-lint compatibility manifest")
	}
}

func TestLoadCompatibilityManifest_NilVersionsMap(t *testing.T) {
	t.Parallel()

	provider := New(Config{
		Client:            http.DefaultClient,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{}`),
	})

	manifest, err := provider.loadCompatibilityManifest()
	if err != nil {
		t.Fatalf("loadCompatibilityManifest: %v", err)
	}
	if manifest.Versions == nil {
		t.Fatal("expected non-nil Versions map after loading manifest with no versions key")
	}
	if len(manifest.Versions) != 0 {
		t.Fatalf("len(manifest.Versions) = %d, want 0", len(manifest.Versions))
	}
}

func TestFetchAllReleasePages_RequestBuildError(t *testing.T) {
	t.Parallel()

	// A URL with a control character causes http.NewRequestWithContext to fail.
	provider := New(Config{
		Client:            http.DefaultClient,
		BaseURL:           "http://invalid\x7f",
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.fetchAllReleasePages(context.Background())
	if err == nil {
		t.Fatal("expected an error for invalid URL")
	}
	if !strings.Contains(err.Error(), "build golangci-lint releases request") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "build golangci-lint releases request")
	}
}

func TestFetchAllReleasePages_ClientDoError(t *testing.T) {
	t.Parallel()

	// Use a server that immediately closes to trigger a client.Do error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.fetchAllReleasePages(context.Background())
	if err == nil {
		t.Fatal("expected an error for client.Do failure")
	}
	if !strings.Contains(err.Error(), "fetch golangci-lint releases") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "fetch golangci-lint releases")
	}
}

func TestFetchAllReleasePages_JSONDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.fetchAllReleasePages(context.Background())
	if err == nil {
		t.Fatal("expected an error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "decode golangci-lint releases") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "decode golangci-lint releases")
	}
}

func TestFindArchive_VersionNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v2.11.2",
				"assets": [
					{"name": "golangci-lint-2.11.2-linux-amd64.tar.gz", "browser_download_url": "https://example.test/v2.11.2"}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.FindArchive(context.Background(), "v99.99.99")
	if err == nil {
		t.Fatal("expected an error for version not found")
	}
	if !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "was not found")
	}
}

func TestParseGoMinor_InvalidVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
	}{
		{"too few parts", "1"},
		{"non-numeric minor", "1.abc"},
		{"go prefix too few parts", "go1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseGoMinor(tt.version)
			if err == nil {
				t.Fatalf("parseGoMinor(%q): expected error, got nil", tt.version)
			}
			if !strings.Contains(err.Error(), "invalid Go version") {
				t.Fatalf("error = %q, want it to contain %q", err.Error(), "invalid Go version")
			}
		})
	}
}

func TestFindArchive_FetchReleasesError(t *testing.T) {
	t.Parallel()

	// Use a closed server to trigger a fetch error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.FindArchive(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatal("expected an error when fetch releases fails")
	}
}

func TestListRemoteLintVersions_FetchReleasesError(t *testing.T) {
	t.Parallel()

	// Use a closed server to trigger a fetch error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	_, err := provider.ListRemoteLintVersions(context.Background(), "1.25.0")
	if err == nil {
		t.Fatal("expected an error when fetch releases fails")
	}
}

func TestFetchAllReleasePages_StopsAfterLimit(t *testing.T) {
	t.Parallel()

	requests := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", `<`+server.URL+`/repos/golangci/golangci-lint/releases?page=next>; rel="next"`)
		_, _ = w.Write([]byte(`[` + strings.TrimRight(strings.Repeat(`{"tag_name":"v2.11.2"},`, 101), ",") + `]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "linux",
		GOARCH:            "amd64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	releases, err := provider.fetchAllReleasePages(context.Background())
	if err != nil {
		t.Fatalf("fetchAllReleasePages: %v", err)
	}
	if len(releases) <= 500 {
		t.Fatalf("len(releases) = %d, want > 500 to trigger limit", len(releases))
	}
	if requests < 5 {
		t.Fatalf("requests = %d, want multiple pages", requests)
	}
}

func TestFindArchive_SkipsMismatchedPlatformAsset(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v2.11.2",
				"assets": [
					{"name": "golangci-lint-2.11.2-linux-amd64.tar.gz", "browser_download_url": "https://example.test/linux"},
					{"name": "golangci-lint-2.11.2-darwin-arm64.zip", "browser_download_url": "https://example.test/darwin", "digest": "sha256:abc123"}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:            server.Client(),
		BaseURL:           server.URL,
		GOOS:              "darwin",
		GOARCH:            "arm64",
		CompatibilityData: []byte(`{"versions":{}}`),
	})

	archive, err := provider.FindArchive(context.Background(), "v2.11.2")
	if err != nil {
		t.Fatalf("FindArchive: %v", err)
	}
	if archive.URL != "https://example.test/darwin" {
		t.Fatalf("archive.URL = %q, want darwin asset URL", archive.URL)
	}
}
