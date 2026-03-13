package queuecleaner

import (
	"os"
	"path/filepath"
	"testing"
)

// These tests run on all platforms to verify the hasHardlinks function's
// basic invariants (empty path, non-existent path, regular file).

func TestHasHardlinks_EmptyPath(t *testing.T) {
	if hasHardlinks("") {
		t.Fatal("expected false for empty path")
	}
}

func TestHasHardlinks_NonExistent(t *testing.T) {
	if hasHardlinks("/this/path/does/not/exist") {
		t.Fatal("expected false for non-existent path")
	}
}

func TestHasHardlinks_RegularFile(t *testing.T) {
	// hasHardlinks should return false for a regular file (not a directory)
	f, err := os.CreateTemp(t.TempDir(), "testfile")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if hasHardlinks(f.Name()) {
		t.Fatal("expected false for regular file")
	}
}

func TestHasHardlinks_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	if hasHardlinks(dir) {
		t.Fatal("expected false for empty directory")
	}
}

func TestHasHardlinks_DirWithSingleFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	// On non-Linux: always false (stub). On Linux: single file has nlink=1, so false.
	if hasHardlinks(dir) {
		t.Fatal("expected false for directory with single unlinked file")
	}
}
