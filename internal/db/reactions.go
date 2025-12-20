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
func ToggleReaction(
	ctx context.Context,
	db *sql.DB,
	userID,
	objectID int64,
	targetReaction int,
	targetType string,

) (int, error) {

	ctx, cancel := context.WithTimeout(ctx, reactionTimeout)
	defer cancel()

	if targetReaction != 1 && targetReaction != -1 {
		return 0, fmt.Errorf("invalid target: %d", targetReaction)
	}

	current, err := getReactionValue(ctx, db, userID, objectID, targetType)
	if err != nil {
		return 0, err
	}

	newValue, err := applyReactionToggle(ctx, db, userID, objectID, current, targetReaction, targetType)
	if err != nil {
		return 0, err
	}

	return newValue, nil
}

// ---------------------------------------------------------
// HELPERS
// ---------------------------------------------------------

// targetType is an indicator showing if object is "post" or "comment"
func getReactionValue(
	ctx context.Context,
	db *sql.DB,
	userID,
	objectID int64,
	targetType string,
) (int, error) {

	var value int
	switch targetType {
	case "post":
		err := db.QueryRowContext(
			ctx,
			`SELECT value FROM reactions WHERE user_id = ? AND post_id = ?`,
			userID, objectID,
		).Scan(&value)

		if err == sql.ErrNoRows {
			return 0, nil
		}
		if err != nil {
			return 0, fmt.Errorf("select reaction: %w", err)
		}
	case "comment":
		err := db.QueryRowContext(
			ctx,
			`SELECT value FROM reactions WHERE user_id = ? AND comment_id = ?`,
			userID, objectID,
		).Scan(&value)

		if err == sql.ErrNoRows {
			return 0, nil
		}
		if err != nil {
			return 0, fmt.Errorf("select reaction: %w", err)
		}
	default:
		return 0, fmt.Errorf("invalid target type: %s", targetType)

	}
	return value, nil
}

// applyReactionToggle toggles like/dislike on a users post or comment.
// Target = 1 (like) or -1 (dislike)
// If current value = target, it deletes reaction (untoggle).
// targetType can be "post" or "comment"
func applyReactionToggle(
	ctx context.Context,
	db *sql.DB,
	userID,
	objectID int64,
	current,
	targetReaction int,
	targetType string,
) (int, error) {

	var idColumn string

	switch targetType {
	case "post":
		idColumn = "post_id"
	case "comment":
		idColumn = "comment_id"
	default:
		return 0, fmt.Errorf("invalid targetType: %s", targetType)
	}

	if current == targetReaction {
		query := fmt.Sprintf(
			`DELETE FROM reactions WHERE user_id = ? AND %s = ?`,
			idColumn,
		)

		_, err := db.ExecContext(ctx, query, userID, objectID)
		if err != nil {
			return 0, fmt.Errorf("delete reaction: %w", err)
		}
		return 0, nil
	}

	query := fmt.Sprintf(`
		INSERT INTO reactions (user_id, %s, value, created_at)
		VALUES (?, ?, ?, datetime('now'))
		ON CONFLICT(user_id, %s)
		DO UPDATE SET value = excluded.value
	`, idColumn, idColumn)

	_, err := db.ExecContext(
		ctx,
		query,
		userID,
		objectID,
		targetReaction,
	)
	if err != nil {
		return 0, fmt.Errorf("upsert reaction: %w", err)
	}

	return targetReaction, nil
}
