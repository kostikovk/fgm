package cmd

import (
	"fmt"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newEnvCmd(application *app.App) *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Print shell environment setup for FGM",
		Long:  "Print shell setup commands that prepend FGM shims to PATH in an interactive shell. In non-interactive sessions, use 'fgm exec -- <command>' for each invocation because environment changes do not persist across separate child processes.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.EnvRenderer == nil {
				return fmt.Errorf("env renderer is not configured")
			}

			lines, err := application.EnvRenderer.Render(cmd.Context(), shell)
			if err != nil {
				return err
			}
			for _, line := range lines {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "override detected shell (zsh, bash, fish, powershell)")

	return cmd
}
