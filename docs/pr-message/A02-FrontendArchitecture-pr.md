# A02: Frontend Architecture & Directory Restructuring

This PR implements the new folder structure for the SPA in the `web/SPA/` directory, following modern Clean Vertical Slices and Screaming Architecture patterns. It also updates the architecture documentation in `docs/SDS.md` and prepares the Go frontend server to serve the SPA assets and entry point.

## Summary of Changes

### 1. SPA Directory Structure
- **Vertical Slices**: Implemented `features/` directory with slices for `auth`, `feed`, `chat`, `post`, and `profile`.
- **Core Infrastructure**: Established `core/` directory for `api`, `router`, `state`, and `utils`.
- **Shared Components**: Created `components/shared/` for reusable UI elements.
- **Asset Management**: Added `assets/` for global CSS, images, and sounds.
- **Entry Points**: Added `index.html` and `main.js` as the root of the SPA.

### 2. Documentation
- **SDS Update**: Updated `docs/SDS.md` (Section 7.0) with the detailed directory structure and technical rationale.

### 3. Frontend Server (Go)
- **Route Preparation**: Added a handler for `/spa/*` in `cmd/frontend/routes.go` to serve the new SPA directory.
- **Non-Breaking Changes**: Ensured existing template-based routes and static asset serving remain fully functional.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **A02**:
> "- the directory tree reflects the approved modern structure
> - docs reflect the current implementation accurately
> - existing functionality behaves as it did before structurally breaking changes"

## Testing & Validation Verified

### Automated Test Suite
- [x] **Backend (Go)**: `make test-backend` — All 439 tests passed (cached/integration).
- [x] **Frontend (SPA Structure)**: `make lint` — Biome check passed (52 files).
- [x] **Build**: `make build-all` — Both frontend and backend binaries built successfully.

### Manual E2E Verification
- [x] **Root Access**: Verified `http://localhost:3000/` still serves the legacy home page.
- [x] **SPA Access**: Verified `http://localhost:3000/spa/` successfully serves the new SPA shell (`index.html`).
- [x] **Static Assets**: Confirmed `/static/` files are still accessible by legacy templates.

## Key Files Impacted
- `docs/SDS.md`
- `cmd/frontend/routes.go`
- `docs/ticket-tracker.md`
- `web/SPA/` (new directory structure)
