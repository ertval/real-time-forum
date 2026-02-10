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

	// If file not found, serve custom 404 page
	if rec.status == http.StatusNotFound {
		http.ServeFile(w, r, c.notFound)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
