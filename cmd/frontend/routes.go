// cmd/frontend/routes.go
package main

import (
	"forum/internal/handlers"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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
	backendURL, err := url.Parse(backendBaseURL)
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

		setNoStoreHeaders(w)
		http.ServeFile(w, r, "./web/templates/home.html")
	})
	return mux
}

func servePostByIDTemplate(path string) http.HandlerFunc {
	templateHandler := serveTemplate(path)
	return func(w http.ResponseWriter, r *http.Request) {
		postID, ok := parseRouteID(r.URL.Path, "view-post", "post")
		if !ok {
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

		status, err := postStatusChecker.GetPostStatus(
			r.Context(),
			postID,
			r.Header.Get("Cookie"),
		)
		if err != nil {
			log.Printf("failed to verify post existence (post_id=%d): %v", postID, err)
			handlers.WriteError(
				w,
				r,
				handlers.NewError(
					"INTERNAL_SERVER_ERROR",
					"failed to load post",
					http.StatusInternalServerError,
				),
			)
			return
		}

		switch status {
		case http.StatusOK:
			templateHandler(w, r)
		case http.StatusNotFound:
			handlers.WriteError(
				w,
				r,
				handlers.NewError("NOT_FOUND", "post not found", http.StatusNotFound),
			)
		case http.StatusBadRequest:
			handlers.WriteError(
				w,
				r,
				handlers.NewError("BAD_REQUEST", "invalid post id", http.StatusBadRequest),
			)
		default:
			log.Printf("unexpected backend status while verifying post %d: %d", postID, status)
			handlers.WriteError(
				w,
				r,
				handlers.NewError(
					"INTERNAL_SERVER_ERROR",
					"failed to load post",
					http.StatusInternalServerError,
				),
			)
		}
	}
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

		setNoStoreHeaders(w)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, path)
	}
}

func setNoStoreHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
