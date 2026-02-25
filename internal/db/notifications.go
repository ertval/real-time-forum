// internal/db/notifications.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const notificationTimeout = 2 * time.Second

type Notification struct {
	ID          int64  `json:"id"`
	RecipientID int64  `json:"recipient_id"`
	ActorID     int64  `json:"actor_id"`
	Type        string `json:"type"`
	PostID      *int64 `json:"post_id,omitempty"`
	CommentID   *int64 `json:"comment_id,omitempty"`
	CreatedAt   string `json:"created_at"`
	IsRead      bool   `json:"is_read"`
}

/* =========================================================
   PUBLIC INSERT (non-transactional)
========================================================= */

func InsertNotification(
	ctx context.Context,
	db *sql.DB,
	recipientID int64,
	actorID int64,
	notificationType string,
	postID *int64,
	commentID *int64,
) error {

	if recipientID == actorID {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, notificationTimeout)
	defer cancel()

	var conflictTarget string

	if postID != nil {
		conflictTarget = `
			ON CONFLICT(actor_id, recipient_id, type, post_id)
			WHERE post_id IS NOT NULL
			DO UPDATE SET
				created_at = excluded.created_at,
				is_read = 0
		`
	} else {
		conflictTarget = `
			ON CONFLICT(actor_id, recipient_id, type, comment_id)
			WHERE comment_id IS NOT NULL
			DO UPDATE SET
				created_at = excluded.created_at,
				is_read = 0
		`
	}

	query := fmt.Sprintf(`
		INSERT INTO notifications
		(recipient_id, actor_id, type, post_id, comment_id, created_at, is_read)
		VALUES (?, ?, ?, ?, ?, strftime('%%Y-%%m-%%dT%%H:%%M:%%SZ','now'), 0)
		%s
	`, conflictTarget)

	_, err := db.ExecContext(ctx, query,
		recipientID,
		actorID,
		notificationType,
		postID,
		commentID,
	)

	return err
}

/* =========================================================
   TX VERSION (used inside ToggleReaction)
========================================================= */

func handleReactionNotificationTx(
	ctx context.Context,
	tx *sql.Tx,
	ownerID,
	userID,
	objectID int64,
	newValue int,
	targetType string,
) error {

	if ownerID == userID {
		return nil
	}

	var notificationType string
	var postID *int64
	var commentID *int64

	switch targetType {

	case "post":
		postID = &objectID
		if newValue == 1 {
			notificationType = "post_like"
		} else {
			notificationType = "post_dislike"
		}

	case "comment":
		commentID = &objectID
		if newValue == 1 {
			notificationType = "comment_like"
		} else {
			notificationType = "comment_dislike"
		}

	default:
		return nil
	}

	var conflictTarget string

	if postID != nil {
		conflictTarget = `
			ON CONFLICT(actor_id, recipient_id, type, post_id)
			WHERE post_id IS NOT NULL
			DO UPDATE SET
				created_at = excluded.created_at,
				is_read = 0
		`
	} else {
		conflictTarget = `
			ON CONFLICT(actor_id, recipient_id, type, comment_id)
			WHERE comment_id IS NOT NULL
			DO UPDATE SET
				created_at = excluded.created_at,
				is_read = 0
		`
	}

	query := fmt.Sprintf(`
		INSERT INTO notifications
		(recipient_id, actor_id, type, post_id, comment_id, created_at, is_read)
		VALUES (?, ?, ?, ?, ?, strftime('%%Y-%%m-%%dT%%H:%%M:%%SZ','now'), 0)
		%s
	`, conflictTarget)

	_, err := tx.ExecContext(ctx, query,
		ownerID,
		userID,
		notificationType,
		postID,
		commentID,
	)

	return err
}
