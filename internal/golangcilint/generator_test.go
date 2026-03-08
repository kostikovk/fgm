package golangcilint

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGeneratorGenerate_BuildsManifestFromUpstreamData(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/golangci/golangci-lint/releases":
			_, _ = w.Write([]byte(`[
				{"tag_name":"v2.11.2","draft":false,"prerelease":false},
				{"tag_name":"v2.4.0","draft":false,"prerelease":false},
				{"tag_name":"v1.64.2","draft":false,"prerelease":false},
				{"tag_name":"v1.60.1","draft":false,"prerelease":false},
				{"tag_name":"v1.56.0","draft":false,"prerelease":false},
				{"tag_name":"v1.54.1","draft":false,"prerelease":false},
				{"tag_name":"v1.39.0","draft":false,"prerelease":false},
				{"tag_name":"v2.12.0-rc1","draft":false,"prerelease":true}
			]`))
		case "/search/issues":
			_, _ = w.Write([]byte(`{
				"items": [
					{"title":"go1.21 support","body":"officially fully supported since [v1.54.1](https://example.test)"},
					{"title":"go1.22 support","body":"EDIT: since v1.56.0 golangci supports go1.22"},
					{"title":"go1.23 support","body":"EDIT: since v1.60.1 golangci-lint supports go1.23"},
					{"title":"go1.24 support","body":"EDIT: since v1.64.2 golangci-lint supports go1.24"},
					{"title":"go1.25 support","body":"EDIT: since v2.4.0 golangci-lint supports go1.25"},
					{"title":"go1.26 support","body":"not supported yet"}
				]
			}`))
		case "/dl/":
			_, _ = w.Write([]byte(`[
				{"version":"go1.26.1","stable":true},
				{"version":"go1.25.8","stable":true},
				{"version":"go1.24.9","stable":true},
				{"version":"go1.26rc2","stable":false}
			]`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	generator := NewGenerator(GeneratorConfig{
		Client:        server.Client(),
		BaseURL:       server.URL,
		GoReleasesURL: server.URL,
		Now: func() time.Time {
			return time.Date(2026, time.March, 8, 12, 0, 0, 0, time.UTC)
		},
	})

	manifest, err := generator.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if manifest.GeneratedAt != "2026-03-08T12:00:00Z" {
		t.Fatalf("GeneratedAt = %q, want %q", manifest.GeneratedAt, "2026-03-08T12:00:00Z")
	}
	if manifest.LatestGoVersions["26"] != "1.26.1" {
		t.Fatalf("LatestGoVersions[26] = %q, want %q", manifest.LatestGoVersions["26"], "1.26.1")
	}
	if manifest.SupportThresholds["25"] != "v2.4.0" {
		t.Fatalf("SupportThresholds[25] = %q, want %q", manifest.SupportThresholds["25"], "v2.4.0")
	}
	if manifest.Versions["v2.11.2"].MaxGoMinor != 25 {
		t.Fatalf("v2.11.2 MaxGoMinor = %d, want %d", manifest.Versions["v2.11.2"].MaxGoMinor, 25)
	}
	if manifest.Versions["v1.54.1"].MaxGoMinor != 21 {
		t.Fatalf("v1.54.1 MaxGoMinor = %d, want %d", manifest.Versions["v1.54.1"].MaxGoMinor, 21)
	}
	if manifest.Versions["v1.39.0"].MaxGoMinor != 20 {
		t.Fatalf("v1.39.0 MaxGoMinor = %d, want %d", manifest.Versions["v1.39.0"].MaxGoMinor, 20)
	}
	if _, ok := manifest.SupportThresholds["26"]; ok {
		t.Fatal("SupportThresholds unexpectedly contains go1.26")
	}
}
