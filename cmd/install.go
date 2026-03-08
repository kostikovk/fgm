package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newInstallCmd(application *app.App, v *viper.Viper) *cobra.Command {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install toolchains into the local FGM store",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.Resolver == nil {
				return fmt.Errorf("resolver is not configured")
			}
			if application.GoInstaller == nil {
				return fmt.Errorf("Go installer is not configured")
			}
			if application.LintInstaller == nil {
				return fmt.Errorf("golangci-lint installer is not configured")
			}

			selection, err := application.Resolver.Current(cmd.Context(), v.GetString(flagChdir))
			if err != nil {
				return err
			}

			goInstallPath, err := application.GoInstaller.InstallGoVersion(cmd.Context(), selection.GoVersion)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Installed Go %s to %s\n", selection.GoVersion, goInstallPath); err != nil {
				return err
			}

			if selection.LintVersion == "" {
				return nil
			}

			lintInstallPath, err := application.LintInstaller.InstallLintVersion(cmd.Context(), selection.LintVersion)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(
				cmd.OutOrStdout(),
				"Installed golangci-lint %s to %s\n",
				selection.LintVersion,
				lintInstallPath,
			)
			return err
		},
	}

	installCmd.AddCommand(newInstallGoCmd(application))
	installCmd.AddCommand(newInstallLintCmd(application))

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

func newInstallLintCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "golangci-lint [version]",
		Short: "Install a golangci-lint version into the local FGM store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.LintInstaller == nil {
				return fmt.Errorf("golangci-lint installer is not configured")
			}

			version := args[0]
			installPath, err := application.LintInstaller.InstallLintVersion(cmd.Context(), version)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Installed golangci-lint %s to %s\n", version, installPath)
			return err
		},
	}
}
