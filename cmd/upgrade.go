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
	var dryRun bool
	var version string

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

			options := app.GoUpgradeOptions{
				WorkDir: v.GetString(flagChdir),
				Version: version,
				DryRun:  dryRun,
			}

			if global {
				result, err := application.GoUpgrader.UpgradeGlobal(cmd.Context(), options)
				if err != nil {
					return err
				}
				if result.DryRun {
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "Would upgrade global Go to %s\n", result.Version)
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Upgraded global Go to %s\n", result.Version)
				return err
			}

			result, err := application.GoUpgrader.UpgradeProject(cmd.Context(), options)
			if err != nil {
				return err
			}
			if result.DryRun {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Would upgrade project Go to %s in %s\n", result.Version, result.Path)
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Upgraded project Go to %s in %s\n", result.Version, result.Path)
			return err
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "upgrade the global default Go version")
	cmd.Flags().BoolVar(&project, "project", false, "upgrade the nearest project Go version")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview the selected Go upgrade without changing anything")
	cmd.Flags().StringVar(&version, "to", "", "upgrade to a specific Go version instead of the latest remote version")

	return cmd
}
