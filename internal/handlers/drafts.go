// internal/handlers/drafts.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	repository "forum/internal/db"
)

func (p *PostsHandler) HandleDraft(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	switch r.Method {

	case http.MethodGet:
		post, err := repository.DraftGet(r.Context(), p.conn, userID)
		if err == sql.ErrNoRows {
			WriteOK(w, nil, nil)
			return
		}
		if err != nil {
			WriteError(w, r, NewError("INTERNAL", "load draft failed", 500))
			return
		}
		WriteOK(w, post, nil)

	case http.MethodPost:
		var req struct {
			Title       string  `json:"title"`
			ImageURL    *string `json:"image_url"`
			Body        string  `json:"body"`
			CategoryIDs []int64 `json:"category_ids"`
			Manual      bool    `json:"manual"`
		}

		var (
			uploadFile multipart.File
			uploadMime string
			uploadPath string
		)

		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		if strings.HasPrefix(contentType, "multipart/form-data") {
			const maxUploadSize = 5 << 20
			r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				var maxErr *http.MaxBytesError
				if errors.As(err, &maxErr) {
					WriteError(w, r, NewError("PAYLOAD_TOO_LARGE", "upload too large", http.StatusRequestEntityTooLarge))
					return
				}
				WriteError(w, r, NewError("BAD_REQUEST", "invalid multipart form", http.StatusBadRequest))
				return
			}
			if r.MultipartForm != nil {
				defer r.MultipartForm.RemoveAll()
			}

			req.Title = r.FormValue("title")
			req.Body = r.FormValue("body")
			if raw := strings.TrimSpace(r.FormValue("image_url")); raw != "" {
				req.ImageURL = &raw
			}
			if raw := strings.TrimSpace(r.FormValue("manual")); raw != "" {
				manual, err := strconv.ParseBool(raw)
				if err != nil {
					WriteError(w, r, NewError("BAD_REQUEST", "invalid manual flag", http.StatusBadRequest))
					return
				}
				req.Manual = manual
			}

			categoryIDs, err := parseCategoryIDs(r.Form["category_ids"])
			if err != nil {
				WriteError(w, r, NewError("BAD_REQUEST", "invalid category_id", http.StatusBadRequest))
				return
			}
			req.CategoryIDs = categoryIDs

			file, _, err := r.FormFile("image")
			if err == nil {
				uploadFile = file
				defer uploadFile.Close()
				mime, err := validateImageType(uploadFile)
				if err != nil {
					WriteError(w, r, NewError("BAD_REQUEST", "unsupported image type", http.StatusBadRequest))
					return
				}
				uploadMime = mime
			} else if err != http.ErrMissingFile {
				WriteError(w, r, NewError("BAD_REQUEST", "invalid image upload", http.StatusBadRequest))
				return
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				WriteError(w, r, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
				return
			}
		}
		if req.ImageURL != nil && strings.TrimSpace(*req.ImageURL) == "" {
			req.ImageURL = nil
		}

		// Validation ONLY for manual "Save Draft"
		if req.Manual {
			if strings.TrimSpace(req.Title) == "" {
				WriteError(w, r, NewError("BAD_REQUEST", "title required", http.StatusBadRequest))
				return
			}
			hasImageUpload := r.MultipartForm != nil && len(r.MultipartForm.File["image"]) > 0
			if strings.TrimSpace(req.Body) == "" && !hasImageUpload && req.ImageURL == nil {
				WriteError(w, r, NewError("BAD_REQUEST", "body required", http.StatusBadRequest))
				return
			}
			if len(req.CategoryIDs) == 0 {
				WriteError(w, r, NewError("BAD_REQUEST", "at least one category required", http.StatusBadRequest))
				return
			}
		}

		if uploadFile != nil {
			imageURL, imagePath, err := saveUploadedImage(uploadFile, uploadMime)
			if err != nil {
				log.Printf("failed to save image: %v", err)
				WriteError(w, r, NewError(
					"INTERNAL_SERVER_ERROR",
					"error saving image",
					http.StatusInternalServerError,
				))
				return
			}
			req.ImageURL = &imageURL
			uploadPath = imagePath
		}

		id, err := repository.DraftCreate(
			r.Context(),
			p.conn,
			userID,
			req.Title,
			req.Body,
			req.CategoryIDs,
			req.ImageURL,
		)
		if err != nil {
			if uploadPath != "" {
				_ = os.Remove(uploadPath)
			}
			log.Printf("draft save failed: %v", err)
			WriteError(w, r, NewError(
				"INTERNAL_SERVER_ERROR",
				"save failed",
				http.StatusInternalServerError,
			))
			return
		}

		WriteOK(w, map[string]int64{"id": id}, nil)

	default:
		MethodNotAllowed(w, r)
	}
}

func (p *PostsHandler) HandleDraftByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		WriteError(w, r, NewError("BAD_REQUEST", "missing draft id", 400))
		return
	}
	idStr := parts[len(parts)-1]

	draftID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || draftID <= 0 {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid draft id", 400))
		return
	}

	switch r.Method {

	case http.MethodPut:
		var req struct {
			Title       string  `json:"title"`
			ImageURL    *string `json:"image_url"`
			Body        string  `json:"body"`
			CategoryIDs []int64 `json:"category_ids"`
			Manual      bool    `json:"manual"`
		}

		var (
			uploadFile multipart.File
			uploadMime string
			uploadPath string
		)

		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		if strings.HasPrefix(contentType, "multipart/form-data") {
			const maxUploadSize = 5 << 20
			r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				var maxErr *http.MaxBytesError
				if errors.As(err, &maxErr) {
					WriteError(w, r, NewError("PAYLOAD_TOO_LARGE", "upload too large", http.StatusRequestEntityTooLarge))
					return
				}
				WriteError(w, r, NewError("BAD_REQUEST", "invalid multipart form", http.StatusBadRequest))
				return
			}
			if r.MultipartForm != nil {
				defer r.MultipartForm.RemoveAll()
			}

			req.Title = r.FormValue("title")
			req.Body = r.FormValue("body")
			if raw := strings.TrimSpace(r.FormValue("image_url")); raw != "" {
				req.ImageURL = &raw
			}
			if raw := strings.TrimSpace(r.FormValue("manual")); raw != "" {
				manual, err := strconv.ParseBool(raw)
				if err != nil {
					WriteError(w, r, NewError("BAD_REQUEST", "invalid manual flag", http.StatusBadRequest))
					return
				}
				req.Manual = manual
			}

			categoryIDs, err := parseCategoryIDs(r.Form["category_ids"])
			if err != nil {
				WriteError(w, r, NewError("BAD_REQUEST", "invalid category_id", http.StatusBadRequest))
				return
			}
			req.CategoryIDs = categoryIDs

			file, _, err := r.FormFile("image")
			if err == nil {
				uploadFile = file
				defer uploadFile.Close()
				mime, err := validateImageType(uploadFile)
				if err != nil {
					WriteError(w, r, NewError("BAD_REQUEST", "unsupported image type", http.StatusBadRequest))
					return
				}
				uploadMime = mime
			} else if err != http.ErrMissingFile {
				WriteError(w, r, NewError("BAD_REQUEST", "invalid image upload", http.StatusBadRequest))
				return
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				WriteError(w, r, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
				return
			}
		}
		if req.ImageURL != nil && strings.TrimSpace(*req.ImageURL) == "" {
			req.ImageURL = nil
		}

		if uploadFile != nil {
			imageURL, imagePath, err := saveUploadedImage(uploadFile, uploadMime)
			if err != nil {
				log.Printf("failed to save image: %v", err)
				WriteError(w, r, NewError(
					"INTERNAL_SERVER_ERROR",
					"error saving image",
					http.StatusInternalServerError,
				))
				return
			}
			req.ImageURL = &imageURL
			uploadPath = imagePath
		}

		// Validation ONLY for manual update
		if req.Manual {
			if strings.TrimSpace(req.Title) == "" {
				WriteError(w, r, NewError("BAD_REQUEST", "title required", http.StatusBadRequest))
				return
			}
			hasImageUpload := r.MultipartForm != nil && len(r.MultipartForm.File["image"]) > 0
			hasImageURL := req.ImageURL != nil && strings.TrimSpace(*req.ImageURL) != ""
			if strings.TrimSpace(req.Body) == "" && !hasImageUpload && !hasImageURL {
				WriteError(w, r, NewError("BAD_REQUEST", "body required", http.StatusBadRequest))
				return
			}
			if len(req.CategoryIDs) == 0 {
				WriteError(w, r, NewError("BAD_REQUEST", "at least one category required", http.StatusBadRequest))
				return
			}
		}

		err = repository.DraftUpdate(
			r.Context(),
			p.conn,
			userID,
			draftID,
			req.Title,
			req.Body,
			req.CategoryIDs,
			req.ImageURL,
		)
		if err == sql.ErrNoRows {
			WriteError(w, r, NewError("NOT_FOUND", "draft not found", http.StatusNotFound))
			return
		}
		if err != nil {
			if uploadPath != "" {
				_ = os.Remove(uploadPath)
			}
			WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "update failed", http.StatusInternalServerError))
			return
		}

		WriteNoContent(w)
		return

	case http.MethodDelete:
		err = repository.DraftDelete(r.Context(), p.conn, userID, draftID)
		if err == sql.ErrNoRows {
			WriteError(w, r, NewError("NOT_FOUND", "draft not found", http.StatusNotFound))
			return
		}
		if err != nil {
			WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "delete failed", http.StatusInternalServerError))
			return
		}

		WriteNoContent(w)
		return

	default:
		MethodNotAllowed(w, r)
	}
}
