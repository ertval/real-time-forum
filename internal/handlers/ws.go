// internal/handlers/ws.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"forum/internal/db"
	"forum/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsHandler struct {
	db *sql.DB
	hub *ws.Hub
}

func NewWsHandler(database *sql.DB, hub *ws.Hub) *WsHandler {
	return &WsHandler{
		db: database,
		hub: hub,
	}
}

func (h *WsHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := ""

	if cookie, err := r.Cookie("session_token"); err == nil {
		token = cookie.Value
	}

	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	session, err := db.GetSessionByToken(r.Context(), h.db, token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID := session.UserID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	firstConnection := h.hub.Add(userID, conn)
	if firstConnection {
		h.hub.BroadcastPresenceUpdate(userID, true)
	}
	h.hub.SendPresenceSnapshot(userID)

	go h.readPump(userID, conn)
	go h.writePump(userID, conn)
}

func (h *WsHandler) readPump(userID int64, conn *websocket.Conn) {
	defer func() {
		remaining := h.hub.Remove(userID, conn)
		if remaining == 0 {
			h.hub.BroadcastPresenceUpdate(userID, false)
		}
		conn.Close()
	}()

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ws.WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			break
		}

		h.handleMessage(userID, msg)
	}
}

func (h *WsHandler) writePump(userID int64, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WsHandler) handleMessage(userID int64, msg ws.WSMessage) {
	switch msg.Type {
	case "dm.send":
		var payload ws.DMSendPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.sendError(userID, "INVALID_PAYLOAD", "invalid message payload")
			return
		}

		if payload.RecipientID == userID {
			h.sendError(userID, "SELF_SEND", "cannot send message to self")
			return
		}

		if payload.Body == "" {
			h.sendError(userID, "EMPTY_BODY", "message body cannot be empty")
			return
		}

		if !h.hub.IsUserOnline(payload.RecipientID) {
			h.sendError(userID, "RECIPIENT_OFFLINE", "recipient is offline")
			return
		}

		// TODO: C06 will handle persistence and actual routing
		_ = userID
	}
}

func (h *WsHandler) sendError(userID int64, code, message string) {
	payload, _ := json.Marshal(map[string]string{"code": code, "message": message})
	errMsg := ws.WSMessage{
		Type:    "chat.error",
		Payload: payload,
	}
	data, _ := json.Marshal(errMsg)
	h.hub.SendToUser(userID, data)
}