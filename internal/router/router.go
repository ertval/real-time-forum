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
	//v1 indicates this is version one of the api

	//Health
	mux.HandleFunc("/api/v1/health", health.Health)
	//Posts
	mux.HandleFunc("/api/v1/posts", posts.Collection)
	mux.HandleFunc("/api/v1/posts/", posts.Item)
	//Users
	mux.HandleFunc("/api/v1/users/register", users.Register)
	mux.HandleFunc("/api/v1/users/me", users.Me)
	mux.HandleFunc("/api/v1/users/", users.Item)
	//Categories
	mux.HandleFunc("/api/v1/categories/", categories.List)
	mux.HandleFunc("/api/v1/categories", categories.List)

	// JSON 404 is served for undefined API routes
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteError(w, handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound))
	})

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteError(w, handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound))
	})

	// Example placeholder handler that uses the DB
	//mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	if err := db.InspectAllTables(database, w); err != nil {
	//		db.HandleRuntimeError(err, "inspecting all tables")
	//		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//		return
	//	}
	//})
	return addMiddlewares(mux)
}

func addMiddlewares(h http.Handler) http.Handler {
	//here (h) is just the param for the returned function that runs instantly w
	h = middleware.CORS("http://localhost:3000")(h) // dev frontend origin
	h = middleware.Recoverer(h)
	h = middleware.Logger(h)
	return h
}
