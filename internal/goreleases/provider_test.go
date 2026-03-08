package goreleases

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProviderListRemoteGoVersions_FiltersByPlatform(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dl/" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/dl/")
		}
		if got := r.URL.Query().Get("mode"); got != "json" {
			t.Fatalf("mode = %q, want %q", got, "json")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"version": "go1.25.2",
				"stable": true,
				"files": [
					{"os": "darwin", "arch": "arm64", "kind": "archive"},
					{"os": "linux", "arch": "amd64", "kind": "archive"}
				]
			},
			{
				"version": "go1.25.1",
				"stable": true,
				"files": [
					{"os": "linux", "arch": "amd64", "kind": "archive"}
				]
			},
			{
				"version": "go1.25rc1",
				"stable": false,
				"files": [
					{"os": "darwin", "arch": "arm64", "kind": "archive"}
				]
			},
			{
				"version": "go1.20.14",
				"stable": true,
				"files": [
					{"os": "darwin", "arch": "arm64", "kind": "archive"}
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
	})
	versions, err := provider.ListRemoteGoVersions(context.Background())
	if err != nil {
		t.Fatalf("ListRemoteGoVersions: %v", err)
	}

	if len(versions) != 2 {
		t.Fatalf("len(versions) = %d, want %d", len(versions), 2)
	}
	if versions[0] != "1.25.2" {
		t.Fatalf("versions[0] = %q, want %q", versions[0], "1.25.2")
	}
	if versions[1] != "1.20.14" {
		t.Fatalf("versions[1] = %q, want %q", versions[1], "1.20.14")
	}
}

func TestProviderFindArchive_ReturnsArchiveMetadataForPlatform(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"version": "go1.25.7",
				"stable": true,
				"files": [
					{
						"filename": "go1.25.7.darwin-arm64.tar.gz",
						"os": "darwin",
						"arch": "arm64",
						"kind": "archive",
						"sha256": "abc123"
					},
					{
						"filename": "go1.25.7.windows-amd64.msi",
						"os": "windows",
						"arch": "amd64",
						"kind": "installer",
						"sha256": "skipme"
					}
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
	})
	archive, err := provider.FindArchive(context.Background(), "1.25.7")
	if err != nil {
		t.Fatalf("FindArchive: %v", err)
	}

	if archive.Filename != "go1.25.7.darwin-arm64.tar.gz" {
		t.Fatalf("archive.Filename = %q, want %q", archive.Filename, "go1.25.7.darwin-arm64.tar.gz")
	}
	if archive.SHA256 != "abc123" {
		t.Fatalf("archive.SHA256 = %q, want %q", archive.SHA256, "abc123")
	}
}

func TestProviderListRemoteGoVersions_ReturnsErrorForInvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:  server.Client(),
		BaseURL: server.URL,
		GOOS:    "darwin",
		GOARCH:  "arm64",
	})
	if _, err := provider.ListRemoteGoVersions(context.Background()); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestProviderNew_DefaultsClientAndBaseURL(t *testing.T) {
	t.Parallel()

	provider := New(Config{
		GOOS:   "linux",
		GOARCH: "amd64",
	})
	if provider.client == nil {
		t.Fatal("client is nil, want non-nil default")
	}
	if provider.baseURL == "" {
		t.Fatal("baseURL is empty, want non-empty default")
	}
}

func TestFindArchive_ErrorForMissingVersion(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"version": "go1.25.2",
				"stable": true,
				"files": [
					{"os": "darwin", "arch": "arm64", "kind": "archive"}
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
	})
	_, err := provider.FindArchive(context.Background(), "1.99.0")
	if err == nil {
		t.Fatal("expected an error for missing version")
	}
	if !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "was not found")
	}
}

func TestFindArchive_SkipsInstallerKind(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"version": "go1.25.7",
				"stable": true,
				"files": [
					{
						"filename": "go1.25.7.darwin-arm64.pkg",
						"os": "darwin",
						"arch": "arm64",
						"kind": "installer",
						"sha256": "skipme"
					}
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
	})
	_, err := provider.FindArchive(context.Background(), "1.25.7")
	if err == nil {
		t.Fatal("expected an error when only installer kind is available")
	}
	if !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "was not found")
	}
}

func TestFetchReleases_ErrorForNon200(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := New(Config{
		Client:  server.Client(),
		BaseURL: server.URL,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})
	_, err := provider.ListRemoteGoVersions(context.Background())
	if err == nil {
		t.Fatal("expected an error for non-200 status")
	}
	if !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "unexpected status")
	}
}

func TestFindArchive_SkipsMismatchedPlatformBeforeMatchingArchive(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"version": "go1.25.7",
				"stable": true,
				"files": [
					{
						"filename": "go1.25.7.linux-amd64.tar.gz",
						"os": "linux",
						"arch": "amd64",
						"kind": "archive",
						"sha256": "skipme"
					},
					{
						"filename": "go1.25.7.darwin-arm64.tar.gz",
						"os": "darwin",
						"arch": "arm64",
						"kind": "archive",
						"sha256": "abc123"
					}
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
	})

	archive, err := provider.FindArchive(context.Background(), "1.25.7")
	if err != nil {
		t.Fatalf("FindArchive: %v", err)
	}
	if archive.Filename != "go1.25.7.darwin-arm64.tar.gz" {
		t.Fatalf("archive.Filename = %q, want darwin archive", archive.Filename)
	}
}

func TestFetchReleases_ClientDoError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	provider := New(Config{
		Client:  server.Client(),
		BaseURL: server.URL,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})

	_, err := provider.ListRemoteGoVersions(context.Background())
	if err == nil {
		t.Fatal("expected fetch releases client error")
	}
	if !strings.Contains(err.Error(), "fetch Go releases") {
		t.Fatalf("error = %q, want fetch error", err)
	}
}

func TestFindArchive_HandlesGoPrefix(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"version": "go1.25.7",
				"stable": true,
				"files": [
					{
						"filename": "go1.25.7.linux-amd64.tar.gz",
						"os": "linux",
						"arch": "amd64",
						"kind": "archive",
						"sha256": "def456"
					}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:  server.Client(),
		BaseURL: server.URL,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})
	archive, err := provider.FindArchive(context.Background(), "go1.25.7")
	if err != nil {
		t.Fatalf("FindArchive: %v", err)
	}
	if archive.Version != "1.25.7" {
		t.Fatalf("archive.Version = %q, want %q", archive.Version, "1.25.7")
	}
	if archive.Filename != "go1.25.7.linux-amd64.tar.gz" {
		t.Fatalf("archive.Filename = %q, want %q", archive.Filename, "go1.25.7.linux-amd64.tar.gz")
	}
	if archive.SHA256 != "def456" {
		t.Fatalf("archive.SHA256 = %q, want %q", archive.SHA256, "def456")
	}
}

func TestFindArchive_FetchReleasesRequestBuildError(t *testing.T) {
	t.Parallel()

	// A URL with a control character causes http.NewRequestWithContext to fail.
	provider := New(Config{
		Client:  http.DefaultClient,
		BaseURL: "http://invalid\x7f",
		GOOS:    "linux",
		GOARCH:  "amd64",
	})

	_, err := provider.FindArchive(context.Background(), "1.25.7")
	if err == nil {
		t.Fatal("expected an error for invalid URL in request build")
	}
	if !strings.Contains(err.Error(), "build Go releases request") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "build Go releases request")
	}
}

func TestFetchReleases_RequestBuildError(t *testing.T) {
	t.Parallel()

	// A URL with a control character causes http.NewRequestWithContext to fail.
	provider := New(Config{
		Client:  http.DefaultClient,
		BaseURL: "http://invalid\x7f",
		GOOS:    "linux",
		GOARCH:  "amd64",
	})

	_, err := provider.ListRemoteGoVersions(context.Background())
	if err == nil {
		t.Fatal("expected an error for invalid URL in request build")
	}
	if !strings.Contains(err.Error(), "build Go releases request") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "build Go releases request")
	}
}

func TestFetchReleases_JSONDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json at all`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:  server.Client(),
		BaseURL: server.URL,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})

	_, err := provider.ListRemoteGoVersions(context.Background())
	if err == nil {
		t.Fatal("expected an error for invalid JSON in response")
	}
	if !strings.Contains(err.Error(), "decode Go releases") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "decode Go releases")
	}
}

func TestFindArchive_SkipsSourceKind(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"version": "go1.25.7",
				"stable": true,
				"files": [
					{
						"filename": "go1.25.7.src.tar.gz",
						"os": "linux",
						"arch": "amd64",
						"kind": "source",
						"sha256": "srcsha"
					}
				]
			}
		]`))
	}))
	defer server.Close()

	provider := New(Config{
		Client:  server.Client(),
		BaseURL: server.URL,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})

	_, err := provider.FindArchive(context.Background(), "1.25.7")
	if err == nil {
		t.Fatal("expected an error when only source kind is available")
	}
	if !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "was not found")
	}
}
