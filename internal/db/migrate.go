// internal/db/migrate.go
package db

import (
	"context"
	"database/sql"
	"fmt"
)

// columnSpec describes a column that must exist on a table after migration.
// Adding a new entry here is the only step needed to evolve a table that
// already exists in a deployed database; fresh databases already have the
// column from forum_schema.sql, so Migrate sees it and skips.
type columnSpec struct {
	table      string
	column     string
	definition string // SQL fragment, e.g. "INTEGER NOT NULL DEFAULT 0"
}

// requiredColumns lists every column added to the schema *after* the initial
// release. forum_schema.sql is authoritative for new databases; this list is
// authoritative for upgrading legacy databases. The two must stay in sync —
// any column in requiredColumns must also exist in forum_schema.sql with the
// same definition.
//
// A table needs entries here only once a column is added to it *after* the
// table first shipped. A brand-new table carries all of its columns in
// forum_schema.sql via CREATE TABLE IF NOT EXISTS, so it needs no entry — but
// once a later release adds a column to a table that already exists in
// deployed databases, that column MUST be listed here: CREATE TABLE IF NOT
// EXISTS is a no-op on an existing table and will not add it. That is why
// private_messages.image_path appears below even though private_messages is
// itself created by the schema file.
var requiredColumns = []columnSpec{
	// C10 — User profile schema extension. NOT NULL with safe defaults so
	// existing user rows remain valid after the ALTER.
	{"users", "age", "INTEGER NOT NULL DEFAULT 0"},
	{"users", "gender", "TEXT NOT NULL DEFAULT ''"},
	{"users", "first_name", "TEXT NOT NULL DEFAULT ''"},
	{"users", "last_name", "TEXT NOT NULL DEFAULT ''"},

	// Image upload feature — posts and comments gained an `image_url`
	// after the initial release. Nullable TEXT so the ALTER cannot fail
	// on populated tables.
	{"posts", "image_url", "TEXT"},
	{"comments", "image_url", "TEXT"},

	// C09 — DM image attachments. private_messages predates this column, so
	// databases created between C02 and C09 need the ALTER. Nullable so it
	// cannot fail on populated tables.
	{"private_messages", "image_path", "TEXT DEFAULT NULL"},
}

// Migrate brings an existing database up to the current schema by adding any
// columns missing from tables that pre-date the current code. Idempotent:
// safe to call on fresh databases (no-op), on already-migrated databases
// (no-op), and on partial-schema fixtures (tables not present are skipped).
//
// Tables that did not exist in the legacy schema (e.g. private_messages) are
// created by forum_schema.sql via CREATE TABLE IF NOT EXISTS, so they need
// no work here.
func Migrate(ctx context.Context, database *sql.DB) error {
	// SQLite's database/sql transactions start in DEFERRED mode, which means
	// two concurrent boots could both observe "column missing" before either
	// runs its ALTER. The forum only ever has a single InitDB caller, so the
	// race window does not exist in practice — if that ever changes, switch
	// the DSN to include `_txlock=immediate` to upgrade every transaction to
	// IMMEDIATE locking.
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return WrapError("begin migration tx", err)
	}
	defer tx.Rollback()

	// Cache PRAGMA results per table. requiredColumns lists multiple columns
	// per table (four on users), so caching avoids redundant introspection.
	seen := map[string]map[string]bool{}

	for _, c := range requiredColumns {
		cols, ok := seen[c.table]
		if !ok {
			cols, err = tableColumns(ctx, tx, c.table)
			if err != nil {
				return WrapError(fmt.Sprintf("introspect %s", c.table), err)
			}
			seen[c.table] = cols
		}
		// Empty result from PRAGMA table_info means the table does not
		// exist. forum_schema.sql is supposed to have created it before
		// Migrate runs; if it didn't, the schema apply or the table list
		// is inconsistent. Skip rather than fail — this keeps Migrate
		// robust against partial-schema test fixtures and against a
		// future schema that retires a table.
		if len(cols) == 0 {
			continue
		}
		if cols[c.column] {
			continue
		}
		// Identifiers come from the hardcoded requiredColumns list — no
		// user input ever reaches this string. Plain %s is correct;
		// adding SQLite-style ("") quoting would be misleading defense.
		stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", c.table, c.column, c.definition)
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return WrapError(fmt.Sprintf("add column %s.%s", c.table, c.column), err)
		}
		// Record the added column so a later spec on the same table
		// observes the new state from the cache.
		cols[c.column] = true
	}

	if err := tx.Commit(); err != nil {
		return WrapError("commit migration tx", err)
	}
	return nil
}

// tableColumns returns the set of columns present on a table via
// PRAGMA table_info. An empty map means the table does not exist;
// PRAGMA table_info returns no rows in that case, not an error.
//
// `table` must be a hardcoded identifier from requiredColumns. PRAGMA
// statements cannot accept bound parameters, so the table name is
// inlined; no user input reaches this string.
func tableColumns(ctx context.Context, tx *sql.Tx, table string) (map[string]bool, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := map[string]bool{}
	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}
