package golangcilint

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProviderGitHubAuthentication(t *testing.T) {
	t.Parallel()

	t.Run("FindArchive", func(t *testing.T) {
		t.Parallel()

		t.Run("should send the token as a bearer credential when a GitHub token is configured", func(t *testing.T) {
			t.Parallel()

			// Arrange
			var gotAuth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth = r.Header.Get("Authorization")
				_, _ = w.Write([]byte(`[{"tag_name": "v2.9.0", "assets": [{"name": "golangci-lint-2.9.0-linux-amd64.tar.gz", "browser_download_url": "https://example.invalid/a.tar.gz"}]}]`))
			}))
			t.Cleanup(server.Close)
			provider := New(Config{BaseURL: server.URL, GOOS: "linux", GOARCH: "amd64", Token: "ghs_test"})

			// Act
			_, err := provider.FindArchive(context.Background(), "v2.9.0")

			// Assert
			if err != nil {
				t.Fatalf("FindArchive: %v", err)
			}
			if gotAuth != "Bearer ghs_test" {
				t.Errorf("Authorization = %q; want the bearer token", gotAuth)
			}
		})

		t.Run("should name GITHUB_TOKEN in the error when the GitHub API rate-limits the request", func(t *testing.T) {
			t.Parallel()

			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			}))
			t.Cleanup(server.Close)
			provider := New(Config{BaseURL: server.URL, GOOS: "linux", GOARCH: "amd64"})

			// Act
			_, err := provider.FindArchive(context.Background(), "v2.9.0")

			// Assert
			if err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
				t.Errorf("err = %v; want a GITHUB_TOKEN hint", err)
			}
		})
	})
}
