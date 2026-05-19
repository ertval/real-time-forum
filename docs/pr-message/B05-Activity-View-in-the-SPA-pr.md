# B05: Activity View in the SPA
<!-- Filename: docs/pr-message/B05-Activity-View-in-the-SPA-pr.md -->

This PR closes ticket `B05` by migrating the Activity screen from the legacy
standalone template architecture into the SPA runtime. `/activity` now renders
inside the shared authenticated SPA shell, preserves the existing activity data
loading and mutation contracts, and removes the project's dependency on hard
page navigation, legacy activity scripts, and the legacy activity template.

## Summary of Changes

### 1. SPA Activity Feature Migration
- **Activity SPA Feature Modules**: Added the activity feature slice following the
  established `{feature}.api.js` / `{feature}.views.js` / `{feature}.page.js`
  separation:
  - `SPA/features/activity/activity.api.js`
  - `SPA/features/activity/activity.views.js`
  - `SPA/features/activity/activity.page.js`
- **SPA Route Integration**: Wired `/activity` through
  `SPA/core/router/routes.js`, `SPA/core/router/render-template.js`, and the
  `runRouteInitializer` branch in `SPA/core/app/create-app.js`.
- **Shared Shell Rendering**: Activity now renders entirely inside the
  authenticated app shell outlet and follows the same lifecycle/navigation
  model as Feed, Post Detail, and Create/Edit Post.

### 2. Activity Runtime & Navigation Refactor
- **SPA Navigation Transitions**: Replaced hard `window.location` navigation with
  SPA route transitions backed by the History APIs (`pushState` /
  `replaceState`) and the shared `data-link` anchor pattern.
- **Stateful Activity Interactions**: Added SPA-managed:
  - filtering (status, items-per-section) with `history.replaceState` for
    deep-link/refresh fidelity
  - pagination wired through the shared `SPA/core/ui/pagination.js` controller
  - collapsible section toggles
  - inline comment editing that reuses `SPA/core/shared/image-picker.js`
  - data-action dispatch for owner-only post controls (status toggle, edit,
    delete) and per-comment edit/delete
- **Reload-Free Mutations**: Post and comment mutations
  (`PATCH /api/v1/posts/:id`, `DELETE /api/v1/posts/:id`,
  `PATCH /api/v1/comments/:id`, `DELETE /api/v1/comments/:id`) now refresh the
  activity view through an in-place `refresh()` instead of triggering a full
  document reload.

### 3. Legacy Cleanup & Documentation
- **Removed Legacy Activity Infrastructure**:
  - deleted `web/templates/activity.html`
  - deleted `web/static/js/activity/activity-page.js`
  - deleted `web/static/js/activity/api-activity.js`
  - deleted `web/static/js/activity/bootstrap-activity.js`
  - deleted `web/static/js/activity/comments-activity.js`
  - deleted `web/static/js/activity/posts-activity.js`
  - deleted `web/static/js/activity/render-activity.js`
  - deleted `web/static/js/activity/sections-activity.js`
  - deleted `web/static/js/activity/state-activity.js`
- **Standalone Activity Removal**: Activity no longer depends on standalone
  templates or legacy render flows; no Go route serves a legacy activity
  template, and no SPA module imports from `web/static/js/activity/*`.
- **Documentation Updates**: Updated `architecture.md`, `docs/SDS.md`,
  `docs/ticket-tracker.md`, and `docs/track-b.md` to reflect the SPA-based
  Activity implementation, the removal of legacy infrastructure, and the
  follow-up note for the orphan `web/static/css/activity.css`.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **B05**:
> "activity is accessible inside the shared shell
> activity data loads correctly
> activity navigation no longer depends on a standalone template"

## Testing & Validation Verified

### Automated Test Suite
- [x] `npx vitest run` — 69/69 tests passing, including the `/activity`
  deep-link route coverage in `SPA/tests/unit/core/app/create-app.test.js`
- [x] `bun run policy` — Biome and Vitest frontend policy gate passing
- [x] `bun x biome check SPA/features SPA/core` — clean
- [x] SPA import-boundary policy tests
  (`SPA/tests/unit/policy/spa-import-paths.test.js`) — passing (confirms no
  module re-introduces legacy `/web/static/*` imports or
  `DOMContentLoaded`-style standalone boot logic)

### QA Checklist
- [x] Verified `/activity` renders inside the authenticated SPA shell
- [x] Verified browser refresh on `/activity` is served by the SPA shell
  catch-all and re-hydrates through `initActivityPage`
- [x] Verified browser back/forward navigation replays through the SPA router
  without a full document reload
- [x] Verified activity API requests use the SPA runtime flow
  (`GET /api/v1/users/activity` with `page`, `per_page`, and optional `status`)
- [x] Verified activity mutations refresh state without full-page reloads
- [x] Verified shell-level logout remains visible while on `/activity`
- [x] Verified SPA modules do not import directly from
  `web/static/js/activity/*`
- [x] Verified no remaining references to `web/templates/activity.html` in
  code, templates, or routing

### Automated Behavioral Verification (E2E)
- [x] SPA route lifecycle tests verify `/activity` initializes through the
  shared shell instead of standalone boot logic
- [x] SPA navigation tests verify shell persistence and logout visibility while
  navigating to and from `/activity`
- [x] Auth-gating tests verify unauthenticated entry to `/activity` resolves to
  the login flow

## Key Files Impacted
- `SPA/features/activity/activity.api.js`
- `SPA/features/activity/activity.views.js`
- `SPA/features/activity/activity.page.js`
- `SPA/core/app/create-app.js`
- `SPA/core/router/routes.js`
- `SPA/core/router/render-template.js`
- `web/templates/activity.html` *(deleted)*
- `web/static/js/activity/*` *(deleted — 8 modules)*
- `architecture.md`
- `docs/SDS.md`
- `docs/ticket-tracker.md`
- `docs/track-b.md`
- `README.md`

## Follow-up Notes
- `web/static/css/activity.css` is now orphaned legacy styling — it is not
  loaded by the SPA bundle and can be safely ported into the SPA stylesheet or
  removed in a future cleanup ticket. Out of scope for B05.
