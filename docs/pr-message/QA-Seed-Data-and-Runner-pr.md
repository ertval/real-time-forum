# QA: Deterministic Seed Data and Runner

This PR adds a repeatable QA seed workflow for the SQLite database without mixing it with application bootstrap data. It introduces a dedicated seed runner, separates bootstrap categories from disposable QA sample data, and provides a deterministic dataset that can be reapplied across environments.

## Summary of Changes

### 1. Bootstrap vs QA Seed Separation
- **Bootstrap Categories Preserved**: Moved the fixed application categories into `internal/db/bootstrap/default_categories.sql`.
- **QA Data Isolated**: Added ordered SQL files under `internal/db/seeds/` for disposable QA data only.

### 2. QA Seed Runner
- **Dedicated Seed Command**: Added `cmd/qa-seed/main.go` to initialize the schema and apply QA seed files in lexical order.
- **Make Entry Point**: Added `make seed-qa` as the standard way to seed a local database.

### 3. Deterministic Sample Dataset
- **Repeatable Reset Flow**: Added a reset script that clears QA-owned tables while leaving bootstrap categories intact.
- **Stable Sample Records**: Added fixed users, posts, post-category joins, comments, reactions, and notifications with deterministic IDs and timestamps.
- **Login-Capable Seeded Users**: Seeded users use a valid bcrypt password hash so seeded accounts can be used for manual QA login.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for this QA support change:
> - QA bootstrap data is separated from disposable sample data
> - a repeatable command exists to recreate the same local dataset
> - seeded users and forum content can be reused across QA runs

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./...` - PASS. Includes `internal/db/qa_seeds_test.go` coverage for row counts and seeded-user login.
- [ ] `bun run policy` - Not run. This PR only adds backend/database seed functionality.
- [ ] `bun x biome check .` - Not run globally. No frontend scope required for this change.
- [x] `go run ./cmd/qa-seed --db-path /tmp/forum-seed-check.db` - PASS. Verified seed runner creates and populates a fresh SQLite database.

### QA Checklist
- [x] Verified seed runner applies bootstrap categories plus QA data to a fresh database file.
- [x] Verified QA seed reset does not delete bootstrap categories.
- [x] Verified seeded data counts match the intended dataset for users, posts, comments, reactions, and notifications.
- [x] Verified seeded users can authenticate with the shared QA password.

### Manual E2E Verification
- [x] Ran the QA seed runner against a fresh DB path and confirmed completion without schema errors.
- [x] Confirmed the runner writes to the exact `--db-path` provided.
- [x] Confirmed the default local workflow is available through `make seed-qa`.

## Key Files Impacted
- `cmd/qa-seed/main.go`
- `internal/db/db.go`
- `internal/db/bootstrap/default_categories.sql`
- `internal/db/seeds/00_reset.sql`
- `internal/db/seeds/10_users.sql`
- `internal/db/seeds/20_posts.sql`
- `internal/db/seeds/21_post_categories.sql`
- `internal/db/seeds/30_comments.sql`
- `internal/db/seeds/40_reactions.sql`
- `internal/db/seeds/50_notifications.sql`
- `internal/db/qa_seeds_test.go`
- `Makefile`
