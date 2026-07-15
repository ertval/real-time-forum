# Deduplicated & Validated Codebase Audit Report

**Date:** 2026-07-14
**Branch:** main
**Project:** real-time-forum (Go SPA + WebSockets — 01-edu real-time-forum)
**Scope:** Consolidation, deduplication, cross-check, and independent re-validation of the four `codebase-audit-*` reports dated 2026-06-26, 2026-06-30, 2026-07-01, and 2026-07-02.

---

## Methodology

This report does **not** re-audit the codebase from scratch. It:

1. **Deduplicates** the four source reports. Their ~135 raw findings collapse into **51 canonical findings** (many appear in 2–4 reports under different IDs; several single-report entries are unique). One source finding (06-26 `DEAD-07`) was already retracted-as-false-positive in its own report and is not carried forward.
2. **Cross-checks** the reports against each other, surfacing every place they disagree on existence or severity — including one contradiction *inside a single report* (07-01's OAuth finding).
3. **Validates** every canonical finding against the **current working tree** (not the state when each report was written). Seven domain agents read the actual cited code — locating it by content, since line numbers had drifted — and returned a verdict + current evidence for each finding.
4. **Adversarially re-checks** the disputed and highest-stakes findings with two independent skeptics instructed to refute the primary verdict in *either* direction (confirm a claimed ghost, or debunk a claimed real bug). All adversarial re-checks agreed with the primary verdicts.

**Verdict vocabulary:** `CONFIRMED` (real as described) · `PARTIAL` (real but overstated / mis-scoped / wrong-severity) · `ACCEPTED` (real but explicitly sanctioned by the design docs) · `GHOST` (false) · `FIXED` (real when written, since resolved).

---

## Executive Summary

### Validated severity counts (this report)

| Severity | Count |
|----------|-------|
| 🔴 Blocking | 0 |
| 🔴 Critical | 0 |
| 🟠 High | 0 |
| 🟡 Medium | 13 |
| 🟢 Low | 37 |
| ⚪ Info | 1 |
| **Total** | **51** |

**Verdict distribution:** 42 CONFIRMED · 6 PARTIAL (overstated) · 3 ACCEPTED (by design) · **0 GHOST** · 0 FIXED.

### Key results

1. **No ghosts.** Every deduplicated finding corresponds to real code. The two findings that appeared in only one source report and that I explicitly treated as ghost-risks both validated as **real**:
   - `cmd/frontend/backend_validator.go` (07-02 only) is genuinely dead — no caller anywhere, including the `cmd/frontend` test files 07-02 itself suspected.
   - The missing chat-roster→profile link (07-01 `PARITY-01` only) is genuinely absent — `chat.roster.views.js` has no `href`/`data-link`/`/profile`, despite tickets A07 and D01 being marked `[x]` Done.

2. **The source reports systematically inflated severity.** Every Blocking / Critical / High claim was downgraded on validation. There are **no** surviving Blocking, Critical, or High findings.
   - 07-02 rated the roster-hidden issue **Blocking**, its sole hard justification being "breaks Playwright test A04-03." That justification is **false**: A04-03 never opens a conversation, so it never exercises (and would never fail from) the roster-hiding code. → **Medium.**
   - 06-30 / 07-01 rated the WebSocket `CheckOrigin` issue **High** (exploitable CSWSH). The session cookie is `SameSite=Lax`, which browsers do not attach to cross-site WS handshakes, and the handler re-validates the cookie before upgrade — the attack does not work. → **Low** (defense-in-depth gap).
   - 06-30 / 07-01 rated the CI gaps **Critical**. They are real but process-only, and the chat features they "don't verify" are covered by the Vitest suite. → **Medium.**
   - 07-01 rated the handler raw-SQL and WS-panic issues **High**; both downgraded (parameterized/scoped SQL with no impact → Low; no demonstrated untrusted panic path → Medium).

3. **Three findings are ACCEPTED-by-design**, not defects: the WebSocket non-reconnect (explicitly deferred by ticket D04), the unused `/posts/draft*` endpoints (documented as intentional in track-b.md), and the schema `DEFAULT` clauses (sanctioned by SDS §8 migration strategy).

4. **The 07-01 internal contradiction is resolved.** Its `ARCH-04` ("OAuth callback mints a real session") and its own `DEAD-04` ("OAuth cannot be configured to run") are *both half-right*: credentials **are** read from `os.Getenv` (so it *can* be configured), but a session is minted only after a genuine provider token exchange with a valid code + state cookie (so it is **not** attacker-forgeable). Net: dead-but-wired code, **Low**, not exploitable.

### What is genuinely worth acting on first (all Medium)

The real, non-cosmetic gaps cluster into two themes:

- **Live-chat correctness drift** — the open conversation composer never reacts to live presence (`BUG-COMPOSER-PRESENCE`), newly-online users never join a loaded roster (`BUG-ROSTER-NEW-USER`), scroll-up history dead-ends on a transient failure (`BUG-HISTORY-DEADEND`), and the roster is hidden while a conversation is open against an explicit "visible at all times" requirement (`BUG-ROSTER-HIDDEN`).
- **Verification gaps** — CI runs no `-race`, no E2E, and no build/compile step (`CI-PIPELINE`); E2E covers only Track A (`CI-E2E-TRACK-A-ONLY`); and one "Done" deliverable (roster→profile link) was never implemented (`CI-PARITY-ROSTER-PROFILE-LINK`).

Everything else is Low/Info — dead code, doc drift, and hardening nits.

---

## Cross-Report Contradiction Resolutions

The primary value of this consolidation is settling the places the four reports disagree. Each was resolved by reading the current code.

| # | Contradiction | Reports in conflict | Resolution (verified against current code) |
|---|---|---|---|
| 1 | **Roster hidden when a conversation opens** — Blocking vs Low | 07-02 (Blocking) vs 06-26 (Low) | **Medium.** Roster *is* hidden (`chat.conversation.page.js:87`), and `docs/requirements.md:57` does say "must be visible at all times" — so it is a real spec deviation (not Low). But 07-02's "breaks A04-03" is **false**: A04-03 (`tickets.test.js:398-447`) only navigates routes; it never clicks a roster row, so the hiding path is never hit and the test does not fail. Recoverable via the back arrow → not Blocking. |
| 2 | **WS `CheckOrigin` = true** — High CSWSH vs Medium | 06-30 / 07-01 (High) vs 06-26 (Medium) | **Low.** `CheckOrigin` does return `true` (`ws.go:22-24`), but `session_token` is `SameSite=Lax`, which browsers do not send on cross-site (script-initiated) WS handshakes, and `HandleWebSocket` re-validates the cookie via `GetSessionByToken` before upgrade (`ws.go:54-58`). CSWSH is doubly mitigated → hardening gap only. |
| 3 | **OAuth backend retained** — "callback mints sessions" vs "can't be configured" | 07-01 `ARCH-04` vs 07-01 `DEAD-04` (same report!) | **Low, not exploitable.** Credentials *are* read from `os.Getenv(GOOGLE_CLIENT_ID/_SECRET/_REDIRECT_URL)` (so it can be configured), but a session mints only after a real Google/GitHub token exchange with a valid code + matching `oauth_state` cookie (`oauth_helpers.go:37-49,51-84`). Dead-but-wired code; not attacker-forgeable. |
| 4 | **WS pumps lack `recover()`** — High DoS vs Medium | 07-01 (High) vs 06-26 (Medium) | **Medium.** Both pumps genuinely run without `recover()` and Recoverer can't catch a goroutine panic → whole-process crash *if* one occurs. But 07-01's "reachable from untrusted `dm.send`" is unsubstantiated: `handleDMSend` guards unmarshal/nil/marshal and the channel-close ordering is race-safe. Real robustness gap, no demonstrated trigger. |
| 5 | **Handler raw SQL (`fetchDraftImageURLByID`)** — High vs Low | 07-01 (High) vs 06-30 (Low) | **Low.** The raw `SELECT` in the handler layer is real (`drafts.go:357-372`) and breaks the layering rule, but it is fully parameterized, authorization-scoped, and side-effect-free. Severity tracks consequence → Low. |
| 6 | **Legacy `web/` tree** — High vs Medium vs Low | 07-01 (High) / 06-30 (Medium) vs 06-26 (Low) | **Low.** The whole legacy MPA tree is dead/unreferenced, but the `/static/` mount must **stay** (the live SPA loads DM images from `/static/uploads/dm/`). Delete the dead files, keep the mount + `web/static/uploads` + `web/errors`. |
| 7 | **`backend_validator.go` dead** — only 07-02 reported it | (unique to 07-02) | **CONFIRMED real, Low.** Every exported symbol is unreferenced repo-wide, including the `cmd/frontend/*_test.go` files. Not a ghost. |
| 8 | **Roster→profile link missing (PARITY-01)** — only 07-01 reported it | (unique to 07-01) | **CONFIRMED real, Medium.** No `href`/`data-link`/`/profile` in `chat.roster.views.js`; the link genuinely does not exist despite A07/D01 marked Done. Not a ghost, and it did not "exist all along" via another path. |

---

## Findings by Domain

Severity shown is the **validated** severity. "Sources" lists the originating IDs across the four reports (`26`=06-26, `30`=06-30, `01`=07-01, `02`=07-02).

### 1) Bugs & Logic Errors

| ID | Sev | Verdict | Finding | Sources | Current location |
|----|-----|---------|---------|---------|------------------|
| BUG-COMPOSER-PRESENCE | 🟡 Med | CONFIRMED | Open conversation composer never subscribes to `PRESENCE_UPDATE`/`SNAPSHOT`; `isOnline` frozen at selection time, so it goes stale when the open recipient's presence flips. | 26·BUG-01, 30·BUG-01, 01·BUG-01 | `chat.conversation.page.js:360-386,192-201`; `chat.conversation.views.js:76-118` |
| BUG-ROSTER-HIDDEN | 🟡 Med | PARTIAL | Online-users roster is hidden while a DM thread is open (`showConversationPanel` sets `hidden`), deviating from requirements.md "visible at all times." **Not Blocking** — does not break A04-03. | 26·ARCH-02, 02·BUG-01 | `chat.conversation.page.js:85-88,215`; `shell.css:229` |
| BUG-ROSTER-NEW-USER | 🟡 Med | CONFIRMED | Roster fetched once on mount; `setPresence`/`moveToTopForMessage` never insert an unknown user, so someone who registers/comes online after load stays invisible until reload. | 26·BUG-04, 30·BUG-03, 01·BUG-05 | `chat.roster.logic.js:29-38,43-63`; `chat.roster.page.js:166-169` |
| BUG-HISTORY-DEADEND | 🟡 Med | CONFIRMED | `fetchConversation` returns `{messages:[],hasMore:false}` for *every* failure; `loadOlderHistory` then sets `hasMore=false`, permanently dead-ending scroll-up history after one transient blip. | 26·BUG-02 | `chat.conversation.api.js:3,23-35`; `chat.conversation.page.js:127,152` |
| BUG-NO-401-RECOVERY | 🟡 Med | CONFIRMED | No shared API layer downgrades the session on a post-boot 401; expiry mid-session leaves a half-authenticated app (dead WS, empty feed, poller stopped, no redirect). Per-flow post-form redirects exist but are not global. | 30·BUG-02 | `create-app.js:389`; `notification.page.js:303` |
| BUG-FEED-UNSEQUENCED | 🟡 Med | CONFIRMED | Feed pager/filter has no `AbortController`/generation guard; the last fetch to resolve wins, so a stale slow response overwrites a newer one. (Was High → Med: transient, self-correcting.) | 01·BUG-02 | `feed.page.js:66,226-251` |
| BUG-ORPHANED-ISLOADING | 🟢 Low | CONFIRMED | `loadOlderHistory` returns on the conversation-switch guard before resetting the captured `state.isLoading`. Benign today (GC'd); latent if state is ever cached/restored. | 26·BUG-03 | `chat.conversation.page.js:126,139-141` |
| BUG-ROSTER-IMG-PREVIEW | 🟢 Low | CONFIRMED | Roster last-message preview is blank for image-only DMs; the `latest` CTE never selects `image_path` and the frontend only updates preview for non-empty bodies. | 26·BUG-05 | `messages.go:118-127`; `chat.roster.logic.js:56-59` |
| BUG-WS-NO-RECONNECT | 🟢 Low | **ACCEPTED** | No `close` listener on the chat socket, so a server-side close leaves it permanently dead until reload. **Explicitly deferred by ticket D04** (`chat-socket.js:12-13`, `track-d.md:91,99`). | 26·BUG-06 | `chat-socket.js:83-129` |
| BUG-POST-NAV-RACE | 🟢 Low | CONFIRMED | Create/edit submit `await`s then unconditionally `navigate()`s with no teardown guard; a user who leaves mid-request gets yanked back. UI-only. | 01·BUG-03 | `post.page.js:617,640,646,684` |
| BUG-REACTION-REENTRANCY | 🟢 Low | CONFIRMED | No in-flight guard on reaction clicks; a rapid double-click fires two concurrent toggles → transient count/pressed desync until reload. | 01·BUG-04 | `post.reactions.bindings.js:63` |
| BUG-REACTION-SCOPE | 🟢 Low | CONFIRMED | Opposite-reaction lookup is rooted at the whole `[data-post-id]` container, not the `.reactions` wrapper; correct today only by DOM ordering. Latent. | 01·BUG-06 | `post.reactions.bindings.js:100` |
| BUG-REACTION-CONTRACT | 🟢 Low | CONFIRMED | Reactions are server-authoritative on a `click`-bound `<button>`, whereas SDS §7.4.2 specifies a delegated `change` on a checkbox with optimistic-flip-and-revert. Spec/code mismatch, no runtime bug. | 30·BUG-05 | `post.reactions.bindings.js:120,88-104` |
| BUG-FRAME-DROP | 🟢 Low | CONFIRMED | All server→client sends use `select{…default:}` drop-on-full (64-slot buffer); a saturated client silently loses `chat.error`/presence frames. Documented "drop rather than block" tradeoff. | 30·BUG-04 | `ws.go:315-324`; `connection_manager.go:138-195,25` |

### 2) Security & Hardening

| ID | Sev | Verdict | Finding | Sources | Current location |
|----|-----|---------|---------|---------|------------------|
| SEC-NO-CSP | 🟡 Med | CONFIRMED | No CSP (meta or header), and no `X-Frame-Options`/`X-Content-Type-Options` anywhere. Defense-in-depth gap; **not** a live XSS (all sinks use `escapeHTML`). | 26·SEC-02, 30·SEC-01, 01·SEC-03, 02·SEC-01 | `index.html:1-39`; `routes.go:80-84` |
| SEC-WS-PANIC | 🟡 Med | PARTIAL | WS read/write pump goroutines have no `recover()`; Recoverer can't catch them → a panic crashes the whole process. No demonstrated untrusted trigger (was High). | 26·SEC-04, 01·SEC-01 | `ws.go:88-89,92-164`; `middleware.go:53-72` |
| SEC-WS-CHECKORIGIN | 🟢 Low | PARTIAL | `CheckOrigin` returns `true` (no Origin validation). CSWSH neutralized by `SameSite=Lax` + in-handler session check (was High). | 26·SEC-03, 30·SEC-03, 01·SEC-04 | `ws.go:19-25` |
| SEC-COOKIE-SECURE | 🟢 Low | CONFIRMED | All six `session_token` set-cookie sites omit `Secure`. Deployment-dependent; dev topology is http-only so `Secure` would break it (was Med). | 26·SEC-05, 30·SEC-06, 01·SEC-02 | `users.go:135-241`; `auth.go:67-74`; `oauth_helpers.go:27-76` |
| SEC-INLINE-ONCLICK | 🟢 Low | CONFIRMED | The sole inline handler in the SPA (`onclick="window.history.back()"` in the profile "not found" branch). Static literal → no XSS; blocks a strict `script-src` (was Med). | 26·SEC-01, 30·SEC-02 | `profile.page.js:76` |
| SEC-CSS-URL-ESCAPER | 🟢 Low | CONFIRMED | `escapeHTML` (HTML escaper) used inside a CSS `url()` string. Very low exploitability — `imageUrl` is a server-generated upload path. | 30·SEC-05 | `post-card.views.js:155` |
| SEC-UNHANDLEDREJECTION | 🟢 Low | CONFIRMED | No global `unhandledrejection`/`error` handler; fire-and-forget rejections vanish. Observability gap. | 26·SEC-06, 30·SEC-04, 01·SEC-06 | `main.js:1-23` |
| SEC-AGE-UPPER-BOUND | 🟢 Low | CONFIRMED | Registration validates only `age <= 0`; any large `int` is accepted and shown on the public profile. Data-quality gap. | 01·SEC-05 | `users_helpers.go:49` |
| SEC-HTTP-TIMEOUTS | 🟢 Low | CONFIRMED | Both servers use bare `http.ListenAndServe` with no `Read/Write/Idle` timeouts (Slowloris exposure). | 01·SEC-07 | `cmd/backend/server.go:70`; `cmd/frontend/main.go:19` |

### 3) Architecture & Guideline Drift

| ID | Sev | Verdict | Finding | Sources | Current location |
|----|-----|---------|---------|---------|------------------|
| ARCH-CORE-STATE | 🟢 Low | CONFIRMED | Mandated Proxy-based `core/state` layer (SDS §7.0) is unimplemented — dir holds only `.gitkeep`, zero `new Proxy` in SPA. State is ad-hoc per feature (was Med). | 26·ARCH-01 | `SPA/core/state/.gitkeep`; `SDS.md:444` |
| ARCH-WS-ROUTE-BYPASS | 🟢 Low | PARTIAL | `/ws` is registered via bare `mux.HandleFunc`, skipping the `Auth` wrapper — but the handler re-validates the cookie before upgrade. Acceptable-by-design for WS; consistency nit (was Med). | 02·ARCH-01 | `router.go:303`; `ws.go:44-58` |
| ARCH-HANDLER-RAW-SQL | 🟢 Low | CONFIRMED | `fetchDraftImageURLByID` runs raw `SELECT` in the handler layer (breaks db/handler separation). Parameterized, scoped, no impact (was High). | 30·ARCH-05, 01·ARCH-03 | `drafts.go:357-372` |
| ARCH-BEARER-FALLBACK | 🟢 Low | CONFIRMED | `Auth` falls back to an `Authorization: Bearer <token>` header (opaque session token). Non-cookie path the design says shouldn't exist; SPA never uses it. | 30·ARCH-04, 01·ARCH-07 | `auth.go:31-37` |
| ARCH-MIDDLEWARE-DRIFT | 🟢 Low | CONFIRMED | Actual middleware order is Recoverer→Logger→CORS (not the documented Logger→Recoverer→CORS), and `OptionalAuth` doesn't exist. Pure doc drift (was Med). | 30·ARCH-01, 01·ARCH-01, 01·ARCH-02 | `router.go:337-342`; `AGENTS.md:99` |
| ARCH-ARCHITECTURE-MD | 🟢 Low | CONFIRMED | `architecture.md` says chat is "in active development" and lists OAuth as a supported auth method; both contradict the shipped/complete state (was Med). | 30·ARCH-02 | `architecture.md:70,81` |
| ARCH-README-DRIFT | 🟢 Low | CONFIRMED | README status block says "Active Development (Wave 3) … Real-Time Chat: In Progress (Proxy ready)" — chat is done and "Proxy ready" is false. | 30·ARCH-03 | `README.md:14,18` |
| ARCH-GOOGLE-FONTS | 🟢 Low | CONFIRMED | SPA shell loads fonts from `fonts.googleapis.com`/`fonts.gstatic.com` at runtime (offline/privacy/CSP concern). | 26·ARCH-03 | `index.html:12-16` |
| ARCH-DB-CTX-PARAM | 🟢 Low | CONFIRMED | `InitDB(dbPath)` and `ApplyQASeeds(db)` don't take `context.Context` first, against the persistence-layer convention. Bootstrap-only. | 01·ARCH-05 | `db.go:34,98` |
| ARCH-SCHEMA-DEFAULTS | ⚪ Info | **ACCEPTED** | `age/gender/first_name/last_name` carry `DEFAULT` clauses absent from SDS §4.1 DDL — but explicitly sanctioned by SDS §8 migration strategy and enforced at the API layer. | 30·ARCH-06, 01·ARCH-08 | `forum_schema.sql:30-33`; `migrate.go:38-41` |

### 4) Dead Code & Unused References

| ID | Sev | Verdict | Finding | Sources | Current location |
|----|-----|---------|---------|---------|------------------|
| DEAD-LEGACY-WEB-TREE | 🟢 Low | CONFIRMED | Entire legacy MPA tree (6 `web/templates`, ~26 `web/static/js`, css, partials, sounds, legacy img) is tracked and unreferenced by any live handler/SPA module. **Keep** the `/static/` mount + `web/static/uploads` + `web/errors` (was High). | 26·DEAD-01, 30·DEAD-01/02/04/05/06, 01·DEAD-01/02, 02·DEAD-01/02 | `web/templates/*`, `web/static/{js,css,partials,sounds,img}/*` |
| DEAD-OAUTH | 🟢 Low | PARTIAL | 4 OAuth routes + handlers retained and wired but not surfaced by the SPA. Configurable via env, but a session mints only after a real provider exchange — **not attacker-forgeable** (was High). | 26·DEAD-02, 30·DEAD-03, 01·DEAD-04/ARCH-04, 02·DEAD-04 | `router.go:206-239`; `oauth_*.go`; `internal/auth/oauth.go` |
| DEAD-BACKEND-VALIDATOR | 🟢 Low | CONFIRMED | `cmd/frontend/backend_validator.go` — every exported symbol unreferenced repo-wide (incl. the `cmd/frontend` test files). Genuinely dead (single-report, **not a ghost**). | 02·DEAD-03 | `cmd/frontend/backend_validator.go:14-79` |
| DEAD-DRAFTS-FIXTURE | 🟢 Low | CONFIRMED | Two Go tests read dead legacy fixtures (`web/templates/create-post.html`, `web/static/js/posts.js`), coupling the suite to dead files and blocking their removal (was Med). | 30·DEAD-09, 01·DEAD-03/ARCH-09 | `drafts_test.go:256-293` |
| DEAD-WS-HUB-TESTONLY | 🟢 Low | CONFIRMED | `Hub.SetCallbacks`/`GetOnlineUserIDs`/`GetConnectionCount` are exercised only by the unit test; the `onConnect`/`onDisconnect` callbacks stay `nil` in production. | 26·DEAD-05 | `connection_manager.go:49,99,116` |
| DEAD-DRAFT-ENDPOINTS | 🟢 Low | **ACCEPTED** | `/api/v1/posts/draft*` routes exist but the SPA creates drafts via `POST /posts` with `status="draft"`. Documented as intentional in track-b.md B08. | 01·DEAD-11 | `router.go:79-95`; `track-b.md:281-284` |
| DEAD-VITEST-ALIAS | 🟢 Low | CONFIRMED | `vitest.config.ts` aliases `/static/js` → `./web/static/js` (dead bundle); no module resolves through it and the policy test forbids it. | 26·DEAD-03, 30·DEAD-08, 01·DEAD-09 | `vitest.config.ts:8` |
| DEAD-UNUSED-SPA-EXPORTS | 🟢 Low | CONFIRMED | `ACTIVITY_DEFAULTS` and `getDefaultFeedState` are exported with zero importers anywhere in `SPA/`. | 26·DEAD-04, 30·DEAD-10, 01·DEAD-05 | `activity.api.js:171`; `feed.state.js:9` |
| DEAD-REEXPORT-DISPOSE | 🟢 Low | CONFIRMED | Dead `renderPostDetailView` re-export in the `post.views.js` barrel (all consumers import directly); `[Symbol.dispose]` never triggered by any `using` (only a direct unit-test call). | 01·DEAD-06 | `post.views.js:3`; `chat-socket.js:137` |
| DEAD-BIOME-IGNORES | 🟢 Low | CONFIRMED | `biome.json` ignores untracked binaries `forum-backend`/`forum-frontend`; `.biomeignore` excludes `web/templates` but not the equally-dead `web/static/{js,css}`. | 26·DEAD-06, 01·DEAD-10 | `biome.json`; `.biomeignore` |
| DEAD-PKG-TOOLING | 🟢 Low | CONFIRMED | Bundle (all true): `package.json` `"main":"index.js"` (no such file) + duplicate `check`≡`lint`/`fix`≡`lint:fix`; `Makefile` `PORT=8080` twice; `package-lock.json` tracked beside `bun.lock`; `.gitignore` has both `.tmp` and `.tmp/`. | 30·DEAD-11/12/13, 01·DEAD-07/08/12, 02·DEAD-05 | `package.json:5,11-14`; `Makefile:20,135`; `.gitignore:17,20` |
| DEAD-E2E-LEGACY-ASSETS | 🟢 Low | CONFIRMED | E2E test `A02-03` asserts HTTP 200 for legacy `/static/{css/base.css,js/auth.js,img/forum-logo.png}`, pinning dead assets as served and blocking cleanup (was Med). | 30·DEAD-07 | `tickets.test.js:185-212` |

### 5) Tests & CI Gaps

| ID | Sev | Verdict | Finding | Sources | Current location |
|----|-----|---------|---------|---------|------------------|
| CI-PIPELINE | 🟡 Med | CONFIRMED | Bundle — **all 7 facets verified true**: (1) no `go test -race`; (2) no Playwright E2E in CI; (3) no `make build`/compile check; (4) no Vitest coverage thresholds; (5) `vet` target exists but is never called; (6) format gate mutates-then-diffs (`biome check --write --unsafe` + `git diff`); (7) triggers reference a non-existent `dev` branch. Process-only (was Critical). | 26·CI-01/03/07/08/09, 30·CI-01/02/03/07/09/10, 01·CI-01/02/03/07/09/10, 02·CI-01 | `.github/workflows/ci.yml:5-39`; `Makefile`; `vitest.config.ts` |
| CI-E2E-TRACK-A-ONLY | 🟡 Med | CONFIRMED | Playwright covers only Track A (A02–A10); no DM send, real-time delivery, presence, roster reorder, offline-composer, or throttled scroll; auth via API injection not the form. Covered by Vitest, not E2E (was Critical). | 26·CI-05, 30·CI-04, 01·CI-04 | `tickets.test.js:143,39,57` |
| CI-CONN-MGR-CONCURRENCY-TEST | 🟡 Med | CONFIRMED | `connection_manager_test.go` has only sequential tests; with no concurrent exerciser, `-race` can't surface a presence-map race at the unit level. | 26·CI-02 | `connection_manager_test.go:13-117` |
| CI-MIDDLEWARE-UNIT-TESTS | 🟡 Med | CONFIRMED | No `_test.go` in `internal/middleware` or `internal/router`; the `Auth` chokepoint has no isolated unit test (only indirect coverage). | 30·CI-08, 01·CI-08 | `internal/middleware/`, `internal/router/` |
| CI-PARITY-ROSTER-PROFILE-LINK | 🟡 Med | CONFIRMED | Roster→profile link (A07 gate / D01 work item, both `[x]` Done) does not exist — no `href`/`data-link`/`/profile` in the roster view. Real "Done-but-absent" gap; profiles still reachable via author links (was High). | 01·PARITY-01 | `chat.roster.views.js:22-33`; `ticket-tracker.md:99,121` |
| CI-TRACEABILITY | 🟢 Low | PARTIAL | The 2026-04-18 coverage report is stale (foreign `file:///home/ertval/` paths, "Not Started"), and the D07 report over-cites `chat.roster.logic.test.js` for *initial* roster ordering (that test only covers live reorder). But a Q→test mapping *does* exist (D07 §2), so "no matrix exists" is overstated (was High). | 26·CI-06, 30·CI-05/06, 01·CI-06/DOC-01 | `audit-e2e-coverage-2026-04-18.md`; `D07-...-2026-06-20.md:47-48` |

---

## Severity Reconciliation (source claims → validated)

Only findings whose validated severity differs from the highest source claim are listed.

| Finding | Source max | Validated | Why corrected |
|---|---|---|---|
| BUG-ROSTER-HIDDEN | Blocking (02) | **Medium** | 07-02's "breaks A04-03" is false; recoverable spec deviation. |
| CI-PIPELINE | Critical (30/01) | **Medium** | Real but process-only; features covered elsewhere. |
| CI-E2E-TRACK-A-ONLY | Critical (30/01) | **Medium** | Chat covered by Vitest; E2E-shape gap, not untested behavior. |
| SEC-WS-CHECKORIGIN | High (30/01) | **Low** | `SameSite=Lax` + in-handler session check neutralize CSWSH. |
| SEC-WS-PANIC | High (01) | **Medium** | Real gap; no demonstrated untrusted panic path. |
| ARCH-HANDLER-RAW-SQL | High (01) | **Low** | Parameterized, scoped, no impact — convention breach only. |
| DEAD-LEGACY-WEB-TREE | High (01) | **Low** | Inert dead files; mount must stay. |
| DEAD-OAUTH | High (01·ARCH-04) | **Low** | Not attacker-forgeable; dead-but-wired. |
| BUG-FEED-UNSEQUENCED | High (01) | **Medium** | Transient, self-correcting display race. |
| CI-PARITY-ROSTER-PROFILE-LINK | High (01) | **Medium** | Real gap; profiles reachable via author links. |
| CI-TRACEABILITY | High (01) | **Low** | A Q→test map exists (D07 §2); doc-hygiene + one bad citation. |
| BUG-COMPOSER-PRESENCE | High (26/30/01) | **Medium** | Real stale-UI defect; recoverable by reselecting. |
| SEC-NO-CSP | High (30/01) | **Medium** | Defense-in-depth gap; escaping currently correct, no live XSS. |
| BUG-POST-NAV-RACE, BUG-REACTION-REENTRANCY | Medium (01) | **Low** | Narrow, self-healing, UI-only. |
| SEC-COOKIE-SECURE, SEC-INLINE-ONCLICK | Medium (30/01) | **Low** | Deployment-dependent / static-literal, no exploit. |
| ARCH-WS-ROUTE-BYPASS | Medium (02) | **Low** | Equivalent in-handler auth; by-design for WS. |
| ARCH-MIDDLEWARE-DRIFT, ARCH-ARCHITECTURE-MD, ARCH-CORE-STATE | Medium (30/01/26) | **Low** | Doc/architecture drift, no runtime defect. |
| DEAD-DRAFTS-FIXTURE, DEAD-E2E-LEGACY-ASSETS | Medium (30/01) | **Low** | Test-hygiene, currently passing, no runtime risk. |
| ARCH-SCHEMA-DEFAULTS | Low (30/01) | **Info (Accepted)** | Sanctioned by SDS §8. |

**Retracted-at-source (not carried forward):** 06-26 `DEAD-07` (claimed test-only SPA exports were dead) was already retracted in that report as a false positive — the exports run in production via same-module composition. Confirmed still not a finding.

---

## Recommended Fix Order

No Blocking / Critical / High items. Order below is by validated severity and effort.

**Tier 1 — Medium, product-facing (fix first):**
1. `BUG-COMPOSER-PRESENCE` — subscribe the open conversation to `PRESENCE_UPDATE`/`SNAPSHOT` and re-derive the composer.
2. `BUG-ROSTER-NEW-USER` — debounced roster refetch when a presence/dm event references an unknown user.
3. `BUG-HISTORY-DEADEND` — distinguish a transient fetch failure from genuine end-of-history; don't clear `hasMore` on error.
4. `BUG-ROSTER-HIDDEN` — show roster + conversation side-by-side on wide viewports; keep the toggle only for narrow widths.
5. `BUG-NO-401-RECOVERY` — centralize 401 handling (downgrade session, stop chat/notifications, redirect to `/login`).
6. `BUG-FEED-UNSEQUENCED` — add an `AbortController`/generation token to `renderPosts`.

**Tier 2 — Medium, verification (fix alongside):**
7. `CI-PIPELINE` — add `go build ./...`, a `-race` target, a Playwright E2E job, Vitest coverage thresholds, `go vet`; make the format gate read-only; drop the `dev` trigger.
8. `CI-E2E-TRACK-A-ONLY` — add a two-context chat E2E (DM round-trip, presence, roster reorder, offline composer, throttled scroll).
9. `CI-PARITY-ROSTER-PROFILE-LINK` — add the roster→profile link (and a test), or correct the A07/D01 "Done" status.
10. `CI-CONN-MGR-CONCURRENCY-TEST`, `CI-MIDDLEWARE-UNIT-TESTS` — add a concurrent Add/Remove stress test under `-race` and isolated `Auth` middleware tests.

**Tier 3 — Medium, hardening:**
11. `SEC-NO-CSP` — emit CSP + `X-Frame-Options` + `nosniff` from the frontend server (allowlist the Google Fonts origins or self-host per `ARCH-GOOGLE-FONTS`).
12. `SEC-WS-PANIC` — add `defer/recover()` to both WS pumps.

**Tier 4 — Low, cleanup (batch):**
13. Dead code: delete the legacy `web/` tree (keep the `/static/` mount, `web/static/uploads`, `web/errors`), the `A02-03` E2E test, the two `drafts_test.go` fixture tests, the Vitest alias, unused SPA exports, the dead re-export, and de-wire/remove OAuth routes — in the correct order (remove the tests/aliases *before* the files).
14. Tooling: `package.json` `main` + dup scripts, `Makefile` PORT, `package-lock.json`, `.gitignore` `.tmp`, `.biomeignore` gaps.
15. Hardening nits: `SEC-WS-CHECKORIGIN` (Origin allowlist), `SEC-COOKIE-SECURE` (conditional `Secure`), `SEC-INLINE-ONCLICK`, `SEC-CSS-URL-ESCAPER`, `SEC-UNHANDLEDREJECTION`, `SEC-AGE-UPPER-BOUND`, `SEC-HTTP-TIMEOUTS`.
16. Doc/convention: `ARCH-MIDDLEWARE-DRIFT`, `ARCH-ARCHITECTURE-MD`, `ARCH-README-DRIFT`, `ARCH-CORE-STATE`, `ARCH-BEARER-FALLBACK`, `ARCH-HANDLER-RAW-SQL`, `ARCH-DB-CTX-PARAM`, `CI-TRACEABILITY`, `ARCH-WS-ROUTE-BYPASS`.
17. Remaining Low bugs: `BUG-ORPHANED-ISLOADING`, `BUG-ROSTER-IMG-PREVIEW`, `BUG-POST-NAV-RACE`, `BUG-REACTION-REENTRANCY`, `BUG-REACTION-SCOPE`, `BUG-REACTION-CONTRACT`, `BUG-FRAME-DROP`.

**No action needed (Accepted-by-design):** `BUG-WS-NO-RECONNECT` (D04 scope), `DEAD-DRAFT-ENDPOINTS` (track-b.md B08), `ARCH-SCHEMA-DEFAULTS` (SDS §8). Consider adding an explicit code comment so future audits don't re-flag them.

---

## Notes on Verification

- **Every finding was read against the current working tree** (post the 06-30/07-01/07-02 audit merges), not the state each report was written against. Nothing had been silently fixed in the interim — all 51 remain present, and no source claim turned out to be a ghost.
- **Confirmed-safe surfaces** (re-verified, consistent with all four source reports): SQL is parameterized throughout the RTF surface; user-content render paths route through `escapeHTML`; sessions are `HttpOnly` and validated on every protected request and on the WS upgrade; passwords use bcrypt; DM send is fail-closed; uploads are content-sniffed with UUID filenames; CORS is single-origin with credentials; and `go test -race` passes clean today (the gap is that CI never runs it).
- **Adversarial agreement:** the two independent skeptic passes over the 9 disputed/high-stakes findings agreed with all primary verdicts, including the two single-report findings (`DEAD-BACKEND-VALIDATOR`, `CI-PARITY-ROSTER-PROFILE-LINK`) that were most at risk of being ghosts.

---

*End of report.*
