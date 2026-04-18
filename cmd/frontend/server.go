// cmd/frontend/server.go
package main

import (
	"net/http"
)

type CustomFileServer struct {
	handler  http.Handler
	notFound string
	fs       http.FileSystem
}

func NewCustomFileServer(fs http.FileSystem, notFoundPage string) http.Handler {
	return &CustomFileServer{
		handler:  http.FileServer(fs),
		notFound: notFoundPage,
		fs:       fs,
	}
}

func (c *CustomFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// For SPA, we want to serve index.html for any path that doesn't
	// correspond to an actual file (like /login, /activity, etc.)

	// Check if the path exists in the configured SPA filesystem.
	f, err := c.fs.Open(r.URL.Path)
	if err == nil {
		defer f.Close()
		stat, err := f.Stat()
		if err == nil && !stat.IsDir() {
			// File exists and is not a directory, serve it normally
			c.handler.ServeHTTP(w, r)
			return
		}
	}

	// Otherwise, serve the SPA shell
	// Ensure we set the correct content type for the shell
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, c.notFound)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	// Don't call underlying WriteHeader yet!
	// This is slightly dangerous if someone else calls Write()
	// because then Go will implicitly WriteHeader(200).
	if code != http.StatusNotFound {
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == http.StatusNotFound {
		return len(b), nil // Drop the "404 page not found" body
	}
	return r.ResponseWriter.Write(b)
}
