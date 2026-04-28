// internal/handlers/chats.go
package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/db"
	"forum/internal/middleware"
)

type ChatsHandler struct {
	conn *sql.DB
}

func NewChatsHandler(database *sql.DB) *ChatsHandler {
	return &ChatsHandler{conn: database}
}

/*--------------------------
  HANDLE CHAT MESSAGES
--------------------------*/

func (h *ChatsHandler) HandleChatMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getChatHistory(w, r)
	default:
		MethodNotAllowed(w, r)
	}
}

/*--------------------------
  GET CHAT HISTORY
--------------------------*/

func (h *ChatsHandler) getChatHistory(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasSuffix(path, "/messages") {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid path, expected /chats/{userID}/messages", http.StatusBadRequest))
		return
	}

	path = strings.TrimSuffix(path, "/messages")
	path = strings.TrimPrefix(path, "/api/v1/chats/")
	if path == "" {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid path", http.StatusBadRequest))
		return
	}

	recipientID, err := strconv.ParseInt(path, 10, 64)
	if err != nil || recipientID <= 0 {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid recipient user ID", http.StatusBadRequest))
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil || userID == 0 {
		WriteError(w, r, NewError("UNAUTHORIZED", "user not authenticated", http.StatusUnauthorized))
		return
	}

	beforeIDStr := r.URL.Query().Get("before_id")
	var beforeID int64
	if beforeIDStr != "" {
		beforeID, err = strconv.ParseInt(beforeIDStr, 10, 64)
		if err != nil || beforeID <= 0 {
			WriteError(w, r, NewError("BAD_REQUEST", "invalid before_id", http.StatusBadRequest))
			return
		}
	}

	messages, err := db.GetMessageHistory(r.Context(), h.conn, userID, recipientID, beforeID)
	if err != nil {
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "failed to retrieve messages", http.StatusInternalServerError))
		return
	}

	usersMap := make(map[int64]string)
	if len(messages) > 0 {
		uniqueSenders := make(map[int64]bool)
		for _, msg := range messages {
			uniqueSenders[msg.SenderID] = true
		}
		senderIDs := make([]int64, 0, len(uniqueSenders))
		for senderID := range uniqueSenders {
			senderIDs = append(senderIDs, senderID)
		}

		users, err := db.GetUsersByIDs(r.Context(), h.conn, senderIDs)
		if err != nil {
			for senderID := range uniqueSenders {
				usersMap[senderID] = "unknown"
			}
		} else {
			for _, user := range users {
				usersMap[user.ID] = user.Username
			}
		}
	}

	hasMore := len(messages) == 10

	messagesData := make([]map[string]any, len(messages))
	for i, msg := range messages {
		senderUsername := usersMap[msg.SenderID]
		messagesData[i] = map[string]any{
			"id":              msg.ID,
			"sender_id":       msg.SenderID,
			"recipient_id":    msg.RecipientID,
			"sender_username": senderUsername,
			"body":            msg.Body,
			"created_at":      msg.CreatedAt,
		}
	}

	resp := struct {
		Messages []map[string]any `json:"messages"`
		HasMore  bool             `json:"has_more"`
	}{
		Messages: messagesData,
		HasMore:  hasMore,
	}

	WriteOK(w, resp, nil)
}
