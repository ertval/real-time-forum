package handlers

import (
	"net/http"

	repository "forum/internal/db"
)

func (p *PostsHandler) PublicList(w http.ResponseWriter, r *http.Request) {
	page, per := sanitizePagination(r)
	sort := sanitizeSort(r)

	result, err := repository.ListPublicPosts(
		r.Context(),
		p.conn,
		repository.ListPublicPostsParams{
			Page:    page,
			PerPage: per,
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

	meta := buildMeta(page, per, result.Total, map[string]any{
		"sort": sort,
	})

	WriteOK(w, result.Posts, meta)
}
