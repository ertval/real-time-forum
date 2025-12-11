package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

//
// ---------------------------------------------------------
// DATA STRUCTURES
// ---------------------------------------------------------
//

// Comment represents a user comment under a post.
type Comment struct {
	ID              int64  `json:"id"`
	PostID          int64  `json:"post_id"`
	UserID          int64  `json:"user_id"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
	Body            string `json:"body"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// Pagination input.
type ListCommentsParams struct {
	PostID  int64
	Page    int
	PerPage int
}

// Returned from a paginated comment list.
type ListCommentsResult struct {
	Comments []Comment
	Total    int
}

//
// ---------------------------------------------------------
// LIST COMMENTS (WITH PAGINATION)
// ---------------------------------------------------------
//

// ListCommentsByPost returns paginated comments for a specific post.
func ListCommentsByPost(ctx context.Context, db *sql.DB, p ListCommentsParams) (ListCommentsResult, error) {

	// Sanitize pagination.
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	offset := (p.Page - 1) * p.PerPage

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Count total comments.
	var total int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM comments WHERE post_id = ?`,
		p.PostID,
	).Scan(&total); err != nil {
		return ListCommentsResult{}, fmt.Errorf("count comments: %w", err)
	}

	// Query paginated comments.
	rows, err := db.QueryContext(ctx, `
		SELECT id, post_id, user_id, parent_comment_id, body, created_at, updated_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?`,
		p.PostID, p.PerPage, offset)
	if err != nil {
		return ListCommentsResult{}, fmt.Errorf("list comments: %w", err)
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
			return ListCommentsResult{}, fmt.Errorf("scan comment: %w", err)
		}

		if parentID.Valid {
			id := parentID.Int64
			c.ParentCommentID = &id
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return ListCommentsResult{}, fmt.Errorf("iterate comments: %w", err)
	}

	return ListCommentsResult{
		Comments: comments,
		Total:    total,
	}, nil
}

//
// ---------------------------------------------------------
// CREATE COMMENT
// ---------------------------------------------------------
//

// CreateCommentInput contains the data needed to insert a new comment.
type CreateCommentInput struct {
	PostID          int64
	UserID          int64
	ParentCommentID *int64
	Body            string
}

// CreateComment inserts a new comment (top-level or nested).
func CreateComment(ctx context.Context, db *sql.DB, in CreateCommentInput) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const q = `
		INSERT INTO comments (post_id, user_id, parent_comment_id, body, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`

	res, err := db.ExecContext(ctx, q,
		in.PostID, in.UserID, in.ParentCommentID, in.Body)
	if err != nil {
		return 0, fmt.Errorf("create comment: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return id, nil
}

//
// ---------------------------------------------------------
// GET COMMENT BY ID
// ---------------------------------------------------------
//

// GetComment returns a single comment by its ID.
func GetComment(ctx context.Context, db *sql.DB, id int64) (Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const q = `
		SELECT id, post_id, user_id, parent_comment_id, body, created_at, updated_at
		FROM comments
		WHERE id = ?
	`

	var c Comment
	var parentID sql.NullInt64

	err := db.QueryRowContext(ctx, q, id).
		Scan(&c.ID, &c.PostID, &c.UserID, &parentID, &c.Body, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return Comment{}, err // caller handles ErrNoRows
	}

	if parentID.Valid {
		v := parentID.Int64
		c.ParentCommentID = &v
	}

	return c, nil
}
