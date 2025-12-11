package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	repository "forum/internal/db"
)

type PostsHandler struct {
	conn *sql.DB
}

func NewPostsHandler(database *sql.DB) *PostsHandler {
	return &PostsHandler{conn: database}
}

//
// ─────────────────────────────────────────────────────────────
//  /api/v1/posts → GET list, POST create
// ─────────────────────────────────────────────────────────────
//

func (p *PostsHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		// Pagination
		page, per := sanitizePagination(r)

		result, err := repository.ListPosts(
			r.Context(),
			p.conn,
			repository.ListPostsParams{
				Page:    page,
				PerPage: per,
			},
		)
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error listing posts", http.StatusInternalServerError))
			return
		}

		meta := buildMeta(page, per, result.Total, nil)
		WriteOK(w, result.Posts, meta)

	case http.MethodPost:
		var in struct {
			Title       string  `json:"title"`
			Body        string  `json:"body"`
			CategoryIDs []int64 `json:"category_ids"`
		}

		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
			return
		}

		if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Body) == "" {
			WriteError(w, NewError("BAD_REQUEST", "title and body required", http.StatusBadRequest))
			return
		}

		// TODO: replace with authenticated user
		const fakeAuthorID = 1

		postID, err := repository.CreatePostWithCategories(r.Context(), p.conn,
			fakeAuthorID,
			in.Title,
			in.Body,
			in.CategoryIDs,
		)
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError))
			return
		}

		post, err := repository.GetPost(r.Context(), p.conn, postID)
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "post created but failed to load", http.StatusInternalServerError))
			return
		}

		WriteCreated(w, post)

	default:
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
	}
}

//
// ─────────────────────────────────────────────────────────────
//  /api/v1/posts/{id} (+ comments & like subroutes)
// ─────────────────────────────────────────────────────────────
//

func (p *PostsHandler) Item(w http.ResponseWriter, r *http.Request) {
	tail := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	parts := strings.Split(strings.Trim(tail, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		WriteError(w, NewError("NOT_FOUND", "route not found", http.StatusNotFound))
		return
	}

	idStr := parts[0]
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid post ID", http.StatusBadRequest))
		return
	}

	//
	// /posts/{id}
	//
	if len(parts) == 1 {
		switch r.Method {

		case http.MethodGet:
			post, err := repository.GetPost(r.Context(), p.conn, postID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					WriteError(w, NewError("NOT_FOUND", "post not found", http.StatusNotFound))
					return
				}
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error getting post", http.StatusInternalServerError))
				return
			}
			WriteOK(w, post, nil)

		case http.MethodPatch:
			var in struct {
				Title *string `json:"title"`
				Body  *string `json:"body"`
			}

			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				WriteError(w, NewError("BAD_REQUEST", "invalid JSON", http.StatusBadRequest))
				return
			}

			if in.Title == nil && in.Body == nil {
				WriteError(w, NewError("BAD_REQUEST", "nothing to update", http.StatusBadRequest))
				return
			}

			err = repository.UpdatePost(r.Context(), p.conn, postID, repository.UpdatePostInput{
				Title: in.Title,
				Body:  in.Body,
			})
			if err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error updating post", http.StatusInternalServerError))
				return
			}

			WriteOK(w, map[string]string{"status": "updated"}, nil)

		case http.MethodDelete:
			if err := repository.DeletePost(r.Context(), p.conn, postID); err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error deleting post", http.StatusInternalServerError))
				return
			}
			WriteNoContent(w)

		default:
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		}
		return
	}

	//
	// Subroutes: /posts/{id}/like /posts/{id}/comments
	//

	switch parts[1] {

	//
	// /posts/{id}/like
	//
	case "like":
		if r.Method != http.MethodPost {
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
			return
		}

		const fakeUserID int64 = 1

		liked, count, err := repository.TogglePostLike(r.Context(), p.conn, fakeUserID, postID)
		if err != nil {
			log.Printf("TogglePostLike failed for user=%d post=%d: %v", fakeUserID, postID, err)
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error toggling like", http.StatusInternalServerError))
			return
		}

		WriteOK(w, map[string]any{
			"post_id": postID,
			"liked":   liked,
			"likes":   count,
		}, nil)

	//
	// /posts/{id}/comments
	//
	case "comments":
		switch r.Method {

		case http.MethodGet:
			page, per := sanitizePagination(r)

			res, err := repository.ListCommentsByPost(r.Context(), p.conn, repository.ListCommentsParams{
				PostID:  postID,
				Page:    page,
				PerPage: per,
			})
			if err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error listing comments", http.StatusInternalServerError))
				return
			}

			meta := buildMeta(page, per, res.Total, map[string]any{"post_id": postID})
			WriteOK(w, res.Comments, meta)

		case http.MethodPost:
			var in struct {
				Body            string `json:"body"`
				ParentCommentID *int64 `json:"parent_comment_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
				return
			}

			if strings.TrimSpace(in.Body) == "" {
				WriteError(w, NewError("BAD_REQUEST", "body required", http.StatusBadRequest))
				return
			}

			const fakeUserID int64 = 1

			commentID, err := repository.CreateComment(r.Context(), p.conn, repository.CreateCommentInput{
				PostID:          postID,
				UserID:          fakeUserID,
				ParentCommentID: in.ParentCommentID,
				Body:            in.Body,
			})
			if err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error creating comment", http.StatusInternalServerError))
				return
			}

			comment, err := repository.GetComment(r.Context(), p.conn, commentID)
			if err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "comment created but failed to load", http.StatusInternalServerError))
				return
			}

			WriteCreated(w, comment)

		default:
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		}

	default:
		WriteError(w, NewError("NOT_FOUND", "route not found", http.StatusNotFound))
	}
}

//
// ─────────────────────────────────────────────────────────────
//  PUBLIC POSTS ENDPOINT (/api/v1/posts/public)
// ─────────────────────────────────────────────────────────────
//

func (p *PostsHandler) PublicList(w http.ResponseWriter, r *http.Request) {
	page, per := sanitizePagination(r)
	sort := sanitizeSort(r)

	result, err := repository.ListPublicPosts(r.Context(), p.conn, repository.ListPublicPostsParams{
		Page:    page,
		PerPage: per,
		SortBy:  sort,
	})
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to list posts", http.StatusInternalServerError))
		return
	}

	meta := buildMeta(page, per, result.Total, map[string]any{"sort": sort})
	WriteOK(w, result.Posts, meta)
}
