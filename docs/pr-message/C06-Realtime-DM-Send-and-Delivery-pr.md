# C06: Realtime DM Send and Delivery
<!-- Filename: docs/pr-message/C06-Realtime-DM-Send-and-Delivery-pr.md -->

Implements the inbound side of the chat WebSocket: clients can now send a
`dm.send` event over `/ws`; the server validates the payload, persists the
message, and emits `dm.message` to both sender and recipient. Invalid sends
receive a typed `chat.error` frame on the sender's connection only. This
completes the realtime delivery half of the chat MVP (the persistence layer
landed with C02, history with C03, presence with C04, roster with C05).

## Summary of Changes

### 1. Inbound event routing (`internal/handlers/ws.go`)
- `readPump` no longer discards inbound frames. Each text frame is unmarshalled
  into `ws.WSMessage` and dispatched by `type`. Malformed JSON is silently
  dropped (matches the existing presence-only contract — connections do not
  close on a bad frame).
- The read deadline is refreshed on every successful read so an actively
  chatting client does not get evicted by the 60-second timeout, while idle
  clients still rely on pong refresh.
- A `switch msg.Type` keeps the dispatch table small and extensible; today
  only `dm.send` is wired, but `typing` / `read` events plug in here later
  with no further plumbing.

### 2. `dm.send` handler (`internal/handlers/ws.go`)
- New `dmSendPayload` struct + `handleDMSend(senderID, senderClient, payload)`.
- Validation order — fail fast, cheapest check first:
  1. **Malformed JSON**: `INVALID_PAYLOAD` when the payload cannot be
     decoded into `dmSendPayload`.
  2. **Missing recipient**: `INVALID_RECIPIENT` when `recipient_id == 0`
     (the zero value of a missing or null field). Distinguished from
     `INVALID_PAYLOAD` so clients can surface a precise diagnostic.
  3. **Self-send**: `SELF_SEND` when `recipient_id == senderID`. Caught at
     the handler layer so the error code is precise; `db.CreateMessage`
     also rejects this, but its error is a generic "sender and recipient
     must be different" string we don't want to leak to the wire. The
     duplication is documented in code as intentional defense-in-depth.
  4. **Empty body**: `EMPTY_BODY` after `strings.TrimSpace`. Whitespace-only
     bodies are treated as empty (matches SDS § 6.1 "reject empty message
     body" and the repository-layer guard).
  5. **Recipient online**: `RECIPIENT_OFFLINE` via `hub.IsUserOnline`.
     Checked **before** the DB write so we never persist a message we
     refused to deliver — keeps history honest about what was actually
     sent during a session. The TOCTOU between this check and
     `SendToUser` (recipient may disconnect in between) is documented:
     the message is persisted and recoverable via REST history on
     reconnect.
- Persists via the existing `db.CreateMessage` under a `context.WithTimeout(5s)`
  derived from `context.Background()`. The bounded context caps how long
  a stuck DB call can pin the read pump; if the connection dies during the
  call, the timer still ensures the goroutine returns.
- Resolves `sender_username` from a **cache on the `*ws.Client` populated
  once at upgrade time** — no DB lookup per send. Falls back to `"unknown"`
  if the upgrade-time lookup happened to fail.
- Emits `dm.message` to **both** the sender and the recipient via
  `hub.SendToUser`. Echoing back to the sender is what makes other tabs
  of the same sender (and the sender's own composer) update without
  needing an HTTP refetch. `SendToUser`'s `ErrUserOffline` return is
  ignored — the persistence covers the offline case.
- On `db.CreateMessage` failure the sender gets a generic `INTERNAL_ERROR`;
  the DB error string is **never** echoed onto the wire.

### 3. Wire-shape and error-code typing (`internal/handlers/ws.go`)
- Replaced all `map[string]any` event payloads with typed structs:
  `dmMessagePayload` and `chatErrorPayload`, each carrying explicit
  `json` tags. The JSON contract is now locked at compile time and
  cannot drift if a field is renamed or added.
- Introduced named constants for every `chat.error` code
  (`INVALID_PAYLOAD`, `INVALID_RECIPIENT`, `SELF_SEND`, `EMPTY_BODY`,
  `RECIPIENT_OFFLINE`, `INTERNAL_ERROR`) so the canonical set lives in
  one place and is searchable.
- `marshalEvent(type, payload)` helper centralises envelope serialization.
  The discarded `json.Marshal` errors are documented as unreachable for
  typed structs with primitive fields.

### 4. Error event helper (`internal/handlers/ws.go`)
- `sendChatError(c, code, message)` enqueues a `chat.error` frame on a
  single client's send channel (non-blocking, drops on a full buffer like
  the rest of the hub). Errors are **sender-scoped** by design: the
  recipient is never told that someone tried to message them and failed.

### 5. Read-pump dispatch (`internal/handlers/ws.go`)
- `readPump` now resets the read deadline on every successful frame so
  actively chatting clients are not evicted by the 60-second idle timer.
- Dispatch is **synchronous** by design — one frame, one DB call, one
  emit, then the next frame. This gives per-connection ordering for free
  and was deliberately commented so future readers do not "optimize"
  with `go h.handleDMSend(...)`, which would lose ordering.
- Malformed top-level JSON is silently dropped (no `chat.error` because
  we have no `type` to attribute it to). Documented in code.

### 6. Hub client caching (`internal/ws/connection_manager.go`)
- `Client` gains a `Username` field, populated once at upgrade time and
  reused for every outbound `dm.message`. Eliminates a `GetUsersByIDs`
  DB lookup per send.

### 7. Tests (`internal/tests/dm_send_test.go`)

Integration tests that drive real WebSocket connections through the
test API (no mocked hub):

- `TestDMSend_DeliveredToSenderAndRecipient` — happy path. Verifies
  every field of the `dm.message` payload (`id`, `sender_id`,
  `recipient_id`, `sender_username`, `body`, `created_at`) on **both**
  connections — guards against single-sided delivery regressions.
- `TestDMSend_MessageIsPersisted` — sends a DM over WS, then queries
  `GET /api/v1/chats/{aliceID}/messages` as bob and asserts the body
  appears. Cross-layer guard: the realtime path and the history API
  must agree on what was stored.
- `TestDMSend_SelfSendRejected` — alice → alice yields `SELF_SEND`.
- `TestDMSend_EmptyBodyRejected` — empty string yields `EMPTY_BODY`.
- `TestDMSend_WhitespaceBodyRejected` — `"   "` yields `EMPTY_BODY`
  (regression guard for the trim).
- `TestDMSend_RecipientOfflineRejected` — recipient never connects to
  WS; sender gets `RECIPIENT_OFFLINE` and the DB is untouched.
- `TestDMSend_MissingRecipientRejected` — `dm.send` with no `recipient_id`
  yields `INVALID_RECIPIENT`.
- `TestDMSend_MalformedPayloadRejected` — `dm.send` with payload that
  is a JSON string instead of an object yields `INVALID_PAYLOAD`.
- `TestDMSend_MalformedJSONIgnored` — non-JSON top-level frame is
  silently dropped; the next valid `dm.send` still works.
- `TestDMSend_UnknownTypeIgnored` — frame with unknown `type`
  (e.g. `"typing"`) is silently dropped.
- `TestDMSend_DBErrorReturnsInternalError` — the `private_messages`
  table is dropped mid-session; the resulting `chat.error` carries
  `INTERNAL_ERROR` and the wire message must not leak any SQLite
  internals (asserted by substring check on `"sql"`, `"table"`,
  `"private_messages"`).

Per-frame ordering: the server dispatches frames strictly in order, so
the "silently dropped" tests use a deterministic technique — send the
suspect frame, then send a valid `dm.send`, and assert the **very next**
event is the `dm.message` for the follow-up. If the suspect frame had
triggered a `chat.error`, that error would arrive first and the body
mismatch would fail the test. No timing-dependent "absence of event"
probes needed (`assertNextIsDMMessage` helper).

### 8. Hub direct unit tests (`internal/ws/connection_manager_test.go`)

New file covering the three Hub methods that had zero coverage before
this PR — public API surface introduced by earlier Track C tickets
(C01/C04) that no integration test exercised:

- `TestGetOnlineUserIDs_EmptyHub` — empty hub returns no ids.
- `TestGetOnlineUserIDs_ReflectsAddAndRemove` — last-Remove cleanup
  contract (no stale ids after disconnect).
- `TestGetConnectionCount` — per-user count is the size of the active
  client set; `Add` returns `firstConnection=true` only on the first
  tab.
- `TestSetCallbacks_FireOnFirstAndLast` — `onConnect` fires only on
  offline→online, `onDisconnect` only on online→offline.
- `TestSetCallbacks_NilSafe` — hubs without callbacks must not panic.

Test helpers added:
- `sendDMSend(conn, recipientID, body)` — writes a typed `dm.send` frame.
- `tryReadWSMessage(conn, timeout)` — non-fatal read with a deadline.
- `drainUntilType(conn, type, timeout)` — reads frames until one of the
  target type appears, skipping presence noise. This is what lets the
  rejection tests survive interleaved presence broadcasts without
  fragile read-count assertions.
- `getChatMessages(h, token, otherUserID, beforeID)` — REST history
  fetch, used by the persistence cross-check.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C06**:
> "valid direct messages are persisted and delivered to sender and recipient;
> invalid sends are rejected with the agreed contract; delivered payloads
> match the SDS event shape."

- ✅ Valid DM is persisted and delivered to **both** parties
  (`TestDMSend_DeliveredToSenderAndRecipient`, `TestDMSend_MessageIsPersisted`)
- ✅ Self-send, empty body, whitespace body, offline recipient, and
  malformed payload are each rejected with the agreed `chat.error` codes
- ✅ `dm.message` payload shape matches SDS § 5.5 exactly (verified
  field-by-field, not just decoded)

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./... -timeout 180s` — **all packages pass** (23.9 s)
- [x] `go test ./internal/tests/ -run TestDMSend -v` — **11/11 pass**
- [x] `go test ./internal/ws/ -v` — **5/5 pass**
- [x] `go build ./...` — **clean**
- [x] `go vet ./...` — **clean**

### Coverage (handlers + ws packages)

C06 introduces `handleDMSend` and friends; their final coverage:

| Function | Coverage |
|---|---|
| `handlers.handleDMSend` | **96.4 %** |
| `handlers.readPump` | **90.5 %** |
| `handlers.HandleWebSocket` | 95.2 % |
| `handlers.sendChatError` / `marshalEvent` | 100 % |

This PR also clears pre-existing 0%-coverage gaps in the `ws.Hub` public
API (introduced by C01/C04, never covered by integration tests):

| Function | On `main` before this PR | After |
|---|---|---|
| `ws.SetCallbacks` | 0 % | **100 %** |
| `ws.GetOnlineUserIDs` | 0 % | **100 %** |
| `ws.GetConnectionCount` | 0 % | **100 %** |
| `ws.Remove` | 81.8 % | **90.9 %** |

Remaining uncovered branches in `ws.go` are infrastructure-bound and
have no deterministic trigger in test: WebSocket upgrade failure
(`upgrader.Upgrade` error), the pong handler closure (fires on a real
keepalive timer), and the write-pump close-channel race path.

### QA Checklist
- [x] Two users connect to `/ws`; one sends `dm.send`; both receive
  `dm.message` with matching `id`, `sender_id`, `body`, `created_at`.
- [x] A DM sent over WS appears in `GET /api/v1/chats/{userID}/messages`
  for both participants (history layer agrees with realtime layer).
- [x] `dm.send` with `recipient_id == senderID` produces
  `chat.error: SELF_SEND` on the sender, nothing on the wire elsewhere.
- [x] `dm.send` with `body: ""` and `body: "   "` both produce
  `chat.error: EMPTY_BODY`; nothing is persisted.
- [x] `dm.send` to a user who is not connected produces
  `chat.error: RECIPIENT_OFFLINE`; nothing is persisted.
- [x] `dm.send` with no `recipient_id` produces
  `chat.error: INVALID_PAYLOAD`; nothing is persisted.
- [x] Malformed JSON on the socket is silently ignored (connection
  stays open, no `chat.error` sent — matches the presence-only
  pre-C06 behavior on the same socket).

## Key Files Impacted
- `internal/handlers/ws.go` — read-pump dispatch, `handleDMSend`,
  `sendChatError`, `marshalEvent`, typed wire payloads, named error codes
- `internal/ws/connection_manager.go` — `Client.Username` cache field
- `internal/ws/connection_manager_test.go` — Hub direct unit tests (new)
- `internal/tests/dm_send_test.go` — integration tests + WS test helpers
- `docs/ticket-tracker.md` — C05 and C06 marked done, counts updated
- `docs/pr-message/C06-Realtime-DM-Send-and-Delivery-pr.md` — this file

## Audit Findings Addressed

An independent cold-start audit raised 14 items (PASS-WITH-NITS, zero
criticals). All actionable items were folded back in:

- ✅ Derived context with timeout for the DB call (nit #1)
- ✅ Username cached on `*ws.Client` at upgrade (nit #2)
- ✅ Distinct `INVALID_RECIPIENT` vs `INVALID_PAYLOAD` codes (nit #3)
- ✅ Self-send defense-in-depth documented (nit #4)
- ✅ `json.Marshal` error policy documented; typed structs prevent the
  failure mode (nit #5)
- ✅ TOCTOU window between `IsUserOnline` and `SendToUser` documented
  with the persist-and-recover rationale (nit #6)
- ✅ Silent malformed-frame drop documented (nit #7)
- ✅ Saturated-send-channel drop policy documented (nit #8)
- ✅ Canonical `chat.error` code set named and documented (nit #9)
- ✅ Synchronous-dispatch invariant documented to prevent future "go h.handle…"
  optimizations (nit #10)
- ✅ Test type-comparison now uses `json.Unmarshal` not quoted-string
  equality (nit #12)
- ✅ Dead `_ = bobID` removed (nit #13)
- ➖ Time-based test drains kept at 1–2 s (nit #14); the bounds are
  already generous relative to the local test runtime and will be
  re-evaluated if CI flakes surface.
