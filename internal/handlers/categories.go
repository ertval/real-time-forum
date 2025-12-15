package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	repository "forum/internal/db"
)

type CategoriesHandler struct {
	conn *sql.DB
}

func NewCategoriesHandler(database *sql.DB) *CategoriesHandler {
	return &CategoriesHandler{conn: database}
}

// ---------------------------------------------------------
// HandleCategories: GET / POST
// ---------------------------------------------------------

func (c *CategoriesHandler) HandleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.listCategories(w, r)
	case http.MethodPost:
		c.createCategory(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (c *CategoriesHandler) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := repository.ListCategories(r.Context(), c.conn)
	if err != nil {
		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"could not list categories",
			http.StatusInternalServerError,
		))
		return
	}

	WriteOK(w, categories, nil)
}

func (c *CategoriesHandler) createCategory(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		strings.TrimSpace(req.Name) == "" {

		WriteError(w, NewError(
			"BAD_REQUEST",
			"invalid or missing name",
			http.StatusBadRequest,
		))
		return
	}

	slug := slugify(req.Name)

	id, err := repository.CreateCategory(
		r.Context(),
		c.conn,
		req.Name,
		slug,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			WriteError(w, NewError(
				"CONFLICT",
				"category name already exists",
				http.StatusConflict,
			))
			return
		}

		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"failed to create category",
			http.StatusInternalServerError,
		))
		return
	}

	category, err := repository.GetCategory(r.Context(), c.conn, id)
	if err != nil {
		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"created but failed to load category",
			http.StatusInternalServerError,
		))
		return
	}

	WriteCreated(w, category)
}

// ---------------------------------------------------------
// HandleCategory: GET / PATCH / DELETE
// ---------------------------------------------------------

func (c *CategoriesHandler) HandleCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/v1/categories/")
	if err != nil {
		WriteError(w, NewError(
			"BAD_REQUEST",
			"invalid category ID",
			http.StatusBadRequest,
		))
		return
	}

	switch r.Method {
	case http.MethodGet:
		c.getCategory(w, r, id)
	case http.MethodPatch:
		c.updateCategory(w, r, id)
	case http.MethodDelete:
		c.deleteCategory(w, r, id)
	default:
		methodNotAllowed(w)
	}

}

func (c *CategoriesHandler) getCategory(w http.ResponseWriter, r *http.Request, id int64) {
	category, err := repository.GetCategory(r.Context(), c.conn, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, NewError(
				"NOT_FOUND",
				"category not found",
				http.StatusNotFound,
			))
			return
		}

		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"failed to load category",
			http.StatusInternalServerError,
		))
		return
	}

	WriteOK(w, category, nil)
}

func (c *CategoriesHandler) updateCategory(w http.ResponseWriter, r *http.Request, id int64) {
	var req struct {
		Name *string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError(
			"BAD_REQUEST",
			"invalid json",
			http.StatusBadRequest,
		))
		return
	}

	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		WriteError(w, NewError(
			"BAD_REQUEST",
			"name cannot be empty",
			http.StatusBadRequest,
		))
		return
	}

	if err := repository.UpdateCategoryName(
		r.Context(),
		c.conn,
		id,
		*req.Name,
	); err != nil {

		if isUniqueConstraint(err) {
			WriteError(w, NewError(
				"CONFLICT",
				"category name already exists",
				http.StatusConflict,
			))
			return
		}

		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"failed to update category",
			http.StatusInternalServerError,
		))
		return
	}

	category, err := repository.GetCategory(r.Context(), c.conn, id)
	if err != nil {
		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"updated but failed to load category",
			http.StatusInternalServerError,
		))
		return
	}

	WriteOK(w, category, nil)
}

func (c *CategoriesHandler) deleteCategory(w http.ResponseWriter, r *http.Request, id int64) {
	if err := repository.DeleteCategory(r.Context(), c.conn, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, NewError(
				"NOT_FOUND",
				"category not found",
				http.StatusNotFound,
			))
			return
		}

		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"failed to delete category",
			http.StatusInternalServerError,
		))
		return
	}

	WriteNoContent(w)
}
