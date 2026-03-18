package cmd

import (
	"fmt"
	"strings"

	"github.com/kostikovk/fgm/internal/fgmconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newPinCmd(v *viper.Viper) *cobra.Command {
	pinCmd := &cobra.Command{
		Use:   "pin",
		Short: "Pin repo-level toolchain policy",
	}

	pinCmd.AddCommand(newPinLintCmd(v))

	return pinCmd
}

func newPinLintCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "golangci-lint [version|auto]",
		Short: "Pin the repo golangci-lint version in .fgm.toml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := strings.TrimSpace(args[0])
			if version == "" {
				return fmt.Errorf("version must not be empty")
			}

			path, err := fgmconfig.SaveNearest(v.GetString(flagChdir), fgmconfig.File{
				Toolchain: fgmconfig.ToolchainConfig{GolangCILint: version},
			})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Pinned golangci-lint %s in %s\n", version, path)
			return err
		},
	}
}
