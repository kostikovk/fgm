package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/goenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newExecCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var installMissing bool
	var goToolchain string
	var allowUnselectedLint bool

	cmd := &cobra.Command{
		Use:   "exec [--install] [--gotoolchain local|inherit] [--allow-unselected-lint] -- [command] [args...]",
		Short: "Run a command with the resolved Go and installed selected golangci-lint on PATH",
		Long:  "Run a child process with the workspace's resolved Go binary directory (and installed selected golangci-lint directory, when present) prepended to PATH. Pass --install to automatically install missing resolved Go and golangci-lint versions before running the command. The child runs with GOTOOLCHAIN=local and without GOROOT so the selected Go is the Go that runs. When the selected golangci-lint is not installed, running golangci-lint fails with an install hint (pass --install, or --allow-unselected-lint to use whatever is on PATH) and other commands warn if golangci-lint on PATH is a different binary. Pass --gotoolchain inherit (or set FGM_EXEC_GOTOOLCHAIN=inherit) to keep the caller's values. Use this in non-interactive sessions where shell profile changes do not persist across processes.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command provided")
			}
			if application.Executor == nil {
				return fmt.Errorf("executor is not configured")
			}

			modeValue := optsForCommand(cmd).withDefaults().Getenv(goenv.ModeEnvVar)
			if cmd.Flags().Changed("gotoolchain") {
				modeValue = goToolchain
			}
			mode, err := goenv.ParseMode(modeValue)
			if err != nil {
				return err
			}

			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			if installMissing {
				if err := ensureExecToolchainsInstalled(cmd, application, workDir); err != nil {
					return err
				}
			}

			return application.Executor.Exec(
				cmd.Context(),
				workDir,
				args,
				app.ExecOptions{GoToolchain: string(mode), AllowUnselectedLint: allowUnselectedLint},
				cmd.InOrStdin(),
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			)
		},
	}

	cmd.Flags().BoolVar(&installMissing, "install", false, "install missing resolved Go and golangci-lint versions before executing")
	cmd.Flags().StringVar(&goToolchain, "gotoolchain", "", "GOTOOLCHAIN handling for the child: local (default) or inherit (overrides "+goenv.ModeEnvVar+")")
	cmd.Flags().BoolVar(&allowUnselectedLint, "allow-unselected-lint", false, "let golangci-lint resolve from the inherited PATH when the selected version is not installed")
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func ensureExecToolchainsInstalled(cmd *cobra.Command, application *app.App, workDir string) error {
	if application.Resolver == nil {
		return fmt.Errorf("resolver is not configured")
	}
	if application.GoStore == nil {
		return fmt.Errorf("local Go version store is not configured")
	}
	if application.GoInstaller == nil {
		return fmt.Errorf("go installer is not configured")
	}

	ctx := cmd.Context()
	selection, err := application.Resolver.Current(ctx, workDir)
	if err != nil {
		return err
	}

	ui := uiFromCmd(cmd)
	errUI := *ui
	errUI.Out = ui.Err

	installedAny := false
	hasGo, err := application.GoStore.HasGoVersion(ctx, selection.GoVersion)
	if err != nil {
		return err
	}
	if !hasGo {
		goInstallPath, err := application.GoInstaller.InstallGoVersion(ctx, selection.GoVersion)
		if err != nil {
			return err
		}
		installedAny = true
		goMsg := errUI.FormatOutcome(
			fmt.Sprintf("Installed Go %s to %s", selection.GoVersion, goInstallPath),
			"Installed",
			"Go",
			selection.GoVersion,
			"to "+goInstallPath,
		)
		if _, err := fmt.Fprintln(cmd.ErrOrStderr(), goMsg); err != nil {
			return err
		}
	}

	if selection.LintVersion != "" && application.LintStore != nil {
		_, lintErr := application.LintStore.LintBinaryPath(ctx, selection.LintVersion)
		if lintErr != nil {
			if !errors.Is(lintErr, os.ErrNotExist) {
				return lintErr
			}
			if application.LintInstaller == nil {
				return fmt.Errorf("golangci-lint installer is not configured")
			}
			lintInstallPath, err := application.LintInstaller.InstallLintVersion(ctx, selection.LintVersion)
			if err != nil {
				return err
			}
			installedAny = true
			lintMsg := errUI.FormatOutcome(
				fmt.Sprintf("Installed golangci-lint %s to %s", selection.LintVersion, lintInstallPath),
				"Installed",
				"golangci-lint",
				selection.LintVersion,
				"to "+lintInstallPath,
			)
			if _, err := fmt.Fprintln(cmd.ErrOrStderr(), lintMsg); err != nil {
				return err
			}
		}
	}

	if installedAny {
		refreshUpdateCache(ctx, application.FGMUpdateChecker)
		if err := ensureShimsIfSupported(application.GoStore); err != nil {
			return err
		}
	}
	return nil
}
