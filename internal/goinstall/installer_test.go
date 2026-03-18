package goinstall

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
	"strconv"
	"strings"
	"testing"

	"github.com/kostikovk/fgm/internal/goreleases"
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

func TestInstallerNew_DefaultsHttpClient(t *testing.T) {
	t.Parallel()

	installer := New(Config{
		Root:     t.TempDir(),
		Provider: stubArchiveProvider{},
	})
	if installer == nil {
		t.Fatalf("New returned nil")
	}
	if installer.client == nil {
		t.Fatalf("client is nil, want http.DefaultClient")
	}
}

func TestInstallGoVersion_ReturnsExistingWhenAlreadyInstalled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	version := "1.25.7"
	goBinaryPath := filepath.Join(root, "go", version, "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBinaryPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(goBinaryPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	called := false
	installer := New(Config{
		Root: root,
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				called = true
				return goreleases.Archive{}, fmt.Errorf("should not be called")
			},
		},
	})

	installPath, err := installer.InstallGoVersion(context.Background(), version)
	if err != nil {
		t.Fatalf("InstallGoVersion: %v", err)
	}
	if called {
		t.Fatalf("provider was called, but binary already existed")
	}
	if installPath != filepath.Join(root, "go", version) {
		t.Fatalf("installPath = %q, want %q", installPath, filepath.Join(root, "go", version))
	}
}

func TestDownloadArchive_ErrorForNon200Status(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.25.7.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.25.7.darwin-arm64.tar.gz",
					SHA256:   "abc123",
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.25.7")
	if err == nil {
		t.Fatalf("expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "unexpected status")
	}
}

func TestDownloadArchive_ErrorForBadChecksum(t *testing.T) {
	t.Parallel()

	archiveBytes := createTarGzArchive(t, map[string]string{
		"go/bin/go": "#!/bin/sh\n",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(archiveBytes)))
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.25.7.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.25.7.darwin-arm64.tar.gz",
					SHA256:   "badchecksum",
				}, nil
			},
		},
		ProgressWriter: io.Discard,
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.25.7")
	if err == nil {
		t.Fatalf("expected error for bad checksum, got nil")
	}
	if !strings.Contains(err.Error(), "verify Go archive checksum") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "verify Go archive checksum")
	}
}

func TestDownloadArchive_SkipsChecksumWhenEmpty(t *testing.T) {
	t.Parallel()

	archiveBytes := createTarGzArchive(t, map[string]string{
		"go/bin/go": "#!/bin/sh\n",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(archiveBytes)))
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.25.7.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.25.7.darwin-arm64.tar.gz",
					SHA256:   "",
				}, nil
			},
		},
		ProgressWriter: io.Discard,
	})

	installPath, err := installer.InstallGoVersion(context.Background(), "1.25.7")
	if err != nil {
		t.Fatalf("InstallGoVersion: %v", err)
	}
	if _, err := os.Stat(filepath.Join(installPath, "bin", "go")); err != nil {
		t.Fatalf("stat extracted go binary: %v", err)
	}
}

func TestDownloadArchive_WithoutProgressWriter(t *testing.T) {
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
		ProgressWriter: nil,
	})

	installPath, err := installer.InstallGoVersion(context.Background(), "1.25.7")
	if err != nil {
		t.Fatalf("InstallGoVersion: %v", err)
	}
	if _, err := os.Stat(filepath.Join(installPath, "bin", "go")); err != nil {
		t.Fatalf("stat extracted go binary: %v", err)
	}
}

func TestInstallGoVersion_CreateTempArchiveError(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(tmpFile, []byte("block"), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	t.Setenv("TMPDIR", tmpFile)

	installer := New(Config{
		Root: t.TempDir(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.25.7.darwin-arm64.tar.gz",
					URL:      "https://example.test/archive.tar.gz",
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.25.7")
	if err == nil {
		t.Fatal("expected create temp archive error")
	}
	if !strings.Contains(err.Error(), "create temp archive") {
		t.Fatalf("error = %q, want create temp archive error", err)
	}
}

func TestInstallGoVersion_CreateTempExtractDirError(t *testing.T) {
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

	archiveBytes := createTarGzArchive(t, map[string]string{
		"go/bin/go": "#!/bin/sh\n",
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = os.Setenv("TMPDIR", blockedTmp)
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	installer := New(Config{
		Root:   t.TempDir(),
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.25.7.darwin-arm64.tar.gz",
					URL:      server.URL,
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.25.7")
	if err == nil {
		t.Fatal("expected create temp extract dir error")
	}
	if !strings.Contains(err.Error(), "create temp extract dir") {
		t.Fatalf("error = %q, want create temp extract dir error", err)
	}
}

func TestDownloadArchive_RewindError(t *testing.T) {
	t.Parallel()

	archiveBytes := createTarGzArchive(t, map[string]string{
		"go/bin/go": "#!/bin/sh\n",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveBytes)
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

	err = installer.downloadArchive(context.Background(), goreleases.Archive{
		Version:  "1.25.7",
		Filename: "go1.25.7.darwin-arm64.tar.gz",
		URL:      server.URL,
	}, writer)
	_ = writer.Close()
	<-done
	if err == nil {
		t.Fatal("expected rewind error")
	}
	if !strings.Contains(err.Error(), "rewind Go archive") {
		t.Fatalf("error = %q, want rewind error", err)
	}
}

func TestInstallGoVersion_FindArchiveError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	installer := New(Config{
		Root: root,
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{}, fmt.Errorf("archive not found")
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatalf("expected error from FindArchive, got nil")
	}
	if !strings.Contains(err.Error(), "archive not found") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "archive not found")
	}
}

func TestInstallGoVersion_ExtractError(t *testing.T) {
	t.Parallel()

	// Serve corrupt data that is not a valid tar.gz archive.
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
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.99.0.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.99.0.darwin-arm64.tar.gz",
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatalf("expected error from archive.Extract, got nil")
	}
}

func TestInstallGoVersion_ExtractedDirNotFound(t *testing.T) {
	t.Parallel()

	// Create a valid tar.gz archive that does NOT contain a "go/" directory.
	archiveBytes := createTarGzArchive(t, map[string]string{
		"notgo/bin/go": "#!/bin/sh\n",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.99.0.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.99.0.darwin-arm64.tar.gz",
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatalf("expected error for missing extracted go directory, got nil")
	}
	if !strings.Contains(err.Error(), "find extracted go directory") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "find extracted go directory")
	}
}

func TestInstallGoVersion_MkdirAllError(t *testing.T) {
	t.Parallel()

	archiveBytes := createTarGzArchive(t, map[string]string{
		"go/bin/go": "#!/bin/sh\n",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	// Use a file (not a directory) as root so MkdirAll for filepath.Dir(installDir) fails.
	root := t.TempDir()
	blockingFile := filepath.Join(root, "go")
	if err := os.WriteFile(blockingFile, []byte("not a dir"), 0o444); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.99.0.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.99.0.darwin-arm64.tar.gz",
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatalf("expected error from MkdirAll, got nil")
	}
	if !strings.Contains(err.Error(), "create managed Go root") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "create managed Go root")
	}
}

func TestInstallGoVersion_RenameError(t *testing.T) {
	t.Parallel()

	archiveBytes := createTarGzArchive(t, map[string]string{
		"go/bin/go": "#!/bin/sh\n",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	// Make root/go read-only so that Rename(extractedGoDir, installDir) fails.
	// installDir does not exist, so RemoveAll is a no-op. MkdirAll succeeds because
	// root/go already exists. Rename fails because root/go is not writable.
	root := t.TempDir()
	parentDir := filepath.Join(root, "go")
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.99.0.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.99.0.darwin-arm64.tar.gz",
					SHA256:   "",
				}, nil
			},
		},
	})

	if err := os.Chmod(parentDir, 0o555); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	defer func() { _ = os.Chmod(parentDir, 0o755) }()

	_, err := installer.InstallGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatalf("expected error from Rename, got nil")
	}
	// Could be "move Go installation into place" or "prepare install directory" depending on what fails first.
	if !strings.Contains(err.Error(), "move Go installation into place") && !strings.Contains(err.Error(), "prepare install directory") {
		t.Fatalf("error = %q, want it to contain rename-related message", err.Error())
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
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.99.0.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.99.0.darwin-arm64.tar.gz",
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatalf("expected error from HTTP Do, got nil")
	}
	if !strings.Contains(err.Error(), "download Go archive") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "download Go archive")
	}
}

func TestDownloadArchive_IoCopyError(t *testing.T) {
	t.Parallel()

	// Serve a response where Content-Length is larger than the body, causing
	// an unexpected EOF during io.Copy.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "999999")
		_, _ = w.Write([]byte("short"))
		// Close connection prematurely.
	}))
	defer server.Close()

	root := t.TempDir()
	installer := New(Config{
		Root:   root,
		Client: server.Client(),
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.99.0.darwin-arm64.tar.gz",
					URL:      server.URL + "/go1.99.0.darwin-arm64.tar.gz",
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatalf("expected error from io.Copy, got nil")
	}
	if !strings.Contains(err.Error(), "write Go archive") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "write Go archive")
	}
}

func TestDownloadArchive_InvalidURL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	installer := New(Config{
		Root: root,
		Provider: stubArchiveProvider{
			findArchiveFn: func(ctx context.Context, version string) (goreleases.Archive, error) {
				return goreleases.Archive{
					Version:  version,
					Filename: "go1.99.0.darwin-arm64.tar.gz",
					URL:      "://invalid-url",
					SHA256:   "",
				}, nil
			},
		},
	})

	_, err := installer.InstallGoVersion(context.Background(), "1.99.0")
	if err == nil {
		t.Fatalf("expected error for invalid URL, got nil")
	}
	if !strings.Contains(err.Error(), "build archive request") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "build archive request")
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
