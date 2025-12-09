package router

import (
	"database/sql"
	"forum/internal/handlers"
	"forum/internal/middleware"
	"net/http"
)

func NewRouter(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	health := handlers.NewHealth()
	posts := handlers.NewPosts(database)
	users := handlers.NewUsers(database)
	categories := handlers.NewCategories(database)

	// ---------------------------
	// PUBLIC ROUTES (no auth)
	// ---------------------------
	mux.HandleFunc("/api/v1/health", health.Health)

	// Categories (GET list, POST create, GET single, DELETE single)
	mux.HandleFunc("/api/v1/categories", categories.Collection)
	mux.HandleFunc("/api/v1/categories/", categories.Item)

	// Posts:
	// GET /posts → public
	// POST /posts → requires auth
	mux.HandleFunc("/api/v1/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// wrap only POST with Auth middleware
			middleware.Auth(database)(http.HandlerFunc(posts.Collection)).ServeHTTP(w, r)
			return
		}
		posts.Collection(w, r)
	})

	// Single post + subroutes (comments, like) — currently all public
	mux.HandleFunc("/api/v1/posts/", posts.Item)

	// Users:
	mux.HandleFunc("/api/v1/users/register", users.Register)
	mux.HandleFunc("/api/v1/users/login", users.Login)

	// Auth required routes
	mux.Handle("/api/v1/users/logout", middleware.Auth(database)(http.HandlerFunc(users.Logout)))
	mux.Handle("/api/v1/users/me", middleware.Auth(database)(http.HandlerFunc(users.Me)))
	mux.Handle("/api/v1/users/", middleware.Auth(database)(http.HandlerFunc(users.Item)))

	// ---------------------------
	// API NOT FOUND (JSON 404)
	// ---------------------------
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteError(w, handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound))
	})

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteError(w, handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound))
	})

	// Wrap everything with middleware (CORS, Recoverer, Logger)
	return addMiddlewares(mux)
}

func addMiddlewares(h http.Handler) http.Handler {
	h = middleware.CORS("http://localhost:3000")(h) // dev frontend origin
	h = middleware.Recoverer(h)
	h = middleware.Logger(h)
	return h
}
