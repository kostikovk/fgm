package repocheck

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const versionGuardedMakefile = `GOBIN ?= $(shell go env GOPATH)/bin
GOLANGCI_LINT := $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION := 2.9.0

install-golangci-lint:
	@if ! test -x $(GOLANGCI_LINT) || ! $(GOLANGCI_LINT) version | grep -q "version $(GOLANGCI_LINT_VERSION)"; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(GOBIN) v$(GOLANGCI_LINT_VERSION); \
	fi

check-golangci: install-golangci-lint
	@$(GOLANGCI_LINT) run --timeout=2m
`

func TestServiceCheckToolBypassNoise(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		files    map[string]string
		wantC7   string
		wantC3   bool
		wantHint string
	}{
		{
			name:  "should not report C7 when the run target reinstalls the selected version whenever another version is found",
			files: map[string]string{"go.mod": goMod126, "Makefile": versionGuardedMakefile},
		},
		{
			name:   "should report only C3 when the version-guarded install pins another version than .fgm.toml",
			files:  map[string]string{"go.mod": goMod126, "Makefile": versionGuardedMakefile, ".fgm.toml": "[toolchain]\ngolangci_lint = \"v2.11.0\"\n"},
			wantC3: true,
		},
		{
			name: "should report one C7 with a hint when the install guard does not check the version",
			files: map[string]string{"go.mod": goMod126, "Makefile": "path := $(shell go env GOPATH)\n\n" +
				"install:\n\t@which $(path)/bin/golangci-lint > /dev/null || curl -sSfL https://example.com/install.sh | sh -s -- -b $(path)/bin v2.5.0\n\n" +
				"lint: install\n\t$(path)/bin/golangci-lint run\n"},
			wantC7:   "Makefile:7",
			wantHint: "the install target does not check the installed version",
		},
		{
			name:   "should report one finding when a run line mentions the binary twice",
			files:  map[string]string{"go.mod": goMod126, "Makefile": "lint:\n\t$(GOPATH)/bin/golangci-lint run ./a && $(GOPATH)/bin/golangci-lint run ./b\n"},
			wantC7: "Makefile:2",
		},
		{
			name:  "should not report probes that only test, locate or ask the version of the binary",
			files: map[string]string{"go.mod": goMod126, "Makefile": "probe:\n\ttest -x $(GOPATH)/bin/golangci-lint\n\tcommand -v $(GOPATH)/bin/golangci-lint\n\t$(GOPATH)/bin/golangci-lint --version\n\t[ -f $(GOPATH)/bin/golangci-lint ] && echo ok\n"},
		},
		{
			name:   "should follow prerequisites without looping when targets depend on each other",
			files:  map[string]string{"go.mod": goMod126, "Makefile": "a: b\n\t$(GOPATH)/bin/golangci-lint run\nb: a\n\techo b\n"},
			wantC7: "Makefile:2",
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
			var c7 []string
			hasC3 := false
			message := ""
			for _, finding := range result.Findings {
				switch finding.ID {
				case "C7":
					c7 = append(c7, fmt.Sprintf("%s:%d", finding.File, finding.Line))
					message = finding.Message
				case "C3":
					hasC3 = true
				}
			}
			if strings.Join(c7, "; ") != tc.wantC7 || hasC3 != tc.wantC3 {
				t.Errorf("C7 = %q, C3 = %v; want %q, %v\n%+v", strings.Join(c7, "; "), hasC3, tc.wantC7, tc.wantC3, result.Findings)
			}
			if tc.wantHint != "" && !strings.Contains(message, tc.wantHint) {
				t.Errorf("message = %q; want the hint %q", message, tc.wantHint)
			}
		})
	}
}
