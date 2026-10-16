package cmd

import (
	"fmt"
	"strings"

	"github.com/kostikovk/fgm/internal/fgmconfig"
	"github.com/kostikovk/fgm/internal/resolve"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newPinCmd(v *viper.Viper) *cobra.Command {
	pinCmd := &cobra.Command{
		Use:   "pin",
		Short: "Pin repo-level toolchain policy",
	}

	pinCmd.AddCommand(newPinGoCmd(v))
	pinCmd.AddCommand(newPinLintCmd(v))

	return pinCmd
}

func newPinLintCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "golangci-lint [version|auto]",
		Short: "Pin the repo golangci-lint version in .fgm.toml",
		Long:  "Record the repository golangci-lint version policy (an explicit version or 'auto') in the nearest .fgm.toml. Pinning does not install the linter or validate compatibility upfront; run 'fgm install' and 'fgm doctor' afterward.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := strings.TrimSpace(args[0])
			if version == "" {
				return fmt.Errorf("version must not be empty")
			}

			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			path, err := fgmconfig.UpdateNearest(workDir, func(file *fgmconfig.File) {
				file.Toolchain.GolangCILint = version
			})
			if err != nil {
				return err
			}

			ui := uiFromCmd(cmd)
			msg := ui.FormatOutcome(
				fmt.Sprintf("Pinned golangci-lint %s in %s", version, path),
				"Pinned",
				"golangci-lint",
				version,
				"in "+path,
			)
			_, err = fmt.Fprintln(cmd.OutOrStdout(), msg)
			return err
		},
	}
}

func newPinGoCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "go [version|latest-patch]",
		Short: "Pin the repo Go version policy in .fgm.toml",
		Long:  "Record the repository Go version policy in the nearest .fgm.toml: an explicit version (it must not be lower than the go directive) or 'latest-patch' (the newest released patch of the go directive's minor). Other keys in .fgm.toml are kept; comments are not. Pinning does not install Go; run 'fgm install' afterward.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			policy, err := fgmconfig.ParseGoPolicyValue(args[0])
			if err != nil {
				return err
			}

			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}
			if policy != fgmconfig.LatestPatchPolicy {
				if err := resolve.ValidateGoPin(workDir, policy); err != nil {
					return err
				}
			}

			path, err := fgmconfig.UpdateNearest(workDir, func(file *fgmconfig.File) {
				file.Toolchain.Go = policy
			})
			if err != nil {
				return err
			}

			ui := uiFromCmd(cmd)
			msg := ui.FormatOutcome(
				fmt.Sprintf("Pinned Go %s in %s", policy, path),
				"Pinned",
				"Go",
				policy,
				"in "+path,
			)
			_, err = fmt.Fprintln(cmd.OutOrStdout(), msg)
			return err
		},
	}
}
