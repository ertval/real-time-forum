package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// InspectAllTables queries all tables in the connected database and prints their contents.
func InspectAllTables(database *sql.DB) error {
	// Step 1: Get all table names
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

	if len(tables) == 0 {
		fmt.Println("No tables found.")
		return nil
	}

	fmt.Println("=== Inspecting all tables ===")

	// Step 2: Iterate through each table and query all rows
	for _, table := range tables {
		fmt.Printf("\nTable: %s\n", table)
		fmt.Println(strings.Repeat("-", len(table)+8))

		query := fmt.Sprintf("SELECT * FROM %s LIMIT 50;", table)
		r, err := database.Query(query)
		if err != nil {
			fmt.Printf("  (error querying table: %v)\n", err)
			continue
		}

		cols, _ := r.Columns()
		if len(cols) == 0 {
			fmt.Println("  (no columns)")
			r.Close()
			continue
		}

		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		for r.Next() {
			if err := r.Scan(valuePtrs...); err != nil {
				fmt.Printf("  scan error: %v\n", err)
				continue
			}
			for i, col := range cols {
				val := values[i]
				var s string
				switch v := val.(type) {
				case []byte:
					s = string(v)
				case nil:
					s = "NULL"
				default:
					s = fmt.Sprintf("%v", v)
				}
				fmt.Printf("%s=%s  ", col, s)
			}
			fmt.Println()
		}
		r.Close()
	}
	fmt.Println("\n=== Done ===")
	return nil
}
