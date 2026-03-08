package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCurrentCmd(application *app.App, v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the resolved toolchain for the current workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := application.Resolver.Current(cmd.Context(), v.GetString(flagChdir))
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "go %s\n", selection.GoVersion); err != nil {
				return err
			}
			if selection.LintVersion != "" {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "golangci-lint %s\n", selection.LintVersion); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
