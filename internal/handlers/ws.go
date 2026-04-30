// internal/handlers/ws.go
package handlers

import (
	"database/sql"
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
	db  *sql.DB
	hub *ws.Hub
}

func NewWsHandler(database *sql.DB, hub *ws.Hub) *WsHandler {
	return &WsHandler{
		db:  database,
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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client, firstConnection := h.hub.Add(session.UserID, conn)

	// Enqueue presence messages before starting goroutines. The Send channel is
	// buffered so this is safe. Doing it here — rather than after launching
	// readPump — eliminates a race where readPump's defer (Remove + offline
	// broadcast) could fire before the online broadcast is enqueued, leaving
	// other clients with a stale "online" event for a user who has already gone.
	//
	// Every new connection receives a snapshot so it knows who is online, even
	// when the same user opens a second tab. The presence.update broadcast only
	// fires on an offline→online state transition (first connection).
	h.hub.SendSnapshotToClient(client)
	if firstConnection {
		h.hub.BroadcastPresenceUpdate(session.UserID, true)
	}

	go h.readPump(session.UserID, client)
	go h.writePump(client)
}

func (h *WsHandler) readPump(userID int64, c *ws.Client) {
	defer func() {
		remaining := h.hub.Remove(userID, c)
		if remaining == 0 {
			h.hub.BroadcastPresenceUpdate(userID, false)
		}
		close(c.Send)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(4096)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			break
		}
	}
}

// writePump is the sole goroutine writing to the connection, preventing
// concurrent-write data races. It delivers messages from the client's send
// channel and sends periodic pings to keep the connection alive.
func (h *WsHandler) writePump(c *ws.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
