package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

// InitDB opens (or creates) the file, applies schema once, and returns the live *sql.DB.
func InitDB(dbPath, schemaPath string) (*sql.DB, error) {
	// DSN toggles: FK enforcement, WAL for concurrency, sane timeouts.
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL", dbPath)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Quick connectivity check
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// (Belt & suspenders) ensure FKs are on for this connection.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		db.Close()
		return nil, err
	}

	// Apply schema from file
	schema, err := os.ReadFile(schemaPath)
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
