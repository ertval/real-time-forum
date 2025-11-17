package handlers

import (
	"database/sql"
	"forum/internal/db"
	"net/http"
)

type Categories struct{ db *sql.DB }

func NewCategories(database *sql.DB) *Categories { return &Categories{db: database} }

func (c *Categories) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}
	cats, err := db.ListCategories(r.Context(), c.db)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error listing categories", http.StatusInternalServerError))
	}
	WriteOK(w, cats, nil)
}
