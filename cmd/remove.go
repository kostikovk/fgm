package cmd

import (
	"fmt"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newRemoveCmd(application *app.App) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove toolchains from the local FGM store",
	}

	removeCmd.AddCommand(newRemoveGoCmd(application))
	removeCmd.AddCommand(newRemoveLintCmd(application))

	return removeCmd
}

func newRemoveGoCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "go [version]",
		Short: "Remove an FGM-managed Go version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.GoStore == nil {
				return fmt.Errorf("local Go version store is not configured")
			}

			version := args[0]
			removedPath, err := application.GoStore.DeleteGoVersion(cmd.Context(), version)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed Go %s from %s\n", version, removedPath)
			return err
		},
	}
}

func newRemoveLintCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "golangci-lint [version]",
		Short: "Remove an FGM-managed golangci-lint version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.LintStore == nil {
				return fmt.Errorf("local golangci-lint version store is not configured")
			}

			version := args[0]
			removedPath, err := application.LintStore.DeleteLintVersion(cmd.Context(), version)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed golangci-lint %s from %s\n", version, removedPath)
			return err
		},
	}
}
