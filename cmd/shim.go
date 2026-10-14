package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/shim"
	"github.com/spf13/cobra"
)

func newShimCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:    "__shim [tool] [args...]",
		Short:  "Internal shim entrypoint",
		Hidden: true,
		Args:   cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "go":
				goResolver := application.GoResolver
				if goResolver == nil {
					goResolver = application.Resolver
				}
				if application.GoStore == nil || goResolver == nil {
					return fmt.Errorf("shim dependencies are not configured")
				}

				workDir, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("resolve working directory: %w", err)
				}

				resolver := shim.New(goResolver, application.GoStore)
				binaryPath, err := resolver.ResolveGoBinary(cmd.Context(), workDir)
				if err != nil {
					return err
				}

				return runShimBinary(cmd, binaryPath, args[1:], os.Environ())

			case "golangci-lint":
				if application.Resolver == nil || application.LintStore == nil {
					return fmt.Errorf("shim dependencies are not configured")
				}

				workDir, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("resolve working directory: %w", err)
				}

				resolver := shim.NewWithLint(application.Resolver, application.GoStore, application.LintStore)
				lintExec, err := resolver.ResolveLint(cmd.Context(), workDir)
				if err != nil {
					return err
				}

				env := os.Environ()
				if lintExec.GoBinaryDir != "" {
					env = prependDirToPathEnv(env, lintExec.GoBinaryDir)
				}

				return runShimBinary(cmd, lintExec.BinaryPath, args[1:], env)

			default:
				return fmt.Errorf("unsupported shim tool %q", args[0])
			}
		},
	}
}

func runShimBinary(cmd *cobra.Command, binaryPath string, args []string, env []string) error {
	execCmd := exec.CommandContext(cmd.Context(), binaryPath, args...)
	execCmd.Stdin = cmd.InOrStdin()
	execCmd.Stdout = cmd.OutOrStdout()
	execCmd.Stderr = cmd.ErrOrStderr()
	execCmd.Env = env
	if err := execCmd.Run(); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func prependDirToPathEnv(env []string, dir string) []string {
	updated := make([]string, 0, len(env)+1)
	found := false
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			existing := entry[len("PATH="):]
			if existing != "" {
				updated = append(updated, "PATH="+dir+string(os.PathListSeparator)+existing)
			} else {
				updated = append(updated, "PATH="+dir)
			}
			found = true
			continue
		}
		updated = append(updated, entry)
	}
	if !found {
		updated = append(updated, "PATH="+dir)
	}
	return updated
}
