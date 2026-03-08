package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/koskosovu4/fgm/internal/golangcilint"
)

func main() {
	output := flag.String(
		"output",
		filepath.Join("internal", "golangcilint", "compatibility.json"),
		"output manifest path",
	)
	flag.Parse()

	generator := golangcilint.NewGenerator(golangcilint.GeneratorConfig{
		Client:      http.DefaultClient,
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
	})

	manifest, err := generator.Generate(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	content = append(content, '\n')

	if err := os.WriteFile(*output, content, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Updated %s\n", *output)
}
