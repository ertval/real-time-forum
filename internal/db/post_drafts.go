// internal/db/post_drafts.go

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

func UpsertDraft(
	ctx context.Context,
	db *sql.DB,
	userID int64,
	title, body string,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var existingID int64
	err := db.QueryRowContext(ctx, `
		SELECT id
		FROM posts
		WHERE author_id = ? AND status = 'draft'
		LIMIT 1
	`, userID).Scan(&existingID)

	if err == nil {
		// 2️⃣ UPDATE existing draft
		_, err = db.ExecContext(ctx, `
			UPDATE posts
			SET title = ?, body = ?, updated_at = datetime('now')
			WHERE id = ?
		`, title, body, existingID)

		return existingID, err
	}

	if err != sql.ErrNoRows {
		return 0, err
	}

	// 3️⃣ INSERT new draft
	res, err := db.ExecContext(ctx, `
		INSERT INTO posts (author_id, title, body, status)
		VALUES (?, ?, ?, 'draft')
	`, userID, title, body)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func GetLatestDraftByUser(
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
		ORDER BY updated_at DESC
		LIMIT 1
	`, userID).Scan(&d.ID, &d.Title, &d.Body, &d.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &d, nil
}

func DeleteDraft(ctx context.Context, db *sql.DB, draftID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM posts WHERE id = ? AND status = 'draft'`,
		draftID,
	)
	return err
}

func CreateDraft(
	ctx context.Context,
	db *sql.DB,
	authorID int64,
	title, body string,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	res, err := db.ExecContext(ctx, `
		INSERT INTO posts (author_id, title, body, status)
		VALUES (?, ?, ?, 'draft')
	`, authorID, title, body)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func GetMyDraft(
	ctx context.Context,
	db *sql.DB,
	userID int64,
) (*Draft, error) {
	return GetLatestDraftByUser(ctx, db, userID)
}
