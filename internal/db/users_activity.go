// internal/db/users_activity.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type UserActivityCommentPost struct {
	ID       int64  `json:"id"`
	AuthorID int64  `json:"author_id"`
	Author   string `json:"author"`
	Title    string `json:"title"`
}

type UserActivityComment struct {
	ID              int64                   `json:"id"`
	PostID          int64                   `json:"post_id"`
	UserID          int64                   `json:"user_id"`
	Username        string                  `json:"username"`
	ParentCommentID *int64                  `json:"parent_comment_id,omitempty"`
	Body            string                  `json:"body"`
	ImageURL        *string                 `json:"image_url"`
	CreatedAt       string                  `json:"created_at"`
	UpdatedAt       string                  `json:"updated_at,omitempty"`
	Likes           int                     `json:"likes"`
	Dislikes        int                     `json:"dislikes"`
	Post            UserActivityCommentPost `json:"post"`
}

type ListUserCommentsWithPostParams struct {
	UserID  int64
	Page    int
	PerPage int
}

type ListUserCommentsWithPostResult struct {
	Comments []UserActivityComment
	Total    int
}

func ListUserCommentsWithPost(
	ctx context.Context,
	db *sql.DB,
	p ListUserCommentsWithPostParams,
) (ListUserCommentsWithPostResult, error) {
	p.Page, p.PerPage = normalizePagination(p.Page, p.PerPage)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	total, err := countUserComments(ctx, db, p.UserID)
	if err != nil {
		return ListUserCommentsWithPostResult{}, err
	}

	comments, err := fetchUserCommentsWithPost(ctx, db, p)
	if err != nil {
		return ListUserCommentsWithPostResult{}, err
	}

	return ListUserCommentsWithPostResult{
		Comments: comments,
		Total:    total,
	}, nil
}

func countUserComments(ctx context.Context, db *sql.DB, userID int64) (int, error) {
	var total int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM comments WHERE user_id = ?`,
		userID,
	).Scan(&total); err != nil {
		return 0, fmt.Errorf("count user comments: %w", err)
	}
	return total, nil
}

func fetchUserCommentsWithPost(
	ctx context.Context,
	db *sql.DB,
	p ListUserCommentsWithPostParams,
) ([]UserActivityComment, error) {
	offset := (p.Page - 1) * p.PerPage

	rows, err := db.QueryContext(ctx, `
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			cu.username,
			c.parent_comment_id,
			c.body,
			c.image_url,
			c.created_at,
			c.updated_at,
			post.id,
			post.author_id,
			pu.username,
			post.title,
			IFNULL(rc.likes, 0) AS likes,
			IFNULL(rc.dislikes, 0) AS dislikes
		FROM comments c
		JOIN users cu ON cu.id = c.user_id
		JOIN posts post ON post.id = c.post_id
		JOIN users pu ON pu.id = post.author_id
		LEFT JOIN (
			SELECT
				comment_id,
				SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END) AS likes,
				SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END) AS dislikes
			FROM reactions
			WHERE comment_id IS NOT NULL
			GROUP BY comment_id
		) rc ON rc.comment_id = c.id
		WHERE c.user_id = ?
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`, p.UserID, p.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query user comments: %w", err)
	}
	defer rows.Close()

	comments := make([]UserActivityComment, 0)

	for rows.Next() {
		var comment UserActivityComment
		var parentID sql.NullInt64
		var imageURL sql.NullString

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&parentID,
			&comment.Body,
			&imageURL,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.Post.ID,
			&comment.Post.AuthorID,
			&comment.Post.Author,
			&comment.Post.Title,
			&comment.Likes,
			&comment.Dislikes,
		); err != nil {
			return nil, fmt.Errorf("scan user comment: %w", err)
		}

		if parentID.Valid {
			id := parentID.Int64
			comment.ParentCommentID = &id
		}
		if imageURL.Valid {
			comment.ImageURL = &imageURL.String
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows user comments: %w", err)
	}

	return comments, nil
}
