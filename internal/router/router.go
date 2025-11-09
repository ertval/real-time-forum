package router

import (
	"forum/internal/handlers"
	"forum/internal/middleware"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	health := handlers.NewHealth()
	posts := handlers.NewPosts()
	users := handlers.NewUsers()
	//v1 indicates this is version one of the api

	//Health
	mux.HandleFunc("/api/v1/health", health.Health)
	//Posts
	mux.HandleFunc("/api/v1/posts", posts.Collection)
	mux.HandleFunc("/api/v1/posts/", posts.Item)
	//Users
	mux.HandleFunc("/api/v1/users/me", users.Me)
	mux.HandleFunc("/api/v1/users/", users.Item)

	//Catches all undefined routes and servers json 404 message
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteError(w, handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound))
	})
	return addMiddlewares(mux)
}

func addMiddlewares(h http.Handler) http.Handler {
	//here (h) is just the param for the returned function that runs instantly w
	h = middleware.CORS("http://localhost:5173")(h) // dev frontend origin
	h = middleware.Recoverer(h)
	h = middleware.Logger(h)
	return h
}
