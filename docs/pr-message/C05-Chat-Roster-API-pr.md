# C05: Chat Roster API
<!-- Filename: docs/pr-message/C05-Chat-Roster-API-pr.md -->

Implements `GET /api/v1/chats` — the persistent chat roster endpoint that lists
every other user with presence and last-message metadata, ordered for direct
rendering in the chat sidebar.

## Summary of Changes

### 1. DB layer (`internal/db/messages.go`)
- New `RosterEntry` struct + `GetChatRoster(ctx, db, viewerID)` function.
- Single SQL pass using a `ROW_NUMBER()` window over `private_messages` to find
  the most recent message per pair, then a `LEFT JOIN` against `users`. No
  per-user N+1; one query for the whole roster.
- Sort key: `(has_history DESC, last_message_at DESC, msg_id DESC, LOWER(username) ASC)`.
  The `msg_id DESC` tiebreak makes the order deterministic even when two pair
  conversations share the same `created_at` second.

### 2. Handler (`internal/handlers/chats.go`)
- `ChatsHandler` now takes a `*ws.Hub` so it can decorate each roster row with
  `is_online`.
- New `HandleChatRoster` for `GET /api/v1/chats`:
  - 401 when unauthenticated
  - 500 on DB error
  - returns `{ data: [...] }` matching SDS § 5.3
  - response is a typed struct (`rosterResponseEntry`), not a `map[string]any` —
    locks the JSON contract at compile time
  - `last_message_at` / `last_message_preview` / `last_sender_id` use pointer
    fields and serialize as JSON `null` for users without history
- Method enforcement is delegated to `middleware.AllowMethods` at the route
  layer (no redundant in-handler `r.Method` switch).
- `last_message_preview` is capped at 200 chars at the SQL level
  (`SUBSTR(body, 1, 200)`), so a multi-MB DM body never inflates the roster
  payload. SQLite's `substr()` is codepoint-aware on TEXT, so the cap cannot
  split a multi-byte character.

### 3. Router (`internal/router/router.go`)
- Registers `apiPrefix + "/chats"` → `HandleChatRoster`, auth-gated.
- The `/api/v1/chats/` subtree continues to route to the C03 history handler.
- `NewChatsHandler` call updated to pass the hub.

### 4. Tests

**Handler-level (`internal/tests/chat_roster_test.go`):**
- `TestChatRoster_Unauthorized` — guests get 401
- `TestChatRoster_OnlyOtherUsers` — viewer is excluded from their own roster
- `TestChatRoster_PresenceReflected` — `is_online` matches `hub.IsUserOnline`
- `TestChatRoster_OrderingByLastMessage` — history-first by recency, then
  alphabetical (case-insensitive: "Adam" < "zach")
- `TestChatRoster_DBError_Returns500` — drives the 500 branch by dropping
  `private_messages` after login (auth still succeeds, roster query fails)
- `TestChatRoster_ResponseShape` — every entry carries exactly the field names
  defined in SDS § 5.3 (no extras, none missing)
- `TestChatRoster_EmptyWhenNoOtherUsers` — empty roster serializes as
  `"data": []`, never `null`
- `TestChatRoster_LatestSpansBothDirections` — last-message fields reflect the
  most recent row in either direction, not just the viewer's sends
- `TestChatRoster_PreviewTruncated` — long DM bodies don't flow through the
  roster verbatim (regression guard for the SQL truncation)
- `TestChatRoster_LastMessageMetadata` — last-message fields match the most
  recent message in the pair; `null` for users without history

**DB-level (`internal/db/messages_test.go`):**
- `TestGetChatRoster_QueryErrorPropagates` — closed DB triggers
  `QueryContext` error; the wrap is asserted
- `TestGetChatRoster_ScanErrorPropagates` — synthetic schema with non-numeric
  `id` triggers the row-scan error path; the wrap is asserted

**Coverage** (run with `-coverpkg=./internal/db,./internal/handlers`):
- `db.GetChatRoster` — **92.3 %**
- `handlers.HandleChatRoster` — **90.0 %**
- `handlers.NewChatsHandler` — 100 %

The remaining ~10 % on each is the post-iteration `rows.Err()` check, which
fires only on connection failure mid-iteration and has no deterministic
trigger in `database/sql`.

**Test plumbing:**
- `internal/tests/helpers_test.go` gains a sibling helper
  `newTestAPIWithHub` so roster tests can mark users online via
  `hub.Add(uid, nil)` without standing up a real WebSocket.
- `forum_auth_access_test.go` re-asserts that `/api/v1/chats` is now wired
  and 401-gated for guests.

## Verification Gate Satisfaction
- ✅ Roster entries include `is_online` and last-message metadata
- ✅ Users with history sort by latest activity (DESC)
- ✅ Users without history sort alphabetically after active conversations

## Testing
- `go test ./...` — all packages pass
- `go build ./...` — clean
- `go vet ./internal/db/ ./internal/handlers/ ./internal/router/` — clean

## Notes
- This branch is stacked on `medvall/C03` because C05 extends `chats.go`. Once
  C03 merges, rebase onto main.
