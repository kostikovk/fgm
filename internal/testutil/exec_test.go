package testutil

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecuteCommand(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	root := &cobra.Command{
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if got, want := strings.Join(args, ","), "alpha,beta"; got != want {
				t.Fatalf("args = %q, want %q", got, want)
			}
			if _, err := cmd.OutOrStdout().Write([]byte("stdout")); err != nil {
				t.Fatalf("write stdout: %v", err)
			}
			if _, err := cmd.ErrOrStderr().Write([]byte("stderr")); err != nil {
				t.Fatalf("write stderr: %v", err)
			}
			return expectedErr
		},
	}

	stdout, stderr, err := ExecuteCommand(t, root, "alpha", "beta")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("err = %v, want %v", err, expectedErr)
	}
	if stdout != "stdout" {
		t.Fatalf("stdout = %q, want %q", stdout, "stdout")
	}
	if !strings.Contains(stderr, "stderr") {
		t.Fatalf("stderr = %q, want it to contain %q", stderr, "stderr")
	}
}
