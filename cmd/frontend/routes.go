// cmd/frontend/routes.go
package main

import (
	"forum/internal/handlers"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
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

	// Favicon (served from root with cache headers to prevent flicker)
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
	mux.HandleFunc("/edit-post", serveTemplate("./web/templates/edit-post.html"))
	mux.HandleFunc("/edit-post/", serveTemplate("./web/templates/edit-post.html"))
	mux.HandleFunc("/forgot-password", serveTemplate("./web/templates/forgot-password.html"))
	mux.HandleFunc("/login", serveTemplate("./web/templates/login.html"))
	mux.HandleFunc("/register", serveTemplate("./web/templates/register.html"))
	mux.HandleFunc("/activity", serveTemplate("./web/templates/activity.html"))
	servePostByID := servePostByIDTemplate("./web/templates/view-post.html")
	mux.HandleFunc("/view-post", servePostByID)
	mux.HandleFunc("/view-post/", servePostByID)
	mux.HandleFunc("/post", servePostByID)
	mux.HandleFunc("/post/", servePostByID)

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
	return mux
}

func servePostByIDTemplate(path string) http.HandlerFunc {
	templateHandler := serveTemplate(path)
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := parsePostIDFromFrontendPath(r.URL.Path); !ok {
			handlers.WriteError(
				w,
				r,
				handlers.NewError(
					"BAD_REQUEST",
					"invalid post id",
					http.StatusBadRequest,
				),
			)
			return
		}

		templateHandler(w, r)
	}
}

func parsePostIDFromFrontendPath(path string) (int64, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		return 0, false
	}

	route := parts[0]
	if route != "view-post" && route != "post" {
		return 0, false
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
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
