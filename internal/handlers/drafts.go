//internal/handlers/drafts.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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
			Title string `json:"title"`
			Body  string `json:"body"`
			CategoryIDs []int64 `json:"category_ids"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, NewError("BAD_REQUEST", "invalid json", 400))
			return
		}

		if strings.TrimSpace(req.Title) == "" {
			WriteError(w, r, NewError("BAD_REQUEST", "title required", 400))
			return
		}
		if strings.TrimSpace(req.Body) == "" {
			WriteError(w, r, NewError("BAD_REQUEST", "body required", 400))
			return
		}

		if len(req.CategoryIDs) == 0 {
			WriteError(w, r, NewError("BAD_REQUEST", "At least one category required", 400))
			return
		}

		id, err := repository.DraftCreate(
			r.Context(),
			p.conn,
			userID,
			req.Title,
			req.Body,
			req.CategoryIDs,
		)
		if err != nil {
			log.Printf("draft save failed: %v", err)
			WriteError(w, r, NewError("INTERNAL", "save failed", 500))
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

	// extract id from URL: /posts/draft/{id}
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

	case http.MethodPut:
		var req struct {
			Title       string  `json:"title"`
			Body        string  `json:"body"`
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
		if strings.TrimSpace(req.Body) == "" {
			WriteError(w, r, NewError("BAD_REQUEST", "body required", http.StatusBadRequest))
			return
		}

		err = repository.DraftUpdate(r.Context(), p.conn, userID, draftID, req.Title, req.Body, req.CategoryIDs)
		if err == sql.ErrNoRows {
			WriteError(w, r, NewError("NOT_FOUND", "draft not found", http.StatusNotFound))
			return
		}
		if err != nil {
			WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "update failed", http.StatusInternalServerError))
			return
		}

		WriteNoContent(w)
		return
	default:
		MethodNotAllowed(w, r)
		return
	}
}
