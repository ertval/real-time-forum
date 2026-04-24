# C01: Authenticated WebSocket Endpoint and Connection Manager
<!-- Filename: docs/pr-message/C01-WebSocket-Endpoint-pr.md -->

Implements the WebSocket transport layer that all real-time features depend on. Adds `GET /ws` with session-cookie authentication, a thread-safe connection Hub, and multi-tab connection lifecycle tracking.

This is a resubmission addressing the four findings from the first review (2× P1, 2× P2).

## Summary of Changes

### 1. Connection Manager (`internal/ws/connection_manager.go`)
- **`Client` struct**: wraps a `*websocket.Conn` with a buffered `send chan []byte` so all writes are serialized through a single goroutine — eliminates the concurrent-write race flagged in review
- **Stale-map fix**: `Hub.Remove` now calls `delete(h.connections, userID)` when the last connection closes; `GetOnlineUserIDs` and `SendPresenceSnapshot` can no longer iterate over empty user entries from disconnected users
- **Channel-based send**: `BroadcastPresenceUpdate`, `SendToUser`, and `SendPresenceSnapshot` enqueue via `client.Send` rather than calling `conn.WriteMessage` directly

### 2. WebSocket Handler (`internal/handlers/ws.go`)
- **Auth**: validates `session_token` cookie before upgrade; returns `401` if missing or invalid
- **Lifecycle**: `readPump` reads until EOF then calls `Hub.Remove`; `writePump` is the sole writer — drains `client.Send` channel and sends periodic pings
- **Scope**: removed out-of-scope C04 presence broadcast calls and C06 `dm.send` handling that were present in the first submission

### 3. Backend Integration
- **Router**: `internal/router/router.go` accepts `*ws.Hub` and registers `/ws`
- **Server**: `cmd/backend/server.go` creates the Hub and passes it to the router

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C01**:
> - authenticated users can open WebSocket connections
> - unauthenticated connections are rejected
> - multi-tab connection lifecycle is tracked reliably

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./internal/... -timeout 60s` — All packages pass
- [x] `go test ./internal/tests/... -race -run "TestWebSocket"` — All 5 WS tests pass, no data races detected
- [x] `go build ./...` — Builds without errors

### QA Checklist
- [x] `TestWebSocket_Unauthenticated` — no cookie → `401`
- [x] `TestWebSocket_InvalidSession` — bad token → `401`
- [x] `TestWebSocket_ValidConnection` — valid session → `101 Switching Protocols`
- [x] `TestWebSocket_MultiTab` — same user dials twice → both connections alive
- [x] `TestWebSocket_DisconnectCleansUp` — close frame → server removes client cleanly

### Manual E2E Verification (websocat)
- [x] No cookie → `401 Unauthorized`
- [x] Invalid session token → `401 Unauthorized`
- [x] Valid session → WebSocket upgrades, connection held open for 2s without error
- [x] Multi-tab → two simultaneous `websocat` processes with the same token both stay alive

## Key Files Impacted
- `internal/ws/connection_manager.go` (new)
- `internal/handlers/ws.go` (new)
- `internal/tests/ws_test.go` (new)
- `internal/router/router.go` (modified)
- `cmd/backend/server.go` (modified)
- `go.mod` / `go.sum` (gorilla/websocket dependency)
