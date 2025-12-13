package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const reactionTimeout = 2 * time.Second

// ---------------------------------------------------------
// PUBLIC API
// ---------------------------------------------------------

// TogglePostLike toggles a user's "like" on a post.
//
// Behavior:
//   - No reaction        → insert like
//   - Existing like      → remove (unlike)
//   - Existing dislike   → switch to like
//
// Returns:
//
//	liked     — whether the post is liked after the toggle
//	likeCount — total number of likes for the post
func TogglePostLike(
	ctx context.Context,
	db *sql.DB,
	userID,
	postID int64,
) (liked bool, likeCount int, err error) {

	ctx, cancel := context.WithTimeout(ctx, reactionTimeout)
	defer cancel()

	value, err := getReactionValue(ctx, db, userID, postID)
	if err != nil {
		return false, 0, err
	}

	value, err = applyReactionToggle(ctx, db, userID, postID, value)
	if err != nil {
		return false, 0, err
	}

	likeCount, err = countPostLikes(ctx, db, postID)
	if err != nil {
		return value == 1, 0, err
	}

	return value == 1, likeCount, nil
}

//
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
	err := db.QueryRowContext(ctx,
		`SELECT value FROM reactions WHERE user_id = ? AND post_id = ?`,
		userID, postID,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return 0, nil // no reaction
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
	currentValue int,
) (int, error) {

	switch currentValue {

	case 0:
		// No reaction → insert like
		_, err := db.ExecContext(ctx, `
			INSERT INTO reactions (user_id, post_id, value, created_at)
			VALUES (?, ?, 1, datetime('now'))
		`, userID, postID)
		if err != nil {
			return 0, fmt.Errorf("insert reaction: %w", err)
		}
		return 1, nil

	case 1:
		// Like exists → remove
		_, err := db.ExecContext(ctx,
			`DELETE FROM reactions WHERE user_id = ? AND post_id = ?`,
			userID, postID,
		)
		if err != nil {
			return 0, fmt.Errorf("delete reaction: %w", err)
		}
		return 0, nil

	case -1:
		// Dislike → switch to like
		_, err := db.ExecContext(ctx,
			`UPDATE reactions SET value = 1 WHERE user_id = ? AND post_id = ?`,
			userID, postID,
		)
		if err != nil {
			return 0, fmt.Errorf("update reaction: %w", err)
		}
		return 1, nil

	default:
		return 0, fmt.Errorf("unexpected reaction value: %d", currentValue)
	}
}

func countPostLikes(
	ctx context.Context,
	db *sql.DB,
	postID int64,
) (int, error) {

	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM reactions WHERE post_id = ? AND value = 1`,
		postID,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("count likes: %w", err)
	}

	return count, nil
}
