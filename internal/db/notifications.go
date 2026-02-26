package db

import (
	"context"
	"database/sql"
	"time"
)

const notificationTimeout = 2 * time.Second

type Notification struct {
	ID            int64  `json:"id"`
	RecipientID   int64  `json:"recipient_id"`
	ActorID       int64  `json:"actor_id"`
	ActorUsername string `json:"actor_username"`
	Type          string `json:"type"`
	PostID        *int64 `json:"post_id,omitempty"`
	CommentID     *int64 `json:"comment_id,omitempty"`
	CreatedAt     string `json:"created_at"`
	IsRead        bool   `json:"is_read"`
}

/* =========================================================
   MAIN INSERT (POSTS & COMMENTS – NOT REACTIONS)
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

	query := `
		INSERT INTO notifications
			(recipient_id, actor_id, type, post_id, comment_id, created_at, is_read)
		VALUES (?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%SZ','now'), 0)
	`

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
   REACTION NOTIFICATIONS (LIKE/DISLIKE)
   — ALWAYS INSERT NEW, NEVER CONFLICT
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

	query := `
		INSERT INTO notifications
			(recipient_id, actor_id, type, post_id, comment_id, created_at, is_read)
		VALUES (?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%SZ','now'), 0)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		ownerID,
		userID,
		notificationType,
		postID,
		commentID,
	)

	return err
}
