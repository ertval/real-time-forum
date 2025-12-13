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

type Comment struct {
	ID              int64  `json:"id"`
	PostID          int64  `json:"post_id"`
	UserID          int64  `json:"user_id"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
	Body            string `json:"body"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

type ListCommentsParams struct {
	PostID  int64
	Page    int
	PerPage int
}

type ListCommentsResult struct {
	Comments []Comment
	Total    int
}

//
// ---------------------------------------------------------
// LIST COMMENTS (ORCHESTRATOR)
// ---------------------------------------------------------
//

func ListCommentsByPost(
	ctx context.Context,
	db *sql.DB,
	p ListCommentsParams,
) (ListCommentsResult, error) {

	normalizeCommentsPagination(&p)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	total, err := countCommentsByPost(ctx, db, p.PostID)
	if err != nil {
		return ListCommentsResult{}, err
	}

	comments, err := fetchCommentsByPost(ctx, db, p)
	if err != nil {
		return ListCommentsResult{}, err
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

type CreateCommentInput struct {
	PostID          int64
	UserID          int64
	ParentCommentID *int64
	Body            string
}

func CreateComment(
	ctx context.Context,
	db *sql.DB,
	in CreateCommentInput,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const q = `
		INSERT INTO comments (post_id, user_id, parent_comment_id, body, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`

	res, err := db.ExecContext(ctx, q,
		in.PostID,
		in.UserID,
		in.ParentCommentID,
		in.Body,
	)
	if err != nil {
		return 0, fmt.Errorf("create comment: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

//
// ---------------------------------------------------------
// GET COMMENT
// ---------------------------------------------------------
//

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

	err := db.QueryRowContext(ctx, q).
		Scan(
			&c.ID,
			&c.PostID,
			&c.UserID,
			&parentID,
			&c.Body,
			&c.CreatedAt,
			&c.UpdatedAt,
		)

	if err != nil {
		return Comment{}, err
	}

	if parentID.Valid {
		v := parentID.Int64
		c.ParentCommentID = &v
	}

	return c, nil
}
