package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const reactionTimeout = 2 * time.Second

// TogglePostLike toggles a user's reaction on a post.
// Target == 1 (like) or -1 (dislike)
func TogglePostReaction(
	ctx context.Context,
	db *sql.DB,
	userID,
	postID int64,
	target int,
) (int, error) {

	ctx, cancel := context.WithTimeout(ctx, reactionTimeout)
	defer cancel()

	if target != 1 && target != -1 {
		return 0, fmt.Errorf("invalid target: %d", target)
	}

	current, err := getReactionValue(ctx, db, userID, postID)
	if err != nil {
		return 0, err
	}

	newValue, err := applyReactionToggle(ctx, db, userID, postID, current, target)
	if err != nil {
		return 0, err
	}

	return newValue, nil
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

// applyReactionToggle toggles like/dislike on a users post.
// Target = 1 (like) or -1 (dislike)
// If current value = target, it deletes reaction (untoggle).
func applyReactionToggle(
	ctx context.Context,
	db *sql.DB,
	userID,
	postID int64,
	current,
	target int,
) (int, error) {
	if current == target {
		// Untoggle: delete reaction
		_, err := db.ExecContext(ctx, `DELETE FROM reactions WHERE user_id = ? AND post_id = ?`, userID, postID)
		if err != nil {
			return 0, fmt.Errorf("delete reaction: %w", err)
		}
		return 0, nil
	} else {
		// Set to target: insert or update
		_, err := db.ExecContext(ctx, `
            INSERT INTO reactions (user_id, post_id, value, created_at)
            VALUES (?, ?, ?, datetime('now'))
            ON CONFLICT(user_id, post_id) DO UPDATE SET value = excluded.value
        `, userID, postID, target)
		if err != nil {
			return 0, fmt.Errorf("upsert reaction: %w", err)
		}
		return target, nil
	}
}
