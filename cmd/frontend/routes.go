// cmd/frontend/routes.go
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	cspHeaderPolicy = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; img-src 'self' data: blob: /static/uploads/; connect-src 'self' ws: wss:; frame-ancestors 'none'; object-src 'none'; base-uri 'self'; form-action 'self';"
)

var backendBaseURL = "http://localhost:8080"

// SecurityHeaders injects standard browser security headers into all frontend responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", cspHeaderPolicy)
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// NewMux builds the frontend server mux, including static SPA delivery and the
// backend proxy routes used for REST and authenticated WebSocket traffic.
func NewMux() http.Handler {
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

	// Route both /api/ and /ws through the same backend proxy so the frontend
	// server stays the browser-facing entry point while preserving session-cookie
	// authentication for both HTTP requests and WebSocket upgrades.
	proxyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host = backendURL.Host
		proxy.ServeHTTP(w, r)
	})

	mux.Handle("/api/", proxyHandler)
	mux.Handle("/ws", proxyHandler)

	/*-----------------------------
	  SPA Catch-all
	-----------------------------*/
	// This handler serves files from /SPA/ if they exist, 
	// otherwise it serves /SPA/index.html.
	// This enables SPA client-side routing.
	spaFileServer := NewCustomFileServer(http.Dir("./SPA"), "./SPA/index.html")

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setNoStoreHeaders(w)
		spaFileServer.ServeHTTP(w, r)
	}))

	return SecurityHeaders(mux)
}

// setNoStoreHeaders prevents browsers from caching the SPA shell so auth-gated
// routes always revalidate through the frontend server.
func setNoStoreHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
