package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

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

func ListCommentsByPost(ctx context.Context, db *sql.DB, p ListCommentsParams) (ListCommentsResult, error) {
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

	//Total comments
	var total int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments WHERE post_id = ?`, p.PostID).Scan(&total); err != nil {
		return ListCommentsResult{}, fmt.Errorf("count comments: %w", err)
	}

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

	var out []Comment
	for rows.Next() {
		var c Comment
		//has to be nullable because comments can be top level
		var parent sql.NullInt64
		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.UserID,
			&parent,
			&c.Body,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return ListCommentsResult{}, fmt.Errorf("scan comment: %w", err)
		}
		if parent.Valid {
			v := parent.Int64
			c.ParentCommentID = &v
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return ListCommentsResult{}, fmt.Errorf("rows comments: %w", err)
	}
	return ListCommentsResult{Comments: out, Total: total}, nil
}

type CreateCommentInput struct {
	PostID          int64
	UserID          int64
	ParentCommentID *int64
	Body            string
}

func CreateComment(ctx context.Context, db *sql.DB, in CreateCommentInput) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	q := `INSERT INTO comments (post_id, user_id, parent_comment_id, body, created_at, updated_at)
    	 VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`
	res, err := db.ExecContext(ctx, q, in.PostID, in.UserID, in.ParentCommentID, in.Body)
	if err != nil {
		return 0, fmt.Errorf("create comment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func GetComment(ctx context.Context, db *sql.DB, id int64) (Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	q := `
	SELECT id, post_id, user_id, parent_comment_id, body, created_at, updated_at
	FROM comments
	WHERE id = ?
	`
	var c Comment
	var parent sql.NullInt64
	if err := db.QueryRowContext(ctx, q, id).Scan(&c.ID, &c.PostID, &c.UserID, &parent, &c.Body, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return Comment{}, err
		}
		return Comment{}, fmt.Errorf("get comment: %w", err)
	}
	if parent.Valid {
		v := parent.Int64
		c.ParentCommentID = &v
	}
	return c, nil
}
