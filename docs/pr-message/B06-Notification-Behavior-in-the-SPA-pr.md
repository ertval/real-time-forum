# B06: Notification Behavior in the SPA
<!-- Filename: docs/pr-message/B06-Notification-Behavior-in-the-SPA-pr.md -->

This PR closes ticket `B06` by migrating the notification system from the legacy
standalone scripts into the SPA runtime. The notification bell now lives in the
shared authenticated SPA shell and preserves the existing notification behavior —
5-second polling, the unread badge and count, mark-one-read and mark-all-read,
and notification-click deep-link navigation — while removing the legacy
`window.location.href` hard navigation in favor of the SPA router. The backend
notification API contracts are reused unchanged. Toast and sound feedback from the
legacy UI are intentionally out of scope for this ticket.

## Summary of Changes

### 1. SPA Notification Feature Migration
- **Notification SPA Feature Modules**: Added the notification feature slice
  following the established `{feature}.api.js` / `{feature}.views.js` /
  `{feature}.page.js` separation:
  - `SPA/features/notification/notification.api.js` — REST client for
    `GET /api/v1/notifications`, `PATCH /api/v1/notifications/{id}/read`, and
    `PATCH /api/v1/notifications/read-all`, returning uniform
    `{ ok, status, ... }` results
  - `SPA/features/notification/notification.views.js` — bell/badge/dropdown
    markup, item markup carrying `data-notification-{type,post-id,comment-id}`
    for stateless destination resolution, and the message formatter
  - `SPA/features/notification/notification.page.js` — `createNotificationCenter`
    lifecycle controller (polling, badge/dropdown render, mark-read, click →
    navigate)
  - `SPA/features/notification/notification.css`
- **Persistent Shell Mounting**: Mounted the bell in the persistent authenticated
  shell (`SPA/features/shell/shell.views.js`) so it survives client-side route
  changes; only `.app-shell__outlet` is swapped on navigation, leaving the bell
  and its running poll interval intact.

### 2. Polling Lifecycle & Read Actions
- **Polling Lifecycle**: Wired the 5-second polling lifecycle through
  `SPA/core/app/create-app.js` — started after authenticated boot and after
  login `onSuccess`, and stopped on logout, app teardown, and any `401`
  response. A `pollHandle` guard prevents duplicate intervals; the controller
  re-binds if the shell (and bell DOM) is rebuilt on a logout → login cycle.
- **Unread Badge & Count**: Badge reflects the server `unread_count`; mark
  actions update it optimistically (`max(n-1, 0)` / `0`) and the next poll
  reconciles from server truth.
- **Mark-One / Mark-All Read**: Single read `PATCH`es
  `/api/v1/notifications/{id}/read`; mark-all `PATCH`es
  `/api/v1/notifications/read-all`. Both update local state immediately, handle
  failures gracefully (no optimistic change on failure), and leave reconciliation
  to the next poll.

### 3. Notification-Click Deep-Link Navigation
- **Reload-Free Navigation**: Replaced the legacy
  `window.location.href = '/view-post/{id}?highlight=…'` navigation
  (`web/static/js/notifications.js`) with SPA router navigation. Clicking a
  notification marks it read (if unread) and routes to
  `/posts/{id}?highlight={comment_id|last}` without a full document reload;
  already-read notifications still navigate.
- **Comment Deep-Link Highlighting**: Added
  `SPA/features/post/comment-highlight.js` and `comment-highlight.css`, called
  from `initPostDetailPage` after comments render. It reads the `highlight`
  query param, resolves the legacy `#comment-{id}` intent against the SPA
  `[data-comment-id]` markup, supports `highlight=last`, retries for
  asynchronously-rendered comments, scrolls the target into view, and applies a
  temporary visual highlight.

### Out-of-Scope Note
- Toast and sound feedback from the legacy
  `web/static/js/notifications.js` (the `uiNotify` toast and `playNotification`
  sound on new notifications) were **not** migrated. The B06 work list scopes to
  polling, badge/count, mark-read, and click/deep-link navigation; toast and
  sound are intentionally deferred.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **B06**:
> - notification polling works from the SPA shell
> - unread badge/count behavior is preserved
> - mark-read and mark-all-read still work
> - notification click behavior still navigates correctly

## Testing & Validation Verified

### Automated Test Suite
- [x] `npx vitest run` — 127/127 tests passing, including the notification
  feature-slice suites
  (`SPA/tests/unit/features/notification/notification.{api,page,views}.test.js`),
  the comment-highlight suite
  (`SPA/tests/unit/features/post/comment-highlight.test.js`), and the polling
  lifecycle assertions added to
  `SPA/tests/unit/core/app/create-app.test.js`
- [x] `bun run policy` — Biome and Vitest frontend policy gate passing
- [x] `bun x biome check SPA/features SPA/core` — clean
- [x] SPA import-boundary policy tests
  (`SPA/tests/unit/policy/spa-import-paths.test.js`) — passing (confirms no
  module re-introduces legacy `/web/static/*` imports or
  `DOMContentLoaded`-style standalone boot logic)

### QA Checklist
- [x] Verified the notification bell renders in the authenticated SPA shell and
  is absent on public routes (`/login`, `/register`)
- [x] Verified polling starts after authenticated boot and after login, and
  stops on logout and on `401`
- [x] Verified polling survives client-side navigation between protected routes
  without creating duplicate intervals
- [x] Verified the unread badge and count update on poll and on mark actions
- [x] Verified mark-one-read (`PATCH /api/v1/notifications/{id}/read`) and
  mark-all-read (`PATCH /api/v1/notifications/read-all`)
- [x] Verified notification click marks read and navigates via the SPA router
  (no `window.location.href`)
- [x] Verified deep-link highlighting for `highlight={commentId}` and
  `highlight=last` on the post detail view
- [x] Verified browser back/forward replays through the SPA router without a
  full document reload
- [x] Verified the SPA notification flow uses no full-page reloads

### Automated Behavioral Verification (E2E)
- [ ] Real-browser Playwright coverage for the bell → click → mark-read →
  deep-link → highlight chain is not yet added; recommended under D05
  (SPA and Forum Frontend Regression Coverage). Current coverage is
  unit/integration with faithful DOM mocks.

## Audit Traceability

- Ticket: B06 - Notification Behavior in the SPA
- Source: RTF-29
- Phase: P2
- Verification Gate: PASS
- Backend Changes: None
- API Contract Changes: None
- Migration Type: Legacy → SPA

## Key Files Impacted
- `SPA/features/notification/notification.api.js` *(new)*
- `SPA/features/notification/notification.views.js` *(new)*
- `SPA/features/notification/notification.page.js` *(new)*
- `SPA/features/notification/notification.css` *(new)*
- `SPA/features/post/comment-highlight.js` *(new)*
- `SPA/features/post/comment-highlight.css` *(new)*
- `SPA/features/shell/shell.views.js` *(mount notification bell)*
- `SPA/features/post/post.page.js` *(highlight after comments render)*
- `SPA/core/app/create-app.js` *(polling lifecycle + inject router navigate)*
- `SPA/assets/css/main.css` *(import notification + comment-highlight CSS)*
- `SPA/tests/unit/features/notification/notification.api.test.js` *(new — 5 tests)*
- `SPA/tests/unit/features/notification/notification.page.test.js` *(new — 26 tests)*
- `SPA/tests/unit/features/notification/notification.views.test.js` *(new — 6 tests)*
- `SPA/tests/unit/features/post/comment-highlight.test.js` *(new — 5 tests)*
- `SPA/tests/unit/core/app/create-app.test.js` *(polling lifecycle tests)*
- `architecture.md`
- `docs/ticket-tracker.md`
- `docs/track-b.md`

## Follow-up Notes
- **Toast/sound (deferred):** porting the legacy `uiNotify` toast and
  `playNotification` sound (and copying `web/static/sounds/notification.mp3`
  into `SPA/assets/sounds/`) can be picked up as a small follow-up if product
  wants full parity with the legacy notification UX.
- **E2E coverage:** add a Playwright scenario for the full notification
  navigation + highlight chain under D05.
- **Idle polling:** polling continues while the tab is hidden (matches legacy);
  an optional `visibilitychange` pause could reduce idle requests.

## Regression Review

Verified:

- No backend regressions
- No API contract regressions
- No authentication regressions
- No routing regressions
- No full page reloads introduced
- Existing notification behavior preserved
