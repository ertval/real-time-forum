# C07: Database Migration Strategy
<!-- Filename: docs/pr-message/C07-Database-Migration-Strategy-pr.md -->

Adds a lightweight migration step that brings a pre-existing SQLite database
forward to the current schema without losing data. The driver is that
`forum_schema.sql` uses `CREATE TABLE IF NOT EXISTS`, which is a no-op on
tables that already exist with an older shape — so a forum upgraded from a
pre-C10 build would crash on the first `SELECT users.age` until those columns
are added. C07 closes that gap.

## Summary of Changes

### 1. DB layer (`internal/db/migrate.go`)
- New file. Exposes a single public function: `Migrate(ctx, *sql.DB) error`.
- The migration model is **column introspection**: an in-code list of
  `columnSpec{table, column, definition}` describes every column added to the
  schema *after* the initial release. `Migrate` introspects each table via
  `PRAGMA table_info` and runs `ALTER TABLE … ADD COLUMN` only for columns
  that are missing.
- `requiredColumns` covers every known post-initial-release column on tables
  that existed in the initial schema:
  - **C10** — `users.age`, `users.gender`, `users.first_name`,
    `users.last_name` (each `NOT NULL` with a safe default so the ALTER
    cannot fail on pre-existing rows).
  - **Image upload feature** — `posts.image_url`, `comments.image_url`
    (nullable `TEXT`, same reason).
- The whole migration runs in a single transaction; either all required
  columns are added or none are.
- **Missing target tables are skipped**, not treated as errors. Production
  reaches `Migrate` after `forum_schema.sql` has guaranteed every current
  table exists, but the lenient behaviour keeps the function robust against
  partial-schema test fixtures and against a future schema that retires a
  table.
- **PRAGMA results are cached per table** so a table with multiple specs
  (e.g. `users` has four C10 columns) is introspected exactly once.
- New private-message tables/indexes that C02 added are already handled by
  the `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS` lines in
  `forum_schema.sql`, so no work is needed for them here.

### 2. Boot wiring (`internal/db/db.go`)
- `InitDB` now calls `Migrate(context.Background(), db)` immediately after
  `forum_schema.sql` is applied and before category seeding. The order is
  deliberate:
  1. `CREATE TABLE IF NOT EXISTS` — covers fresh databases and any tables
     a legacy database is missing.
  2. `Migrate` — adds any columns missing from legacy tables.
  3. Seed default categories — runs against a schema that is now guaranteed
     to be current.
- `Migrate` returns an error on failure; `InitDB` closes the DB handle and
  propagates the error, so a half-upgraded database never reaches the rest
  of the app.

### 3. Tests (`internal/tests/migrate_test.go`)

Ten tests, each scoped to one piece of the contract:

- `TestMigrate_AddsC10ColumnsToLegacyUsers` — opens a hand-crafted database
  with the pre-C10 `users` schema and asserts that after `Migrate` all four
  C10 columns exist. Also asserts the *precondition* (legacy table starts
  without those columns) so a future schema drift won't silently invalidate
  the test.
- `TestMigrate_PreservesLegacyUserRows` — pre-inserts a user with no C10
  fields, runs Migrate, and reads back username/email and the new defaults.
  Proves "existing users remain usable" by data, not just by schema shape.
- `TestMigrate_IdempotentOnLegacyDB` — runs Migrate three times in a row,
  asserts no errors. Documents the contract that `Migrate` is safe to call
  unconditionally on every boot.
- `TestMigrate_NoOpOnFreshDB` — runs Migrate against a fresh DB created via
  the standard `setupTestDB` (which applies the current `forum_schema.sql`,
  so all columns are present). Asserts no errors and no column loss.
- `TestMigrate_RetainedTablesUnaffected` — pre-populates a `posts` row,
  runs Migrate, reads back. Proves retained forum data is untouched.
- `TestMigrate_AddsImageURLToLegacyPostsAndComments` — opens a database
  with the pre-image-upload `posts`/`comments` schemas alongside legacy
  `users`, runs Migrate, asserts `image_url` exists on both tables.
- `TestMigrate_SkipsMissingTables` — runs Migrate on a database that has
  only `users` (no `posts`, no `comments`). Asserts no error, and that the
  user columns still get added. Hardens the "lenient on partial schemas"
  contract.
- `TestMigrate_BeginTxErrorPropagates` — closes the DB before calling
  Migrate. Asserts the BeginTx error is surfaced, not swallowed.
- `TestInitDB_AppliesSchemaAndMigratesLegacyFile` — the only test that
  exercises C07's `db.go` wiring change. Writes a pre-C10 users table to
  a real temp file, opens it with `db.InitDB`, asserts the C10 columns
  are now present, that a pre-existing row survived with the new
  defaults, and that a second `InitDB` against the same file is a no-op.
- `TestMigrate_AppCanWriteAfterMigrate` — the integration canary: legacy
  schema → Migrate → `db.CreateUser` (which writes to every C10 column).
  If any column had been omitted from the migration list the INSERT would
  fail. This is the practical embodiment of the C07 verification gate.

Two `legacy…Schema` SQL fragments are kept verbatim in the test file as
inline documentation of what "pre-C10 users" and "pre-image-upload
posts/comments" looked like — future readers don't have to dig through
git history to know what was migrated away from.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **C07**:
> "schema changes can be applied to an existing database; existing users
> remain usable; retained forum data is preserved."

- ✅ Schema changes apply to existing databases
  (`TestMigrate_AddsC10ColumnsToLegacyUsers`,
  `TestMigrate_IdempotentOnLegacyDB`)
- ✅ Existing users remain usable
  (`TestMigrate_PreservesLegacyUserRows`,
  `TestMigrate_AppCanWriteAfterMigrate`)
- ✅ Retained forum data is preserved
  (`TestMigrate_RetainedTablesUnaffected`)

## Testing & Validation Verified

### Automated Test Suite
- [x] `go test ./... -timeout 180s` — **all packages pass** (21.3 s)
- [x] `go test ./internal/tests/ -run "TestMigrate|TestInitDB" -v` — **10/10 pass**
- [x] `go test ./... -race -timeout 300s` — **clean** (no data races)
- [x] `go build ./...` — **clean**
- [x] `go vet ./...` — **clean**

### Coverage (internal/db)

| Function | Before PR | After |
|---|---|---|
| `db.InitDB` | 0 % | **50.0 %** |
| `db.Migrate` | — (new) | **87.0 %** |
| `db.tableColumns` | — (new) | 81.8 % |

Remaining uncovered lines are error paths with no deterministic trigger
in test: file-open / schema-embed failures in `InitDB`; `PRAGMA table_info`
query failure, scan failure, and transaction commit failure inside
`Migrate`. The BeginTx failure path is covered by
`TestMigrate_BeginTxErrorPropagates`.

### QA Checklist
- [x] Fresh-database boot: `InitDB` runs schema, then `Migrate` is a no-op
  (verified by `TestMigrate_NoOpOnFreshDB` and full suite).
- [x] Legacy-database boot: `InitDB` runs schema (no-op on existing tables),
  then `Migrate` adds the missing C10 columns and existing user rows survive.
- [x] Boot is atomic: a `Migrate` failure aborts `InitDB` and propagates the
  error — the half-upgraded database never reaches the rest of the app.
- [x] Migration is repeat-safe: running `Migrate` multiple times has no
  cumulative effect.

## Design Notes

- The migration system is intentionally **introspection-based**, not
  version-tracked. A `schema_migrations` table would be more enterprise-y
  but the only foreseeable schema evolution from here is adding columns
  to existing tables — exactly what introspection handles cleanly. Adding
  a tracking table can happen later if the project grows a need for
  destructive migrations (column drops, type changes) that introspection
  cannot express.
- `requiredColumns` is the single source of truth for *upgrading* a legacy
  schema. It is paired with `forum_schema.sql`, which is the single source
  of truth for *creating* a fresh one. The PR adds an explicit comment in
  `migrate.go` reminding the next author to keep the two in sync.
- `PRAGMA table_info` is the SQLite-native way to introspect; it cannot be
  parameter-bound, so the table name is identifier-quoted (`%q`). Callers
  only pass internal, fixed table names — there is no user-input path into
  this string.

## Key Files Impacted
- `internal/db/migrate.go` — new; the migration engine
- `internal/db/db.go` — `InitDB` now calls `Migrate` after schema apply
- `internal/tests/migrate_test.go` — new; nine migration tests
- `docs/ticket-tracker.md` — C07 marked done, counts updated
- `docs/pr-message/C07-Database-Migration-Strategy-pr.md` — this file

## Audit Findings Addressed

An independent cold-start audit raised six items (PASS-WITH-NITS, zero
criticals). Actionable items folded back in:

- ✅ `requiredColumns` extended with `posts.image_url` and `comments.image_url`
  (image upload columns were post-initial-release, omitting them was a
  half-job — nit #1)
- ✅ Concurrent `Migrate` race documented; explicit comment points at
  `_txlock=immediate` as the DSN-level fix should the project ever grow
  multiple boot callers (nit #2)
- ✅ `%q` identifier quoting replaced with plain `%s` and comment corrected
  — Go's `%q` produces Go-syntax escapes, not SQLite identifier quoting,
  and since callers only pass hardcoded constants no quoting is required
  (nit #3)
- ✅ Missing failure-path test added: `TestMigrate_BeginTxErrorPropagates`
  proves error propagation on a closed DB; `TestMigrate_SkipsMissingTables`
  proves the lenient missing-table behaviour (nit #6)
- ➖ `context.Background()` in `InitDB` left as-is; `InitDB` does not take
  a context and adding one is out of C07's scope (nit #4)
- ➖ `legacyUsersSchema` fidelity confirmed by the auditor against the
  initial-release schema; no action (nit #5)
