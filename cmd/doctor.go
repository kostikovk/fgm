package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDoctorCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var fix bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check FGM installation and environment health",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.Doctor == nil {
				return fmt.Errorf("doctor service is not configured")
			}

			findings, err := application.Doctor.Diagnose(cmd.Context(), v.GetString(flagChdir))
			if err != nil {
				return err
			}

			for _, f := range findings {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", f.Severity, f.Message); err != nil {
					return err
				}
			}

			if !fix {
				// Print hint for fixable issues.
				for _, f := range findings {
					if f.FixKind == "shell_profile" {
						if _, err := fmt.Fprintln(cmd.OutOrStdout(), `Run: eval "$(fgm env)" or use --fix to add it to your shell profile`); err != nil {
							return err
						}
					}
				}
				return nil
			}

			// Apply fixes.
			for _, f := range findings {
				if f.FixKind == "shell_profile" {
					if err := applyShellProfileFix(cmd, application); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "auto-fix detected issues")
	return cmd
}

func applyShellProfileFix(cmd *cobra.Command, application *app.App) error {
	if application.ProfileInstaller == nil {
		return fmt.Errorf("profile installer is not configured")
	}

	// Prompt if stdin is a terminal.
	if isTerminal(cmd.InOrStdin()) {
		fmt.Fprint(cmd.OutOrStdout(), "Add shell integration to your profile? [y/N] ")
		var answer string
		fmt.Fscanln(cmd.InOrStdin(), &answer)
		if answer != "y" && answer != "Y" {
			return nil
		}
	}

	profilePath, modified, err := application.ProfileInstaller.InstallProfile("")
	if err != nil {
		return fmt.Errorf("failed to install shell profile: %w", err)
	}

	if modified {
		fmt.Fprintf(cmd.OutOrStdout(), "Added shell integration to %s\n", profilePath)
		fmt.Fprintf(cmd.OutOrStdout(), "Restart your shell or run: source %s\n", profilePath)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Shell integration already present in %s\n", profilePath)
	}
	return nil
}

func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
