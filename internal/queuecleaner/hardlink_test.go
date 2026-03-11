package queuecleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasHardlinksEmptyPath(t *testing.T) {
	if hasHardlinks("") {
		t.Error("expected false for empty path")
	}
}

func TestHasHardlinksNonExistent(t *testing.T) {
	if hasHardlinks("/tmp/does-not-exist-lurkarr-test-12345") {
		t.Error("expected false for non-existent path")
	}
}

func TestHasHardlinksSingleFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "single.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A newly created file has nlink=1 — no hardlinks.
	if hasHardlinks(f) {
		t.Error("expected false for file with nlink=1")
	}
}

func TestHasHardlinksLinkedFile(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "original.txt")
	if err := os.WriteFile(original, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	linked := filepath.Join(dir, "linked.txt")
	if err := os.Link(original, linked); err != nil {
		t.Fatal(err)
	}

	// File with a hardlink has nlink=2.
	if !hasHardlinks(original) {
		t.Error("expected true for file with nlink=2")
	}
	if !hasHardlinks(linked) {
		t.Error("expected true for the hardlinked copy too")
	}
}

func TestHasHardlinksDirectoryNoLinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	if hasHardlinks(dir) {
		t.Error("expected false for directory with only nlink=1 files")
	}
}

func TestHasHardlinksDirectoryWithLink(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "file1.txt")
	if err := os.WriteFile(f1, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("other"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a hardlink to file1 outside this directory.
	externalDir := t.TempDir()
	if err := os.Link(f1, filepath.Join(externalDir, "linked.txt")); err != nil {
		t.Fatal(err)
	}

	// The directory should show hardlinks because file1 now has nlink=2.
	if !hasHardlinks(dir) {
		t.Error("expected true for directory containing a hardlinked file")
	}
}

func TestHasHardlinksNestedDirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(sub, "nested.txt")
	if err := os.WriteFile(f, []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	externalDir := t.TempDir()
	if err := os.Link(f, filepath.Join(externalDir, "linked.txt")); err != nil {
		t.Fatal(err)
	}

	// Should find the hardlink in a nested subdirectory.
	if !hasHardlinks(dir) {
		t.Error("expected true for directory with hardlinked file in subdirectory")
	}
}
