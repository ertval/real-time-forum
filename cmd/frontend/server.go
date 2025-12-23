package main

import "net/http"

type CustomFileServer struct {
	root http.FileSystem
}

// serves u
func (c CustomFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := c.root.Open(r.URL.Path)
	if err != nil {
		http.ServeFile(w, r, "./web/errors/404.html") // Or absolute path
		return
	}
	defer f.Close()
	http.FileServer(c.root).ServeHTTP(w, r)
}
