package cmd

import (
	"fmt"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newExecCmd(application *app.App, v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "exec -- [command] [args...]",
		Short: "Run a command with the resolved Go and installed selected golangci-lint on PATH",
		Long:  "Run a child process with the workspace's resolved Go binary directory (and installed selected golangci-lint directory, when present) prepended to PATH. Use this in non-interactive sessions where shell profile changes do not persist across processes.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command provided")
			}
			if application.Executor == nil {
				return fmt.Errorf("executor is not configured")
			}

			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			return application.Executor.Exec(
				cmd.Context(),
				workDir,
				args,
				cmd.InOrStdin(),
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}
}
