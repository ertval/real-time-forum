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

const (
	ReactionTargetPost    = "post"
	ReactionTargetComment = "comment"
)

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
		MethodNotAllowed(w, r)
	}
}

// ============================================================
// LIST POSTS
// ============================================================

func (p *PostsHandler) listPosts(w http.ResponseWriter, r *http.Request) {
	page, perPage := sanitizePagination(r)

	// Filter by category (optional)
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil || categoryID <= 0 {
			WriteError(w, r, NewError("BAD_REQUEST", "invalid category_id", http.StatusBadRequest))
			return
		}

		result, err := repository.ListPostsByCategory(
			r.Context(),
			p.conn,
			repository.ListPostsByCategoryParams{
				CategoryID: categoryID,
				Page:       page,
				PerPage:    perPage,
			},
		)
		if err != nil {
			log.Printf("failed to list posts by category: %v", err)
			WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error listing posts", http.StatusInternalServerError))
			return
		}

		pagination := buildPaginationInfo(page, perPage, result.Total, map[string]any{
			"category_id": categoryID,
		})

		WriteOK(w, result.Posts, pagination)
		return
	}

	// Default: list all posts
	result, err := repository.ListPosts(
		r.Context(),
		p.conn,
		repository.ListPostsParams{
			Page:    page,
			PerPage: perPage,
		},
	)
	if err != nil {
		log.Printf("failed to list posts: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error listing posts", http.StatusInternalServerError))
		return
	}

	pagination := buildPaginationInfo(page, perPage, result.Total, nil)
	WriteOK(w, result.Posts, pagination)
}

// ListPublicPosts handles:
// GET /api/v1/posts/public
func (p *PostsHandler) ListPublicPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w, r)
		return
	}

	p.listPosts(w, r)
}

// ============================================================
// CREATE POST
// ============================================================

func (p *PostsHandler) createPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req struct {
		Title       string  `json:"title"`
		Body        string  `json:"body"`
		Status      string  `json:"status"`
		CategoryIDs []int64 `json:"category_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		WriteError(w, r, NewError("BAD_REQUEST", "title required", http.StatusBadRequest))
		return
	}

	status := "published"
	if strings.ToLower(req.Status) == "draft" {
		status = "draft"
	}

	postID, err := repository.CreatePostWithCategories(
		r.Context(),
		p.conn,
		userID,
		req.Title,
		req.Body,
		status,
		req.CategoryIDs,
	)

	if err != nil {
		log.Printf("failed to create post: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error creating post", http.StatusInternalServerError))
		return
	}

	post, err := repository.GetPost(r.Context(), p.conn, postID)
	if err != nil {
		log.Printf("failed to load post after creation: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "post created but failed to load", http.StatusInternalServerError))
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
			MethodNotAllowed(w, r)
		}

	case "like":
		if r.Method == http.MethodPost {
			p.handleReaction(w, r, postID, 1, ReactionTargetPost)
			return
		}
		MethodNotAllowed(w, r)

	case "dislike":
		if r.Method == http.MethodPost {
			p.handleReaction(w, r, postID, -1, ReactionTargetPost)
			return
		}
		MethodNotAllowed(w, r)

	case "comments":
		switch r.Method {
		case http.MethodGet:
			p.listComments(w, r, postID)
		case http.MethodPost:
			p.createComment(w, r, postID)
		default:
			MethodNotAllowed(w, r)
		}

	default:
		notFound(w, r)
	}
}

// ============================================================
// POST ACTIONS
// ============================================================

func (p *PostsHandler) getPost(w http.ResponseWriter, r *http.Request, postID int64) {
	post, err := repository.GetPost(r.Context(), p.conn, postID)
	if err != nil {
		log.Printf("failed to load post: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r)
			return
		}
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error loading post", http.StatusInternalServerError))
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
		WriteError(w, r, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	if req.Title == nil && req.Body == nil {
		WriteError(w, r, NewError("BAD_REQUEST", "nothing to update", http.StatusBadRequest))
		return
	}

	if err := repository.UpdatePost(
		r.Context(),
		p.conn,
		postID,
		repository.UpdatePostInput{Title: req.Title, Body: req.Body},
	); err != nil {
		log.Printf("failed to update post: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error updating post", http.StatusInternalServerError))
		return
	}

	WriteOK(w, map[string]string{"status": "updated"}, nil)
}

func (p *PostsHandler) deletePost(w http.ResponseWriter, r *http.Request, postID int64) {
	if _, ok := requireUserID(w, r); !ok {
		return
	}

	if err := repository.DeletePost(r.Context(), p.conn, postID); err != nil {
		log.Printf("failed to delete post: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error deleting post", http.StatusInternalServerError))
		return
	}

	WriteNoContent(w)
}

// ============================================================
// REACTIONS
// ============================================================

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
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error toggling reaction", http.StatusInternalServerError))
		return
	}

	var likesCount, dislikesCount int

	switch targetType {
	case ReactionTargetPost:
		likesCount, dislikesCount, err = repository.CountReactionsForPost(r.Context(), p.conn, objectID)
	case ReactionTargetComment:
		likesCount, dislikesCount, err = repository.CountReactionsForComment(r.Context(), p.conn, objectID)
	default:
		WriteError(w, r, NewError("BAD_REQUEST", "invalid target type", http.StatusBadRequest))
		return
	}

	if err != nil {
		log.Printf("failed to count reactions: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error counting reactions", http.StatusInternalServerError))
		return
	}

	idKey := "post_id"
	if targetType == ReactionTargetComment {
		idKey = "comment_id"
	}

	WriteOK(w, map[string]any{
		idKey:            objectID,
		"reaction":       reaction,
		"likes_count":    likesCount,
		"dislikes_count": dislikesCount,
	}, nil)
}

// ============================================================
// COMMENTS
// ============================================================

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
		log.Printf("failed to list comments: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error listing comments", http.StatusInternalServerError))
		return
	}

	pagination := buildPaginationInfo(page, perPage, listResult.Total, map[string]any{
		"post_id": postID,
	})

	WriteOK(w, listResult.Comments, pagination)
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
		WriteError(w, r, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	if strings.TrimSpace(req.Body) == "" {
		WriteError(w, r, NewError("BAD_REQUEST", "body required", http.StatusBadRequest))
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
		log.Printf("failed to create comment: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error creating comment", http.StatusInternalServerError))
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

// GET /api/v1/posts/{id}/comments
func (p *PostsHandler) ListPostComments(w http.ResponseWriter, r *http.Request) {
	postID, ok := getPostIDFromURL(w, r)
	if !ok {
		return
	}
	p.listComments(w, r, postID)
}

// POST /api/v1/posts/{id}/comments
func (p *PostsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	postID, ok := getPostIDFromURL(w, r)
	if !ok {
		return
	}
	p.createComment(w, r, postID)
}

// ============================================================
// MY POSTS
// ============================================================

func (p *PostsHandler) ListMyPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	page, perPage := sanitizePagination(r)

	result, err := repository.ListPostsByAuthor(
		r.Context(),
		p.conn,
		repository.ListPostsByAuthorParams{
			AuthorID: userID,
			Page:     page,
			PerPage:  perPage,
		},
	)
	if err != nil {
		log.Printf("failed to list posts by author: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error listing user posts", http.StatusInternalServerError))
		return
	}

	pagination := buildPaginationInfo(page, perPage, result.Total, map[string]any{
		"author_id": userID,
	})

	WriteOK(w, result.Posts, pagination)
}

func (p *PostsHandler) ListLikedPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	page, perPage := sanitizePagination(r)

	result, err := repository.ListPostsLikedByUser(
		r.Context(),
		p.conn,
		repository.ListPostsLikedByUserParams{
			UserID:  userID,
			Page:    page,
			PerPage: perPage,
		},
	)
	if err != nil {
		log.Printf("failed to list liked posts: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error listing liked posts", http.StatusInternalServerError))
		return
	}

	pagination := buildPaginationInfo(page, perPage, result.Total, map[string]any{
		"user_id": userID,
	})

	WriteOK(w, result.Posts, pagination)
}
