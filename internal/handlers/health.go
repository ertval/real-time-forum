package handlers

import (
	"net/http"
)

type Health struct{}

func NewHealth() *Health {
	return &Health{}
}

func (hnd *Health) Health(w http.ResponseWriter, r *http.Request) {
	WriteOK(w, map[string]any{"status": "ok"}, nil)
}
