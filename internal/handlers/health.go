package handlers

import (
	"net/http"
)

type Health struct{}

// NewHealth returns a simple healthcheck handler.
func NewHealth() *Health {
	return &Health{}
}

// Health responds with a basic API health status.
func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	WriteOK(w, map[string]any{"status": "ok"}, nil)
}
