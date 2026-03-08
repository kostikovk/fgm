package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilenameSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "tar.gz suffix",
			filename: "go1.21.0.linux-amd64.tar.gz",
			want:     ".tar.gz",
		},
		{
			name:     "zip suffix",
			filename: "go1.21.0.windows-amd64.zip",
			want:     ".zip",
		},
		{
			name:     "unknown suffix",
			filename: "go1.21.0.linux-amd64.tar.bz2",
			want:     "",
		},
		{
			name:     "empty string",
			filename: "",
			want:     "",
		},
		{
			name:     "no suffix",
			filename: "somefile",
			want:     "",
		},
		{
			name:     "only extension",
			filename: ".tar.gz",
			want:     ".tar.gz",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := FilenameSuffix(tc.filename)
			if got != tc.want {
				t.Errorf("FilenameSuffix(%q) = %q, want %q", tc.filename, got, tc.want)
			}
		})
	}
}

func TestExtractTarGz(t *testing.T) {
	t.Parallel()

	// Build a tar.gz archive in memory containing:
	//   mydir/
	//   mydir/hello.txt  (content: "hello tar.gz")
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "test.tar.gz")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive file: %v", err)
		}
		defer f.Close()

		gw := gzip.NewWriter(f)
		defer gw.Close()

		tw := tar.NewWriter(gw)
		defer tw.Close()

		// Directory entry.
		dirHdr := &tar.Header{
			Typeflag: tar.TypeDir,
			Name:     "mydir/",
			Mode:     0o755,
		}
		if err := tw.WriteHeader(dirHdr); err != nil {
			t.Fatalf("write tar dir header: %v", err)
		}

		// File entry.
		content := []byte("hello tar.gz")
		fileHdr := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "mydir/hello.txt",
			Mode:     0o644,
			Size:     int64(len(content)),
		}
		if err := tw.WriteHeader(fileHdr); err != nil {
			t.Fatalf("write tar file header: %v", err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatalf("write tar file content: %v", err)
		}
	}()

	destDir := t.TempDir()
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	// Verify the directory was created.
	dirPath := filepath.Join(destDir, "mydir")
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("stat extracted dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected %q to be a directory", dirPath)
	}

	// Verify the file was extracted with correct content.
	filePath := filepath.Join(destDir, "mydir", "hello.txt")
	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(got) != "hello tar.gz" {
		t.Errorf("file content = %q, want %q", string(got), "hello tar.gz")
	}
}

func TestExtractZip(t *testing.T) {
	t.Parallel()

	// Build a zip archive containing:
	//   subdir/
	//   subdir/world.txt  (content: "hello zip")
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "test.zip")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create zip file: %v", err)
		}
		defer f.Close()

		zw := zip.NewWriter(f)
		defer zw.Close()

		// Directory entry.
		dirHdr := &zip.FileHeader{
			Name:   "subdir/",
			Method: zip.Deflate,
		}
		dirHdr.SetMode(0o755 | os.ModeDir)
		if _, err := zw.CreateHeader(dirHdr); err != nil {
			t.Fatalf("create zip dir header: %v", err)
		}

		// File entry.
		fileHdr := &zip.FileHeader{
			Name:   "subdir/world.txt",
			Method: zip.Deflate,
		}
		fileHdr.SetMode(0o644)
		fw, err := zw.CreateHeader(fileHdr)
		if err != nil {
			t.Fatalf("create zip file header: %v", err)
		}
		if _, err := fw.Write([]byte("hello zip")); err != nil {
			t.Fatalf("write zip file content: %v", err)
		}
	}()

	destDir := t.TempDir()
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	// Verify the directory was created.
	dirPath := filepath.Join(destDir, "subdir")
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("stat extracted dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected %q to be a directory", dirPath)
	}

	// Verify the file was extracted with correct content.
	filePath := filepath.Join(destDir, "subdir", "world.txt")
	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(got) != "hello zip" {
		t.Errorf("file content = %q, want %q", string(got), "hello zip")
	}
}

func TestExtract_UnsupportedFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		archiveName string
	}{
		{
			name:        "tar.bz2 file",
			archiveName: "archive.tar.bz2",
		},
		{
			name:        "plain text file",
			archiveName: "file.txt",
		},
		{
			name:        "no extension",
			archiveName: "archive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			archiveDir := t.TempDir()
			archivePath := filepath.Join(archiveDir, tc.archiveName)

			// Write a dummy file so the path exists.
			if err := os.WriteFile(archivePath, []byte("not an archive"), 0o644); err != nil {
				t.Fatalf("write dummy file: %v", err)
			}

			destDir := t.TempDir()
			err := Extract(archivePath, destDir)
			if err == nil {
				t.Errorf("Extract(%q) expected error for unsupported format, got nil", archivePath)
			}
		})
	}
}

func TestSafeJoin(t *testing.T) {
	t.Parallel()

	t.Run("normal path accepted", func(t *testing.T) {
		t.Parallel()
		base := "/tmp/dest"
		got, err := safeJoin(base, "subdir/file.txt")
		if err != nil {
			t.Fatalf("safeJoin() unexpected error: %v", err)
		}
		want := filepath.Join(base, "subdir", "file.txt")
		if got != want {
			t.Errorf("safeJoin() = %q, want %q", got, want)
		}
	})

	t.Run("path traversal rejected", func(t *testing.T) {
		t.Parallel()
		_, err := safeJoin("/tmp/dest", "../../../etc/passwd")
		if err == nil {
			t.Error("safeJoin() expected error for path traversal, got nil")
		}
	})

	t.Run("path traversal via tar.gz rejected", func(t *testing.T) {
		t.Parallel()

		// Build a tar.gz archive that contains a path-traversal entry.
		archiveDir := t.TempDir()
		archivePath := filepath.Join(archiveDir, "traversal.tar.gz")

		func() {
			f, err := os.Create(archivePath)
			if err != nil {
				t.Fatalf("create archive: %v", err)
			}
			defer f.Close()

			gw := gzip.NewWriter(f)
			defer gw.Close()

			tw := tar.NewWriter(gw)
			defer tw.Close()

			content := []byte("evil content")
			hdr := &tar.Header{
				Typeflag: tar.TypeReg,
				Name:     "../../../etc/passwd",
				Mode:     0o644,
				Size:     int64(len(content)),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				t.Fatalf("write traversal header: %v", err)
			}
			if _, err := tw.Write(content); err != nil {
				t.Fatalf("write traversal content: %v", err)
			}
		}()

		destDir := t.TempDir()
		err := Extract(archivePath, destDir)
		if err == nil {
			t.Error("Extract() with path traversal entry expected error, got nil")
		}
	})

	t.Run("path traversal via zip rejected", func(t *testing.T) {
		t.Parallel()

		// Build a zip archive that contains a path-traversal entry.
		archiveDir := t.TempDir()
		archivePath := filepath.Join(archiveDir, "traversal.zip")

		func() {
			f, err := os.Create(archivePath)
			if err != nil {
				t.Fatalf("create archive: %v", err)
			}
			defer f.Close()

			zw := zip.NewWriter(f)
			defer zw.Close()

			hdr := &zip.FileHeader{
				Name:   "../../../etc/passwd",
				Method: zip.Deflate,
			}
			hdr.SetMode(0o644)
			fw, err := zw.CreateHeader(hdr)
			if err != nil {
				t.Fatalf("create traversal zip header: %v", err)
			}
			if _, err := fw.Write([]byte("evil content")); err != nil {
				t.Fatalf("write traversal zip content: %v", err)
			}
		}()

		destDir := t.TempDir()
		err := Extract(archivePath, destDir)
		if err == nil {
			t.Error("Extract() with path traversal entry expected error, got nil")
		}
	})

	t.Run("base directory itself accepted", func(t *testing.T) {
		t.Parallel()
		base := "/tmp/dest"
		got, err := safeJoin(base, ".")
		if err != nil {
			t.Fatalf("safeJoin() unexpected error: %v", err)
		}
		want := filepath.Clean(base)
		if got != want {
			t.Errorf("safeJoin() = %q, want %q", got, want)
		}
	})
}

func TestExtractTarGz_RejectsEscapingEntry(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "escape.tar.gz")
	func() {
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer file.Close()

		gzipWriter := gzip.NewWriter(file)
		defer gzipWriter.Close()

		tarWriter := tar.NewWriter(gzipWriter)
		defer tarWriter.Close()

		content := []byte("boom")
		if err := tarWriter.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "../escape.txt",
			Mode:     0o644,
			Size:     int64(len(content)),
		}); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			t.Fatalf("write tar content: %v", err)
		}
	}()

	err := Extract(archivePath, t.TempDir())
	if err == nil {
		t.Fatal("expected escaping tar entry to fail")
	}
	if !strings.Contains(err.Error(), "escapes destination") {
		t.Fatalf("error = %q, want escape error", err)
	}
}

func TestExtractZip_RejectsEscapingEntry(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "escape.zip")
	func() {
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer file.Close()

		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		writer, err := zipWriter.Create("../escape.txt")
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := writer.Write([]byte("boom")); err != nil {
			t.Fatalf("write zip content: %v", err)
		}
	}()

	err := Extract(archivePath, t.TempDir())
	if err == nil {
		t.Fatal("expected escaping zip entry to fail")
	}
	if !strings.Contains(err.Error(), "escapes destination") {
		t.Fatalf("error = %q, want escape error", err)
	}
}

func TestExtractTarGz_ErrorForInvalidGzip(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "invalid.tar.gz")
	if err := os.WriteFile(archivePath, []byte("not gzip"), 0o644); err != nil {
		t.Fatalf("write invalid archive: %v", err)
	}

	err := Extract(archivePath, t.TempDir())
	if err == nil {
		t.Fatal("expected invalid gzip archive to fail")
	}
	if !strings.Contains(err.Error(), "open gzip reader") {
		t.Fatalf("error = %q, want gzip reader error", err)
	}
}

func TestExtractZip_ErrorForInvalidArchive(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "invalid.zip")
	if err := os.WriteFile(archivePath, []byte("not zip"), 0o644); err != nil {
		t.Fatalf("write invalid archive: %v", err)
	}

	err := Extract(archivePath, t.TempDir())
	if err == nil {
		t.Fatal("expected invalid zip archive to fail")
	}
	if !strings.Contains(err.Error(), "open zip archive") {
		t.Fatalf("error = %q, want zip archive error", err)
	}
}

func TestExtractTarGz_ErrorCreatingDirectory(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "dir.tar.gz")
	func() {
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer file.Close()

		gzipWriter := gzip.NewWriter(file)
		defer gzipWriter.Close()

		tarWriter := tar.NewWriter(gzipWriter)
		defer tarWriter.Close()

		if err := tarWriter.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     "nested",
			Mode:     0o755,
		}); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
	}()

	destination := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(destination, []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	err := Extract(archivePath, destination)
	if err == nil {
		t.Fatal("expected directory creation error")
	}
	if !strings.Contains(err.Error(), "create tar directory") {
		t.Fatalf("error = %q, want tar directory error", err)
	}
}

func TestExtractTarGz_ErrorCreatingFile(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "file.tar.gz")
	func() {
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer file.Close()

		gzipWriter := gzip.NewWriter(file)
		defer gzipWriter.Close()

		tarWriter := tar.NewWriter(gzipWriter)
		defer tarWriter.Close()

		content := []byte("hello")
		if err := tarWriter.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "nested",
			Mode:     0o644,
			Size:     int64(len(content)),
		}); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			t.Fatalf("write tar content: %v", err)
		}
	}()

	destination := t.TempDir()
	if err := os.Mkdir(filepath.Join(destination, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir existing directory: %v", err)
	}

	err := Extract(archivePath, destination)
	if err == nil {
		t.Fatal("expected file creation error")
	}
	if !strings.Contains(err.Error(), "create tar file") {
		t.Fatalf("error = %q, want tar file error", err)
	}
}

func TestExtractZip_ErrorCreatingDirectory(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "dir.zip")
	func() {
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer file.Close()

		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		header := &zip.FileHeader{Name: "nested/"}
		header.SetMode(0o755 | os.ModeDir)
		if _, err := zipWriter.CreateHeader(header); err != nil {
			t.Fatalf("create zip dir header: %v", err)
		}
	}()

	destination := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(destination, []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	err := Extract(archivePath, destination)
	if err == nil {
		t.Fatal("expected zip directory creation error")
	}
	if !strings.Contains(err.Error(), "create zip directory") {
		t.Fatalf("error = %q, want zip directory error", err)
	}
}

func TestExtractZip_ErrorCreatingFile(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "file.zip")
	func() {
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer file.Close()

		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		header := &zip.FileHeader{Name: "nested", Method: zip.Deflate}
		header.SetMode(0o644)
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			t.Fatalf("create zip file header: %v", err)
		}
		if _, err := writer.Write([]byte("hello")); err != nil {
			t.Fatalf("write zip content: %v", err)
		}
	}()

	destination := t.TempDir()
	if err := os.Mkdir(filepath.Join(destination, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir existing directory: %v", err)
	}

	err := Extract(archivePath, destination)
	if err == nil {
		t.Fatal("expected zip file creation error")
	}
	if !strings.Contains(err.Error(), "create zip file") {
		t.Fatalf("error = %q, want zip file error", err)
	}
}

func TestExtractTarGz_ErrorCreatingParentDirectory(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "nested.tar.gz")
	func() {
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer file.Close()

		gzipWriter := gzip.NewWriter(file)
		defer gzipWriter.Close()

		tarWriter := tar.NewWriter(gzipWriter)
		defer tarWriter.Close()

		content := []byte("hello")
		if err := tarWriter.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "nested/file.txt",
			Mode:     0o644,
			Size:     int64(len(content)),
		}); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			t.Fatalf("write tar content: %v", err)
		}
	}()

	destination := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(destination, []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	err := Extract(archivePath, destination)
	if err == nil {
		t.Fatal("expected tar parent directory creation error")
	}
	if !strings.Contains(err.Error(), "create tar parent directory") {
		t.Fatalf("error = %q, want tar parent directory error", err)
	}
}

func TestExtractZip_ErrorCreatingParentDirectory(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "nested.zip")
	func() {
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer file.Close()

		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		header := &zip.FileHeader{Name: "nested/file.txt", Method: zip.Deflate}
		header.SetMode(0o644)
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			t.Fatalf("create zip header: %v", err)
		}
		if _, err := writer.Write([]byte("hello")); err != nil {
			t.Fatalf("write zip content: %v", err)
		}
	}()

	destination := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(destination, []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	err := Extract(archivePath, destination)
	if err == nil {
		t.Fatal("expected zip parent directory creation error")
	}
	if !strings.Contains(err.Error(), "create zip parent directory") {
		t.Fatalf("error = %q, want zip parent directory error", err)
	}
}

func TestExtractTarGz_ErrorReadingEntryPayload(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "truncated.tar.gz")
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	content := []byte("hello world")
	if err := tarWriter.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "file.txt",
		Mode:     0o644,
		Size:     int64(len(content) + 5),
	}); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("write tar content: %v", err)
	}
	_ = tarWriter.Close()
	_ = gzipWriter.Close()

	if err := os.WriteFile(archivePath, buffer.Bytes(), 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	err := Extract(archivePath, t.TempDir())
	if err == nil {
		t.Fatal("expected tar payload read error")
	}
	if !strings.Contains(err.Error(), "write tar file") {
		t.Fatalf("error = %q, want tar payload error", err)
	}
}

func TestExtractZip_ErrorOpeningFilePayload(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "unsupported.zip")
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	header := &zip.FileHeader{Name: "file.txt", Method: zip.Store}
	header.SetMode(0o644)
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		t.Fatalf("create zip header: %v", err)
	}
	if _, err := writer.Write([]byte("hello")); err != nil {
		t.Fatalf("write zip payload: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}

	data := buffer.Bytes()
	for i := 0; i+10 < len(data); i++ {
		if bytes.Equal(data[i:i+4], []byte("PK\x03\x04")) {
			data[i+8] = 99
			data[i+9] = 0
			break
		}
	}
	for i := 0; i+12 < len(data); i++ {
		if bytes.Equal(data[i:i+4], []byte("PK\x01\x02")) {
			data[i+10] = 99
			data[i+11] = 0
			break
		}
	}
	if err := os.WriteFile(archivePath, data, 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	err = Extract(archivePath, t.TempDir())
	if err == nil {
		t.Fatal("expected zip file open error")
	}
	if !strings.Contains(err.Error(), "open zip file") {
		t.Fatalf("error = %q, want zip open error", err)
	}
}

func TestExtractZip_ErrorReadingFilePayload(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "truncated-payload.zip")
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	header := &zip.FileHeader{Name: "file.txt", Method: zip.Store}
	header.SetMode(0o644)
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		t.Fatalf("create zip header: %v", err)
	}
	if _, err := writer.Write([]byte("hello")); err != nil {
		t.Fatalf("write zip payload: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}

	data := append([]byte(nil), buffer.Bytes()...)
	setLE32 := func(offset int, value uint32) {
		data[offset] = byte(value)
		data[offset+1] = byte(value >> 8)
		data[offset+2] = byte(value >> 16)
		data[offset+3] = byte(value >> 24)
	}
	for i := 0; i+30 < len(data); i++ {
		if bytes.Equal(data[i:i+4], []byte("PK\x03\x04")) {
			setLE32(i+18, 10)
			setLE32(i+22, 10)
			break
		}
	}
	for i := 0; i+46 < len(data); i++ {
		if bytes.Equal(data[i:i+4], []byte("PK\x01\x02")) {
			setLE32(i+20, 10)
			setLE32(i+24, 10)
			break
		}
	}
	if err := os.WriteFile(archivePath, data, 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	err = Extract(archivePath, t.TempDir())
	if err == nil {
		t.Fatal("expected zip payload read error")
	}
	if !strings.Contains(err.Error(), "write zip file") {
		t.Fatalf("error = %q, want zip payload error", err)
	}
}

func TestProgressWriter(t *testing.T) {
	t.Parallel()

	t.Run("reports at 10 percent intervals", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		pw := NewProgressWriter(&buf, "test", 100)

		// Write 50 bytes in chunks of 10.
		for range 5 {
			chunk := make([]byte, 10)
			n, err := pw.Write(chunk)
			if err != nil {
				t.Fatalf("Write: %v", err)
			}
			if n != 10 {
				t.Fatalf("Write returned %d, want 10", n)
			}
		}

		output := buf.String()
		if !strings.Contains(output, "10%") {
			t.Errorf("expected 10%% in output, got %q", output)
		}
		if !strings.Contains(output, "50%") {
			t.Errorf("expected 50%% in output, got %q", output)
		}
	})

	t.Run("no output when total is zero", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		pw := NewProgressWriter(&buf, "test", 0)

		chunk := make([]byte, 100)
		if _, err := pw.Write(chunk); err != nil {
			t.Fatalf("Write: %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("expected no output for zero total, got %q", buf.String())
		}
	})
}

func TestExtractZip_PreservesExecutablePermissions(t *testing.T) {
	t.Parallel()

	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "exec.zip")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create zip file: %v", err)
		}
		defer f.Close()

		zw := zip.NewWriter(f)
		defer zw.Close()

		fileHdr := &zip.FileHeader{
			Name:   "run.sh",
			Method: zip.Deflate,
		}
		fileHdr.SetMode(0o755)
		fw, err := zw.CreateHeader(fileHdr)
		if err != nil {
			t.Fatalf("create zip file header: %v", err)
		}
		if _, err := fw.Write([]byte("#!/bin/sh\necho hello")); err != nil {
			t.Fatalf("write zip file content: %v", err)
		}
	}()

	destDir := t.TempDir()
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	filePath := filepath.Join(destDir, "run.sh")
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat extracted file: %v", err)
	}

	mode := info.Mode()
	if mode&0o111 == 0 {
		t.Errorf("expected executable permissions on %q, got %v", filePath, mode)
	}
}

func TestExtractTarGz_ErrorForCorruptGzip(t *testing.T) {
	t.Parallel()

	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "corrupt.tar.gz")

	if err := os.WriteFile(archivePath, []byte("this is not gzip data"), 0o644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	destDir := t.TempDir()
	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected an error for corrupt gzip data")
	}
	if !strings.Contains(err.Error(), "gzip") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "gzip")
	}
}

func TestExtractTarGz_ErrorForCorruptTar(t *testing.T) {
	t.Parallel()

	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "badtar.tar.gz")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create file: %v", err)
		}
		defer f.Close()

		gw := gzip.NewWriter(f)
		// Write non-tar data inside a valid gzip stream.
		if _, err := gw.Write([]byte("this is not valid tar data")); err != nil {
			t.Fatalf("write gzip content: %v", err)
		}
		if err := gw.Close(); err != nil {
			t.Fatalf("close gzip writer: %v", err)
		}
	}()

	destDir := t.TempDir()
	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected an error for corrupt tar data inside gzip")
	}
	if !strings.Contains(err.Error(), "tar") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "tar")
	}
}

func TestExtractTarGz_ErrorForMissingFile(t *testing.T) {
	t.Parallel()

	destDir := t.TempDir()
	err := Extract(filepath.Join(destDir, "nonexistent.tar.gz"), destDir)
	if err == nil {
		t.Fatal("expected error for missing tar.gz file, got nil")
	}
	if !strings.Contains(err.Error(), "open tar.gz archive") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "open tar.gz archive")
	}
}

func TestExtractZip_ErrorForMissingFile(t *testing.T) {
	t.Parallel()

	destDir := t.TempDir()
	err := Extract(filepath.Join(destDir, "nonexistent.zip"), destDir)
	if err == nil {
		t.Fatal("expected error for missing zip file, got nil")
	}
	if !strings.Contains(err.Error(), "open zip archive") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "open zip archive")
	}
}

func TestExtractTarGz_HandlesSymlinks(t *testing.T) {
	t.Parallel()

	// Build a tar.gz archive containing:
	//   mydir/
	//   mydir/hello.txt  (regular file)
	//   mydir/link.txt   (symlink -> hello.txt)
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "symlink.tar.gz")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive file: %v", err)
		}
		defer f.Close()

		gw := gzip.NewWriter(f)
		defer gw.Close()

		tw := tar.NewWriter(gw)
		defer tw.Close()

		// Directory entry.
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     "mydir/",
			Mode:     0o755,
		}); err != nil {
			t.Fatalf("write dir header: %v", err)
		}

		// Regular file entry.
		content := []byte("symlink target")
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "mydir/hello.txt",
			Mode:     0o644,
			Size:     int64(len(content)),
		}); err != nil {
			t.Fatalf("write file header: %v", err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatalf("write file content: %v", err)
		}

		// Symlink entry pointing to hello.txt.
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeSymlink,
			Name:     "mydir/link.txt",
			Linkname: "hello.txt",
		}); err != nil {
			t.Fatalf("write symlink header: %v", err)
		}
	}()

	destDir := t.TempDir()
	// The current implementation does not handle TypeSymlink in its switch.
	// The symlink entry will fall through the switch and be silently skipped.
	// This should not produce an error.
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	// Verify the regular file was extracted.
	filePath := filepath.Join(destDir, "mydir", "hello.txt")
	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(got) != "symlink target" {
		t.Errorf("file content = %q, want %q", string(got), "symlink target")
	}

	// The symlink entry should have been silently skipped (no symlink created).
	linkPath := filepath.Join(destDir, "mydir", "link.txt")
	_, err = os.Lstat(linkPath)
	if err == nil {
		t.Error("expected symlink to not be created (skipped by implementation), but it exists")
	}
}

func TestExtractTarGz_SkipsUnknownTypes(t *testing.T) {
	t.Parallel()

	// Build a tar.gz archive containing a regular file followed by an entry
	// with an unknown type flag byte. The unknown entry should be silently
	// skipped without error.
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "unknown.tar.gz")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive file: %v", err)
		}
		defer f.Close()

		gw := gzip.NewWriter(f)
		defer gw.Close()

		tw := tar.NewWriter(gw)
		defer tw.Close()

		// Regular file entry.
		content := []byte("good file")
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "valid.txt",
			Mode:     0o644,
			Size:     int64(len(content)),
		}); err != nil {
			t.Fatalf("write file header: %v", err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatalf("write file content: %v", err)
		}

		// Entry with a non-standard type flag (TypeChar = '3' is unusual
		// and not handled by the switch in extractTarGz).
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeChar,
			Name:     "chardev",
			Mode:     0o644,
			Size:     0,
		}); err != nil {
			t.Fatalf("write unknown type header: %v", err)
		}

		// Another unknown type: TypeFifo.
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeFifo,
			Name:     "fifo",
			Mode:     0o644,
			Size:     0,
		}); err != nil {
			t.Fatalf("write fifo header: %v", err)
		}
	}()

	destDir := t.TempDir()
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	// Verify the valid file was extracted.
	got, err := os.ReadFile(filepath.Join(destDir, "valid.txt"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(got) != "good file" {
		t.Errorf("file content = %q, want %q", string(got), "good file")
	}

	// Verify unknown type entries were skipped (not created on disk).
	for _, name := range []string{"chardev", "fifo"} {
		path := filepath.Join(destDir, name)
		if _, err := os.Lstat(path); err == nil {
			t.Errorf("expected %q to not be created (unknown type), but it exists", path)
		}
	}
}

func TestExtractTarGz_SymlinkPathTraversalRejected(t *testing.T) {
	t.Parallel()

	// Build a tar.gz with a symlink whose Name escapes the destination.
	// safeJoin should catch this before the symlink is even considered.
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "symtraversal.tar.gz")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive file: %v", err)
		}
		defer f.Close()

		gw := gzip.NewWriter(f)
		defer gw.Close()

		tw := tar.NewWriter(gw)
		defer tw.Close()

		// Symlink with a path-traversal Name.
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeSymlink,
			Name:     "../../../etc/evil-link",
			Linkname: "/etc/passwd",
		}); err != nil {
			t.Fatalf("write symlink header: %v", err)
		}
	}()

	destDir := t.TempDir()
	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error for symlink path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "escapes destination") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "escapes destination")
	}
}

func TestExtractTarGz_DirPathTraversalRejected(t *testing.T) {
	t.Parallel()

	// Build a tar.gz with a TypeDir entry whose name escapes the destination.
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "dirtraversal.tar.gz")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive file: %v", err)
		}
		defer f.Close()

		gw := gzip.NewWriter(f)
		defer gw.Close()

		tw := tar.NewWriter(gw)
		defer tw.Close()

		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     "../../../etc/evil-dir/",
			Mode:     0o755,
		}); err != nil {
			t.Fatalf("write dir header: %v", err)
		}
	}()

	destDir := t.TempDir()
	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error for directory path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "escapes destination") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "escapes destination")
	}
}

func TestExtractZip_DirPathTraversalRejected(t *testing.T) {
	t.Parallel()

	// Build a zip with a directory entry whose name escapes the destination.
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "dirtraversal.zip")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive file: %v", err)
		}
		defer f.Close()

		zw := zip.NewWriter(f)
		defer zw.Close()

		hdr := &zip.FileHeader{
			Name:   "../../../etc/evil-dir/",
			Method: zip.Deflate,
		}
		hdr.SetMode(0o755 | os.ModeDir)
		if _, err := zw.CreateHeader(hdr); err != nil {
			t.Fatalf("create zip dir header: %v", err)
		}
	}()

	destDir := t.TempDir()
	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error for zip directory path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "escapes destination") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "escapes destination")
	}
}

func TestExtractZip_FilePathTraversalRejected(t *testing.T) {
	t.Parallel()

	// Build a zip with a file entry whose name escapes the destination.
	// This exercises the safeJoin error path for zip file entries (line 111).
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "filetraversal.zip")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive file: %v", err)
		}
		defer f.Close()

		zw := zip.NewWriter(f)
		defer zw.Close()

		hdr := &zip.FileHeader{
			Name:   "../../../etc/shadow",
			Method: zip.Deflate,
		}
		hdr.SetMode(0o644)
		fw, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatalf("create zip file header: %v", err)
		}
		if _, err := fw.Write([]byte("evil")); err != nil {
			t.Fatalf("write zip file content: %v", err)
		}
	}()

	destDir := t.TempDir()
	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error for zip file path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "escapes destination") {
		t.Fatalf("error = %q, want it to contain %q", err.Error(), "escapes destination")
	}
}

func TestExtractZip_MultipleFilesExtracted(t *testing.T) {
	t.Parallel()

	// Build a zip with multiple files to exercise the full zip extraction
	// loop including file open, copy, and close paths.
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "multi.zip")

	files := map[string]string{
		"a.txt":        "content a",
		"dir/b.txt":    "content b",
		"dir/c.sh":     "#!/bin/sh\necho c",
		"dir/sub/d.go": "package main",
	}

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create zip file: %v", err)
		}
		defer f.Close()

		zw := zip.NewWriter(f)
		defer zw.Close()

		// Create directory entries first.
		for _, dir := range []string{"dir/", "dir/sub/"} {
			hdr := &zip.FileHeader{Name: dir, Method: zip.Deflate}
			hdr.SetMode(0o755 | os.ModeDir)
			if _, err := zw.CreateHeader(hdr); err != nil {
				t.Fatalf("create dir header %q: %v", dir, err)
			}
		}

		for name, content := range files {
			hdr := &zip.FileHeader{Name: name, Method: zip.Deflate}
			if strings.HasSuffix(name, ".sh") {
				hdr.SetMode(0o755)
			} else {
				hdr.SetMode(0o644)
			}
			fw, err := zw.CreateHeader(hdr)
			if err != nil {
				t.Fatalf("create file header %q: %v", name, err)
			}
			if _, err := fw.Write([]byte(content)); err != nil {
				t.Fatalf("write file %q: %v", name, err)
			}
		}
	}()

	destDir := t.TempDir()
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	for name, wantContent := range files {
		path := filepath.Join(destDir, name)
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %q: %v", name, err)
		}
		if string(got) != wantContent {
			t.Errorf("file %q content = %q, want %q", name, string(got), wantContent)
		}
	}

	// Check executable permission on .sh file.
	shInfo, err := os.Stat(filepath.Join(destDir, "dir/c.sh"))
	if err != nil {
		t.Fatalf("stat c.sh: %v", err)
	}
	if shInfo.Mode()&0o111 == 0 {
		t.Errorf("expected executable permissions on c.sh, got %v", shInfo.Mode())
	}
}

func TestExtractTarGz_MultipleTypesInOneArchive(t *testing.T) {
	t.Parallel()

	// Build a tar.gz that exercises the directory creation, file creation,
	// symlink (skipped), and unknown type (skipped) paths all in one pass.
	archiveDir := t.TempDir()
	archivePath := filepath.Join(archiveDir, "mixed.tar.gz")

	func() {
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("create archive: %v", err)
		}
		defer f.Close()

		gw := gzip.NewWriter(f)
		defer gw.Close()

		tw := tar.NewWriter(gw)
		defer tw.Close()

		// Directory.
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     "mixed/",
			Mode:     0o755,
		}); err != nil {
			t.Fatalf("write dir header: %v", err)
		}

		// Regular file.
		content := []byte("hello mixed")
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "mixed/file.txt",
			Mode:     0o644,
			Size:     int64(len(content)),
		}); err != nil {
			t.Fatalf("write file header: %v", err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatalf("write file content: %v", err)
		}

		// Symlink (will be skipped).
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeSymlink,
			Name:     "mixed/link",
			Linkname: "file.txt",
		}); err != nil {
			t.Fatalf("write symlink header: %v", err)
		}

		// Block device (unknown type, will be skipped).
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeBlock,
			Name:     "mixed/blockdev",
			Mode:     0o644,
			Size:     0,
		}); err != nil {
			t.Fatalf("write block header: %v", err)
		}

		// Another regular file after the skipped entries, to confirm the
		// loop continues correctly.
		content2 := []byte("after skipped entries")
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "mixed/after.txt",
			Mode:     0o644,
			Size:     int64(len(content2)),
		}); err != nil {
			t.Fatalf("write second file header: %v", err)
		}
		if _, err := tw.Write(content2); err != nil {
			t.Fatalf("write second file content: %v", err)
		}
	}()

	destDir := t.TempDir()
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	// Verify regular files were extracted.
	for _, tc := range []struct {
		name    string
		content string
	}{
		{"mixed/file.txt", "hello mixed"},
		{"mixed/after.txt", "after skipped entries"},
	} {
		got, err := os.ReadFile(filepath.Join(destDir, tc.name))
		if err != nil {
			t.Fatalf("read %q: %v", tc.name, err)
		}
		if string(got) != tc.content {
			t.Errorf("file %q = %q, want %q", tc.name, string(got), tc.content)
		}
	}

	// Verify skipped entries were not created.
	for _, name := range []string{"mixed/link", "mixed/blockdev"} {
		if _, err := os.Lstat(filepath.Join(destDir, name)); err == nil {
			t.Errorf("expected %q to not be created, but it exists", name)
		}
	}
}
