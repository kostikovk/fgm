package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCurrentCmd(application *app.App, v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the resolved toolchain for the current workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			selection, err := application.Resolver.Current(cmd.Context(), workDir)
			if err != nil {
				return err
			}

			ui := uiFromCmd(cmd)
			colWidth := len("go")
			if selection.LintVersion != "" {
				colWidth = len("golangci-lint")
			}

			goLine := ui.FormatCurrentLine("go", selection.GoVersion, goSourceLabel(selection), colWidth)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), goLine); err != nil {
				return err
			}
			if selection.LintVersion != "" {
				lintLine := ui.FormatCurrentLine("golangci-lint", selection.LintVersion, lintSourceLabel(workDir, selection), colWidth)
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), lintLine); err != nil {
					return err
				}
			}
			for _, warning := range selection.Warnings {
				if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "warning: "+warning); err != nil {
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

func goSourceLabel(selection app.Selection) string {
	if selection.GoSource == "latest-patch" {
		return fmt.Sprintf("latest-patch of %s %s", filepath.Base(selection.GoSourceFile), selection.GoBase)
	}
	return sourceLabel(selection.GoSource)
}

func lintSourceLabel(workDir string, selection app.Selection) string {
	if selection.LintSource == "detected" {
		return fmt.Sprintf("detected: %s:%d", currenttoolchain.DisplayPath(workDir, selection.LintSourceFile), selection.LintSourceLine)
	}
	return sourceLabel(selection.LintSource)
}

func sourceLabel(source string) string {
	switch source {
	case "":
		return ""
	case "global", "config", "detected", "local", "remote", "latest-patch":
		return source
	default:
		return filepath.Base(source)
	}
}
