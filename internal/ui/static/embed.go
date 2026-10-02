package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed css/* js/* img/*
var staticFS embed.FS

// FS returns an http.FileSystem backed by the embedded static assets.
func FS() http.FileSystem {
	return http.FS(staticFS)
}

// SubFS returns an fs.FS sub-rooted at the static files directory.
func SubFS() fs.FS {
	return staticFS
}
