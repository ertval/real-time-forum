# Track A: Resolve issues #57, #60, #66

This PR resolves Gitea issues **#57** (Frontend security headers & shell hardening), **#60** (Frontend & infra cleanup: legacy web/ tree, pins & tooling), and **#66** (CI pipeline gaps).

Closes #57
Closes #60
Closes #66

## Summary of Changes

### 1. Frontend Security Headers & Shell Hardening (#57)
- **Security Headers Middleware**: Added `SecurityHeaders` middleware in `cmd/frontend/routes.go` emitting `Content-Security-Policy`, `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, and `Referrer-Policy: strict-origin-when-cross-origin` across all frontend responses (SPA catch-all, `/static/`, `/errors/`, `/favicon.ico`).
- **Inline onclick Removal**: Replaced inline `onclick="window.history.back()"` in `SPA/features/profile/profile.page.js` with declarative `data-action="back"` event delegation.
- **Global Error and Promise Rejection Handlers**: Created `SPA/core/errors/error-handler.js` to catch and log unhandled exceptions and promise rejections, initialized during SPA bootstrap in `SPA/main.js`.
- **Unit & Server Coverage**: Added `cmd/frontend/server_test.go` (`TestSecurityHeaders`), `SPA/tests/unit/core/errors/error-handler.test.js`, and `SPA/tests/unit/features/profile/profile.page.test.js`.

### 2. Legacy Web Tree & Dead Code Cleanup (#60)
- **Deleted Legacy MPA Assets**: Removed orphaned legacy assets in `web/templates/*` (6 files), `web/static/js/*` (26 files), `web/static/css/*`, `web/static/partials/*`, `web/static/sounds/*`, and `web/static/img/*`.
- **Preserved Live Static Surfaces**: Kept `/static/` routing for `web/static/uploads/`, `web/static/favicon.ico`, and `web/errors/`.
- **Removed Stale Pins**: Removed obsolete test `A02-03: legacy /static assets remain accessible` from `SPA/tests/e2e/tickets.test.js` and dead template tests from `internal/tests/drafts_test.go`.
- **Deleted Dead Go Code**: Removed `cmd/frontend/backend_validator.go` and defined `backendBaseURL` in `cmd/frontend/routes.go`.
- **Tooling & Config Cleanup**: Removed `/static/js` alias in `vitest.config.ts`, cleaned up `.biomeignore`, anchored binaries in `.gitignore`, and refactored `SPA/tests/integration/frontend_behavior.test.mjs` to import `setupImagePicker` directly.

### 3. CI Pipeline Gaps & Race Detector (#66)
- **CI Workflow Enhancements**: Updated `.github/workflows/ci.yml` to remove non-existent `dev` branch triggers, and added `make build`, `make vet`, `make test-race`, and `make test-e2e` steps with read-only lint/policy verification.
- **Makefile Target**: Added `test-race` target in `Makefile` (`go test -race -timeout 20m ./...`).
- **Race Condition & Timeout Hardening**: Extended `userTimeout` in `internal/db/users.go` from 2s to 10s to prevent SQLite context deadline timeouts during high-concurrency bcrypt operations under `-race`.

---

## Verification Gate Satisfaction

This PR fully satisfies all verification gates for issues #57, #60, and #66:
- Frontend HTTP responses emit strict CSP and defense-in-depth security headers.
- No inline scripts/handlers exist in the SPA DOM markup.
- Legacy MPA templates, styles, scripts, and tests are completely removed.
- Full CI test suite including Go race detector and Playwright E2E executes cleanly.
- PR Audit report generated and verified: `docs/audit-reports/pr-audit-ekaramet-A-57-security-cleanup-ci.md` (**PASS**).

---

## Testing & Validation Verified

### Automated Test Suite
- [x] `make build` — PASS (both `forum-backend` and `forum-frontend` built cleanly)
- [x] `make vet` — PASS (0 warnings/errors)
- [x] `make test` — PASS (Go tests + Vitest + 20 Playwright E2E tests)
- [x] `make test-race` — PASS (all backend tests passed under `-race`)
- [x] `bun run policy` — PASS (102 files checked by Biome, 37 test files / 336 unit & integration tests passed)
- [x] `npm run policy` — PASS (Exit code 0)

### Key Files Impacted
- `cmd/frontend/routes.go`
- `cmd/frontend/server_test.go`
- `cmd/frontend/backend_validator.go` (deleted)
- `SPA/main.js`
- `SPA/core/errors/error-handler.js` (new)
- `SPA/features/profile/profile.page.js`
- `SPA/tests/unit/core/errors/error-handler.test.js` (new)
- `SPA/tests/unit/features/profile/profile.page.test.js` (new)
- `SPA/tests/e2e/tickets.test.js`
- `SPA/tests/integration/frontend_behavior.test.mjs`
- `internal/db/users.go`
- `internal/tests/drafts_test.go`
- `.github/workflows/ci.yml`
- `Makefile`
- `vitest.config.ts`
- `.gitignore`
- `.biomeignore`
- `docs/audit-reports/pr-audit-ekaramet-A-57-security-cleanup-ci.md`
- `docs/pr-message/A57-SecurityCleanupAndCI-pr.md`
