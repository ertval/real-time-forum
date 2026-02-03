// internal/db/db.go
package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

/*---------------
  EMBEDED FILES
---------------*/

//go:embed forum_schema.sql
var schemaFS embed.FS

//go:embed seeds/categories.sql
var categoriesSeed string

// InitDB opens/creates the SQLite database,
// applies PRAGMA options via DSN,
// loads the embedded schema,
// seeds default categories,
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

	/*----------------------
	  LOAD EMBEDED SCHEMA
	----------------------*/

	schema, err := schemaFS.ReadFile("forum_schema.sql")
	if err != nil {
		db.Close()
		return nil, WrapError("read schema file", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, WrapError("apply schema", MapSQLError(err))
	}

	/*-----------------
	  SEED CATEGORIES
	-----------------*/

	if categoriesSeed != "" {
		if _, err := db.Exec(categoriesSeed); err != nil {
			db.Close()
			return nil, WrapError("seed categories", MapSQLError(err))
		}
	}

	return db, nil
}
