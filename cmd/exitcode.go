package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// ExitError carries a process exit code. A silent ExitError means the command already
// reported its outcome on stdout/stderr, so no error line is printed.
type ExitError struct {
	Code   int
	Err    error
	Silent bool
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit status %d", e.Code)
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// ExitCode returns the process exit code for an error returned by the root command:
// 0 for nil, the ExitError code when present, 1 otherwise.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := errors.AsType[*ExitError](err); ok {
		return exitErr.Code
	}
	return 1
}

// usageError marks an invalid invocation (exit status 2).
func usageError(err error) error {
	return &ExitError{Code: 2, Err: err}
}

// usageArgs wraps a positional-argument validator so its failures exit with status 2.
func usageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return usageError(err)
		}
		return nil
	}
}

func isSilent(err error) bool {
	exitErr, ok := errors.AsType[*ExitError](err)
	return ok && exitErr.Silent
}
