package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newEnvCmd(application *app.App) *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Print shell environment setup for FGM",
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
