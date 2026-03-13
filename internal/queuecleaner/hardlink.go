//go:build linux

package queuecleaner

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// hasHardlinks checks whether any file at or under path has more than one
// hard link (nlink > 1). This is used to protect files that are hardlinked
// into media libraries from being deleted by the cleaner.
//
// Returns false if the path is empty, does not exist, or cannot be read.
func hasHardlinks(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Lstat(path)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return fileHasHardlinks(info)
	}

	// Walk the directory and check each regular file.
	found := false
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found {
			return fs.SkipAll
		}
		if d.IsDir() {
			return nil
		}
		if checkFileHardlinks(p) {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}

// checkFileHardlinks stats a path and returns true if it has nlink > 1.
func checkFileHardlinks(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fileHasHardlinks(info)
}

// fileHasHardlinks returns true if the given FileInfo represents a file with nlink > 1.
func fileHasHardlinks(info fs.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	return stat.Nlink > 1
}
