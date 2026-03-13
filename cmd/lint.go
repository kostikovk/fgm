package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/lintconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagPreset      = "preset"
	flagWithImports = "with-imports"
	flagForce       = "force"
)

func newLintCmd(application *app.App, v *viper.Viper) *cobra.Command {
	lintCmd := &cobra.Command{
		Use:   "lint",
		Short: "Manage golangci-lint configuration",
	}

	lintCmd.AddCommand(newLintInitCmd(application, v))
	lintCmd.AddCommand(newLintDoctorCmd(application, v))

	return lintCmd
}

func newLintInitCmd(application *app.App, v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a best-practice .golangci.yml",
		Long:  "Generate a golangci-lint v2 configuration file tuned to the project's Go version.",
		RunE: func(cmd *cobra.Command, args []string) error {
			preset, _ := cmd.Flags().GetString(flagPreset)
			effectivePreset, err := lintconfig.NormalizePreset(preset)
			if err != nil {
				return err
			}

			if application.LintConfigGenerator == nil {
				return fmt.Errorf("lint config generator is not available")
			}

			workDir := v.GetString(flagChdir)
			withImports, _ := cmd.Flags().GetBool(flagWithImports)
			force, _ := cmd.Flags().GetBool(flagForce)

			_, err = application.LintConfigGenerator.Generate(cmd.Context(), app.LintConfigOptions{
				WorkDir:     workDir,
				Preset:      effectivePreset,
				WithImports: withImports,
				Force:       force,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated .golangci.yml (preset: %s)\n", effectivePreset)
			return nil
		},
	}

	cmd.Flags().String(flagPreset, "standard", "linter preset: minimal, standard, strict")
	cmd.Flags().Bool(flagWithImports, false, "include import ordering configuration (gci)")
	cmd.Flags().Bool(flagForce, false, "overwrite existing .golangci.yml")

	return cmd
}

func newLintDoctorCmd(application *app.App, v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Audit existing golangci-lint configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.LintDoctor == nil {
				return fmt.Errorf("lint doctor is not available")
			}

			workDir := v.GetString(flagChdir)
			findings, err := application.LintDoctor.Diagnose(cmd.Context(), workDir)
			if err != nil {
				return err
			}

			for _, f := range findings {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", f.Severity, f.Message)
			}
			return nil
		},
	}
}
