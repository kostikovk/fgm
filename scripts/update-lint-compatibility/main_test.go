package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/golangcilint"
)

type fakeGenerator struct {
	generate func(context.Context) (golangcilint.Manifest, error)
}

func (f fakeGenerator) Generate(ctx context.Context) (golangcilint.Manifest, error) {
	return f.generate(ctx)
}

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestRunWritesManifest(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "compatibility.json")
	var capturedToken string
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run(
		[]string{"-output", outputPath},
		&stdout,
		&stderr,
		func(key string) string {
			if key == "GITHUB_TOKEN" {
				return "secret-token"
			}
			return ""
		},
		os.WriteFile,
		func(config golangcilint.GeneratorConfig) manifestGenerator {
			capturedToken = config.GitHubToken
			return fakeGenerator{
				generate: func(context.Context) (golangcilint.Manifest, error) {
					return golangcilint.Manifest{}, nil
				},
			}
		},
	)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if capturedToken != "secret-token" {
		t.Fatalf("capturedToken = %q, want %q", capturedToken, "secret-token")
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(content), "\"versions\": null") {
		t.Fatalf("content = %q, want manifest JSON", string(content))
	}
	if stdout.String() != "Updated "+outputPath+"\n" {
		t.Fatalf("stdout = %q, want updated message", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunReturnsGeneratorError(t *testing.T) {
	t.Parallel()

	err := run(
		nil,
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(string) string { return "" },
		os.WriteFile,
		func(golangcilint.GeneratorConfig) manifestGenerator {
			return fakeGenerator{
				generate: func(context.Context) (golangcilint.Manifest, error) {
					return golangcilint.Manifest{}, errors.New("generate failed")
				},
			}
		},
	)
	if err == nil || !strings.Contains(err.Error(), "generate failed") {
		t.Fatalf("err = %v, want generator error", err)
	}
}

func TestRunReturnsWriteFileError(t *testing.T) {
	t.Parallel()

	err := run(
		nil,
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(string) string { return "" },
		func(string, []byte, os.FileMode) error { return errors.New("write failed") },
		func(golangcilint.GeneratorConfig) manifestGenerator {
			return fakeGenerator{
				generate: func(context.Context) (golangcilint.Manifest, error) {
					return golangcilint.Manifest{}, nil
				},
			}
		},
	)
	if err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("err = %v, want write error", err)
	}
}

func TestRunReturnsStdoutError(t *testing.T) {
	t.Parallel()

	err := run(
		nil,
		failingWriter{},
		&bytes.Buffer{},
		func(string) string { return "" },
		func(string, []byte, os.FileMode) error { return nil },
		func(golangcilint.GeneratorConfig) manifestGenerator {
			return fakeGenerator{
				generate: func(context.Context) (golangcilint.Manifest, error) {
					return golangcilint.Manifest{}, nil
				},
			}
		},
	)
	if err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("err = %v, want stdout write error", err)
	}
}

func TestMainWritesErrorAndExits(t *testing.T) {
	oldStdout := stdoutWriter
	oldStderr := stderrWriter
	oldGetEnv := getEnv
	oldWriteFile := writeFile
	oldNewGenerator := newGenerator
	oldExit := exitMain
	oldArgs := os.Args
	defer func() {
		stdoutWriter = oldStdout
		stderrWriter = oldStderr
		getEnv = oldGetEnv
		writeFile = oldWriteFile
		newGenerator = oldNewGenerator
		exitMain = oldExit
		os.Args = oldArgs
	}()

	var stderr bytes.Buffer
	stdoutWriter = &bytes.Buffer{}
	stderrWriter = &stderr
	getEnv = func(string) string { return "" }
	writeFile = os.WriteFile
	newGenerator = func(golangcilint.GeneratorConfig) manifestGenerator {
		return fakeGenerator{
			generate: func(context.Context) (golangcilint.Manifest, error) {
				return golangcilint.Manifest{}, errors.New("generate failed")
			},
		}
	}
	exitCode := 0
	exitMain = func(code int) {
		exitCode = code
		panic("exit")
	}
	os.Args = []string{"update-lint-compatibility"}

	defer func() {
		if recovered := recover(); recovered != "exit" {
			t.Fatalf("recovered = %v, want exit panic", recovered)
		}
		if exitCode != 1 {
			t.Fatalf("exitCode = %d, want 1", exitCode)
		}
		if stderr.String() != "generate failed\n" {
			t.Fatalf("stderr = %q, want %q", stderr.String(), "generate failed\n")
		}
	}()

	main()
	t.Fatal("main should have exited")
}
