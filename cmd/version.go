package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newVersionCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show FGM build information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "fgm %s\n", fallbackValue(application.BuildInfo.Version, "dev")); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "commit %s\n", fallbackValue(application.BuildInfo.Commit, "unknown")); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "date %s\n", fallbackValue(application.BuildInfo.Date, "unknown")); err != nil {
				return err
			}
			return nil
		},
	}
}

func fallbackValue(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
