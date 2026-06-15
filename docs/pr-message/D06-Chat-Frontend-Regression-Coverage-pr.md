# D06: Chat Frontend Regression Coverage
<!-- Filename: docs/pr-message/D06-Chat-Frontend-Regression-Coverage-pr.md -->

This PR closes the regression-coverage gate for the realtime chat UI delivered across D01–D04. Those feature tickets each shipped focused unit tests that exercise one module in isolation (roster **or** conversation). D06 adds the missing **integrated** layer: a single suite that wires `initChatRoster` and `initChatConversation` onto one shared document, so the behaviours that only emerge from their interaction — one inbound `dm.message` reordering the roster *and* rendering live in the open thread, a roster click driving the conversation, presence frames repainting while a thread is open — are explicitly protected against regression. No production code changed; this is a pure test ticket.

## Summary of Changes

### 1. Integrated chat regression suite (`SPA/tests/integration/chat_regression.test.mjs`)
- **Shared-document harness**: both chat modules are booted on one fake `documentRef` (the same wiring `create-app` performs at authenticated boot), so a single dispatched WebSocket frame fans out to every registered handler — exactly as DOM event propagation does in the browser. The fake DOM follows the project's established node-environment stub style (Vitest runs `environment: 'node'`, no jsdom): real view modules render the HTML strings we assert on, while imperative nodes (scroll container, message list, error banner) are recording stubs.
- **Offline composer behaviour**: selecting an offline user disables the composer and shows the offline note while keeping prior history readable; selecting an online user re-enables it.
- **Empty-state conversations**: selecting a user with no history renders the valid empty state.
- **Presence rendering & selectability**: a `presence.snapshot` repaints per-user online/offline state across the roster with no refetch; a `presence.update` flips a single user; an **offline** rostered row remains clickable and opens its conversation (offline users stay selectable).
- **Live messaging across both modules**: one `dm.message` for the open thread floats its partner to the top of the roster **and** appends live to the conversation (viewport pinned to bottom); a `dm.message` for a different thread reorders the roster but does **not** append to the open conversation; a `chat:error` surfaces the conversation error banner; composer submit emits the outbound send for the active recipient and clears the input.
- **Incremental history loading**: scrolling to the top requests the older batch with `before_id`, prepends exactly 10 items, and shifts `scrollTop` to keep the viewport anchored.

### 2. Tracker
- **`docs/ticket-tracker.md`**: D06 marked `Done` (`[x]`) with a link to this PR; summary snapshot updated (Done 32 → 33, Not Started 5 → 4).

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **D06**:
> - offline composer behavior is explicitly protected
> - offline history readability and empty-state conversations are tested
> - live message rendering and send-error handling are covered
> - presence rendering, roster reordering, and history-loading behavior are tested

Mapping:
- **offline composer** → `D06 — offline composer behaviour` (disabled composer + offline note).
- **offline history readability / empty state** → `D06 — offline composer behaviour` (history readable while offline) and `D06 — empty-state conversations`.
- **live message rendering + send-error handling** → `D06 — live messaging across roster and conversation` (live append, error banner, outbound send).
- **presence rendering / roster reordering / history-loading** → `D06 — presence rendering and selectability`, the single-frame reorder+append test, and `D06 — incremental history loading`.

This integrated suite complements (does not replace) the per-feature unit suites authored during D01–D04, which retain finer-grained coverage of the throttle/no-burst paths, view rendering, and pure roster state transitions.

## Testing & Validation Verified

### Automated Test Suite
- [x] `bun run test` (full Vitest) — **PASS**: 35 files, **312 tests** (301 baseline + 11 new). 
- [x] `bun x vitest run SPA/tests/integration/chat_regression.test.mjs` — **PASS**: 11/11 in isolation.
- [x] `bun run check` (`biome check .`) — **PASS**: 144 files checked, no diagnostics.

### QA Checklist
- [x] New suite runs green in isolation (11/11) and as part of the full run.
- [x] Full-suite re-run confirms zero regressions in the 301 pre-existing tests.
- [x] No production source modified — diff is the new test file plus tracker/PR docs.

### Automated Behavioral Verification (E2E)
- [x] Cross-module realtime behaviour (one `dm.message` → roster reorder + live conversation append) verified through the shared-document integration test rather than a duplicated unit stub.
- [N/A] No new Playwright E2E added: D06's gate is satisfied at the unit/integration layer, mirroring the D05 regression precedent; chat is also exercised structurally by the existing `A04-03` chat-layout E2E.

## Key Files Impacted
- `SPA/tests/integration/chat_regression.test.mjs` (new)
- `docs/ticket-tracker.md` (D06 marked Done)
- `docs/pr-message/D06-Chat-Frontend-Regression-Coverage-pr.md` (new)
