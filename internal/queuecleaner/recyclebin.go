package queuecleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// moveToRecycleBin moves the contents at sourcePath into a date-stamped
// subdirectory under recyclePath. Uses os.Rename for efficiency when source
// and destination are on the same filesystem; returns an error otherwise.
func moveToRecycleBin(sourcePath, recyclePath string) error {
	datePath := filepath.Join(recyclePath, time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(datePath, 0o750); err != nil {
		return fmt.Errorf("create recycle dir: %w", err)
	}

	dest := filepath.Join(datePath, filepath.Base(sourcePath))

	// Ensure uniqueness if a same-name folder already exists in today's bin.
	if _, err := os.Stat(dest); err == nil {
		dest = fmt.Sprintf("%s_%d", dest, time.Now().UnixNano())
	}

	if err := os.Rename(sourcePath, dest); err != nil {
		return fmt.Errorf("move to recycle bin: %w", err)
	}
	return nil
}
