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
			if application.GoImporter == nil {
				return fmt.Errorf("Go importer is not configured")
			}

			imported, err := application.GoImporter.ImportAuto(cmd.Context())
			if err != nil {
				return err
			}
			if len(imported) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No Go installations were imported.")
				return err
			}

			for _, item := range imported {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Imported Go %s from %s\n", item.Version, item.Path); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
