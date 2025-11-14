package handlers

import (
	"database/sql"
	"net/http"
)

type Categories struct{ db *sql.DB }

func NewCategories(database *sql.DB) *Categories { return &Categories{db: database} }

func (c *Categories) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}
	// TODO: Fetch categories from DB
	data := []map[string]any{
		{"id": 1, "name": "General"},
		{"id": 2, "name": "Tech"},
	}
	WriteOK(w, data, nil)
}
