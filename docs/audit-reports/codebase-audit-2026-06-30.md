# Codebase Analysis & Audit Report

**Date:** 2026-06-30
**Branch:** main
**Project:** real-time-forum (Go SPA + WebSockets — 01-edu real-time-forum)
**Scope:** Full SPA review + real-time-forum backend surface — 5 parallel analysis passes

---

## Methodology

Five parallel analysis passes were executed across the codebase:
1. **Bugs & Logic Errors** — SPA + RTF backend runtime behavior; WS lifecycle/presence, DM rules, pagination, routing, auth gating; `go test -race ./...` executed (exit 0, no data races).
2. **Dead Code & Unused References** — orphaned legacy `web/templates`/`web/static` assets, retained OAuth, stale config/tooling, unused SPA exports.
3. **Architecture, Boundary Violations & Guideline Drift** — AGENTS.md design rules, backend layering, SPA structure, SDS contract integrity, audit-question behavioral coverage, track-ownership consistency.
4. **Code Quality & Security** — XSS sinks, CSP, SQL injection, auth/authz gating, WS-upgrade origin, file upload, concurrency, cookie flags.
5. **Tests & CI Gaps** — unit/integration/E2E coverage vs SDS §10 and audit.md, race detection, CI pipeline holes, ticket parity, traceability.

Each pass was evidence-driven and read-only. The SPA was audited in full; the Go backend was audited only across the real-time-forum surface (the pre-existing earlier-forum CRUD code was treated as read-only context, with the touched set derived from `git diff 14495d3..HEAD -- internal/ cmd/`). Findings include concrete file/line references and suggested remediations.

**Evidence baseline (verified during this audit):**
- `go test ./...` → all pass; `go test -race -count=1 ./...` → exit 0, **no data races** (incl. the `internal/ws` hub and the 269s `internal/tests` integration suite under `-race`).
- `bun run test` (Vitest) → 332 tests / 35 files pass.
- Playwright E2E (`SPA/tests/e2e/tickets.test.js`) → 21 tests, **Track A only** (A02–A10).
- Both required `private_messages` indexes present (`forum_schema.sql:240-244`).
- All SPA user-content render paths run through `escapeHTML`; all RTF backend SQL parameterized; sessions HttpOnly + validated on every protected request and WS upgrade; bcrypt password hashing; DM send fail-closed; uploads content-sniffed with server-generated UUID filenames.

---

## Executive Summary

| Severity | Count |
|----------|-------|
| 🔴 Blocking | 0 |
| 🔴 Critical | 3 |
| 🟠 High | 7 |
| 🟡 Medium | 14 |
| 🟢 Low / Info | 16 |

**Total findings:** 40

**Top risks:**
1. **CI does not verify the product end-to-end** — the entire chat/DM/feed/login journey (the mandatory audit features) has **zero Playwright coverage** and CI never runs E2E at all (CI-02, CI-04). Real WS/proxy/routing regressions ship green.
2. **`-race` is mandated by SDS §10.4 but never run in CI or the Makefile** (CI-01) — the only concurrent subsystem (the WS hub) could regress into a data race undetected. (It passes today.)
3. **No compile check in CI** (CI-03) — `go test ./...` doesn't build `cmd/backend`, `cmd/qa-seed`, `internal/middleware`, or `internal/router`; a compile error in those ships green.
4. **Live presence is not wired into the open conversation composer** (BUG-01) — a partner going offline/online mid-chat leaves the composer in a stale enabled/disabled state, violating the audit's "message online users" requirement.
5. **Defense-in-depth gaps on the web tier** — no Content-Security-Policy (SEC-01) and a permissive WebSocket `CheckOrigin` (SEC-03, CSWSH) mean a single missed escape or a cross-origin handshake becomes directly exploitable.

A large secondary theme is **legacy-migration debt**: the SPA migration left the entire `web/templates/` tree and 26 `web/static/js/*` files orphaned (and still publicly served), plus retained-but-unwired OAuth, kept alive only by a few stale tests.

---

## 1) Bugs & Logic Errors

### BUG-01: Conversation composer does not react to live presence changes ⬆ HIGH
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: D (Tickets: D02, D04) — contract SDS §7.5
- `SPA/features/chat/chat.conversation.page.js` (~L77, L181-216 — `isOnline` set only at selection time; L360-386 — subscribes to `chat:user-selected`/`DM_MESSAGE`/`ERROR` but **not** `PRESENCE_UPDATE`)
- `SPA/features/chat/chat.conversation.views.js` (~L76-118 — `renderComposer(isOnline)` rendered once per selection)

**Problem:** The composer's enabled/disabled state is computed from `detail.isOnline` at the instant a roster row is clicked, then baked into `innerHTML`. The conversation page never subscribes to `PRESENCE_UPDATE` (only the roster page does, `chat.roster.page.js:142-149`), so while a conversation is open the composer never updates when the partner's presence flips.
**Impact:** Partner goes **offline** mid-chat → composer stays enabled; the user sends and the backend rejects with `chat.error RECIPIENT_OFFLINE` (`internal/handlers/ws.go:265-267`). Partner comes **online** mid-chat → composer stays disabled until the row is reselected, so the user cannot message an online user — a direct hit on the audit requirement "send private messages to the users who are online."

**Fix:** Add a `PRESENCE_UPDATE` listener in `initChatConversation` that, when `detail.userId === state.recipientId`, updates `state.isOnline` and toggles the composer controls:
```js
documentRef.addEventListener(WS_EVENTS.PRESENCE_UPDATE, (event) => {
  const d = event?.detail;
  if (!d || Number(d.userId) !== state.recipientId) return;
  state.isOnline = Boolean(d.isOnline);
  const composer = activeRoot.querySelector(COMPOSER_SELECTOR);
  for (const el of composer?.querySelectorAll('[data-conversation-input],[data-conversation-send],[data-conversation-attach]') ?? [])
    state.isOnline ? el.removeAttribute('disabled') : el.setAttribute('disabled','');
  // also update the offline note + placeholder
});
```
**Tests to add:** Vitest — open a conversation with an online user, dispatch `PRESENCE_UPDATE {userId, isOnline:false}` → assert composer disabled; dispatch `isOnline:true` → assert re-enabled; presence for a *different* user must not touch the open composer.

---

### BUG-02: Session expiry mid-session leaves the SPA half-authenticated (no global 401 recovery) ⬆ MEDIUM
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: A (Tickets: A05, A06) — interacts with B06; contract SDS §7.2
- `SPA/core/app/create-app.js` (~L33-51 — `resolveSession` checks 401 only at boot; L228-301 — route rendering never re-checks auth post-boot)
- `SPA/features/notification/notification.page.js` (~L90-103, L303 — `onUnauthorized: stop` only stops the poller)

**Problem:** Auth is enforced only at boot and on explicit login/logout. There is no shared API layer that observes a 401 from any post-boot fetch (feed, roster, history, post detail) and downgrades the session. The notification center is the only consumer that detects 401, and its sole reaction is to stop its own interval — it does not touch app state, the socket, or routing.
**Impact:** When a session expires while the app is open, the user stays on an authenticated route with notifications stopped, a dead WS, and every REST call returning 401 → empty feed/roster, no redirect to `/login`, no recovery and no prompt to re-authenticate.

**Fix:** Centralize 401 handling — add an app-level `onUnauthorized` (set `state.isAuthenticated=false`, stop notifications, stop chat, `goTo('/login', true)`) and route feature fetches through a wrapper that invokes it on `response.status === 401`. Minimum: wire the notification center's `onUnauthorized` to the app's logout-finalize logic.
**Tests to add:** Integration — boot authenticated, stub a feature fetch to return 401, assert redirect to `/login`, polling stopped, socket closed.

---

### BUG-03: Newly-online users absent from the loaded roster never appear without a manual reload ⬆ MEDIUM
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: D (Tickets: D01, D04) — backend contract C04/C05
- `SPA/features/chat/chat.roster.logic.js` (~L29-38 — `setPresence` leaves unknown users untouched)
- `SPA/features/chat/chat.roster.page.js` (~L142-149 — PRESENCE_UPDATE only calls `setPresence`; roster fetched once at init L166-169)

**Problem:** The roster is fetched once on mount. A `presence.update` for a user **not already** in `state.entries` is intentionally a no-op, so a user who registers/logs in after the viewer loaded the app comes online but is never added to the roster.
**Impact:** The audit requires the online-users section be visible at all times and allow messaging online users. A user who joins after you loaded the app is invisible until a hard reload — in a fresh two-browser test (one user registers after the other is logged in), the real-time presence/messaging check can fail.

**Fix:** On a `presence.update` for an unknown `userId`, re-fetch the roster (cheap, debounced) — option (a), matching the C05 source of truth — or extend the `presence.update`/`presence.snapshot` payload with username to insert a new entry.
**Tests to add:** Vitest — roster loaded with [A]; dispatch `PRESENCE_UPDATE {userId: B(new), isOnline:true}`; assert B appears.

---

### BUG-04: `chat.error`/presence frames silently dropped to slow clients with no recovery signal ⬆ LOW
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: C (Tickets: C04, C06) — frontend D04
- `internal/handlers/ws.go` (~L315-324 — `sendChatError` non-blocking `select/default` drop)
- `internal/ws/connection_manager.go` (~L136-145, L161-167, L192-196 — all Hub sends drop-on-full; 64-slot buffer at L25)

**Problem:** Every server→client enqueue uses `select { case c.Send <- data: default: }`. On a saturated buffer the frame is dropped silently. For `dm.message` to the recipient this is acceptable (history backfills on reconnect), but for **`chat.error` to the sender** and **edge-triggered presence transitions** a drop means the sender thinks a send succeeded when it was rejected, or a client shows stale online/offline state indefinitely (presence is never re-snapshotted after connect).
**Impact:** Edge cases only (a client that stops draining its socket). Low likelihood in normal use.

**Fix:** On a full send buffer treat the client as dead and close the connection (readPump/writePump tear down and re-snapshot on reconnect) rather than silently dropping presence/error frames; or periodically re-broadcast a presence snapshot.
**Tests to add:** Go — fill a client's `Send` buffer, trigger a `chat.error`, assert the connection is closed rather than the frame silently dropped.

---

### BUG-05: Reaction engine deviates from the documented optimistic-flip-and-revert contract ⬆ LOW
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: B (Ticket: B07) — contract SDS §7.4.2
- `SPA/features/post/post.reactions.bindings.js` (~L101-104, L120-122 — binds a `click` listener using `aria-pressed`, updates DOM only after the server responds; not optimistic)
- `SPA/features/post/post.reactions.logic.js` (~L29-56 — `normalizeReactionState` correctly reconciles both payload shapes)

**Problem:** The implementation is server-authoritative (await response, then set state) on a `click`-bound button, whereas SDS §7.4.2 specifies a delegated `change` listener on a checkbox with an optimistic flip reverted on failure/unauthenticated. The dual-payload-shape normalization the SDS calls out **is** implemented and correct; only the binding mechanism/interaction model differs.
**Impact:** No runtime bug — reactions work and failed/unauthorized requests correctly leave state untouched (`bindings.js:88-97`). This is a spec/contract mismatch, flagged for traceability.

**Fix:** Update SDS §7.4.2 to describe the server-authoritative click model (recommended — the current code is the safer pattern), or refactor to the documented optimistic `change` model.
**Tests to add:** None for correctness; if aligning to SDS, add a Vitest test for optimistic flip + revert on a failed POST.

---

## 2) Dead Code & Unused References

> **Removal ordering:** the stale tests DEAD-07 and DEAD-09 are the only remaining references keeping the legacy assets "alive" — they must be removed **before/with** DEAD-01/02/04/06.

### DEAD-01: Orphaned legacy HTML templates (entire `web/templates/`) ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References (also surfaced by Agent 3 as ARCH-04)
**Files:** Ownership: B (Tickets: B01, B03, B04) + A (A10)
- `web/templates/create-post.html`, `edit-post.html`, `home.html`, `login.html`, `register.html`, `view-post.html` (6 files)

**Problem:** No Go handler parses or serves them (`grep ParseFiles`/`templates/` in `internal/`+`cmd/` non-test = none); the frontend server never exposes `web/templates`. Wholly superseded by SPA routes/views. The only references are stale fixture tests (DEAD-09). Note: not served at runtime, so the "single HTML shell" rule is not breached in production — this is dead-on-disk debt.

**Fix:** Delete `web/templates/` after removing the DEAD-09 fixture tests.

---

### DEAD-02: Orphaned legacy frontend JS — still publicly served via `/static/js/` ⬆ HIGH
**Origin:** 2. Dead Code & Unused References (merged with Agent 3 ARCH-04 — "still-served `web/static/js`" boundary angle)
**Files:** Ownership: B (B0x feed/post/reactions/notifications) + A (A0x auth/header/logout) + D (D04 auth-modal)
- 26 files in `web/static/js/` — `auth.js`, `auth-modal.js`, `create-post.js`, `edit-post.js`, `home.js`, `posts.js`, `view-post.js`, `reactions.js`, `notifications.js`, `header*.js`, `login.js`, `register.js`, `logout.js`, `post-actions.js`, `post-form.js`, `pagination.js`, `category.js`, `helpers.js`, `utils.js`, `image-picker.js`, `password-toggle.js`, `settings*.js`, `sound-effects.js`, `ui-messages.js`
- Served by `cmd/frontend/routes.go:19-25` (`/static/` file server)

**Problem:** These are loaded only by the orphaned templates (DEAD-01) — a closed legacy loop. The SPA imports exclusively from `SPA/` (enforced by `SPA/tests/unit/policy/spa-import-paths.test.js`, which forbids `/static/js/`). They remain **reachable as static files** through `/static/` even though nothing in the SPA loads them, presenting a contradictory second implementation and a maintenance/confusion hazard. The tracker's noted "B04 create-post hard-reload leak" (`web/static/js/create-post.js:319,373` → `window.location.href`) lives here and is moot once the file is deleted.

**Fix:** Delete `web/static/js/` (after DEAD-09); at minimum stop serving `/static/js/`.

---

### DEAD-03: Retained-but-unwired OAuth — dead routes still registered ⬆ HIGH
**Origin:** 2. Dead Code & Unused References (related doc-drift in ARCH-02)
**Files:** Ownership: C (C11 auth/registration domain; Go predates RTF but is orphaned from the product)
- `internal/auth/oauth.go`, `internal/handlers/oauth_github.go`, `oauth_google.go`, `oauth_helpers.go`
- Routes still registered at `internal/router/router.go:206-239` (`/auth/google`, `/auth/google/callback`, `/auth/github`, `/auth/github/callback`)

**Problem:** Per AGENTS.md "Key Design Rules 3" and SDS §6.3, OAuth is not part of the product. The SPA deliberately exposes no OAuth UI (`SPA/tests/unit/core/app/create-app.test.js:389-391` asserts the endpoints are absent). These are 4 reachable-but-unused backend endpoints plus ~14KB of unreferenced handler/helper code. This is the AGENTS.md "retained" exception — flagged as dead/retained.

**Fix:** If removal is desired, drop the 4 route registrations and the 4 OAuth Go files together; the related legacy assets `web/static/img/google-logo.svg`, `github-logo.svg`, `web/static/css/auth-modal.css`, `web/static/partials/auth-modal.html` are orphaned with them. Otherwise document explicitly as retained-non-product. Either way, reconcile `architecture.md` (ARCH-02).

---

### DEAD-04: Orphaned legacy CSS (`web/static/css/*`) ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References (also Agent 3 ARCH-04)
**Files:** Ownership: B/A (mirrors DEAD-02 owners)
- 13 files in `web/static/css/` — `auth-modal.css`, `base.css`, `create-post.css`, `header.css`, `home.css`, `login-register.css`, `pagination.css`, `post-navigation.css`, `posts.css`, `settings.css`, `ui-messages.css`, `view-post.css`, `components/password-field.css`

**Problem:** Loaded only by the orphaned templates (DEAD-01). The SPA uses `SPA/assets/css/*` + feature-co-located CSS. (`web/static/css/activity.css` was already deleted under B05; these are the remaining same-class orphans. `web/errors/common.css` is NOT dead — served by `error.html`.)

**Fix:** Delete `web/static/css/` (after DEAD-07, since `base.css` is asserted by the stale e2e test).

---

### DEAD-05: Orphaned legacy HTML partials & sound assets ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: A (header partial) + D (auth-modal partial) + B (settings/notification sounds)
- `web/static/partials/auth-modal.html`, `header.html`, `settings-modal.html` (3)
- `web/static/sounds/delete.mp3`, `notification.mp3`, `reaction.mp3`, `upload.mp3` (4)

**Problem:** Partials were fetched by legacy `header-loader.js`/`settings-loader.js` (themselves dead, DEAD-02). The SPA plays no audio — no `.mp3`/`new Audio`/`sounds/` reference in `SPA/` non-test code; `SPA/assets/sounds/` is empty (`.gitkeep` only).

**Fix:** Delete `web/static/partials/` and `web/static/sounds/`.

---

### DEAD-06: Orphaned legacy images (`web/static/img/*` except favicon) ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References (also Agent 3 ARCH-04)
**Files:** Ownership: A/B (header/branding/post-action icons)
- 13 files in `web/static/img/` — `forum-logo.png`, `forum-logo-night.png`, `camera.png`, `close.png`, `create-post.png`, `delete.png`, `draft.png`, `edit.png`, `my-activity.png`, `paperclip.png`, `publish.png`, `github-logo.svg`, `google-logo.svg`

**Problem:** The SPA uses its own `SPA/assets/img/*`. Legacy images are referenced only by orphaned templates/CSS and the stale e2e test (`forum-logo.png`). (`web/static/favicon.ico` is NOT dead — served at `/favicon.ico`.)

**Fix:** Delete legacy `web/static/img/*` except `favicon.ico` (google/github SVGs tie to DEAD-03).

---

### DEAD-07: Stale e2e test asserting orphaned legacy assets remain served ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: D (D05/D06 — e2e suite is D-owned)
- `SPA/tests/e2e/tickets.test.js` (~L185-213 — test `A02-03: legacy /static assets remain accessible` asserts HTTP 200 for `/static/css/base.css`, `/static/js/auth.js`, `/static/img/forum-logo.png`)

**Problem:** This test exists solely to keep orphaned legacy assets (DEAD-02/04/06) alive and contradicts both the SPA migration intent and the import-path policy test. It actively blocks removal.

**Fix:** Delete the `A02-03` test block; it gates removal of DEAD-02/04/06.

---

### DEAD-08: Stale `vitest.config.ts` alias to legacy JS ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: A (A01 tooling)
- `vitest.config.ts` (~L6 — `'/static/js': path.resolve(__dirname, './web/static/js')`)

**Problem:** This resolve alias only supports imports of legacy `web/static/js/*`, which the import-path policy forbids and no active test imports. Once DEAD-02 is removed it points at nothing.

**Fix:** Remove the `resolve.alias` block.

---

### DEAD-09: Stale backend tests pinning orphaned legacy assets as fixtures ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: B (B04/B08 surface; tests in `internal/tests`)
- `internal/tests/drafts_test.go` (~L256-293 — `TestCreatePostTemplate_HasImagePreviewControls` reads `../../web/templates/create-post.html`; `TestCreatePostJS_ConditionallyRendersPostImageMarkup` reads `../../web/static/js/posts.js`)

**Problem:** Both assert on legacy template/JS the SPA replaced; they test product surface no longer served and are the only non-e2e references keeping DEAD-01/DEAD-02 referenced.

**Fix:** Delete both test functions; must go before DEAD-01/DEAD-02 file removal.

---

### DEAD-10: Unused SPA exports (fully dead) ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: B (feed B01; activity B05)
- `SPA/features/feed/feed.state.js` (~L9 — `export function getDefaultFeedState()`, never imported, superseded by `getFeedQueryState`)
- `SPA/features/activity/activity.api.js` (~L171 — `export const ACTIVITY_DEFAULTS`, never imported, unused internally)

**Problem:** Zero references across `SPA/` (src + tests). Biome's `noUnusedVariables: error` does not catch unused *exports*.

**Fix:** Remove both declarations.

---

### DEAD-11: `package.json` redundant scripts + invalid `main` ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: A (A01 tooling)
- `package.json` — `"main": "index.js"` (no such file; project is Go + ES-module SPA); duplicate pairs `"lint"`≡`"check"` (both `biome check .`) and `"lint:fix"`≡`"fix"` (both `biome check --write --unsafe .`)

**Fix:** Drop `main`; consolidate the duplicate `check`/`fix` aliases (Makefile uses `lint`/`lint:fix`, keep those).

---

### DEAD-12: Makefile duplicate `PORT` definition ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: A (A01 infra/Makefile)
- `Makefile` (~L20 `PORT = 8080` and ~L135 `PORT = 8080` in the Docker block — second redefines to the same value)

**Fix:** Remove the duplicate at line 135.

---

### DEAD-13: Dual lockfiles tracked ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: A (A01 tooling; AGENTS.md mandates Bun)
- `bun.lock` and `package-lock.json` both tracked. AGENTS.md/SDS declare Bun the required runtime; `package-lock.json` (npm) is redundant. `biome.json:18-19` already excludes both; `Makefile:126` only falls back to `npm install` if Bun is absent.

**Fix:** Remove `package-lock.json` from tracking (keep `bun.lock`); optionally gitignore it.

---

## 3) Architecture, Boundary Violations & Guideline Drift

> The architecture is fundamentally sound: split-server topology, single-HTML SPA shell, vanilla-JS frontend, layered backend, session-cookie auth, WebSocket-only chat, and SDS event/REST contracts are all correctly implemented and tested. Findings are drift/hygiene, not functional breaks. (Agent 3's original ARCH-04 — dead legacy assets still served — is merged into the Dead Code cluster DEAD-01/02/04/06.)

### ARCH-01: Documented middleware flow does not match implementation ⬆ MEDIUM
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** AGENTS.md Code Conventions (Middleware): *"Request flow: Logger → Recoverer → CORS → OptionalAuth → Auth → Handler."*
**Files:** Ownership: C (C01, C08) + cross-cutting A (A05)
- `internal/router/router.go` (~L337-342 `addMiddlewares`), `internal/middleware/middleware.go`

**Problem:** Actual wrap order is `CORS(Logger(Recoverer(handler)))` → execution `CORS → Logger → Recoverer → Auth`, not Logger-first. There is **no `OptionalAuth` middleware** anywhere (grep returns nothing); `Auth` is applied per-route, not as a documented chain stage. Logger wrapping Recoverer means a recovered panic logs status 200 unless Recoverer sets it first.
**Impact:** Functionally acceptable (panics still recovered, CORS-first short-circuits OPTIONS), but the canonical AGENTS.md flow is unimplemented as written — clearest case of guideline drift.

**Fix:** Either reorder to `Logger(Recoverer(CORS(...)))` to match the doc, or update AGENTS.md to the actual order and drop the non-existent `OptionalAuth` stage.

---

### ARCH-02: `architecture.md` stale — lists OAuth as supported auth and chat as "in development" ⬆ MEDIUM
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift (related to DEAD-03)
**Violated rule:** AGENTS.md Key Design Rule 3: *"Session-cookie auth — HttpOnly cookies. No JWT. No OAuth for the real-time forum."*
**Files:** Ownership: C (C07, C08) + D (D07)
- `architecture.md` (~L80-89 "Google & GitHub OAuth" under §4 Authentication Model; ~L70 "Chat and messaging areas are in active development"; ~L23/L68 feature-slice list omits `chat`/`profile`)

**Problem:** Contradicts the tracker (37/37 done, chat complete) and SDS §6.3 ("OAuth is not part of the target product design"). An auditor reading this would conclude OAuth gates auth and chat is unfinished. The SPA UI exposes no OAuth/guest entry (D10 satisfied), so this is doc drift, not a real auth breach.

**Fix:** Mark OAuth as retained/non-product legacy, update §3 chat status to complete, add chat/profile to the slice lists.

---

### ARCH-03: `README.md` project-status drift ⬆ LOW
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** A02/A07 gate "docs reflect the current implementation accurately."
**Files:** Ownership: D (D07)
- `README.md` (~L14-18 "Active Development (Wave 3)", "Real-Time Chat (Wave 3): In Progress")

**Problem:** Tracker reports all 6 waves / 37 tickets done; README says Wave 3 in progress. Misleads on completeness; no code impact.

**Fix:** Update the status block to reflect completion.

---

### ARCH-04: `Authorization: Bearer` token fallback in Auth middleware ⬆ LOW
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** AGENTS.md Key Design Rule 3: *"Session-cookie auth — HttpOnly cookies. No JWT."*
**Files:** Ownership: A (A05) + C (C01)
- `internal/middleware/auth.go` (~L32-37)

**Problem:** After the cookie check, Auth falls back to a `Bearer <token>` Authorization header (the session token, not a JWT). Doesn't strictly violate "No JWT," but introduces a non-cookie auth path the design says shouldn't exist. The WS handler (`internal/handlers/ws.go:45-49`) correctly uses cookie-only; the Bearer path is unused by the SPA (always `credentials: 'include'`).

**Fix:** Remove the Bearer fallback for full conformance, or document it as an intentional test/tooling affordance.

---

### ARCH-05: Handler layer issues a raw SQL query (layering breach) ⬆ LOW
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** AGENTS.md Code Conventions (Handlers): *"internal/handlers/ — no direct database/sql queries (must call db layer)."*
**Files:** Ownership: B (B08 drafts — pre-existing earlier-forum code within the RTF-touched `drafts.go`)
- `internal/handlers/drafts.go` (~L357-372 `fetchDraftImageURLByID` runs `db.QueryRowContext(...)` with inline SQL directly in the handler package)

**Problem:** A `SELECT image_url FROM posts ...` is embedded in the handler layer instead of `internal/db/`, breaking db/handler separation. (Note: `internal/db/messages.go` `CreateMessage` does INSERT-then-SELECT without a transaction, but that is a single write + read-back, not a multi-step write — the transaction rule does not apply, acceptable.)

**Fix:** Move the query into an `internal/db/` repository function (e.g. `db.GetDraftImageURL(ctx, ...)`).

---

### ARCH-06: Schema deviates from SDS column DDL for new user fields ⬆ LOW
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** SDS §4.1 DDL (`age INTEGER NOT NULL`, `gender TEXT NOT NULL`, no defaults)
**Files:** Ownership: C (C10, C07)
- `internal/db/forum_schema.sql` (~L30-33 `age INTEGER NOT NULL DEFAULT 0`, `gender/first_name/last_name TEXT NOT NULL DEFAULT ''`); migration mirror at `internal/db/migrate.go:38-41`

**Problem:** Schema adds `DEFAULT 0`/`DEFAULT ''` not present in SDS §4.1. This is a deliberate, correct migration-safety choice (existing rows need backfill values), explicitly sanctioned by SDS §8, and C11 registration validation enforces non-empty profile fields at the API layer so the defaults can't be reached by new registrations. Flagged only as a literal deviation from §4.1.

**Fix:** None required; optionally reconcile SDS §4.1 wording with the migration-safe DDL.

---

## 4) Code Quality & Security

> Overall posture is strong: all SPA user-content render paths run through `escapeHTML`; all RTF backend SQL is parameterized; sessions are HttpOnly + validated on every protected request and on WS upgrade; bcrypt hashing; DM send fail-closed; uploads content-sniffed with UUID filenames; `go test -race` passes clean. Findings are hardening gaps, not active exploits.

### SEC-01: No Content-Security-Policy (meta tag or HTTP headers) ⬆ HIGH
**Origin:** 4. Code Quality & Security
**Files:** Ownership: A (A10 owns `SPA/index.html`) + A01/D09 (frontend-server header emission)
- `SPA/index.html` (~L1-39 — no CSP meta)
- `cmd/frontend/routes.go` (~L29, L81-83 — sets only Cache-Control/Pragma/Expires; no CSP, X-Frame-Options, or X-Content-Type-Options anywhere in `cmd/frontend/`)

**Problem:** The app renders user-generated content into `innerHTML` across ~15 sites. Escaping is currently correct everywhere, but with zero CSP there is no defense-in-depth: a single missed escape, a dependency issue, or the inline handler (SEC-02) becomes directly exploitable XSS. No clickjacking or MIME-sniff protection either. The page loads Google Fonts cross-origin (`index.html:12-16`), which a CSP must allowlist.

**Fix:** Emit a CSP from the frontend server (it controls all responses incl. the shell), e.g. `default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; object-src 'none'`. Add `X-Content-Type-Options: nosniff` and `X-Frame-Options: DENY`. `script-src 'self'` requires removing the inline handler (SEC-02).

---

### SEC-02: Inline `onclick` event handler in rendered markup ⬆ MEDIUM
**Origin:** 4. Code Quality & Security
**Files:** Ownership: A (A07 owns `SPA/features/profile/`)
- `SPA/features/profile/profile.page.js` (~L76 — `<button class="auth-button" onclick="window.history.back()">Go Back</button>`)

**Problem:** Violates the mandated event-delegation/`addEventListener` convention (AGENTS.md, SDS §7.0) and is the one inline handler in the codebase. It also blocks adopting a strict `script-src 'self'` CSP (SEC-01) — inline handlers require `'unsafe-inline'`.

**Fix:** Render the button without the inline handler and bind via `addEventListener` in `initProfilePage`, calling `windowRef.history.back()`.

---

### SEC-03: WebSocket upgrade accepts any Origin (CSWSH) ⬆ HIGH
**Origin:** 4. Code Quality & Security *(Agent rated Medium; raised to High — enables authenticated session-riding over WS)*
**Files:** Ownership: C (C01)
- `internal/handlers/ws.go` (~L19-25 — `upgrader.CheckOrigin` unconditionally `return true`)

**Problem:** Cross-Site WebSocket Hijacking. The WS authenticates purely via the `session_token` cookie (`ws.go:44-58`), which is `SameSite=Lax`. Browsers send cookies on cross-origin WebSocket handshakes and `Lax` does not reliably block them for WS upgrades, so any external origin can open an authenticated `/ws` for a logged-in victim, receive their `presence.snapshot`/`dm.message` events, and send DMs as them.
**Impact:** Authenticated impersonation + private-message disclosure across origins.

**Fix:** Implement `CheckOrigin` to allowlist the configured frontend origin (the `FRONTEND_URL` value already used for CORS at `internal/router/router.go:33-35`), rejecting mismatched/empty origins.

---

### SEC-04: No global `unhandledrejection` / error handler in the SPA ⬆ LOW
**Origin:** 4. Code Quality & Security
**Files:** Ownership: A (A03/A10; bootstrap in `SPA/main.js`)
- `SPA/main.js` (no handler registered — confirmed absent; only a per-image `addEventListener('error')` exists in `SPA/core/shared/image-picker.js:207`)

**Problem:** A rejected promise anywhere in the async UI flows (fetch failures, WS handlers, route initializers) surfaces only to the console with no user feedback and no central logging hook. Robustness/observability gap, explicitly called out by the audit.

**Fix:** Register `window.addEventListener('unhandledrejection', …)` (and optionally `'error'`) in `main.js` to log and optionally surface a non-blocking notice.

---

### SEC-05: HTML-context escaping used inside a CSS `url()` context ⬆ LOW
**Origin:** 4. Code Quality & Security
**Files:** Ownership: B (B01/B07 feed/post card rendering)
- `SPA/features/post/post-card.views.js` (~L155 — `style="background-image: url('${escapeHTML(imageUrl)}')"`)

**Problem:** `escapeHTML` is an HTML-entity escaper, not a CSS escaper. The value sits in an inline `style` attribute (CSS nested in HTML); the browser HTML-decodes the attribute before CSS parsing, so an entity-encoded `'` round-trips to a literal `'` and could break out of the CSS string. Exploitability is currently very low because `imageUrl` is a server-generated upload path (`/static/uploads/<uuid>.<ext>`), not free-form input — but the escaping is contextually wrong and becomes a real CSS-injection vector if the field ever carries user-controlled data. The companion `<img src="${escapeHTML(imageUrl)}">` on the next line is in a correct HTML-attribute context and is fine.

**Fix:** Set the background via JS (`el.style.backgroundImage = ...`) using a proper URL/CSS escaper, or validate the URL against the known upload-path prefix before interpolation.

---

### SEC-06: Session cookies omit the `Secure` flag ⬆ LOW
**Origin:** 4. Code Quality & Security
**Files:** Ownership: C (C11 / auth handlers; pre-existing forum-auth code reused by RTF)
- `internal/handlers/users.go` (~L135-141 register, L193-199 login, L234-240 logout); `internal/middleware/auth.go` (~L66-75 clear) — all set `HttpOnly: true, SameSite: Lax` but never `Secure: true`

**Problem:** Over plain HTTP (the documented dev topology, ports 3000/8080) this is expected, but if the app is ever served over TLS the session cookie would still be transmittable over HTTP, enabling interception/downgrade. `HttpOnly` + `SameSite=Lax` are correctly set.

**Fix:** Set `Secure: true` conditionally when serving over HTTPS (env-flag driven), or unconditionally in production builds.

---

## 5) Tests & CI Gaps

### CI-01: `-race` never run in CI or Makefile despite SDS §10.4 mandate ⬆ CRITICAL
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: C (C08) + D (D07); SDS §10.4
- `Makefile` (~L92-94 `test-backend` = plain `go test ./...`), `.github/workflows/ci.yml` (~L35-36 `make test-backend`), `internal/ws/connection_manager.go` + `internal/ws/*`

**Problem:** The WS hub/presence/DM broadcast is the only concurrent (goroutine + channel + shared-map) subsystem, and SDS §10.4 explicitly requires race verification. No build path runs `-race`. A data race in the Hub would pass CI silently. (It passes today — verified this audit — so adding it is non-breaking.)

**Fix:** Add a `test-race` target (`go test -race ./internal/ws/... ./internal/tests/...`) and a CI step invoking it. At minimum gate the `ws` package.
**Tests to add:** the CI/Makefile target itself.

---

### CI-02: CI does not run E2E (Playwright) — E2E regressions pass CI ⬆ CRITICAL
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (A01) + D (D07)
- `.github/workflows/ci.yml` (~L35-39 — runs only `test-backend` + `test-frontend`), `Makefile` (~L84 `test` includes `test-e2e` but CI never calls `make test`/`make test-e2e`)

**Problem:** `make test-frontend` = `biome check && vitest run`; Playwright is excluded (`vitest.config.ts:14`). No CI job ever starts the real servers or runs Playwright, so any full-stack regression (proxy, routing fallback, auth-gate redirect, shell stability) ships green.

**Fix:** Add a separate CI job running `make test-e2e` (already free-ports + spins fresh servers via `reuseExistingServer:false`); use `playwright install --with-deps chromium` and upload the HTML report artifact.

---

### CI-03: CI does not run `make build` (compile check) ⬆ HIGH
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (A01)
- `.github/workflows/ci.yml` (no `make build` step), `Makefile` (~L36)

**Problem:** A compile error in any non-test Go file in packages with no `_test.go` (`cmd/backend`, `cmd/qa-seed`, `internal/middleware`, `internal/router`, `internal/env`) would NOT be caught — `go test ./...` only builds packages with tests or imported by them, and `cmd/backend`/`cmd/qa-seed` are not imported by the test tree.

**Fix:** Add `make build` (or `go build ./...`) to the quality-gate job before tests.

---

### CI-04: E2E suite covers Track A only; entire chat/forum journey unverified at E2E tier ⬆ CRITICAL
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: D (D06, D07); SDS §10.2.2; audit.md chat questions (L45-75)
- `SPA/tests/e2e/tickets.test.js` (~L143-723 — only A02/A03/A04/A05/A06/A07/A10 `describe` blocks)

**Problem:** No Playwright coverage for: two-browser real-time DM delivery (audit.md:57-63); roster ordering by last-message/alphabetical (audit.md:45-51); message format showing username + date (audit.md:53-55); last-10 history + scroll-up pagination with throttle (audit.md:65-75); disabled composer for an offline user (SDS §10.2); feed-has-no-comments / post-detail-has-comments in a real browser (audit.md:33-43); and login/registration through the actual rendered form (every E2E auths via `page.request.post` API injection, `tickets.test.js:38-67`). These mandatory audit requirements are only verified at the Vitest tier with **mocked** sockets/DOM — never against the real backend/WebSocket/proxy stack.

**Fix:** Add Playwright specs for (a) two-context real-time DM round-trip, (b) roster ordering after messaging, (c) message bubble shows username+timestamp, (d) scroll-up loads 10 older without API spam, (e) offline composer disabled, (f) feed-vs-detail comment rendering, (g) form-based login + register. The harness already supports `browser.newContext()` (used in A02-02).

---

### CI-05: Stale audit-coverage report is misleading and never updated ⬆ MEDIUM
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: D (D07)
- `docs/audit-reports/audit-e2e-coverage-2026-04-18.md` (entire file)

**Problem:** Dated 2026-04-18, states C05/D01/D02/D04/C06/B01–B04 are "Not Started / Nothing" with hard-coded `file:///home/ertval/...` paths. The tracker now shows all 37 tickets done. Factually wrong; would mislead an auditor.

**Fix:** Regenerate or supersede with a current traceability matrix (see CI-06); mark explicitly superseded if retained.

---

### CI-06: No audit traceability matrix exists ⬆ HIGH
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: D (D07 — final acceptance gate = "comprehensive audit pass using full audit.md checklist")
- none — `docs/audit.md` questions are not mapped to any test

**Problem:** No artifact maps each audit.md question → the proving test. Several questions have no proving test at any tier: "chat users organized by last message" (audit.md:47 — backend covered `chat_roster_test.go:123`, but no frontend E2E renders the ordered roster); "alphabetic order for a no-history user" (audit.md:49-51 — the no-prior-message alphabetical tiebreak isn't separately asserted); "did the other user receive a notification of a new DM?" (audit.md:57-59 — no test asserts a DM produces a user-visible notification/unread indicator).

**Fix:** Add `docs/audit-reports/traceability-matrix.md` with one row per audit question → test file:line, and close the uncovered rows above.

---

### CI-07: No Vitest coverage thresholds; coverage not measured ⬆ MEDIUM
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (A01) + D (D05/D06)
- `vitest.config.ts` (~L10-15 — no `coverage` block), `package.json` (~L9 `test` = `vitest run`, no `--coverage`)

**Problem:** No coverage provider/thresholds, so coverage regressions are invisible. E2E exclusion is correct (`exclude: ['SPA/tests/e2e/**']`), but `include` is repo-wide with no `coverage.include` scoping to `SPA/`.

**Fix:** Add a `coverage` block (provider `v8`, `include: ['SPA/**']`, thresholds ~70–80%) and run `vitest run --coverage` in CI as a ratchet.

---

### CI-08: `internal/middleware` and `internal/router` have no co-located unit tests ⬆ MEDIUM
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (A05) + C (C01)
- `internal/middleware/auth.go`, `internal/middleware/middleware.go`, `internal/router/router.go` (no `_test.go` — `go test ./...` reports "no test files")

**Problem:** The `Auth` middleware (the chokepoint that injects `userID` and rejects unauthenticated requests, incl. the WS-upgrade gate) has no isolated unit test. It is exercised transitively via `internal/tests/`, so behavior is verified, but edge cases (expired session, malformed cookie, `session_version` mismatch, context injection) are only indirectly covered.

**Fix:** Add `internal/middleware/auth_test.go` with table-driven cases for valid/expired/invalid/missing cookie + context propagation. Lower priority than CI-01/02/04 given existing integration coverage.

---

### CI-09: No `go vet` in CI ⬆ LOW
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (A01)
- `.github/workflows/ci.yml` (~L28-33 — runs `make lint` (biome) + `make format` then `git diff --exit-code`); a `vet` target exists at `Makefile:117` but is uncalled

**Problem:** Formatting drift is caught; Go static issues (shadowing, printf, unreachable) are not. The `git diff` format gate itself is sound.

**Fix:** Add `make vet` (`go vet ./...`) to the quality-gate job.

---

### CI-10: CI trigger branches `[main, dev]` may miss feature-branch PRs ⬆ LOW
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: A (A01)
- `.github/workflows/ci.yml` (~L3-7 — `pull_request` targets base branches `main`/`dev` only)

**Problem:** PRs whose base is a long-lived integration/feature branch (the repo uses `asmyrogl/*` branches per git log) won't run CI. Feature→feature merges before main are ungated.

**Fix:** Standardize PR base to `main`/`dev`, or broaden/drop the `pull_request.branches` filter so all PRs run CI.

---

## Cross-Reference: Finding ID Mapping

| Consolidated ID | Agent 1 | Agent 2 | Agent 3 | Agent 4 | Agent 5 | Track Ownership | Description |
|----------------|---------|---------|---------|---------|---------|-----------------|-------------|
| BUG-01 | BUG-01 | — | — | — | — | D | Composer ignores live presence updates |
| BUG-02 | BUG-02 | — | — | — | — | A | No global 401 recovery; half-auth state |
| BUG-03 | BUG-03 | — | — | — | — | D | Newly-online users never added to roster |
| BUG-04 | BUG-04 | — | — | — | — | C | Drop-on-full hides chat.error/presence |
| BUG-05 | BUG-05 | — | — | — | — | B | Reaction binding deviates from SDS §7.4.2 |
| DEAD-01 | — | DEAD-01 | ARCH-04 | — | — | B/A | Orphaned `web/templates/*` |
| DEAD-02 | — | DEAD-02 | ARCH-04 | — | — | B/A/D | Orphaned `web/static/js/*` still served |
| DEAD-03 | — | DEAD-03 | (ARCH-02) | — | — | C | Retained OAuth routes/files unwired |
| DEAD-04 | — | DEAD-04 | ARCH-04 | — | — | B/A | Orphaned `web/static/css/*` |
| DEAD-05 | — | DEAD-05 | — | — | — | A/D/B | Orphaned partials + sound assets |
| DEAD-06 | — | DEAD-06 | ARCH-04 | — | — | A/B | Orphaned legacy images |
| DEAD-07 | — | DEAD-07 | — | — | — | D | Stale e2e test keeps legacy assets alive |
| DEAD-08 | — | DEAD-08 | — | — | — | A | Stale vitest `/static/js` alias |
| DEAD-09 | — | DEAD-09 | — | — | — | B | Stale backend fixture tests |
| DEAD-10 | — | DEAD-10 | — | — | — | B | Unused SPA exports |
| DEAD-11 | — | DEAD-11 | — | — | — | A | package.json redundant scripts + main |
| DEAD-12 | — | DEAD-12 | — | — | — | A | Makefile duplicate PORT |
| DEAD-13 | — | DEAD-13 | — | — | — | A | Dual lockfiles tracked |
| ARCH-01 | — | — | ARCH-01 | — | (CI-08) | C/A | Middleware flow doc ≠ code; no OptionalAuth |
| ARCH-02 | — | (DEAD-03) | ARCH-02 | — | — | C/D | architecture.md OAuth/chat stale |
| ARCH-03 | — | — | ARCH-03 | — | — | D | README status drift |
| ARCH-04 | — | — | ARCH-05 | — | — | A/C | Bearer token fallback in Auth |
| ARCH-05 | — | — | ARCH-06 | — | — | B | Handler raw SQL (drafts) |
| ARCH-06 | — | — | ARCH-07 | — | — | C | Schema DEFAULTs deviate from SDS §4.1 |
| SEC-01 | — | — | — | SEC-01 | — | A | No Content-Security-Policy |
| SEC-02 | — | — | — | SEC-02 | — | A | Inline onclick handler |
| SEC-03 | — | — | — | SEC-03 | — | C | WS CheckOrigin accepts any origin (CSWSH) |
| SEC-04 | — | — | — | SEC-04 | — | A | No unhandledrejection handler |
| SEC-05 | — | — | — | SEC-05 | — | B | HTML escaper used in CSS url() context |
| SEC-06 | — | — | — | SEC-06 | — | C | Session cookies lack Secure flag |
| CI-01 | (race ran clean) | — | — | (race clean) | CI-01 | C/D | `-race` not in CI/Makefile (SDS §10.4) |
| CI-02 | — | — | — | — | CI-02 | A/D | CI never runs E2E |
| CI-03 | — | — | — | — | CI-03 | A | CI never runs compile/build check |
| CI-04 | — | — | — | — | CI-04 | D | E2E covers Track A only |
| CI-05 | — | — | — | — | CI-05 | D | Stale audit-coverage report |
| CI-06 | — | — | — | — | CI-06 | D | No traceability matrix |
| CI-07 | — | — | — | — | CI-07 | A/D | No coverage thresholds |
| CI-08 | — | — | (ARCH-01) | — | CI-08 | A/C | No middleware/router unit tests |
| CI-09 | — | — | — | — | CI-09 | A | No `go vet` in CI |
| CI-10 | — | — | — | — | CI-10 | A | CI trigger branches may miss PRs |

---

## Recommended Fix Order

### Phase 1 — Critical (must fix before relying on CI as a quality gate)
1. **CI-03**: Add `go build ./...` to CI quality-gate (cheapest, closes the compile-check hole) (A)
2. **CI-01**: Add a `-race` target + CI step for `internal/ws` / `internal/tests` (SDS §10.4) — non-breaking, passes today (C/D)
3. **CI-02**: Add a CI job that runs `make test-e2e` with `playwright install` + report artifact (A/D)
4. **CI-04**: Author the missing chat/feed/login Playwright specs so the mandatory audit features are verified in a real browser (D)

### Phase 2 — High Severity (immediate follow-up)
5. **BUG-01**: Wire `PRESENCE_UPDATE` into the open conversation composer (D)
6. **SEC-03**: Restrict WS `CheckOrigin` to the configured frontend origin (CSWSH) (C)
7. **SEC-01**: Emit CSP + `X-Frame-Options` + `nosniff` from the frontend server (A)
8. **CI-06**: Create `docs/audit-reports/traceability-matrix.md` and close the uncovered audit questions (D)
9. **DEAD-02 / DEAD-03**: Remove (or stop serving) orphaned `web/static/js/*`; remove/explicitly-retain OAuth routes+files (B/A, C)

### Phase 3 — Medium Severity
10. **BUG-02 / BUG-03**: Centralize 401 recovery; refresh roster on unknown-user presence (A, D)
11. **SEC-02**: Replace the inline `onclick` with `addEventListener` (prereq for strict CSP) (A)
12. **ARCH-01 / ARCH-02**: Reconcile middleware-flow doc with code; fix architecture.md OAuth/chat status (C/A, C/D)
13. **DEAD-01/04/05/06/07/08**: Delete orphaned templates/css/partials/sounds/images + the stale e2e test + vitest alias (B/A/D) — *do DEAD-07/09 first*
14. **CI-05 / CI-07 / CI-08**: Refresh/supersede the stale coverage report; add coverage thresholds; add middleware unit tests (D, A, A/C)

### Phase 4 — Low Severity (maintenance)
15. **BUG-04 / BUG-05**: Close slow-client on full buffer; reconcile reaction model with SDS §7.4.2 (C, B)
16. **SEC-04 / SEC-05 / SEC-06**: Add `unhandledrejection` handler; CSS-safe escaping for `url()`; conditional `Secure` cookie (A, B, C)
17. **ARCH-03/04/05/06**: README status; remove Bearer fallback; move drafts raw SQL into db layer; reconcile schema DDL wording (D, A/C, B, C)
18. **DEAD-09/10/11/12/13**: Remove stale fixture tests, unused exports, redundant scripts, duplicate Makefile PORT, dual lockfile (B, A)
19. **CI-09 / CI-10**: Add `go vet` to CI; broaden PR trigger branches (A)

---

## Notes

- **No data races.** `go test -race -count=1 ./...` exits 0 across `internal/ws`, `internal/handlers`, `internal/db`, and the 269s `internal/tests` integration suite. The WS hub's send-on-closed-channel path was traced exhaustively and is safe (map deletion + `close()` are serialized under the Hub mutex, with `close()` strictly after `Remove`).
- **Backend SDS §10.1 coverage is comprehensive** — extended registration, login-by-username/email, auth-gating rejection, persistence, latest-10/`before_id` history, roster ordering + offline rejection, and presence connect/disconnect + multi-connection are all covered by `internal/tests/` and co-located `_test.go`.
- **Confirmed-clean security surface:** all RTF `internal/db` SQL parameterized (`GetUsersByIDs` uses `fmt.Sprintf` only to build `?` placeholders, data passed as args); authorization scoping correct (history scoped to the requesting pair, roster excludes self, post/comment edit-delete ownership-gated); uploads use a 20MB `MaxBytesReader`, content-sniffed MIME, server-generated UUID filename — no path traversal; CORS is single-origin with credentials, not wildcard; no `localStorage`/`sessionStorage` usage at all.
- **No committed build artifacts** — `git ls-files` shows no `forum-backend`/`forum-frontend`, `test-results/`, `playwright-report/`, or `.tmp/`; `.gitignore` covers them.
- **Both required `private_messages` indexes exist** (`forum_schema.sql:240-244`, SDS §4.2); no N+1 surfaced in roster/history.
- **Minor cleanup candidate (not a finding):** `Hub.SetCallbacks`/`onConnect`/`onDisconnect` (`internal/ws/connection_manager.go:44-52,69-71,91-93`) is dead — never called; presence is broadcast directly in `ws.go`.
- **Scope discipline:** all backend findings stay within the real-time-forum surface; no new bugs were raised against the untouched earlier-forum CRUD code. ARCH-05 (drafts raw SQL) is pre-existing but lives in an RTF-touched file, flagged for layering hygiene only.

---

*End of report.*
