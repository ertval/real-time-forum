// /internal/handlers/posts_public.go
package handlers

import (
	"log"
	"net/http"

	repository "forum/internal/db"
)

// ============================================================
// PUBLIC POSTS
// ============================================================

// PublicList handles:
// GET /api/v1/posts/public
func (p *PostsHandler) PublicList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w, r)
		return
	}

	p.publicListPosts(w, r)
}

// ------------------------------------------------------------
// Internal helpers
// ------------------------------------------------------------

func (p *PostsHandler) publicListPosts(w http.ResponseWriter, r *http.Request) {
	page, perPage := sanitizePagination(r)
	sort := sanitizeSort(r)

	result, err := repository.ListPublicPosts(
		r.Context(),
		p.conn,
		repository.ListPublicPostsParams{
			Page:    page,
			PerPage: perPage,
			SortBy:  sort,
		},
	)
	if err != nil {
		log.Printf("failed to list public posts: %v", err)
		WriteError(w, r, NewError(
			"INTERNAL_SERVER_ERROR",
			"failed to list posts",
			http.StatusInternalServerError,
		))
		return
	}

	// ------------------------------------------------------------
	// Build pagination metadata
	// ------------------------------------------------------------

	totalPages := 0
	if perPage > 0 {
		totalPages = (result.Total + perPage - 1) / perPage
	}

	meta := &Meta{
		Pagination: &PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	}

	WriteOK(w, result.Posts, meta)
}
