package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	repository "forum/internal/db"
	"net/http"
	"strings"
)

// ============================================================
// HandleComment: /api/v1/comment/{id}
// ============================================================

func (p *PostsHandler) HandleComment(w http.ResponseWriter, r *http.Request) {
	commentID, action, ok := resolveCommentRoute(w, r)
	if !ok {
		return
	}

	switch action {

	case "":
		switch r.Method {
		case http.MethodGet:
			p.getComment(w, r, commentID)
		case http.MethodPatch:
			p.updateComment(w, r, commentID)
		case http.MethodDelete:
			p.deleteComment(w, r, commentID)
		default:
			MethodNotAllowed(w)
		}
	case "like":
		if r.Method == http.MethodPost {
			p.handleReaction(w, r, commentID, 1, "comment")
			return
		}
		MethodNotAllowed(w)

	case "dislike":
		if r.Method == http.MethodPost {
			p.handleReaction(w, r, commentID, -1, "comment")
			return
		}
		MethodNotAllowed(w)

	default:
		notFound(w)
	}
}

func (p *PostsHandler) getComment(w http.ResponseWriter, r *http.Request, commentID int64) {
	comment, err := repository.GetComment(r.Context(), p.conn, commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error loading comment", http.StatusInternalServerError))
		return
	}
	WriteOK(w, comment, nil)
}

func (p *PostsHandler) updateComment(w http.ResponseWriter, r *http.Request, commentID int64) {
	userID, ok := requireUserID(w, r)

	if !ok {
		return
	}

	//fetching comment to check ownership
	comment, err := repository.GetComment(r.Context(), p.conn, commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error loading comment", http.StatusInternalServerError))
		return
	}

	if comment.UserID != userID {
		WriteError(w, NewError("FORBIDDEN", "you are not allowed to update this comment", http.StatusForbidden))
		return
	}

	var req struct {
		Body *string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	if req.Body == nil || strings.TrimSpace(*req.Body) == "" {
		WriteError(w, NewError("BAD_REQUEST", "body required and cannot be empty", http.StatusBadRequest))
		return
	}

	if err := repository.UpdateComment(
		r.Context(),
		p.conn,
		commentID,
		repository.UpdateCommentInput{Body: req.Body},
	); err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error updating comment", http.StatusInternalServerError))
		return
	}

	updatedComment, _ := repository.GetComment(r.Context(), p.conn, commentID)
	WriteOK(w, updatedComment, nil)
}

func (p *PostsHandler) deleteComment(w http.ResponseWriter, r *http.Request, commentID int64) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	// Fetching comment to check ownership
	comment, err := repository.GetComment(r.Context(), p.conn, commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error loading comment", http.StatusInternalServerError))
		return
	}
	if comment.UserID != userID {
		WriteError(w, NewError("FORBIDDEN", "you are not allowed to delete this comment", http.StatusForbidden))
		return
	}

	if err := repository.DeleteComment(r.Context(), p.conn, commentID); err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "error deleting comment", http.StatusInternalServerError))
		return
	}
	WriteNoContent(w)
}
