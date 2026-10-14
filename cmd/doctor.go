package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

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

			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			findings, err := application.Doctor.Diagnose(cmd.Context(), workDir)
			if err != nil {
				return err
			}

			ui := uiFromCmd(cmd)
			for _, f := range findings {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), ui.FormatFinding(f.Severity, f.Message)); err != nil {
					return err
				}
			}

			if !fix {
				// Print hint for fixable issues.
				for _, f := range findings {
					if f.FixKind == "shell_profile" {
						hint := `Run: eval "$(fgm env)" or use --fix to add it to your shell profile`
						if _, err := fmt.Fprintln(cmd.OutOrStdout(), ui.FormatEmptyNotice(hint)); err != nil {
							return err
						}
					}
				}
				return nil
			}

			// Apply fixes.
			if err := ensureShimsIfSupported(application.GoStore); err != nil {
				return err
			}
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

	cmd.Flags().BoolVar(&fix, "fix", false, "add shell integration to your shell profile when missing")
	return cmd
}

func applyShellProfileFix(cmd *cobra.Command, application *app.App) error {
	if application.ProfileInstaller == nil {
		return fmt.Errorf("profile installer is not configured")
	}

	// Prompt if stdin is a terminal.
	if isTerminal(cmd.InOrStdin()) {
		if _, err := fmt.Fprint(cmd.OutOrStdout(), "Add shell integration to your profile? [y/N] "); err != nil {
			return err
		}
		var answer string
		if _, err := fmt.Fscanln(cmd.InOrStdin(), &answer); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if answer != "y" && answer != "Y" {
			return nil
		}
	}

	profilePath, modified, err := application.ProfileInstaller.InstallProfile("")
	if err != nil {
		return fmt.Errorf("failed to install shell profile: %w", err)
	}

	ui := uiFromCmd(cmd)
	if modified {
		addedLine := ui.FormatOutcome(
			fmt.Sprintf("Added shell integration to %s", profilePath),
			"Added",
			"shell integration",
			"",
			"to "+profilePath,
		)
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), addedLine); err != nil {
			return err
		}
		restartLine := ui.FormatEmptyNotice(fmt.Sprintf("Restart your shell or run: source %s", profilePath))
		_, err = fmt.Fprintln(cmd.OutOrStdout(), restartLine)
		return err
	}
	presentLine := ui.FormatEmptyNotice(fmt.Sprintf("Shell integration already present in %s", profilePath))
	_, err = fmt.Fprintln(cmd.OutOrStdout(), presentLine)
	return err
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

func ensureShimsIfSupported(goStore app.GoLocalVersionStore) error {
	if goStore == nil || runtime.GOOS == "windows" {
		return nil
	}
	return goStore.EnsureShims()
}
