package static

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// GetFS returns the filesystem for the static assets, rooted at "dist".
func GetFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
