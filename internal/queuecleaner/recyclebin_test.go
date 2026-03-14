package queuecleaner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMoveToRecycleBin(t *testing.T) {
	// Create temp source dir with a file inside
	srcDir := t.TempDir()
	contentDir := filepath.Join(srcDir, "MyMovie.2024")
	if err := os.MkdirAll(contentDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, "movie.mkv"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create temp recycle bin destination
	recycleDir := t.TempDir()

	if err := moveToRecycleBin(contentDir, recycleDir); err != nil {
		t.Fatalf("moveToRecycleBin() error: %v", err)
	}

	// Source should no longer exist
	if _, err := os.Stat(contentDir); !os.IsNotExist(err) {
		t.Error("source directory should have been moved")
	}

	// Destination should exist under date-stamped folder
	dateFolder := time.Now().Format("2006-01-02")
	movedDir := filepath.Join(recycleDir, dateFolder, "MyMovie.2024")
	if _, err := os.Stat(movedDir); err != nil {
		t.Errorf("moved directory not found at %s: %v", movedDir, err)
	}

	// File inside should be intact
	data, err := os.ReadFile(filepath.Join(movedDir, "movie.mkv"))
	if err != nil {
		t.Fatalf("failed to read moved file: %v", err)
	}
	if string(data) != "data" {
		t.Errorf("file content = %q, want %q", string(data), "data")
	}
}

func TestMoveToRecycleBin_DuplicateName(t *testing.T) {
	srcDir := t.TempDir()
	recycleDir := t.TempDir()

	// Create two source dirs with the same base name (sequentially)
	for i := 0; i < 2; i++ {
		contentDir := filepath.Join(srcDir, "MyMovie")
		if err := os.MkdirAll(contentDir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := moveToRecycleBin(contentDir, recycleDir); err != nil {
			t.Fatalf("iteration %d: moveToRecycleBin() error: %v", i, err)
		}
	}

	// Both should exist (second one gets unique suffix)
	dateFolder := time.Now().Format("2006-01-02")
	entries, err := os.ReadDir(filepath.Join(recycleDir, dateFolder))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries in recycle bin, got %d", len(entries))
	}
}
