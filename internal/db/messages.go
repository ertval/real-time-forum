// internal/db/messages.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type PrivateMessage struct {
	ID          int64  `json:"id"`
	SenderID    int64  `json:"sender_id"`
	RecipientID int64  `json:"recipient_id"`
	Body        string `json:"body"`
	CreatedAt   string `json:"created_at"`
}

type CreateMessageRequest struct {
	SenderID    int64
	RecipientID int64
	Body        string
}

// CreateMessage persists a new private message and returns the stored record.
func CreateMessage(ctx context.Context, database *sql.DB, req CreateMessageRequest) (PrivateMessage, error) {
	if req.SenderID == req.RecipientID {
		return PrivateMessage{}, fmt.Errorf("sender and recipient must be different users")
	}
	if strings.TrimSpace(req.Body) == "" {
		return PrivateMessage{}, fmt.Errorf("message body cannot be empty")
	}

	result, err := database.ExecContext(ctx,
		`INSERT INTO private_messages (sender_id, recipient_id, body)
		 VALUES (?, ?, ?)`,
		req.SenderID, req.RecipientID, req.Body,
	)
	if err != nil {
		return PrivateMessage{}, fmt.Errorf("create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return PrivateMessage{}, fmt.Errorf("last insert id: %w", err)
	}

	var msg PrivateMessage
	err = database.QueryRowContext(ctx,
		`SELECT id, sender_id, recipient_id, body, created_at
		 FROM private_messages WHERE id = ?`, id,
	).Scan(&msg.ID, &msg.SenderID, &msg.RecipientID, &msg.Body, &msg.CreatedAt)
	if err != nil {
		return PrivateMessage{}, fmt.Errorf("fetch created message: %w", err)
	}

	return msg, nil
}

// GetMessageHistory returns up to 10 messages exchanged between two users,
// ordered oldest-first (render-ready). Pass beforeID > 0 to paginate backwards
// through history; pass 0 to get the latest 10.
func GetMessageHistory(ctx context.Context, database *sql.DB, userA, userB, beforeID int64) ([]PrivateMessage, error) {
	const limit = 10

	var rows *sql.Rows
	var err error

	if beforeID > 0 {
		rows, err = database.QueryContext(ctx,
			`SELECT id, sender_id, recipient_id, body, created_at FROM private_messages
			 WHERE sender_id = ? AND recipient_id = ? AND id < ?
			 UNION ALL
			 SELECT id, sender_id, recipient_id, body, created_at FROM private_messages
			 WHERE sender_id = ? AND recipient_id = ? AND id < ?
			 ORDER BY id DESC
			 LIMIT ?`,
			userA, userB, beforeID, userB, userA, beforeID, limit,
		)
	} else {
		rows, err = database.QueryContext(ctx,
			`SELECT id, sender_id, recipient_id, body, created_at FROM private_messages
			 WHERE sender_id = ? AND recipient_id = ?
			 UNION ALL
			 SELECT id, sender_id, recipient_id, body, created_at FROM private_messages
			 WHERE sender_id = ? AND recipient_id = ?
			 ORDER BY id DESC
			 LIMIT ?`,
			userA, userB, userB, userA, limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("get message history: %w", err)
	}
	defer rows.Close()

	var msgs []PrivateMessage
	for rows.Next() {
		var m PrivateMessage
		if err := rows.Scan(&m.ID, &m.SenderID, &m.RecipientID, &m.Body, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Reverse to chronological order (oldest first)
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}
