# C09: DM Image Upload Backend (Bonus)

Adds image attachments to private messages: a multipart upload endpoint that
stores an image and returns its URL, an optional `image_url` on the `dm.send`
WebSocket event, propagation of `image_url` on `dm.message`, and `image_url` in
the chat-history API. This is the backend half of the DM-image bonus feature
(SDS §4.4 / §5.7); the inline rendering frontend is D08. Implements ticket C09
and unblocks D08.

## Summary of Changes

### 1. Schema & Migration (`internal/db`)
- **`image_path` column**: added `image_path TEXT DEFAULT NULL` to
  `private_messages` in `forum_schema.sql`.
- **Legacy migration**: because `private_messages` predates this column,
  `CREATE TABLE IF NOT EXISTS` is a no-op on already-deployed databases (the
  C07 problem). Added `{"private_messages", "image_path", "TEXT DEFAULT NULL"}`
  to `requiredColumns` so `Migrate` ALTERs existing databases, and corrected the
  migrate comment that previously cited `private_messages` as a table that never
  needs an entry. Nullable, so the ALTER cannot fail on populated tables.

### 2. Message Repository (`internal/db/messages.go`)
- **`PrivateMessage.ImagePath *string`** and **`CreateMessageRequest.ImagePath
  string`** added. Empty request value is stored as SQL `NULL` (no attachment),
  read back via `sql.NullString` so absent images serialize as omitted.
- **`CreateMessage` / `GetMessageHistory`** updated to write and read
  `image_path`. The data layer stays free of `net/http`/`json` per AGENTS.md.

### 3. Upload Endpoint & Transport (`internal/handlers`, `internal/router`)
- **`POST /api/v1/chats/{userID}/images`**: new dispatcher in `HandleChatMessages`
  routes `/messages` (GET → history) and `/images` (POST → upload).
  `uploadDMImage` reuses the existing multipart/validation helpers, saves to
  `web/static/uploads/dm/`, and returns `{ "data": { "image_url": ... } }`
  (SDS §5.7). It does not create a message — the sender includes the URL in a
  subsequent `dm.send`.
- **Shared save helper**: `saveUploadedImage` refactored into a thin wrapper over
  `saveUploadedImageToSubdir`, so the `dm/` subdirectory path is shared without
  duplicating the type→extension and copy logic; post/comment behaviour is
  unchanged.
- **`dm.send` / `dm.message`**: `dm.send` now accepts an optional `image_url`;
  `dm.message` echoes it (`omitempty`). When non-empty, `image_url` is
  validated against `^/static/uploads/dm/<file>.<jpg|png|gif>$` plus a `..`
  guard and rejected with a new `chat.error` code `INVALID_IMAGE` — the value
  is rendered as an `<img src>` by D08, so it is bound to files our own endpoint
  produced rather than trusted.
- **History API**: `image_url` added to each history message when present.

### Design notes
- **Body remains required even with an image** — SDS §6.1 ("reject empty
  message body") and the `CHECK (length(trim(body)) > 0)` on `private_messages`.
  Image-only messages would require a CHECK-dropping table rebuild and are out
  of scope.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C09**:
> - DM image upload saves the file to disk and returns a valid URL
> - messages with images are persisted with the `image_path` column populated
> - `dm.message` events include `image_url` when an image is attached
> - history API returns `image_url` for messages that have images

| Gate item | Where satisfied |
|-----------|-----------------|
| Upload saves to disk + returns URL | `chats.go` `uploadDMImage` → `posts_helpers.go` `saveUploadedImageToSubdir(..,"dm")` |
| `image_path` persisted | `forum_schema.sql`, `messages.go` `CreateMessage` |
| `dm.message` carries `image_url` | `ws.go` `dmMessagePayload` + `handleDMSend` |
| History returns `image_url` | `messages.go` `GetMessageHistory` + `chats.go` `getChatHistory` |

## Testing & Validation Verified

### Automated Test Suite
- [x] `go build ./...` — clean.
- [x] `go vet ./...` — clean.
- [x] `go test ./internal/...` — all packages pass.
- [x] `go test -race ./internal/...` — pass (race detector clean).

### Coverage of new/changed paths
| Function | Coverage |
|----------|----------|
| `isValidDMImageURL` | 100.0% |
| `handleDMSend` | 97.1% |
| `CreateMessage` | 90.0% |
| `parseChatTargetID` | 88.9% |
| `Migrate` | 87.0% |
| `GetMessageHistory` | 88.5% |
| `uploadDMImage` | 82.6% |
| `saveUploadedImageToSubdir` | 78.3% |

### QA Checklist (new tests)
- [x] `internal/tests/dm_images_test.go` — upload saves JPEG/PNG/GIF to `dm/`
      and returns URL; rejects non-image (400), missing file (400), oversized
      (413), bad recipient id (400), wrong method (405), unauthenticated (401);
      unknown chat subroute (400); dm.send with image delivered to both parties
      and surfaced in history; invalid `image_url` → `INVALID_IMAGE` with zero
      persistence; empty body + image still `EMPTY_BODY`.
- [x] `internal/tests/messages_test.go` — `CreateMessage` persists/omits
      `image_path`; `GetMessageHistory` returns it.
- [x] `internal/tests/migrate_test.go` — legacy `private_messages` (no
      `image_path`) gains the column, legacy row preserved with NULL,
      idempotent on re-run.

## Independent Audit (cold-start)

A fresh-context audit evaluated the change against SDS §4.4/§5.7/§6.1,
`docs/audit.md`, `docs/requirements.md`, and `AGENTS.md` without reading the
plan. **Verdict: PASS** — no critical findings; all four gate items confirmed
with file:line evidence; build/vet/tests green; path-traversal/injection on
`image_url` confirmed safe; migration confirmed to ALTER existing (not just
fresh) databases; no DB error strings leak to the wire; data-layer boundary
respected. Non-blocking nits, accepted by design:
1. **Orphaned uploads** — a file uploaded but never attached (or attached via a
   rejected `dm.send`) is not reaped. Inherent to the SDS upload-then-send
   design; a janitor would be a separate ticket.
2. **Upload endpoint does not validate recipient existence/self-send** —
   `{userID}` is parsed for route conformance only; real validation occurs on
   `dm.send`, and the file is not bound to the recipient.
3. **`PrivateMessage.ImagePath` json tag is cosmetic** — consistent with the
   struct's other tags (handlers build their own response shapes).

## Key Files Impacted
- `internal/db/forum_schema.sql`
- `internal/db/migrate.go`
- `internal/db/messages.go`
- `internal/handlers/chats.go`
- `internal/handlers/posts_helpers.go`
- `internal/handlers/ws.go`
- `internal/router/router.go`
- `internal/tests/dm_images_test.go`
- `internal/tests/messages_test.go`
- `internal/tests/migrate_test.go`
