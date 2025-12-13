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
	// MIDDLEWARE INSTANCES (micro-optimization)
	// ---------------------------------------------------------
	auth := middleware.Auth(database)

	// ---------------------------------------------------------
	// PUBLIC ROUTES
	// ---------------------------------------------------------
	mux.HandleFunc(apiPrefix+"/health", health.Health)

	// Categories
	mux.HandleFunc(apiPrefix+"/categories", categories.Collection)
	mux.HandleFunc(apiPrefix+"/categories/", categories.Item)

	// Public posts listing
	mux.HandleFunc(apiPrefix+"/posts/public", posts.PublicList)

	// ---------------------------------------------------------
	// POSTS
	// ---------------------------------------------------------

	// /posts → GET public, POST auth
	mux.HandleFunc(apiPrefix+"/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auth(http.HandlerFunc(posts.Collection)).ServeHTTP(w, r)
			return
		}
		posts.Collection(w, r)
	})

	// /posts/{id}, /posts/{id}/comments, /posts/{id}/like
	mux.HandleFunc(apiPrefix+"/posts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auth(http.HandlerFunc(posts.Item)).ServeHTTP(w, r)
			return
		}
		posts.Item(w, r)
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
		auth(http.HandlerFunc(users.Item)),
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
	handler = middleware.CORS(frontendOrigin)(handler)
	return handler
}
