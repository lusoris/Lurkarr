package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var buildFS embed.FS

// BuildFS returns the frontend build filesystem rooted at the build/ directory.
// Returns nil if the build only contains the .gitkeep placeholder.
func BuildFS() fs.FS {
	if _, err := fs.Stat(buildFS, "build/index.html"); err != nil {
		return nil
	}
	sub, err := fs.Sub(buildFS, "build")
	if err != nil {
		return nil
	}
	return sub
}
