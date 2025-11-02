package db

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
)

// InspectAllTablesTo writes a text dump of all tables to the provided writer.
func InspectAllTables(database *sql.DB, out io.Writer) error {
	// 1) list tables
	rows, err := database.Query(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name;
	`)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan table name: %w", err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("list tables err: %w", err)
	}

	for _, table := range tables {
		fmt.Fprintf(out, "\nTable: %s\n", table)
		fmt.Fprintln(out, strings.Repeat("-", len(table)+8))

		r, err := database.Query("SELECT * FROM " + table + " LIMIT 50;")
		if err != nil {
			fmt.Fprintf(out, "  (error querying table: %v)\n", err)
			continue
		}

		cols, _ := r.Columns()
		if len(cols) == 0 {
			fmt.Fprintln(out, "  (no columns)")
			r.Close()
			continue
		}

		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		for r.Next() {
			if err := r.Scan(ptrs...); err != nil {
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
			fmt.Fprintf(out, "  rows err: %v\n", err)
		}
		r.Close()
	}
	fmt.Fprintln(out, "\n=== Done ===")
	return nil
}
