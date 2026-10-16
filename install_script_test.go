package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func fakeReleaseArchive(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	tw := tar.NewWriter(gz)
	script := []byte("#!/bin/sh\necho fgm-test\n")
	if err := tw.WriteHeader(&tar.Header{Name: "fgm", Mode: 0o755, Size: int64(len(script))}); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(script); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buffer.Bytes()
}

func TestInstallScript(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("install.sh is POSIX-only")
	}
	archiveOS := runtime.GOOS
	archiveArch := runtime.GOARCH
	archiveName := fmt.Sprintf("fgm_1.2.3_%s_%s.tar.gz", archiveOS, archiveArch)

	runInstall := func(t *testing.T, checksum func(archive []byte) string) (string, string, error) {
		t.Helper()
		archive := fakeReleaseArchive(t)
		var mu sync.Mutex
		var requested []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			requested = append(requested, r.URL.Path)
			mu.Unlock()
			switch r.URL.Path {
			case "/v1.2.3/" + archiveName:
				_, _ = w.Write(archive)
			case "/v1.2.3/checksums.txt":
				_, _ = fmt.Fprintf(w, "%s  %s\n", checksum(archive), archiveName)
			default:
				http.NotFound(w, r)
			}
		}))
		t.Cleanup(server.Close)

		installDir := t.TempDir()
		cmd := exec.Command("sh", "install.sh")
		cmd.Env = append(os.Environ(),
			"FGM_VERSION=v1.2.3",
			"FGM_RELEASE_BASE_URL="+server.URL,
			"INSTALL_DIR="+installDir,
		)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		mu.Lock()
		defer mu.Unlock()
		if len(requested) == 0 {
			t.Fatalf("install.sh did not use FGM_RELEASE_BASE_URL")
		}
		return installDir, stderr.String(), err
	}

	t.Run("install.sh", func(t *testing.T) {
		t.Parallel()

		t.Run("should install the pinned version when the checksum matches", func(t *testing.T) {
			t.Parallel()

			// Arrange
			valid := func(archive []byte) string {
				sum := sha256.Sum256(archive)
				return hex.EncodeToString(sum[:])
			}

			// Act
			installDir, stderr, err := runInstall(t, valid)

			// Assert
			if err != nil {
				t.Fatalf("install.sh: %v\nstderr:\n%s", err, stderr)
			}
			if _, err := os.Stat(filepath.Join(installDir, "fgm")); err != nil {
				t.Errorf("installed binary missing: %v", err)
			}
		})

		t.Run("should fail without installing when the checksum does not match", func(t *testing.T) {
			t.Parallel()

			// Arrange
			wrong := func([]byte) string { return strings.Repeat("0", 64) }

			// Act
			installDir, stderr, err := runInstall(t, wrong)

			// Assert
			if err == nil {
				t.Fatal("install.sh succeeded; want a checksum failure")
			}
			if !strings.Contains(stderr, "checksum mismatch") {
				t.Errorf("stderr = %q; want a checksum mismatch message", stderr)
			}
			if _, err := os.Stat(filepath.Join(installDir, "fgm")); !os.IsNotExist(err) {
				t.Errorf("binary installed despite mismatch (stat err %v)", err)
			}
		})
	})
}
