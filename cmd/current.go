package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCurrentCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var asJSON bool
	var all bool

	cmd := &cobra.Command{
		Use:   "current [--json] [--all]",
		Short: "Show the resolved toolchain for the current workspace",
		Long:  "Show the resolved Go and golangci-lint versions with their sources. --json prints a versioned machine-readable document (schema 1, never styled). --all reports every module under the working directory: the use entries of a go.work, otherwise every go.mod below it (skipping vendor, testdata, node_modules and dot-directories). Exit status: 0 when every module resolved, 1 when at least one failed, 2 for an invalid invocation.",
		Args:  usageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}
			if !asJSON && !all {
				return printCurrentText(cmd, application, workDir)
			}
			if application.CurrentReporter == nil {
				return fmt.Errorf("current reporter is not configured")
			}

			if !all {
				report, err := application.CurrentReporter.Report(cmd.Context(), workDir)
				if err != nil {
					return err
				}
				return writeJSON(cmd.OutOrStdout(), report)
			}

			reports, err := application.CurrentReporter.ReportAll(cmd.Context(), workDir)
			if err != nil {
				return err
			}
			if asJSON {
				err = writeJSON(cmd.OutOrStdout(), reports)
			} else {
				err = printReportLines(cmd.OutOrStdout(), workDir, reports)
			}
			if err != nil {
				return err
			}
			for _, report := range reports {
				if report.Error != "" {
					return &ExitError{Code: 1, Err: fmt.Errorf("at least one module failed to resolve"), Silent: true}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "print machine-readable JSON (schema 1)")
	cmd.Flags().BoolVar(&all, "all", false, "report every module under the working directory")
	return cmd
}

func printCurrentText(cmd *cobra.Command, application *app.App, workDir string) error {
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
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func printReportLines(w io.Writer, workDir string, reports []app.CurrentReport) error {
	for _, report := range reports {
		parts := []string{currenttoolchain.DisplayPath(workDir, report.Dir)}
		switch {
		case report.Error != "":
			parts = append(parts, "error: "+report.Error)
		default:
			parts = append(parts, formatLabeledLine("go", report.Go.Version, goReportLabel(report.Go)))
			if report.GolangCILint != nil {
				parts = append(parts, formatLabeledLine("golangci-lint", report.GolangCILint.Version, lintReportLabel(report.GolangCILint)))
			}
			for _, warning := range report.Warnings {
				parts = append(parts, "warning: "+warning)
			}
		}
		if _, err := fmt.Fprintln(w, strings.Join(parts, "  ")); err != nil {
			return err
		}
	}
	return nil
}

func formatCurrentLine(tool string, version string, source string) string {
	return formatLabeledLine(tool, version, sourceLabel(source))
}

func formatLabeledLine(tool string, version string, label string) string {
	line := fmt.Sprintf("%s %s", tool, version)
	if label != "" {
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

func goReportLabel(report *app.GoReport) string {
	if report.Source == "latest-patch" {
		return fmt.Sprintf("latest-patch of %s %s", path.Base(report.SourceFile), report.Base)
	}
	return report.Source
}

func lintSourceLabel(workDir string, selection app.Selection) string {
	if selection.LintSource == "detected" {
		return fmt.Sprintf("detected: %s:%d", currenttoolchain.DisplayPath(workDir, selection.LintSourceFile), selection.LintSourceLine)
	}
	return sourceLabel(selection.LintSource)
}

func lintReportLabel(report *app.LintReport) string {
	if report.Source == "detected" {
		return fmt.Sprintf("detected: %s:%d", report.SourceFile, report.SourceLine)
	}
	return report.Source
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
