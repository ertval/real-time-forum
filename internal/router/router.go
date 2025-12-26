package router

import (
	"database/sql"
	"net/http"

	"forum/internal/handlers"
	"forum/internal/middleware"
)

const (
	apiPrefix      = "/api/v1"
	frontendOrigin = "http://localhost:3000"
)

func NewRouter(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// ---------------------------------------------------------
	// HANDLERS
	// ---------------------------------------------------------
	health := handlers.NewHealthHandler()
	posts := handlers.NewPostsHandler(database)
	users := handlers.NewUsersHandler(database)
	categories := handlers.NewCategoriesHandler(database)

	// ---------------------------------------------------------
	// MIDDLEWARE INSTANCES
	// ---------------------------------------------------------
	auth := middleware.Auth(database)

	// ---------------------------------------------------------
	// PUBLIC ROUTES (NO AUTH)
	// ---------------------------------------------------------

	// Healthcheck
	mux.HandleFunc(apiPrefix+"/health", health.Health)

	// Categories (read-only)
	mux.HandleFunc(apiPrefix+"/categories", categories.HandleCategories)
	mux.HandleFunc(apiPrefix+"/categories/", categories.HandleCategory)
	mux.HandleFunc(apiPrefix+"/categories/view", categories.ListCategoriesWithPosts)

	// Public posts
	mux.HandleFunc(apiPrefix+"/posts/public", posts.PublicList)

	// ---------------------------------------------------------
	// POSTS
	// ---------------------------------------------------------

	// /posts → GET (public), POST (auth required)
	mux.HandleFunc(apiPrefix+"/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			posts.HandlePosts(w, r)
		case http.MethodPost:
			auth(http.HandlerFunc(posts.HandlePosts)).ServeHTTP(w, r)
		default:
			handlers.MethodNotAllowed(w, r)
		}
	})

	// /posts/{id}, /posts/{id}/comments, likes, updates
	mux.HandleFunc(apiPrefix+"/posts/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			posts.HandlePost(w, r)
		case http.MethodPost, http.MethodPatch, http.MethodDelete:
			auth(http.HandlerFunc(posts.HandlePost)).ServeHTTP(w, r)
		default:
			handlers.MethodNotAllowed(w, r)
		}
	})

	// My posts (auth)
	mux.Handle(
		apiPrefix+"/posts/mine",
		auth(http.HandlerFunc(posts.ListMyPosts)),
	)

	// Liked posts (auth)
	mux.Handle(
		apiPrefix+"/posts/liked",
		auth(http.HandlerFunc(posts.ListLikedPosts)),
	)

	// ---------------------------------------------------------
	// USERS
	// ---------------------------------------------------------

	// Auth endpoints
	mux.HandleFunc(apiPrefix+"/users/register", users.Register)
	mux.HandleFunc(apiPrefix+"/users/login", users.Login)

	mux.Handle(
		apiPrefix+"/users/logout",
		auth(http.HandlerFunc(users.Logout)),
	)

	mux.Handle(
		apiPrefix+"/users/me",
		auth(http.HandlerFunc(users.Me)),
	)

	mux.Handle(
		apiPrefix+"/users/",
		auth(http.HandlerFunc(users.HandleUser)),
	)

	// ---------------------------------------------------------
	// COMMENTS
	// ---------------------------------------------------------

	mux.HandleFunc(apiPrefix+"/comments/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			posts.HandleComment(w, r)
		case http.MethodPost, http.MethodPatch, http.MethodDelete:
			auth(http.HandlerFunc(posts.HandleComment)).ServeHTTP(w, r)
		default:
			handlers.MethodNotAllowed(w, r)
		}
	})

	// ---------------------------------------------------------
	// API NOT FOUND (JSON ONLY)
	// ---------------------------------------------------------
	mux.HandleFunc("/api", notFoundJSON)
	mux.HandleFunc("/api/", notFoundJSON)

	// ---------------------------------------------------------
	// STATIC ERROR PAGES (FRONTEND)
	// ---------------------------------------------------------
	mux.Handle(
		"/errors/",
		http.StripPrefix(
			"/errors/",
			http.FileServer(http.Dir("./web/errors")),
		),
	)

	// ---------------------------------------------------------
	// GLOBAL MIDDLEWARE
	// ---------------------------------------------------------
	return addMiddlewares(mux)
}

func notFoundJSON(w http.ResponseWriter, r *http.Request) {
	handlers.WriteError(
		w,
		r,
		handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound),
	)
}

func addMiddlewares(handler http.Handler) http.Handler {
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)
	handler = middleware.EnableCORS(frontendOrigin)(handler)
	return handler
}
