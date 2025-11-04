package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Embed the SQL schema directly into the binary.
// This ensures the database schema is always available at runtime,
// even if the .sql file isn't present on disk (prevents path errors).
//
//go:embed forum_schema.sql
var schemaFS embed.FS

// InitDB opens (or creates) a SQLite database file, applies the schema, and returns *sql.DB.
// Foreign keys are enforced, WAL mode is enabled for concurrency, and busy timeout is 5s.
func InitDB(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL", dbPath)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, WrapError("open database", err)
	}

	// Connectivity check by pinging the DB
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, WrapError("ping database", err)
	}

	// Ensure foreign keys are enforced
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		db.Close()
		return nil, WrapError("enable foreign_keys", err)
	}

	// Read schema from embedded SQL file
	schema, err := schemaFS.ReadFile("forum_schema.sql")
	if err != nil {
		db.Close()
		return nil, WrapError("read embedded schema file", err)
	}

	// Apply schema to database
	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, WrapError("apply schema", MapSQLError(err))
	}

	LogInfo("SQLite database initialized at %s", dbPath)
	return db, nil
}
