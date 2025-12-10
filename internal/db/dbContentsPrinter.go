package db

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
)

// InspectAllTablesTo writes a readable dump of all tables to the provided writer.
// Useful for debugging database state.
func InspectAllTables(database *sql.DB, out io.Writer) error {
	LogInfo("Inspecting SQLite tables...")

	// ---------------------------------------------------------
	// 1) Fetch all user tables (skip internal sqlite_% tables)
	// ---------------------------------------------------------
	rows, err := database.Query(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name;
	`)
	if err != nil {
		return WrapError("list tables", MapSQLError(err))
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return WrapError("scan table name", err)
		}
		tableNames = append(tableNames, t)
	}
	if err := rows.Err(); err != nil {
		return WrapError("iterate tables", err)
	}

	// ---------------------------------------------------------
	// 2) Iterate tables and dump content
	// ---------------------------------------------------------
	for _, table := range tableNames {

		fmt.Fprintf(out, "\nTable: %s\n", table)
		fmt.Fprintln(out, strings.Repeat("-", len(table)+8))

		r, err := database.Query("SELECT * FROM " + table + " LIMIT 50;")
		if err != nil {
			LogWarn("Query failed for table %s: %v", table, err)
			fmt.Fprintf(out, "  (error querying table: %v)\n", err)
			continue
		}

		cols, err := r.Columns()
		if err != nil {
			LogWarn("Failed to read column names for %s: %v", table, err)
			fmt.Fprintf(out, "  (error reading columns: %v)\n", err)
			_ = r.Close()
			continue
		}

		if len(cols) == 0 {
			fmt.Fprintln(out, "  (no columns)")
			_ = r.Close()
			continue
		}

		// Prepare scan buffers
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		// ---------------------------------------------------------
		// 3) Print rows
		// ---------------------------------------------------------
		for r.Next() {
			if err := r.Scan(ptrs...); err != nil {
				LogWarn("Scan error in table %s: %v", table, err)
				fmt.Fprintf(out, "  scan error: %v\n", err)
				continue
			}

			for i, col := range cols {
				val := formatValue(values[i])
				fmt.Fprintf(out, "%s=%s  ", col, val)
			}
			fmt.Fprintln(out)
		}
		if err := r.Err(); err != nil {
			LogWarn("Row iteration error in %s: %v", table, err)
			fmt.Fprintf(out, "  (rows error: %v)\n", err)
		}

		_ = r.Close()
	}

	fmt.Fprintln(out, "\n=== Inspection Complete ===")
	LogInfo("Completed inspection of %d tables", len(tableNames))
	return nil
}

// formatValue converts scanned SQL values into human-readable strings.
func formatValue(v interface{}) string {
	switch vv := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(vv)
	default:
		return fmt.Sprintf("%v", vv)
	}
}
