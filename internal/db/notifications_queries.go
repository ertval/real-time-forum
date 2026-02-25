// internal/db/notifications_queries.go
package db

import (
	"context"
	"database/sql"
	"time"
)

const notificationsQueryTimeout = 2 * time.Second

type ListNotificationsResult struct {
	Notifications []Notification `json:"notifications"`
	UnreadCount   int            `json:"unread_count"`
}

func ListUserNotifications(
	ctx context.Context,
	db *sql.DB,
	userID int64,
) (ListNotificationsResult, error) {

	ctx, cancel := context.WithTimeout(ctx, notificationsQueryTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT id, recipient_id, actor_id, type, post_id, comment_id, created_at, is_read
		FROM notifications
		WHERE recipient_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return ListNotificationsResult{}, err
	}
	defer rows.Close()

	var result []Notification

	for rows.Next() {
		var n Notification
		var postID sql.NullInt64
		var commentID sql.NullInt64
		var isRead int

		if err := rows.Scan(
			&n.ID,
			&n.RecipientID,
			&n.ActorID,
			&n.Type,
			&postID,
			&commentID,
			&n.CreatedAt,
			&isRead,
		); err != nil {
			return ListNotificationsResult{}, err
		}

		if postID.Valid {
			n.PostID = &postID.Int64
		}
		if commentID.Valid {
			n.CommentID = &commentID.Int64
		}

		n.IsRead = isRead == 1

		result = append(result, n)
	}

	var unread int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE recipient_id = ? AND is_read = 0`,
		userID,
	).Scan(&unread)

	if err != nil {
		return ListNotificationsResult{}, err
	}

	return ListNotificationsResult{
		Notifications: result,
		UnreadCount:   unread,
	}, nil
}

func MarkNotificationRead(
	ctx context.Context,
	db *sql.DB,
	userID,
	notificationID int64,
) error {

	ctx, cancel := context.WithTimeout(ctx, notificationsQueryTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = 1
		WHERE id = ? AND recipient_id = ?
	`, notificationID, userID)

	return err
}

func MarkAllNotificationsRead(
	ctx context.Context,
	db *sql.DB,
	userID int64,
) error {

	ctx, cancel := context.WithTimeout(ctx, notificationsQueryTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = 1
		WHERE recipient_id = ?
	`, userID)

	return err
}
