package cmd

import (
	"fmt"

	"github.com/kostikovk/fgm/internal/app"
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
		Short: "Automatically import existing Go and golangci-lint installations from common locations",
		Long:  "Discover existing Go and golangci-lint installations in standard system paths, Go toolchains cached in the module cache by Go's own toolchain switching, and golang.org/dl SDKs (~/sdk), and register them as symlinks in the FGM store without modifying the global default Go selection. Toolchain directories whose bin/go is missing or whose VERSION file disagrees with the directory name are listed as skipped.",
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

			ui := uiFromCmd(cmd)
			importedAny := len(importedLint) > 0
			for _, item := range importedGo {
				importedAny = importedAny || item.SkipReason == ""
			}
			for _, item := range importedGo {
				if item.SkipReason == "" {
					continue
				}
				notice := fmt.Sprintf("Skipped Go %s at %s (%s): %s", item.Version, item.Path, item.Source, item.SkipReason)
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), ui.FormatEmptyNotice(notice)); err != nil {
					return err
				}
			}
			if !importedAny {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), ui.FormatEmptyNotice("No Go or golangci-lint installations were imported."))
				return err
			}

			for _, item := range importedGo {
				if item.SkipReason != "" {
					continue
				}
				from := "from " + item.Path
				if item.Source != "" {
					from += " (" + item.Source + ")"
				}
				msg := ui.FormatOutcome(
					fmt.Sprintf("Imported Go %s %s", item.Version, from),
					"Imported",
					"Go",
					item.Version,
					from,
				)
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
					return err
				}
			}
			for _, item := range importedLint {
				msg := ui.FormatOutcome(
					fmt.Sprintf("Imported golangci-lint %s from %s", item.Version, item.Path),
					"Imported",
					"golangci-lint",
					item.Version,
					"from "+item.Path,
				)
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
					return err
				}
			}
			return ensureShimsIfSupported(application.GoStore)
		},
	}
}
