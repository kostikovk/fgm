package cmd

import (
	"fmt"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newUpgradeCmd(application *app.App, v *viper.Viper) *cobra.Command {
	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade toolchains to newer versions",
	}

	upgradeCmd.AddCommand(newUpgradeGoCmd(application, v))
	upgradeCmd.AddCommand(newUpgradeLintCmd(application, v))

	return upgradeCmd
}

func newUpgradeGoCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var global bool
	var project bool
	var dryRun bool
	var version string
	var withLint bool

	cmd := &cobra.Command{
		Use:   "go",
		Short: "Upgrade Go globally or for the current project",
		Long:  "Upgrade Go globally (--global) or in the nearest project go.work/go.mod (--project). Use --dry-run to preview the target version and --to <version> to freeze an explicit target; --with-lint preserves explicit .fgm.toml pins.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if global && project {
				return fmt.Errorf("--global and --project are mutually exclusive")
			}
			if !global && !project {
				return fmt.Errorf("provide --global or --project")
			}
			if application.GoUpgrader == nil {
				return fmt.Errorf("go upgrader is not configured")
			}

			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			options := app.GoUpgradeOptions{
				WorkDir:  workDir,
				Version:  version,
				DryRun:   dryRun,
				WithLint: withLint,
			}

			ui := uiFromCmd(cmd)
			if global {
				result, err := application.GoUpgrader.UpgradeGlobal(cmd.Context(), options)
				if err != nil {
					return err
				}
				if result.DryRun {
					goMsg := ui.FormatOutcome(
						fmt.Sprintf("Would upgrade global Go to %s", result.Version),
						"Would upgrade global to",
						"Go",
						result.Version,
						"",
					)
					if _, err = fmt.Fprintln(cmd.OutOrStdout(), goMsg); err != nil {
						return err
					}
					if result.LintVersion != "" {
						lintMsg := ui.FormatOutcome(
							fmt.Sprintf("Would install golangci-lint %s", result.LintVersion),
							"Would install",
							"golangci-lint",
							result.LintVersion,
							"",
						)
						_, err = fmt.Fprintln(cmd.OutOrStdout(), lintMsg)
					}
					return err
				}
				goMsg := ui.FormatOutcome(
					fmt.Sprintf("Upgraded global Go to %s", result.Version),
					"Upgraded global to",
					"Go",
					result.Version,
					"",
				)
				if _, err = fmt.Fprintln(cmd.OutOrStdout(), goMsg); err != nil {
					return err
				}
				if result.LintVersion != "" {
					lintMsg := ui.FormatOutcome(
						fmt.Sprintf("Installed golangci-lint %s", result.LintVersion),
						"Installed",
						"golangci-lint",
						result.LintVersion,
						"",
					)
					_, err = fmt.Fprintln(cmd.OutOrStdout(), lintMsg)
				}
				return err
			}

			result, err := application.GoUpgrader.UpgradeProject(cmd.Context(), options)
			if err != nil {
				return err
			}
			if result.DryRun {
				goMsg := ui.FormatOutcome(
					fmt.Sprintf("Would upgrade project Go to %s in %s", result.Version, result.Path),
					"Would upgrade project to",
					"Go",
					result.Version,
					"in "+result.Path,
				)
				if _, err = fmt.Fprintln(cmd.OutOrStdout(), goMsg); err != nil {
					return err
				}
				if result.LintVersion != "" {
					lintMsg := ui.FormatOutcome(
						fmt.Sprintf("Would install golangci-lint %s", result.LintVersion),
						"Would install",
						"golangci-lint",
						result.LintVersion,
						"",
					)
					_, err = fmt.Fprintln(cmd.OutOrStdout(), lintMsg)
				}
				return err
			}
			goMsg := ui.FormatOutcome(
				fmt.Sprintf("Upgraded project Go to %s in %s", result.Version, result.Path),
				"Upgraded project to",
				"Go",
				result.Version,
				"in "+result.Path,
			)
			if _, err = fmt.Fprintln(cmd.OutOrStdout(), goMsg); err != nil {
				return err
			}
			if result.LintVersion != "" {
				lintMsg := ui.FormatOutcome(
					fmt.Sprintf("Installed golangci-lint %s", result.LintVersion),
					"Installed",
					"golangci-lint",
					result.LintVersion,
					"",
				)
				if _, err = fmt.Fprintln(cmd.OutOrStdout(), lintMsg); err != nil {
					return err
				}
			}
			return ensureShimsIfSupported(application.GoStore)
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "upgrade the global default Go version")
	cmd.Flags().BoolVar(&project, "project", false, "upgrade the nearest project Go version")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview the selected Go upgrade without changing anything")
	cmd.Flags().StringVar(&version, "to", "", "upgrade to a specific Go version instead of the latest remote version")
	cmd.Flags().BoolVar(&withLint, "with-lint", false, "also install the matching or pinned golangci-lint version")

	return cmd
}

func newUpgradeLintCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var dryRun bool
	var version string

	cmd := &cobra.Command{
		Use:   "golangci-lint",
		Short: "Upgrade golangci-lint to the latest compatible version",
		Long:  "Install the latest compatible (or pinned) golangci-lint version for the workspace's resolved Go version, or use --to <version> to install an explicit version.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.LintUpgrader == nil {
				return fmt.Errorf("golangci-lint upgrader is not configured")
			}

			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			result, err := application.LintUpgrader.Upgrade(cmd.Context(), app.LintUpgradeOptions{
				WorkDir: workDir,
				Version: version,
				DryRun:  dryRun,
			})
			if err != nil {
				return err
			}

			ui := uiFromCmd(cmd)
			if result.DryRun {
				msg := ui.FormatOutcome(
					fmt.Sprintf("Would install golangci-lint %s (for Go %s)", result.Version, result.GoVersion),
					"Would install",
					"golangci-lint",
					result.Version,
					fmt.Sprintf("(for Go %s)", result.GoVersion),
				)
				_, err = fmt.Fprintln(cmd.OutOrStdout(), msg)
				return err
			}
			msg := ui.FormatOutcome(
				fmt.Sprintf("Installed golangci-lint %s (for Go %s)", result.Version, result.GoVersion),
				"Installed",
				"golangci-lint",
				result.Version,
				fmt.Sprintf("(for Go %s)", result.GoVersion),
			)
			if _, err = fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
				return err
			}
			return ensureShimsIfSupported(application.GoStore)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview the selected upgrade without installing")
	cmd.Flags().StringVar(&version, "to", "", "upgrade to a specific version instead of the latest compatible")

	return cmd
}
