package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/koskosovu4/fgm/cmd"
	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/currenttoolchain"
	"github.com/koskosovu4/fgm/internal/doctor"
	"github.com/koskosovu4/fgm/internal/envsetup"
	"github.com/koskosovu4/fgm/internal/execenv"
	"github.com/koskosovu4/fgm/internal/goimport"
	"github.com/koskosovu4/fgm/internal/goinstall"
	"github.com/koskosovu4/fgm/internal/golangcilint"
	"github.com/koskosovu4/fgm/internal/golocal"
	"github.com/koskosovu4/fgm/internal/goreleases"
	"github.com/koskosovu4/fgm/internal/goupgrade"
	"github.com/koskosovu4/fgm/internal/lintimport"
	"github.com/koskosovu4/fgm/internal/lintinstall"
	"github.com/koskosovu4/fgm/internal/lintlocal"
	"github.com/koskosovu4/fgm/internal/resolve"
)

type rootCommand interface {
	ExecuteContext(ctx context.Context) error
}

var (
	defaultGoRoot           = golocal.DefaultRoot
	getEnv                  = os.Getenv
	stderrWriter  io.Writer = os.Stderr
	newRootCmd              = func(application *app.App) rootCommand {
		return cmd.NewRootCmd(application)
	}
	exitMain = os.Exit
)

func main() {
	if err := run(context.Background(), stderrWriter, getEnv, defaultGoRoot, newRootCmd); err != nil {
		_, _ = fmt.Fprintln(stderrWriter, err)
		exitMain(1)
	}
}

func run(
	ctx context.Context,
	stderr io.Writer,
	getenv func(string) string,
	resolveRoot func() (string, error),
	buildRoot func(application *app.App) rootCommand,
) error {
	goRoot, err := resolveRoot()
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: 5 * time.Minute}

	goStore := golocal.New(goRoot, getenv("PATH"))
	lintStore := lintlocal.New(goRoot)
	goReleaseProvider := goreleases.New(goreleases.Config{
		Client: httpClient,
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	})
	lintReleaseProvider := golangcilint.New(golangcilint.Config{
		Client: httpClient,
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	})
	goInstaller := goinstall.New(goinstall.Config{
		Root:           goRoot,
		Client:         http.DefaultClient,
		Provider:       goReleaseProvider,
		ProgressWriter: stderr,
	})
	lintInstaller := lintinstall.New(lintinstall.Config{
		Root:           goRoot,
		Client:         http.DefaultClient,
		Provider:       lintReleaseProvider,
		ProgressWriter: stderr,
	})
	goImporter := goimport.New(goimport.Config{
		Candidates: goimport.DefaultCandidates(getenv("PATH")),
		Registry:   goStore,
	})
	lintImporter := lintimport.New(lintimport.Config{
		Candidates: lintimport.DefaultCandidates(getenv("PATH")),
		Registry:   lintStore,
	})
	goUpgrader := goupgrade.New(goupgrade.Config{
		RemoteProvider:     goReleaseProvider,
		Installer:          goInstaller,
		LintRemoteProvider: lintReleaseProvider,
		LintInstaller:      lintInstaller,
		GlobalStore:        goStore,
	})
	currentResolver := currenttoolchain.New(currenttoolchain.Config{
		GoResolver:         resolve.New(goStore),
		LintStore:          lintStore,
		LintRemoteProvider: lintReleaseProvider,
	})
	doctorService := doctor.New(doctor.Config{
		Resolver:           currentResolver,
		GoStore:            goStore,
		LintStore:          lintStore,
		LintRemoteProvider: lintReleaseProvider,
		PathEnv:            getenv("PATH"),
	})
	executor := execenv.New(execenv.Config{
		Resolver:    currentResolver,
		GoLocator:   goStore,
		LintLocator: lintStore,
		PathEnv:     getenv("PATH"),
	})
	envRenderer := envsetup.New(envsetup.Config{
		GoStore:   goStore,
		ShellPath: getenv("SHELL"),
		GOOS:      runtime.GOOS,
	})
	application := &app.App{
		Resolver:           currentResolver,
		GoStore:            goStore,
		LintStore:          lintStore,
		GoRemoteProvider:   goReleaseProvider,
		LintRemoteProvider: lintReleaseProvider,
		GoInstaller:        goInstaller,
		LintInstaller:      lintInstaller,
		GoImporter:         goImporter,
		LintImporter:       lintImporter,
		GoUpgrader:         goUpgrader,
		Doctor:             doctorService,
		Executor:           executor,
		EnvRenderer:        envRenderer,
	}

	return buildRoot(application).ExecuteContext(ctx)
}
