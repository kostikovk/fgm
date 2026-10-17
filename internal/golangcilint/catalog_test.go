package golangcilint

import "testing"

func TestProviderCatalog(t *testing.T) {
	t.Parallel()

	data := []byte(`{
  "latest_go_versions": {"19": "1.19.13", "26": "1.26.9"},
  "versions": {
    "v1.64.8": {"min_go_minor": 22, "max_go_minor": 24},
    "v2.9.0": {"min_go_minor": 24, "max_go_minor": 26}
  }
}`)

	t.Run("LatestGoPatch", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name      string
			goVersion string
			want      string
			wantOK    bool
		}{
			{name: "should return the newest catalog patch when the minor is known", goVersion: "1.26.0", want: "1.26.9", wantOK: true},
			{name: "should return the newest catalog patch when the version has no patch component", goVersion: "go1.19", want: "1.19.13", wantOK: true},
			{name: "should report not found when the minor is missing from the catalog", goVersion: "1.28.0"},
			{name: "should return the newest catalog patch of the minor when the version is a release candidate", goVersion: "1.26rc1", want: "1.26.9", wantOK: true},
			{name: "should return the newest catalog patch of the minor when the version is a go-prefixed beta", goVersion: "go1.26beta1", want: "1.26.9", wantOK: true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Arrange
				provider := New(Config{CompatibilityData: data})

				// Act
				got, ok, err := provider.LatestGoPatch(tc.goVersion)

				// Assert
				if err != nil {
					t.Fatalf("LatestGoPatch: %v", err)
				}
				if got != tc.want || ok != tc.wantOK {
					t.Errorf("LatestGoPatch(%q) = %q, %v; want %q, %v", tc.goVersion, got, ok, tc.want, tc.wantOK)
				}
			})
		}
	})

	t.Run("LintCompatibility", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name           string
			goVersion      string
			lintVersion    string
			wantCompatible bool
			wantKnown      bool
		}{
			{name: "should report compatible when the Go minor is inside the linter range", goVersion: "1.26.3", lintVersion: "v2.9.0", wantCompatible: true, wantKnown: true},
			{name: "should report incompatible when the Go minor is above the linter range", goVersion: "1.26.3", lintVersion: "v1.64.8", wantKnown: true},
			{name: "should accept a version without the v prefix", goVersion: "1.25.0", lintVersion: "2.9.0", wantCompatible: true, wantKnown: true},
			{name: "should report unknown when the linter is missing from the catalog", goVersion: "1.26.3", lintVersion: "v2.99.0"},
			{name: "should judge a release candidate by its minor when the Go version is a prerelease", goVersion: "1.26rc2", lintVersion: "v2.9.0", wantCompatible: true, wantKnown: true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Arrange
				provider := New(Config{CompatibilityData: data})

				// Act
				compatible, known, err := provider.LintCompatibility(tc.goVersion, tc.lintVersion)

				// Assert
				if err != nil {
					t.Fatalf("LintCompatibility: %v", err)
				}
				if compatible != tc.wantCompatible || known != tc.wantKnown {
					t.Errorf("LintCompatibility = %v, %v; want %v, %v", compatible, known, tc.wantCompatible, tc.wantKnown)
				}
			})
		}
	})
}
