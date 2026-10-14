package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagChdir = "chdir"
)

func resolveWorkDir(v *viper.Viper) (string, error) {
	dir := v.GetString(flagChdir)
	if dir == "" {
		dir = "."
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	return absDir, nil
}

// NewRootCmd builds the root Cobra command for fgm.
func NewRootCmd(application *app.App) *cobra.Command {
	return NewRootCmdWithUI(application, UIOptions{})
}

// NewRootCmdWithUI builds the root Cobra command with explicit UI capability options.
func NewRootCmdWithUI(application *app.App, opts UIOptions) *cobra.Command {
	v := viper.New()
	v.SetEnvPrefix("FGM")
	v.AutomaticEnv()
	v.SetDefault(flagChdir, ".")

	rootCmd := &cobra.Command{
		Use:          "fgm",
		Short:        "Manage Go and golangci-lint toolchains",
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmd.SetContext(withUIOptions(cmd.Context(), opts))
		},
		PersistentPostRun: func(cmd *cobra.Command, _ []string) {
			maybePrintCachedUpdateNotice(cmd, application)
		},
	}

	rootCmd.PersistentFlags().String(flagChdir, ".", "working directory for repo resolution")

	_ = v.BindPFlag(flagChdir, rootCmd.PersistentFlags().Lookup(flagChdir))

	rootCmd.AddCommand(newCurrentCmd(application, v))
	rootCmd.AddCommand(newDoctorCmd(application, v))
	rootCmd.AddCommand(newEnvCmd(application))
	rootCmd.AddCommand(newExecCmd(application, v))
	rootCmd.AddCommand(newImportCmd(application))
	rootCmd.AddCommand(newInstallCmd(application, v))
	rootCmd.AddCommand(newPinCmd(v))
	rootCmd.AddCommand(newRemoveCmd(application))
	rootCmd.AddCommand(newUpgradeCmd(application, v))
	rootCmd.AddCommand(newUseCmd(application))
	rootCmd.AddCommand(newVersionCmd(application))
	rootCmd.AddCommand(newVersionsCmd(application, v))
	rootCmd.AddCommand(newShimCmd(application))
	rootCmd.AddCommand(newLintCmd(application, v))

	configureFangHelp(rootCmd, opts)

	return rootCmd
}

func maybePrintCachedUpdateNotice(cmd *cobra.Command, application *app.App) {
	if cmd == nil || application == nil || application.FGMUpdateChecker == nil {
		return
	}
	switch cmd.Name() {
	case "__shim", "env", "exec", "version", "doctor":
		return
	}
	resolvedOpts := optsForCommand(cmd).withDefaults()
	if strings.TrimSpace(resolvedOpts.Getenv("FGM_NO_UPDATE_NOTIFIER")) != "" {
		return
	}
	ui := uiFromCmd(cmd)
	if !ui.Err.Styled {
		return
	}
	status, ok := application.FGMUpdateChecker.CachedNotice(cmd.Context())
	if !ok || !status.UpdateAvailable {
		return
	}
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), ui.FormatUpdateNotice(status))
}

func refreshUpdateCache(ctx context.Context, checker app.FGMUpdateChecker) {
	if checker == nil {
		return
	}
	_, _, _ = checker.CheckLatest(ctx)
}
