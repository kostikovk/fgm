package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newUpgradeCmd(application *app.App, v *viper.Viper) *cobra.Command {
	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade toolchains to newer versions",
	}

	upgradeCmd.AddCommand(newUpgradeGoCmd(application, v))

	return upgradeCmd
}

func newUpgradeGoCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var global bool
	var project bool

	cmd := &cobra.Command{
		Use:   "go",
		Short: "Upgrade Go globally or for the current project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if global == project {
				return fmt.Errorf("choose --global or --project")
			}
			if application.GoUpgrader == nil {
				return fmt.Errorf("Go upgrader is not configured")
			}

			if global {
				result, err := application.GoUpgrader.UpgradeGlobal(cmd.Context())
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Upgraded global Go to %s\n", result.Version)
				return err
			}

			result, err := application.GoUpgrader.UpgradeProject(cmd.Context(), v.GetString(flagChdir))
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Upgraded project Go to %s in %s\n", result.Version, result.Path)
			return err
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "upgrade the global default Go version")
	cmd.Flags().BoolVar(&project, "project", false, "upgrade the nearest project Go version")

	return cmd
}
