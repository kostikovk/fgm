package repocheck

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestServiceCheckToolBypass(t *testing.T) {
	t.Parallel()

	makefile := func(recipe string) map[string]string {
		return map[string]string{"go.mod": goMod126, "Makefile": "lint:\n\t" + recipe + "\n"}
	}

	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{name: "should report C7 when a recipe runs golangci-lint from $(shell go env GOPATH)/bin", files: makefile("$(shell go env GOPATH)/bin/golangci-lint run ./..."), want: "C7 WARN Makefile:2"},
		{name: "should report C7 when a recipe runs golangci-lint from $(GOPATH)/bin", files: makefile("$(GOPATH)/bin/golangci-lint run"), want: "C7 WARN Makefile:2"},
		{name: "should report C7 when a recipe runs golangci-lint from $$(go env GOPATH)/bin", files: makefile("$$(go env GOPATH)/bin/golangci-lint run"), want: "C7 WARN Makefile:2"},
		{name: "should report C7 when a recipe runs golangci-lint from $(GOBIN)", files: makefile("$(GOBIN)/golangci-lint run"), want: "C7 WARN Makefile:2"},
		{name: "should report C7 when a recipe runs golangci-lint from ~/go/bin", files: makefile("~/go/bin/golangci-lint run"), want: "C7 WARN Makefile:2"},
		{name: "should report C7 when a recipe runs golangci-lint from $HOME/go/bin", files: makefile("$$HOME/go/bin/golangci-lint run"), want: "C7 WARN Makefile:2"},
		{name: "should report C7 when a recipe runs go by an absolute path", files: makefile("/usr/local/go/bin/go build ./..."), want: "C7 WARN Makefile:2"},
		{name: "should not report C7 when a recipe runs golangci-lint through PATH", files: makefile("golangci-lint run ./...")},
		{name: "should not report C7 when a recipe runs a project-local binary", files: makefile("./bin/golangci-lint run && $(LOCALBIN)/golangci-lint run && $(CURDIR)/bin/golangci-lint run")},
		{name: "should not report C7 when a line only installs the linter", files: makefile("curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(GOPATH)/bin v2.5.0")},
		{name: "should not report C7 when the invocation is commented out", files: makefile("# $(GOPATH)/bin/golangci-lint run")},
		{
			name:  "should report C7 when a workflow run step runs the linter from GOPATH",
			files: map[string]string{"go.mod": goMod126, ".github/workflows/ci.yml": "jobs:\n  lint:\n    steps:\n      - run: |\n          go version\n          $(go env GOPATH)/bin/golangci-lint run\n"},
			want:  "C7 WARN .github/workflows/ci.yml:6",
		},
		{
			name: "should not report C7 when the workflow step works in another module",
			files: map[string]string{
				"a/go.mod":                 goMod126,
				"b/go.mod":                 goMod126,
				".github/workflows/ci.yml": "jobs:\n  lint:\n    defaults:\n      run:\n        working-directory: b\n    steps:\n      - run: $(go env GOPATH)/bin/golangci-lint run\n",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeRepo(t, tc.files)
			dir := root
			if _, ok := tc.files["a/go.mod"]; ok {
				dir = root + "/a"
			}

			// Act
			result, err := newService().Check(context.Background(), dir)

			// Assert
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			var got []string
			for _, finding := range result.Findings {
				if finding.ID == "C7" {
					got = append(got, fmt.Sprintf("%s %s %s:%d", finding.ID, finding.Severity, finding.File, finding.Line))
				}
			}
			if strings.Join(got, "; ") != tc.want {
				t.Errorf("C7 findings = %q; want %q\n%+v", strings.Join(got, "; "), tc.want, result.Findings)
			}
		})
	}
}
