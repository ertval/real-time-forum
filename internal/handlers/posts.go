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

type Posts struct{ db *sql.DB }

func NewPosts(database *sql.DB) *Posts { return &Posts{db: database} }

// Collection is the main endpoint for posts
func (p *Posts) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//lists posts
		page := parseIntOr(r.URL.Query().Get("page"), 1)
		per := parseIntOr(r.URL.Query().Get("per_page"), 20)
		//falls back to 100 so no one can get say 10,000 elements
		if per > 100 {
			per = 100
		}
		if page < 1 {
			page = 1
		}
		res, err := db.ListPosts(r.Context(), p.db, db.ListPostsParams{Page: page, PerPage: per})
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error listing posts", http.StatusInternalServerError))
			return
		}
		meta := map[string]any{"page": page, "per_page": per, "total": res.Total}
		WriteOK(w, res.Posts, meta)

	case http.MethodPost:
		var in struct {
			Title      string `json:"title"`
			Body       string `json:"body"`
			CategoryID *int64 `json:"category_id"`
			//authorID to be added once auth is implemented.
			//For now hardcoded to 1
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" {
			WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
			return
		}
		if strings.TrimSpace(in.Body) == "" || strings.TrimSpace(in.Title) == "" {
			WriteError(w, NewError("BAD_REQUEST", "title and body required", http.StatusBadRequest))
			return
		}

		//TODO once auth exists, take authorID from authentication. For now its hard-coded
		const fakeAuthorID = 1
		newID, err := db.CreatePost(r.Context(), p.db, db.CreatePostInput{
			AuthorID:   fakeAuthorID,
			Title:      in.Title,
			Body:       in.Body,
			CategoryID: in.CategoryID,
		})
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error creating post", http.StatusInternalServerError))
			return
		}
		//Gets after creating to also grab auto-filled fields from database
		post, err := db.GetPost(r.Context(), p.db, newID)
		if err != nil {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "post created but failed to load", http.StatusInternalServerError))
			return
		}
		WriteCreated(w, post)
	default:
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
	}
}

// Item is the endpoint for a single post and subresources such as /comments, /like
func (p *Posts) Item(w http.ResponseWriter, r *http.Request) {
	tail := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	parts := strings.Split(strings.Trim(tail, "/"), "/")

	//case nothing past tail or empty ID
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

	//case /api/v1/posts/{id}
	if len(parts) == 1 {
		switch r.Method {
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
		case http.MethodDelete:
			if err := db.DeletePost(r.Context(), p.db, postID); err != nil {
				WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error deleting post", http.StatusInternalServerError))
				return
			}
			WriteNoContent(w)
		default:
			WriteError(w, NewError("METHOD_NOT_ALLOWED", " method not allowed", http.StatusMethodNotAllowed))
		}
		return
	}
	//case /api/v1/posts/{id}/...
	switch parts[1] {
	case "like":
		if r.Method != http.MethodPost {
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
			return
		}
		WriteOK(w, map[string]string{"status": "toggled"}, nil)

	case "comments":
		switch r.Method {
		case http.MethodGet:
			//list comments
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
			//TODO: once auth exists, use authorID from authentication. For now its hard-coded
			const fakeAuthorID = 1

			commentID, err := db.CreateComment(r.Context(), p.db, db.CreateCommentInput{
				PostID:          postID,
				UserID:          fakeAuthorID,
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

// Unused at the moment
func (p *Posts) List(w http.ResponseWriter, r *http.Request) {
	//Requests page number & per page elements
	//default page num is 1 and per page elements is 20
	page := parseIntOr(r.URL.Query().Get("page"), 1)
	per := parseIntOr(r.URL.Query().Get("per_page"), 20)
	//falls back to 100 so no one can get say 10,000 elements
	if per > 100 {
		per = 100
	}
	if page < 1 {
		page = 1
	}
	//TODO Populate data from database by linking it to API's
	data := []any{}
	meta := map[string]any{"page": page, "per_page": per, "total": 0}
	WriteOK(w, data, meta)
}

func parseIntOr(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
