package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newImportCmd(application *app.App) *cobra.Command {
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import existing toolchains into FGM",
	}

	importCmd.AddCommand(newImportAutoCmd(application))

	return importCmd
}

func newImportAutoCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "auto",
		Short: "Automatically import existing Go installations from common locations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.GoImporter == nil && application.LintImporter == nil {
				return fmt.Errorf("importers are not configured")
			}

			importedGo := []app.ImportedGo(nil)
			if application.GoImporter != nil {
				var err error
				importedGo, err = application.GoImporter.ImportAuto(cmd.Context())
				if err != nil {
					return err
				}
			}

			importedLint := []app.ImportedLint(nil)
			if application.LintImporter != nil {
				var err error
				importedLint, err = application.LintImporter.ImportAuto(cmd.Context())
				if err != nil {
					return err
				}
			}

			if len(importedGo) == 0 && len(importedLint) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No Go or golangci-lint installations were imported.")
				return err
			}

			for _, item := range importedGo {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Imported Go %s from %s\n", item.Version, item.Path); err != nil {
					return err
				}
			}
			for _, item := range importedLint {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Imported golangci-lint %s from %s\n", item.Version, item.Path); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
