package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kostikovk/fgm/internal/app"
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

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), formatCurrentLine("go", selection.GoVersion, selection.GoSource)); err != nil {
				return err
			}
			if selection.LintVersion != "" {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), formatCurrentLine("golangci-lint", selection.LintVersion, selection.LintSource)); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func formatCurrentLine(tool string, version string, source string) string {
	line := fmt.Sprintf("%s %s", tool, version)
	if label := sourceLabel(source); label != "" {
		line += " (" + label + ")"
	}
	return line
}

func sourceLabel(source string) string {
	switch source {
	case "":
		return ""
	case "global", "config", "local", "remote":
		return source
	default:
		return filepath.Base(source)
	}
}
