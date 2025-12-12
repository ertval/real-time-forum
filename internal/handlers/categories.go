package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	repository "forum/internal/db"
)

type CategoriesHandler struct {
	conn *sql.DB
}

func NewCategoriesHandler(database *sql.DB) *CategoriesHandler {
	return &CategoriesHandler{conn: database}
}

//
// ---------------------------------------------------------
// COLLECTION HANDLER: GET (list) / POST (create)
// ---------------------------------------------------------
//

func (c *CategoriesHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		c.handleList(w, r)

	case http.MethodPost:
		c.handleCreate(w, r)

	default:
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
	}
}

//
// GET /api/v1/categories
//

func (c *CategoriesHandler) handleList(w http.ResponseWriter, r *http.Request) {
	categories, err := repository.ListCategories(r.Context(), c.conn)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "could not list categories", http.StatusInternalServerError))
		return
	}
	WriteOK(w, categories, nil)
}

//
// POST /api/v1/categories
//

func (c *CategoriesHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Name) == "" {
		WriteError(w, NewError("BAD_REQUEST", "invalid or missing name", http.StatusBadRequest))
		return
	}

	slug := slugify(in.Name)

	id, err := repository.CreateCategory(r.Context(), c.conn, in.Name, slug)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			WriteError(w, NewError("CONFLICT", "category name already exists", http.StatusConflict))
			return
		}
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to create category", http.StatusInternalServerError))
		return
	}

	category, err := repository.GetCategory(r.Context(), c.conn, id)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "created but failed to load", http.StatusInternalServerError))
		return
	}

	WriteCreated(w, category)
}

//
// ---------------------------------------------------------
// ITEM HANDLER: GET / PATCH / DELETE
// ---------------------------------------------------------
//

func (c *CategoriesHandler) Item(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/v1/categories/")
	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid category ID", http.StatusBadRequest))
		return
	}

	switch r.Method {

	case http.MethodGet:
		c.handleGet(w, r, id)

	case http.MethodPatch:
		c.handlePatch(w, r, id)

	case http.MethodDelete:
		c.handleDelete(w, r, id)

	default:
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
	}
}

//
// GET /api/v1/categories/{id}
//

func (c *CategoriesHandler) handleGet(w http.ResponseWriter, r *http.Request, id int64) {
	category, err := repository.GetCategory(r.Context(), c.conn, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, NewError("NOT_FOUND", "category not found", http.StatusNotFound))
			return
		}
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to load category", http.StatusInternalServerError))
		return
	}
	WriteOK(w, category, nil)
}

//
// PATCH /api/v1/categories/{id}
//

func (c *CategoriesHandler) handlePatch(w http.ResponseWriter, r *http.Request, id int64) {
	var in struct {
		Name *string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	if in.Name == nil || strings.TrimSpace(*in.Name) == "" {
		WriteError(w, NewError("BAD_REQUEST", "name cannot be empty", http.StatusBadRequest))
		return
	}

	err := repository.UpdateCategoryName(r.Context(), c.conn, id, *in.Name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			WriteError(w, NewError("CONFLICT", "category name already exists", http.StatusConflict))
			return
		}
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to update category", http.StatusInternalServerError))
		return
	}

	category, _ := repository.GetCategory(r.Context(), c.conn, id)
	WriteOK(w, category, nil)
}

//
// DELETE /api/v1/categories/{id}
//

func (c *CategoriesHandler) handleDelete(w http.ResponseWriter, r *http.Request, id int64) {
	err := repository.DeleteCategory(r.Context(), c.conn, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, NewError("NOT_FOUND", "category not found", http.StatusNotFound))
			return
		}
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to delete category", http.StatusInternalServerError))
		return
	}

	WriteNoContent(w)
}

//
// ---------------------------------------------------------
// HELPERS
// ---------------------------------------------------------

func parseID(path, prefix string) (int64, error) {
	raw := strings.TrimPrefix(path, prefix)
	return strconv.ParseInt(raw, 10, 64)
}

func slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}
