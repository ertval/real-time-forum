# C03: Chat History API
<!-- Filename: docs/pr-message/C03-Chat-History-API-pr.md -->

Implements the chat history API endpoint for retrieving private message history between the authenticated user and another user. This completes the backend persistence layer for chat history.

## Summary of Changes

### 1. Backend Handler
- **New Chat Handler**: Created `internal/handlers/chats.go` with `ChatsHandler` struct and `HandleChatMessages` method
- **GET /api/v1/chats/{userID}/messages**: Returns the latest 10 messages in chronological (oldest-first) order
- **Pagination**: Supports `before_id` query parameter to load older messages in batches of 10
- **Response Format**: Returns `has_more` boolean to indicate if more history exists
- **Optimization**: Added `db.GetUsersByIDs` to batch-fetch sender usernames in single query (avoids N+1)

### 2. Router Integration
- Added routes for `/api/v1/chats` and `/api/v1/chats/` to handle both path formats
- Protected by auth middleware (401 for unauthenticated users)
- Uses existing repository function `db.GetMessageHistory` for data retrieval

### 3. Test Updates
- Updated `TestChatRoutes_GuestCannotSuccessfullyAccess_CurrentStage` to expect 401 (Unauthorized) instead of 404, reflecting that the routes are now wired

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket C03:
> - the first request returns the latest 10 messages
> - older history returns in batches of 10
> - the API exposes whether more history exists

The implementation:
- Returns latest 10 messages when no `before_id` is provided
- Returns up to 10 older messages when `before_id` is provided
- Includes `has_more: true/false` in response to indicate additional history

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./...` — All tests pass, including existing `GetMessageHistory` repository tests and the updated auth access test
- [x] `go build ./...` — Build succeeds with no errors

### QA Checklist
- [x] Added new handler tests in `internal/tests/chat_history_test.go`:
  - `TestChatHistory_Unauthorized` — verifies unauthenticated users get 401
  - `TestChatHistory_AuthenticatedNoMessages` — empty conversation returns empty array
  - `TestChatHistory_WithMessages` — verifies 10 messages returned with has_more flag
  - `TestChatHistory_ResponseFormat` — verifies response has "data" wrapper
  - `TestChatHistory_SenderUsernameIncluded` — verifies sender_username is returned
  - `TestChatHistory_InvalidBeforeID` — verifies before_id validation
- [x] Repository layer (`GetMessageHistory`) already has comprehensive test coverage for:
  - Both directions (messages from both users)
  - 10-message limit
  - BeforeID pagination
  - Conversation isolation
  - Empty conversations

## Key Files Impacted
- `internal/handlers/chats.go` — New chat handler
- `internal/router/router.go` — Added chat routes
- `internal/tests/forum_auth_access_test.go` — Updated test expectations
- `internal/tests/chat_history_test.go` — New comprehensive handler tests