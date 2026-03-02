// Internal/handlers/comments.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	repository "forum/internal/db"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

/*-------------------------------------
  HandleComment: /api/v1/comment/{id}
-------------------------------------*/

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
			MethodNotAllowed(w, r)
		}
	case "like":
		if r.Method == http.MethodPost {
			p.handleReaction(w, r, commentID, 1, "comment")
			return
		}
		MethodNotAllowed(w, r)

	case "dislike":
		if r.Method == http.MethodPost {
			p.handleReaction(w, r, commentID, -1, "comment")
			return
		}
		MethodNotAllowed(w, r)

	default:
		notFound(w, r)
	}
}

func (p *PostsHandler) getComment(w http.ResponseWriter, r *http.Request, commentID int64) {
	comment, err := repository.GetCommentWithAuthor(r.Context(), p.conn, commentID)
	if err != nil {
		log.Printf("failed to load comment: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r)
			return
		}
		WriteError(
			w,
			r,
			NewError("INTERNAL_SERVER_ERROR", "error loading comment", http.StatusInternalServerError),
		)
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
	comment, err := repository.GetCommentWithAuthor(r.Context(), p.conn, commentID)
	if err != nil {
		log.Printf("failed to load comment for update: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r)
			return
		}
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error loading comment", http.StatusInternalServerError))
		return
	}

	if comment.UserID != userID {
		WriteError(w, r, NewError("FORBIDDEN", "you are not allowed to update this comment", http.StatusForbidden))
		return
	}

	updateReq := struct {
		Body           *string
		ImageURL       *string
		HasImageUpdate bool
		RemoveImage    bool

		UploadFile     multipart.File
		UploadMime     string
		UploadPath     string
		HasImageUpload bool
	}{}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		cleanupMultipartForm, ok := parseMultipartForm(w, r)
		if !ok {
			return
		}
		defer cleanupMultipartForm()

		if value, exists := multipartFirstValue(r.MultipartForm, "body"); exists {
			updateReq.Body = value
		}

		if value, exists := multipartFirstTrimmedValue(r.MultipartForm, "image_url"); exists {
			if value != nil && *value == "" {
				updateReq.ImageURL = nil
			} else {
				updateReq.ImageURL = value
			}
			updateReq.HasImageUpdate = true
		}

		removeImage, err := multipartOptionalBool(r.MultipartForm, "remove_image")
		if err != nil {
			WriteError(w, r, NewError("BAD_REQUEST", "invalid remove_image flag", http.StatusBadRequest))
			return
		}
		if removeImage != nil {
			updateReq.RemoveImage = *removeImage
		}

		file, mime, hasUpload, ok := parseImageUpload(w, r)
		if !ok {
			return
		}
		if hasUpload {
			updateReq.UploadFile = file
			defer updateReq.UploadFile.Close()
			updateReq.UploadMime = mime
			updateReq.HasImageUpload = true
		}
	} else {
		var req struct {
			Body        *string `json:"body"`
			ImageURL    *string `json:"image_url"`
			RemoveImage *bool   `json:"remove_image"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
			return
		}

		updateReq.Body = req.Body

		if req.ImageURL != nil {
			trimmed := strings.TrimSpace(*req.ImageURL)
			if trimmed == "" {
				updateReq.ImageURL = nil
			} else {
				updateReq.ImageURL = &trimmed
			}
			updateReq.HasImageUpdate = true
		}

		if req.RemoveImage != nil {
			updateReq.RemoveImage = *req.RemoveImage
		}
	}

	if updateReq.RemoveImage && updateReq.HasImageUpload {
		WriteError(w, r, NewError(
			"BAD_REQUEST",
			"remove_image cannot be combined with image upload",
			http.StatusBadRequest,
		))
		return
	}

	if updateReq.RemoveImage && updateReq.HasImageUpdate && updateReq.ImageURL != nil {
		WriteError(w, r, NewError(
			"BAD_REQUEST",
			"remove_image cannot be combined with image_url",
			http.StatusBadRequest,
		))
		return
	}

	if updateReq.UploadFile != nil {
		savedImageURL, imagePath, err := saveUploadedImage(updateReq.UploadFile, updateReq.UploadMime)
		if err != nil {
			log.Printf("failed to save updated comment image: %v", err)
			WriteError(w, r, NewError(
				"INTERNAL_SERVER_ERROR",
				"error saving image",
				http.StatusInternalServerError,
			))
			return
		}
		updateReq.ImageURL = &savedImageURL
		updateReq.HasImageUpdate = true
		updateReq.UploadPath = imagePath
	}

	if updateReq.RemoveImage && !updateReq.HasImageUpload {
		updateReq.ImageURL = nil
		updateReq.HasImageUpdate = true
	}

	if updateReq.Body == nil && !updateReq.HasImageUpdate {
		WriteError(w, r, NewError("BAD_REQUEST", "nothing to update", http.StatusBadRequest))
		return
	}

	finalBody := comment.Body
	if updateReq.Body != nil {
		finalBody = *updateReq.Body
	}

	finalImageURL := comment.ImageURL
	if updateReq.HasImageUpdate {
		finalImageURL = updateReq.ImageURL
	}

	if strings.TrimSpace(finalBody) == "" && finalImageURL == nil {
		if updateReq.UploadPath != "" {
			_ = os.Remove(updateReq.UploadPath)
		}
		WriteError(w, r, NewError("BAD_REQUEST", "body required", http.StatusBadRequest))
		return
	}

	if err := repository.UpdateComment(
		r.Context(),
		p.conn,
		commentID,
		repository.UpdateCommentInput{
			Body:           updateReq.Body,
			ImageURL:       updateReq.ImageURL,
			HasImageUpdate: updateReq.HasImageUpdate,
		},
	); err != nil {
		if updateReq.UploadPath != "" {
			_ = os.Remove(updateReq.UploadPath)
		}
		log.Printf("failed to update comment: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error updating comment", http.StatusInternalServerError))
		return
	}

	if imageURLToCleanup, shouldCleanup := replacedOrRemovedImageURL(comment.ImageURL, finalImageURL); shouldCleanup {
		if err := maybeDeleteUploadedImageByURL(r.Context(), p.conn, imageURLToCleanup); err != nil {
			log.Printf("failed to cleanup replaced comment image (comment_id=%d, image_url=%q): %v", commentID, imageURLToCleanup, err)
		}
	}

	updatedComment, err := repository.GetCommentWithAuthor(r.Context(), p.conn, commentID)
	if err != nil {
		log.Printf("failed to load updated comment: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "comment updated but failed to load", http.StatusInternalServerError))
	}
	WriteOK(w, updatedComment, nil)
}

func (p *PostsHandler) deleteComment(w http.ResponseWriter, r *http.Request, commentID int64) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	// Fetching comment to check ownership
	comment, err := repository.GetCommentWithAuthor(r.Context(), p.conn, commentID)
	if err != nil {
		log.Printf("failed to load comment for deletion: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r)
			return
		}
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error loading comment", http.StatusInternalServerError))
		return
	}
	if comment.UserID != userID {
		WriteError(w, r, NewError("FORBIDDEN", "you are not allowed to delete this comment", http.StatusForbidden))
		return
	}

	if err := repository.DeleteComment(r.Context(), p.conn, commentID); err != nil {
		log.Printf("failed to delete comment: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error deleting comment", http.StatusInternalServerError))
		return
	}

	if comment.ImageURL != nil {
		if err := maybeDeleteUploadedImageByURL(r.Context(), p.conn, *comment.ImageURL); err != nil {
			log.Printf("failed to cleanup comment image after delete (comment_id=%d, image_url=%q): %v", commentID, *comment.ImageURL, err)
		}
	}
	WriteNoContent(w)
}
