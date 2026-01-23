package db

import (
	"context"
	"database/sql"
	"time"
)

type Draft struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	UpdatedAt string `json:"updated_at"`
}

func DraftUpsert(
	ctx context.Context,
	db *sql.DB,
	userID int64,
	title, body string,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var id int64
	err := db.QueryRowContext(ctx, `
		SELECT id FROM posts
		WHERE author_id = ? AND status = 'draft'
		LIMIT 1
	`, userID).Scan(&id)

	if err == nil {
		_, err = db.ExecContext(ctx, `
			UPDATE posts
			SET title = ?, body = ?, updated_at = datetime('now')
			WHERE id = ?
		`, title, body, id)
		return id, err
	}

	if err != sql.ErrNoRows {
		return 0, err
	}

	res, err := db.ExecContext(ctx, `
		INSERT INTO posts (author_id, title, body, status)
		VALUES (?, ?, ?, 'draft')
	`, userID, title, body)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func DraftGet(
	ctx context.Context,
	db *sql.DB,
	userID int64,
) (*Draft, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var d Draft
	err := db.QueryRowContext(ctx, `
		SELECT id, title, body, updated_at
		FROM posts
		WHERE author_id = ? AND status = 'draft'
		LIMIT 1
	`, userID).Scan(&d.ID, &d.Title, &d.Body, &d.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &d, nil
}

func DraftDeleteByUser(
	ctx context.Context,
	db *sql.DB,
	userID int64,
) error {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		DELETE FROM posts
		WHERE author_id = ? AND status = 'draft'
	`, userID)

	return err
}
