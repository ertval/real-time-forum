// Internal/handlers/posts_helpers.go
package handlers

import (
	"database/sql"
	"errors"
	repository "forum/internal/db"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func resolvePostRoute(w http.ResponseWriter, r *http.Request) (postID int64, action string, ok bool) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		notFound(w, r)
		return 0, "", false
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid post id", http.StatusBadRequest))
		return 0, "", false
	}

	if len(parts) > 1 {
		action = strings.Split(parts[1], "?")[0]
	}

	return id, action, true
}

func parsePostStatusFilter(r *http.Request) (*string, error) {
	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))

	switch status {
	case "", "all":
		return nil, nil
	case "draft", "published":
		return &status, nil
	default:
		return nil, errors.New("invalid status filter")
	}
}

func (p *PostsHandler) requirePostAuthor(w http.ResponseWriter, r *http.Request, postID int64, userID int64) bool {
	authorID, err := repository.GetPostAuthorID(r.Context(), p.conn, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r)
			return false
		}
		log.Printf("failed to load post author: %v", err)
		WriteError(w, r, NewError(
			"INTERNAL_SERVER_ERROR",
			"error loading post",
			http.StatusInternalServerError,
		))
		return false
	}

	if authorID != userID {
		WriteError(w, r, NewError(
			"FORBIDDEN",
			"not allowed",
			http.StatusForbidden,
		))
		return false
	}

	return true
}
