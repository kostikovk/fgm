package goinstall

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/koskosovu4/fgm/internal/goreleases"
)

// ArchiveProvider finds downloadable Go archives for specific versions.
type ArchiveProvider interface {
	FindArchive(ctx context.Context, version string) (goreleases.Archive, error)
}

// Installer downloads and extracts Go toolchains into the FGM-managed store.
type Installer struct {
	root           string
	client         *http.Client
	provider       ArchiveProvider
	progressWriter io.Writer
}

// Config configures an Installer.
type Config struct {
	Root           string
	Client         *http.Client
	Provider       ArchiveProvider
	ProgressWriter io.Writer
}

// New constructs an Installer.
func New(config Config) *Installer {
	if config.Client == nil {
		config.Client = http.DefaultClient
	}

	return &Installer{
		root:           config.Root,
		client:         config.Client,
		provider:       config.Provider,
		progressWriter: config.ProgressWriter,
	}
}

// InstallGoVersion downloads and extracts the requested Go version into the managed store.
func (i *Installer) InstallGoVersion(ctx context.Context, version string) (string, error) {
	installDir := filepath.Join(i.root, "go", version)
	goBinaryPath := filepath.Join(installDir, "bin", "go")
	if _, err := os.Stat(goBinaryPath); err == nil {
		return installDir, nil
	}

	archive, err := i.provider.FindArchive(ctx, version)
	if err != nil {
		return "", err
	}

	tempArchive, err := os.CreateTemp("", "fgm-go-archive-*"+archiveFilenameSuffix(archive.Filename))
	if err != nil {
		return "", fmt.Errorf("create temp archive: %w", err)
	}
	defer os.Remove(tempArchive.Name())
	defer tempArchive.Close()

	if err := i.downloadArchive(ctx, archive, tempArchive); err != nil {
		return "", err
	}

	tempExtractDir, err := os.MkdirTemp("", "fgm-go-extract-*")
	if err != nil {
		return "", fmt.Errorf("create temp extract dir: %w", err)
	}
	defer os.RemoveAll(tempExtractDir)

	if err := extractArchive(tempArchive.Name(), tempExtractDir); err != nil {
		return "", err
	}

	extractedGoDir := filepath.Join(tempExtractDir, "go")
	if _, err := os.Stat(extractedGoDir); err != nil {
		return "", fmt.Errorf("find extracted go directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(installDir), 0o755); err != nil {
		return "", fmt.Errorf("create managed Go root: %w", err)
	}
	if err := os.RemoveAll(installDir); err != nil {
		return "", fmt.Errorf("prepare install directory: %w", err)
	}
	if err := os.Rename(extractedGoDir, installDir); err != nil {
		return "", fmt.Errorf("move Go installation into place: %w", err)
	}

	return installDir, nil
}

func (i *Installer) downloadArchive(ctx context.Context, archive goreleases.Archive, destination *os.File) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archive.URL, nil)
	if err != nil {
		return fmt.Errorf("build archive request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("download Go archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download Go archive: unexpected status %s", resp.Status)
	}

	if i.progressWriter != nil {
		_, _ = fmt.Fprintf(i.progressWriter, "Downloading Go %s...\n", archive.Version)
	}

	hash := sha256.New()
	writer := io.MultiWriter(destination, hash)
	if i.progressWriter != nil {
		totalBytes, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		writer = io.MultiWriter(destination, hash, &progressWriter{
			out:     i.progressWriter,
			label:   "download",
			total:   totalBytes,
			written: 0,
			lastPct: -1,
		})
	}

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("write Go archive: %w", err)
	}

	if i.progressWriter != nil {
		_, _ = fmt.Fprintln(i.progressWriter, "Download complete.")
	}

	if archive.SHA256 != "" {
		gotChecksum := hex.EncodeToString(hash.Sum(nil))
		if !strings.EqualFold(gotChecksum, archive.SHA256) {
			return fmt.Errorf("verify Go archive checksum: got %s, want %s", gotChecksum, archive.SHA256)
		}
	}

	if _, err := destination.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind Go archive: %w", err)
	}

	return nil
}

type progressWriter struct {
	out     io.Writer
	label   string
	total   int64
	written int64
	lastPct int
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.written += int64(n)

	if w.total > 0 {
		pct := min(int((w.written*100)/w.total), 100)
		if pct != w.lastPct && pct%10 == 0 {
			w.lastPct = pct
			_, _ = fmt.Fprintf(w.out, "Download progress: %d%%\n", pct)
		}
	}

	return n, nil
}

func archiveFilenameSuffix(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".tar.gz"):
		return ".tar.gz"
	case strings.HasSuffix(filename, ".zip"):
		return ".zip"
	default:
		return ""
	}
}

func extractArchive(archivePath string, destination string) error {
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"):
		return extractTarGz(archivePath, destination)
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZip(archivePath, destination)
	default:
		return fmt.Errorf("unsupported archive format for %s", archivePath)
	}
}

func extractTarGz(archivePath string, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open tar.gz archive: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		targetPath, err := safeJoin(destination, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create tar directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("create tar parent directory: %w", err)
			}
			output, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create tar file: %w", err)
			}
			if _, err := io.Copy(output, tarReader); err != nil {
				output.Close()
				return fmt.Errorf("write tar file: %w", err)
			}
			if err := output.Close(); err != nil {
				return fmt.Errorf("close tar file: %w", err)
			}
		}
	}
}

func extractZip(archivePath string, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath, err := safeJoin(destination, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create zip directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("create zip parent directory: %w", err)
		}

		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("open zip file: %w", err)
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			src.Close()
			return fmt.Errorf("create zip file: %w", err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return fmt.Errorf("write zip file: %w", err)
		}
		if err := src.Close(); err != nil {
			dst.Close()
			return fmt.Errorf("close zip source: %w", err)
		}
		if err := dst.Close(); err != nil {
			return fmt.Errorf("close zip destination: %w", err)
		}
	}

	return nil
}

func safeJoin(base string, name string) (string, error) {
	cleanName := filepath.Clean(name)
	targetPath := filepath.Join(base, cleanName)
	relativePath, err := filepath.Rel(base, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve archive path: %w", err)
	}
	if strings.HasPrefix(relativePath, "..") {
		return "", fmt.Errorf("archive entry %q escapes destination", name)
	}
	return targetPath, nil
}
