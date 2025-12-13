package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Embed database schema inside the binary.
//
//go:embed forum_schema.sql
var schemaFS embed.FS

// InitDB opens/creates the SQLite database,
// applies PRAGMA options via DSN,
// loads the embedded schema,
// and returns a ready-to-use *sql.DB.
func InitDB(dbPath string) (*sql.DB, error) {

	dsn := fmt.Sprintf(
		"%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL",
		dbPath,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, WrapError("open database", err)
	}

	// Ensure the database is reachable
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, WrapError("ping database", err)
	}

	// Load embedded schema
	schema, err := schemaFS.ReadFile("forum_schema.sql")
	if err != nil {
		db.Close()
		return nil, WrapError("read schema file", err)
	}

	// Apply schema (CREATE TABLE IF NOT EXISTS ...)
	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, WrapError("apply schema", MapSQLError(err))
	}

	return db, nil
}
