package db

import (
	"context"
	"database/sql"
	"fmt"
)

// ------------------------------------------------------------
// COMMENTS HELPERS (PRIVATE)
// ------------------------------------------------------------

// Normalize pagination for comments
func normalizeCommentsPagination(p *ListCommentsParams) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

// Count comments for a specific post
func countCommentsByPost(
	ctx context.Context,
	db *sql.DB,
	postID int64,
) (int, error) {

	var total int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM comments WHERE post_id = ?`,
		postID,
	).Scan(&total); err != nil {
		return 0, fmt.Errorf("count comments: %w", err)
	}
	return total, nil
}

// Fetch paginated comments for a post
func fetchCommentsByPost(
	ctx context.Context,
	db *sql.DB,
	p ListCommentsParams,
) ([]Comment, error) {

	offset := (p.Page - 1) * p.PerPage

	rows, err := db.QueryContext(ctx, `
		SELECT id, post_id, user_id, parent_comment_id, body, created_at, updated_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`, p.PostID, p.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var c Comment
		var parentID sql.NullInt64

		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.UserID,
			&parentID,
			&c.Body,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}

		if parentID.Valid {
			id := parentID.Int64
			c.ParentCommentID = &id
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows comments: %w", err)
	}

	return comments, nil
}
