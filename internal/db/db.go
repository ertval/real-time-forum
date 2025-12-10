package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Embed database schema inside the binary.
var (
	//go:embed forum_schema.sql
	schemaFS embed.FS
)

// InitDB opens/creates the SQLite database file,
// applies PRAGMA settings through DSN,
// loads the embedded schema, and returns *sql.DB.
func InitDB(dbPath string) (*sql.DB, error) {

	// Configure SQLite via DSN parameters.
	dsn := fmt.Sprintf(
		"%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL",
		dbPath,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, WrapError("open database", err)
	}

	// Verify connection before continuing.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, WrapError("ping database", err)
	}

	// Read SQL schema from embedded file.
	schema, err := schemaFS.ReadFile("forum_schema.sql")
	if err != nil {
		db.Close()
		return nil, WrapError("load schema file", err)
	}

	// Apply schema (should contain CREATE TABLE IF NOT EXISTS).
	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, WrapError("apply schema", MapSQLError(err))
	}

	LogInfo("SQLite database initialized at %s", dbPath)
	return db, nil
}
