# AAA in Go

Use the package's existing `_test.go` location and package convention. Prefer standard `testing` when no assertion library is already established.

## Contents

1. [A single behavior](#a-single-behavior)
2. [Table-driven subtests](#table-driven-subtests)
3. [Go-specific decisions](#go-specific-decisions)
4. [Validation](#validation)

## A single behavior

This complete test illustrates the three phases using a standard-library function. In a project, invoke the actual behavior under test.

```go
package example

import (
	"strings"
	"testing"
)

func TestTrimSpace_RemovesSurroundingWhitespace(t *testing.T) {
	t.Run("should remove surrounding whitespace when input has spaces and line breaks", func(t *testing.T) {
		// Arrange
		input := " \tinvoice\n"
		want := "invoice"

		// Act
		got := strings.TrimSpace(input)

		// Assert
		if got != want {
			t.Errorf("TrimSpace(%q) = %q; want %q", input, got, want)
		}
	})
}
```

## Table-driven subtests

Define scenario data outside the loop, then arrange fresh fixtures inside each `t.Run` when the subject or dependencies are mutable. Share only immutable scenario data. Do not add `t.Parallel()` unless all accessed state and dependencies support concurrent use. Honor the repository's Go version when selecting helpers and language features.

This complete example checks a value on success and the presence of an error on failure. If a project's contract promises an error identity or type, assert that as well.

```go
package example

import (
	"strconv"
	"testing"
)

func TestAtoi(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "should return 42 when input is a positive integer string", input: "42", want: 42},
		{name: "should return -7 when input is a negative integer string", input: "-7", want: -7},
		{name: "should return zero when input is the string zero", input: "0", want: 0},
		{name: "should return an error when input is not an integer string", input: "forty-two", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input, want, wantErr := tc.input, tc.want, tc.wantErr

			// Act
			got, err := strconv.Atoi(input)

			// Assert
			if wantErr {
				if err == nil {
					t.Fatalf("Atoi(%q) returned no error; want an error", input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Atoi(%q) returned an unexpected error: %v", input, err)
			}
			if got != want {
				t.Errorf("Atoi(%q) = %d; want %d", input, got, want)
			}
		})
	}
}
```

## Go-specific decisions

- Call `t.Helper()` in testing helpers so failures point to the caller. Prefer `t.Fatalf` for failed fixture setup or a prerequisite that makes later checks unsafe; use `t.Errorf` for independent mismatches. `Fatalf` stops the current test goroutine, so do not call it from a spawned goroutine; return its result to the test goroutine instead. See [testing.T](https://pkg.go.dev/testing#T).
- Register release work with `t.Cleanup` when allocating resources, or use `t.TempDir` for temporary filesystem fixtures. Cleanup belongs to resource setup, not to the behavior asserted. See [Cleanup](https://pkg.go.dev/testing#T.Cleanup) and [TempDir](https://pkg.go.dev/testing#T.TempDir).
- Test successful calls with `err == nil`. For expected sentinel errors that may be wrapped, use `errors.Is(err, wantErr)` with the existing sentinel; for an error type, use `errors.As`. Avoid error-string comparisons unless the exact message is the contract. See [errors](https://pkg.go.dev/errors).
- Mock responses and expectations are arranged before the action; verify contractual interactions afterward. Keep cleanup-backed mock verification if the existing framework requires it. Reuse the project's chosen tooling rather than selecting a mocking library as part of an AAA refactor.

## Validation

Adapt these commands to the module, test names, and repository instructions:

```sh
gofmt -w path/to/changed_test.go
go test ./path/to/package -run '^TestBehavior$' -count=1
go test ./path/to/package
go vet ./path/to/package
```

Use `go test -race` when relevant to changed concurrent behavior and supported by the environment. Expanding to `go test ./...` depends on project scope and available integration services.
