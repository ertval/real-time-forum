# C08: Backend Test Coverage for Auth, Messaging, and Presence

Verifies and closes out backend test coverage for the auth, messaging, and
presence flows. A coverage audit found the C08 verification gate was **already
satisfied** by tests delivered incrementally in C03–C06; this PR documents that
mapping and adds a focused set of tests to close the genuinely-untested real
paths (session cleanup, single-active-session invalidation, auth endpoint
error paths). Implements ticket C08; a prerequisite for D07.

## Summary of Changes

### 1. Session lifecycle coverage (`internal/tests/session_lifecycle_test.go`, new)
- **`TestCleanupSessions_RemovesInvalidAndExpired`**: covers `db.CleanupSessions`
  (previously 0% — it runs in the backend's periodic cleanup goroutine at
  `cmd/backend/server.go:52`). Seeds a valid, an invalidated, and an expired
  session (across distinct users, per the partial unique index) and asserts only
  the valid one survives.
- **`TestSession_NewLoginInvalidatesPreviousSession`**: covers the
  single-active-session rule — `CreateSession` bumps `session_version` and
  invalidates prior sessions. Asserts a re-login issues a new token and the old
  token is rejected (401) while the new one works (200).

### 2. Auth endpoint error paths (`internal/tests/auth_errors_test.go`, new)
- Malformed JSON body on register/login → `400`.
- `GET` on the POST-only register/login routes → `405` (the `AllowMethods`
  contract).

### 3. Dead-code removal (`internal/middleware/auth.go`)
- Removed **`OptionalAuth`** — surfaced by this ticket's audit as dead code with
  no non-test callers. Rather than write a test pinning unused code (test
  over-engineering), the function is deleted. `go build`/`go vet`/`go test ./...`
  confirm nothing depended on it.

### 4. Documentation
- Coverage report (below) mapping every gate clause to its proving test.
- `docs/ticket-tracker.md` → C08 `[x]`.

### Deliberately out of scope (avoiding test over-engineering)
- The remaining uncovered lines in `bumpSessionVersionTx` /
  `invalidateUserSessionsTx` are SQL-error branches; exercising them would
  require sabotage with little behavioral value.
- No re-assertion of flows already covered by C03–C06.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C08**:
> - backend tests cover the major auth and chat flows
> - both supported login modes are explicitly tested
> - offline-recipient and multi-connection presence cases are covered

### Gate clause → proving test

| Gate clause | Proving test(s) |
|---|---|
| Major auth flows — registration (+ extended profile fields) | `users_test.go`: `TestUserRegistration`, `TestC11_Registration_MissingProfileFields`, `TestC11_Registration_ValidPayloadSucceeds` |
| Major auth flows — login / wrong-password / non-existent | `users_test.go`: `TestUserLogin_WrongPassword`, `TestUserLogin_NonExistentUsername`, `TestUserLogin_NonExistentEmail` |
| Major auth flows — logout / session invalidation | `users_test.go`: `TestUserLogout`, `TestUserMe_NoCookie`, `TestUserMe_InvalidCookie`; `session_lifecycle_test.go` (new) |
| Auth-gating | `forum_auth_access_test.go` |
| **Both login modes explicitly tested** | `users_test.go`: `TestUserLoginByUsername` **+** `TestUserLoginByEmail` |
| Chat — roster | `chat_roster_test.go` |
| Chat — history | `chat_history_test.go`, `messages_test.go` |
| Chat — DM send/delivery | `dm_send_test.go` |
| Chat — WebSocket auth | `ws_test.go`: `TestWebSocket_Unauthenticated`, `TestWebSocket_InvalidSession` |
| **Offline-recipient** | `dm_send_test.go`: `TestDMSend_RecipientOfflineRejected` |
| **Multi-connection presence** | `presence_test.go`: `TestPresence_SecondTabDoesNotBroadcast`, `TestPresence_NonLastDisconnectDoesNotBroadcast`; `ws_test.go`: `TestWebSocket_MultiTab` |

## Testing & Validation Verified

### Automated Test Suite
- [x] `go build ./...` — clean.
- [x] `go vet ./...` — clean.
- [x] `go test ./internal/...` — all packages pass.
- [x] `go test -race ./internal/...` — pass (race detector clean).

### Coverage movement (new tests)
| Function | Before | After |
|----------|--------|-------|
| `db.CleanupSessions` | 0.0% | 100.0% |
| `handlers.Register` | 63.2% | 73.7% |
| `handlers.Login` | 63.2% | 73.7% |

## Independent Audit (cold-start)

A fresh-context audit independently re-verified each gate clause against
`docs/track-c.md` / `docs/SDS.md` (without reading the plan) and confirmed the
"already satisfied by C03–C06" thesis clause-by-clause. **Verdict: PASS** — no
critical findings; the two new test files assert real behavior and are stable
across repeated runs; build/vet/tests green. Audit nits addressed: a test
comment was corrected to describe the sessions index as a *partial* unique index
(one valid session per user); the `OptionalAuth` dead-code finding is now
resolved by deleting the function in this PR (the only production change).

## Key Files Impacted
- `internal/tests/session_lifecycle_test.go` (new)
- `internal/tests/auth_errors_test.go` (new)
- `internal/middleware/auth.go` (removed dead `OptionalAuth`)
- `docs/ticket-tracker.md`
- `docs/pr-message/C08-Backend-Test-Coverage-pr.md` (new)
