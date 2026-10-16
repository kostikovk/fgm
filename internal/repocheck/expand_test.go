package repocheck

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestServiceCheckToolBypassVariables(t *testing.T) {
	t.Parallel()

	withMakefile := func(files map[string]string) map[string]string {
		files["go.mod"] = goMod126
		return files
	}

	cases := []struct {
		name        string
		files       map[string]string
		want        string
		wantMessage string
	}{
		{
			name: "should report the run line when the GOPATH path is stored in a variable",
			files: withMakefile(map[string]string{"Makefile": "path := $(shell go env GOPATH)\n\n" +
				"install-golangci-lint:\n\t@which $(path)/bin/golangci-lint > /dev/null || curl -sSfL https://example.com/install.sh | sh -s -- -b $(path)/bin v2.5.0\n\n" +
				"check-golangci: install-golangci-lint\n\t@$(path)/bin/golangci-lint run --timeout=5m\n"}),
			want:        "C7 WARN Makefile:7",
			wantMessage: "golangci-lint is invoked as $(path)/bin/golangci-lint (expands to $(shell go env GOPATH)/bin/golangci-lint)",
		},
		{
			name:  "should report a GOBIN variable assigned from go env GOPATH",
			files: withMakefile(map[string]string{"Makefile": "GOBIN ?= $(shell go env GOPATH)/bin\n\nlint:\n\t$(GOBIN)/golangci-lint run\n"}),
			want:  "C7 WARN Makefile:4",
		},
		{
			name:  "should report a chain of variables when they resolve to GOPATH/bin",
			files: withMakefile(map[string]string{"Makefile": "A := $(shell go env GOPATH)\nB := $(A)/bin\n\nlint:\n\t$(B)/golangci-lint run\n"}),
			want:  "C7 WARN Makefile:5",
		},
		{
			name:  "should report a variable referenced with braces",
			files: withMakefile(map[string]string{"Makefile": "TOOLS = $(shell go env GOPATH)/bin\n\nlint:\n\t${TOOLS}/golangci-lint run\n"}),
			want:  "C7 WARN Makefile:4",
		},
		{
			name:  "should report a variable defined in an included makefile",
			files: withMakefile(map[string]string{"Makefile": "include build/tools.mk\n\nlint:\n\t$(BIN)/golangci-lint run\n", "build/tools.mk": "BIN := $(shell go env GOPATH)/bin\n"}),
			want:  "C7 WARN Makefile:4",
		},
		{
			name:  "should not report a project-local bin directory reached through a variable",
			files: withMakefile(map[string]string{"Makefile": "LOCALBIN := $(shell pwd)/bin\nGOBIN := $(CURDIR)/bin\n\nlint:\n\t$(LOCALBIN)/golangci-lint run\n\t$(GOBIN)/golangci-lint run\n"}),
		},
		{
			name:  "should not report a variable that is only set in the environment",
			files: withMakefile(map[string]string{"Makefile": "lint:\n\t$(TOOLS_DIR)/golangci-lint run\n"}),
		},
		{
			name:  "should not report and should terminate when variables reference each other",
			files: withMakefile(map[string]string{"Makefile": "A = $(B)\nB = $(A)\n\nlint:\n\t$(A)/bin/golangci-lint run\n"}),
		},
		{
			name:  "should not report a variable that uses an unknown function",
			files: withMakefile(map[string]string{"Makefile": "LINT := $(shell which golangci-lint)\n\nlint:\n\t$(LINT) run\n"}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			root := writeRepo(t, tc.files)

			// Act
			result, err := newService().Check(context.Background(), root)

			// Assert
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			var got []string
			message := ""
			for _, finding := range result.Findings {
				if finding.ID == "C7" {
					got = append(got, fmt.Sprintf("%s %s %s:%d", finding.ID, finding.Severity, finding.File, finding.Line))
					message = finding.Message
				}
			}
			if strings.Join(got, "; ") != tc.want {
				t.Errorf("C7 findings = %q; want %q\n%+v", strings.Join(got, "; "), tc.want, result.Findings)
			}
			if tc.wantMessage != "" && !strings.Contains(message, tc.wantMessage) {
				t.Errorf("message = %q; want it to contain %q", message, tc.wantMessage)
			}
		})
	}
}
