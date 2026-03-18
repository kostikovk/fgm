package cmd

import (
	"fmt"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newUseCmd(application *app.App) *cobra.Command {
	useCmd := &cobra.Command{
		Use:   "use",
		Short: "Select active toolchain versions",
	}

	useCmd.AddCommand(newUseGoCmd(application))

	return useCmd
}

func newUseGoCmd(application *app.App) *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:   "go [version]",
		Short: "Select the active Go version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !global {
				return fmt.Errorf("only --global is supported for now")
			}
			if application.GoStore == nil {
				return fmt.Errorf("local Go version store is not configured")
			}

			version := args[0]
			ok, err := application.GoStore.HasGoVersion(cmd.Context(), version)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("go version %s is not installed", version)
			}

			if err := application.GoStore.SetGlobalGoVersion(cmd.Context(), version); err != nil {
				return err
			}
			if err := application.GoStore.EnsureShims(); err != nil {
				return err
			}

			_, err = fmt.Fprintf(
				cmd.OutOrStdout(),
				"Selected Go %s as the global default.\nAdd %s to PATH ahead of other Go installations.\n",
				version,
				application.GoStore.ShimDir(),
			)
			return err
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "set the global Go version")

	return cmd
}
