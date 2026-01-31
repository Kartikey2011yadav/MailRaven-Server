package static

import (
	"io/fs"
	"net/http"
	"os"
	"strings"
)

// Handler returns a http.Handler that serves the SPA.
func Handler(fsys fs.FS) http.Handler {
	// We wrap the fs.FS to fallback to index.html on 404s for route paths
	// However, http.FileServer calls Open just to check if it exists?
	// The standard way to do SPA fallback in Go with http.FileServer is checking beforehand.
	
	fileServer := http.FileServer(http.FS(fsys))
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Clean path
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in the FS
		_, err := fs.Stat(fsys, path)
		if os.IsNotExist(err) {
			// If not found, serve index.html
			// Modify request to point to index.html
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
