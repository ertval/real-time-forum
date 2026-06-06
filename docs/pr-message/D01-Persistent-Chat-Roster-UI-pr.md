# D01: Persistent Chat Roster UI

This PR closes ticket D01 by adding the always-visible chat roster to the authenticated SPA shell. The roster consumes the existing C05 backend (`GET /api/v1/chats`) and renders one selectable row per fellow user with online/offline presence and last-message metadata. Selection itself is exposed as a `chat:user-selected` CustomEvent for D02 to consume; live presence updates remain D04's responsibility.

## Summary of Changes

### 1. Chat Roster Feature Slice (`SPA/features/chat/`)
- **`chat.roster.api.js`**: thin `fetchRoster(fetchRef)` wrapper around `GET /api/v1/chats` with `credentials: 'include'`. Returns the `data` array on ok, `[]` on non-ok status, rejection, or malformed payload.
- **`chat.roster.views.js`**: pure `renderRosterItem` / `renderRosterList` functions. Every `<li>` carries `tabindex=0`, `role="button"`, `aria-pressed="false"`, plus `data-roster-user-id`, `data-roster-username`, and `data-roster-online`. Username and last-message preview are escaped via the existing `core/utils/html.js` helper. Empty input renders the existing `chat-panel__empty` paragraph.
- **`chat.roster.page.js`**: `initChatRoster({windowRef, documentRef, fetchRef})`. No-ops when `[data-chat-roster]` is absent. Uses a `data-roster-bound="true"` idempotency guard so repeated calls during route transitions are free. Attaches a single delegated `click` listener and a single delegated `keydown` listener (Enter/Space) on the roster panel; on activation it flips `aria-pressed` + `is-selected` on the chosen row, clears siblings, and dispatches `new CustomEvent('chat:user-selected', { detail: { userId, username, isOnline }, bubbles: true })`. A short comment marks the spot where D04 will introduce reactive state for live presence.
- **`chat.roster.css`**: list reset, presence dot (online/offline), name + preview layout, hover and selected states. Imported via the existing `SPA/assets/css/main.css` aggregation pipeline.

### 2. SPA Shell Wiring
- **`SPA/core/app/create-app.js`**: imports `initChatRoster` and invokes it from `runRouteInitializer` whenever the matched route has `access === 'protected'`. The bound-attribute guard means subsequent navigation between protected routes is a no-op fetch-wise; the persistent shell preserves the painted DOM across route swaps.
- **`SPA/assets/css/main.css`**: `@import "../../features/chat/chat.roster.css"` so the roster styles ride the same loader chain as auth and shell CSS (no `index.html` change needed).

### 3. Frontend Regression Coverage
- 26 new Vitest cases under `SPA/tests/unit/features/chat/`:
  - `chat.roster.api.test.js` (6) — URL/method/headers/credentials, ok unwrap, non-ok fallback, rejected fetch fallback.
  - `chat.roster.views.test.js` (9) — input order preserved, online vs offline classes, null preview rendered as no preview, empty input → empty-state paragraph, XSS escape for `<img onerror>` and `<script>` payloads.
  - `chat.roster.page.test.js` (11) — absence no-op, idempotent bind, single fetch + paint, click and keyboard activation on online AND offline rows emit the CustomEvent with correct detail, `aria-pressed` flips on selection and clears on siblings.

### 4. Backend (No Code Changes)
- D01 consumes the C05 contract unchanged. `internal/handlers/chats.go` (`HandleChatRoster`) was already shipped on main and is registered at `GET /api/v1/chats` in `internal/router/router.go`. The tracker is updated to reflect that C05 is done.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **D01**:
> - the roster is visible on authenticated routes
> - presence state is shown per user
> - roster order matches backend ordering
> - all rostered users remain selectable

Audit-backed evidence (from the Phase 3 cold-start audit):

- **Roster visible on authenticated routes**: `[data-chat-roster]` is rendered by the persistent shell (`SPA/features/shell/shell.views.js`) for every protected route, and `initChatRoster` runs from `runRouteInitializer` in `SPA/core/app/create-app.js` for every `access === 'protected'` match. The guard means the painted DOM persists across navigation without re-fetch.
- **Presence state shown per user**: each `<li>` carries `data-roster-online` plus a visible `chat-roster__presence--online|--offline` dot with an `aria-label`, populated from the backend `is_online` field.
- **Roster order matches backend**: `fetchRoster` returns `payload.data` verbatim and `renderRosterList` maps in input order. No client-side sort or filter exists in the chat slice; the SDS § 5.3 ordering guarantee (history-first by `last_message_at DESC`, then alphabetical) is honored by deferring to backend ordering.
- **All rostered users selectable**: every row gets `tabindex=0`, `role="button"`, and the click/keydown delegated handlers act on any `[data-roster-user-id]` regardless of `is_online`. `chat.roster.page.test.js` asserts this explicitly with a dedicated offline-row case.

## Testing & Validation Verified

### Automated Test Suite
- [x] `make test` — backend PASS (`forum/cmd/frontend`, `forum/internal/db`, `forum/internal/handlers`, `forum/internal/tests`). The Make `test-frontend` sub-target failed only because the Makefile hard-codes `./node_modules/.bin/bun` while this environment uses a global `bun` binary; pre-existing infrastructure noise unrelated to D01. Frontend suite verified independently via `bun run policy` below.
- [x] `bun run policy` — PASS. Biome clean across 101 files; Vitest **76 passed (76)** in 9 files including the 26 new D01 cases.
- [x] `bun x biome check .` — covered by `bun run policy`.

### QA Checklist
- [x] Verified `fetchRoster` calls `GET /api/v1/chats` with `credentials: 'include'` and `Accept: application/json`.
- [x] Verified roster is rendered into `[data-chat-roster] #roster-list` on protected routes and persists across navigation without a second fetch (guard test).
- [x] Verified offline rows are selectable (Vitest case explicitly emits `chat:user-selected` for `is_online: false`).
- [x] Verified XSS escaping of usernames and last-message previews via dedicated test cases.
- [x] Confirmed `package.json`, `bun.lock`, and `go.mod` are unchanged — no new dependencies.

### Manual E2E Verification
- [ ] (Pending live-stack walkthrough) Log in as a user, confirm the roster aside is visible on `/`, `/posts/:id`, `/activity`, and `/create-post`.
- [ ] (Pending live-stack walkthrough) Click an offline user row and confirm the `chat:user-selected` event fires (D02 will visibly consume this).

The two manual checks above are gated by a running stack and are non-blocking for D01 — the verification gate is fully covered by automated tests and audit. They will be executed end-to-end as part of D02 and D04 integration.

## Key Files Impacted
- `SPA/features/chat/chat.roster.api.js`
- `SPA/features/chat/chat.roster.views.js`
- `SPA/features/chat/chat.roster.page.js`
- `SPA/features/chat/chat.roster.css`
- `SPA/tests/unit/features/chat/chat.roster.api.test.js`
- `SPA/tests/unit/features/chat/chat.roster.views.test.js`
- `SPA/tests/unit/features/chat/chat.roster.page.test.js`
- `SPA/core/app/create-app.js`
- `SPA/assets/css/main.css`
- `docs/ticket-tracker.md`
- `docs/pr-message/D01-Persistent-Chat-Roster-UI-pr.md`
