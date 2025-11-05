package router

import (
	"forum/internal/handlers"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	health := handlers.NewHealth()
	//v1 indicates this is version one of the api
	mux.HandleFunc("/api/v1/health", health.Health)

	return mux
}
