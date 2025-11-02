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

// InitDB opens (or creates) the file, applies schema once, and returns the live *sql.DB.
func InitDB(dbPath string) (*sql.DB, error) {
	// DSN toggles: FK enforcement, WAL for concurrency, 5 sec timeouts.
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL", dbPath)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	//Connectivity check by pinging the DB
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	//Ensures FKs are on for this connection.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		db.Close()
		return nil, err
	}

	//Reads the schema from the embedded file.
	schema, err := schemaFS.ReadFile("forum_schema.sql")
	if err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
