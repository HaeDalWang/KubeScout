package ui

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// GetHandler returns the file system for the embedded frontend.
// It effectively "strips" the "dist" prefix so the server sees the root.
func GetFileSystem() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
