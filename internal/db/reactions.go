package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// TogglePostLike toggles a user's "like" on a post.
// - If no reaction exists → insert like (1)
// - If like exists (1) → remove it (unlike)
// - If dislike exists (-1) → switch to like (1)
// Returns:
//
//	liked (bool)   — whether it is liked after the toggle
//	likeCount (int) — total number of likes for the post
func TogglePostLike(ctx context.Context, db *sql.DB, userID, postID int64) (bool, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// ----------------------------------------
	// 1. Read existing reaction (if any)
	// ----------------------------------------
	var value int
	err := db.QueryRowContext(ctx,
		`SELECT value FROM reactions WHERE user_id = ? AND post_id = ?`,
		userID, postID,
	).Scan(&value)

	switch {
	case err == sql.ErrNoRows:
		// ----------------------------------------
		// No prior reaction → insert like
		// ----------------------------------------
		_, execErr := db.ExecContext(ctx, `
			INSERT INTO reactions (user_id, post_id, value, created_at)
			VALUES (?, ?, 1, datetime('now'))
		`, userID, postID)
		if execErr != nil {
			return false, 0, fmt.Errorf("insert reaction: %w", execErr)
		}
		value = 1 // reaction is now like

	case err != nil:
		// Unexpected DB error
		return false, 0, fmt.Errorf("select reaction: %w", err)

	default:
		// ----------------------------------------
		// Reaction exists → toggle it
		// ----------------------------------------
		switch value {
		case 1:
			// Removing a like (unlike)
			_, execErr := db.ExecContext(ctx,
				`DELETE FROM reactions WHERE user_id = ? AND post_id = ?`,
				userID, postID,
			)
			if execErr != nil {
				return false, 0, fmt.Errorf("delete reaction: %w", execErr)
			}
			value = 0 // no reaction exists now

		case -1:
			// Switching dislike → like
			_, execErr := db.ExecContext(ctx,
				`UPDATE reactions SET value = 1 WHERE user_id = ? AND post_id = ?`,
				userID, postID,
			)
			if execErr != nil {
				return false, 0, fmt.Errorf("update reaction: %w", execErr)
			}
			value = 1 // now liked

		default:
			// Should never happen unless data corruption
			return false, 0, fmt.Errorf("unexpected reaction value: %d", value)
		}
	}

	// ----------------------------------------
	// 2. Count likes
	// ----------------------------------------
	var likeCount int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM reactions WHERE post_id = ? AND value = 1`,
		postID,
	).Scan(&likeCount)
	if err != nil {
		return value == 1, 0, fmt.Errorf("count likes: %w", err)
	}

	return value == 1, likeCount, nil
}
