# A06: Global Logout Across the Forum

This PR closes ticket A06 by wiring global logout behavior into the persistent authenticated shell and validating that logout works consistently from every protected forum route. The implementation preserves server-side session invalidation and returns users to the auth flow after logout.

## Summary of Changes

### 1. SPA App Controller Logout Flow
- **Centralized Logout Action Handling**: Added a shell-level logout action path in `SPA/core/app/create-app.js` using global event delegation with `[data-action="logout"]`.
- **Logout API Invocation**: Added `performLogout()` to call `POST /api/v1/users/logout` with cookie credentials.
- **Auth State Reset and Redirect**: After logout attempt, app state is set to unauthenticated and navigation is replaced to `/login`.

### 2. Frontend Regression Coverage
- **Logout Interaction Harness**: Extended test browser mock in `SPA/tests/unit/main.test.js` with `clickLogout()` helper.
- **A06 Functional Tests**: Added tests that verify API call shape, auth-state reset, redirect to login, and logout usability from all protected routes.

### 3. Backend Contract Confirmation (No Code Changes)
- **Existing Logout Contract Reused**: `POST /api/v1/users/logout` already invalidates session token and clears cookie in `internal/handlers/users.go`.
- **Route Guarded by Auth Middleware**: Logout route remains protected in `internal/router/router.go`.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **A06**:
> - logout is visible and usable from every authenticated route
> - logout clears the session and returns the user to auth flow
> - logout no longer depends on create/edit page layout

Audit-backed evidence:
- **Visible and usable everywhere**: Logout control is rendered by the persistent authenticated shell and tested across `/`, `/post/:id`, `/create-post`, `/edit-post/:id`, and `/activity`.
- **Session cleared + auth flow return**: Frontend calls `POST /api/v1/users/logout`, resets `isAuthenticated` to `false`, and replaces route to `/login`; backend invalidates session token and expires cookie.
- **Layout independence**: Logout logic is centralized in app-level shell click handling, not page-specific templates.

## Testing & Validation Verified

### Automated Test Suite
- [x] `make test` - PASS. Go tests pass including `TestUserLogout` session invalidation behavior and full backend regression suite.
- [x] `bun run policy` - PASS. Biome check and Vitest suite pass, including new A06 logout tests.
- [x] `bun x biome check .` - PASS (also covered by policy pipeline).

### QA Checklist
- [x] Verified logout action calls `POST /api/v1/users/logout` with `credentials: include`.
- [x] Verified logout from protected routes returns UI to `/login` and removes authenticated shell.
- [x] Verified `isAuthenticated` app state resets to `false` after logout flow.
- [x] Verified protected-route guard remains effective for unauthenticated state.

### Manual E2E Verification
- [x] Logged in, navigated to protected screens, and confirmed logout control is present in the shared shell.
- [x] Triggered logout and confirmed return to auth flow (`/login`) from multiple protected routes.
- [x] Confirmed post-logout protected content is not rendered in unauthenticated state.

## Key Files Impacted
- `SPA/core/app/create-app.js`
- `SPA/tests/unit/main.test.js`
- `docs/ticket-tracker.md`
- `docs/pr-message/A06-Global-Logout-Across-the-Forum-pr.md`
