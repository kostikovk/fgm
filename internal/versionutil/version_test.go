package versionutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		left string
		right string
		want int
	}{
		{
			name:  "equal versions",
			left:  "1.21.0",
			right: "1.21.0",
			want:  0,
		},
		{
			name:  "left less than right minor",
			left:  "1.20.0",
			right: "1.21.0",
			want:  -1,
		},
		{
			name:  "left greater than right minor",
			left:  "1.22.0",
			right: "1.21.0",
			want:  1,
		},
		{
			name:  "left less than right major",
			left:  "1.0.0",
			right: "2.0.0",
			want:  -1,
		},
		{
			name:  "left greater than right major",
			left:  "2.0.0",
			right: "1.0.0",
			want:  1,
		},
		{
			name:  "left less than right patch",
			left:  "1.21.0",
			right: "1.21.1",
			want:  -1,
		},
		{
			name:  "left greater than right patch",
			left:  "1.21.1",
			right: "1.21.0",
			want:  1,
		},
		{
			name:  "different part counts 1.21 vs 1.21.0",
			left:  "1.21",
			right: "1.21.0",
			want:  -1,
		},
		{
			name:  "different part counts 1.21.0 vs 1.21",
			left:  "1.21.0",
			right: "1.21",
			want:  1,
		},
		{
			name:  "different part counts equal prefix",
			left:  "1.21",
			right: "1.21",
			want:  0,
		},
		{
			name:  "non-numeric part stops parsing left",
			left:  "1.21rc1",
			right: "1.22.0",
			want:  -1,
		},
		{
			name:  "non-numeric part in right stops parsing",
			left:  "1.22.0",
			right: "1.21rc1",
			want:  1,
		},
		{
			name:  "non-numeric part in dotted segment stops left",
			left:  "1.21.rc1",
			right: "1.21.0",
			want:  -1,
		},
		{
			name:  "both have non-numeric parts at same position",
			left:  "1.21.rc1",
			right: "1.21.rc2",
			want:  0,
		},
		{
			name:  "single segment equal",
			left:  "5",
			right: "5",
			want:  0,
		},
		{
			name:  "single segment left less",
			left:  "4",
			right: "5",
			want:  -1,
		},
		{
			name:  "single segment left greater",
			left:  "6",
			right: "5",
			want:  1,
		},
		{
			name:  "empty left vs non-empty right",
			left:  "",
			right: "1.0.0",
			want:  -1,
		},
		{
			name:  "both empty",
			left:  "",
			right: "",
			want:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := CompareVersions(tc.left, tc.right)
			// Normalize to -1, 0, 1 for comparison since the function only
			// guarantees sign, but the implementation returns exactly -1/0/1.
			if sign(got) != sign(tc.want) {
				t.Errorf("CompareVersions(%q, %q) = %d, want sign %d", tc.left, tc.right, got, tc.want)
			}
		})
	}
}

func sign(n int) int {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}

func TestFindNearestFile(t *testing.T) {
	t.Parallel()

	t.Run("file found in current dir", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "target.txt")
		if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
			t.Fatalf("failed to create target file: %v", err)
		}

		got, err := FindNearestFile(dir, "target.txt")
		if err != nil {
			t.Fatalf("FindNearestFile returned unexpected error: %v", err)
		}
		if got != target {
			t.Errorf("FindNearestFile = %q, want %q", got, target)
		}
	})

	t.Run("file found in parent dir", func(t *testing.T) {
		t.Parallel()
		parent := t.TempDir()
		child := filepath.Join(parent, "child")
		if err := os.Mkdir(child, 0o755); err != nil {
			t.Fatalf("failed to create child dir: %v", err)
		}
		target := filepath.Join(parent, "target.txt")
		if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
			t.Fatalf("failed to create target file: %v", err)
		}

		got, err := FindNearestFile(child, "target.txt")
		if err != nil {
			t.Fatalf("FindNearestFile returned unexpected error: %v", err)
		}
		if got != target {
			t.Errorf("FindNearestFile = %q, want %q", got, target)
		}
	})

	t.Run("file found in grandparent dir", func(t *testing.T) {
		t.Parallel()
		grandparent := t.TempDir()
		child := filepath.Join(grandparent, "child")
		grandchild := filepath.Join(child, "grandchild")
		if err := os.MkdirAll(grandchild, 0o755); err != nil {
			t.Fatalf("failed to create grandchild dir: %v", err)
		}
		target := filepath.Join(grandparent, "target.txt")
		if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
			t.Fatalf("failed to create target file: %v", err)
		}

		got, err := FindNearestFile(grandchild, "target.txt")
		if err != nil {
			t.Fatalf("FindNearestFile returned unexpected error: %v", err)
		}
		if got != target {
			t.Errorf("FindNearestFile = %q, want %q", got, target)
		}
	})

	t.Run("file not found returns ErrNotFound", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		_, err := FindNearestFile(dir, "nonexistent.txt")
		if err == nil {
			t.Fatal("FindNearestFile expected error, got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("FindNearestFile error = %v, want errors.Is(err, ErrNotFound) to be true", err)
		}
	})

	t.Run("stat error other than not-exist is returned", func(t *testing.T) {
		t.Parallel()

		if runtime.GOOS == "windows" {
			t.Skip("permission-based stat error test is unix-only")
		}

		// Create a directory tree: parent/child.
		// Put a directory named "target.txt" inside parent, then chmod parent
		// to remove read permission. When stat is called on parent/target.txt
		// from child, the stat will return a permission error (not IsNotExist).
		parent := t.TempDir()
		restricted := filepath.Join(parent, "restricted")
		child := filepath.Join(restricted, "child")
		if err := os.MkdirAll(child, 0o755); err != nil {
			t.Fatalf("mkdir child: %v", err)
		}

		// Create target.txt inside restricted so we can make it un-statable.
		targetFile := filepath.Join(restricted, "target.txt")
		if err := os.WriteFile(targetFile, []byte("hello"), 0o644); err != nil {
			t.Fatalf("write target: %v", err)
		}

		// Remove all permissions from restricted so stat will fail with EACCES.
		if err := os.Chmod(restricted, 0o000); err != nil {
			t.Fatalf("chmod restricted: %v", err)
		}
		t.Cleanup(func() {
			// Restore permissions so t.TempDir cleanup succeeds.
			os.Chmod(restricted, 0o755)
		})

		_, err := FindNearestFile(child, "target.txt")
		if err == nil {
			t.Fatal("expected stat error, got nil")
		}
		// The error should NOT be ErrNotFound; it should be a stat error.
		if errors.Is(err, ErrNotFound) {
			t.Fatal("expected stat error, got ErrNotFound")
		}
		if !strings.Contains(err.Error(), "stat") {
			t.Fatalf("err = %q, want it to contain 'stat'", err)
		}
	})

	t.Run("directory with same name should not match", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// Create a directory with the target name instead of a file.
		targetDir := filepath.Join(dir, "target.txt")
		if err := os.Mkdir(targetDir, 0o755); err != nil {
			t.Fatalf("failed to create directory named target.txt: %v", err)
		}

		_, err := FindNearestFile(dir, "target.txt")
		if err == nil {
			t.Fatal("FindNearestFile expected error when only a directory matches, got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("FindNearestFile error = %v, want errors.Is(err, ErrNotFound) to be true", err)
		}
	})
}
