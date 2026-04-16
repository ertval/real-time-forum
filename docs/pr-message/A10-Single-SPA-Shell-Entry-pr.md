# A10: Single SPA Shell Entry
<!-- Filename: docs/pr-message/A10-Single-SPA-Shell-Entry-pr.md -->

This PR implements the single SPA shell entry point for the Real-Time Forum. It marks the transition from a traditional multi-page template-based application to a modern Single Page Application (SPA), as required by the project specifications.

## Summary of Changes

### 1. SPA Foundations (web/SPA/)
- **Core Shell**: Created `index.html` as the single entry point, featuring a premium design system with CSS variables, Google Fonts (Inter, Outfit), and a sophisticated dark mode palette.
- **Boot Logic**: Initialized `main.js` with a robust boot sequence, including a loading overlay and dynamic content injection.
- **Design System**: Implemented `assets/css/main.css` with a comprehensive set of tokens for typography, colors, shadows, and transitions (ES2026+ style).

### 2. Frontend Server (cmd/frontend/)
- **SPA Routing**: Updated `routes.go` to serve the SPA shell for all non-api, non-static routes, enabling seamless client-side navigation.
- **Real-Time Proxy**: Added a dedicated proxy handler for `/ws` to support the upcoming WebSocket implementation.
- **Legacy Cleanup**: Removed all individual template routes and consolidated the routing logic.
- **Robustness Fix**: Refined `server.go`'s `CustomFileServer` and `statusRecorder` to correctly intercept 404s and serve the SPA shell with a 200 OK status, essential for sub-route direct access.

### 3. Tooling & Configuration
- **Biome Integration**: Updated `biome.json` to include the `web/SPA` directory in linting and formatting checks.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **A10**:
> "- the app is served through one HTML document
> - direct navigation to supported SPA routes resolves successfully
> - static assets load correctly
> - /api/ proxy behavior remains intact"

## Testing & Validation Verified

### Automated Test Suite
- [x] **Backend (Go)**: `go test -v ./cmd/frontend` — Verified SPA routing success for root, sub-routes (/login, /register), and static assets.
- [x] **Linting (Biome)**: `node node_modules/.bin/biome check web/SPA/` — Passed with automated fixes applied.

### Manual E2E Verification
- [x] Verified that visiting the root `/` loads the new premium SPA shell.
- [x] Verified that visiting a sub-route like `/login` directly also loads the SPA shell (no 404).
- [x] Verified that static assets (main.js, main.css) are served correctly from the `web/SPA` directory.
- [x] Verified that `/api/` proxying remains functional for backend communication.

## Key Files Impacted
- `web/SPA/index.html`
- `web/SPA/main.js`
- `web/SPA/assets/css/main.css`
- `cmd/frontend/routes.go`
- `cmd/frontend/server.go`
- `biome.json`
- `docs/ticket-tracker.md`
