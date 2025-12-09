package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

// ---------------------------------------------------------
// LIST ALL
// ---------------------------------------------------------

func ListCategories(ctx context.Context, db *sql.DB) ([]Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx,
		`SELECT id, name, slug, created_at FROM categories ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var result []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil
}

// ---------------------------------------------------------
// GET SINGLE CATEGORY
// ---------------------------------------------------------

func GetCategory(ctx context.Context, db *sql.DB, id int64) (Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var c Category
	err := db.QueryRowContext(ctx,
		`SELECT id, name, slug, created_at FROM categories WHERE id = ?`,
		id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt)

	if err != nil {
		return Category{}, err
	}

	return c, nil
}

// ---------------------------------------------------------
// CREATE CATEGORY
// ---------------------------------------------------------

func CreateCategory(ctx context.Context, db *sql.DB, name, slug string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	res, err := db.ExecContext(ctx,
		`INSERT INTO categories (name, slug, created_at)
		 VALUES (?, ?, datetime('now'))`,
		name, slug,
	)
	if err != nil {
		return 0, fmt.Errorf("create category: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

// ---------------------------------------------------------
// UPDATE CATEGORY NAME (slug remains unchanged)
// ---------------------------------------------------------

func UpdateCategoryName(ctx context.Context, db *sql.DB, id int64, newName string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE categories SET name = ? WHERE id = ?`,
		newName, id,
	)
	if err != nil {
		return fmt.Errorf("update category name: %w", err)
	}

	return nil
}

// ---------------------------------------------------------
// DELETE CATEGORY
// ---------------------------------------------------------

func DeleteCategory(ctx context.Context, db *sql.DB, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM categories WHERE id = ?`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}
