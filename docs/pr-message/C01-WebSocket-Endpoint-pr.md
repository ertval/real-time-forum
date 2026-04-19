# C01: Authenticated WebSocket Endpoint and Connection Manager
<!-- Filename: docs/pr-message/C01-WebSocket-Endpoint-pr.md -->

Implements the WebSocket transport layer for real-time chat functionality. Adds `GET /ws` endpoint with session-cookie authentication, presence tracking, and multi-tab connection support.

## Summary of Changes

### 1. WebSocket Infrastructure
- **Gorilla WebSocket**: Added `github.com/gorilla/websocket` dependency to `go.mod`
- **Connection Manager**: Created `internal/ws/connection_manager.go` with thread-safe Hub struct

### 2. WebSocket Handler
- **HandleWebSocket**: New handler in `internal/handlers/ws.go` that:
  - Validates session cookies during upgrade
  - Rejects unauthenticated connections with 401
  - Manages connection lifecycle per user
  - Supports multiple connections (tabs) via connection counting

### 3. Presence System
- **On Connect**: Sends `presence.snapshot` to new client, broadcasts `presence.update` if first connection
- **On Disconnect**: Removes connection, broadcasts `presence.update` if last connection
- **Message Routing**: Basic `dm.send` parsing with validation (self-send, empty body, offline recipient)

### 4. Backend Integration
- **Router Update**: Modified `internal/router/router.go` to accept Hub and register `/ws` route
- **Server Init**: Modified `cmd/backend/server.go` to create Hub and pass to router

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C01**:
> - authenticated users can open WebSocket connections
> - unauthenticated connections are rejected with 401
> - multi-tab connection lifecycle is tracked reliably (connection count)

## Testing & Validation Verified

### Automated Test Suite
- [x] `make test` — All tests pass (Go backend + Vitest frontend)
- [x] `bun run lint` — Biome check: 82 files, no issues
- [x] `go build ./...` — Project builds without errors

### QA Checklist
- [x] Verified unauthenticated WS upgrade returns 401
- [x] Verified invalid session token returns 401
- [x] Verified Hub tracks connection count correctly
- [x] Verified router signature change compiles with all existing tests

### Manual E2E Verification
- [ ] **Not yet tested**: Real browser WS connection with multiple tabs (requires running servers)

## Key Files Impacted
- `internal/ws/connection_manager.go` (new)
- `internal/handlers/ws.go` (new)
- `internal/router/router.go` (modified)
- `cmd/backend/server.go` (modified)
- `go.mod` (modified)
- `internal/tests/ws_test.go` (new)
- `internal/tests/users_test.go` (modified - router signature)
- `internal/tests/helpers_test.go` (modified - router signature)
- `internal/tests/forum_auth_access_test.go` (modified - 401 vs 404)