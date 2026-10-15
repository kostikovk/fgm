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

func newPruneCmd(application *app.App, v *viper.Viper) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove unused Go and golangci-lint versions from the local FGM store",
		Long:  "Remove locally installed Go and golangci-lint versions that are neither the global default nor selected by the current workspace. Use --dry-run to preview what would be removed.",
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

			keepGo := make(map[string]struct{})
			keepLint := make(map[string]struct{})

			var globalGoVersion string
			if application.GoStore != nil {
				if globalGo, ok, err := application.GoStore.GlobalGoVersion(ctx); err != nil {
					return err
				} else if ok && globalGo != "" {
					globalGoVersion = globalGo
					keepGo[globalGo] = struct{}{}
				}
			}

			if application.Resolver != nil {
				if selection, err := application.Resolver.Current(ctx, workDir); err == nil {
					if selection.GoVersion != "" {
						keepGo[selection.GoVersion] = struct{}{}
					}
					if selection.LintVersion != "" {
						keepLint[selection.LintVersion] = struct{}{}
					}
				}
			}

			ui := uiFromCmd(cmd)
			prunedCount := 0

			if application.GoStore != nil {
				versions, err := application.GoStore.ListLocalGoVersions(ctx)
				if err != nil {
					return err
				}
				storeRoot := filepath.Clean(filepath.Dir(application.GoStore.ShimDir()))
				for _, version := range versions {
					if _, keep := keepGo[version]; keep {
						continue
					}
					if !isManagedGoVersion(ctx, application.GoStore, storeRoot, version) {
						continue
					}
					prunedCount++
					if dryRun {
						msg := ui.FormatOutcome(
							fmt.Sprintf("Would remove Go %s", version),
							"Would remove",
							"Go",
							version,
							"",
						)
						if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
							return err
						}
						continue
					}
					removedPath, err := application.GoStore.DeleteGoVersion(ctx, version)
					if err != nil {
						return err
					}
					msg := ui.FormatOutcome(
						fmt.Sprintf("Removed Go %s from %s", version, removedPath),
						"Removed",
						"Go",
						version,
						"from "+removedPath,
					)
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
						return err
					}
				}
			}

			if application.LintStore != nil {
				versions, err := application.LintStore.ListLocalLintVersions(ctx)
				if err != nil {
					return err
				}
				if globalGoVersion != "" && application.LintRemoteProvider != nil {
					if globalLint := pickCompatibleInstalledLint(ctx, application.LintRemoteProvider, globalGoVersion, versions); globalLint != "" {
						keepLint[globalLint] = struct{}{}
					}
				}
				for _, version := range versions {
					if _, keep := keepLint[version]; keep {
						continue
					}
					prunedCount++
					if dryRun {
						msg := ui.FormatOutcome(
							fmt.Sprintf("Would remove golangci-lint %s", version),
							"Would remove",
							"golangci-lint",
							version,
							"",
						)
						if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
							return err
						}
						continue
					}
					removedPath, err := application.LintStore.DeleteLintVersion(ctx, version)
					if err != nil {
						return err
					}
					msg := ui.FormatOutcome(
						fmt.Sprintf("Removed golangci-lint %s from %s", version, removedPath),
						"Removed",
						"golangci-lint",
						version,
						"from "+removedPath,
					)
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
						return err
					}
				}
			}

			if prunedCount == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), ui.FormatEmptyNotice("No unused Go or golangci-lint versions to prune."))
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview unused toolchains without removing them")
	return cmd
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
