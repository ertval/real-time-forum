package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------
// DATA STRUCTURES
// ---------------------------------------------------------

type Comment struct {
	ID              int64  `json:"id"`
	PostID          int64  `json:"post_id"`
	UserID          int64  `json:"user_id"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
	Body            string `json:"body"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	Likes           int    `json:"likes"`
	Dislikes        int    `json:"dislikes"`
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

// ---------------------------------------------------------
// LIST COMMENTS (ORCHESTRATOR)
// ---------------------------------------------------------

func ListCommentsByPost(
	ctx context.Context,
	db *sql.DB,
	params ListCommentsParams,
) (ListCommentsResult, error) {

	normalizeCommentsPagination(&params)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	total, err := countCommentsByPost(ctx, db, params.PostID)
	if err != nil {
		return ListCommentsResult{}, err
	}

	comments, err := fetchCommentsByPost(ctx, db, params)
	if err != nil {
		return ListCommentsResult{}, err
	}

	// Attach reactions AFTER fetching comments
	if err := attachCommentReactions(ctx, db, comments); err != nil {
		return ListCommentsResult{}, err
	}

	return ListCommentsResult{
		Comments: comments,
		Total:    total,
	}, nil
}

// ---------------------------------------------------------
// CREATE COMMENT
// ---------------------------------------------------------

type CreateCommentInput struct {
	PostID          int64
	UserID          int64
	ParentCommentID *int64
	Body            string
}

func CreateComment(
	ctx context.Context,
	db *sql.DB,
	input CreateCommentInput,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const query = `
		INSERT INTO comments (post_id, user_id, parent_comment_id, body, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`

	res, err := db.ExecContext(ctx, query,
		input.PostID,
		input.UserID,
		input.ParentCommentID,
		input.Body,
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

// ---------------------------------------------------------
// GET COMMENT
// ---------------------------------------------------------

func GetComment(ctx context.Context, db *sql.DB, id int64) (Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	const query = `
		SELECT id, post_id, user_id, parent_comment_id, body, created_at, updated_at
		FROM comments
		WHERE id = ?
	`

	var comment Comment
	var parentID sql.NullInt64

	err := db.QueryRowContext(ctx, query, id).
		Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&parentID,
			&comment.Body,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
	if err != nil {
		return Comment{}, err
	}

	if parentID.Valid {
		id := parentID.Int64
		comment.ParentCommentID = &id
	}

	likes, dislikes, err := CountReactionsForComment(ctx, db, id)
	if err != nil {
		return Comment{}, err
	}
	comment.Likes = likes
	comment.Dislikes = dislikes

	return comment, nil
}

// ---------------------------------------------------------
// PATCH COMMENT
// ---------------------------------------------------------

type UpdateCommentInput struct {
	Body *string
}

func UpdateComment(ctx context.Context, db *sql.DB, id int64, in UpdateCommentInput) error {
	setParts := []string{}
	args := []any{}

	if in.Body != nil {
		setParts = append(setParts, "body = ?")
		args = append(args, *in.Body)
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = datetime('now')")
	args = append(args, id)

	query := `UPDATE comments SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// ------------------------------------------------------------
// DELETE COMMENT
// ------------------------------------------------------------

func DeleteComment(ctx context.Context, db *sql.DB, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `DELETE FROM comments WHERE id = ?`, id)
	return err
}
