package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes a SQLite database file (empty for now)
func InitDB(filepath string) error {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Test connection
	return db.Ping()
}
