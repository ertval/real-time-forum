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
	mux.HandleFunc("/api/v1/health", health.Health)
	mux.HandleFunc("/api/v1/posts", posts.List)
	mux.HandleFunc("/api/v1/users/me", users.Me)

	//Catches all undefined routes and servers json 404 message
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteError(w, handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound))
	})
	return withMiddlewares(mux)
}

func withMiddlewares(h http.Handler) http.Handler {
	h = middleware.CORS("http://localhost:5173")(h) // dev frontend origin
	h = middleware.Recoverer(h)
	h = middleware.Logger(h)
	return h
}
