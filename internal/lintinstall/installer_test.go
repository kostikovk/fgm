package lintinstall

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/golangcilint"
)

type stubArchiveProvider struct {
	findArchiveFn func(ctx context.Context, version string) (golangcilint.Archive, error)
}

func (s stubArchiveProvider) FindArchive(ctx context.Context, version string) (golangcilint.Archive, error) {
	return s.findArchiveFn(ctx, version)
}

func TestInstallerInstallLintVersion_DownloadsAndExtractsBinary(t *testing.T) {
	t.Parallel()

	archiveContent := buildTarGzArchive(t, "golangci-lint-2.11.2-darwin-arm64/golangci-lint", []byte("#!/bin/sh\necho lint\n"))
	sum := sha256.Sum256(archiveContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	root := t.TempDir()
	var progress bytes.Buffer
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				if version != "v2.11.2" {
					t.Fatalf("version = %q, want %q", version, "v2.11.2")
				}
				return golangcilint.Archive{
					Version:  "v2.11.2",
					Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   hex.EncodeToString(sum[:]),
				}, nil
			},
		},
		ProgressWriter: &progress,
	})

	installPath, err := installer.InstallLintVersion(context.Background(), "v2.11.2")
	if err != nil {
		t.Fatalf("InstallLintVersion: %v", err)
	}

	expectedBinary := filepath.Join(root, "golangci-lint", "v2.11.2", "golangci-lint")
	if installPath != filepath.Join(root, "golangci-lint", "v2.11.2") {
		t.Fatalf("installPath = %q, want %q", installPath, filepath.Join(root, "golangci-lint", "v2.11.2"))
	}
	if _, err := os.Stat(expectedBinary); err != nil {
		t.Fatalf("stat installed binary: %v", err)
	}
	if !strings.Contains(progress.String(), "Downloading golangci-lint v2.11.2") {
		t.Fatalf("progress = %q, want download message", progress.String())
	}
}

func buildTarGzArchive(t *testing.T, name string, content []byte) []byte {
	t.Helper()

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	header := &tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("Close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("Close gzip writer: %v", err)
	}

	return buffer.Bytes()
}
