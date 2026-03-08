package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newVersionsCmd(application *app.App) *cobra.Command {
	versionsCmd := &cobra.Command{
		Use:   "versions",
		Short: "List available toolchain versions",
	}

	versionsCmd.AddCommand(newVersionsGoCmd(application))

	return versionsCmd
}

func newVersionsGoCmd(application *app.App) *cobra.Command {
	var local bool
	var remote bool

	cmd := &cobra.Command{
		Use:   "go",
		Short: "List Go toolchain versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if local == remote {
				return fmt.Errorf("choose --local or --remote")
			}

			var versions []string
			var err error

			if local {
				if application.GoStore == nil {
					return fmt.Errorf("local Go version store is not configured")
				}
				versions, err = application.GoStore.ListLocalGoVersions(cmd.Context())
			} else {
				if application.GoRemoteProvider == nil {
					return fmt.Errorf("remote Go version provider is not configured")
				}
				versions, err = application.GoRemoteProvider.ListRemoteGoVersions(cmd.Context())
			}
			if err != nil {
				return err
			}

			currentVersion := ""
			if application.Resolver != nil {
				workDir, err := cmd.Flags().GetString(flagChdir)
				if err != nil {
					return err
				}
				selection, err := application.Resolver.Current(cmd.Context(), workDir)
				if err == nil {
					currentVersion = selection.GoVersion
				}
			}

			for _, version := range versions {
				line := version
				if version == currentVersion {
					line = "* " + version
				}
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&local, "local", false, "list locally installed Go versions")
	cmd.Flags().BoolVar(&remote, "remote", false, "list remotely available Go versions")

	return cmd
}
