package router

import (
	"database/sql"
	"forum/internal/handlers"
	"forum/internal/middleware"
	"net/http"
)

const apiPrefix = "/api/v1"

func NewRouter(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// Handlers
	health := handlers.NewHealthHandler()
	posts := handlers.NewPostsHandler(database)
	users := handlers.NewUsersHandler(database)
	categories := handlers.NewCategoriesHandler(database)

	// ---------------------------
	// PUBLIC ROUTES
	// ---------------------------
	mux.HandleFunc(apiPrefix+"/health", health.Health)

	// Categories (public GETs, private POST/DELETE depending on design)
	mux.HandleFunc(apiPrefix+"/categories", categories.Collection)
	mux.HandleFunc(apiPrefix+"/categories/", categories.Item)

	// Posts:
	// GET = public
	// POST = auth required
	mux.HandleFunc(apiPrefix+"/posts/public", posts.PublicList)

	mux.HandleFunc(apiPrefix+"/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Only protect POST with Auth
			middleware.Auth(database)(http.HandlerFunc(posts.Collection)).ServeHTTP(w, r)
			return
		}
		posts.Collection(w, r)
	})

	// Single post + nested resources (comments, like)
	mux.HandleFunc(apiPrefix+"/posts/", posts.Item)

	// ---------------------------
	// USERS
	// ---------------------------
	mux.HandleFunc(apiPrefix+"/users/register", users.Register)
	mux.HandleFunc(apiPrefix+"/users/login", users.Login)

	// Auth-protected user routes
	mux.Handle(apiPrefix+"/users/logout", middleware.Auth(database)(http.HandlerFunc(users.Logout)))
	mux.Handle(apiPrefix+"/users/me", middleware.Auth(database)(http.HandlerFunc(users.Me)))
	mux.Handle(apiPrefix+"/users/", middleware.Auth(database)(http.HandlerFunc(users.Item)))

	// ---------------------------
	// NOT FOUND HANDLERS (JSON)
	// ---------------------------
	mux.HandleFunc("/api", notFoundJSON)
	mux.HandleFunc("/api/", notFoundJSON)

	// ---------------------------
	// GLOBAL MIDDLEWARE
	// ---------------------------
	return addMiddlewares(mux)
}

func notFoundJSON(w http.ResponseWriter, r *http.Request) {
	handlers.WriteError(w, handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound))
}

const frontendOrigin = "http://localhost:3000"

func addMiddlewares(handler http.Handler) http.Handler {
	// Order matters: Logger → Recoverer → CORS
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)
	handler = middleware.CORS(frontendOrigin)(handler)

	return handler
}
