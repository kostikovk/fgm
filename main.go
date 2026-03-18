package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/kostikovk/fgm/cmd"
	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/currenttoolchain"
	"github.com/kostikovk/fgm/internal/doctor"
	"github.com/kostikovk/fgm/internal/envsetup"
	"github.com/kostikovk/fgm/internal/execenv"
	"github.com/kostikovk/fgm/internal/goimport"
	"github.com/kostikovk/fgm/internal/goinstall"
	"github.com/kostikovk/fgm/internal/golangcilint"
	"github.com/kostikovk/fgm/internal/golocal"
	"github.com/kostikovk/fgm/internal/goreleases"
	"github.com/kostikovk/fgm/internal/goupgrade"
	"github.com/kostikovk/fgm/internal/lintconfig"
	"github.com/kostikovk/fgm/internal/lintimport"
	"github.com/kostikovk/fgm/internal/lintinstall"
	"github.com/kostikovk/fgm/internal/lintlocal"
	"github.com/kostikovk/fgm/internal/lintupgrade"
	"github.com/kostikovk/fgm/internal/resolve"
)

type rootCommand interface {
	ExecuteContext(ctx context.Context) error
}

var (
	buildVersion            = "dev"
	buildCommit             = "unknown"
	buildDate               = "unknown"
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
	lintConfigService, err := lintconfig.New(lintconfig.Config{
		Resolver: currentResolver,
	})
	if err != nil {
		return err
	}
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
		LintUpgrader: lintupgrade.New(lintupgrade.Config{
			Resolver:       currentResolver,
			RemoteProvider: lintReleaseProvider,
			Installer:      lintInstaller,
		}),
		BuildInfo: app.BuildInfo{
			Version: buildVersion,
			Commit:  buildCommit,
			Date:    buildDate,
		},
		Doctor:              doctorService,
		Executor:            executor,
		EnvRenderer:         envRenderer,
		ProfileInstaller:    envRenderer,
		LintConfigGenerator: lintConfigService,
		LintDoctor:          lintConfigService,
	}

	return buildRoot(application).ExecuteContext(ctx)
}
