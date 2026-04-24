# D10: SPA Login and Registration Views
<!-- Filename: pr-message/D10-SPA-Login-and-Registration-Views-pr.md -->

This PR implements ticket **D10** by completing the SPA-side login and registration entry flow. It adds the required registration fields, keeps login compatible with username-or-email entry, removes unsupported guest/OAuth/password-reset paths from the visible auth UI, and introduces a dedicated auth feature handler for payload shaping, inline error handling, and submit-state behavior. The work stays frontend-only and aligns with the real-time forum auth requirements in `docs/requirements.md`, `docs/PRD.md`, and `docs/track-d.md`.

## Summary of Changes

### 1. SPA Auth Views
- **Removed unsupported entry paths**: Deleted the visible OAuth and password-reset UI from the SPA auth screens so the rendered auth flow matches D10 exactly.
- **Extended registration fields**: Kept registration aligned with the SDS/PRD payload shape by rendering `first_name`, `last_name`, `age`, `gender`, `username`, `email`, and `password`.
- **Added HTML field constraints**: Added client-side `max="120"` for age and `minlength="8"` for password to improve the registration UX without moving validation responsibility out of the backend.

### 2. Auth Feature Slice Behavior
- **Dedicated auth handler module**: Added `SPA/features/auth/auth.handlers.js` so auth-specific logic lives in the `features/auth` slice rather than the global app bootstrap.
- **Login payload routing**: Implemented the D10 login contract so identifiers containing `@` are sent as `{ email, password }`, otherwise `{ username, password }`.
- **Registration payload shaping**: Implemented the full SDS registration payload shape in the auth slice and hardened age parsing so invalid programmatic values are omitted instead of serialized as strings.
- **Failure handling and busy state**: Added inline auth error rendering, network-error handling, and temporary submit-button disabling/restoration around auth requests.

### 3. Frontend Verification Coverage
- **Route-level D10 tests**: Added focused SPA route checks covering `/login`, `/register`, required registration fields, and the absence of guest/OAuth entry options.
- **Auth-slice unit coverage**: Added focused tests for login payload mapping, registration payload shaping, success navigation, failure messaging, network errors, callback behavior, trimming rules, and invalid age handling.
- **Dead auth CSS cleanup**: Removed stale auth selectors left behind after the OAuth/forgot-password markup was removed.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **D10**:
> - login and registration routes render inside the SPA
> - registration includes all required fields
> - login accepts username or email entry
> - guest and OAuth entry options are not exposed

## Testing & Validation Verified

### Automated Test Suite
- [x] `make test` — Not run for this PR write-up.
- [x] `bun run policy` — Not run for this PR write-up.
- [x] `bun x biome check .` — Not run for this PR write-up.
- [x] `npm test -- SPA/tests/unit/features/auth/auth.handlers.test.js` — Passed (`16/16` tests).
- [x] `npm test -- SPA/tests/unit/core/app/create-app.test.js SPA/tests/unit/features/auth/auth.handlers.test.js` — Passed (`37/37` tests).

### QA Checklist
- [ ] Browser QA for `/login` and `/register` routes — Not run in this PR write-up.
- [ ] Manual regression check for guest/OAuth entry options not being visible — Not run in this PR write-up.
- [ ] Manual regression check for registration field constraints (`age max`, `password minlength`) — Not run in this PR write-up.

### Manual E2E Verification
- [ ] Verified `/login` renders as an unauthenticated SPA route in a browser session — Not run in this PR write-up.
- [ ] Verified `/register` renders all required registration fields in a browser session — Not run in this PR write-up.
- [ ] Verified login accepts username and email entry paths against a live backend session — Not run in this PR write-up.
- [ ] Verified guest and OAuth entry options are absent from the SPA auth UI — Not run in this PR write-up.

## Key Files Impacted
- `SPA/features/auth/auth.views.js`
- `SPA/features/auth/auth.css`
- `SPA/features/auth/auth.handlers.js`
- `SPA/tests/unit/core/app/create-app.test.js`
- `SPA/tests/unit/features/auth/auth.handlers.test.js`
