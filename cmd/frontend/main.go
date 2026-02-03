// cmd/frontend/main.go
package main

import (
	"forum/internal/handlers"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
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

	/* --------------------------
	   Pages (HTML templates)
	---------------------------*/

	mux.HandleFunc("/create-post", serveTemplate("./web/templates/create-post.html"))
	mux.HandleFunc("/forgot-password", serveTemplate("./web/templates/forgot-password.html"))
	mux.HandleFunc("/login", serveTemplate("./web/templates/login.html"))
	mux.HandleFunc("/register", serveTemplate("./web/templates/register.html"))
	mux.HandleFunc("/my-posts/", serveTemplate("./web/templates/my-posts.html"))
	mux.Handle("/view-post/", serveTemplate("./web/templates/view-post.html"))
	mux.Handle("/my-liked-posts", serveTemplate("./web/templates/my-liked-posts.html"))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HIT / handler: %s", r.URL.Path)

		if r.URL.Path != "/" {
			handlers.WriteError(w, r, handlers.NewError(
				"NOT_FOUND",
				"route not found",
				http.StatusNotFound,
			))
			return
		}
		http.ServeFile(w, r, "./web/templates/home.html")
	})

	log.Println("Frontend running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

func serveTemplate(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(path); err != nil {
			handlers.WriteError(
				w,
				r,
				handlers.NewError(
					"TEMPLATE_NOT_FOUND",
					"Internal server error",
					http.StatusInternalServerError,
				),
			)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, path)
	}
}
