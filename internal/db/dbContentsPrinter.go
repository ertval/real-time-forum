package db

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
)

// InspectAllTablesTo writes a text dump of all tables to the provided writer.
func InspectAllTables(database *sql.DB, out io.Writer) error {
	LogInfo("Starting table inspection")

	// 1) List user tables
	rows, err := database.Query(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name;
	`)
	if err != nil {
		return WrapError("list tables", MapSQLError(err))
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return WrapError("scan table name", err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return WrapError("iterate tables", err)
	}

	// 2) Iterate through tables and print data
	for _, table := range tables {
		fmt.Fprintf(out, "\nTable: %s\n", table)
		fmt.Fprintln(out, strings.Repeat("-", len(table)+8))

		r, err := database.Query("SELECT * FROM " + table + " LIMIT 50;")
		if err != nil {
			LogWarn("Error querying table %s: %v", table, err)
			fmt.Fprintf(out, "  (error querying table: %v)\n", err)
			continue
		}

		cols, _ := r.Columns()
		if len(cols) == 0 {
			fmt.Fprintln(out, "  (no columns)")
			_ = r.Close()
			continue
		}

		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		for r.Next() {
			if err := r.Scan(ptrs...); err != nil {
				LogWarn("Scan error in table %s: %v", table, err)
				fmt.Fprintf(out, "  scan error: %v\n", err)
				continue
			}
			for i, col := range cols {
				v := values[i]
				var s string
				switch vv := v.(type) {
				case []byte:
					s = string(vv)
				case nil:
					s = "NULL"
				default:
					s = fmt.Sprintf("%v", vv)
				}
				fmt.Fprintf(out, "%s=%s  ", col, s)
			}
			fmt.Fprintln(out)
		}
		if err := r.Err(); err != nil {
			LogWarn("Row iteration error in table %s: %v", table, err)
			fmt.Fprintf(out, "  rows err: %v\n", err)
		}
		_ = r.Close()
	}

	fmt.Fprintln(out, "\n=== Done ===")
	LogInfo("Completed table inspection for %d tables", len(tables))
	return nil
}
