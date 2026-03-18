package golangcilint

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func newResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestNewGenerator_Defaults(t *testing.T) {
	t.Parallel()

	generator := NewGenerator(GeneratorConfig{})
	if generator.client != http.DefaultClient {
		t.Fatal("expected default client")
	}
	if generator.baseURL != defaultBaseURL {
		t.Fatalf("baseURL = %q, want %q", generator.baseURL, defaultBaseURL)
	}
	if generator.goReleasesURL != defaultGoReleasesURL {
		t.Fatalf("goReleasesURL = %q, want %q", generator.goReleasesURL, defaultGoReleasesURL)
	}
	if generator.now == nil {
		t.Fatal("expected default clock")
	}
}

func TestFetchLintReleases_SetsAuthorizationHeader(t *testing.T) {
	t.Parallel()

	client := newHTTPClient(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want Bearer header", got)
		}
		if req.URL.Path != "/repos/golangci/golangci-lint/releases" {
			t.Fatalf("path = %q, want releases path", req.URL.Path)
		}
		return newResponse(http.StatusOK, `[{"tag_name":"v2.11.2"}]`), nil
	})

	generator := NewGenerator(GeneratorConfig{
		Client:      client,
		BaseURL:     "https://example.test",
		GitHubToken: "test-token",
	})

	releases, err := generator.fetchLintReleases(context.Background())
	if err != nil {
		t.Fatalf("fetchLintReleases: %v", err)
	}
	if len(releases) != 1 || releases[0].TagName != "v2.11.2" {
		t.Fatalf("releases = %#v, want v2.11.2", releases)
	}
}

func TestFetchLintReleases_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		resp *http.Response
		want string
	}{
		{
			name: "non-200 status",
			resp: newResponse(http.StatusBadGateway, `[]`),
			want: "unexpected status",
		},
		{
			name: "invalid json",
			resp: newResponse(http.StatusOK, `{`),
			want: "decode golangci-lint releases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(GeneratorConfig{
				Client: newHTTPClient(func(req *http.Request) (*http.Response, error) {
					return tt.resp, nil
				}),
				BaseURL: "https://example.test",
			})

			_, err := generator.fetchLintReleases(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err, tt.want)
			}
		})
	}
}

func TestFetchLintReleases_RequestBuildAndClientError(t *testing.T) {
	t.Parallel()

	generator := NewGenerator(GeneratorConfig{
		Client:  http.DefaultClient,
		BaseURL: "http://invalid\x7f",
	})
	if _, err := generator.fetchLintReleases(context.Background()); err == nil || !strings.Contains(err.Error(), "build golangci-lint releases request") {
		t.Fatalf("err = %v, want request build error", err)
	}

	generator = NewGenerator(GeneratorConfig{
		Client: newHTTPClient(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("client boom")
		}),
		BaseURL: "https://example.test",
	})
	if _, err := generator.fetchLintReleases(context.Background()); err == nil || !strings.Contains(err.Error(), "fetch golangci-lint releases") {
		t.Fatalf("err = %v, want client error", err)
	}
}

func TestFetchSupportThresholds_SortsAndSkipsInvalidItems(t *testing.T) {
	t.Parallel()

	client := newHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/search/issues" {
			t.Fatalf("path = %q, want issues path", req.URL.Path)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want Bearer header", got)
		}
		return newResponse(http.StatusOK, `{
			"items": [
				{"title":"go1.22 support","body":"since v1.56.0"},
				{"title":"go1.24 support","body":"supported since v1.64.2"},
				{"title":"go1.22 support","body":"since v1.56.0 and then since v1.57.1"},
				{"title":"not a support issue","body":"since v1.0.0"},
				{"title":"go1.23 support","body":"not yet"}
			]
		}`), nil
	})

	generator := NewGenerator(GeneratorConfig{
		Client:      client,
		BaseURL:     "https://example.test",
		GitHubToken: "test-token",
	})

	thresholds, err := generator.fetchSupportThresholds(context.Background())
	if err != nil {
		t.Fatalf("fetchSupportThresholds: %v", err)
	}
	if len(thresholds) != 3 {
		t.Fatalf("len(thresholds) = %d, want 3", len(thresholds))
	}
	if thresholds[0].GoMinor != 22 || thresholds[0].SinceVersion != "v1.56.0" {
		t.Fatalf("thresholds[0] = %#v, want go1.22 -> v1.56.0", thresholds[0])
	}
	if thresholds[1].GoMinor != 22 || thresholds[1].SinceVersion != "v1.57.1" {
		t.Fatalf("thresholds[1] = %#v, want go1.22 -> v1.57.1", thresholds[1])
	}
	if thresholds[2].GoMinor != 24 || thresholds[2].SinceVersion != "v1.64.2" {
		t.Fatalf("thresholds[2] = %#v, want go1.24 -> v1.64.2", thresholds[2])
	}
}

func TestFetchSupportThresholds_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		resp *http.Response
		want string
	}{
		{
			name: "non-200 status",
			resp: newResponse(http.StatusBadGateway, `{}`),
			want: "unexpected status",
		},
		{
			name: "invalid json",
			resp: newResponse(http.StatusOK, `{`),
			want: "decode golangci-lint issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(GeneratorConfig{
				Client: newHTTPClient(func(req *http.Request) (*http.Response, error) {
					return tt.resp, nil
				}),
				BaseURL: "https://example.test",
			})

			_, err := generator.fetchSupportThresholds(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err, tt.want)
			}
		})
	}
}

func TestFetchSupportThresholds_RequestBuildAndClientError(t *testing.T) {
	t.Parallel()

	generator := NewGenerator(GeneratorConfig{
		Client:  http.DefaultClient,
		BaseURL: "http://invalid\x7f",
	})
	if _, err := generator.fetchSupportThresholds(context.Background()); err == nil || !strings.Contains(err.Error(), "build golangci-lint issues request") {
		t.Fatalf("err = %v, want request build error", err)
	}

	generator = NewGenerator(GeneratorConfig{
		Client: newHTTPClient(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("client boom")
		}),
		BaseURL: "https://example.test",
	})
	if _, err := generator.fetchSupportThresholds(context.Background()); err == nil || !strings.Contains(err.Error(), "fetch golangci-lint issues") {
		t.Fatalf("err = %v, want client error", err)
	}
}

func TestFetchLatestGoVersions_FiltersStableAndHighestPatch(t *testing.T) {
	t.Parallel()

	client := newHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/dl/" {
			t.Fatalf("path = %q, want /dl/", req.URL.Path)
		}
		return newResponse(http.StatusOK, `[
			{"version":"go1.25.7","stable":true},
			{"version":"go1.25.8","stable":true},
			{"version":"go1.24.9","stable":true},
			{"version":"go1.17.13","stable":true},
			{"version":"go1.26rc1","stable":false},
			{"version":"broken","stable":true}
		]`), nil
	})

	generator := NewGenerator(GeneratorConfig{
		Client:        client,
		GoReleasesURL: "https://example.test",
	})

	latest, err := generator.fetchLatestGoVersions(context.Background())
	if err != nil {
		t.Fatalf("fetchLatestGoVersions: %v", err)
	}
	if latest["25"] != "1.25.8" {
		t.Fatalf("latest[25] = %q, want %q", latest["25"], "1.25.8")
	}
	if latest["24"] != "1.24.9" {
		t.Fatalf("latest[24] = %q, want %q", latest["24"], "1.24.9")
	}
	if _, ok := latest["17"]; ok {
		t.Fatal("unexpected tracked go1.17 entry")
	}
}

func TestFetchLatestGoVersions_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		resp *http.Response
		want string
	}{
		{
			name: "non-200 status",
			resp: newResponse(http.StatusBadGateway, `[]`),
			want: "unexpected status",
		},
		{
			name: "invalid json",
			resp: newResponse(http.StatusOK, `{`),
			want: "decode Go releases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(GeneratorConfig{
				Client:        newHTTPClient(func(req *http.Request) (*http.Response, error) { return tt.resp, nil }),
				GoReleasesURL: "https://example.test",
			})

			_, err := generator.fetchLatestGoVersions(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err, tt.want)
			}
		})
	}
}

func TestFetchLatestGoVersions_RequestBuildAndClientError(t *testing.T) {
	t.Parallel()

	generator := NewGenerator(GeneratorConfig{
		Client:        http.DefaultClient,
		GoReleasesURL: "http://invalid\x7f",
	})
	if _, err := generator.fetchLatestGoVersions(context.Background()); err == nil || !strings.Contains(err.Error(), "build Go releases request") {
		t.Fatalf("err = %v, want request build error", err)
	}

	generator = NewGenerator(GeneratorConfig{
		Client:        newHTTPClient(func(req *http.Request) (*http.Response, error) { return nil, fmt.Errorf("client boom") }),
		GoReleasesURL: "https://example.test",
	})
	if _, err := generator.fetchLatestGoVersions(context.Background()); err == nil || !strings.Contains(err.Error(), "fetch Go releases") {
		t.Fatalf("err = %v, want client error", err)
	}
}

func TestBuildManifestEntries_UsesThresholdFallbackAndSkipsInvalidReleases(t *testing.T) {
	t.Parallel()

	entries := buildManifestEntries(
		[]release{
			{TagName: "v1.54.1"},
			{TagName: "v1.64.2"},
			{TagName: "v2.4.0"},
			{TagName: "main"},
			{TagName: "v2.5.0", Draft: true},
			{TagName: "v2.6.0", Prerelease: true},
		},
		[]supportThreshold{
			{GoMinor: 21, SinceVersion: "v1.54.1"},
			{GoMinor: 24, SinceVersion: "v1.64.2"},
			{GoMinor: 25, SinceVersion: "v2.4.0"},
		},
		26,
	)

	if entries["v1.54.1"].MaxGoMinor != 21 {
		t.Fatalf("v1.54.1 MaxGoMinor = %d, want 21", entries["v1.54.1"].MaxGoMinor)
	}
	if entries["v1.64.2"].MaxGoMinor != 24 {
		t.Fatalf("v1.64.2 MaxGoMinor = %d, want 24", entries["v1.64.2"].MaxGoMinor)
	}
	if entries["v2.4.0"].MaxGoMinor != 26 {
		t.Fatalf("v2.4.0 MaxGoMinor = %d, want 26", entries["v2.4.0"].MaxGoMinor)
	}
	if _, ok := entries["main"]; ok {
		t.Fatal("unexpected entry for non-semver tag")
	}
	if _, ok := entries["v2.5.0"]; ok {
		t.Fatal("unexpected entry for draft release")
	}
	if _, ok := entries["v2.6.0"]; ok {
		t.Fatal("unexpected entry for prerelease")
	}
}

func TestBuildManifestEntries_EmptyThresholds(t *testing.T) {
	t.Parallel()

	entries := buildManifestEntries([]release{{TagName: "v1.54.1"}}, nil, 26)
	if len(entries) != 0 {
		t.Fatalf("len(entries) = %d, want 0", len(entries))
	}
}

func TestParseSupportThreshold(t *testing.T) {
	t.Parallel()

	threshold, ok := parseSupportThreshold(issueItem{
		Title: "Go1.25 Support",
		Body:  "since v2.3.0 and later since [v2.4.0](https://example.test)",
	})
	if !ok {
		t.Fatal("expected support threshold to parse")
	}
	if threshold.GoMinor != 25 || threshold.SinceVersion != "v2.4.0" {
		t.Fatalf("threshold = %#v, want go1.25 -> v2.4.0", threshold)
	}

	if _, ok := parseSupportThreshold(issueItem{Title: "go1.25 support", Body: "not yet"}); ok {
		t.Fatal("expected issue without since version to be rejected")
	}
	if _, ok := parseSupportThreshold(issueItem{Title: "not a support issue", Body: "since v2.4.0"}); ok {
		t.Fatal("expected invalid title to be rejected")
	}
}

func TestGoMinorString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		want    string
		ok      bool
	}{
		{version: "go1.25.7", want: "25", ok: true},
		{version: "1.24.9", want: "24", ok: true},
		{version: "go1", want: "", ok: false},
		{version: "go1.x.0", want: "", ok: false},
	}

	for _, tt := range tests {
		got, ok := goMinorString(tt.version)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("goMinorString(%q) = (%q, %v), want (%q, %v)", tt.version, got, ok, tt.want, tt.ok)
		}
	}
}

func TestCompareGoVersion(t *testing.T) {
	t.Parallel()

	if compareGoVersion("2.0.0", "1.99.99") <= 0 {
		t.Fatal("expected 2.0.0 > 1.99.99")
	}
	if compareGoVersion("1.25.8", "1.25.7") <= 0 {
		t.Fatal("expected 1.25.8 > 1.25.7")
	}
	if compareGoVersion("1.26.0", "1.25.9") <= 0 {
		t.Fatal("expected 1.26.0 > 1.25.9")
	}
	if compareGoVersion("1.25.7", "1.25.7") != 0 {
		t.Fatal("expected equal versions to compare as zero")
	}
}

func TestGenerate_PropagatesFetchError(t *testing.T) {
	t.Parallel()

	generator := NewGenerator(GeneratorConfig{
		Client: newHTTPClient(func(req *http.Request) (*http.Response, error) {
			return newResponse(http.StatusInternalServerError, `[]`), nil
		}),
		BaseURL:       "https://example.test",
		GoReleasesURL: "https://example.test",
		Now: func() time.Time {
			return time.Unix(0, 0)
		},
	})

	_, err := generator.Generate(context.Background())
	if err == nil {
		t.Fatal("expected Generate to propagate fetch error")
	}
	if !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("error = %q, want unexpected status", err)
	}
}

func TestGenerate_PropagatesLaterFetchErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   roundTripFunc
		want string
	}{
		{
			name: "support threshold fetch",
			fn: func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/repos/golangci/golangci-lint/releases":
					return newResponse(http.StatusOK, `[]`), nil
				case "/search/issues":
					return newResponse(http.StatusBadGateway, `{}`), nil
				default:
					return newResponse(http.StatusOK, `[]`), nil
				}
			},
			want: "golangci-lint issues",
		},
		{
			name: "go releases fetch",
			fn: func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/repos/golangci/golangci-lint/releases":
					return newResponse(http.StatusOK, `[]`), nil
				case "/search/issues":
					return newResponse(http.StatusOK, `{"items":[]}`), nil
				case "/dl/":
					return newResponse(http.StatusBadGateway, `[]`), nil
				default:
					return newResponse(http.StatusOK, `[]`), nil
				}
			},
			want: "Go releases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(GeneratorConfig{
				Client:        newHTTPClient(tt.fn),
				BaseURL:       "https://example.test",
				GoReleasesURL: "https://example.test",
				Now:           func() time.Time { return time.Unix(0, 0) },
			})

			_, err := generator.Generate(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err, tt.want)
			}
		})
	}
}
