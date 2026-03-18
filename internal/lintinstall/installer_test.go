package lintinstall

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/golangcilint"
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

func TestInstallerNew_DefaultsHttpClient(t *testing.T) {
	t.Parallel()

	installer := New(Config{
		Root:     t.TempDir(),
		Provider: stubArchiveProvider{},
	})
	if installer == nil {
		t.Fatalf("New returned nil")
	}
	if installer.client != http.DefaultClient {
		t.Fatalf("client = %v, want http.DefaultClient", installer.client)
	}
}

func TestInstallLintVersion_ReturnsExistingWhenAlreadyInstalled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	version := "v2.11.2"
	installDir := filepath.Join(root, "golangci-lint", version)
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	binaryPath := filepath.Join(installDir, "golangci-lint")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho lint\n"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	installer := New(Config{
		Root: root,
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				t.Fatalf("FindArchive should not be called when binary already exists")
				return golangcilint.Archive{}, nil
			},
		},
	})

	result, err := installer.InstallLintVersion(context.Background(), version)
	if err != nil {
		t.Fatalf("InstallLintVersion: %v", err)
	}
	if result != installDir {
		t.Fatalf("installPath = %q, want %q", result, installDir)
	}
}

func TestDownloadArchive_ErrorForNon200Status(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v2.11.2",
					Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   "abc123",
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatalf("expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "unexpected status")
	}
}

func TestDownloadArchive_ErrorForBadChecksum(t *testing.T) {
	t.Parallel()

	archiveContent := buildTarGzArchive(t, "golangci-lint-2.11.2-darwin-arm64/golangci-lint", []byte("#!/bin/sh\necho lint\n"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v2.11.2",
					Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   "badchecksum",
				}, nil
			},
		},
		ProgressWriter: &bytes.Buffer{},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatalf("expected error for bad checksum")
	}
	if !strings.Contains(err.Error(), "verify golangci-lint archive checksum") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "verify golangci-lint archive checksum")
	}
}

func TestDownloadArchive_SkipsChecksumWhenEmpty(t *testing.T) {
	t.Parallel()

	archiveContent := buildTarGzArchive(t, "golangci-lint-2.11.2-darwin-arm64/golangci-lint", []byte("#!/bin/sh\necho lint\n"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v2.11.2",
					Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   "",
				}, nil
			},
		},
		ProgressWriter: &bytes.Buffer{},
	})

	installPath, err := installer.InstallLintVersion(context.Background(), "v2.11.2")
	if err != nil {
		t.Fatalf("InstallLintVersion: %v", err)
	}

	expectedBinary := filepath.Join(root, "golangci-lint", "v2.11.2", "golangci-lint")
	if _, err := os.Stat(expectedBinary); err != nil {
		t.Fatalf("stat installed binary: %v", err)
	}
	if installPath != filepath.Join(root, "golangci-lint", "v2.11.2") {
		t.Fatalf("installPath = %q, want %q", installPath, filepath.Join(root, "golangci-lint", "v2.11.2"))
	}
}

func TestFindLintBinary_ErrorWhenNotFound(t *testing.T) {
	t.Parallel()

	archiveContent := buildTarGzArchive(t, "golangci-lint-2.11.2-darwin-arm64/other-binary", []byte("#!/bin/sh\necho other\n"))
	sum := sha256.Sum256(archiveContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v2.11.2",
					Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   hex.EncodeToString(sum[:]),
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatalf("expected error when binary not found in archive")
	}
	if !strings.Contains(err.Error(), "binary not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "binary not found")
	}
}

func TestDownloadArchive_WithoutProgressWriter(t *testing.T) {
	t.Parallel()

	archiveContent := buildTarGzArchive(t, "golangci-lint-2.11.2-darwin-arm64/golangci-lint", []byte("#!/bin/sh\necho lint\n"))
	sum := sha256.Sum256(archiveContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:           root,
		Client:         server.Client(),
		ProgressWriter: nil,
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v2.11.2",
					Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   hex.EncodeToString(sum[:]),
				}, nil
			},
		},
	})

	installPath, err := installer.InstallLintVersion(context.Background(), "v2.11.2")
	if err != nil {
		t.Fatalf("InstallLintVersion: %v", err)
	}

	expectedBinary := filepath.Join(root, "golangci-lint", "v2.11.2", "golangci-lint")
	if _, err := os.Stat(expectedBinary); err != nil {
		t.Fatalf("stat installed binary: %v", err)
	}
	if installPath != filepath.Join(root, "golangci-lint", "v2.11.2") {
		t.Fatalf("installPath = %q, want %q", installPath, filepath.Join(root, "golangci-lint", "v2.11.2"))
	}
}

func TestInstallLintVersion_CreateTempArchiveError(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(tmpFile, []byte("block"), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	t.Setenv("TMPDIR", tmpFile)

	installer := New(Config{
		Root: t.TempDir(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  version,
					Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
					URL:      "https://example.test/archive.tar.gz",
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatal("expected create temp archive error")
	}
	if !strings.Contains(err.Error(), "create temp archive") {
		t.Fatalf("error = %q, want create temp archive error", err)
	}
}

func TestInstallLintVersion_CreateTempExtractDirError(t *testing.T) {
	writableTmp := t.TempDir()
	blockedTmp := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blockedTmp, []byte("block"), 0o644); err != nil {
		t.Fatalf("write blocked tmp file: %v", err)
	}

	originalTmp := os.Getenv("TMPDIR")
	if err := os.Setenv("TMPDIR", writableTmp); err != nil {
		t.Fatalf("set TMPDIR: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("TMPDIR", originalTmp)
	})

	archiveContent := buildTarGzArchive(t, "golangci-lint-2.11.2-darwin-arm64/golangci-lint", []byte("#!/bin/sh\n"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = os.Setenv("TMPDIR", blockedTmp)
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	installer := New(Config{
		Root:   t.TempDir(),
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  version,
					Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
					URL:      server.URL,
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v2.11.2")
	if err == nil {
		t.Fatal("expected create temp extract dir error")
	}
	if !strings.Contains(err.Error(), "create temp extract dir") {
		t.Fatalf("error = %q, want create temp extract dir error", err)
	}
}

func TestDownloadLintArchive_RewindError(t *testing.T) {
	t.Parallel()

	archiveContent := buildTarGzArchive(t, "golangci-lint-2.11.2-darwin-arm64/golangci-lint", []byte("#!/bin/sh\n"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	installer := New(Config{
		Root:   t.TempDir(),
		Client: server.Client(),
	})

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	defer func() { _ = reader.Close() }()
	defer func() { _ = writer.Close() }()

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, reader)
		close(done)
	}()

	err = installer.downloadArchive(context.Background(), golangcilint.Archive{
		Version:  "v2.11.2",
		Filename: "golangci-lint-2.11.2-darwin-arm64.tar.gz",
		URL:      server.URL,
		SHA256:   "",
	}, writer)
	_ = writer.Close()
	<-done
	if err == nil {
		t.Fatal("expected rewind error")
	}
	if !strings.Contains(err.Error(), "rewind golangci-lint archive") {
		t.Fatalf("error = %q, want rewind error", err)
	}
}

func TestInstallLintVersion_FindArchiveError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	installer := New(Config{
		Root: root,
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{}, fmt.Errorf("archive not found")
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v9.99.0")
	if err == nil {
		t.Fatalf("expected error from FindArchive, got nil")
	}
	if !strings.Contains(err.Error(), "archive not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "archive not found")
	}
}

func TestInstallLintVersion_ExtractError(t *testing.T) {
	t.Parallel()

	corruptData := []byte("this is not a valid tar.gz archive")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(corruptData)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v9.99.0",
					Filename: "golangci-lint-9.99.0-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v9.99.0")
	if err == nil {
		t.Fatalf("expected error from archive.Extract, got nil")
	}
}

func TestInstallLintVersion_MkdirAllError(t *testing.T) {
	t.Parallel()

	archiveContent := buildTarGzArchive(t, "golangci-lint-9.99.0-darwin-arm64/golangci-lint", []byte("#!/bin/sh\n"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	// Place a file at the path where MkdirAll(installDir) needs to create a directory.
	root := t.TempDir()
	blockingFile := filepath.Join(root, "golangci-lint")
	if err := os.WriteFile(blockingFile, []byte("not a dir"), 0o444); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v9.99.0",
					Filename: "golangci-lint-9.99.0-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v9.99.0")
	if err == nil {
		t.Fatalf("expected error from MkdirAll, got nil")
	}
	// Could be "create install directory" or "prepare install directory" depending on which MkdirAll/RemoveAll fails.
	if !strings.Contains(err.Error(), "install directory") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "install directory")
	}
}

func TestInstallLintVersion_RemoveAllError(t *testing.T) {
	t.Parallel()

	archiveContent := buildTarGzArchive(t, "golangci-lint-9.99.0-darwin-arm64/golangci-lint", []byte("#!/bin/sh\n"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveContent)
	}))
	defer server.Close()

	// Create the parent directory as read-only so RemoveAll(installDir) fails when
	// installDir exists inside a non-writable parent.
	root := t.TempDir()
	parentDir := filepath.Join(root, "golangci-lint")
	installDir := filepath.Join(parentDir, "v9.99.0")
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Place a file inside installDir and make parentDir read-only.
	if err := os.WriteFile(filepath.Join(installDir, "dummy"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Chmod(parentDir, 0o555); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	defer func() { _ = os.Chmod(parentDir, 0o755) }()

	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v9.99.0",
					Filename: "golangci-lint-9.99.0-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v9.99.0")
	if err == nil {
		t.Fatalf("expected error from RemoveAll, got nil")
	}
	if !strings.Contains(err.Error(), "prepare install directory") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "prepare install directory")
	}
}

func TestFindLintBinary_ErrorWhenDirNotReadable(t *testing.T) {
	t.Parallel()

	// Create a directory tree with an unreadable subdirectory to trigger
	// the error return inside WalkDir's callback (line 159-161).
	root := t.TempDir()
	subDir := filepath.Join(root, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Chmod(subDir, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	defer func() { _ = os.Chmod(subDir, 0o755) }()

	_, err := findLintBinary(root)
	if err == nil {
		t.Fatalf("expected error from findLintBinary on unreadable dir, got nil")
	}
	if !strings.Contains(err.Error(), "scan extracted golangci-lint archive") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "scan extracted golangci-lint archive")
	}
}

func TestDownloadArchive_HttpDoError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // Close immediately so requests fail.

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v9.99.0",
					Filename: "golangci-lint-9.99.0-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v9.99.0")
	if err == nil {
		t.Fatalf("expected error from HTTP Do, got nil")
	}
	if !strings.Contains(err.Error(), "download golangci-lint archive") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "download golangci-lint archive")
	}
}

func TestDownloadArchive_IoCopyError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "999999")
		_, _ = w.Write([]byte("short"))
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v9.99.0",
					Filename: "golangci-lint-9.99.0-darwin-arm64.tar.gz",
					URL:      server.URL,
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v9.99.0")
	if err == nil {
		t.Fatalf("expected error from io.Copy, got nil")
	}
	if !strings.Contains(err.Error(), "write golangci-lint archive") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "write golangci-lint archive")
	}
}

func TestDownloadArchive_InvalidURL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	installer := New(Config{
		Root: root,
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (golangcilint.Archive, error) {
				return golangcilint.Archive{
					Version:  "v9.99.0",
					Filename: "golangci-lint-9.99.0-darwin-arm64.tar.gz",
					URL:      "://invalid-url",
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallLintVersion(context.Background(), "v9.99.0")
	if err == nil {
		t.Fatalf("expected error for invalid URL, got nil")
	}
	if !strings.Contains(err.Error(), "build archive request") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "build archive request")
	}
}

func TestFindLintBinary_WalkDirError(t *testing.T) {
	t.Parallel()

	// Call findLintBinary on a non-existent directory to trigger WalkDir error.
	_, err := findLintBinary("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatalf("expected error from findLintBinary on non-existent dir, got nil")
	}
	if !strings.Contains(err.Error(), "scan extracted golangci-lint archive") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "scan extracted golangci-lint archive")
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
