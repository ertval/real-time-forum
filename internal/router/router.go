package router

import (
	"database/sql"
	"net/http"

	"forum/internal/handlers"
	"forum/internal/middleware"
)

const apiPrefix = "/api/v1"
const frontendOrigin = "http://localhost:3000"

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
	mux.HandleFunc(apiPrefix+"/health", health.Health)

	// Categories (read-only)
	mux.HandleFunc(apiPrefix+"/categories", categories.HandleCategories)
	mux.HandleFunc(apiPrefix+"/categories/", categories.HandleCategory)

	// Public posts listing
	mux.HandleFunc(apiPrefix+"/posts/public", posts.PublicList)

	// ---------------------------------------------------------
	// POSTS
	// ---------------------------------------------------------

	// /posts → GET public, POST requires auth
	mux.HandleFunc(apiPrefix+"/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			posts.HandlePosts(w, r)
		case http.MethodPost:
			auth(http.HandlerFunc(posts.HandlePosts)).ServeHTTP(w, r)
		default:
			handlers.MethodNotAllowed(w)
		}
	})

	// /posts/{id}, /posts/{id}/comments, /posts/{id}/like /posts/{id}/dislike
	mux.HandleFunc(apiPrefix+"/posts/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// View post OR list comments → public
			posts.HandlePost(w, r)

		case http.MethodPost, http.MethodPatch, http.MethodDelete:
			// create comment, like, update/delete post → auth required
			auth(http.HandlerFunc(posts.HandlePost)).ServeHTTP(w, r)

		default:
			handlers.MethodNotAllowed(w)
		}
	})

	// ---------------------------------------------------------
	// USERS
	// ---------------------------------------------------------
	mux.HandleFunc(apiPrefix+"/users/register", users.Register)
	mux.HandleFunc(apiPrefix+"/users/login", users.Login)

	mux.Handle(apiPrefix+"/users/logout",
		auth(http.HandlerFunc(users.Logout)),
	)
	mux.Handle(apiPrefix+"/users/me",
		auth(http.HandlerFunc(users.Me)),
	)
	mux.Handle(apiPrefix+"/users/",
		auth(http.HandlerFunc(users.HandleUser)),
	)

	// ---------------------------------------------------------
	// NOT FOUND (API ONLY)
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
		handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound),
	)
}

func addMiddlewares(handler http.Handler) http.Handler {
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)
	handler = middleware.EnableCORS(frontendOrigin)(handler)
	return handler
}
