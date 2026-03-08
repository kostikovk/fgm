package cmd

import (
	"github.com/koskosovu4/fgm/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagChdir   = "chdir"
	flagConfig  = "config"
	flagVerbose = "verbose"
)

// NewRootCmd builds the root Cobra command for fgm.
func NewRootCmd(application *app.App) *cobra.Command {
	v := viper.New()
	v.SetEnvPrefix("FGM")
	v.AutomaticEnv()
	v.SetDefault(flagChdir, ".")

	rootCmd := &cobra.Command{
		Use:   "fgm",
		Short: "Manage Go and golangci-lint toolchains",
	}

	rootCmd.PersistentFlags().String(flagChdir, ".", "working directory for repo resolution")
	rootCmd.PersistentFlags().String(flagConfig, "", "config file path")
	rootCmd.PersistentFlags().Bool(flagVerbose, false, "verbose output")

	_ = v.BindPFlag(flagChdir, rootCmd.PersistentFlags().Lookup(flagChdir))
	_ = v.BindPFlag(flagConfig, rootCmd.PersistentFlags().Lookup(flagConfig))
	_ = v.BindPFlag(flagVerbose, rootCmd.PersistentFlags().Lookup(flagVerbose))

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
	rootCmd.AddCommand(newVersionsCmd(application, v))
	rootCmd.AddCommand(newShimCmd(application))

	return rootCmd
}
