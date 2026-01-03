package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	mux := http.NewServeMux()

	// ---------------------------------------------------------
	// Static assets (/static/*)
	// ---------------------------------------------------------
	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("./web/static")),
		),
	)

	// ---- API proxy to backend (8080) ----
	backendURL, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	apiProxy := httputil.NewSingleHostReverseProxy(backendURL)

	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// preserve host headers appropriately (optional; usually fine either way)
		r.Host = backendURL.Host
		apiProxy.ServeHTTP(w, r)
	}))

	// ---------------------------------------------------------
	// Pages (HTML templates)
	// ---------------------------------------------------------

	mux.HandleFunc("/", serveTemplate("./web/templates/home.html"))

	mux.HandleFunc("/home", serveTemplate("./web/templates/home.html"))
	mux.HandleFunc("/create-post", serveTemplate("./web/templates/create-post.html"))
	mux.HandleFunc("/forgot-password", serveTemplate("./web/templates/forgot-password.html"))
	mux.HandleFunc("/login", serveTemplate("./web/templates/login.html"))
	mux.HandleFunc("/register", serveTemplate("./web/templates/register.html"))
	mux.Handle("/view-post/", serveTemplate("./web/templates/view-post.html"))

	log.Println("Frontend running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

func serveTemplate(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
