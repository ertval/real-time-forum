// internal/handlers/chats.go
package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/db"
	"forum/internal/middleware"
	"forum/internal/ws"
)

type ChatsHandler struct {
	conn *sql.DB
	hub  *ws.Hub
}

func NewChatsHandler(database *sql.DB, hub *ws.Hub) *ChatsHandler {
	return &ChatsHandler{conn: database, hub: hub}
}

/*--------------------------
  HANDLE CHAT ROSTER
--------------------------*/

// rosterResponseEntry is the JSON shape for a single roster row.
// Last-message fields are pointers so they can be encoded as `null` for users
// without history, matching SDS § 5.3.
type rosterResponseEntry struct {
	UserID             int64   `json:"user_id"`
	Username           string  `json:"username"`
	IsOnline           bool    `json:"is_online"`
	LastMessageAt      *string `json:"last_message_at"`
	LastMessagePreview *string `json:"last_message_preview"`
	LastSenderID       *int64  `json:"last_sender_id"`
}

func (h *ChatsHandler) HandleChatRoster(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil || userID == 0 {
		WriteError(w, r, NewError("UNAUTHORIZED", "user not authenticated", http.StatusUnauthorized))
		return
	}

	entries, err := db.GetChatRoster(r.Context(), h.conn, userID)
	if err != nil {
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "failed to load roster", http.StatusInternalServerError))
		return
	}

	roster := make([]rosterResponseEntry, len(entries))
	for i, e := range entries {
		row := rosterResponseEntry{
			UserID:   e.UserID,
			Username: e.Username,
			IsOnline: h.hub.IsUserOnline(e.UserID),
		}
		if e.LastMessageAt != "" {
			at := e.LastMessageAt
			preview := e.LastMessagePreview
			sender := e.LastSenderID
			row.LastMessageAt = &at
			row.LastMessagePreview = &preview
			row.LastSenderID = &sender
		}
		roster[i] = row
	}

	WriteOK(w, roster, nil)
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

	messages, hasMore, err := db.GetMessageHistory(r.Context(), h.conn, userID, recipientID, beforeID)
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
			WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "failed to retrieve user info", http.StatusInternalServerError))
			return
		}
		for _, user := range users {
			usersMap[user.ID] = user.Username
		}
		// Graceful fallback for orphaned senders (user deleted after sending).
		for senderID := range uniqueSenders {
			if _, found := usersMap[senderID]; !found {
				usersMap[senderID] = "unknown"
			}
		}
	}

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
