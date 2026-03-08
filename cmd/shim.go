package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/shim"
	"github.com/spf13/cobra"
)

func newShimCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:    "__shim [tool] [args...]",
		Short:  "Internal shim entrypoint",
		Hidden: true,
		Args:   cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.GoStore == nil || application.Resolver == nil {
				return fmt.Errorf("shim dependencies are not configured")
			}

			switch args[0] {
			case "go":
				workDir, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("resolve working directory: %w", err)
				}

				resolver := shim.New(application.Resolver, application.GoStore)
				binaryPath, err := resolver.ResolveGoBinary(cmd.Context(), workDir)
				if err != nil {
					return err
				}

				execCmd := exec.CommandContext(cmd.Context(), binaryPath, args[1:]...)
				execCmd.Stdin = cmd.InOrStdin()
				execCmd.Stdout = cmd.OutOrStdout()
				execCmd.Stderr = cmd.ErrOrStderr()
				execCmd.Env = os.Environ()
				return execCmd.Run()
			default:
				return fmt.Errorf("unsupported shim tool %q", args[0])
			}
		},
	}
}
