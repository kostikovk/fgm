package testutil

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

// ExecuteCommand runs a Cobra command and returns captured output.
func ExecuteCommand(t *testing.T, root *cobra.Command, args ...string) (string, string, error) {
	t.Helper()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)

	err := root.Execute()
	return stdout.String(), stderr.String(), err
}
