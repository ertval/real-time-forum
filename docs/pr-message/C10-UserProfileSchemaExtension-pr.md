# C10: User Profile Schema Extension
<!-- Filename: docs/pr-message/C10-UserProfileSchemaExtension-pr.md -->

Extends the `users` table and repository layer with the four profile fields required by the real-time forum specification (`age`, `gender`, `first_name`, `last_name`). This is a prerequisite for C11 (registration validation), C02 (private messages), and C07 (migration strategy), and directly satisfies the audit requirement that the registration form collects age, gender, first name, and last name.

## Summary of Changes

### 1. Database Schema
- **New columns**: Added `age INTEGER NOT NULL DEFAULT 0`, `gender TEXT NOT NULL DEFAULT ''`, `first_name TEXT NOT NULL DEFAULT ''`, `last_name TEXT NOT NULL DEFAULT ''` to the `users` table in `forum_schema.sql`. Defaults are set to allow safe migration of existing rows (C07).

### 2. Repository Layer
- **`User` struct**: Extended with `Age int`, `Gender string`, `FirstName string`, `LastName string` and corresponding `json` tags.
- **`CreateUserRequest`**: Extended with all four profile fields so callers can supply them on registration.
- **`GetUser`**: SELECT and Scan updated to include all four new columns.
- **`insertUser`**: Signature and INSERT statement updated to persist all four fields.

### 3. Test Suite
- **New test file** (`internal/tests/users_profile_schema_test.go`): Four tests covering field persistence at the repository layer, existing-read safety, multi-user isolation, and end-to-end API response verification via `GET /api/v1/users/me`.
- **Existing test helpers updated**: All registration payloads and `db.CreateUser` calls across the test suite now include valid profile data, ensuring C11 validation can be added without breaking existing tests.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C10**:
> "New users persist all required profile fields. Existing user reads do not break. Repository coverage exists for the new fields."

- `TestUserProfileSchema_FieldsPersisted` — verifies all four fields are written and read back correctly
- `TestUserProfileSchema_ExistingReadIntact` — verifies pre-existing users are unaffected
- `TestUserProfileSchema_MultipleUsers` — verifies isolation across multiple users
- `TestUserProfileSchema_APIResponseIncludesProfileFields` — verifies fields appear in the `GET /api/v1/users/me` JSON response

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./internal/...` — All 3 packages PASS (`internal/db`, `internal/handlers`, `internal/tests`)
- [ ] `bun run policy` — N/A (backend-only ticket, no frontend changes)
- [ ] `bun x biome check .` — N/A (no JS/TS files changed)

### QA Checklist
- [x] `go test ./internal/... -coverpkg=./internal/...` — `CreateUser` 90.9%, `GetUser` 88.9%, `insertUser` 84.6%
- [x] All existing tests pass without modification to test logic (only payload data updated)
- [x] Schema column order matches SDS §4.1 (`age`, `gender`, `first_name`, `last_name`)
- [x] File-level header comments added to all new and modified files

### Manual E2E Verification
- [x] `POST /api/v1/users/register` with full profile payload via curl — returns `201` with user id
- [x] `GET /api/v1/users/me` with session cookie — returns all four profile fields (`age`, `gender`, `first_name`, `last_name`) with correct values in JSON response
- [ ] Browser registration form — D10 (SPA Login and Registration Views) must implement the submit handler before full browser E2E is possible

## Key Files Impacted
- `internal/db/forum_schema.sql`
- `internal/db/users.go`
- `internal/db/users_helpers.go`
- `internal/tests/users_profile_schema_test.go`
- `internal/tests/helpers_test.go`
- `internal/tests/helpers_post_test.go`
- `internal/tests/reactions_test.go`
- `internal/tests/notifications_test.go`
- `internal/tests/posts_delete_test.go`
- `internal/tests/posts_my_test.go`
- `internal/tests/users_test.go`
- `docs/ticket-tracker.md`
