# C11: Registration API Contract
<!-- Filename: docs/pr-message/C11-RegistrationAPIContract-pr.md -->

Adds validation for the four profile fields introduced in C10 (`age`, `gender`, `first_name`, `last_name`) to the registration flow. Registration now rejects missing or blank profile fields at the db layer, satisfying the SDS §5.2 validation rules and unblocking C08 (backend test coverage).

## Summary of Changes

### 1. Repository Layer — Validation
- **`validateCreateUser`** (`internal/db/users_helpers.go`): Extended with profile field validation — `age` must be a positive integer; `gender`, `first_name`, and `last_name` must be non-empty non-whitespace strings. All existing username, email, and password rules are unchanged.

### 2. Test Suite
- **`TestC11_Registration_MissingProfileFields`** (`internal/tests/users_test.go`): 8 rejection cases covering zero age, negative age, empty/whitespace gender, empty/whitespace first name, and empty/whitespace last name — all expect `400`.
- **`TestC11_Registration_ValidPayloadSucceeds`**: Confirms a fully valid extended payload returns `201` and sets a session cookie.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C11**:
> "Registration rejects missing required profile fields. Valid extended payloads succeed. Successful registration still creates a session."

- `TestC11_Registration_MissingProfileFields` — 8 cases, all rejected with `400`
- `TestC11_Registration_ValidPayloadSucceeds` — valid payload returns `201` with session cookie

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./internal/...` — All 3 packages PASS (`internal/db`, `internal/handlers`, `internal/tests`)
- [x] `bun run policy` — 23 frontend tests PASS, Biome clean (82 files, no fixes applied)
- [x] `bun x biome check .` — 82 files checked, no fixes applied

### QA Checklist
- [x] All pre-existing tests pass — no regressions introduced
- [x] Existing username, email, and password validation rules verified unchanged

### Manual E2E Verification
- [x] `POST /api/v1/users/register` with `age=0` — returns `400` with `"age must be a positive integer"`
- [x] `POST /api/v1/users/register` with valid full payload — returns `201` with user id and session cookie
- [ ] Browser registration form — blocked on D10 (SPA Login and Registration Views)

## Key Files Impacted
- `internal/db/users_helpers.go`
- `internal/tests/users_test.go`
- `docs/ticket-tracker.md`
