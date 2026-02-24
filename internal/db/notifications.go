// internal/db/notifications.go
package db

import (
	"context"
	"database/sql"
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

func InsertNotification(
	ctx context.Context,
	db *sql.DB,
	recipientID int64,
	actorID int64,
	notificationType string,
	postID *int64,
	commentID *int64,
) error {

	// Do not notify yourself
	if recipientID == actorID {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, notificationTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		INSERT OR IGNORE INTO notifications
		(recipient_id, actor_id, type, post_id, comment_id)
		VALUES (?, ?, ?, ?, ?)
	`,
		recipientID,
		actorID,
		notificationType,
		postID,
		commentID,
	)

	return err
}

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

	if targetType == "post" {
		postID = &objectID
		if newValue == 1 {
			notificationType = "post_like"
		} else {
			notificationType = "post_dislike"
		}
	} else {
		commentID = &objectID
		if newValue == 1 {
			notificationType = "post_like"
		} else {
			notificationType = "post_dislike"
		}
	}

	_, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO notifications
		(recipient_id, actor_id, type, post_id, comment_id)
		VALUES (?, ?, ?, ?, ?)
	`,
		ownerID,
		userID,
		notificationType,
		postID,
		commentID,
	)

	return err
}
