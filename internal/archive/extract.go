package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FilenameSuffix returns the archive extension for a filename.
func FilenameSuffix(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".tar.gz"):
		return ".tar.gz"
	case strings.HasSuffix(filename, ".zip"):
		return ".zip"
	default:
		return ""
	}
}

// Extract extracts an archive to a destination directory.
func Extract(archivePath string, destination string) error {
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
			mode := os.FileMode(header.Mode) & 0o755
			output, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
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
		mode := file.Mode() & 0o755
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
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
