# Codebase Analysis & Audit Report

**Date:** 2026-06-26
**Branch:** asmyrogl/chat-ux-and-ui-polish
**Project:** real-time-forum (Go SPA + WebSockets — 01-edu real-time-forum)
**Scope:** Full SPA review + real-time-forum backend surface — 5 parallel analysis passes

---

## Methodology

Five parallel analysis passes were executed across the codebase:
1. **Bugs & Logic Errors** — WebSocket lifecycle/presence, DM send/delivery, history pagination, roster ordering, SPA routing/boot, reaction normalization, notification polling, migration safety.
2. **Dead Code & Unused References** — orphaned legacy `web/*` bundle, retained-but-unwired OAuth, unused SPA/Go exports, stale tooling config, repo artifacts.
3. **Architecture, Boundary Violations & Guideline Drift** — `AGENTS.md` Key Design Rules, SDS contracts (§5.x event shapes, §7.x frontend model), backend layering, single-shell rule, allowed deps, track-ownership consistency, audit-question behavioral coverage.
4. **Code Quality & Security** — XSS sinks, DOM safety, forbidden tech, CSP, SQL injection, session/cookie hardening, auth gating + WS-upgrade auth, authorization, file-upload safety, concurrency, error handling.
5. **Tests & CI Gaps** — Go + Vitest unit/integration coverage, Playwright E2E mapping vs `docs/audit.md`, race detection, ticket parity, audit traceability, coverage config, CI pipeline.

Each pass was evidence-driven and read-only. The SPA was audited in full; the Go backend was audited only across the real-time-forum surface (computed from the diff against the PRD/SDS planning commit `14495d3`); the pre-existing earlier-forum CRUD code was treated as read-only context. Findings include concrete file/line references and suggested remediations. The orchestrator independently verified every High-severity and security finding against the source before consolidating; one security finding (SEC-03) was down-scoped from the originating agent's impact assessment after verifying the `SameSite=Lax` mitigation.

---

## Executive Summary

| Severity | Count |
|----------|-------|
| 🔴 Blocking | 0 |
| 🔴 Critical | 0 |
| 🟠 High | 4 |
| 🟡 Medium | 10 |
| 🟢 Low / Info | 16 |

**Total active findings: 30** (from 35 raw findings; 4 merged across overlapping passes, and 1 — DEAD-07 — retracted after verification). Every finding was independently re-verified against the source by a dedicated adversarial pass: **26 confirmed real, 5 corrected (overstated/mis-scoped), 1 retracted, 0 ghosts.** Severity counts above reflect the post-verification reassessments (SEC-01 and CI-03 downgraded High→Low; DEAD-01 Medium→Low). See **Finding Verification** below.

**Top risks:**
1. **The headline real-time feature is never verified end-to-end (CI-04, CI-05, CI-01).** CI runs no Playwright E2E (CI-04) and no `go test -race` (CI-01), and the sole E2E spec covers only Track A — there is zero browser-level proof of real-time DM delivery, presence, roster ordering, or throttled scroll (CI-05). Real-time regressions and data races can merge green.
2. **The open conversation composer is frozen to live presence (BUG-01).** It never subscribes to `presence.update`/`presence.snapshot`, so a recipient going offline mid-thread leaves the composer enabled (sends silently fail with `RECIPIENT_OFFLINE`), and a recipient coming online leaves it disabled until reselection.
3. **WebSocket subsystem hardening gaps (SEC-04, SEC-03).** The per-connection read/write goroutines run with no `recover()` outside the Recoverer middleware — a single panic on any connection (reachable from untrusted `dm.send` input) crashes the whole backend — and the upgrader's `CheckOrigin` accepts every origin (a cross-site-WS weakness currently mitigated by the `SameSite=Lax` session cookie).
4. **Missing security headers / CSP (SEC-02).** The frontend serves no Content-Security-Policy, `X-Frame-Options`, or `X-Content-Type-Options`, so the app is clickjackable and a future escaping regression in the many `innerHTML` sinks would have no secondary mitigation. (User content is currently escaped via `escapeHTML` at the sinks, so there is no active XSS today.)
5. **Retired/dead surface still live (DEAD-02, DEAD-01).** Four OAuth routes remain registered in the RTF-modified router despite SDS §6.3 retiring OAuth, and the orphaned pre-SPA `web/templates/*` + `web/static/{js,css}` bundle is still present as static assets.

---

## Finding Verification

After consolidation, every finding was put through a second, independent **adversarial verification pass** — one skeptic per finding, each instructed to re-read the cited code and try to *refute* its claim, defaulting to "ghost" if it could not find concrete evidence. Outcome: **26 real, 5 partial (corrected below), 0 ghosts**, with **1 finding (DEAD-07) retracted**. The corrections are applied to the findings, counts, and fix order throughout this report.

| Finding | Verdict | Correction applied |
|---------|---------|--------------------|
| SEC-01 | partial | Downgraded **High → Low** and de-securitized: the inline `onclick="window.history.back()"` is a static literal with no user data interpolated and there is no CSP to break it, so it is a code-convention nit, not a security/XSS issue. |
| CI-03 | partial | Downgraded **High → Low**: `make test-backend` runs `go test ./...`, which already compiles every package *including* the `cmd/` main packages, so ordinary compile breaks already fail CI. The residual gap (build-/linker-only issues + producing the artifacts) is narrow. |
| DEAD-01 | partial | Downgraded **Medium → Low** and fix corrected: the `/static/` mount must **not** be removed — it actively serves `/static/uploads/` (DM/comment images the SPA depends on) and `favicon.ico`. Only the `web/templates/`, `web/static/js`, and `web/static/css` subtrees are dead. |
| DEAD-07 | **retracted** | All six "test-only" exports actually run in production via same-module composition; `loadEditablePost` is the real implementation that `loadPostForEdit` delegates to (not a duplicate). Nothing is dead — finding removed from the active set. |
| CI-09 | partial | Reworded: only the non-existent `dev` trigger is inert; `main` is fully active and gates every PR into `main`, so there is no coverage gap — just cosmetic config cruft. (Already Low.) |

All other 25 findings were confirmed **real** with their file/line references intact. Two were independently judged to trend *higher* than rated — **SEC-04** (remotely-triggerable full-server-crash DoS, trends High; kept Medium pending a demonstrated panic path) and **BUG-06** (trends Medium for a chat app) — and are noted in their sections.

---

## 1) Bugs & Logic Errors

### BUG-01: Open conversation composer never reacts to live presence changes ⬆ High
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: D (Tickets: D04, D07)
- `SPA/features/chat/chat.conversation.page.js` (~L77, L181-216, L360-386)
- `SPA/features/chat/chat.conversation.views.js` (~L76-145)

**Problem:** The conversation panel captures the recipient's online state once at selection time (`state.isOnline = Boolean(detail.isOnline)`, ~L194) and renders the composer enabled/disabled from it. The conversation page subscribes only to `WS_EVENTS.DM_MESSAGE` and `WS_EVENTS.ERROR` (verified: listeners at L360, L367, L384 — no `PRESENCE_UPDATE`/`PRESENCE_SNAPSHOT`). The roster page consumes presence events, but the open thread does not. Root cause: presence is a per-selection snapshot, not a live-bound value.

**Impact:** If the recipient goes offline while the thread is open, the composer stays enabled and a Send silently fails (backend rejects with `RECIPIENT_OFFLINE`, surfaced only as a small banner). If the recipient comes online, the composer stays disabled and the user cannot message a now-online peer until they reselect the row. Both are incorrect realtime transitions for the product's core feature.

**Fix:** Subscribe the conversation slice to presence events and re-derive the composer when the open recipient's presence flips:
```js
function updateComposerPresence(isOnline) {
  state.isOnline = isOnline;
  for (const sel of ['[data-conversation-input]','[data-conversation-send]','[data-conversation-attach]','[data-conversation-image-input]']) {
    const el = activeRoot.querySelector(sel); if (!el) continue;
    isOnline ? el.removeAttribute('disabled') : el.setAttribute('disabled','');
  }
}
documentRef.addEventListener(WS_EVENTS.PRESENCE_UPDATE, (e) => {
  const d = e?.detail; if (!d || !state.recipientId) return;
  if (Number(d.userId) === state.recipientId) updateComposerPresence(Boolean(d.isOnline));
});
documentRef.addEventListener(WS_EVENTS.PRESENCE_SNAPSHOT, (e) => {
  if (!state.recipientId) return;
  const online = (e?.detail?.users ?? []).some(u => Number(u?.user_id) === state.recipientId && u?.is_online);
  updateComposerPresence(online);
});
```

**Tests to add:** Vitest — render a conversation with recipient online, dispatch `presence-update {userId: recipient, isOnline:false}`, assert input/send/attach disable and the offline note appears; dispatch `isOnline:true` and assert they re-enable; add a `presence.snapshot` test excluding the open recipient and assert the composer disables.

---

### BUG-02: Scroll-up history pagination permanently dead-ends on a transient fetch failure ⬆ Medium
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: C (Tickets: C03, D04)
- `SPA/features/chat/chat.conversation.api.js` (~L7-36)
- `SPA/features/chat/chat.conversation.page.js` (~L125-157)

**Problem:** `fetchConversation` swallows all errors (network, 5xx, malformed JSON) and returns `{ messages: [], hasMore: false }` (verified: `EMPTY_CONVERSATION` returned on `!response.ok` and in `catch`). `loadOlderHistory` then takes the empty branch and sets `state.hasMore = false`; the guard `!state.hasMore` blocks every future load for that conversation. An empty result is indistinguishable from a transient failure, and a failure is treated as a terminal "no more history" signal.

**Impact:** A single momentary network blip or transient backend 500 while scrolling up permanently disables further history loading for that open conversation, even though older messages exist. The user must close and reopen the thread to recover — silent data-availability loss on the chat history hot path.

**Fix:** Distinguish failure from genuine end-of-history. Expose an `ok` flag from `fetchConversation` (or throw), and do not clear `hasMore` on failure:
```js
let older;
try { older = await fetchConversation(fetchRef, state.userId, { beforeId: state.oldestId }); }
catch { state.isLoading = false; return; } // keep hasMore so a later scroll retries
if (conversationState !== state) return;
if (older.messages.length > 0) { /* prepend */ state.hasMore = older.hasMore; }
else if (older.ok) { state.hasMore = false; } // only terminal on a real empty page
state.isLoading = false;
```

**Tests to add:** Vitest — mock `fetchConversation` to reject once (or return `ok:false`) on the first older-history load; assert `state.hasMore` stays true and a subsequent scroll re-triggers `loadOlderHistory`; then mock a real empty page and assert `hasMore` becomes false.

---

### BUG-03: `loadOlderHistory` leaves orphaned `state.isLoading=true` when switching conversations mid-flight ⬆ Low
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: C (Tickets: C03, D04)
- `SPA/features/chat/chat.conversation.page.js` (~L125-157)

**Problem:** `loadOlderHistory` captures `const state = conversationState`, sets `state.isLoading = true`, then awaits the fetch. If the user selects a different conversation mid-flight, `onUserSelected` replaces `conversationState` with a fresh object; the early-return guard `if (conversationState !== state) return;` bails before resetting the captured state's `isLoading`. Benign today (orphaned state is GC'd) but a latent foot-gun if conversation state is ever cached/restored.

**Impact:** No user-visible failure today. Latent correctness risk if conversation state is later cached/restored rather than recreated — the restored object would be stuck loading forever.

**Fix:** Reset the captured state's flag in the abandon path before returning:
```js
if (conversationState !== state) {
  state.isLoading = false; // release orphaned state defensively
  return;
}
```

**Tests to add:** Vitest — start `loadOlderHistory`, switch `conversationState` before the fetch resolves, resolve the fetch, and assert the original state object has `isLoading === false`.

---

### BUG-04: Newly-registered online users never appear in an already-loaded roster ⬆ Medium
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: D (Tickets: D04, C05)
- `SPA/features/chat/chat.roster.logic.js` (~L27-38)
- `SPA/features/chat/chat.roster.page.js` (~L142-149, L166-169)
- `internal/handlers/chats.go` (~L40-72)

**Problem:** The roster is built once per mount from `GET /api/v1/chats`. Live updates arrive via `presence.update`, but `setPresence` only mutates entries that already exist (its own comment: "Unknown users (not in the roster) are left untouched"), and `applySnapshot` only reconciles `is_online` on existing entries. A user who registers after the viewer's roster fetch and comes online produces neither a new row nor an online dot until the viewer reloads/re-logs in. Root cause: no path to insert a user discovered only via a presence event.

**Impact:** Violates the requirement that the online-users list reflects all registered users in realtime. A user registering during the viewer's session is invisible/unmessageable until a full roster refetch (mount/login only).

**Fix:** On a `presence.update`/`presence.snapshot` referencing a `userId` not present in `state.entries`, trigger a debounced `fetchRoster()` and re-render. A refetch avoids needing `username` on the presence wire (`presence.update` carries only `user_id`/`is_online`).

**Tests to add:** Vitest — load roster `[A]`; dispatch `presence.update {userId: B (unknown), isOnline:true}`; assert a roster refetch fires and B appears. Go — assert a second user registering+connecting is returned by `GET /api/v1/chats` for an existing viewer.

---

### BUG-05: Roster last-message preview is blank for image-only DMs ⬆ Low
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: C (Tickets: C05, C09)
- `internal/db/messages.go` (~L122-136)
- `SPA/features/chat/chat.roster.logic.js` (~L43-63)

**Problem:** `GetChatRoster` computes `last_message_preview` as `COALESCE(SUBSTR(l.body,1,?2),'')` (verified — the `latest` CTE does not even select `image_path`). Image-only DMs are now allowed (empty body + `image_path`), so when the latest message in a pair is image-only the preview is empty. The frontend `moveToTopForMessage` likewise only updates the preview when `message.body` is a non-empty string, so a live image-only DM leaves a stale/blank preview.

**Impact:** The roster row for a conversation whose latest message is an image shows no preview text (or a stale one), making it look like there was no recent activity. Minor UX regression introduced by the image-only DM feature.

**Fix:** SQL — select `l.image_path` in the `latest` CTE and fall back when body is empty:
```sql
COALESCE(NULLIF(SUBSTR(l.body,1,?2),''),
         CASE WHEN l.image_path IS NOT NULL THEN '[image]' ELSE '' END) AS last_message_preview
```
Mirror on the frontend: in `moveToTopForMessage` use `'[image]'` when body is empty but `message.image_url` is set.

**Tests to add:** Go (`messages_test.go`) — insert an image-only DM as the latest message for a pair, assert `GetChatRoster` returns a non-empty (or `[image]`) preview. Vitest — dispatch a `dm-message` with empty body + `image_url` and assert the roster preview updates rather than staying blank.

---

### BUG-06: WebSocket never reconnects after a server-side close; realtime chat dies silently for the session ⬆ Low
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: D (Tickets: D04, D09)
- `SPA/core/realtime/chat-socket.js` (~L83-129)
- `SPA/core/app/create-app.js` (~L141-166)

**Problem:** `createChatSocket` has no `close` listener (verified: only `message` and `error` listeners), so a transport close fired by the server never nulls the `socket` reference. `open()` short-circuits while `socket` is truthy, so it can never re-establish; `isOpen()` then returns false permanently and every `dm.send` fails. `onSendMessage` does surface `NOT_CONNECTED` (good), but there is no recovery path short of a full reload. Automatic reconnect/backoff is explicitly out of D04 scope, but the missing close-handler turns a recoverable drop into a permanently dead socket.

**Impact:** After any server-side socket drop (deploy, idle timeout, network change), realtime chat is silently dead for the session: presence stops and every send fails with `NOT_CONNECTED` until the user reloads.

**Fix:** Add a `close` listener that nulls `socket` so `open()` can re-establish (optionally with bounded backoff):
```js
socket.addEventListener('close', () => { socket = null; /* optional: scheduleReconnect() */ });
```

**Tests to add:** Vitest — open the socket, fire a `close` event on the mock WebSocket, then call `open()` and assert a new socket is constructed (current behavior short-circuits and never reconnects).

---

## 2) Dead Code & Unused References

### DEAD-01: Orphaned legacy frontend bundle (`web/templates/*`, `web/static/js/*`, `web/static/css/*`) left after the SPA migration ⬆ Low
**Origin:** 2. Dead Code & Unused References (also found by 3. Architecture & Guideline Drift)
**Files:** Ownership: A (Tickets: A10, A02); some files B (B04, B05)
- `web/templates/{home,login,register,create-post,edit-post,view-post}.html`
- `web/static/js/*` (26 files), `web/static/css/*` (13 files), `web/static/partials/*`
- `cmd/frontend/routes.go` (~L19-31)

**Problem:** After the SPA migration, `SPA/index.html` loads only `/assets/css/main.css` and `/main.js`. No served HTML references anything under `web/static` or `web/templates` (grep across `SPA/`, `cmd/`, `internal/` shows zero references to `web/templates`); the only consumers of `web/static/js/*` are the `web/templates/*.html` files, which are themselves never served — the frontend mux has no template route, only a `/static/` `FileServer` and the SPA catch-all. The migration deleted some legacy assets (activity template/CSS) but left 6 templates, 26 JS, 13 CSS, and 3 partials orphaned, and the `/static/` mount still publicly serves this dead JS/CSS (including the stale create-post hard-reload to `/activity`).

**Impact:** Dead code bloats the repo and the served surface, can confuse maintainers, duplicates logic that has diverged from the SPA, and presents a stale attack surface. Contradicts the SPA-only delivery model and SDS §7.4.1's deletion precedent.

**Fix (corrected after verification):** Delete only the genuinely-orphaned subtrees once confirmed unreferenced: `git rm -r web/templates web/static/js web/static/css web/static/partials`. **Do NOT remove the `/static/` `FileServer` mount in `cmd/frontend/routes.go`** — it is live and load-bearing: it serves `/static/uploads/` (DM/comment/post image uploads the SPA renders) and `favicon.ico`. Retain `web/errors/` (used by `response.go`) and `web/static/uploads/`. *(Originally rated Medium; downgraded to Low after verification — the orphaned assets are inert static files never loaded by the SPA, and the mount itself must stay.)*

**Tests to add:** Extend `SPA/tests/unit/policy/spa-import-paths.test.js` (or add a repo-hygiene test) to assert `web/templates` and `web/static/js` no longer exist, preventing reintroduction.

---

### DEAD-02: Retired OAuth handlers fully wired in the RTF-modified router despite SDS §6.3 ⬆ Medium
**Origin:** 2. Dead Code & Unused References (also found by 3. Architecture & Guideline Drift)
**Files:** Ownership: C (Tickets: C01, C11)
- `internal/router/router.go` (~L203-239)
- `internal/handlers/oauth_google.go`, `internal/handlers/oauth_github.go`, `internal/auth/oauth.go`

**Problem:** The router still registers four OAuth routes — `/auth/google`, `/auth/google/callback`, `/auth/github`, `/auth/github/callback` — wired to `GoogleStart`/`GoogleCallback`/`GithubStart`/`GithubCallback` (verified at L207-237). SDS §6.3 states "OAuth is not part of the target product design"; `AGENTS.md` says "No OAuth for the real-time forum." The SPA contains zero references to these flows. `router.go` is an RTF-touched file (in scope), so the live wiring is an RTF-surfaced concern even though the handlers predate RTF. The hard rule (OAuth must not gate forum/chat access) is satisfied — these routes are independent — but the retired surface is still reachable.

**Impact:** Live but unused auth endpoints are dead product surface and an unnecessary attack/maintenance surface (callback handlers, state cookies, external-provider config) for a session-cookie-only product. They contradict the documented auth design and can mislead auditors into thinking OAuth is supported.

**Fix:** Remove the four OAuth route registrations from `internal/router/router.go` (L203-239) and delete the now-unreferenced handlers (`oauth_google.go`, `oauth_github.go`, `oauth_helpers.go`) and `internal/auth/oauth.go`. Per SDS §8, OAuth DB tables may remain temporarily, but the HTTP surface should not be wired.

**Tests to add:** Add a router test asserting `GET /api/v1/auth/google` and `/auth/github` return 404 (route not registered), locking in OAuth retirement.

---

### DEAD-03: Stale Vitest alias mapping `/static/js` to the dead legacy bundle ⬆ Low
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: A (Tickets: A02)
- `vitest.config.ts` (~L4-9)

**Problem:** `vitest.config.ts` defines `resolve.alias['/static/js'] = ./web/static/js`. No production or test module imports through `/static/js` — the only references to that string are the config itself and the policy/e2e tests, which list `/static/js/` as a *forbidden* import fragment. The alias points exclusively at dead code (see DEAD-01).

**Impact:** Stale tooling config keeps the dead legacy path alive in the test resolver, inviting accidental imports of removed/divergent legacy modules and contradicting the policy test that bans those paths.

**Fix:** Remove the alias:
```ts
export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: ['**/*.test.{js,mjs,ts}'],
    exclude: ['SPA/tests/e2e/**'],
  },
});
```

**Tests to add:** —

---

### DEAD-04: Unused SPA exports `ACTIVITY_DEFAULTS` and `getDefaultFeedState` ⬆ Low
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: B (Tickets: B05, B01)
- `SPA/features/activity/activity.api.js` (~L171-176)
- `SPA/features/feed/feed.state.js` (~L9-15)

**Problem:** `ACTIVITY_DEFAULTS` and `getDefaultFeedState` are exported but have zero references anywhere in the repository — no production import and no test import (verified: grep finds only the `export` declarations). The activity defaults are redundant: the same `DEFAULT_PAGE`/`DEFAULT_PER_PAGE`/`ALLOWED_*` values are consumed directly elsewhere in `activity.api.js`.

**Impact:** Dead exported surface increases maintenance burden and implies an API contract nothing consumes. Note `biome.json` enables `correctness.noUnusedVariables: error`, but exported-but-unreferenced symbols escape that check.

**Fix:** Delete both exports. If a default-state helper is wanted, wire it into the feed initializer rather than leaving it unreferenced.

**Tests to add:** —

---

### DEAD-05: `internal/ws` Hub methods `SetCallbacks`, `GetOnlineUserIDs`, `GetConnectionCount` are exported test-only surface ⬆ Low
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: C (Tickets: C01, C04, C08)
- `internal/ws/connection_manager.go` (~L44-122)
- `internal/ws/connection_manager_test.go` (~L5-90)

**Problem:** `Hub` exposes `SetCallbacks`, `GetOnlineUserIDs`, and `GetConnectionCount`, but production code never calls them — the only callers are `connection_manager_test.go`. Production presence/DM paths use `Add`, `Remove`, `IsUserOnline`, `SendToUser`, `SendSnapshotToClient`, `BroadcastPresenceUpdate`. Three exported methods exist solely to be tested.

**Impact:** Exported-for-test-only methods widen the Hub's public surface, imply capabilities the running system never uses, and give false confidence that wired behavior is covered.

**Fix:** Either wire these into the production presence flow if intended (e.g. use `SetCallbacks` for the connect/disconnect hooks instead of the inline logic in `handlers/ws.go`), or unexport them (lowercase) so they are package-internal, or remove them with their tests.

**Tests to add:** If kept, add a handler-level test proving the production path invokes them so they are no longer test-only; otherwise delete the unit tests with the methods.

---

### DEAD-06: Stale Biome ignore entries for non-existent `forum-backend`/`forum-frontend` binaries ⬆ Low
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: A (Tickets: A01)
- `biome.json` (~L16-17)

**Problem:** `biome.json` `files.includes` lists `!forum-backend` and `!forum-frontend` to exclude committed binaries from linting. Those binaries are not tracked (`git ls-files` shows none) and are already gitignored. The ignore entries target files that do not exist in the repo.

**Impact:** Harmless but misleading config residue that suggests committed binaries that aren't present.

**Fix:** Remove the two stale ignore lines, relying on `useIgnoreFile: true` + `.gitignore`.

**Tests to add:** —

---

### DEAD-07: ~~Test-only SPA exports never imported by production code~~ — RETRACTED (verification: not dead) ⬆ Info
**Origin:** 2. Dead Code & Unused References
**Status:** **Retracted after verification — this was a false positive.**
**Files:** `SPA/features/chat/chat.conversation.views.js`, `SPA/features/feed/feed.state.js`, `SPA/features/post/post.api.js`, `SPA/features/notification/notification.page.js`

**Why retracted:** Independent verification found that all six named exports **do** run in production, via same-module internal composition — they are merely *exported* for unit testability without being cross-module imported: `formatTimestamp` (used in `renderMessage`), `renderMessageImage` and `renderComposer` (used in `renderConversation`), `buildFeedUrl` (called by `setFeedQueryState`), and `resolveNotificationDestination` (called by `destinationFromItem`). Critically, `loadEditablePost` is **not** a dead near-duplicate — it is the actual implementation that the production-used `loadPostForEdit` (called in `post.page.js`) delegates to. Nothing here is dead code.

**Residual (informational only):** the one-line `loadPostForEdit` → `loadEditablePost` passthrough wrapper is a trivial style nit with no runtime or maintenance risk; no action required.

---

## 3) Architecture, Boundary Violations & Guideline Drift

### ARCH-01: Mandated Proxy-based global state layer (`core/state`) is unimplemented — `.gitkeep` only ⬆ Medium
**Origin:** 3. Architecture & Guideline Drift
**Violated rule:** SDS §7.0 — "`core/state/`: Global state management using Proxy-based reactivity". *(Verification note: this Proxy-reactivity mandate is **only** in SDS §7.0; `AGENTS.md` does not mention Proxy/reactive state — it says only `core/ → State, Router, and API Logic` and "UI state comes from REST calls and WebSocket events". Cite SDS §7.0 as the sole normative source.)*
**Files:** Ownership: A (Tickets: A02, A03)
- `SPA/core/state/.gitkeep` (~L1)
- `SPA/features/feed/feed.state.js` (~L1-40)

**Problem:** SDS §7.0 and `AGENTS.md` define `core/state/` as Proxy-based reactive global state. In the actual tree `SPA/core/state/` contains only `.gitkeep` (verified: `git ls-files` returns just the `.gitkeep`), and a project-wide search for `new Proxy` over `SPA/` (excluding tests) returns zero matches (verified). State is held in per-feature ad-hoc objects (`feed.state.js`, local `state` in `chat.roster.page.js`/`chat.conversation.page.js`, and `create-app.js`'s closure). The foundational A02/A03 state-management deliverable was never built.

**Impact:** Architecture drifts from the documented design. State is duplicated and uncoordinated across slices (e.g. current-user id is resolved independently in both the roster and conversation pages), with no single source of truth for auth/session/presence. A structural/maintainability gap, not a runtime bug — the app functions with local state.

**Fix:** Either implement the mandated layer (a minimal Proxy-based store in `SPA/core/state/store.js` with `get/set/subscribe`, then migrate session/auth/presence into it) **or** update SDS §7.0 and `AGENTS.md` to remove the Proxy-reactivity requirement and the empty folder if local state is a deliberate choice. Do not leave the docs and code contradicting each other.

**Tests to add:** Vitest unit test for the store's get/set/subscribe/unsubscribe semantics, plus an integration test asserting a single source of truth for the current-user id.

---

### ARCH-02: Online-users roster is hidden whenever a conversation is open (single shared panel slot) ⬆ Low
**Origin:** 3. Architecture & Guideline Drift
**Violated rule:** SDS §7.5 — "roster is visible on every authenticated route"; `docs/audit.md` — "Is there a section designed to show online users?"
**Files:** Ownership: D (Tickets: D01, D02)
- `SPA/features/chat/chat.conversation.page.js` (~L85-93)
- `SPA/features/shell/shell.views.js` (~L34-42)

**Problem:** The shell renders the roster (`data-chat-roster`) and conversation (`data-chat-active`) as two sections sharing one `<aside>` slot. `showConversationPanel()` sets `hidden` on the roster; `showRosterPanel()` reverses it (the recent "unified chat panel" commit). While a DM thread is open, the online-users list is removed from view. The roster is present on every route but not *simultaneously visible* with an open conversation.

**Impact:** Borderline against the audit's online-users requirement. A reviewer checking "is the online-users section always visible" could fail it while a conversation is open. Functionally the roster still receives presence updates and reappears via the back arrow — a UX/interpretation risk, not a hard break.

**Fix:** On wide viewports, show the roster and the open conversation side by side instead of toggling; keep the toggle only for narrow/mobile widths. Gate the `hidden` on `rosterRoot` behind a `matchMedia('(max-width: 720px)')` check and lay the two `chat-panel` sections in a column/grid in `shell.css` so both render together on desktop.

**Tests to add:** Vitest DOM test asserting opening a conversation on a wide layout leaves `[data-chat-roster]` without `hidden`, while it is hidden on a narrow layout.

---

### ARCH-03: SPA shell depends on an external Google Fonts CDN at runtime ⬆ Low
**Origin:** 3. Architecture & Guideline Drift
**Violated rule:** `AGENTS.md` — "Don't use external CSS frameworks — vanilla CSS only"; the self-hosted vanilla-frontend intent (web fonts are not explicitly banned, hence Low).
**Files:** Ownership: A (Tickets: A10)
- `SPA/index.html` (~L11-16)

**Problem:** The single HTML shell preconnects to and loads stylesheets from `fonts.googleapis.com`/`fonts.gstatic.com` (Fraunces, Inter, Outfit) — verified. The frontend is otherwise designed to be fully self-hosted by the Go frontend server.

**Impact:** Offline or network-restricted environments (including a likely sandboxed grading environment) render with fallback fonts and incur external requests the project otherwise avoids. Weakens the self-contained posture. Cosmetic, not functional.

**Fix:** Self-host the `woff2` files under `SPA/assets/fonts/` and declare them with `@font-face` in `SPA/assets/css/tokens.css` (or `base.css`), then remove the three font `<link>` tags from `index.html`. (Coordinate with SEC-02's CSP `font-src`.)

**Tests to add:** —

---

## 4) Code Quality & Security

### SEC-01: Inline `onclick` handler in profile error markup (convention violation, not a security issue) ⬆ Low
**Origin:** 4. Code Quality & Security
**Files:** Ownership: A (Tickets: A07)
- `SPA/features/profile/profile.page.js` (~L72-78)

**Problem:** The profile "User not found" error state is injected via `innerHTML` and contains an inline handler: `<button class="auth-button" onclick="window.history.back()">Go Back</button>` (verified — the only inline handler in the entire SPA). `AGENTS.md` mandates event delegation/`addEventListener` for all dynamic DOM.

**Impact (corrected after verification):** This is a **code-convention/style nit, not a security or functional bug.** The handler is a static literal with **no user/profile data interpolated**, so there is no XSS vector, and there is currently no CSP to break it, so the button works today. The forward-looking concern is that a future strict CSP (see SEC-02) without `unsafe-inline` would make this button silently dead — so it is worth fixing alongside any CSP rollout. *(Originally rated High/security; downgraded to Low after verification.)*

**Violated rule:** `AGENTS.md` — "Event delegation for dynamic DOM elements."

**Fix:** Render the button without the inline handler and bind via `addEventListener`:
```js
contentEl.innerHTML = `
  <div class="profile-error">
    <h3>User not found</h3>
    <p>The profile you are looking for does not exist or has been removed.</p>
    <button class="auth-button" data-profile-back type="button">Go Back</button>
  </div>`;
contentEl.querySelector('[data-profile-back]')
  ?.addEventListener('click', () => windowRef.history.back());
```
(pass `windowRef` into `renderProfileData`).

**Tests to add:** Vitest — render the "not found" branch into a jsdom container and assert the HTML contains no `onclick` attribute and a click on `[data-profile-back]` invokes a mocked `history.back()`.

---

### SEC-02: No Content-Security-Policy on SPA shell or frontend server ⬆ Medium
**Origin:** 4. Code Quality & Security
**Files:** Ownership: D (Tickets: D09, D04)
- `SPA/index.html` (~L1-39)
- `cmd/frontend/routes.go` (~L78-83)

**Problem:** There is no CSP anywhere: `index.html` has no CSP `<meta http-equiv>` (verified), and the frontend server (which sets `Cache-Control`/`Pragma`/`Expires` on the shell) emits no CSP/`X-Frame-Options`/`X-Content-Type-Options` headers. The app renders user-generated content (post/comment/DM bodies, usernames) through many `innerHTML` sinks (the `escapeHTML` helper is used 76× across the SPA). Those paths are currently escaped, but CSP is the standard defense-in-depth layer.

**Impact:** Loss of defense-in-depth: a single future escaping regression in any `innerHTML` site becomes fully exploitable stored XSS with no secondary mitigation. Absence of `frame-ancestors`/`X-Frame-Options` also leaves the SPA clickjackable.

**Fix:** Emit CSP from the frontend server when serving the shell (must allow the Google Fonts origins used today, or remove them per ARCH-03):
```go
w.Header().Set("Content-Security-Policy",
  "default-src 'self'; img-src 'self' data:; "+
  "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
  "font-src https://fonts.gstatic.com; "+
  "connect-src 'self' ws: wss:; frame-ancestors 'none'")
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
```
After adding CSP, fix SEC-01 so the policy can omit `unsafe-inline` for scripts.

**Tests to add:** `cmd/frontend/server_test.go` — request the SPA shell route and assert the response carries a CSP header (and `X-Frame-Options`).

---

### SEC-03: WebSocket upgrader `CheckOrigin` accepts every origin ⬆ Medium
**Origin:** 4. Code Quality & Security
**Files:** Ownership: C (Tickets: C01)
- `internal/handlers/ws.go` (~L19-25)

**Problem:** The gorilla/websocket upgrader sets `CheckOrigin: func(r *http.Request) bool { return true }` (verified), unconditionally accepting cross-origin upgrades. The WS upgrade authenticates via the ambient `session_token` cookie, so origin validation is the missing layer that would prevent a malicious page from opening an authenticated socket. The REST API is protected by single-origin CORS; the WS endpoint bypasses that.

**Impact (verified/down-scoped):** This is the classic Cross-Site WebSocket Hijacking shape, **but the practical exploitability is limited** because the session cookie is `SameSite=Lax` (verified in `users.go`/`auth.go`): modern browsers do not attach `Lax` cookies to cross-site WebSocket handshakes (they are not top-level navigations), so a third-party page generally cannot ride the victim's session. The always-`true` `CheckOrigin` remains a real defense-in-depth weakness — it relies entirely on `SameSite` enforcement, offers no protection for non-browser clients or older/edge-case `SameSite` behavior, and is bad practice for a cookie-authenticated upgrade. Hence Medium, not Critical.

**Violated rule:** Audit requirement — "WS upgrade rejects invalid/expired sessions; no guest access" (the upgrade should not be cross-site forgeable, defense-in-depth alongside `SameSite`).

**Fix:** Validate the `Origin` header against the configured frontend origin:
```go
var allowedWSOrigin = func() string {
  if v := os.Getenv("FRONTEND_URL"); v != "" { return v }
  return "http://localhost:3000"
}()

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024, WriteBufferSize: 1024,
  CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    return origin == "" || origin == allowedWSOrigin // allow empty for same-origin/native clients
  },
}
```

**Tests to add:** Go test against an `httptest` server — an upgrade with a foreign `Origin` is rejected (non-101), while the configured frontend origin (or no `Origin`) upgrades successfully.

---

### SEC-04: WebSocket read/write goroutines run without panic recovery ⬆ Medium
**Origin:** 4. Code Quality & Security
**Files:** Ownership: C (Tickets: C01, C06)
- `internal/handlers/ws.go` (~L88-133)
- `internal/middleware/middleware.go` (~L53-72)

**Problem:** The Recoverer middleware only wraps the HTTP handler chain. After `HandleWebSocket` upgrades, it spawns `go h.readPump(...)` and `go h.writePump(...)` (verified L88-89); these goroutines are detached from the middleware chain and have no `recover()` (verified — `readPump`'s defer only does `Remove`/`close`/`Conn.Close`). A panic inside `readPump → handleDMSend → json/db` would not be caught and would crash the entire backend process, taking down all connected users.

**Impact:** A single panic on any connection's goroutine crashes the whole server (all sessions, all chat, all REST) rather than dropping one connection — a denial-of-service/availability risk for the realtime subsystem that Recoverer is meant to prevent for HTTP. *(Verification flagged this as trending **High** because the panic path is reachable from untrusted `dm.send` input; it is kept Medium pending a demonstrated panic trigger — there is no known panic in the current DM path, so it is a robustness/defense-in-depth gap rather than a live crash.)*

**Violated rule:** `AGENTS.md` middleware request flow lists "Recoverer" as the panic-safety layer; the WS goroutines escape it. Audit scope — "Can an exception crash the WS read/write loop?"

**Fix:** Add a `defer/recover` at the top of each pump so a panic terminates only that connection:
```go
func (h *WsHandler) readPump(userID int64, c *ws.Client) {
  defer func() {
    if rec := recover(); rec != nil {
      log.Printf("PANIC in readPump (user %d): %v\n%s", userID, rec, debug.Stack())
    }
    remaining := h.hub.Remove(userID, c)
    if remaining == 0 { h.hub.BroadcastPresenceUpdate(userID, false) }
    close(c.Send)
    c.Conn.Close()
  }()
  // ...
}
```
(mirror in `writePump`).

**Tests to add:** Go test — inject a handler/db path that panics for a crafted `dm.send` frame and assert the process survives (the connection is closed and the user removed from the hub) rather than the test binary aborting.

---

### SEC-05: Session cookie not marked `Secure` ⬆ Low
**Origin:** 4. Code Quality & Security
**Files:** Ownership: C (Tickets: C01)
- `internal/handlers/users.go` (~L135-141, L193, L234)
- `internal/middleware/auth.go` (~L66-75)
- `internal/handlers/oauth_helpers.go` (~L69-76) — fifth session-cookie site (same `HttpOnly`+`SameSite=Lax`, no `Secure`; surfaced during verification)

**Problem:** All `session_token` cookies (register, login, logout, `clearSessionCookie`, and the OAuth login flow) are set with `HttpOnly:true` and `SameSite:Lax` but never `Secure`. Over plain HTTP the cookie is sent in clear; on a TLS deployment, the absence of `Secure` means it can still be transmitted over an http:// downgrade. The cookie is the sole bearer of the authenticated session.

**Impact:** On any TLS deployment, an attacker able to force/observe an http:// request to the domain can capture the session cookie (session hijacking). Low for the current localhost/dev posture; a real gap for production-like deployments.

**Violated rule:** Audit scope — "Session cookies — `HttpOnly`, `Secure` where applicable, `SameSite`."

**Fix:** Add `Secure` conditionally (driven by `r.TLS != nil` or an `APP_SECURE_COOKIES` env) to every `SetCookie` call, via a shared `sessionCookie(...)` helper, so local HTTP dev is unaffected while TLS deployments get `Secure`.

**Tests to add:** Go handler test — when TLS / the secure flag is enabled, assert `Set-Cookie` includes `Secure`; when not, assert it is omitted.

---

### SEC-06: No global `unhandledrejection`/`error` handler in the SPA ⬆ Low
**Origin:** 4. Code Quality & Security
**Files:** Ownership: D (Tickets: D04)
- `SPA/index.html` (~L36-38)
- `SPA/main.js` (bootstrap)

**Problem:** A repository-wide grep for `unhandledrejection`, `window.onerror`, and a global error listener returns no results in `SPA/`. Many async flows (`fetchConversation`, `uploadDMImage`, `loadOlderHistory`, `fillProfileFields`) are invoked fire-and-forget; a rejected promise silently disappears with no user-facing surfacing and no logging hook.

**Impact:** Async failures can leave the UI in a partially-updated state with no feedback, and there is no global place to log/telemetry them. The audit explicitly asks whether an `unhandledrejection` handler is installed; it is not.

**Fix:** Install global handlers during bootstrap (`main.js`):
```js
window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason);
  // optionally surface a non-blocking toast/banner
});
window.addEventListener('error', (event) => {
  console.error('Uncaught error:', event.error ?? event.message);
});
```

**Tests to add:** Vitest — after bootstrapping the app in jsdom, dispatch an `unhandledrejection` event and assert the registered handler runs.

---

## 5) Tests & CI Gaps

### CI-01: CI never runs `go test -race`; SDS §10.4 race-detection requirement unmet ⬆ High
**Origin:** 5. Tests & CI Gaps (consolidates two facets: pipeline-flag gap + concurrency integration tests not under `-race`)
**Files:** Ownership: C (Tickets: C08, C04, C01)
- `.github/workflows/ci.yml` (~L33-37)
- `Makefile` (~L117-119)
- `internal/ws/connection_manager.go` (~L30-117)
- `internal/tests/presence_test.go` (~L85-660), `internal/tests/dm_send_test.go`

**Problem:** The connection manager guards its presence map with a `sync.RWMutex` and the DM/presence subsystem is goroutine+channel based. SDS §10.4 requires "goroutine/channel patterns do not introduce data races (verified with `-race`)." Neither the Makefile `test-backend` (`go test ./...`) nor CI (`make test-backend`) ever passes `-race` (verified). Critically, the presence *integration* tests (`presence_test.go`, `dm_send_test.go`) **do** drive concurrent multi-connection WebSocket flows — the most race-prone surface — yet also run without `-race`, so any genuine race goes unreported despite good functional coverage.

**Impact:** Concurrency bugs (data races on the presence/connection map, broadcast-vs-disconnect races) pass CI undetected and can reach production as flaky presence/online state or crashes under concurrent multi-tab load.

**Violated rule:** SDS §10.4 race-verification requirement; the audit prompt also requires `go test -race` for the WS/presence code.

**Fix:** Add a race-enabled target covering the concurrent surfaces and run it in CI:
```makefile
test-backend-race:
	@mkdir -p $(TEST_CACHE_DIR) $(TEST_TMP_DIR)
	@GOCACHE=$(TEST_CACHE_DIR) GOTMPDIR=$(TEST_TMP_DIR) \
	  go test -race ./internal/ws/... ./internal/handlers/... ./internal/tests/...
```
```yaml
      - name: Run Backend Race Tests
        run: make test-backend-race
```

**Tests to add:** Wire `make test-backend-race` into the CI quality-gate (minimum: `internal/ws`, `internal/handlers/ws.go`, `internal/tests/presence_test.go`). See CI-02 for the missing concurrent unit test that makes `-race` meaningful at the unit level.

---

### CI-02: No concurrent test exercises the mutex-guarded connection manager, so `-race` cannot catch a presence-map race at the unit level ⬆ Medium
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: C (Tickets: C08, C01)
- `internal/ws/connection_manager_test.go` (~L1-130)
- `internal/ws/connection_manager.go` (~L55-180)

**Problem:** `connection_manager_test.go` contains only sequential single-goroutine tests (no `go func`, no `sync.WaitGroup`). Even with CI-01 fixed, the race detector only instruments paths actually executed concurrently — purely sequential unit tests report clean even if the locking is wrong. (Integration-level concurrency exists in `presence_test.go`; this gap is specifically the unit-level guarantee for the Hub.)

**Impact:** The mutex protecting the presence/connection map is untested for correctness under contention at the unit level; a missing/incorrect lock would not be caught by `-race` from these tests.

**Violated rule:** SDS §10.4 race verification is only meaningful with concurrent access; the connection-manager contract (last-disconnect cleanup, per-user multi-tab set) is concurrency-sensitive.

**Fix:** Add a concurrency stress test in `connection_manager_test.go`:
```go
func TestConnectionManager_ConcurrentAddRemove(t *testing.T) {
  h := NewConnectionManager()
  var wg sync.WaitGroup
  for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(uid int64) {
      defer wg.Done()
      c := newTestConn()
      h.Add(uid, c); _ = h.GetOnlineUserIDs(); _ = h.GetConnectionCount(uid); h.Remove(uid, c)
    }(int64(i % 10))
  }
  wg.Wait()
}
```
Run under `go test -race -count=1 ./internal/ws/...`.

**Tests to add:** `TestConnectionManager_ConcurrentAddRemove` plus a concurrent same-user multi-tab variant, executed under `-race`.

---

### CI-03: CI has no explicit `make build` / artifact step ⬆ Low
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (Tickets: A01)
- `.github/workflows/ci.yml` (~L20-37)
- `Makefile` (~L27-37)

**Problem (corrected after verification):** The CI quality-gate never runs `make build` (verified). However, the originally-claimed impact was **overstated**: `make test-backend` runs `go test ./...`, which compiles *every* package in the module **including** the `cmd/backend`, `cmd/frontend`, and `cmd/qa-seed` main packages (they carry no build tags). An ordinary compile/type error in any of those already fails `go test ./...` and therefore fails CI. The genuine residual gap is narrow: build-/linker-only problems that `go test` does not surface (rare for this pure-Go app), and the fact that the deployable binaries themselves are never produced or smoke-checked in CI.

**Impact:** Small. Practically all compile breaks are already caught by `go test ./...`; adding `make build` is cheap insurance that also produces the actual artifacts the grader runs via `make run`. *(Originally rated High; downgraded to Low after verification.)*

**Violated rule:** CI quality-gate intent — defense-in-depth; produce and smoke-check the deployable artifacts.

**Fix:** Add an explicit build step before tests (cheap, fast):
```yaml
      - name: Build
        run: make build   # exercises build-backend + build-frontend, produces the deployable artifacts
```

**Tests to add:** Add `make build` as a CI step.

---

### CI-04: CI never runs Playwright E2E (`make test-e2e`); full-stack/real-time regressions can pass CI ⬆ High
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: D (Tickets: D06, D07)
- `.github/workflows/ci.yml` (~L30-37)
- `Makefile` (~L99-105)

**Problem:** `make test` = `test-backend` + `test-frontend` + `test-e2e`, but CI runs only `test-backend` and `test-frontend` (verified). The Playwright suite — the only full-stack frontend+backend+WebSocket verification (SDS §10.2.2) — never runs in CI. (`test-e2e` self-skips when local TCP listeners are unavailable, so it must run on a runner that allows port binding; GitHub-hosted ubuntu does.)

**Impact:** Routing, auth-gating, proxy, and any real browser-level regressions merge with green CI. Combined with CI-05, real-time DM behavior is never exercised end-to-end anywhere.

**Violated rule:** SDS §10.2.2 — "E2E tests use Playwright to verify the full stack … Run with `make test-e2e`."

**Fix:** Add an E2E step (GitHub-hosted runners permit localhost binding; `make deps` already runs `bun x playwright install chromium`):
```yaml
      - name: Run E2E Tests
        run: make test-e2e
```
Upload `playwright-report` on failure.

**Tests to add:** Wire `make test-e2e` into CI as a dedicated step/job (it boots both servers via the `webServer` config).

---

### CI-05: Playwright E2E covers only Track A; zero E2E for chat/DM/presence/roster/history ⬆ High
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: D (Tickets: D06, D04, D03, D07)
- `SPA/tests/e2e/tickets.test.js` (~L143-712)

**Problem:** The sole Playwright spec (21 tests) verifies only A02-A10: shell serving, SPA routing, back/forward, deep-links, auth gating, logout, profile fields, static assets, `/api` proxy. Its only chat references assert that the `[data-chat-roster]`/`[data-chat-active]` *containers stay visible/stable in layout* — they never send a DM, never assert real-time delivery to a second session, never test presence transitions, roster reordering by last message, the disabled-composer-for-offline state, or throttled history scroll. `docs/audit.md` explicitly requires two-browser real-time DM delivery and scroll-event-throttled history loading; none of these has an E2E proof.

**Impact:** The headline real-time-forum feature has no end-to-end automated verification. The audit's two-browser DM/notification question is unprovable by the current suite; WebSocket/presence/scroll-throttle regressions would not be caught at the integration boundary.

**Violated rule:** SDS §10.2.2 multi-step user journeys; `docs/audit.md` real-time DM and scroll-throttle questions; D06 verification gate.

**Fix:** Add a multi-context Playwright spec (`SPA/tests/e2e/chat.test.js`) using two browser contexts: register+login user A and user B, open the A→B conversation, send a message, assert it appears in B's page **without reload** (`expect(locator).toHaveText`, no `page.reload`), assert a presence/notification indicator, seed >20 messages, scroll up and assert the next batch of 10 loads with the scroll throttled (bound the number of history fetches). Use state-driven waits (`expect.poll`/`waitForResponse`), never `waitForTimeout`.

**Tests to add:** `SPA/tests/e2e/chat.test.js` — two-context real-time DM delivery, presence online/offline transition, roster reorder by last message, disabled composer for offline user, throttled upward-scroll history load.

---

### CI-06: No current audit-traceability matrix; the only coverage report is stale and several audit questions are E2E-uncovered ⬆ Medium
**Origin:** 5. Tests & CI Gaps (consolidates the stale-report and missing-matrix facets)
**Files:** Ownership: D (Tickets: D07)
- `docs/audit-reports/audit-e2e-coverage-2026-04-18.md` (~L1-95)
- `docs/ticket-tracker.md` (~L95-135)
- `docs/audit.md` (~L1-90), `SPA/tests/e2e/tickets.test.js`

**Problem:** `audit-e2e-coverage-2026-04-18.md` is the only audit-traceability artifact and is badly out of date (verified): it marks Registration/Login as "Not Started*" and Forum/Chat/Private Messaging/History as "Not Started — Nothing" and references a foreign absolute path (`file:///home/ertval/...`), while the tracker marks all those tickets (B01-B08, C05, C06, D01-D04) `[x]` complete. Mapping `audit.md` questions to proving tests reveals specific E2E-uncovered questions: "Did the other user receive a notification?", "Did the other user receive the message in real time, without refreshing?" (only Vitest with a mocked socket), roster ordering by last message / alphabetical (only Vitest unit), and scroll-throttled history loading (only `throttle.test.js`/`chat.conversation.history.test.js`, never a real browser scroll).

**Impact:** There is effectively no current audit-traceability matrix. D07 (Final Acceptance Validation — the only `[ ]` ticket) has no up-to-date map of each audit question to its proving test, so acceptance gaps (e.g. CI-05) are easy to miss. Headline behaviors are proven only by mocked unit tests; a passing Vitest suite can coexist with a broken live experience.

**Violated rule:** `docs/audit.md` acceptance-criteria coverage; D07 gate intent (comprehensive audit pass mapping every audit question).

**Fix:** Add `docs/audit-traceability.md` enumerating each `audit.md` question with its test reference and a covered-at level (unit/integration/e2e); date it and fix the `file://` paths. Flag the four E2E-uncovered rows and close them with the CI-05 two-context Playwright tests.

**Tests to add:** Traceability doc + the CI-05 two-context chat E2E tests to raise the real-time/roster/scroll questions from unit-only to E2E coverage.

---

### CI-07: Vitest has no coverage collection or thresholds configured ⬆ Low
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: D (Tickets: D05, D06)
- `vitest.config.ts` (~L1-18)
- `package.json` (~L9-16)

**Problem:** `vitest.config.ts` correctly includes `**/*.test.{js,mjs,ts}` and excludes `SPA/tests/e2e/**` (good), but defines no `coverage` block — no provider, no include/exclude, no thresholds — and the scripts run `vitest run` without `--coverage`. Frontend regression tickets D05/D06 therefore have no objective coverage measurement or floor.

**Impact:** Frontend coverage can regress without any signal; no enforced minimum and no report of which SPA modules are untested.

**Fix:** Add a coverage block and gate on it:
```ts
coverage: {
  provider: 'v8',
  include: ['SPA/**/*.{js,mjs}'],
  exclude: ['SPA/tests/**', '**/*.test.*'],
  thresholds: { statements: 70, branches: 60, functions: 70, lines: 70 },
}
```
and run `vitest run --coverage` in the `test` script / CI.

**Tests to add:** Enable `vitest run --coverage` with thresholds; config + gate change, no new cases.

---

### CI-08: CI format gate runs `biome check --write --unsafe` then `git diff --exit-code`, allowing unsafe auto-rewrites to silently pass ⬆ Medium
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (Tickets: A01)
- `.github/workflows/ci.yml` (~L23-31)
- `Makefile` (~L107-113)
- `package.json` (~L11-14)

**Problem:** CI runs `make lint` → `make format` → `git diff --exit-code`. `make format` = `go fmt ./...` + `bun run lint:fix`, and `lint:fix` = `biome check --write --unsafe .`. The `--unsafe` flag applies potentially behavior-changing auto-fixes *during the CI gate*; if those rewrites are idempotent, the `git diff` passes — meaning CI auto-applies unsafe transforms as the pass condition rather than failing the author. The check should be read-only.

**Impact:** The gate masks the difference between "already clean" and "clean only after an unsafe rewrite" and relies on unsafe transforms being idempotent; a non-idempotent unsafe fix could make the gate flaky.

**Violated rule:** CI quality-gate intent — format/lint checks should be read-only verification, not mutate-then-diff with `--unsafe`.

**Fix:** Use non-mutating checks in CI:
```yaml
      - name: Lint & Format Check
        run: |
          make lint                                   # biome check . (read-only)
          gofmt -l . | tee /tmp/gofmt.out; test ! -s /tmp/gofmt.out
```
Add a read-only `format-check` Make target (`gofmt -l`, `biome check` without `--write`) and drop the mutate-then-diff approach.

**Tests to add:** Add a read-only `format-check` target used by CI.

---

### CI-09: `ci.yml` references a non-existent `dev` branch in its triggers (cosmetic config cruft) ⬆ Low
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (Tickets: A01)
- `.github/workflows/ci.yml` (~L3-7)

**Problem (corrected after verification):** `ci.yml` triggers on push/PR to `main` and `dev`. There is **no `dev` branch** anywhere in the repo (verified: `git branch -a` lists only `main` plus per-author feature branches), so the `dev` half of both triggers never matches. The `main` trigger is fully active. **There is no coverage gap** — feature branches PR into `main`, which is in both the push and `pull_request` branch lists, so those PRs are gated. (The original "feature branches skip the gate" framing was incorrect and has been removed.)

**Impact:** Cosmetic only — a dead branch name in the trigger lists; no missed CI coverage.

**Fix:** Drop `dev` from the triggers (`branches: [ main ]`), or, if every PR regardless of base should run, remove the branch filter on `pull_request:`.

**Tests to add:** —

---

## Cross-Reference: Finding ID Mapping

| Consolidated ID | Agent 1 | Agent 2 | Agent 3 | Agent 4 | Agent 5 | Track Ownership | Description |
|----------------|---------|---------|---------|---------|---------|-----------------|-------------|
| BUG-01 | BUG-01 | — | — | — | — | D | Conversation composer frozen to live presence |
| BUG-02 | BUG-02 | — | — | — | — | C | History pagination dead-ends on transient failure |
| BUG-03 | BUG-03 | — | — | — | — | C | Orphaned `isLoading` on conversation switch |
| BUG-04 | BUG-04 | — | — | — | — | D | New online users never appear in loaded roster |
| BUG-05 | BUG-05 | — | — | — | — | C | Roster preview blank for image-only DMs |
| BUG-06 | BUG-06 | — | — | — | — | D | WS never reconnects after server-side close |
| DEAD-01 | — | DEAD-01 | ARCH-05 | — | — | A / B | Orphaned legacy `web/*` bundle still served |
| DEAD-02 | — | DEAD-02 | ARCH-04 | — | — | C | OAuth routes still wired despite retirement |
| DEAD-03 | — | DEAD-03 | — | — | — | A | Stale Vitest `/static/js` alias |
| DEAD-04 | — | DEAD-04 | — | — | — | B | Unused exports `ACTIVITY_DEFAULTS`/`getDefaultFeedState` |
| DEAD-05 | — | DEAD-05 | — | — | — | C | `ws` Hub exported test-only methods |
| DEAD-06 | — | DEAD-06 | — | — | — | A | Stale Biome ignore entries for absent binaries |
| DEAD-07 | — | DEAD-07 | — | — | — | D | ~~Test-only SPA exports~~ — **RETRACTED** (verification: exports are production-used) |
| ARCH-01 | — | — | ARCH-01 | — | — | A | `core/state` Proxy layer unimplemented |
| ARCH-02 | — | — | ARCH-02 | — | — | D | Roster hidden while conversation open |
| ARCH-03 | — | — | ARCH-03 | — | — | A | External Google Fonts CDN dependency |
| SEC-01 | — | — | — | SEC-01 | — | A | Inline `onclick` handler in profile error view |
| SEC-02 | — | — | — | SEC-02 | — | D | No Content-Security-Policy |
| SEC-03 | — | — | — | SEC-03 | — | C | WS `CheckOrigin` accepts every origin |
| SEC-04 | — | — | — | SEC-04 | — | C | WS goroutines have no panic recovery |
| SEC-05 | — | — | — | SEC-05 | — | C | Session cookie not marked `Secure` |
| SEC-06 | — | — | — | SEC-06 | — | D | No global `unhandledrejection` handler |
| CI-01 | — | — | — | — | CI-01, CI-10 | C | CI never runs `go test -race` (incl. presence integration tests) |
| CI-02 | — | — | — | — | CI-02 | C | No concurrent connection-manager unit test |
| CI-03 | — | — | — | — | CI-03 | A | CI does not run `make build` |
| CI-04 | — | — | — | — | CI-04 | D | CI never runs Playwright E2E |
| CI-05 | — | — | — | — | CI-05 | D | E2E covers only Track A; no chat/DM/presence E2E |
| CI-06 | — | — | — | — | CI-06, CI-11 | D | Stale coverage report + no audit-traceability matrix |
| CI-07 | — | — | — | — | CI-07 | D | No Vitest coverage/thresholds |
| CI-08 | — | — | — | — | CI-08 | A | CI format gate uses `--unsafe` mutate-then-diff |
| CI-09 | — | — | — | — | CI-09 | A | CI trigger branches `main`/`dev` may be inert |

---

## Recommended Fix Order

### Phase 1 — Blocking & Critical (must fix before any merge)
*No Blocking or Critical findings. The codebase has no show-stopping defects in scope.*

*(Severities reflect the post-verification reassessment: SEC-01 and CI-03 moved High→Low, DEAD-01 moved Medium→Low, DEAD-07 retracted.)*

### Phase 2 — High Severity (immediate follow-up)
1. **CI-04**: Run `make test-e2e` (Playwright) in CI. (Track D)
2. **CI-05**: Add two-context chat E2E (real-time DM delivery, presence, roster reorder, offline composer, throttled scroll). (Track D)
3. **CI-01**: Add a `-race` backend target covering `internal/ws`, `internal/handlers`, `internal/tests` and run it in CI. (Track C)
4. **BUG-01**: Wire the conversation composer to `presence.update`/`presence.snapshot` so its enabled state tracks the open recipient live. (Track D)

### Phase 3 — Medium Severity
5. **SEC-04**: Add `defer/recover` to the WS read/write pumps so one panic can't crash the server. (Track C)
6. **SEC-03**: Validate the WS `Origin` header against the configured frontend origin. (Track C)
7. **SEC-02**: Emit CSP + `X-Frame-Options`/`X-Content-Type-Options` from the frontend server. (Track D)
8. **BUG-02**: Stop treating a transient history-fetch failure as terminal `hasMore=false`. (Track C)
9. **BUG-04**: Refetch the roster (debounced) when a presence event references an unknown user. (Track D)
10. **DEAD-02**: Remove the OAuth route registrations and unreferenced OAuth handlers. (Track C)
11. **ARCH-01**: Implement the Proxy-based `core/state` store, or update SDS §7.0 to match the local-state reality. (Track A)
12. **CI-08**: Make the CI format gate read-only (`gofmt -l`, `biome check` without `--write`). (Track A)
13. **CI-06**: Regenerate the audit-traceability matrix and date it; flag E2E-uncovered questions. (Track D)
14. **CI-02**: Add a concurrent Add/Remove stress test for the connection manager under `-race`. (Track C)

### Phase 4 — Low Severity (maintenance)
15. **BUG-05**: Fall back to `[image]` for image-only DM roster previews (SQL + frontend). (Track C)
16. **BUG-06**: Add a WS `close` listener that nulls the socket so it can re-establish. (Track D)
17. **BUG-03**: Reset orphaned `isLoading` in the conversation-switch abandon path. (Track C)
18. **SEC-01**: Replace the inline `onclick` in the profile error view with `addEventListener` (convention; bundle with any CSP rollout). (Track A)
19. **SEC-05**: Set the session cookie `Secure` conditionally on TLS (all five cookie sites incl. `oauth_helpers.go`). (Track C)
20. **SEC-06**: Install global `unhandledrejection`/`error` handlers in `main.js`. (Track D)
21. **ARCH-02**: Show roster and open conversation side-by-side on wide viewports. (Track D)
22. **ARCH-03**: Self-host the web fonts and drop the Google Fonts CDN links. (Track A)
23. **DEAD-01**: Delete the orphaned `web/templates`/`web/static/{js,css}` subtrees — keep the `/static/` mount (it serves uploads). (Track A/B)
24. **DEAD-03**: Remove the stale Vitest `/static/js` alias. (Track A)
25. **DEAD-04**: Delete unused `ACTIVITY_DEFAULTS`/`getDefaultFeedState` exports. (Track B)
26. **DEAD-05**: Wire, unexport, or remove the test-only `ws.Hub` methods. (Track C)
27. **DEAD-06**: Remove stale Biome ignore entries for absent binaries. (Track A)
28. **CI-03**: Add `make build` to CI as cheap artifact insurance (compile breaks are already caught by `go test ./...`). (Track A)
29. **CI-07**: Add Vitest coverage collection + thresholds. (Track D)
30. **CI-09**: Drop the dead `dev` branch from `ci.yml` triggers. (Track A)

*(DEAD-07 retracted after verification — not a real finding; no action.)*

---

## Notes

- **Every finding was adversarially re-verified** (one independent skeptic per finding, re-reading the cited code, instructed to refute and default to "ghost" on weak evidence): **26 confirmed real, 5 corrected, 1 retracted (DEAD-07 — the named "test-only" exports are in fact production-used via same-module composition), 0 ghosts.** SEC-01 and CI-03 were downgraded High→Low and DEAD-01 Medium→Low after the verifier showed the original impact was overstated; those corrections are reflected in the findings, severity counts, and fix order above. See the **Finding Verification** section.
- **Confirmed-safe patterns (verified during the audit):**
  - **SQL injection:** all in-scope `internal/db/` queries (`messages.go`, `chats.go`, `users.go`, `migrate.go`) are parameterized — no string-concatenated SQL.
  - **XSS render paths:** user-content render sites (DMs, posts, comments, usernames, profiles, notifications) route through `escapeHTML` (used at all checked sinks); SEC-02 is about the missing defense-in-depth layer and SEC-01 about a static inline handler — neither is a confirmed live XSS.
  - **Auth:** `bcrypt` password hashing; the session cookie is `HttpOnly` + `SameSite=Lax`; the session is validated (including expiry) on every protected request **and** on the WS upgrade (`GetSessionByToken` in `ws.go`); unauthenticated WS upgrades are rejected with `401`.
  - **Presence correctness:** `presence.update` fires only on the offline↔online edge (first connect / last disconnect); a second tab gets a snapshot but no duplicate broadcast — confirmed by `internal/tests/presence_test.go`. The connection map is `sync.RWMutex`-guarded and `go vet ./...` is clean; `go test -race ./internal/ws/...` passes today (the gap is that CI never runs it — CI-01).
  - **Schema/indexes:** both SDS §4.2 `private_messages` indexes exist in **both** `forum_schema.sql` (`idx_pm_sender`, `idx_pm_recipient`) and `migrate.go` — no missing-index finding.
  - **Layering:** `internal/db/` has no `net/http`/JSON/request-parsing; handlers issue no raw SQL; the single-shell rule, ES-modules-only, and allowed-Go-deps-only constraints all hold.
- **SEC-03 caveat:** the originating security agent rated the always-`true` `CheckOrigin` as a fully exploitable Cross-Site WebSocket Hijack. The orchestrator down-scoped it to Medium after verifying the cookie is `SameSite=Lax`, which prevents modern browsers from attaching the session cookie to cross-site WS handshakes. It remains a genuine defense-in-depth weakness worth fixing, but is not directly exploitable from a third-party browser page under current `SameSite` enforcement.
- **Backend scope discipline:** all backend findings stay within the real-time-forum surface (the diff against planning commit `14495d3`). The untouched earlier-forum CRUD code was treated as read-only context; no new bugs were raised against it. DEAD-02/SEC-03/SEC-04/SEC-05 touch files that are either RTF-added (`ws.go`) or RTF-modified (`router.go`, `users.go`, `auth.go`).
- **OAuth retention:** per SDS §8, OAuth DB tables may remain temporarily; DEAD-02 recommends removing only the **HTTP route wiring and unreferenced handlers**, not necessarily the tables.
- **No Blocking/Critical defects** were found in scope: the core forum and chat flows are functionally implemented and backed by strong backend unit/integration coverage. The dominant risk class is **verification** (CI not exercising E2E or `-race`, and no real-time/two-client E2E exists) rather than broken behavior — which is exactly the gap most likely to let a future regression in the headline feature merge undetected.

---

*End of report.*
