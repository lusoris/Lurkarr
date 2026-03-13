//go:build !linux

package queuecleaner

import (
	"io/fs"
	"os"
	"path/filepath"
)

// hasHardlinks checks whether any file at or under path has more than one
// hard link. On non-Linux platforms this always returns false because the
// syscall.Stat_t nlink check is Linux-specific.
func hasHardlinks(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Lstat(path)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	found := false
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return fs.SkipAll
		}
		return nil
	})
	return found
}
