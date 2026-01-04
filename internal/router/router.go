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
	// MIDDLEWARE
	// ---------------------------------------------------------
	auth := middleware.Auth(database)

	// ---------------------------------------------------------
	// HEALTH
	// ---------------------------------------------------------
	mux.HandleFunc(apiPrefix+"/health", health.Health)

	// ---------------------------------------------------------
	// CATEGORIES (PUBLIC)
	// ---------------------------------------------------------
	mux.HandleFunc(apiPrefix+"/categories", categories.HandleCategories)
	mux.HandleFunc(apiPrefix+"/categories/", categories.HandleCategory)
	mux.HandleFunc(apiPrefix+"/categories/view", categories.ListCategoriesWithPosts)

	// ---------------------------------------------------------
	// POSTS COLLECTION
	// ---------------------------------------------------------
	// GET  /posts        → list
	// POST /posts        → create (auth)
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

	// ---------------------------------------------------------
	// POSTS ITEM + COMMENTS + REACTIONS
	// ---------------------------------------------------------
	// Handles:
	// GET    /posts/{id}
	// PATCH  /posts/{id}
	// DELETE /posts/{id}
	// POST   /posts/{id}/like
	// POST   /posts/{id}/dislike
	// GET    /posts/{id}/comments
	// POST   /posts/{id}/comments
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

	// ---------------------------------------------------------
	// USER POSTS
	// ---------------------------------------------------------
	mux.Handle(
		apiPrefix+"/posts/mine",
		auth(http.HandlerFunc(posts.ListMyPosts)),
	)

	mux.Handle(
		apiPrefix+"/posts/liked",
		auth(http.HandlerFunc(posts.ListLikedPosts)),
	)

	// ---------------------------------------------------------
	// USERS
	// ---------------------------------------------------------
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
	// API FALLBACK (JSON 404)
	// ---------------------------------------------------------
	mux.HandleFunc("/api", notFoundJSON)
	mux.HandleFunc("/api/", notFoundJSON)

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
