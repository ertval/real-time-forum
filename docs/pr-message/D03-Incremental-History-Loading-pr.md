# D03: Incremental History Loading
<!-- Filename: docs/pr-message/D03-Incremental-History-Loading-pr.md -->

This PR adds scroll-up pagination to the active conversation panel introduced in D02. When the user scrolls toward the top of an open conversation, the next batch of 10 older messages is fetched from the C03 history endpoint and prepended without disturbing the user's viewport. This satisfies the audit requirement that "the last 10 messages load initially, with scroll-up pagination that is throttled/debounced and does not spam the API."

## Summary of Changes

### 1. Chat Conversation Page (`chat.conversation.page.js`)
- **Per-conversation paging state**: each user selection now tracks `userId`, the `oldestId` currently rendered, the backend `hasMore` flag, and an `isLoading` guard. Replacing this state on selection invalidates any older-history fetch still in flight for the previous conversation (prevents cross-conversation prepends).
- **Throttled scroll loading**: a throttled `scroll` listener is attached to `[data-conversation-scroll]` after every render. When the viewport is within `SCROLL_LOAD_THRESHOLD_PX` (80px) of the top and more history exists, it requests the next older batch via `before_id`.
- **Stable scroll preservation**: before prepending, the scroll geometry is captured; after the prepend the container's `scrollTop` is advanced by exactly the height delta, so the previously-visible message stays anchored in place.
- **Burst suppression**: the leading-edge throttle plus the `isLoading`/`hasMore` guards ensure a stream of scroll events produces at most one in-flight request, and paging stops cleanly once the backend reports `has_more: false`.

### 2. Chat Conversation API (`chat.conversation.api.js`)
- **`before_id` paging**: `fetchConversation` now accepts an optional `{ beforeId }` argument and appends `?before_id=<id>` to the C03 request when a positive, finite id is supplied. The no-argument signature is unchanged, so the initial latest-10 load is untouched.

### 3. Chat Conversation Views (`chat.conversation.views.js`)
- **`renderMessageItems`**: extracted an items-only renderer (no `<ul>` wrapper) so older batches can be prepended into the existing message list. `renderMessageList` now delegates to it, preserving previous output.

### 4. Shared Utilities (`core/utils/throttle.js`)
- **`throttle(fn, wait)`**: a new leading-edge throttle with a trailing replay and a `cancel()` method. Timer-based (no `Date.now`) so it is deterministic under fake timers.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **D03**:
> - older messages load in batches of 10
> - repeated scroll events do not cause burst requests
> - the viewport remains usable after prepending history

- **Batches of 10**: scrolling to the top issues `GET /api/v1/chats/{id}/messages?before_id=<oldest>` and prepends the returned batch (unit-tested with a 10-message older page).
- **No burst requests**: 25 synchronous scroll events produce exactly one older-history request (throttle + in-flight guard), and paging halts when `has_more` is false.
- **Usable viewport**: scroll position is preserved across the prepend (verified: `previousTop + heightDelta`), keeping the prior message anchored.

## Testing & Validation Verified

### Automated Test Suite
- [x] `make test-backend` — all Go packages pass (`internal/tests` 31.8s, `db`, `handlers`, `ws` ok).
- [x] `bun x vitest run` — 23 files, **212 tests passed** (includes new throttle, API `before_id`, and incremental-loading suites).
- [x] `make test-e2e` — Playwright **21 passed**.
- [x] `bun x biome check` — clean (formatting auto-applied to the page module).

### QA Checklist
- [x] New unit tests: `SPA/tests/unit/core/utils/throttle.test.js` (5 tests — leading edge, burst collapse, trailing replay, cancel).
- [x] New unit tests: `SPA/tests/unit/features/chat/chat.conversation.history.test.js` (6 tests — batch load, burst suppression, scroll preservation, empty/non-top/exhausted cases).
- [x] Extended `chat.conversation.api.test.js` with `before_id` presence/omission coverage.
- [x] Existing D01/D02 conversation and roster suites still green (no regressions).

### Automated Behavioral Verification (E2E)
- [x] Full ticket E2E regression (`make test-e2e`) passes — chat roster/history endpoints exercised, no regressions.
- [x] Infrastructure Check: `make verify-infra` — all sanity checks passed (build, process management, API proxy).

> Note: dedicated chat-flow E2E coverage is owned by **D06 (Chat Frontend Regression Coverage)**; D03 is verified through the unit suites above plus the existing E2E regression.

## Key Files Impacted
- `SPA/features/chat/chat.conversation.page.js`
- `SPA/features/chat/chat.conversation.api.js`
- `SPA/features/chat/chat.conversation.views.js`
- `SPA/core/utils/throttle.js`
- `SPA/tests/unit/core/utils/throttle.test.js`
- `SPA/tests/unit/features/chat/chat.conversation.history.test.js`
- `SPA/tests/unit/features/chat/chat.conversation.api.test.js`
