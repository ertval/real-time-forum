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

// ============================================================
// HandleCategories: /api/v1/categories
// ============================================================

func (c *CategoriesHandler) HandleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.listCategories(w, r)
	case http.MethodPost:
		c.createCategory(w, r)
	default:
		MethodNotAllowed(w)
	}
}

// ------------------------------------------------------------
// Internal helpers
// ------------------------------------------------------------

func resolveCategoryID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := parseID(r.URL.Path, "/api/v1/categories/")
	if err != nil || id <= 0 {
		WriteError(w, NewError(
			"BAD_REQUEST",
			"invalid category ID",
			http.StatusBadRequest,
		))
		return 0, false
	}
	return id, true
}

// ------------------------------------------------------------
// LIST CATEGORIES
// ------------------------------------------------------------

func (c *CategoriesHandler) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := repository.ListCategories(r.Context(), c.conn)
	if writeHandlerError(w, err, "could not list categories") {
		return
	}

	WriteOK(w, categories, nil)
}

// ------------------------------------------------------------
// CREATE CATEGORY
// ------------------------------------------------------------

func (c *CategoriesHandler) createCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		strings.TrimSpace(req.Name) == "" {

		WriteError(w, NewError(
			"BAD_REQUEST",
			"name is required",
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

		writeHandlerError(w, err, "failed to create category")
		return
	}

	category, err := repository.GetCategory(r.Context(), c.conn, id)
	if writeHandlerError(w, err, "category created but failed to load") {
		return
	}

	WriteCreated(w, category)
}

// ============================================================
// HandleCategory: /api/v1/categories/{id}
// ============================================================

func (c *CategoriesHandler) HandleCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := resolveCategoryID(w, r)
	if !ok {
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
		MethodNotAllowed(w)
	}
}

// ------------------------------------------------------------
// GET CATEGORY
// ------------------------------------------------------------

func (c *CategoriesHandler) getCategory(w http.ResponseWriter, r *http.Request, id int64) {
	category, err := repository.GetCategory(r.Context(), c.conn, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}

		writeHandlerError(w, err, "failed to load category")
		return
	}

	WriteOK(w, category, nil)
}

// ------------------------------------------------------------
// UPDATE CATEGORY (PATCH)
// ------------------------------------------------------------

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
			"name is required",
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

		writeHandlerError(w, err, "failed to update category")
		return
	}

	category, err := repository.GetCategory(r.Context(), c.conn, id)
	if writeHandlerError(w, err, "category updated but failed to load") {
		return
	}

	WriteOK(w, category, nil)
}

// ------------------------------------------------------------
// DELETE CATEGORY
// ------------------------------------------------------------

func (c *CategoriesHandler) deleteCategory(w http.ResponseWriter, r *http.Request, id int64) {
	if err := repository.DeleteCategory(r.Context(), c.conn, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}

		writeHandlerError(w, err, "failed to delete category")
		return
	}

	WriteNoContent(w)
}
