package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	repository "forum/internal/db"
)

type PostsHandler struct {
	conn *sql.DB
}

func NewPostsHandler(database *sql.DB) *PostsHandler {
	return &PostsHandler{conn: database}
}

// ============================================================
// HandlePosts: /api/v1/posts
// ============================================================

func (p *PostsHandler) HandlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		p.listPosts(w, r)
	case http.MethodPost:
		p.createPost(w, r)
	default:
		MethodNotAllowed(w)
	}
}

func (p *PostsHandler) listPosts(w http.ResponseWriter, r *http.Request) {
	page, perPage := sanitizePagination(r)

	result, err := repository.ListPosts(
		r.Context(),
		p.conn,
		repository.ListPostsParams{
			Page:    page,
			PerPage: perPage,
		},
	)
	if err != nil {
		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"error listing posts",
			http.StatusInternalServerError,
		))
		return
	}

	paginationInfo := buildPaginationInfo(page, perPage, result.Total, nil)
	WriteOK(w, result.Posts, paginationInfo)
}

func (p *PostsHandler) createPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req struct {
		Title       string  `json:"title"`
		Body        string  `json:"body"`
		CategoryIDs []int64 `json:"category_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Body) == "" {
		WriteError(w, NewError("BAD_REQUEST", "title and body required", http.StatusBadRequest))
		return
	}

	postID, err := repository.CreatePostWithCategories(
		r.Context(),
		p.conn,
		userID,
		req.Title,
		req.Body,
		req.CategoryIDs,
	)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error creating post", http.StatusInternalServerError))
		return
	}

	// Reload post to return the full persisted representation
	post, err := repository.GetPost(r.Context(), p.conn, postID)
	if err != nil {
		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"post created but failed to load",
			http.StatusInternalServerError,
		))
		return
	}

	WriteCreated(w, post)
}

// ============================================================
// HandlePost: /api/v1/posts/{id}
// ============================================================

func (p *PostsHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	postID, action, ok := resolvePostRoute(w, r)
	if !ok {
		return
	}

	switch action {

	case "":
		switch r.Method {
		case http.MethodGet:
			p.getPost(w, r, postID)
		case http.MethodPatch:
			p.updatePost(w, r, postID)
		case http.MethodDelete:
			p.deletePost(w, r, postID)
		default:
			MethodNotAllowed(w)
		}

	case "like":
		if r.Method == http.MethodPost {
			p.handleReaction(w, r, postID, 1, "post")
			return
		}
		MethodNotAllowed(w)

	case "dislike":
		if r.Method == http.MethodPost {
			p.handleReaction(w, r, postID, -1, "post")
			return
		}
		MethodNotAllowed(w)

	case "comments":
		switch r.Method {
		case http.MethodGet:
			p.listComments(w, r, postID)
		case http.MethodPost:
			p.createComment(w, r, postID)
		default:
			MethodNotAllowed(w)
		}

	default:
		notFound(w)
	}
}

// ============================================================
// POST ACTIONS
// ============================================================

func (p *PostsHandler) getPost(w http.ResponseWriter, r *http.Request, postID int64) {
	post, err := repository.GetPost(r.Context(), p.conn, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error loading post", http.StatusInternalServerError))
		return
	}
	WriteOK(w, post, nil)
}

func (p *PostsHandler) updatePost(w http.ResponseWriter, r *http.Request, postID int64) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}

	var req struct {
		Title *string `json:"title"`
		Body  *string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	if req.Title == nil && req.Body == nil {
		WriteError(w, NewError("BAD_REQUEST", "nothing to update", http.StatusBadRequest))
		return
	}

	if err := repository.UpdatePost(
		r.Context(),
		p.conn,
		postID,
		repository.UpdatePostInput{Title: req.Title, Body: req.Body},
	); err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error updating post", http.StatusInternalServerError))
		return
	}

	WriteOK(w, map[string]string{"status": "updated"}, nil)
}

func (p *PostsHandler) deletePost(w http.ResponseWriter, r *http.Request, postID int64) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}

	if err := repository.DeletePost(r.Context(), p.conn, postID); err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error deleting post", http.StatusInternalServerError))
		return
	}

	WriteNoContent(w)
}

func (p *PostsHandler) handleReaction(
	w http.ResponseWriter,
	r *http.Request,
	objectID int64,
	targetReaction int,
	targetType string,
) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	reaction, err := repository.ToggleReaction(
		r.Context(),
		p.conn,
		userID,
		objectID,
		targetReaction,
		targetType,
	)
	if err != nil {
		log.Printf("ToggleReaction failed: %v", err)
		WriteError(
			w,
			NewError("INTERNAL_SERVER_ERROR", "error toggling reaction", http.StatusInternalServerError),
		)
		return
	}
	var likesCount, dislikesCount int
	switch targetType {
	case "post":
		likesCount, dislikesCount, err = repository.CountReactionsForPost(r.Context(), p.conn, objectID)
	case "comment":
		likesCount, dislikesCount, err = repository.CountReactionsForComment(r.Context(), p.conn, objectID)
	default:
		WriteError(w, NewError("BAD_REQUEST", "invalid target type", http.StatusBadRequest))
		return
	}
	if err != nil {
		log.Printf("CountPostLikes failed: %v", err)
		WriteError(
			w,
			NewError("INTERNAL_SERVER_ERROR", "error counting likes", http.StatusInternalServerError),
		)
		return
	}

	idKey := "post_id"
	if targetType == "comment" {
		idKey = "comment_id"
	}

	WriteOK(w, map[string]any{
		idKey:            objectID,
		"reaction":       reaction,
		"likes_count":    likesCount,
		"dislikes_count": dislikesCount,
	}, nil)
}

func (p *PostsHandler) listComments(w http.ResponseWriter, r *http.Request, postID int64) {
	page, perPage := sanitizePagination(r)

	listResult, err := repository.ListCommentsByPost(
		r.Context(),
		p.conn,
		repository.ListCommentsParams{
			PostID:  postID,
			Page:    page,
			PerPage: perPage,
		},
	)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error listing comments", http.StatusInternalServerError))
		return
	}

	paginationInfo := buildPaginationInfo(page, perPage, listResult.Total, map[string]any{
		"post_id": postID,
	})

	WriteOK(w, listResult.Comments, paginationInfo)
}

func (p *PostsHandler) createComment(w http.ResponseWriter, r *http.Request, postID int64) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req struct {
		Body            string `json:"body"`
		ParentCommentID *int64 `json:"parent_comment_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	if strings.TrimSpace(req.Body) == "" {
		WriteError(w, NewError("BAD_REQUEST", "body required", http.StatusBadRequest))
		return
	}

	commentID, err := repository.CreateComment(
		r.Context(),
		p.conn,
		repository.CreateCommentInput{
			PostID:          postID,
			UserID:          userID,
			ParentCommentID: req.ParentCommentID,
			Body:            req.Body,
		},
	)
	if err != nil {
		WriteError(w, NewError(
			"INTERNAL_SERVER_ERROR",
			"error creating comment",
			http.StatusInternalServerError,
		))
		return
	}

	comment := repository.Comment{
		ID:              commentID,
		PostID:          postID,
		UserID:          userID,
		ParentCommentID: req.ParentCommentID,
		Body:            req.Body,
	}

	WriteCreated(w, comment)
}
