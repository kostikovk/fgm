package goinstall

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

	arch, err := i.provider.FindArchive(ctx, version)
	if err != nil {
		return "", err
	}

	tempArchive, err := os.CreateTemp("", "fgm-go-archive-*"+archive.FilenameSuffix(arch.Filename))
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

	tempExtractDir, err := os.MkdirTemp("", "fgm-go-extract-*")
	if err != nil {
		return "", fmt.Errorf("create temp extract dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempExtractDir)
	}()

	if err := archive.Extract(tempArchive.Name(), tempExtractDir); err != nil {
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

func (i *Installer) downloadArchive(ctx context.Context, arch goreleases.Archive, destination *os.File) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, arch.URL, nil)
	if err != nil {
		return fmt.Errorf("build archive request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("download Go archive: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download Go archive: unexpected status %s", resp.Status)
	}

	if i.progressWriter != nil {
		_, _ = fmt.Fprintf(i.progressWriter, "Downloading Go %s...\n", arch.Version)
	}

	hash := sha256.New()
	writer := io.MultiWriter(destination, hash)
	if i.progressWriter != nil {
		totalBytes, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		writer = io.MultiWriter(destination, hash, archive.NewProgressWriter(i.progressWriter, "download", totalBytes))
	}

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("write Go archive: %w", err)
	}

	if i.progressWriter != nil {
		_, _ = fmt.Fprintln(i.progressWriter, "Download complete.")
	}

	if arch.SHA256 != "" {
		gotChecksum := hex.EncodeToString(hash.Sum(nil))
		if !strings.EqualFold(gotChecksum, arch.SHA256) {
			return fmt.Errorf("verify Go archive checksum: got %s, want %s", gotChecksum, arch.SHA256)
		}
	}

	if _, err := destination.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind Go archive: %w", err)
	}

	return nil
}
