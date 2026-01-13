// internal/db/comments_helpers.go
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
func normalizeCommentsPagination(params *ListCommentsParams) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
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

// Fetch paginated comments for a post (NO reactions here)
func fetchCommentsByPost(
	ctx context.Context,
	db *sql.DB,
	params ListCommentsParams,
) ([]Comment, error) {

	offset := (params.Page - 1) * params.PerPage

	rows, err := db.QueryContext(ctx, `
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			u.username,
			c.parent_comment_id,
			c.body,
			c.created_at,
			c.updated_at
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
		LIMIT ? OFFSET ?
	`, params.PostID, params.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var comment Comment
		var parentID sql.NullInt64

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&parentID,
			&comment.Body,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}

		if parentID.Valid {
			id := parentID.Int64
			comment.ParentCommentID = &id
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows comments: %w", err)
	}

	return comments, nil
}

// Attach like/dislike counts to each comment
func attachCommentReactions(
	ctx context.Context,
	db *sql.DB,
	comments []Comment,
) error {

	for i := range comments {
		likes, dislikes, err := CountReactionsForComment(ctx, db, comments[i].ID)
		if err != nil {
			return fmt.Errorf("attach comment reactions: %w", err)
		}
		comments[i].Likes = likes
		comments[i].Dislikes = dislikes
	}
	return nil
}
