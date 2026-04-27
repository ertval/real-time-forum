# C02: Private Messages Schema and Repository Layer
<!-- Filename: docs/pr-message/C02-PrivateMessagesSchema-pr.md -->

Adds the `private_messages` table, required indexes, and the two repository functions that C03–C06 depend on. Pure persistence layer — no HTTP handlers, no WebSocket events.

## Summary of Changes

### 1. Schema (`internal/db/forum_schema.sql`)
- **`private_messages` table**: `id`, `sender_id`, `recipient_id`, `body`, `created_at`; FK cascades on both users; `CHECK (sender_id <> recipient_id)`; `CHECK (length(trim(body)) > 0)`
- **`idx_pm_sender`**: `(sender_id, recipient_id, created_at DESC)` — covers outbound pair lookups
- **`idx_pm_recipient`**: `(recipient_id, sender_id, created_at DESC)` — covers inbound pair lookups

### 2. Repository (`internal/db/messages.go`)
- **`CreateMessage`**: inserts a row and returns the full persisted `PrivateMessage`; rejects self-send and whitespace-only body at the application layer (matching the DB constraint)
- **`GetMessageHistory`**: pair-based lookup using `UNION ALL` (one branch per index) so each index can be used directly; returns up to 10 messages ordered oldest-first; supports `beforeID > 0` for backwards pagination

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C02**:
> - messages can be persisted between two users
> - pair-based history queries work
> - schema and repository tests cover the new table

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./internal/... -timeout 60s` — all packages pass
- [x] `go build ./...` — builds without errors
- [x] `bun run policy` — biome: 82 files, no issues; vitest: 23 tests passing

### QA Checklist
- [x] `TestCreateMessage_Persists` — row written, all fields returned correctly
- [x] `TestCreateMessage_SelfSendRejected` — self-send returns error
- [x] `TestCreateMessage_EmptyBodyRejected` — empty body returns error
- [x] `TestCreateMessage_WhitespaceBodyRejected` — whitespace-only body returns error (matches DB `trim` constraint)
- [x] `TestGetMessageHistory_BothDirections` — messages from both sides of a conversation returned in chronological order
- [x] `TestGetMessageHistory_LimitsTen` — never returns more than 10 messages
- [x] `TestGetMessageHistory_LimitsTenLatest` — the 10 returned are the newest, not the oldest
- [x] `TestGetMessageHistory_BeforeIDPagination` — older page returns messages with IDs below the pivot
- [x] `TestGetMessageHistory_ConversationIsolation` — messages from unrelated pairs do not appear
- [x] `TestGetMessageHistory_EmptyConversation` — returns empty slice with no error

### Manual E2E Verification
Started the real backend server, registered two users via curl, then verified the schema and queries directly in `sqlite3`.

- [x] `private_messages` table and both indexes (`idx_pm_sender`, `idx_pm_recipient`) present after server boot
- [x] Self-send rejected at DB level — `CHECK constraint failed: sender_id <> recipient_id`
- [x] Whitespace-only body rejected at DB level — `CHECK constraint failed: length(trim(body)) > 0`
- [x] Latest-10 query returns IDs 6–15 (newest), not 1–10
- [x] Pagination (`beforeID=6`) returns IDs 1–5 in chronological order
- [x] Conversation isolation — inserting an unrelated pair message does not change the alice-bob count
- [x] `EXPLAIN QUERY PLAN` confirms both UNION ALL branches hit an index (`SEARCH private_messages USING INDEX idx_pm_recipient`) — no full table scans

## Key Files Impacted
- `internal/db/forum_schema.sql` (modified — table + 2 indexes)
- `internal/db/messages.go` (new)
- `internal/tests/messages_test.go` (new)
