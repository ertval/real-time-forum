package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed forum_schema.sql
var schemaFS embed.FS

// InitDB initializes SQLite with WAL, FK enforcement, busy timeout, etc.
// NOTE: WAL works only on file-backed DBs (NOT :memory:)
func InitDB(dbPath string) (*sql.DB, error) {

	// DSN options configure SQLite pragmas automatically.
	dsn := fmt.Sprintf(
		"%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL",
		dbPath,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, WrapError("open database", err)
	}

	// Ensure connection is valid before continuing.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, WrapError("ping database", err)
	}

	// Load embedded schema file.
	schemaBytes, err := schemaFS.ReadFile("forum_schema.sql")
	if err != nil {
		db.Close()
		return nil, WrapError("load schema", err)
	}

	// Apply schema. Should contain only CREATE IF NOT EXISTS.
	if _, err := db.Exec(string(schemaBytes)); err != nil {
		db.Close()
		return nil, WrapError("apply schema", MapSQLError(err))
	}

	LogInfo("SQLite database initialized at %s", dbPath)
	return db, nil
}
