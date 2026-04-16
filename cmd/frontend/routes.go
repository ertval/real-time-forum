// cmd/frontend/routes.go
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	/* ----------------------------
	   Static assets (/static/*)
	------------------------------*/
	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("./web/static")),
		),
	)

	// Favicon (served from root with cache headers)
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeFile(w, r, "./web/static/favicon.ico")
	})

	/*-----------------------------
	  Error assets (/errors/*)
	-----------------------------*/
	mux.Handle(
		"/errors/",
		http.StripPrefix(
			"/errors/",
			http.FileServer(http.Dir("./web/errors")),
		),
	)

	// ---- API and WebSocket proxy to backend (8080) ----
	backendURL, err := url.Parse(backendBaseURL)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	// Single proxy handler for both REST and WebSocket
	proxyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host = backendURL.Host
		proxy.ServeHTTP(w, r)
	})

	mux.Handle("/api/", proxyHandler)
	mux.Handle("/ws", proxyHandler)

	/*-----------------------------
	  SPA Catch-all
	-----------------------------*/
	// This handler serves files from /web/SPA/ if they exist, 
	// otherwise it serves /web/SPA/index.html.
	// This enables SPA client-side routing.
	spaFileServer := NewCustomFileServer(http.Dir("./web/SPA"), "./web/SPA/index.html")

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setNoStoreHeaders(w)
		spaFileServer.ServeHTTP(w, r)
	}))

	return mux
}

func setNoStoreHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
