package handlers

import (
	"net/http"
	"strings"
)

type Users struct{}

func NewUsers() *Users { return &Users{} }

func (u *Users) Item(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	if id == "" || strings.Contains(id, "/") {
		WriteError(w, NewError("NOT_FOUND", "route not found", http.StatusNotFound))
		return
	}
	if r.Method != http.MethodGet {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}
	WriteOK(w, map[string]any{"id": id, "username": "stub"}, nil)
}

func (u *Users) Me(w http.ResponseWriter, r *http.Request) {
	WriteError(w, NewError("UNAUTHORIZED", "login required", http.StatusUnauthorized))
}
