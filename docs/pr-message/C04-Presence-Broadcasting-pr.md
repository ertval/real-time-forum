# C04: Presence Broadcasting
<!-- Filename: docs/pr-message/C04-Presence-Broadcasting-pr.md -->

Extends the WebSocket layer from C01 with presence tracking: every connected client receives a snapshot of who is online on connect, and a targeted broadcast whenever any user's online state transitions.

## Summary of Changes

### 1. Connection Manager (`internal/ws/connection_manager.go`)
- **`Hub.Add` first-connection flag**: returns a `bool` indicating whether this is the user's first connection, enabling transition-only online broadcasts
- **`SendSnapshotToClient(*Client)`**: sends `presence.snapshot` to a specific client — not to all tabs of the same user — so second-tab connects don't flood existing tabs with duplicate snapshots
- **`BroadcastPresenceUpdate`**: enqueues `presence.update` to every connected client via buffered send channels; never writes to connections directly
- **`SendToUser` race fix**: holds the read lock for the full iteration so a concurrent `Remove` cannot mutate the clients map while it is being ranged
- **`SetCallbacks` deadlock warning**: documented that callbacks fire under the write lock and must not call any Hub broadcast method

### 2. WebSocket Handler (`internal/handlers/ws.go`)
- **Presence sequencing**: snapshot and online broadcast are enqueued *before* `readPump`/`writePump` goroutines start, eliminating a race where a fast disconnect could broadcast offline before the online event was queued
- **Per-connection snapshot**: `SendSnapshotToClient` is called unconditionally for every new connection (not just the first), so second-tab users also receive the current online roster
- **Transition-only online broadcast**: `BroadcastPresenceUpdate(userID, true)` fires only when `firstConnection` is true (offline→online transition)
- **Last-disconnect offline broadcast**: `readPump` defer calls `Hub.Remove`; only broadcasts offline when `remaining == 0`
- **Read limit**: raised from 512 → 4096 bytes to accommodate future C06 `dm.send` payloads

### 3. Tests (`internal/tests/presence_test.go`)
- 12 integration tests covering: snapshot delivery on connect, online/offline state-transition broadcasts, correct user IDs in every event, second-tab snapshot delivery without spurious online broadcast, multi-observer fanout, and full reconnect cycles

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C04**:
> - new clients receive a presence snapshot
> - first connection marks a user online
> - last disconnect marks a user offline

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./... -timeout 60s` — All packages pass
- [x] `go test ./internal/tests/ -run TestPresence -race -count=3 -timeout 120s` — All 12 presence tests pass with race detector, 3 runs

### QA Checklist
- [x] `TestPresence_SnapshotDeliveredOnConnect` — new client receives `presence.snapshot`
- [x] `TestPresence_FirstConnectBroadcastsOnline` — first connection emits `presence.update {is_online: true}`
- [x] `TestPresence_SecondTabReceivesSnapshot` — second tab for the same user receives a snapshot
- [x] `TestPresence_SecondTabDoesNotBroadcast` — second tab does not emit an online broadcast
- [x] `TestPresence_LastDisconnectBroadcastsOffline` — last disconnect emits `presence.update {is_online: false}`
- [x] `TestPresence_NonLastDisconnectDoesNotBroadcast` — non-last disconnect emits no broadcast
- [x] `TestPresence_SnapshotContainsCorrectUserIDs` — snapshot payload contains the expected user IDs
- [x] `TestPresence_OnlineBroadcastContainsCorrectUserID` — online update carries the connecting user's actual ID
- [x] `TestPresence_OfflineBroadcastContainsCorrectUserID` — offline update carries the disconnecting user's actual ID
- [x] `TestPresence_SnapshotExcludesOfflineUsers` — disconnected users do not appear in subsequent snapshots
- [x] `TestPresence_OfflineBroadcastReachesAllObservers` — offline event fans out to all connected observers
- [x] `TestPresence_ReconnectCycle` — observer sees online → offline → online in the correct sequence

### Audit Findings Resolved
- **Race condition (broadcast ordering)**: presence messages moved before goroutine start — eliminated the window where readPump's defer could fire before the online event was enqueued
- **`SendToUser` data race**: concurrent map read/write between iteration and `Remove`; fixed by holding the read lock for the full iteration (consistent with `BroadcastPresenceUpdate`)
- **Snapshot targeting**: replaced `SendPresenceSnapshot(userID)` (sent to all tabs) with `SendSnapshotToClient(*Client)` (sent only to the new connection)

## Key Files Impacted
- `internal/ws/connection_manager.go` (modified)
- `internal/handlers/ws.go` (modified)
- `internal/tests/presence_test.go` (new)
