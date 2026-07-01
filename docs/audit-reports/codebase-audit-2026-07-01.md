# Codebase Analysis & Audit Report

**Date:** 2026-07-01
**Branch:** chbaikas/codebase-audit
**Project:** real-time-forum (Go SPA + WebSockets — 01-edu real-time-forum)
**Scope:** Full SPA review + real-time-forum backend surface — 5 parallel analysis passes

---

## Methodology

Five parallel analysis passes were executed across the codebase:
1. **Bugs & Logic Errors** — SPA (core + all feature slices) and the RTF backend surface (WS/presence, DM send/delivery, roster, history pagination, migrations, auth-on-boot, reactions, notifications). Verified with `go build ./...`, `go vet ./...`, and `go test -race` across `internal/ws`, `internal/handlers`, `internal/db`, and `internal/tests`.
2. **Dead Code & Unused References** — SPA unused exports/imports, orphaned legacy `web/templates`/`web/static/js`/`web/static/css` assets, retained-but-unwired OAuth backend, stale config/tooling, and repository artifacts.
3. **Architecture, Boundary Violations & Guideline Drift** — `AGENTS.md` layering rules, SPA structure/naming conventions, WS/REST contract fidelity against `docs/SDS.md`, track-ownership consistency, and audit-question structural coverage.
4. **Code Quality & Security** — XSS/DOM-safety sweep, forbidden-tech sweep, CSP, SQL injection, session-cookie hardening, auth gating, authorization scoping, password hashing, DM/file-upload validation, WS concurrency safety, CORS.
5. **Tests & CI Gaps** — Go/SPA unit and integration coverage, E2E coverage vs. `docs/audit.md`, race-detection gating, ticket-tracker parity, audit-traceability mapping, CI pipeline gaps, test flakiness, and index/performance checks.

Each pass was evidence-driven and read-only: every finding below was independently verified by the authoring agent via direct file reads, grep evidence, or real command execution (`go build`, `go vet`, `go test -race`, `bun run test`) rather than assumed from filenames or docs. The SPA was audited in full; the Go backend was audited only across the real-time-forum surface (diffed against baseline commit `14495d3`, the first RTF planning commit) — the pre-existing earlier-forum code was treated as read-only context. Two agents (Architecture and Security) independently converged on the same WebSocket `CheckOrigin` finding, which is merged below as a cross-validated, higher-confidence item.

---

## Executive Summary

| Severity | Count |
|----------|-------|
| 🔴 Blocking | 0 |
| 🔴 Critical | 3 |
| 🟠 High | 10 |
| 🟡 Medium | 11 |
| 🟢 Low / Info | 12 |

**Top risks:**
1. **No `go test -race` gate in CI** — the WS/presence goroutine+channel subsystem is race-clean today but nothing in the pipeline would catch a future regression (CI-01).
2. **CI never runs Playwright E2E, and E2E itself only covers Track A** — the entire mandatory chat/DM/roster/history journey and the core forum create/comment flow are proven only by mocked-socket Vitest tests, never a real browser against the real backend (CI-02, CI-04).
3. **WebSocket upgrade accepts any `Origin` (`CheckOrigin` always `true`)** — a cross-site page can open an authenticated WS connection using the victim's cookie and drive `dm.send` on their behalf (ARCH-06/SEC-04, confirmed independently by two agents).
4. **Live OAuth callbacks can still mint real sessions**, contradicting the explicit SDS §6.3/PRD statement that OAuth is not part of the product and must not gate forum/chat access (ARCH-04).
5. **A large orphaned legacy multi-page frontend tree** (`web/templates/`, `web/static/js/`, `web/static/css/`, partials, sounds — ~50 files) remains fully unreachable in the repo, unlike the fully-cleaned precedent set by the B05 activity migration (DEAD-01).

---

## 1) Bugs & Logic Errors

### BUG-01: Chat composer does not react to live presence changes for the open recipient ⬆ HIGH
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: Track D (Tickets: D02, D04, D06)
- `SPA/features/chat/chat.conversation.page.js` (~L49-424, `initChatConversation`)
- `SPA/features/chat/chat.conversation.views.js` (~L76-118, `renderComposer`)
- Contrast: `SPA/features/chat/chat.roster.page.js` (~L142-149, roster *does* subscribe to `PRESENCE_UPDATE`)

**Problem:** `initChatConversation` subscribes to `chat:user-selected`, `WS_EVENTS.DM_MESSAGE`, and `WS_EVENTS.ERROR`, but never `WS_EVENTS.PRESENCE_UPDATE`/`PRESENCE_SNAPSHOT`. `state.isOnline` is captured once at selection time and baked into the composer's `disabled` markup by `renderComposer(isOnline)`. If the open recipient's presence flips while their conversation is active, the composer's enabled/disabled state never updates.
**Impact:** Violates the requirement that the composer clearly and continuously indicate send-availability (PRD §7.2, track-d.md D02 gate). Users get a late, reactive `chat.error RECIPIENT_OFFLINE` instead of a proactively disabled composer. No test exercises "conversation open, then presence flips for the open recipient" — confirmed absent in both `chat.conversation.page.test.js` and `chat_regression.test.mjs` (independently reconfirmed by Agent 5 as well).

**Fix:**
```js
documentRef.addEventListener(WS_EVENTS.PRESENCE_UPDATE, (event) => {
	const { userId, isOnline } = event?.detail ?? {};
	if (Number(userId) !== state.recipientId) return;
	state.isOnline = Boolean(isOnline);
	updateComposerEnabled(activeRoot, state.isOnline);
});
```

**Tests to add:** Vitest: open a conversation with an online user, dispatch `PRESENCE_UPDATE {userId, isOnline:false}` for that user, assert composer disables without a history refetch; dispatch `isOnline:true`, assert re-enable; verify presence for a *different* user does not touch the open composer.

---

### BUG-02: Feed pagination/filter requests are unsequenced — a stale response can overwrite a newer one ⬆ HIGH
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: Track B (Tickets: B01)
- `SPA/features/feed/feed.page.js` (~L66-109 `renderPosts`, ~L216-246 call sites)

**Problem:** Every pager/per-page/category interaction calls `renderPosts` with no `AbortController`, generation counter, or in-flight guard. Whichever `fetch` resolves last overwrites the DOM/pager state, regardless of which request was issued last.
**Impact:** Rapid pagination or filter changes (normal usage, especially on slower connections) can leave the feed showing results for a superseded selection while the URL/state already reflects the newer one.

**Fix:**
```js
let requestToken = 0;
async function renderPosts(elements, state, pager, fetchRef, windowRef) {
	const token = ++requestToken;
	const payload = await loadPosts(fetchRef, windowRef, state);
	if (token !== requestToken) return;
	// ...apply payload
}
```

**Tests to add:** Unit test resolving two `loadPosts` promises out of order (second call's fetch resolves first); assert DOM/pager reflect the last-*issued* request, not the last-*resolved* one.

---

### SEC-01: WebSocket read/write-pump goroutines have no panic recovery ⬆ HIGH
**Origin:** 1. Bugs & Logic Errors (concurrency scope) / 4. Code Quality & Security
**Files:** Ownership: Track C (Tickets: C01, C06)
- `internal/handlers/ws.go` (~L88-90 `go h.readPump(...)`/`go h.writePump(...)`, ~L92-134 `readPump`, ~L139-164 `writePump`)
- Contrast: `internal/middleware/middleware.go` (~L53-72 `Recoverer` — only covers the synchronous upgrade request, not the long-lived pump goroutines)

**Problem:** `middleware.Recoverer` wraps only the HTTP handler chain during the `/ws` upgrade request. Once `HandleWebSocket` spawns `readPump`/`writePump` and returns, those goroutines run unsupervised — `grep -n "recover()" internal/handlers/ws.go internal/ws/connection_manager.go` finds nothing.
**Impact:** A panic in either loop (a future nil-deref, a marshal panic, or any bug introduced while extending message types) crashes the entire process, taking down every connected user's session — a single malformed frame from any authenticated client is a DoS vector against the whole server.

**Fix:**
```go
func (h *WsHandler) readPump(userID int64, c *ws.Client) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in readPump user=%d: %v\n%s", userID, r, debug.Stack())
		}
		remaining := h.hub.Remove(userID, c)
		if remaining == 0 {
			h.hub.BroadcastPresenceUpdate(userID, false)
		}
		close(c.Send)
		c.Conn.Close()
	}()
	// ...
}
```
(Same pattern for `writePump`.)

**Tests to add:** Inject a deliberate panic path (behind a test-only hook) in `readPump`/`writePump` and assert the connection is torn down cleanly and the server process survives; assert `Remove`/presence-broadcast still fire on the panic path.

---

### BUG-03: In-flight post create/edit submission can force navigation after the user already left ⬆ MEDIUM
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: Track B (Tickets: B04)
- `SPA/features/post/post.page.js` (~L617-647 `handleCreateSubmit`, ~L649-685 `handleEditSubmit`)

**Problem:** Both handlers `await createPost`/`updatePost` then unconditionally call `context.navigate(...)`, with no cancellation flag or `AbortController` tied to route teardown.
**Impact:** On a slow request, a user who manually navigates away before it resolves gets yanked back to `/posts/{id}` or `/activity` once the stale request finally completes — a confusing forced-navigation race. No data corruption; UI-only.

**Fix:** Track a per-page "stale" flag set on route teardown and check it before acting on the resolved promise, or pass an `AbortSignal` tied to the route-swap teardown into the underlying fetch.

**Tests to add:** Simulate a delayed `createPost` promise, trigger teardown/navigate-away before it resolves, then resolve it, and assert `navigate` was not called after teardown.

---

### BUG-04: No re-entrancy guard on reaction clicks — rapid double-click races the toggle ⬆ MEDIUM
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: Track B (Tickets: B07)
- `SPA/features/post/post.reactions.bindings.js` (~L63-105 `handleReactionClick`)

**Problem:** Nothing disables the button or short-circuits a second click while the two `await`s (`getCurrentUser`, `postReaction`) are outstanding; a fast double-click/double-tap fires the handler twice concurrently against a toggle-semantics endpoint.
**Impact:** Transient UI/server desync for reaction counts and pressed-state until next reload; self-heals but is a real logic gap on a click-driven feature.

**Fix:**
```js
const pending = new WeakSet();
async function handleReactionClick(fetchRef, event) {
	const button = getReactionButton(event.target);
	if (!button || pending.has(button)) return;
	pending.add(button);
	try { /* existing body */ } finally { pending.delete(button); }
}
```

**Tests to add:** Fire two rapid `click` events on the same button before the mocked `postReaction` promise resolves; assert only one request is issued or that final DOM state is deterministic.

---

### BUG-05: Roster `moveToTopForMessage` silently no-ops for a message from an unknown counterpart ⬆ LOW
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: Track D (Tickets: D01, D04)
- `SPA/features/chat/chat.roster.logic.js` (~L43-63 `moveToTopForMessage`)

**Problem:** If a `dm.message` event's counterpart id isn't present in the client's in-memory roster entries (roster fetch raced/failed, or a new user registered mid-session), the function returns the list unchanged with no refetch trigger.
**Impact:** Low — self-heals on next load/reload; the message itself still renders in the open conversation. Reachable only if the initial roster fetch fails or races a new registration.

**Fix:** On `index === -1`, trigger a one-off roster refetch instead of silently dropping the update.

**Tests to add:** Unit test asserting a `dm.message` for an id absent from `state.entries` triggers `fetchRoster()`.

---

### BUG-06: Reaction "opposite button" lookup is unscoped beyond the container ⬆ LOW
**Origin:** 1. Bugs & Logic Errors
**Files:** Ownership: Track B (Tickets: B07)
- `SPA/features/post/post.reactions.bindings.js` (~L99-104)

**Problem:** `scope.querySelector('[data-reaction="..."]')` resolves the first matching descendant in the whole post/comment container, not scoped to the specific `.reactions` wrapper that was clicked. Correct today only because markup never nests reaction groups more than once per container.
**Impact:** Latent/defensive-only; fragile to future markup changes (e.g., nested comment-preview reactions).

**Fix:** Scope the query via `button.closest('[data-reaction-scope]')` before calling `querySelector`.

**Tests to add:** None required immediately; would be covered incidentally by BUG-04's fix if broader reaction-scope tests are added.

---

## 2) Dead Code & Unused References

### DEAD-01: Entire orphaned legacy multi-page frontend tree (~50 files) still tracked and served ⬆ HIGH
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track A / Track B (Tickets: A10, B04 — both `[x]` Done; B05's activity-cleanup is the established precedent for full deletion)
- `web/templates/{create-post,edit-post,home,login,register,view-post}.html` (6 files)
- `web/static/js/{auth,auth-modal,category,create-post,edit-post,header,header-loader,helpers,home,image-picker,login,logout,notifications,pagination,password-toggle,post-actions,post-form,posts,reactions,register,settings,settings-loader,sound-effects,ui-messages,utils,view-post}.js` (26 files)
- `web/static/css/*` (13 files), `web/static/partials/*` (3 files), `web/static/sounds/*` (4 files)

**What is dead/unused and why:** `cmd/frontend/routes.go` mounts `web/static` only as a raw file server (nothing `<script>`/`<link>`s into it); zero `template.ParseFiles`/`template.Must` calls anywhere load `web/templates/*.html` (the sole template load in the repo is the unrelated `web/errors/error.html`). `SPA/index.html` loads only `/assets/css/main.css` and `/main.js`. Grepping every filename in this set shows references only within the dead set itself (a self-contained dead import graph). `SPA/tests/unit/policy/spa-import-paths.test.js` already actively forbids importing these paths, confirming the project already treats this tree as legacy-quarantined — it just was never deleted, unlike the fully-cleaned B05 activity-migration precedent (`web/templates/activity.html` + `web/static/js/activity/*` were deleted in that ticket).

**Suggested removal:** Delete all five subtrees in a follow-up cleanup ticket mirroring B05. Keep `web/static/img/forum-logo.png`/`favicon.ico` if the A02-03 static-asset E2E smoke check is meant to persist; otherwise retire that gate too. Note: `internal/tests/drafts_test.go` (DEAD-03 below) must be updated first since it reads two of these files directly.

---

### DEAD-02: Orphaned images with zero live consumer ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track A / Track B (Tickets: A10, B04/B05)
- `web/static/img/forum-logo-night.png` — zero references anywhere in the repo
- `web/static/img/{camera,close}.png`, `my-activity.png`, `{github-logo,google-logo}.svg` — referenced only by the dead files in DEAD-01/DEAD-04

**What is dead/unused and why:** `grep -rn "forum-logo-night"` returns no matches at all; the others are reachable only through the already-orphaned legacy tree or the retired OAuth login page.
**Suggested removal:** Delete alongside DEAD-01; `forum-logo-night.png` can be removed immediately with zero dependency chain.

---

### DEAD-03: `internal/tests/drafts_test.go` asserts against dead legacy files ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track C (Tickets: C08)
- `internal/tests/drafts_test.go` (~L256-293, `TestCreatePostTemplate_HasImagePreviewControls`, `TestCreatePostJS_ConditionallyRendersPostImageMarkup`)

**What is dead/unused and why:** These tests read `web/templates/create-post.html` and `web/static/js/posts.js` — files that (per DEAD-01) are never served to a real user. The live SPA create/edit-post rendering (`SPA/features/post/post.views.js`/`post-card.views.js`) has its own separate coverage; these two Go tests give false assurance since they can never fail from a live-product regression.
**Suggested removal:** Delete both test functions when `web/templates`/`web/static/js` are removed (DEAD-01).

---

### DEAD-04: Retained-but-unwired OAuth backend (612 lines) ⬆ MEDIUM
**Origin:** 2. Dead Code & Unused References / 3. Architecture (cross-referenced as ARCH-04 — see security impact there; this entry captures the dead-code framing)
**Files:** Ownership: Unmappable to a specific A–D ticket by design (explicitly retained pre-RTF code per AGENTS.md/SDS §6.3); closest related ticket is Track D **D10** (removed the OAuth entry points from the UI, `[x]` Done)
- `internal/auth/oauth.go` (154 lines), `internal/handlers/oauth_google.go` (166 lines), `internal/handlers/oauth_github.go` (208 lines), `internal/handlers/oauth_helpers.go` (84 lines)

**What is dead/unused and why:** `internal/router/router.go` still registers live routes to these handlers (see ARCH-04 — this is not fully dead, since the routes are reachable), but `internal/env/env.go` defines no client ID/secret config, so the feature cannot be configured to run in this deployment, and no test references any OAuth handler. The frontend was correctly stripped of all OAuth entry points (D10 verified clean).
**Suggested removal:** Delete `internal/auth/oauth.go` and the three `oauth_*.go` handler files entirely, or — if retained deliberately per SDS §9's "OAuth tables may remain temporarily" — add an explicit top-of-file comment and de-register the routes so they cannot mint sessions (see ARCH-04 for the security framing of why leaving the *routes* live is the higher-severity part of this finding).

---

### DEAD-05: Two fully unused SPA exports ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track B (Tickets: B08, B01)
- `SPA/features/activity/activity.api.js` (~L171-176, `ACTIVITY_DEFAULTS`)
- `SPA/features/feed/feed.state.js` (~L9-15, `getDefaultFeedState`)

**What is dead/unused and why:** Grep for each symbol name across `SPA/` returns only the definition line — no importer anywhere, including tests.
**Suggested removal:** Remove both, or add a comment documenting why they're kept as a public default for future consumers.

---

### DEAD-06: Dead re-export and unused disposal-protocol method ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track B (Tickets: B03) / Track D (Tickets: D04)
- `SPA/features/post/post.views.js` (~L3, `export { renderPostDetailView } from './post-detail.views.js'`)
- `SPA/core/realtime/chat-socket.js` (~L136-139, `[Symbol.dispose]() { close(); }`)

**What is dead/unused and why:** Every real consumer of `renderPostDetailView` imports directly from `post-detail.views.js`, never through the `post.views.js` barrel. `createChatSocket`'s teardown is always an explicit `chatSocket?.close?.()` call — nothing in the codebase invokes the `using`/disposal protocol that would trigger `Symbol.dispose`.
**Suggested removal:** Drop the dead re-export line. Either wire `create-app.js` teardown through `using` to make the pattern real, or remove the unused `[Symbol.dispose]` method.

---

### DEAD-07: Duplicate `package.json` scripts, never invoked ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track A (Tickets: A01)
- `package.json` (~L13-14, `"check"`/`"fix"` — byte-identical to `"lint"`/`"lint:fix"`)

**What is dead/unused and why:** `Makefile`, `README.md`, `AGENTS.md`, and `docs/track-a.md` exclusively reference `lint`/`lint:fix`; grep for `bun run check`/`bun run fix` returns zero hits anywhere.
**Suggested removal:** Remove the `check`/`fix` aliases, or standardize on one name per operation.

---

### DEAD-08: `package-lock.json` tracked alongside `bun.lock` in a Bun-only project ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track A (Tickets: A01)
- `package-lock.json` (top-level, 56,942 bytes)

**What is dead/unused and why:** CI sets up Bun only and never runs `npm ci`/`npm install`; the only npm reference anywhere is the Makefile's emergency fallback for machines without Bun. `git log` shows it was last touched in the same commits that introduced Bun/Biome tooling — a one-time `npm install` artifact, not a maintained lockfile.
**Suggested removal:** Delete `package-lock.json` (keep `bun.lock` as sole source of truth), or explicitly document the npm-fallback duality if it's intentional.

---

### DEAD-09: Stale Vitest path alias for legacy `web/static/js` ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track A (Tickets: A01, A02)
- `vitest.config.ts` (~L5-8, `'/static/js': path.resolve(__dirname, './web/static/js')`)

**What is dead/unused and why:** No test or source file imports through this alias; the SPA's own policy test (`spa-import-paths.test.js`) actively forbids importing from that legacy path, contradicting the still-configured escape hatch.
**Suggested removal:** Remove the alias.

---

### DEAD-10: Biome lints dead legacy JS/CSS; ignore files are inconsistent ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track A (Tickets: A01)
- `biome.json` (~L10-20), `.biomeignore` (~L1-4)

**What is dead/unused and why:** `.biomeignore` excludes `web/templates/`/`web/errors/` but not `web/static/js/`/`web/static/css/`, which are equally orphaned (DEAD-01) yet still linted/formatted by `bun x biome check`, giving false confidence that dead code is "fine."
**Suggested removal:** Add `web/static/js/` and `web/static/css/` to `.biomeignore` for consistency until DEAD-01 lands.

---

### DEAD-11: Redundant `/api/v1/posts/draft*` REST endpoints (intentional, but flagged per audit mandate) ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track B (Tickets: B08)
- `internal/router/router.go` (~L79-95, `HandleDraft`/`HandleDraftByID` routes)

**What is dead/unused and why:** `docs/track-b.md` B08 notes explicitly state draft creation reuses `POST /api/v1/posts` with `status="draft"` and that these endpoints are "intentionally left unused." Confirmed zero SPA references; only `internal/tests/drafts_test.go` exercises them.
**Suggested removal:** No functional change required (explicitly accepted by the ticket), but flag for a future ticket to delete or repurpose this redundant parallel API surface.

---

### DEAD-12: Trivial `.gitignore` redundancy ⬆ LOW
**Origin:** 2. Dead Code & Unused References
**Files:** Ownership: Track A (Tickets: A01)
- `.gitignore` (~L17, `.tmp`; ~L20, `.tmp/`)

**What is dead/unused and why:** Both lines ignore the same path in bare vs. trailing-slash form.
**Suggested removal:** Collapse to a single `.tmp/` entry.

---

## 3) Architecture, Boundary Violations & Guideline Drift

### ARCH-06 / SEC-04: WebSocket upgrade accepts any `Origin` — CSWSH exposure on the cookie-auth boundary ⬆ HIGH
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift *and* 4. Code Quality & Security (independently converged — cross-validated finding)
**Violated rule:** AGENTS.md Key Design Rule #3 (session-cookie auth as the security boundary); `docs/SDS.md §5.5`: "only authenticated users may connect... session cookie is validated during upgrade."
**Files:** Ownership: Track C (Tickets: C01)
- `internal/handlers/ws.go` (~L19-25)
```go
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
```

**Problem:** Browsers attach cookies regardless of the initiating page's origin. Because `CheckOrigin` unconditionally returns `true`, a hostile page (`evil.example`) can open `new WebSocket('ws://forum-host/ws')`; the browser attaches the victim's `session_token` cookie, and the server accepts the upgrade. No test in `internal/tests/ws_test.go` exercises Origin validation.
**Impact:** Cross-Site WebSocket Hijacking (CSWSH) — the attacker's page gets a live, authenticated WS session as the victim and can drive `dm.send` on their behalf (send arbitrary DMs from the victim's account, read live presence/roster data). This is the single highest-impact finding intersecting the "session-cookie auth" design rule, since it turns the cookie-based trust model into a cross-origin-exploitable one for the one endpoint that matters most for private-messaging confidentiality.

**Fix:**
```go
CheckOrigin: func(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin == "" || origin == frontendOrigin // reuse the CORS-configured origin
},
```

**Tests to add:** WS upgrade test asserting a request with a foreign `Origin` header is rejected (`http.StatusForbidden` or connection refused), and one with the configured frontend origin succeeds.

---

### ARCH-04: Live OAuth callbacks can still mint real sessions, contradicting SDS §6.3 ⬆ HIGH
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift (cross-referenced with DEAD-04's dead-code framing of the same files)
**Violated rule:** `docs/SDS.md §6.3`: "OAuth is not part of the target product design." AGENTS.md Key Design Rule #3: "No OAuth for the real-time forum (retained code may exist but is not required)." PRD §4 Out of Scope: "OAuth login as a product requirement."
**Files:** Ownership: Unmappable to a specific A–D ticket — OAuth is explicitly cross-cutting retained code with no owning ticket
- `internal/router/router.go` (~L200-224 — registers `/auth/google`, `/auth/google/callback`, `/auth/github`, `/auth/github/callback` with no auth gate)
- `internal/handlers/oauth_google.go` (~L73-87, `GoogleCallback` calls `createSessionAndRedirect` using the same session mechanism as normal login)
- `internal/handlers/oauth_github.go` (same pattern)

**Impact:** Hitting `GET /api/v1/auth/google/callback` with a valid code produces a real `session_token` cookie that the `Auth` middleware accepts for every forum/chat endpoint — directly contradicting the SDS §6.3 rule that retained OAuth code "must not gate forum/chat access." The SPA itself never links to these routes (no `google|github|oauth` references anywhere in `SPA/`), so the normal user journey is unaffected, but the backend attack/entry surface remains fully functional and unauthenticated-gateable.

**Fix:** Remove the 4 OAuth routes/handlers/files entirely (see DEAD-04), or gate them behind a build flag/feature toggle that is off by default so they cannot silently mint sessions in the shipped product.

---

### ARCH-03: Handler layer issues raw SQL directly, bypassing `internal/db/` ⬆ HIGH
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** AGENTS.md: "Handlers (`internal/handlers/`): ... no direct `database/sql` queries — handlers call db layer functions."
**Files:** Ownership: Track B (Tickets: B08)
- `internal/handlers/drafts.go` (~L357-372, `fetchDraftImageURLByID`; called from `HandleDraftByID` at ~L249, ~L323)
```go
func fetchDraftImageURLByID(ctx context.Context, db *sql.DB, userID, draftID int64) (*string, error) {
	var imageURL sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT image_url
		FROM posts
		WHERE id = ? AND author_id = ? AND status = 'draft'
	`, draftID, userID).Scan(&imageURL)
	// ...
}
```

**Impact:** This handler-layer function imports `database/sql` and runs a raw parameterized query directly, exactly the pattern AGENTS.md forbids, while the rest of the same file correctly calls `repository.DraftGet/DraftUpdate/DraftDelete`. It duplicates the `status = 'draft'` business rule outside the db package with no single source of truth and sets a precedent other handlers could copy.

**Fix:** Move the query into `internal/db/` as `db.GetDraftImageURL(ctx, database, userID, draftID)` and call that from the handler.

---

### CI-01: `go test -race` never wired into CI or Makefile despite SDS §10.4 mandate ⬆ CRITICAL
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track C / Track D (Tickets: C08, D07); contract `docs/SDS.md §10.4`
- `Makefile` (~L92-94, `test-backend` = plain `go test ./...`, no `-race`)
- `.github/workflows/ci.yml` (~L35-36, `make test-backend`)
- Concurrency-bearing code: `internal/ws/connection_manager.go`, `internal/handlers/ws.go`

**What is missing:** SDS §10.4 explicitly requires race verification for the goroutine+channel WS/presence subsystem. No Makefile target or CI step runs `-race`. Independently re-verified this audit run by two separate agents: `go test -race -count=1 ./...` → all packages `ok`, zero data races detected (`internal/tests` 397-398s, `internal/ws` ~1s). The suite is clean today, but nothing in the automated pipeline would catch a future regression.

**Fix:**
```makefile
test-race:
	@GOCACHE=$(TEST_CACHE_DIR) GOTMPDIR=$(TEST_TMP_DIR) go test -race ./internal/ws/... ./internal/tests/... ./internal/handlers/...
```
Add a CI step calling `make test-race` (or fold `-race` directly into `test-backend`).

---

### CI-02: CI never runs E2E (Playwright); full-stack regressions ship green ⬆ CRITICAL
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track A / Track D (Tickets: A01, D07)
- `.github/workflows/ci.yml` (~L35-39, only `make test-backend` + `make test-frontend`)
- `Makefile` (~L84, `test` includes `test-e2e`, but CI never calls `make test` or `make test-e2e`)

**What is missing:** No CI job starts real servers or runs Playwright. `make test-frontend` = `bun run policy` = `biome check && vitest run`; `vitest.config.ts` correctly excludes `SPA/tests/e2e/**`. The harness itself is sound (fresh-server enforcement, port-freeing — see CI-04's clean sub-areas) — it is simply never invoked in CI.

**Fix:** Add a job/step: `bun x playwright install --with-deps chromium && make test-e2e`; upload the HTML report as an artifact.

---

### CI-04: E2E suite covers Track A only; the entire mandatory chat/forum journey is unverified in a real browser ⬆ CRITICAL
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track D (Tickets: D06, D07); `docs/SDS.md §10.2.2`; `docs/audit.md` chat questions
- `SPA/tests/e2e/tickets.test.js` (~L143-723 — `test.describe` blocks only for A02, A03, A04, A05, A06, A07, A10; 21 `test(` blocks total)

**What is missing:** No Playwright spec exercises `/posts` create/comment flows via the rendered form (auth E2E uses `page.request.post` API injection). Zero E2E coverage of: two-browser real-time DM delivery, roster ordering, message format, latest-10/`before_id` scroll pagination, offline-composer disabling, feed-vs-detail comment visibility, or a real login-form flow. The existing roster-related test only asserts DOM-region stability, not ordering/presence/messaging behavior.

**Fix:** Add Playwright specs for: two-context DM round trip, roster reorder-on-message, message bubble format, scroll-up pagination without spamming, offline-composer disabled state, feed/detail comment visibility, and a real login-form flow.

---

### CI-03: No compile check (`go build`) in CI ⬆ HIGH
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track A (Tickets: A01)
- `.github/workflows/ci.yml` (no build step)
- `Makefile` (~L36, `build` target exists, unused by CI)

**What is missing:** `go test ./...` does not build packages with no test files that aren't transitively imported by tested packages — confirmed `cmd/backend`, `cmd/qa-seed`, `internal/middleware`, `internal/router`, `internal/env` all report `[no test files]`. A compile error in any of those would not be caught by `test-backend`. `go build ./...` currently exits 0 cleanly, so this gate is non-breaking to add.

**Fix:** Add a `make build` (or `go build ./...`) step to the `quality-gate` job before/alongside tests.

---

### CI-06: No audit-traceability matrix; overstated evidence citation for roster ordering ⬆ HIGH
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track D (Tickets: D07)
- No traceability file exists; closest artifact is `docs/audit-reports/D07-final-acceptance-validation-2026-06-20.md`

**What is missing:** No file maps each `docs/audit.md` question to a proving test. The D07 report's rows for "chat users ordered by last message" / "alphabetical fallback" cite `chat.roster.logic.test.js` as evidence — independently verified that file only tests live reorder-on-new-message (`applySnapshot`/`setPresence`/`moveToTopForMessage`), **not** the initial sort order. The actual ordering logic is backend-only and well-tested there (`internal/tests/chat_roster_test.go`, including an explicit alphabetical-fallback case) — architecturally fine since the frontend trusts backend order, but the D07 report's citation for the frontend-ordering row is wrong.

**Fix:** Add `docs/audit-reports/traceability-matrix.md` (see the full table in the Tests & CI section methodology); correct the roster-ordering evidence citation in the D07 report; add a Vitest test directly asserting the roster's initial recency/alphabetical order.

---

### PARITY-01: A07/D01 marked `[x]` Done but the required roster→profile link does not exist ⬆ HIGH
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track A / Track D (Tickets: A07, D01)
- `SPA/features/chat/chat.roster.views.js` (no `href`/`data-link` attribute anywhere)
- `docs/track-a.md` (~L150-153, A07 gate: "profile is accessible from the chat roster (Blocked by D01)")
- `docs/track-d.md` (~L38, D01 work item: "add links from rostered users to their profiles... as defined in A07")

**What is missing:** Both tickets are `[x]` in `docs/ticket-tracker.md`. The only profile-link E2E test (`A07-02`) exercises `.profile-link` from post-author links only — there is no roster-originated profile link in the roster view code, no Vitest test for it, and no E2E test for it. The D07 acceptance report also only cites "reachable from author links," silently dropping the roster-link half of the requirement; the prior 2026-06-30 codebase audit did not catch this either.

**Fix:** Add a profile anchor/`data-link` to each roster row in `chat.roster.views.js` pointing at `/profile/:id`; add a Vitest assertion and an E2E case exercising it from the roster.

---

### ARCH-01: Documented `OptionalAuth` middleware stage does not exist ⬆ MEDIUM
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** AGENTS.md: "Request flow: Logger → Recoverer → CORS → OptionalAuth → Auth → Handler."
**Files:** Ownership: Unmappable to a single ticket — pure documentation drift (no ticket specifies an `OptionalAuth` stage)
- `internal/router/router.go` (~L337-341); `internal/middleware/` package (absence confirmed — zero matches for `OptionalAuth`)

**Impact:** `internal/middleware/auth.go` only defines a single hard-reject `Auth` middleware applied per-route; there is no soft/optional variant. Pure doc drift with no functional bug, previously flagged in a C08 PR message ("Removed `OptionalAuth`... dead code") — AGENTS.md was never updated after the code changed.

**Fix:** Update AGENTS.md's middleware convention line to describe the actual per-route `Auth`-only model, or reinstate a real `OptionalAuth` stage if soft-auth semantics are still intended anywhere.

---

### ARCH-02: Middleware wrap order does not match documented Logger → Recoverer → CORS sequence ⬆ MEDIUM
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** AGENTS.md: "Request flow: Logger → Recoverer → CORS → OptionalAuth → Auth → Handler."
**Files:** Ownership: Unmappable to a single ticket — closest are A05, C01
- `internal/router/router.go` (~L337-341)
```go
func addMiddlewares(handler http.Handler, frontendOrigin string) http.Handler {
	handler = middleware.EnableCORS(frontendOrigin)(handler)
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)
	return handler
}
```

**Impact:** Actual execution order is Recoverer → Logger → CORS → mux, not the documented Logger → Recoverer → CORS. Recoverer being outermost is arguably safer than the doc implies, but it's still a real drift from the written contract.

**Fix:** Either reorder to `Logger(Recoverer(EnableCORS(...)))` to match AGENTS.md, or update AGENTS.md to state the actual order.

---

### SEC-02: Session cookie is missing the `Secure` attribute ⬆ MEDIUM
**Origin:** 4. Code Quality & Security
**Files:** Ownership: Track C (Tickets: C11)
- `internal/handlers/users.go` (~L135-141 Register, ~L193-199 Login, ~L234-241 Logout clear)
- `internal/middleware/auth.go` (~L66-75, `clearSessionCookie`)

**Impact:** All four `http.Cookie{...}` constructions set `HttpOnly: true`/`SameSite: Lax` but never `Secure: true`. In any deployment reachable over both HTTP and HTTPS, the browser still attaches `session_token` to plaintext requests, letting a network-level attacker (e.g., public Wi-Fi, SSL-stripping) capture the token and hijack the session.

**Fix:**
```go
http.SetCookie(w, &http.Cookie{
	Name:     "session_token",
	Value:    session.Token,
	Path:     "/",
	HttpOnly: true,
	Secure:   env.IsProduction(), // or an explicit COOKIE_SECURE env var
	SameSite: http.SameSiteLaxMode,
})
```

---

### SEC-03: No Content-Security-Policy anywhere (SPA meta tag or backend header) ⬆ MEDIUM
**Origin:** 4. Code Quality & Security
**Files:** Ownership: Track A (Tickets: A10, A02)
- `SPA/index.html` (no CSP meta tag)
- `cmd/frontend/routes.go` (~L70-84, `setNoStoreHeaders` sets cache-control only)

**Impact:** The SPA's escaping discipline is solid today (see "Verified Safe" notes below), but there is zero defense-in-depth. If a future change introduces an unescaped interpolation, there is no CSP to contain the blast radius.

**Fix:** Add a CSP meta tag to `SPA/index.html` (`default-src 'self'; script-src 'self'; connect-src 'self' ws: wss:; object-src 'none'; ...`), or better, emit `Content-Security-Policy` as a response header from the frontend server's catch-all handler.

---

### ARCH-09: Legacy `create-post.html` remains a live Go-test dependency, blocking full B05-style cleanup ⬆ LOW
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift (consolidates with DEAD-01/DEAD-03 — same underlying files, kept separate here for the architecture-conformance angle)
**Files:** Ownership: Track B (Tickets: B04, B08)
- `web/templates/create-post.html`, `web/static/js/create-post.js` (still contains hard `window.location.href` navigations per the tracker's own note)
- `internal/tests/drafts_test.go` (~L257, hardcodes a read of `../../web/templates/create-post.html`)

**Impact:** `web/templates/create-post.html` cannot be deleted without first migrating this Go test — an undocumented gap in the tracker's own B05-completeness notes.

**Fix:** See DEAD-03's fix; migrate the test off the legacy fixture, then delete the legacy files per DEAD-01.

---

### ARCH-05: `internal/db` bootstrap functions violate ctx-first-parameter convention ⬆ LOW
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** AGENTS.md: "Persistence layer (`internal/db/`): ... All functions accept `context.Context` as first parameter."
**Files:** Ownership: Track C (Tickets: C07)
- `internal/db/db.go` (~L34, `InitDB(dbPath string)` — no context param, manufactures its own `context.Background()` internally; ~L98, `ApplyQASeeds(db *sql.DB)`)

**Impact:** Process-bootstrap-only functions (called once at startup), so practical risk is low, but a literal deviation from the stated rule with no documented bootstrap exception.

**Fix:** Accept `ctx context.Context` as the first parameter in both; have callers pass an explicit context rather than the function manufacturing one internally.

---

### ARCH-07: `Authorization: Bearer` fallback in Auth middleware deviates from cookie-only auth rule ⬆ LOW
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** AGENTS.md: "Session-cookie auth — HttpOnly cookies. No JWT." `docs/SDS.md §6.3`: "the cookie remains the only auth mechanism."
**Files:** Ownership: Track A / Track C (Tickets: A05, C01)
- `internal/middleware/auth.go` (~L31-37)

**Impact:** Not a JWT (same opaque session token, just header-transportable), so unreachable from the actual SPA (confirmed no `Authorization` header sent anywhere in `SPA/core/api/`), but it is a live, testable secondary auth path SDS §6.3 says shouldn't exist.

**Fix:** Remove the Bearer fallback, or explicitly document/restrict it to non-production builds.

---

### ARCH-08: `private_messages`/profile schema `DEFAULT` values deviate from SDS §4.1 literal DDL ⬆ LOW
**Origin:** 3. Architecture, Boundary Violations & Guideline Drift
**Violated rule:** `docs/SDS.md §4.1` (no `DEFAULT` clauses specified for `age`/`gender`/`first_name`/`last_name`)
**Files:** Ownership: Track C (Tickets: C10, C07)
- `internal/db/forum_schema.sql` (~L30-33), mirrored in `internal/db/migrate.go` (~L38-41)

**Impact:** Deliberate and correct for migration safety (sanctioned by SDS §8 and enforced at the API layer by C11 registration validation), so no functional defect — purely a literal-text deviation from §4.1's DDL.

**Fix:** None required functionally; optionally reconcile SDS §4.1 wording to acknowledge the migration-safe defaults.

---

## 4) Code Quality & Security

### SEC-04: See **ARCH-06 / SEC-04** above (merged — WebSocket `CheckOrigin` CSWSH). Independently found by both the Architecture and Security passes.

### SEC-05: No upper bound on registration `age` ⬆ LOW
**Origin:** 4. Code Quality & Security
**Files:** Ownership: Track C (Tickets: C11)
- `internal/db/users_helpers.go` (~L48-50, `case req.Age <= 0:`)

**Security impact:** Not directly exploitable, but a user can register with `age: 2147483647`, later displayed on their public profile — a data-quality gap rather than a vulnerability.
**Fix:** `case req.Age <= 0 || req.Age > 150:`

---

### SEC-06: No global `unhandledrejection` handler in the SPA ⬆ LOW
**Origin:** 4. Code Quality & Security
**Files:** Ownership: Track A / Track D (Tickets: A03, D07)
- `SPA/main.js` (full file — only `DOMContentLoaded`, no `window.addEventListener('unhandledrejection', ...)`)

**Security impact:** Not a vulnerability, but an unhandled rejection anywhere silently logs to console with no telemetry hook or user-facing fallback, making production issues harder to diagnose.
**Fix:**
```js
window.addEventListener('unhandledrejection', (event) => {
	console.error('Unhandled promise rejection:', event.reason);
});
```

---

### SEC-07: Neither HTTP server configures read/write/idle timeouts (Slowloris exposure) ⬆ LOW
**Origin:** 4. Code Quality & Security
**Files:** Ownership: Track A / Track C (Tickets: A01, C01 — both servers wired at the infra layer touched by RTF work)
- `cmd/backend/server.go` (~L70, bare `http.ListenAndServe(addr, handler)`)
- `cmd/frontend/main.go` (~L19, bare `http.ListenAndServe(":3000", mux)`)

**Security impact:** Neither server sets `ReadTimeout`/`WriteTimeout`/`IdleTimeout` (no `http.Server{}` struct is constructed at all — both call the package-level `http.ListenAndServe` helper, which uses Go's zero-value defaults, i.e. no timeouts). A client that opens a connection and sends headers/body slowly (or never completes a request) can hold a connection — and the goroutine serving it — open indefinitely, and a moderate number of such connections can exhaust server resources (a classic Slowloris-class DoS). This is independent of the RTF changes themselves but applies to the exact process entry points the RTF work wires up (`internal/router.NewRouter` in `server.go`, the proxy/static mux in the frontend).
**Suggested fix:**
```go
srv := &http.Server{
	Addr:         addr,
	Handler:      handler,
	ReadTimeout:  10 * time.Second,
	WriteTimeout: 30 * time.Second, // longer if large DM-image uploads need more headroom
	IdleTimeout:  120 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```
Apply the same pattern to the frontend's `main.go`. Note the WebSocket `/ws` connection is long-lived by design — if timeouts are added at the `http.Server` level, verify they don't clash with the hijacked WS connection's own deadline management in `internal/handlers/ws.go`'s pumps (which already set their own read/write deadlines independently).

---

### Verified safe (no findings) — Code Quality & Security
- **XSS/DOM sinks:** Every dynamic interpolation site across the SPA (DM bodies, usernames, post/comment content, profile fields, activity, roster, notifications) routes through a single shared, tested `escapeHTML()` helper (`SPA/core/utils/html.js`). No exploitable XSS found, including against adversarial `<img onerror=...>`-style test fixtures.
- **Forbidden tech:** zero `var`, `require(`, `XMLHttpRequest`, React/Vue/Angular/jQuery, or dynamic-content inline event handlers anywhere in shipped SPA code.
- **Storage trust boundary:** the SPA does not use `localStorage`/`sessionStorage` at all — auth state is entirely HttpOnly-cookie-based, so no trust-boundary read-back is possible.
- **SQL injection:** every RTF-surface query uses `?` placeholders; the only `fmt.Sprintf`-built SQL (`internal/db/migrate.go`, `ALTER TABLE`/`PRAGMA table_info`) is fed exclusively from a hardcoded column-name slice literal, never request input.
- **Authorization scoping:** chat history and roster queries are correctly scoped to the authenticated caller; no cross-user data leakage found.
- **DM send fail-closed validation (SDS §6.1):** all required checks (recipient exists, not self, non-empty body, recipient online) implemented in the correct order with distinct `chat.error` codes; all `TestDMSend_*` tests pass.
- **Password hashing:** bcrypt with `DefaultCost`; `PasswordHash` is `json:"-"` tagged, cannot leak via any response.
- **File upload (DM images):** content-type sniffed server-side (not trusted from client headers), size-capped, server-generated UUID filenames, path-traversal-safe; `dm.send`'s `image_url` is additionally re-validated by a strict regex plus explicit `..` rejection.
- **WS concurrency safety:** every access to the connection map is mutex-guarded; a real (non-cached) `go test -race` run across `internal/ws`, `internal/handlers`, `internal/tests` completed clean with zero race warnings.
- **CORS:** configured to a specific origin (never `*`) paired with `Allow-Credentials: true` — the only spec-legal combination.
- **Auth gating / no guest access:** every forum-content and chat route is wrapped in `Auth`; only register/login/static/OAuth-legacy endpoints are unauthenticated.
- **Session validation on WS upgrade:** session cookie is validated via `db.GetSessionByToken` before `upgrader.Upgrade` is ever called; expired/missing/invalid tokens are rejected pre-upgrade.
- **Registration validation:** non-positive age and empty required profile fields are rejected with 400 before any DB write.
- **Migration safety:** idempotent, additive-only, safe defaults; no data-loss path found.

---

## 5) Tests & CI Gaps

(CI-01, CI-02, CI-03, CI-04, CI-06, and PARITY-01 are detailed above in the Architecture section per severity ordering; remaining findings below.)

### CI-08: `internal/middleware` and `internal/router` have zero unit tests ⬆ MEDIUM
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track A / Track C (Tickets: A05, C01)
- Confirmed via `ls internal/middleware/*_test.go internal/router/*_test.go` → none exist; `go test -race` output confirms `[no test files]` for both packages.

**What is missing:** The `Auth` middleware (session validation chokepoint, injects `userID`, gates WS upgrade) has no isolated table-driven test for expired/malformed/missing-cookie edge cases. It is exercised transitively through `internal/tests/`, so behavior isn't unverified, just not isolated.
**Fix:** Add `internal/middleware/auth_test.go` with table-driven valid/expired/invalid/missing-cookie cases plus context-injection assertions.

---

### CI-07: No Vitest coverage thresholds or coverage measurement ⬆ MEDIUM
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track A / Track D (Tickets: A01, D05, D06)
- `vitest.config.ts` (no `coverage` block); `package.json` (`"test": "vitest run"`, no `--coverage`)

**What is missing:** No coverage provider/thresholds means a coverage regression (e.g., a new feature slice shipped with zero tests) is invisible to CI. E2E exclusion is correctly configured; there's simply no ratchet on the included set.
**Fix:** Add `test.coverage: { provider: 'v8', include: ['SPA/**'], exclude: ['SPA/tests/**'], thresholds: { lines: 75, functions: 75, branches: 70, statements: 75 } }` and run `vitest run --coverage` in CI.

---

### DOC-01: Stale `docs/audit-reports/audit-e2e-coverage-2026-04-18.md` remains uncorrected/unmarked ⬆ MEDIUM
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track D (Tickets: D07)
- `docs/audit-reports/audit-e2e-coverage-2026-04-18.md` (entire file)

**What is missing:** Dated 2026-04-18, states most tickets "Not Started," and contains hard-coded paths from a different contributor's machine (`file:///home/ertval/...`). All 37 tickets are now `[x]` Done per the tracker, yet this file was never superseded, retitled, or flagged stale in-place, sitting alongside the current 06-30 and 06-20 reports with no cross-reference. An auditor skimming the directory chronologically would be misled.
**Fix:** Add a superseded/stale banner at the top of the file, or delete it now that later reports supersede it.

---

### CI-09: No `go vet` in CI ⬆ LOW
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track A (Tickets: A01)
- `.github/workflows/ci.yml` (lint = biome only); `Makefile` has a `vet` target never invoked anywhere.

**Fix:** Add `make vet` to the quality-gate job.

---

### CI-10: CI trigger branches `[main, dev]` may miss integration-branch PRs ⬆ LOW
**Origin:** 5. Tests & CI Gaps
**Files:** Ownership: Track A (Tickets: A01)
- `.github/workflows/ci.yml` (~L3-7)

**What is missing:** This repo actively uses long-lived feature/integration branches merged directly to `main`; a PR whose base is one of those branches never triggers CI.
**Fix:** Broaden the `pull_request.branches` filter or standardize all feature PRs to target `main`/`dev` directly.

---

### Verified safe / clean (no findings) — Tests & CI
- **Backend §10.1 coverage:** every SDS §10.1 requirement (extended registration, login-by-username/email, auth-gating rejection, message persistence, latest-10 history, `before_id` pagination, roster ordering with/without history, offline-recipient rejection, presence connect/disconnect, multi-connection presence) is backed by a named test function.
- **`private_messages` indexes (SDS §4.2):** both required indexes present verbatim.
- **N+1/efficiency:** roster and history queries are single-query, no N+1 found.
- **Playwright flakiness:** zero `waitForTimeout` calls; all waits are state-driven. `reuseExistingServer: false` and Makefile port-freeing confirmed unchanged and correct.
- **`git diff --exit-code` format gate:** reliable — confirmed working tree clean and `biome.json` VCS integration correctly configured.
- **Frontend §10.2 requirements:** route transitions, auth gating, logout visibility, feed/detail comment split, activity mount/refresh, disabled composer at selection time, live message rendering, throttle/debounce with scroll-position preservation — all independently confirmed present; `bun run test` passes 332/332 across 35 files.

---

## Cross-Reference: Finding ID Mapping

| Consolidated ID | Agent 1 (Bugs) | Agent 2 (Dead Code) | Agent 3 (Arch) | Agent 4 (Security) | Agent 5 (Tests/CI) | Track Ownership | Description |
|----------------|---------|---------|---------|---------|---------|-----------------|-------------|
| CI-01 | — | — | — | — | CI-01 | Track C/D | `go test -race` not gated in CI |
| CI-02 | — | — | — | — | CI-02 | Track A/D | CI never runs Playwright E2E |
| CI-04 | — | — | — | — | CI-04 | Track D | E2E covers Track A only |
| ARCH-06/SEC-04 | — | — | ARCH-06 | SEC-04 | — | Track C | WS `CheckOrigin` always true (CSWSH) — cross-validated by 2 agents |
| ARCH-04 | — | DEAD-04 (dead-code angle) | ARCH-04 | — | — | Unmappable (retained code) | Live OAuth callbacks mint real sessions |
| DEAD-01 | — | DEAD-01 | — | — | — | Track A/B | Orphaned legacy multi-page frontend tree |
| BUG-01 | BUG-01 | — | — | — | (cross-ref, SEC/BUG note) | Track D | Composer ignores live presence changes |
| BUG-02 | BUG-02 | — | — | — | — | Track B | Unsequenced feed pagination requests |
| SEC-01 | (concurrency scope) | — | — | SEC-01 | — | Track C | No panic recovery in WS pump goroutines |
| ARCH-03 | — | — | ARCH-03 | — | — | Track B | Handler issues raw SQL directly |
| CI-03 | — | — | — | — | CI-03 | Track A | No `go build` compile check in CI |
| CI-06 | — | — | — | — | CI-06 | Track D | No audit-traceability matrix; overstated roster-ordering citation |
| PARITY-01 | — | — | — | — | PARITY-01 | Track A/D | A07/D01 roster→profile link missing despite `[x]` status |
| DEAD-02 | — | DEAD-02 | — | — | — | Track A/B | Orphaned images |
| DEAD-03 | — | DEAD-03 | (ARCH-09 overlap) | — | — | Track C | Test asserts against dead legacy files |
| DEAD-04 | — | DEAD-04 | (ARCH-04 overlap) | — | — | Unmappable | Retained OAuth backend, 612 dead lines |
| BUG-03 | BUG-03 | — | — | — | — | Track B | Post submit/navigate race |
| BUG-04 | BUG-04 | — | — | — | — | Track B | Reaction click re-entrancy |
| ARCH-01 | — | — | ARCH-01 | — | — | Unmappable | `OptionalAuth` doc/code drift |
| ARCH-02 | — | — | ARCH-02 | — | — | Unmappable | Middleware order doc/code drift |
| SEC-02 | — | — | — | SEC-02 | — | Track C | Session cookie missing `Secure` |
| SEC-03 | — | — | — | SEC-03 | — | Track A | No CSP anywhere |
| CI-08 | — | — | — | — | CI-08 | Track A/C | No unit tests for middleware/router |
| CI-07 | — | — | — | — | CI-07 | Track A/D | No Vitest coverage thresholds |
| DOC-01 | — | — | — | — | DOC-01 | Track D | Stale E2E coverage report unmarked |
| ARCH-09 | — | (DEAD-01/03 overlap) | ARCH-09 | — | — | Track B | Legacy create-post test dependency |
| BUG-05 | BUG-05 | — | — | — | — | Track D | Roster no-ops for unknown counterpart |
| BUG-06 | BUG-06 | — | — | — | — | Track B | Reaction opposite-button lookup unscoped |
| ARCH-05 | — | — | ARCH-05 | — | — | Track C | DB bootstrap funcs lack ctx-first param |
| ARCH-07 | — | — | ARCH-07 | — | — | Track A/C | Bearer-token auth fallback |
| ARCH-08 | — | — | ARCH-08 | — | — | Track C | Schema `DEFAULT` vs SDS §4.1 literal DDL |
| SEC-05 | — | — | — | SEC-05 | — | Track C | No upper bound on registration age |
| SEC-06 | — | — | — | SEC-06 | — | Track A/D | No `unhandledrejection` handler |
| SEC-07 | — | — | — | SEC-07 | — | Track A/C | No HTTP server read/write/idle timeouts (Slowloris exposure) |
| CI-09 | — | — | — | — | CI-09 | Track A | No `go vet` in CI |
| CI-10 | — | — | — | — | CI-10 | Track A | CI branch triggers may miss integration branches |
| DEAD-05 | — | DEAD-05 | — | — | — | Track B | Two unused SPA exports |
| DEAD-06 | — | DEAD-06 | — | — | — | Track B/D | Dead re-export + unused disposal method |
| DEAD-07 | — | DEAD-07 | — | — | — | Track A | Duplicate package.json scripts |
| DEAD-08 | — | DEAD-08 | — | — | — | Track A | package-lock.json alongside bun.lock |
| DEAD-09 | — | DEAD-09 | — | — | — | Track A | Stale Vitest alias |
| DEAD-10 | — | DEAD-10 | — | — | — | Track A | Biome lints dead legacy code |
| DEAD-11 | — | DEAD-11 | — | — | — | Track B | Redundant draft REST endpoints (accepted) |
| DEAD-12 | — | DEAD-12 | — | — | — | Track A | .gitignore redundancy |

---

## Recommended Fix Order

### Phase 1 — Blocking & Critical (must fix before any merge)
1. **CI-01**: Wire `go test -race` into CI/Makefile for the WS/presence subsystem (Track C/D)
2. **CI-02**: Add a CI job that installs Playwright and runs `make test-e2e` (Track A/D)
3. **CI-04**: Expand the E2E suite beyond Track A to cover the mandatory chat/DM/roster/history/forum journey (Track D)

### Phase 2 — High Severity (immediate follow-up)
4. **ARCH-06/SEC-04**: Fix WS `CheckOrigin` to validate against the configured frontend origin (Track C)
5. **ARCH-04**: Remove or feature-flag the live OAuth callback routes so they cannot mint sessions (Unmappable/retained code)
6. **ARCH-03**: Move `fetchDraftImageURLByID`'s raw SQL into `internal/db/` (Track B)
7. **CI-03**: Add a `go build ./...` compile-check step to CI (Track A)
8. **CI-06**: Build the audit-traceability matrix and correct the D07 roster-ordering citation (Track D)
9. **PARITY-01**: Add the missing roster→profile link (code) plus tests for A07/D01 (Track A/D)
10. **BUG-01**: Make the chat composer reactive to live presence changes for the open recipient (Track D)
11. **BUG-02**: Sequence feed pagination/filter requests to prevent stale-response overwrite (Track B)
12. **SEC-01**: Add panic recovery to the WS read/write-pump goroutines (Track C)
13. **DEAD-01**: Plan and execute deletion of the orphaned legacy frontend tree (Track A/B)

### Phase 3 — Medium Severity
14. **ARCH-01/ARCH-02**: Reconcile AGENTS.md middleware documentation with actual code (Unmappable)
15. **SEC-02**: Add `Secure` attribute to session cookies for HTTPS deployments (Track C)
16. **SEC-03**: Add a CSP meta tag / header (Track A)
17. **BUG-03**: Guard post create/edit submission against post-navigation races (Track B)
18. **BUG-04**: Add re-entrancy guard to reaction click handler (Track B)
19. **CI-08**: Add unit tests for `internal/middleware`/`internal/router` (Track A/C)
20. **CI-07**: Add Vitest coverage thresholds (Track A/D)
21. **DOC-01**: Mark or delete the stale E2E coverage report (Track D)
22. **DEAD-02/DEAD-03/DEAD-04**: Clean up orphaned images, legacy-dependent test, and dead OAuth backend (Track A/B/C)

### Phase 4 — Low Severity (maintenance)
23. **ARCH-05/ARCH-07/ARCH-08/ARCH-09**: DB bootstrap ctx param, Bearer fallback removal, schema doc reconciliation, legacy create-post test migration (Track A/C)
24. **SEC-05/SEC-06**: Age upper bound, global `unhandledrejection` handler (Track A/C/D)
25. **CI-09/CI-10**: Add `go vet` to CI, broaden CI branch triggers (Track A)
26. **BUG-05/BUG-06**: Roster refetch-on-unknown-id, reaction scope tightening (Track B/D)
27. **DEAD-05 through DEAD-12**: Unused exports, dead re-export, duplicate scripts, stale lockfile/alias/ignore-file cleanup (Track A/B)
28. **SEC-07**: Add explicit `ReadTimeout`/`WriteTimeout`/`IdleTimeout` to both `http.Server` instances (Track A/C)

---

## Notes

- No **Blocking** findings were raised — nothing in this audit prevents the current codebase from functioning correctly for its documented use cases. The three **Critical** findings are all test/CI-gate gaps (no `-race` gate, no E2E in CI, thin E2E coverage), not functional defects in the shipped product; they represent real risk of *undetected future regressions* rather than a currently-broken feature.
- The RTF backend surface is architecturally solid where it counts most: `go build`/`go vet` are clean, `go test -race` passes with zero races across the WS/presence subsystem (independently re-run by three separate agents), SQL injection surface is clean, DM fail-closed validation is complete and fully tested, and file-upload validation is sound.
- Two agents (Architecture and Security) independently converged on the same WebSocket `CheckOrigin` finding without cross-referencing each other's work — this is treated as a higher-confidence, cross-validated finding rather than a duplicate to discard.
- A prior audit report (`docs/audit-reports/codebase-audit-2026-06-30.md`, dated the day before this one) covers overlapping ground; every finding in this report was independently re-verified against the current tree rather than copied forward, and several new findings (BUG-01 through BUG-06, ARCH-03, ARCH-05, ARCH-07 through ARCH-09, PARITY-01, DOC-01, CI-06's specific citation error) were newly surfaced in this pass.
- The `docs/ticket-tracker.md` shows all 37 tickets as `[x]` Done; this audit found one concrete case (PARITY-01) where a ticket's own verification gate is not actually satisfied by the shipped code, and recommends treating "Done" status as a starting point for verification rather than a substitute for it.

---

*End of report.*
