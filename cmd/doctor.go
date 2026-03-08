package cmd

import (
	"fmt"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
)

func newDoctorCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check FGM installation and environment health",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.Doctor == nil {
				return fmt.Errorf("doctor service is not configured")
			}

			lines, err := application.Doctor.Diagnose(cmd.Context())
			if err != nil {
				return err
			}

			for _, line := range lines {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
