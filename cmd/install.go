package cmd

import (
	"fmt"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newInstallCmd(application *app.App, v *viper.Viper) *cobra.Command {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install toolchains into the local FGM store",
		Long:  "Install the workspace's resolved Go version (and resolved golangci-lint version, when applicable) into the local FGM store without changing the global default Go selection.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.Resolver == nil {
				return fmt.Errorf("resolver is not configured")
			}
			if application.GoInstaller == nil {
				return fmt.Errorf("go installer is not configured")
			}
			if application.LintInstaller == nil {
				return fmt.Errorf("golangci-lint installer is not configured")
			}

			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			selection, err := application.Resolver.Current(cmd.Context(), workDir)
			if err != nil {
				return err
			}

			goInstallPath, err := application.GoInstaller.InstallGoVersion(cmd.Context(), selection.GoVersion)
			if err != nil {
				return err
			}
			ui := uiFromCmd(cmd)
			goMsg := ui.FormatOutcome(
				fmt.Sprintf("Installed Go %s to %s", selection.GoVersion, goInstallPath),
				"Installed",
				"Go",
				selection.GoVersion,
				"to "+goInstallPath,
			)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), goMsg); err != nil {
				return err
			}

			if selection.LintVersion != "" {
				lintInstallPath, err := application.LintInstaller.InstallLintVersion(cmd.Context(), selection.LintVersion)
				if err != nil {
					return err
				}
				lintMsg := ui.FormatOutcome(
					fmt.Sprintf("Installed golangci-lint %s to %s", selection.LintVersion, lintInstallPath),
					"Installed",
					"golangci-lint",
					selection.LintVersion,
					"to "+lintInstallPath,
				)
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), lintMsg); err != nil {
					return err
				}
			}

			refreshUpdateCache(cmd.Context(), application.FGMUpdateChecker)
			return ensureShimsIfSupported(application.GoStore)
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
				return fmt.Errorf("go installer is not configured")
			}

			version := args[0]
			installPath, err := application.GoInstaller.InstallGoVersion(cmd.Context(), version)
			if err != nil {
				return err
			}

			ui := uiFromCmd(cmd)
			msg := ui.FormatOutcome(
				fmt.Sprintf("Installed Go %s to %s", version, installPath),
				"Installed",
				"Go",
				version,
				"to "+installPath,
			)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
				return err
			}
			refreshUpdateCache(cmd.Context(), application.FGMUpdateChecker)
			return ensureShimsIfSupported(application.GoStore)
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

			ui := uiFromCmd(cmd)
			msg := ui.FormatOutcome(
				fmt.Sprintf("Installed golangci-lint %s to %s", version, installPath),
				"Installed",
				"golangci-lint",
				version,
				"to "+installPath,
			)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
				return err
			}
			refreshUpdateCache(cmd.Context(), application.FGMUpdateChecker)
			return ensureShimsIfSupported(application.GoStore)
		},
	}
}
