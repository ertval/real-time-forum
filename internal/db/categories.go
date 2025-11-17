package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Category struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

func ListCategories(ctx context.Context, db *sql.DB) ([]Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	q := `SELECT id, name, slug, created_at FROM categories ORDER BY name ASC`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var out []Category
	for rows.Next() {

		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows categories: %w", err)
	}
	return out, nil
}
