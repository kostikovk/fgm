package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newExecCmd(application *app.App, v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "exec -- [command] [args...]",
		Short: "Run a command with the resolved Go toolchain on PATH",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command provided")
			}
			if application.Executor == nil {
				return fmt.Errorf("executor is not configured")
			}

			return application.Executor.Exec(
				cmd.Context(),
				v.GetString(flagChdir),
				args,
				cmd.InOrStdin(),
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}
}
