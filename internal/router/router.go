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

	//Public routes
	mux.HandleFunc("/api/v1/health", health.Health)
	mux.HandleFunc("/api/v1/categories/", categories.List)
	mux.HandleFunc("/api/v1/categories", categories.List)

	//Posts: Get public, Post requires auth
	mux.HandleFunc("/api/v1/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			middleware.Auth(database)(http.HandlerFunc(posts.Collection)).ServeHTTP(w, r)
		} else {
			posts.Collection(w, r)
		}
	})
	mux.HandleFunc("/api/v1/posts/", posts.Item) //For individual posts. Public for now
	//Users: Register, Login are public, Me/Item require auth
	mux.HandleFunc("/api/v1/users/register", users.Register)
	mux.HandleFunc("/api/v1/users/login", users.Login)
	mux.Handle("/api/v1/users/me", middleware.Auth(database)(http.HandlerFunc(users.Me)))
	mux.Handle("/api/v1/users/", middleware.Auth(database)(http.HandlerFunc(users.Item)))

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
