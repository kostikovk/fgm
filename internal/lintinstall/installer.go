package lintinstall

import (
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

	"github.com/koskosovu4/fgm/internal/archive"
	"github.com/koskosovu4/fgm/internal/golangcilint"
)

// ArchiveProvider finds downloadable golangci-lint archives for specific versions.
type ArchiveProvider interface {
	FindArchive(ctx context.Context, version string) (golangcilint.Archive, error)
}

// Installer downloads and extracts golangci-lint into the FGM-managed store.
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
	client := config.Client
	if client == nil {
		client = http.DefaultClient
	}

	return &Installer{
		root:           config.Root,
		client:         client,
		provider:       config.Provider,
		progressWriter: config.ProgressWriter,
	}
}

// InstallLintVersion downloads and extracts the requested golangci-lint version into the managed store.
func (i *Installer) InstallLintVersion(ctx context.Context, version string) (string, error) {
	installDir := filepath.Join(i.root, "golangci-lint", version)
	binaryPath := filepath.Join(installDir, "golangci-lint")
	if _, err := os.Stat(binaryPath); err == nil {
		return installDir, nil
	}

	arch, err := i.provider.FindArchive(ctx, version)
	if err != nil {
		return "", err
	}

	tempArchive, err := os.CreateTemp("", "fgm-lint-archive-*"+archive.FilenameSuffix(arch.Filename))
	if err != nil {
		return "", fmt.Errorf("create temp archive: %w", err)
	}
	defer func() {
		_ = os.Remove(tempArchive.Name())
	}()
	defer func() {
		_ = tempArchive.Close()
	}()

	if err := i.downloadArchive(ctx, arch, tempArchive); err != nil {
		return "", err
	}

	tempExtractDir, err := os.MkdirTemp("", "fgm-lint-extract-*")
	if err != nil {
		return "", fmt.Errorf("create temp extract dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempExtractDir)
	}()

	if err := archive.Extract(tempArchive.Name(), tempExtractDir); err != nil {
		return "", err
	}

	extractedBinary, err := findLintBinary(tempExtractDir)
	if err != nil {
		return "", err
	}

	if err := os.RemoveAll(installDir); err != nil {
		return "", fmt.Errorf("prepare install directory: %w", err)
	}
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create install directory: %w", err)
	}
	if err := os.Rename(extractedBinary, binaryPath); err != nil {
		return "", fmt.Errorf("move golangci-lint into place: %w", err)
	}

	return installDir, nil
}

func (i *Installer) downloadArchive(ctx context.Context, arch golangcilint.Archive, destination *os.File) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, arch.URL, nil)
	if err != nil {
		return fmt.Errorf("build archive request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("download golangci-lint archive: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download golangci-lint archive: unexpected status %s", resp.Status)
	}

	if i.progressWriter != nil {
		_, _ = fmt.Fprintf(i.progressWriter, "Downloading golangci-lint %s...\n", arch.Version)
	}

	hash := sha256.New()
	writer := io.MultiWriter(destination, hash)
	if i.progressWriter != nil {
		totalBytes, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		writer = io.MultiWriter(destination, hash, archive.NewProgressWriter(i.progressWriter, "download", totalBytes))
	}

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("write golangci-lint archive: %w", err)
	}

	if i.progressWriter != nil {
		_, _ = fmt.Fprintln(i.progressWriter, "Download complete.")
	}

	if arch.SHA256 != "" {
		gotChecksum := hex.EncodeToString(hash.Sum(nil))
		if !strings.EqualFold(gotChecksum, arch.SHA256) {
			return fmt.Errorf("verify golangci-lint archive checksum: got %s, want %s", gotChecksum, arch.SHA256)
		}
	}

	if _, err := destination.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind golangci-lint archive: %w", err)
	}

	return nil
}

func findLintBinary(root string) (string, error) {
	var match string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() == "golangci-lint" {
			match = path
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("scan extracted golangci-lint archive: %w", err)
	}
	if match == "" {
		return "", fmt.Errorf("find extracted golangci-lint binary: binary not found")
	}
	return match, nil
}
