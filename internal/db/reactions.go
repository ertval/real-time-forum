package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// TogglePostLike toggles a "like" for a given user & post.
// If no reaction exists it inserts like (value = 1)
// If there is a like (1) it removes it (unlike)
// If there is a dislike (-1) it changes it to like (1)
// It returns whether the post is liked after the toggle, and the total like count.
func TogglePostLike(ctx context.Context, db *sql.DB, userID, postID int64) (liked bool, likeCount int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var value int
	err = db.QueryRowContext(ctx,
		`SELECT value FROM reactions WHERE user_id = ? AND post_id = ?`,
		userID, postID).Scan(&value)

	switch {
	//No reaction yet
	case err == sql.ErrNoRows:
		q := `INSERT INTO reactions (user_id, post_id, value, created_at) VALUES (?, ?, 1, datetime('now'))`
		_, err = db.ExecContext(ctx, q, userID, postID)

		if err != nil {
			return false, 0, fmt.Errorf("insert reaction: %w", err)
		}
		liked = true
	//Some other DB error
	case err != nil:
		return false, 0, fmt.Errorf("select reaction: %w", err)
	//Reaction already exists
	default:
		//Already liked
		if value == 1 {
			q := `DELETE FROM reactions WHERE user_id = ? AND post_id = ?`
			_, err = db.ExecContext(ctx, q, userID, postID)
			if err != nil {
				return false, 0, fmt.Errorf("delete reaction: %w", err)
			}
			liked = false
			//Already disliked
		} else if value == -1 {
			q := `UPDATE reactions SET value = 1 WHERE user_id = ? AND post_id = ?`
			_, err = db.ExecContext(ctx, q, userID, postID)
			if err != nil {
				return false, 0, fmt.Errorf("update reaction: %w", err)
			}
			liked = true
		} else {
			return false, 0, fmt.Errorf("unexpected reaction value: %d", value)
		}
	}
	q := `SELECT COUNT(*) FROM reactions WHERE post_id = ? AND value = 1`
	err = db.QueryRowContext(ctx, q, postID).Scan(&likeCount)
	if err != nil {
		return liked, 0, fmt.Errorf("count likes: %w", err)
	}
	return liked, likeCount, nil
}
