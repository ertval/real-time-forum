package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	db "forum/internal/db"
)

type Posts struct{ db *sql.DB }

func NewPosts(database *sql.DB) *Posts { return &Posts{db: database} }

// ---------------------------------------------------------
//
//	/api/v1/posts  → GET list, POST create (with categories)
//
// ---------------------------------------------------------
func (p *Posts) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	// ----------------------------
	// GET /posts → list with meta
	// ----------------------------
	case http.MethodGet:
		page := parseIntOr(r.URL.Query().Get("page"), 1)
		per := parseIntOr(r.URL.Query().Get("per_page"), 20)

		if per > 100 {
			per = 100
		}
		if page < 1 {
			page = 1
		}

		res, err := db.ListPosts(r.Context(), p.db, db.ListPostsParams{
			Page:    page,
			PerPage: per,
		})
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error listing posts", http.StatusInternalServerError))
			return
		}

		meta := map[string]any{
			"page":     page,
			"per_page": per,
			"total":    res.Total,
		}

		WriteOK(w, res.Posts, meta)

	// -------------------------------------------------
	// POST /posts → create post + optional categories
	// -------------------------------------------------
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

		// TODO: replace fake user when auth fully implemented
		const fakeAuthorID = 1

		// ----------------------------
		// Create with categories
		// ----------------------------
		postID, err := db.CreatePostWithCategories(r.Context(), p.db,
			fakeAuthorID,
			in.Title,
			in.Body,
			in.CategoryIDs,
		)
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError))
			return
		}

		// load full post after creation
		post, err := db.GetPost(r.Context(), p.db, postID)
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "post created but failed to load", http.StatusInternalServerError))
			return
		}

		WriteCreated(w, post)

	default:
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
	}
}

// -----------------------------------------------------------
//
//	/api/v1/posts/{id}  (+ /comments, /like subroutes)
//
// -----------------------------------------------------------
func (p *Posts) Item(w http.ResponseWriter, r *http.Request) {
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

	// ------------------------------------------------------
	// /posts/{id}
	// ------------------------------------------------------
	if len(parts) == 1 {
		switch r.Method {

		// --------------------------------------------------
		// GET /posts/{id}
		// --------------------------------------------------
		case http.MethodGet:
			post, err := db.GetPost(r.Context(), p.db, postID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					WriteError(w, NewError("NOT_FOUND", "post not found", http.StatusNotFound))
					return
				}
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error getting post", http.StatusInternalServerError))
				return
			}
			WriteOK(w, post, nil)

		// --------------------------------------------------
		// PATCH /posts/{id} → update title/body only
		// --------------------------------------------------
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

			if err := db.UpdatePost(r.Context(), p.db, postID, db.UpdatePostInput{
				Title: in.Title,
				Body:  in.Body,
			}); err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error updating post", http.StatusInternalServerError))
				return
			}

			WriteOK(w, map[string]string{"status": "updated"}, nil)

		// --------------------------------------------------
		// DELETE /posts/{id}
		// --------------------------------------------------
		case http.MethodDelete:
			if err := db.DeletePost(r.Context(), p.db, postID); err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error deleting post", http.StatusInternalServerError))
				return
			}
			WriteNoContent(w)

		default:
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		}

		return
	}

	// ------------------------------------------------------
	// /posts/{id}/like
	// ------------------------------------------------------
	switch parts[1] {

	case "like":
		if r.Method != http.MethodPost {
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
			return
		}

		// fake user
		const fakeID int64 = 1

		liked, count, err := db.TogglePostLike(r.Context(), p.db, fakeID, postID)
		if err != nil {
			log.Printf("TogglePostLike failed for user=%d post=%d: %v", fakeID, postID, err)
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error toggling like", http.StatusInternalServerError))
			return
		}

		WriteOK(w, map[string]any{
			"post_id": postID,
			"liked":   liked,
			"likes":   count,
		}, nil)

	// ------------------------------------------------------
	// /posts/{id}/comments
	// ------------------------------------------------------
	case "comments":
		switch r.Method {

		case http.MethodGet:
			page := parseIntOr(r.URL.Query().Get("page"), 1)
			per := parseIntOr(r.URL.Query().Get("per_page"), 20)
			if per > 100 {
				per = 100
			}
			if page < 1 {
				page = 1
			}

			res, err := db.ListCommentsByPost(r.Context(), p.db, db.ListCommentsParams{
				PostID:  postID,
				Page:    page,
				PerPage: per,
			})
			if err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error listing comments", http.StatusInternalServerError))
				return
			}

			meta := map[string]any{
				"page":     page,
				"per_page": per,
				"total":    res.Total,
				"post_id":  postID,
			}

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

			commentID, err := db.CreateComment(r.Context(), p.db, db.CreateCommentInput{
				PostID:          postID,
				UserID:          fakeUserID,
				ParentCommentID: in.ParentCommentID,
				Body:            in.Body,
			})
			if err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error creating comment", http.StatusInternalServerError))
				return
			}

			comment, err := db.GetComment(r.Context(), p.db, commentID)
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

func parseIntOr(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
