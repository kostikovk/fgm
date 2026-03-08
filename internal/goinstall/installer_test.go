package goinstall

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
	"strconv"
	"strings"
	"testing"

	"github.com/koskosovu4/fgm/internal/goreleases"
)

type stubArchiveProvider struct {
	findArchiveFn func(ctx context.Context, version string) (goreleases.Archive, error)
}

func (s stubArchiveProvider) FindArchive(ctx context.Context, version string) (goreleases.Archive, error) {
	return s.findArchiveFn(ctx, version)
}

func TestInstallerInstallGoVersion_DownloadsAndExtractsArchive(t *testing.T) {
	t.Parallel()

	archiveBytes := createTarGzArchive(t, map[string]string{
		"go/bin/go": "#!/bin/sh\n",
	})
	checksum := sha256.Sum256(archiveBytes)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(archiveBytes)))
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	root := t.TempDir()
	var progress bytes.Buffer
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.25.7.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.25.7.darwin-arm64.tar.gz",
					SHA256:   hex.EncodeToString(checksum[:]),
				}, nil
			},
		},
		ProgressWriter: &progress,
	})

	installPath, err := installer.InstallGoVersion(context.Background(), "1.25.7")
	if err != nil {
		t.Fatalf("InstallGoVersion: %v", err)
	}

	if installPath != filepath.Join(root, "go", "1.25.7") {
		t.Fatalf("installPath = %q, want %q", installPath, filepath.Join(root, "go", "1.25.7"))
	}
	if _, err := os.Stat(filepath.Join(installPath, "bin", "go")); err != nil {
		t.Fatalf("stat extracted go binary: %v", err)
	}
	if !strings.Contains(progress.String(), "Downloading Go 1.25.7") {
		t.Fatalf("progress = %q, want start line", progress.String())
	}
	if !strings.Contains(progress.String(), "Download complete.") {
		t.Fatalf("progress = %q, want completion line", progress.String())
	}
}

func createTarGzArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

	buffer := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("WriteHeader: %v", err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("Close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("Close gzip writer: %v", err)
	}

	return buffer.Bytes()
}
