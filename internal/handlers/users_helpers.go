package handlers

import (
	"database/sql"
	"net/http"
)

type UsersHandler struct {
	conn *sql.DB
}

func NewUsersHandler(database *sql.DB) *UsersHandler {
	return &UsersHandler{conn: database}
}

func resolveUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := parseID(r.URL.Path, "/api/v1/users/")
	if err != nil || id <= 0 {
		WriteError(w, r, NewError(
			"BAD_REQUEST",
			"invalid user ID",
			http.StatusBadRequest,
		))
		return 0, false
	}
	return id, true
}
