package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/koskosovu4/fgm/internal/golangcilint"
)

type manifestGenerator interface {
	Generate(ctx context.Context) (golangcilint.Manifest, error)
}

var (
	stdoutWriter io.Writer = os.Stdout
	stderrWriter io.Writer = os.Stderr
	getEnv                 = os.Getenv
	writeFile              = os.WriteFile
	newGenerator           = func(config golangcilint.GeneratorConfig) manifestGenerator {
		return golangcilint.NewGenerator(config)
	}
	exitMain = os.Exit
)

func main() {
	if err := run(os.Args[1:], stdoutWriter, stderrWriter, getEnv, writeFile, newGenerator); err != nil {
		_, _ = fmt.Fprintln(stderrWriter, err)
		exitMain(1)
	}
}

func run(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	getenv func(string) string,
	writeFile func(string, []byte, os.FileMode) error,
	buildGenerator func(config golangcilint.GeneratorConfig) manifestGenerator,
) error {
	fs := flag.NewFlagSet("update-lint-compatibility", flag.ContinueOnError)
	fs.SetOutput(stderr)

	output := fs.String(
		"output",
		filepath.Join("internal", "golangcilint", "compatibility.json"),
		"output manifest path",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	generator := buildGenerator(golangcilint.GeneratorConfig{
		Client:      http.DefaultClient,
		GitHubToken: getenv("GITHUB_TOKEN"),
	})

	manifest, err := generator.Generate(context.Background())
	if err != nil {
		return err
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')

	if err := writeFile(*output, content, 0o644); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "Updated %s\n", *output); err != nil {
		return err
	}

	return nil
}
