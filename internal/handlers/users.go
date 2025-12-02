package handlers

import (
	"database/sql"
	"encoding/json"
	db "forum/internal/db"
	"net/http"
	"strconv"
	"strings"
)

type Users struct{ db *sql.DB }

func NewUsers(database *sql.DB) *Users { return &Users{db: database} }

func (u *Users) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}
	var req db.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}
	id, err := db.CreateUser(r.Context(), u.db, req)
	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", err.Error(), http.StatusBadRequest))
		return
	}
	WriteCreated(w, map[string]any{"id": id, "message": "user created"})
}

func (u *Users) Item(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if len(parts) != 1 {
		WriteError(w, NewError("NOT_FOUND", "route not found", http.StatusNotFound))
		return
	}
	idStr := parts[0]

	userID, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid user ID", http.StatusBadRequest))
		return
	}

	if r.Method != http.MethodGet {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}
	user, err := db.GetUser(r.Context(), u.db, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteError(w, NewError("NOT_FOUND", "user not found", http.StatusNotFound))
		} else {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to get user", http.StatusInternalServerError))
		}
		return
	}
	WriteOK(w, user, nil)
}

func (u *Users) Me(w http.ResponseWriter, r *http.Request) {
	WriteError(w, NewError("UNAUTHORIZED", "login required", http.StatusUnauthorized))
}
