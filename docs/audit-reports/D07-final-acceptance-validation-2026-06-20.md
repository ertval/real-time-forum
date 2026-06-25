# D07 — Final Acceptance Validation (2026-06-20)

This is the final acceptance pass for the Real-Time Forum. It validates the
delivered build against the full `audit.md` checklist, the PRD success criteria,
and the SDS testing strategy. It also confirms the retained legacy features still
function and records the remaining gaps as explicit follow-up items.

**Scope of this ticket:** validation and documentation only. No production source
was modified.

## 1. Automated Verification Results

All commands executed on `2026-06-20`, branch `fix/e2e-harness-stale-server`.

| Gate | Command | Result |
| :--- | :--- | :--- |
| Backend (Go) | `make test-backend` (`go test ./...`) | **PASS** — all packages `ok` |
| Backend race | `go test -race ./internal/ws/ ./internal/handlers/` | **PASS** — no data races |
| Lint / format | `bun run check` (`biome check .`) | **PASS** — 147 files, no diagnostics |
| Frontend (Vitest) | `bun run test` (`vitest run`) | **PASS** — 35 files, **327 tests** |
| E2E (Playwright) | `make test-e2e` | **PASS** — **21 tests** |
| Infrastructure | `make verify-infra` | **PASS** — bootstrap → build → run/stop all green |

`make test` (the aggregate of backend + frontend + E2E) completes with exit code
0. `make verify-infra` confirms a clean `node_modules` bootstrap, a full build of
both servers, and correct process management (servers come up, proxy works,
`make stop` tears them down).

## 2. Functional Audit Checklist (`audit.md`)

Every mandatory audit question is satisfied with automated evidence.

| # | Audit Question | Verdict | Evidence |
| :--- | :--- | :--- | :--- |
| 1 | Allowed packages respected? | ✅ | `go.mod` limited to gorilla/websocket, mattn/go-sqlite3, x/crypto/bcrypt, google/uuid; A01 CI policy + `bun run policy` import-path gate |
| 2 | Register/login required to use the forum? | ✅ | `A05-01` (protected routes redirect to `/login`), `A05-03` (protected APIs return 401 when unauthenticated) |
| 3 | Registration asks nickname, age, gender, first/last name, email, password? | ✅ | D10 register view; `auth.handlers.test.js`; `A07-01` reads back age/gender/first/last after register |
| 4 | Unregistered login fails to enter the forum? | ✅ | Backend C08/C11 auth coverage (`internal/tests`, `internal/handlers`); invalid-credential rejection |
| 5 | Login accepts nickname **or** email + password? | ✅ | C11 contract; `auth.handlers.test.js`; E2E `loginUser` flow |
| 6 | Registered user can log in? | ✅ | `A05-02` (auth entry resolves to forum shell), `A06-02` re-login loop |
| 7 | Logout reachable on every page? | ✅ | `A06-01` (logout visible+enabled on all protected routes), `A06-02` (redirects to `/login`), `A06-03` (post-logout content guarded) |
| 8 | Can create a post? | ✅ | B04 create/edit flows; `post.api.test.js`, `post.page.test.js`; `A07-02` creates a post via API |
| 9 | Can see a previously created post? | ✅ | `feed.state.test.js`, `feed.views.test.js`; feed renders posts |
| 10 | Can comment on a post? | ✅ | B03 post-detail flow; `post-detail.views.test.js`, `comment-highlight.test.js` |
| 11 | Comment visible on post detail (not feed)? | ✅ | B02 (feed has no comments) + B03; `feed.views.test.js` asserts no comment rendering, `post-detail.views.test.js` asserts comments |
| 12 | Section showing online users? | ✅ | D01 roster; `chat.roster.views.test.js`, `chat.roster.live.test.js`; `A04-01/03` roster always visible in shell |
| 13 | Chat users ordered by last message (Discord-style)? | ✅ | `chat.roster.logic.test.js` (recency ordering) |
| 14 | New user with no messages ordered alphabetically? | ✅ | `chat.roster.logic.test.js` (alpha fallback after active conversations) |
| 15 | Message format includes username and date? | ✅ | `chat.conversation.views.test.js` (sender + timestamp rendering) |
| 16 | Recipient receives notification of a new DM? | ✅ | D04; `create-app.chat-socket.test.js`, `chat.conversation.live.test.js`, `chat_regression.test.mjs` |
| 17 | Recipient receives the message in real time (no refresh)? | ✅ | D04; `chat.conversation.live.test.js`, `chat_regression.test.mjs` (live append on `dm.message`) |
| 18 | Opening a >10-message conversation shows only the last 10? | ✅ | C03 history API; `chat.conversation.history.test.js`, `chat.conversation.api.test.js` |
| 19 | Scroll up loads more (uses the scroll event)? | ✅ | D03; `chat.conversation.history.test.js` (scroll-to-top requests older batch with `before_id`) |
| 20 | Loads 10 at a time without spamming the scroll event (Throttle)? | ✅ | `throttle.test.js`, `chat.conversation.history.test.js` (no burst requests) |

### Bonus questions

| Bonus Question | Verdict | Evidence |
| :--- | :--- | :--- |
| Users have profiles? | ✅ | A07; `profile.views.test.js`; `A07-01` (fields), `A07-02` (reachable from author links) |
| Send images through private messages? | ✅ | C09 backend + D08 frontend; `image-picker.test.js`, conversation image rendering tests |
| Uses concurrency (Promises + Go routines/channels)? | ✅ | `internal/ws/connection_manager.go` (goroutines + channels), `go test -race` clean; frontend uses Promises/async throughout API + WS layers |
| Runs quickly / no unnecessary requests? | ✅ | Throttled history loading; roster/presence updated over WS rather than polling; SPA avoids full reloads (`A03-01`) |
| Code obeys good practices / well done overall? | ✅ | Biome clean, layered backend (no SQL in handlers), vertical-slice SPA, 348 automated checks green |

## 3. PRD Success Criteria

All Section 9 success criteria are met:

- ✅ Single HTML shell (`A10`, `A02`)
- ✅ Forum content restricted to authenticated users (`A05`)
- ✅ Extended registration fields present and persisted (`C10`, `C11`, `D10`)
- ✅ Comments removed from the feed (`B02`)
- ✅ Logout reachable from every authenticated screen (`A06`)
- ✅ Online/offline presence (`C04`, `D01`)
- ✅ Prior DM history readable, including with offline users (`C03`, `D02`)
- ✅ Real-time send/receive of private messages (`C06`, `D04`)
- ✅ Older messages load in batches of 10 on upward scroll (`C03`, `D03`)

Bonus success criteria also met: viewable profiles (`A07`), DM image attachments
(`C09`/`D08`), and concurrency patterns (goroutines/channels + Promises).

## 4. SDS Testing-Strategy Conformance

The Vitest hierarchy in `SPA/tests/` matches SDS §10.2.1 exactly:

| Category | SDS Location | Present | Files |
| :--- | :--- | :--- | :--- |
| Unit | `SPA/tests/unit/` | ✅ | 31 suites (core, features, helpers, policy) |
| Integration | `SPA/tests/integration/` | ✅ | `frontend_behavior.test.mjs`, `chat_regression.test.mjs` |
| E2E | `SPA/tests/e2e/` | ✅ | `tickets.test.js` (Playwright, 21 tests) |

Backend test coverage (SDS §10.1) is delivered by C08 in `internal/tests/` and
co-located `_test.go` files. Bonus race verification (SDS §10.4) confirmed via
`go test -race`.

## 5. Retained Legacy Features

Confirmed working and regression-protected (D05 SPA/forum coverage, D06 chat
coverage):

- ✅ Posts and categories (B01, B04)
- ✅ Comments and reactions (B03, B07)
- ✅ Post/comment image uploads (D05 coverage)
- ✅ Drafts (B08)
- ✅ Activity view (B05)
- ✅ Notifications (B06)

## 6. Remaining Gaps & Follow-ups

No mandatory or bonus audit requirement is unmet. The following are
non-blocking cleanups recorded here rather than hidden:

1. **Orphaned legacy `web/static/js/create-post.js`** still performs hard
   `window.location.href` navigations (lines 319, 373). It is dead standalone
   code — the live create-post flow runs inside the SPA (`features/post`,
   verified by `A03-03`/`A04-01`). Follow-up: delete the legacy script and its
   `web/templates` siblings once the legacy `web/` tree is fully retired.
2. **WebSocket auto-reconnect / backoff / disconnected-state recovery** is
   intentionally out of scope for this phase, as declared by D04 and documented
   in `SPA/core/realtime/chat-socket.js`. Follow-up candidate for a resilience
   ticket.
3. **Legacy `web/templates/` + `web/static/`** assets remain in the tree for
   backward-compatible static serving (exercised by `A02-03`). They are not part
   of the SPA and can be pruned in a dedicated cleanup pass.

## 7. Verdict

**ACCEPTED.** The build satisfies every mandatory audit question and PRD success
criterion, all bonus features are implemented and verified, retained legacy
features are confirmed working, and the only open items are explicitly-scoped
non-blocking cleanups. All automated gates (Go, race, Biome, 327 Vitest, 21
Playwright, infrastructure) are green.
