package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	db "forum/internal/db"
)

type Categories struct {
	db *sql.DB
}

func NewCategories(database *sql.DB) *Categories { return &Categories{db: database} }

// ---------------------------------------------------------
// COLLECTION HANDLER: GET (list) / POST (create)
// ---------------------------------------------------------

func (c *Categories) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		categories, err := db.ListCategories(r.Context(), c.db)
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error listing categories", http.StatusInternalServerError))
			return
		}
		WriteOK(w, categories, nil)

	case http.MethodPost:
		var in struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Name) == "" {
			WriteError(w, NewError("BAD_REQUEST", "invalid or missing name", http.StatusBadRequest))
			return
		}

		slug := slugify(in.Name)

		id, err := db.CreateCategory(r.Context(), c.db, in.Name, slug)
		if err != nil {
			// UNIQUE name/slug conflict
			if strings.Contains(err.Error(), "UNIQUE") {
				WriteError(w, NewError("CONFLICT", "category name already exists", http.StatusConflict))
				return
			}
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error creating category", http.StatusInternalServerError))
			return
		}

		category, err := db.GetCategory(r.Context(), c.db, id)
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "created but failed to load", http.StatusInternalServerError))
			return
		}

		WriteCreated(w, category)

	default:
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
	}
}

// ---------------------------------------------------------
// ITEM HANDLER: GET / PATCH / DELETE
// ---------------------------------------------------------

func (c *Categories) Item(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/categories/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid category ID", http.StatusBadRequest))
		return
	}

	switch r.Method {

	case http.MethodGet:
		category, err := db.GetCategory(r.Context(), c.db, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				WriteError(w, NewError("NOT_FOUND", "category not found", http.StatusNotFound))
				return
			}
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error loading category", http.StatusInternalServerError))
			return
		}
		WriteOK(w, category, nil)

	case http.MethodPatch:
		var in struct {
			Name *string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
			return
		}

		if in.Name == nil {
			WriteError(w, NewError("BAD_REQUEST", "nothing to update", http.StatusBadRequest))
			return
		}

		if strings.TrimSpace(*in.Name) == "" {
			WriteError(w, NewError("BAD_REQUEST", "name cannot be empty", http.StatusBadRequest))
			return
		}

		// IMPORTANT: slug NEVER changes (SEO-safe)
		err = db.UpdateCategoryName(r.Context(), c.db, id, *in.Name)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				WriteError(w, NewError("CONFLICT", "category name already exists", http.StatusConflict))
				return
			}
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error updating category", http.StatusInternalServerError))
			return
		}

		category, _ := db.GetCategory(r.Context(), c.db, id)
		WriteOK(w, category, nil)

	case http.MethodDelete:
		err := db.DeleteCategory(r.Context(), c.db, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				WriteError(w, NewError("NOT_FOUND", "category not found", http.StatusNotFound))
				return
			}
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error deleting category", http.StatusInternalServerError))
			return
		}

		WriteNoContent(w)

	default:
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
	}
}

// ---------------------------------------------------------
// HELPER — Slugify
// ---------------------------------------------------------

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
