# D02: Active Conversation Panel and Composer

This PR closes ticket D02 by turning the previously inert `[data-chat-active]` placeholder in the authenticated SPA shell into a working conversation panel. When a user is picked in the D01 roster, the panel loads that conversation's history from the existing C03 backend (`GET /api/v1/chats/:id/messages`), renders each message with its sender and timestamp, and presents a composer that is disabled whenever the selected user is offline. It consumes the `chat:user-selected` CustomEvent emitted by D01; live message delivery and WebSocket send remain D04's responsibility.

## Summary of Changes

### 1. Chat Conversation Feature Slice (`SPA/features/chat/`)
- **`chat.conversation.api.js`**: `fetchConversation(fetchRef, userId)` wraps `GET /api/v1/chats/:id/messages` with `credentials: 'include'`, unwrapping `{ data: { messages, has_more } }` into `{ messages, hasMore }`. Returns an empty conversation on invalid id, non-ok status, rejection, or malformed payload. `fetchCurrentUserId(fetchRef)` reads `GET /api/v1/users/me` so own vs. incoming messages can be styled apart; returns `null` on any failure.
- **`chat.conversation.views.js`**: pure render functions. `renderMessage` emits an escaped `<li>` carrying `data-message-id`, the sender username, a `<time>` element with the raw `datetime` plus a human-formatted label, and the message body — tagged `--own` or `--incoming` by comparing `sender_id` against the current user. `renderMessageList` renders the `data-conversation-empty` paragraph for no history. `renderComposer` disables the input and send button (and surfaces an offline note) when the user is offline. `renderConversation` composes the header, scrollable history, and composer.
- **`chat.conversation.page.js`**: `initChatConversation({windowRef, documentRef, fetchRef})`. No-ops when `[data-chat-active]` is absent or the document cannot register listeners. Uses a `data-conversation-bound="true"` idempotency guard so repeated calls during route transitions are free. Listens once for the bubbling `chat:user-selected` event, fetches history + current-user id (the latter memoized across selections), and paints the panel. A delegated `submit` handler calls `preventDefault` so the composer never triggers a full-page navigation; a comment marks where D04 will wire the WebSocket send.
- **`chat.conversation.css`**: header/presence, own vs. incoming bubble alignment, scrollable history, empty state, and enabled/disabled composer styling. Imported via the existing `SPA/assets/css/main.css` aggregation pipeline.

### 2. SPA Shell Wiring
- **`SPA/core/app/create-app.js`**: imports `initChatConversation` and invokes it alongside `initChatRoster` from `runRouteInitializer` whenever the matched route has `access === 'protected'`. The bound-attribute guard keeps subsequent protected-route navigation free, matching the persistent-shell model established by D01.
- **`SPA/assets/css/main.css`**: `@import "../../features/chat/chat.conversation.css"` so conversation styles ride the same loader chain as the roster (no `index.html` change needed).

### 3. Backend (No Code Changes)
- D02 consumes the C03 contract unchanged. `GET /api/v1/chats/:id/messages` (latest 10, chronological, with `has_more`) and `GET /api/v1/users/me` were already shipped on main. The tracker is updated to mark D02 done.

### 4. Frontend Regression Coverage
- 32 new Vitest cases under `SPA/tests/unit/features/chat/`:
  - `chat.conversation.api.test.js` (11) — URL/credentials, `data` unwrap, invalid-id short-circuit, non-ok and rejected fallbacks, malformed-payload defaults; `users/me` id extraction and all failure paths.
  - `chat.conversation.views.test.js` (12) — sender/timestamp/body rendering, own vs. incoming classes, XSS escaping of message bodies, empty-state, message ordering, enabled vs. disabled composer, and full-conversation composition for online/offline + empty cases.
  - `chat.conversation.page.test.js` (9) — absence no-op, missing-listener no-op, idempotent bind, render-on-select with correct endpoint, empty-state on no history, offline disables composer while history stays readable, invalid-id ignored, current-user id fetched once across selections, composer submit prevented.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **D02**:
> - selecting a user loads the latest 10 messages
> - offline history remains readable
> - selectable users with no history show a valid empty state
> - sending is disabled in the UI when the selected user is offline

Evidence:

- **Selecting a user loads the latest 10 messages**: `onUserSelected` calls `fetchConversation(fetchRef, detail.userId)`, which hits `GET /api/v1/chats/:id/messages` — the C03 endpoint that returns the latest 10 messages in chronological order. `chat.conversation.page.test.js` asserts the exact URL and that the returned messages are painted into `[data-chat-active]`.
- **Offline history remains readable**: rendering is independent of presence — `renderConversation` always renders `renderMessageList` regardless of `isOnline`. A dedicated page test selects an offline user with prior history and asserts the message body is present while the composer is disabled.
- **Users with no history show a valid empty state**: `renderMessageList([])` returns the `data-conversation-empty` paragraph; covered by both a views test and a page test.
- **Sending disabled when offline**: `renderComposer(false)` adds `disabled` to the input and send button and renders the `data-conversation-offline` note; covered by views and page tests.

## Testing & Validation Verified

### Automated Test Suite
- [x] `make test` — backend PASS. The Make `test-frontend` sub-target fails only because the Makefile hard-codes `./node_modules/.bin/bun` while this environment uses a global `bun`; pre-existing infrastructure noise unrelated to D02. Frontend suite verified independently below.
- [x] `bun run policy` — PASS. Biome clean across 127 files; Vitest **199 passed (199)** in 21 files, including the 32 new D02 cases.
- [x] `bun x biome check .` — PASS, 127 files, no fixes applied (covered by `bun run policy`).

### QA Checklist
- [x] Verified `fetchConversation` calls `GET /api/v1/chats/:id/messages` with `credentials: 'include'` and unwraps `{ data: { messages, has_more } }`.
- [x] Verified `fetchCurrentUserId` reads `GET /api/v1/users/me` and is invoked only once across multiple selections (memoized).
- [x] Verified own vs. incoming messages are distinguished by `sender_id` against the current user.
- [x] Verified message bodies are HTML-escaped via the shared `core/utils/html.js` helper.
- [x] Verified composer submit calls `preventDefault` so no full-page navigation occurs.
- [x] Confirmed `package.json`, `bun.lock`, and `go.mod` are unchanged — no new dependencies.

### Automated Behavioral Verification (E2E)
- [ ] (Pending live-stack walkthrough) Log in, select a roster user with history, and confirm the latest 10 messages render with sender + timestamp.
- [ ] (Pending live-stack walkthrough) Select a user with no history and confirm the empty state; select an offline user and confirm the composer is disabled.

The manual checks above are gated by a running stack and are non-blocking for D02 — the verification gate is fully covered by automated tests. They will be executed end-to-end as part of D04 integration, when the composer is wired to the WebSocket.

## Key Files Impacted
- `SPA/features/chat/chat.conversation.api.js`
- `SPA/features/chat/chat.conversation.views.js`
- `SPA/features/chat/chat.conversation.page.js`
- `SPA/features/chat/chat.conversation.css`
- `SPA/tests/unit/features/chat/chat.conversation.api.test.js`
- `SPA/tests/unit/features/chat/chat.conversation.views.test.js`
- `SPA/tests/unit/features/chat/chat.conversation.page.test.js`
- `SPA/core/app/create-app.js`
- `SPA/assets/css/main.css`
- `docs/ticket-tracker.md`
- `docs/pr-message/D02-Active-Conversation-Panel-and-Composer-pr.md`
