package ws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu sync.RWMutex

	connections map[int64]map[*websocket.Conn]bool

	onConnect    func(userID int64)
	onDisconnect func(userID int64)
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[int64]map[*websocket.Conn]bool),
	}
}

func (h *Hub) SetCallbacks(onConnect, onDisconnect func(int64)) {
	h.onConnect = onConnect
	h.onDisconnect = onDisconnect
}

func (h *Hub) Add(userID int64, conn *websocket.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[userID] == nil {
		h.connections[userID] = make(map[*websocket.Conn]bool)
	}

	firstConnection := len(h.connections[userID]) == 0

	h.connections[userID][conn] = true

	if firstConnection && h.onConnect != nil {
		h.onConnect(userID)
	}

	return firstConnection
}

func (h *Hub) Remove(userID int64, conn *websocket.Conn) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[userID] == nil {
		return 0
	}

	delete(h.connections[userID], conn)

	remaining := len(h.connections[userID])

	if remaining == 0 && h.onDisconnect != nil {
		h.onDisconnect(userID)
	}

	return remaining
}

func (h *Hub) GetOnlineUserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var userIDs []int64
	for userID := range h.connections {
		if len(h.connections[userID]) > 0 {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs
}

func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[userID]) > 0
}

func (h *Hub) BroadcastPresenceUpdate(userID int64, isOnline bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := WSMessage{
		Type: "presence.update",
		Payload: json.RawMessage(`{"user_id":` + int64ToJSON(userID) + `,"is_online":` + boolToJSON(isOnline) + `}`),
	}
	data, _ := json.Marshal(msg)

	for userID := range h.connections {
		for conn := range h.connections[userID] {
			conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (h *Hub) SendToUser(userID int64, data []byte) error {
	h.mu.RLock()
	conns := h.connections[userID]
	h.mu.RUnlock()

	if len(conns) == 0 {
		return ErrUserOffline
	}

	for conn := range conns {
		conn.WriteMessage(websocket.TextMessage, data)
	}
	return nil
}

func (h *Hub) SendPresenceSnapshot(userID int64) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]json.RawMessage, 0, len(h.connections))
	for uid := range h.connections {
		users = append(users, json.RawMessage(`{"user_id":`+int64ToJSON(uid)+`,"is_online":true}`))
	}

	payload, _ := json.Marshal(map[string]any{"users": users})
	msg := WSMessage{
		Type:    "presence.snapshot",
		Payload: payload,
	}
	data, _ := json.Marshal(msg)

	for conn := range h.connections[userID] {
		conn.WriteMessage(websocket.TextMessage, data)
	}
	return nil
}

func (h *Hub) GetConnectionCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[userID])
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type DMSendPayload struct {
	RecipientID int64  `json:"recipient_id"`
	Body      string `json:"body"`
}

var (
	ErrUserOffline   = fmt.Errorf("user offline")
	ErrInvalidMsg = fmt.Errorf("invalid message")
)

func int64ToJSON(v int64) string {
	return strconv.FormatInt(v, 10)
}

func boolToJSON(v bool) string {
	if v {
		return "true"
	}
	return "false"
}