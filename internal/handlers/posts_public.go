package handlers

import (
	"net/http"

	repository "forum/internal/db"
)

func (p *PostsHandler) PublicList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	p.publicListPosts(w, r)
}

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
		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"failed to list posts",
			http.StatusInternalServerError,
		))
		return
	}

	paginationMeta := buildPaginationInfo(page, perPage, result.Total, map[string]any{
		"sort": sort,
	})

	WriteOK(w, result.Posts, paginationMeta)
}
