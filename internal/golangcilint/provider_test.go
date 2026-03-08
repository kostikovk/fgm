package golangcilint

import (
	"context"
	"net/http"
	"net/http/httptest"
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
