package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/kostikovk/fgm/internal/report"
	"github.com/kostikovk/fgm/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newPruneCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var dryRun bool
	var all bool
	var force bool

	cmd := &cobra.Command{
		Use:   "prune [--all] [--dry-run] [--force]",
		Short: "Remove unused Go and golangci-lint versions from the local FGM store",
		Long:  "Remove locally installed Go and golangci-lint versions that are neither the global default nor selected by the current workspace. --all keeps every version selected by any module under the working directory (the modules fgm current --all reports). Without --all, prune warns about other modules it does not consider, and refuses to remove anything from a directory without go.work/go.mod unless --force is given. Use --dry-run to preview what would be kept and removed.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if application.GoStore == nil && application.LintStore == nil {
				return fmt.Errorf("local version stores are not configured")
			}

			ctx := cmd.Context()
			workDir, err := resolveWorkDir(v)
			if err != nil {
				return err
			}

			keepGo, keepLint, globalGoVersion, err := resolvePruneKeepSets(ctx, application, workDir, all)
			if err != nil {
				return err
			}

			if !all {
				unconsidered, err := warnUnconsideredModules(cmd, workDir)
				if err != nil {
					return err
				}
				_, inModule := workspace.ModuleRoot(workDir)
				_, inWorkspace := workspace.WorkRoot(workDir)
				if !inModule && !inWorkspace && !force && !dryRun {
					return fmt.Errorf(
						"refusing to prune: %s has no go.work or go.mod (%d Go modules below it); pass --all to keep every module's toolchains, or --force to prune anyway",
						workDir, unconsidered,
					)
				}
			}

			ui := uiFromCmd(cmd)
			goPruned, err := pruneGoVersions(ctx, cmd, ui, application.GoStore, keepGo, dryRun)
			if err != nil {
				return err
			}

			lintPruned, err := pruneLintVersions(ctx, cmd, ui, application, keepLint, globalGoVersion, dryRun)
			if err != nil {
				return err
			}

			if goPruned+lintPruned == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), ui.FormatEmptyNotice("No unused Go or golangci-lint versions to prune."))
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview kept and unused toolchains without removing them")
	cmd.Flags().BoolVar(&all, "all", false, "keep the toolchains selected by every module under the working directory")
	cmd.Flags().BoolVar(&force, "force", false, "prune from a directory without go.work or go.mod without --all")
	return cmd
}

// warnUnconsideredModules warns about modules whose toolchains prune is not keeping: modules
// below a directory without Go metadata, or the other modules of the current repository.
// It returns how many there are.
func warnUnconsideredModules(cmd *cobra.Command, workDir string) (int, error) {
	moduleRoot, inModule := workspace.ModuleRoot(workDir)
	scope := workDir
	if inModule {
		scope = workspace.Boundary(moduleRoot)
	}
	modules, err := report.ModuleDirs(scope)
	if err != nil {
		return 0, err
	}
	others := 0
	for _, module := range modules {
		if !inModule || module != moduleRoot {
			others++
		}
	}
	if others == 0 {
		return 0, nil
	}
	plural := "s"
	if others == 1 {
		plural = ""
	}
	message := fmt.Sprintf("warning: %d Go module%s below %s are not considered; their toolchains may be removed (use --all)", others, plural, workDir)
	if inModule {
		message = fmt.Sprintf("warning: %d other Go module%s in %s are not considered; their toolchains may be removed (run from the repository root with --all)", others, plural, scope)
	}
	_, err = fmt.Fprintln(cmd.ErrOrStderr(), message)
	return others, err
}

// resolvePruneKeepSets maps each version to keep to the reason it is kept.
func resolvePruneKeepSets(
	ctx context.Context,
	application *app.App,
	workDir string,
	all bool,
) (map[string]string, map[string]string, string, error) {
	keepGo := make(map[string]string)
	keepLint := make(map[string]string)

	var globalGoVersion string
	if application.GoStore != nil {
		globalGo, ok, err := application.GoStore.GlobalGoVersion(ctx)
		if err != nil {
			return nil, nil, "", err
		}
		if ok && globalGo != "" {
			globalGoVersion = globalGo
			keepGo[globalGo] = "global default"
		}
	}

	if all && application.Resolver != nil {
		modules, err := report.ModuleDirs(workDir)
		if err != nil {
			return nil, nil, "", err
		}
		for _, module := range modules {
			selection, err := application.Resolver.Current(ctx, module)
			if err != nil {
				// Removing toolchains a module might need is not recoverable offline; stop instead.
				return nil, nil, "", fmt.Errorf("resolve %s: %w", module, err)
			}
			reason := "selected by " + currenttoolchain.DisplayPath(workDir, module)
			keepVersion(keepGo, selection.GoVersion, reason)
			keepVersion(keepLint, selection.LintVersion, reason)
		}
	}

	if application.Resolver != nil {
		selection, err := application.Resolver.Current(ctx, workDir)
		if err != nil {
			if !isNoGoVersionSelectedError(err) {
				return nil, nil, "", err
			}
		} else {
			keepVersion(keepGo, selection.GoVersion, "selected by the current workspace")
			keepVersion(keepLint, selection.LintVersion, "selected by the current workspace")
		}
	}

	return keepGo, keepLint, globalGoVersion, nil
}

func keepVersion(keep map[string]string, version string, reason string) {
	if version == "" {
		return
	}
	if _, kept := keep[version]; !kept {
		keep[version] = reason
	}
}

func printKept(cmd *cobra.Command, tool string, version string, reason string) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "Keep %s %s (%s)\n", tool, version, reason)
	return err
}

func isNoGoVersionSelectedError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no Go version is selected")
}

func pruneGoVersions(
	ctx context.Context,
	cmd *cobra.Command,
	ui *uiRenderer,
	goStore app.GoLocalVersionStore,
	keepGo map[string]string,
	dryRun bool,
) (int, error) {
	if goStore == nil {
		return 0, nil
	}

	versions, err := goStore.ListLocalGoVersions(ctx)
	if err != nil {
		return 0, err
	}

	storeRoot := filepath.Clean(filepath.Dir(goStore.ShimDir()))
	pruned := 0
	for _, version := range versions {
		if reason, keep := keepGo[version]; keep {
			if dryRun {
				if err := printKept(cmd, "Go", version, reason); err != nil {
					return pruned, err
				}
			}
			continue
		}
		if !isManagedGoVersion(ctx, goStore, storeRoot, version) {
			continue
		}
		pruned++
		if dryRun {
			msg := ui.FormatOutcome(
				fmt.Sprintf("Would remove Go %s", version),
				"Would remove",
				"Go",
				version,
				"",
			)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
				return pruned, err
			}
			continue
		}
		removedPath, err := goStore.DeleteGoVersion(ctx, version)
		if err != nil {
			return pruned, err
		}
		msg := ui.FormatOutcome(
			fmt.Sprintf("Removed Go %s from %s", version, removedPath),
			"Removed",
			"Go",
			version,
			"from "+removedPath,
		)
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
			return pruned, err
		}
	}
	return pruned, nil
}

func pruneLintVersions(
	ctx context.Context,
	cmd *cobra.Command,
	ui *uiRenderer,
	application *app.App,
	keepLint map[string]string,
	globalGoVersion string,
	dryRun bool,
) (int, error) {
	if application.LintStore == nil {
		return 0, nil
	}

	versions, err := application.LintStore.ListLocalLintVersions(ctx)
	if err != nil {
		return 0, err
	}
	if globalGoVersion != "" && application.LintRemoteProvider != nil {
		if globalLint := pickCompatibleInstalledLint(ctx, application.LintRemoteProvider, globalGoVersion, versions); globalLint != "" {
			keepVersion(keepLint, globalLint, "compatible with the global default Go")
		}
	}

	pruned := 0
	for _, version := range versions {
		if reason, keep := keepLint[version]; keep {
			if dryRun {
				if err := printKept(cmd, "golangci-lint", version, reason); err != nil {
					return pruned, err
				}
			}
			continue
		}
		pruned++
		if dryRun {
			msg := ui.FormatOutcome(
				fmt.Sprintf("Would remove golangci-lint %s", version),
				"Would remove",
				"golangci-lint",
				version,
				"",
			)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
				return pruned, err
			}
			continue
		}
		removedPath, err := application.LintStore.DeleteLintVersion(ctx, version)
		if err != nil {
			return pruned, err
		}
		msg := ui.FormatOutcome(
			fmt.Sprintf("Removed golangci-lint %s from %s", version, removedPath),
			"Removed",
			"golangci-lint",
			version,
			"from "+removedPath,
		)
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
			return pruned, err
		}
	}
	return pruned, nil
}

func isManagedGoVersion(ctx context.Context, goStore app.GoLocalVersionStore, storeRoot string, version string) bool {
	if storeRoot == "" || storeRoot == "." {
		return true
	}
	binPath, err := goStore.GoBinaryPath(ctx, version)
	if err != nil || binPath == "" {
		return false
	}
	cleanBin := filepath.Clean(binPath)
	prefix := storeRoot + string(filepath.Separator)
	return strings.HasPrefix(cleanBin, prefix)
}

type catalogLintProvider interface {
	ListCatalogLintVersions(ctx context.Context, goVersion string) ([]app.LintVersion, error)
}

func pickCompatibleInstalledLint(
	ctx context.Context,
	provider app.LintRemoteVersionProvider,
	goVersion string,
	localVersions []string,
) string {
	var compatible []app.LintVersion
	var err error
	if cat, ok := provider.(catalogLintProvider); ok {
		compatible, err = cat.ListCatalogLintVersions(ctx, goVersion)
	} else {
		compatible, err = provider.ListRemoteLintVersions(ctx, goVersion)
	}
	if err != nil || len(compatible) == 0 {
		return ""
	}
	compatSet := make(map[string]struct{}, len(compatible))
	for _, item := range compatible {
		compatSet[item.Version] = struct{}{}
	}
	for _, localVer := range localVersions {
		if _, ok := compatSet[localVer]; ok {
			return localVer
		}
	}
	return ""
}
