package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const reactionTimeout = 2 * time.Second

// TogglePostLike toggles a user's like on a post.
//
// Behavior:
//   - no reaction  → like
//   - like exists  → unlike
//   - dislike      → switch to like
//
// Returns whether the post is liked AFTER the operation.
func TogglePostLike(
	ctx context.Context,
	db *sql.DB,
	userID,
	postID int64,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, reactionTimeout)
	defer cancel()

	current, err := getReactionValue(ctx, db, userID, postID)
	if err != nil {
		return false, err
	}

	newValue, err := applyReactionToggle(ctx, db, userID, postID, current)
	if err != nil {
		return false, err
	}

	return newValue == 1, nil
}

// ---------------------------------------------------------
// HELPERS
// ---------------------------------------------------------

func getReactionValue(
	ctx context.Context,
	db *sql.DB,
	userID,
	postID int64,
) (int, error) {

	var value int
	err := db.QueryRowContext(
		ctx,
		`SELECT value FROM reactions WHERE user_id = ? AND post_id = ?`,
		userID, postID,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("select reaction: %w", err)
	}

	return value, nil
}

func applyReactionToggle(
	ctx context.Context,
	db *sql.DB,
	userID,
	postID int64,
	current int,
) (int, error) {

	switch current {

	case 0:
		_, err := db.ExecContext(ctx, `
			INSERT INTO reactions (user_id, post_id, value, created_at)
			VALUES (?, ?, 1, datetime('now'))
		`, userID, postID)
		if err != nil {
			return 0, fmt.Errorf("insert reaction: %w", err)
		}
		return 1, nil

	case 1:
		_, err := db.ExecContext(ctx,
			`DELETE FROM reactions WHERE user_id = ? AND post_id = ?`,
			userID, postID,
		)
		if err != nil {
			return 0, fmt.Errorf("delete reaction: %w", err)
		}
		return 0, nil

	case -1:
		_, err := db.ExecContext(ctx,
			`UPDATE reactions SET value = 1 WHERE user_id = ? AND post_id = ?`,
			userID, postID,
		)
		if err != nil {
			return 0, fmt.Errorf("update reaction: %w", err)
		}
		return 1, nil

	default:
		return 0, fmt.Errorf("unexpected reaction value: %d", current)
	}
}

func CountPostLikes(ctx context.Context, db *sql.DB, postID int64) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	const query = `SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND value = 1`
	var count int
	err := db.QueryRowContext(ctx, query, postID).Scan(&count)
	return count, err
}
