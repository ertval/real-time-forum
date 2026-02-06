// internal/db/reactions.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const reactionTimeout = 2 * time.Second

// ToggleReaction toggles a user's reaction on a post or comment.
// targetReaction = 1 (like) or -1 (dislike)
func ToggleReaction(
	ctx context.Context,
	db *sql.DB,
	userID,
	objectID int64,
	targetReaction int,
	targetType string,
) (int, error) {

	if targetReaction != 1 && targetReaction != -1 {
		return 0, fmt.Errorf("invalid target reaction: %d", targetReaction)
	}

	ctx, cancel := context.WithTimeout(ctx, reactionTimeout)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Check target exists
	switch targetType {
	case "post":
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM posts WHERE id = ?`, objectID).Scan(new(int)); err != nil {
			if err == sql.ErrNoRows {
				return 0, ErrNotFound
			}
			return 0, fmt.Errorf("check post exists: %w", err)
		}
	case "comment":
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM comments WHERE id = ?`, objectID).Scan(new(int)); err != nil {
			if err == sql.ErrNoRows {
				return 0, ErrNotFound
			}
			return 0, fmt.Errorf("check comment exists: %w", err)
		}
	default:
		return 0, fmt.Errorf("invalid target type: %s", targetType)
	}

	current, err := getReactionValueTx(ctx, tx, userID, objectID, targetType)
	if err != nil {
		return 0, err
	}

	newValue, err := applyReactionToggleTx(
		ctx,
		tx,
		userID,
		objectID,
		current,
		targetReaction,
		targetType,
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return newValue, nil
}

/*-------------------
  HELPERS (TX SAFE)
-------------------*/

func getReactionValueTx(
	ctx context.Context,
	tx *sql.Tx,
	userID,
	objectID int64,
	targetType string,
) (int, error) {

	var value int
	var err error

	switch targetType {
	case "post":
		err = tx.QueryRowContext(
			ctx,
			`SELECT value FROM reactions WHERE user_id = ? AND post_id = ?`,
			userID, objectID,
		).Scan(&value)

	case "comment":
		err = tx.QueryRowContext(
			ctx,
			`SELECT value FROM reactions WHERE user_id = ? AND comment_id = ?`,
			userID, objectID,
		).Scan(&value)

	default:
		return 0, fmt.Errorf("invalid target type: %s", targetType)
	}

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("select reaction: %w", err)
	}

	return value, nil
}

func applyReactionToggleTx(
	ctx context.Context,
	tx *sql.Tx,
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
		return 0, fmt.Errorf("invalid target type: %s", targetType)
	}

	if current == targetReaction {
		query := fmt.Sprintf(
			`DELETE FROM reactions WHERE user_id = ? AND %s = ?`,
			idColumn,
		)

		if _, err := tx.ExecContext(ctx, query, userID, objectID); err != nil {
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

	if _, err := tx.ExecContext(
		ctx,
		query,
		userID,
		objectID,
		targetReaction,
	); err != nil {
		return 0, fmt.Errorf("upsert reaction: %w", err)
	}

	return targetReaction, nil
}
