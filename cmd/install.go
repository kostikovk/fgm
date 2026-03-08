package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newInstallCmd(application *app.App) *cobra.Command {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install toolchains into the local FGM store",
	}

	installCmd.AddCommand(newInstallGoCmd(application))

	return installCmd
}

func newInstallGoCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "go [version]",
		Short: "Install a Go version into the local FGM store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.GoInstaller == nil {
				return fmt.Errorf("Go installer is not configured")
			}

			version := args[0]
			installPath, err := application.GoInstaller.InstallGoVersion(cmd.Context(), version)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Installed Go %s to %s\n", version, installPath)
			return err
		},
	}
}
