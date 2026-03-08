package goreleases

import (
	"context"
	"net/http"
	"net/http/httptest"
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

	if len(versions) != 3 {
		t.Fatalf("len(versions) = %d, want %d", len(versions), 3)
	}
	if versions[0] != "1.25.2" {
		t.Fatalf("versions[0] = %q, want %q", versions[0], "1.25.2")
	}
	if versions[1] != "1.25rc1" {
		t.Fatalf("versions[1] = %q, want %q", versions[1], "1.25rc1")
	}
	if versions[2] != "1.20.14" {
		t.Fatalf("versions[2] = %q, want %q", versions[2], "1.20.14")
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
