package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Category represents a forum category.
type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

const categoryTimeout = 2 * time.Second

// ---------------------------------------------------------
// LIST CATEGORIES
// ---------------------------------------------------------

func ListCategories(ctx context.Context, db *sql.DB) ([]Category, error) {
	ctx, cancel := context.WithTimeout(ctx, categoryTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT id, name, slug, created_at
		FROM categories
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}

// ---------------------------------------------------------
// GET CATEGORY
// ---------------------------------------------------------

func GetCategory(ctx context.Context, db *sql.DB, id int64) (Category, error) {
	ctx, cancel := context.WithTimeout(ctx, categoryTimeout)
	defer cancel()

	var c Category

	err := db.QueryRowContext(ctx, `
		SELECT id, name, slug, created_at
		FROM categories
		WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt)

	if err != nil {
		return Category{}, err // ErrNoRows handled by caller
	}

	return c, nil
}

// ---------------------------------------------------------
// CREATE CATEGORY
// ---------------------------------------------------------

func CreateCategory(ctx context.Context, db *sql.DB, name, slug string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, categoryTimeout)
	defer cancel()

	res, err := db.ExecContext(ctx, `
		INSERT INTO categories (name, slug, created_at)
		VALUES (?, ?, datetime('now'))
	`, name, slug)
	if err != nil {
		return 0, fmt.Errorf("create category: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return id, nil
}

// ---------------------------------------------------------
// UPDATE CATEGORY NAME
// ---------------------------------------------------------

func UpdateCategoryName(ctx context.Context, db *sql.DB, id int64, name string) error {
	ctx, cancel := context.WithTimeout(ctx, categoryTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		UPDATE categories
		SET name = ?
		WHERE id = ?
	`, name, id)

	if err != nil {
		return fmt.Errorf("update category name: %w", err)
	}

	return nil
}

// ---------------------------------------------------------
// DELETE CATEGORY
// ---------------------------------------------------------

func DeleteCategory(ctx context.Context, db *sql.DB, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, categoryTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		DELETE FROM categories
		WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	return nil
}
