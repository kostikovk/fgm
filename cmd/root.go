package cmd

import (
	"fmt"
	"path/filepath"

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
	v := viper.New()
	v.SetEnvPrefix("FGM")
	v.AutomaticEnv()
	v.SetDefault(flagChdir, ".")

	rootCmd := &cobra.Command{
		Use:   "fgm",
		Short: "Manage Go and golangci-lint toolchains",
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

	return rootCmd
}
