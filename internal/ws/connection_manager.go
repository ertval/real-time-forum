package ws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// client wraps a websocket connection with a dedicated send channel so that
// all writes to the connection are serialized through a single goroutine.
type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
		Send: make(chan []byte, 64),
	}
}

type Hub struct {
	mu sync.RWMutex

	connections map[int64]map[*Client]bool

	onConnect    func(userID int64)
	onDisconnect func(userID int64)
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[int64]map[*Client]bool),
	}
}

func (h *Hub) SetCallbacks(onConnect, onDisconnect func(int64)) {
	h.onConnect = onConnect
	h.onDisconnect = onDisconnect
}

// Add registers a new client for the given user. Returns the client and whether
// this is the user's first connection.
func (h *Hub) Add(userID int64, conn *websocket.Conn) (*Client, bool) {
	c := NewClient(conn)

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[userID] == nil {
		h.connections[userID] = make(map[*Client]bool)
	}

	firstConnection := len(h.connections[userID]) == 0
	h.connections[userID][c] = true

	if firstConnection && h.onConnect != nil {
		h.onConnect(userID)
	}

	return c, firstConnection
}

// Remove unregisters a client. Deletes the user entry when the last connection
// closes so that GetOnlineUserIDs never returns stale disconnected users.
func (h *Hub) Remove(userID int64, c *Client) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[userID] == nil {
		return 0
	}

	delete(h.connections[userID], c)
	remaining := len(h.connections[userID])

	if remaining == 0 {
		delete(h.connections, userID)
		if h.onDisconnect != nil {
			h.onDisconnect(userID)
		}
	}

	return remaining
}

func (h *Hub) GetOnlineUserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var userIDs []int64
	for userID := range h.connections {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[userID]) > 0
}

func (h *Hub) GetConnectionCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[userID])
}

// BroadcastPresenceUpdate enqueues a presence.update message for every connected
// client via their send channels — never writes to connections directly.
func (h *Hub) BroadcastPresenceUpdate(userID int64, isOnline bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := WSMessage{
		Type: "presence.update",
		Payload: json.RawMessage(
			`{"user_id":` + int64ToJSON(userID) + `,"is_online":` + boolToJSON(isOnline) + `}`,
		),
	}
	data, _ := json.Marshal(msg)

	for _, clients := range h.connections {
		for c := range clients {
			select {
			case c.Send <- data:
			default:
				// slow client — drop rather than block
			}
		}
	}
}

// SendToUser enqueues data for every connection belonging to the given user.
func (h *Hub) SendToUser(userID int64, data []byte) error {
	h.mu.RLock()
	clients := h.connections[userID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		return ErrUserOffline
	}

	for c := range clients {
		select {
		case c.Send <- data:
		default:
		}
	}
	return nil
}

// SendPresenceSnapshot enqueues a presence.snapshot for all of the given
// user's connections.
func (h *Hub) SendPresenceSnapshot(userID int64) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]json.RawMessage, 0, len(h.connections))
	for uid := range h.connections {
		users = append(users, json.RawMessage(
			`{"user_id":`+int64ToJSON(uid)+`,"is_online":true}`,
		))
	}

	payload, _ := json.Marshal(map[string]any{"users": users})
	msg := WSMessage{
		Type:    "presence.snapshot",
		Payload: payload,
	}
	data, _ := json.Marshal(msg)

	for c := range h.connections[userID] {
		select {
		case c.Send <- data:
		default:
		}
	}
	return nil
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

var ErrUserOffline = fmt.Errorf("user offline")

func int64ToJSON(v int64) string {
	return strconv.FormatInt(v, 10)
}

func boolToJSON(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
