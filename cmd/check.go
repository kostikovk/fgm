package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCheckCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var asJSON bool
	var all bool
	var failOn string

	cmd := &cobra.Command{
		Use:   "check [--all] [--json] [--fail-on error|warn]",
		Short: "Check that the repository's Go and golangci-lint versions agree",
		Long: `Offline repository consistency gate for CI and pre-commit. Reads repository files only: no network, installs, or writes.

  C1   ERROR  selected/detected golangci-lint cannot run on the resolved Go
  C2   ERROR  lint config schema does not match the linter major version
  C3   ERROR  golangci-lint pin sources disagree (.fgm.toml, detected pins)
  C4   ERROR  actions/setup-go go-version minor lower than the go directive
  C4b  WARN   actions/setup-go go-version minor differs from the resolved Go
  C5   ERROR  Dockerfile FROM golang:<tag> older than the go directive
  C6   ERROR  go.work go version lower than a used module's go directive
  C7   WARN   a build script runs golangci-lint/go by a path that bypasses fgm exec

Skip checks with [check] skip = ["C5"] in .fgm.toml. Exit status: 0 clean, 1 findings at or above --fail-on, 2 could not run.`,
		Args: usageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			threshold, err := parseFailOn(failOn)
			if err != nil {
				return usageError(err)
			}
			if application.RepoChecker == nil {
				return fmt.Errorf("repository checker is not configured")
			}
			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			var results []app.CheckReport
			if all {
				results, err = application.RepoChecker.CheckAll(cmd.Context(), workDir)
			} else {
				var result app.CheckReport
				result, err = application.RepoChecker.Check(cmd.Context(), workDir)
				results = []app.CheckReport{result}
			}
			if err != nil {
				return &ExitError{Code: 2, Err: err}
			}

			switch {
			case asJSON && all:
				err = writeJSON(cmd.OutOrStdout(), results)
			case asJSON:
				err = writeJSON(cmd.OutOrStdout(), results[0])
			default:
				err = printCheckText(cmd.OutOrStdout(), workDir, results, all)
			}
			if err != nil {
				return err
			}

			code := 0
			for _, result := range results {
				if result.Error != "" {
					code = 2
					break
				}
				for _, finding := range result.Findings {
					if severityRank(finding.Severity) >= threshold {
						code = 1
					}
				}
			}
			if code != 0 {
				return &ExitError{Code: code, Err: fmt.Errorf("fgm check failed"), Silent: true}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "print the current --json envelope plus a findings array")
	cmd.Flags().BoolVar(&all, "all", false, "check every module under the working directory")
	cmd.Flags().StringVar(&failOn, "fail-on", "error", "lowest severity that fails the check: error or warn")
	return cmd
}

func parseFailOn(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "error":
		return severityRank("ERROR"), nil
	case "warn":
		return severityRank("WARN"), nil
	default:
		return 0, fmt.Errorf("invalid --fail-on %q (want error or warn)", value)
	}
}

func severityRank(severity string) int {
	switch severity {
	case "ERROR":
		return 3
	case "WARN":
		return 2
	case "INFO":
		return 1
	}
	return 0
}

func printCheckText(w io.Writer, workDir string, results []app.CheckReport, all bool) error {
	for _, result := range results {
		prefix := ""
		if all {
			prefix = currenttoolchain.DisplayPath(workDir, result.Dir) + ": "
		}
		lines := make([]string, 0, len(result.Findings)+1)
		switch {
		case result.Error != "":
			lines = append(lines, "error: "+result.Error)
		case len(result.Findings) == 0:
			lines = append(lines, "OK no consistency findings")
		}
		for _, finding := range result.Findings {
			location := ""
			if finding.File != "" {
				location = currenttoolchain.DisplayPath(workDir, filepath.Join(result.Dir, filepath.FromSlash(finding.File)))
				if finding.Line > 0 {
					location += fmt.Sprintf(":%d", finding.Line)
				}
				location += " "
			}
			line := fmt.Sprintf("%s %s %s%s", finding.ID, finding.Severity, location, finding.Message)
			if finding.Hint != "" {
				line += " (fix: " + finding.Hint + ")"
			}
			lines = append(lines, line)
		}
		for _, line := range lines {
			if _, err := fmt.Fprintln(w, prefix+line); err != nil {
				return err
			}
		}
	}
	return nil
}
