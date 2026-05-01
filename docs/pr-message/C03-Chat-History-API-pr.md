# C03: Chat History API
<!-- Filename: docs/pr-message/C03-Chat-History-API-pr.md -->

Implements `GET /api/v1/chats/{userID}/messages` — the chat history endpoint that returns paginated private message history between the authenticated user and another user, oldest-first and render-ready.

## Summary of Changes

### 1. Backend Handler (`internal/handlers/chats.go`)
- **`GET /api/v1/chats/{userID}/messages`**: returns the latest 10 messages in chronological order
- **Pagination**: `?before_id=<id>` loads the next older batch of 10
- **`has_more` correctness**: fetches 11 rows internally; if 11 arrive the 11th is discarded and `has_more: true` is set — eliminates the false-positive that occurred when a page landed on exactly 10 messages
- **Sender username**: batch-fetched via `db.GetUsersByIDs` to avoid N+1; DB errors return 500; missing user rows (orphaned sender) fall back to `"unknown"` rather than leaking an empty string to the client
- **Response shape**: `{"data": {"messages": [...], "has_more": bool}}` — matches SDS contract

### 2. DB Layer (`internal/db/messages.go`)
- **`GetMessageHistory` return signature**: changed from `([]PrivateMessage, error)` to `([]PrivateMessage, bool, error)` — the bool is `hasMore`, computed from the sentinel row; callers no longer need to infer it from `len`

### 3. Router (`internal/router/router.go`)
- Routes `/api/v1/chats` and `/api/v1/chats/` registered, auth-gated (401 for unauthenticated)

### 4. Test Updates
- `forum_auth_access_test.go`: chat routes now expect 401 instead of 404
- `messages_test.go`: all `GetMessageHistory` call sites updated to the new three-value return

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C03**:
> - the first request returns the latest 10 messages
> - older history returns in batches of 10
> - the API exposes whether more history exists

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./... -timeout 60s` — all packages pass
- [x] `go build ./...` — clean build

### QA Checklist
- [x] `TestChatHistory_Unauthorized` — unauthenticated requests get 401
- [x] `TestChatHistory_AuthenticatedNoMessages` — empty conversation returns `[]`, `has_more: false`
- [x] `TestChatHistory_WithMessages` — correct 10-message page with `has_more: true`
- [x] `TestChatHistory_ResponseFormat` — response has `data` wrapper per SDS
- [x] `TestChatHistory_SenderUsernameIncluded` — `sender_username` is present
- [x] `TestChatHistory_InvalidBeforeID` — bad `before_id` values return 400
- [x] `TestChatHistory_HasMoreFalse_ExactlyTen` — exactly 10 messages → `has_more: false`
- [x] `TestChatHistory_HasMoreFalse_LastPage` — last page of a 20-message conversation → `has_more: false`
- [x] `TestChatHistory_SenderUsername_ExactMatch` — `sender_username` matches the registered username exactly
- [x] `TestChatHistory_SenderUsername_OrphanedSender` — deleted sender produces `"unknown"`, not `""`

### Audit Findings Resolved
- **`has_more` false positive**: `len(messages) == 10` wrongly returned `true` when the page landed on exactly 10 messages or on the last page of pagination. Fixed by fetching 11 rows and using the sentinel.
- **Silent empty-string username**: `GetUsersByIDs` returning no rows for a missing user produced `sender_username: ""`. Fixed with an explicit `"unknown"` fallback in the non-error path.
- **Silent 200 on DB error**: a `GetUsersByIDs` failure previously fell back to `"unknown"` and returned 200. Now returns 500.
- **Fragile `aliceID+1` in test**: `TestChatHistory_AuthenticatedNoMessages` referenced bob's ID by assumption rather than capturing the return value. Fixed.

## Key Files Impacted
- `internal/handlers/chats.go` (new)
- `internal/db/messages.go` (modified — return signature)
- `internal/router/router.go` (modified)
- `internal/tests/chat_history_test.go` (new)
- `internal/tests/messages_test.go` (modified — call sites)
- `internal/tests/forum_auth_access_test.go` (modified)
