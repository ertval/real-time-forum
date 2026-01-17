// internal/handlers/drafts.go

package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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
		post, err := repository.GetMyDraft(r.Context(), p.conn, userID)
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
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, NewError("BAD_REQUEST", "invalid json", 400))
			return
		}

		if strings.TrimSpace(req.Title) == "" &&
			strings.TrimSpace(req.Body) == "" {
			WriteError(w, r, NewError("BAD_REQUEST", "empty draft", 400))
			return
		}

		id, err := repository.UpsertDraft(
			r.Context(),
			p.conn,
			userID,
			req.Title,
			req.Body,
		)
		if err != nil {
			log.Printf("draft save failed: %v", err)
			WriteError(w, r, NewError("INTERNAL", "save failed", 500))
			return
		}

		WriteOK(w, map[string]int64{"id": id}, nil)

	case http.MethodDelete:
		err := repository.DeleteUserDraft(r.Context(), p.conn, userID)
		if err != nil {
			WriteError(w, r, NewError("INTERNAL", "delete draft failed", 500))
			return
		}
		WriteNoContent(w)

	default:
		MethodNotAllowed(w, r)
	}
}
