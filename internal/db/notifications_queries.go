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
		SELECT 
			n.id,
			n.recipient_id,
			n.actor_id,
			u.username AS actor_username,
			n.type,

			-- ALWAYS RETURN THE REAL POST ID
			COALESCE(n.post_id, p.id) AS post_id,

			n.comment_id,
			n.created_at,
			n.is_read
		FROM notifications n
		JOIN users u ON u.id = n.actor_id
		LEFT JOIN comments c ON c.id = n.comment_id
		LEFT JOIN posts p ON p.id = c.post_id
		WHERE n.recipient_id = ?
		ORDER BY n.created_at DESC
	`, userID)
	if err != nil {
		return ListNotificationsResult{}, err
	}
	defer rows.Close()

	var list []Notification

	for rows.Next() {
		var n Notification
		var postID sql.NullInt64
		var commentID sql.NullInt64
		var isRead int

		err := rows.Scan(
			&n.ID,
			&n.RecipientID,
			&n.ActorID,
			&n.ActorUsername,
			&n.Type,
			&postID,
			&commentID,
			&n.CreatedAt,
			&isRead,
		)
		if err != nil {
			return ListNotificationsResult{}, err
		}

		if postID.Valid {
			n.PostID = &postID.Int64
		}
		if commentID.Valid {
			n.CommentID = &commentID.Int64
		}

		n.IsRead = isRead == 1
		list = append(list, n)
	}

	var unread int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM notifications 
		WHERE recipient_id = ? AND is_read = 0
	`, userID).Scan(&unread)
	if err != nil {
		return ListNotificationsResult{}, err
	}

	return ListNotificationsResult{
		Notifications: list,
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
