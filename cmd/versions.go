package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newVersionsCmd(application *app.App, v *viper.Viper) *cobra.Command {
	versionsCmd := &cobra.Command{
		Use:   "versions",
		Short: "List available toolchain versions",
	}

	versionsCmd.AddCommand(newVersionsGoCmd(application, v))
	versionsCmd.AddCommand(newVersionsLintCmd(application, v))

	return versionsCmd
}

func newVersionsGoCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var local bool
	var remote bool

	cmd := &cobra.Command{
		Use:   "go",
		Short: "List Go toolchain versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if local && remote {
				return fmt.Errorf("--local and --remote are mutually exclusive")
			}
			if !local && !remote {
				return fmt.Errorf("provide --local or --remote")
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
				selection, err := application.Resolver.Current(cmd.Context(), v.GetString(flagChdir))
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

func newVersionsLintCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var local bool
	var remote bool
	var goVersion string

	cmd := &cobra.Command{
		Use:   "golangci-lint",
		Short: "List golangci-lint versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if local && remote {
				return fmt.Errorf("--local and --remote are mutually exclusive")
			}
			if !local && !remote {
				return fmt.Errorf("provide --local or --remote")
			}

			if local {
				if application.LintStore == nil {
					return fmt.Errorf("local golangci-lint version store is not configured")
				}

				versions, err := application.LintStore.ListLocalLintVersions(cmd.Context())
				if err != nil {
					return err
				}
				for _, version := range versions {
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), version); err != nil {
						return err
					}
				}
				return nil
			}
			if application.LintRemoteProvider == nil {
				return fmt.Errorf("remote golangci-lint version provider is not configured")
			}

			targetGoVersion := goVersion
			if targetGoVersion == "" && application.Resolver != nil {
				selection, err := application.Resolver.Current(cmd.Context(), v.GetString(flagChdir))
				if err == nil {
					targetGoVersion = selection.GoVersion
				}
			}
			if targetGoVersion == "" {
				return fmt.Errorf("provide --go or run inside a repo with Go toolchain metadata")
			}

			versions, err := application.LintRemoteProvider.ListRemoteLintVersions(cmd.Context(), targetGoVersion)
			if err != nil {
				return err
			}

			for _, version := range versions {
				line := version.Version
				if version.Recommended {
					line = "* " + line
				}
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&local, "local", false, "list locally installed golangci-lint versions")
	cmd.Flags().BoolVar(&remote, "remote", false, "list remotely available golangci-lint versions")
	cmd.Flags().StringVar(&goVersion, "go", "", "target Go version for compatibility filtering")

	return cmd
}
