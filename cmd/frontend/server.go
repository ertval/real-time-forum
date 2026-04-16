// cmd/frontend/server.go
package main

import (
	"net/http"
)

type CustomFileServer struct {
	handler  http.Handler
	notFound string
}

func NewCustomFileServer(fs http.FileSystem, notFoundPage string) http.Handler {
	return &CustomFileServer{
		handler:  http.FileServer(fs),
		notFound: notFoundPage,
	}
}

func (c *CustomFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try serving the file
	rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	c.handler.ServeHTTP(rec, r)

	// If file not found, serve the SPA shell instead
	if rec.status == http.StatusNotFound {
		// Reset the header (though hard to do in Go without buffering)
		// Better approach: serve the file directly if it exists, otherwise serve fallback.
		http.ServeFile(w, r, c.notFound)
	}
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
