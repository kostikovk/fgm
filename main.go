package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/koskosovu4/fgm/cmd"
	"github.com/koskosovu4/fgm/internal/app"
	"github.com/koskosovu4/fgm/internal/doctor"
	"github.com/koskosovu4/fgm/internal/envsetup"
	"github.com/koskosovu4/fgm/internal/execenv"
	"github.com/koskosovu4/fgm/internal/goimport"
	"github.com/koskosovu4/fgm/internal/goinstall"
	"github.com/koskosovu4/fgm/internal/golangcilint"
	"github.com/koskosovu4/fgm/internal/golocal"
	"github.com/koskosovu4/fgm/internal/goreleases"
	"github.com/koskosovu4/fgm/internal/lintinstall"
	"github.com/koskosovu4/fgm/internal/resolve"
)

func main() {
	goRoot, err := golocal.DefaultRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	goStore := golocal.New(goRoot, os.Getenv("PATH"))
	goReleaseProvider := goreleases.New(goreleases.Config{
		Client: http.DefaultClient,
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	})
	lintReleaseProvider := golangcilint.New(golangcilint.Config{
		Client: http.DefaultClient,
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	})
	goInstaller := goinstall.New(goinstall.Config{
		Root:           goRoot,
		Client:         http.DefaultClient,
		Provider:       goReleaseProvider,
		ProgressWriter: os.Stderr,
	})
	lintInstaller := lintinstall.New(lintinstall.Config{
		Root:           goRoot,
		Client:         http.DefaultClient,
		Provider:       lintReleaseProvider,
		ProgressWriter: os.Stderr,
	})
	goImporter := goimport.New(goimport.DefaultCandidates(os.Getenv("PATH")), goStore)
	doctorService := doctor.New(goStore, os.Getenv("PATH"))
	executor := execenv.New(resolve.New(goStore), goStore, os.Getenv("PATH"))
	envRenderer := envsetup.New(envsetup.Config{
		GoStore:   goStore,
		ShellPath: os.Getenv("SHELL"),
		GOOS:      runtime.GOOS,
	})
	application := app.New(app.Config{
		Resolver:           resolve.New(goStore),
		GoStore:            goStore,
		GoRemoteProvider:   goReleaseProvider,
		LintRemoteProvider: lintReleaseProvider,
		GoInstaller:        goInstaller,
		LintInstaller:      lintInstaller,
		GoImporter:         goImporter,
		Doctor:             doctorService,
		Executor:           executor,
		EnvRenderer:        envRenderer,
	})

	root := cmd.NewRootCmd(application)
	if err := root.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
