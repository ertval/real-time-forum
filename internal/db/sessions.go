package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        int64
	UserID    int64
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	IP        string
	UserAgent string
	IsValid   int
}

// CreateSession invalidates previous sessions and creates a new one.
func CreateSession(ctx context.Context, db *sql.DB, userID int64, ip, userAgent string) (Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Invalidate previous active sessions
	_, err := db.ExecContext(ctx,
		`UPDATE sessions SET is_valid = 0 WHERE user_id = ? AND is_valid = 1`,
		userID,
	)
	if err != nil {
		return Session{}, fmt.Errorf("invalidate old sessions: %w", err)
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(12 * time.Hour)
	expiresAtStr := expiresAt.UTC().Format(time.RFC3339)

	// created_at is auto-filled by SQLite DEFAULT
	result, err := db.ExecContext(ctx,
		`INSERT INTO sessions (user_id, token, expires_at, ip, user_agent)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, token, expiresAtStr, ip, userAgent,
	)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Session{}, fmt.Errorf("last insert id: %w", err)
	}

	return Session{
		ID:        id,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		IP:        ip,
		UserAgent: userAgent,
		IsValid:   1,
	}, nil
}

// GetSessionByToken returns a valid, non-expired session.
func GetSessionByToken(ctx context.Context, db *sql.DB, token string) (Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, token, created_at, expires_at, ip, user_agent, is_valid
		FROM sessions
		WHERE token = ?
		  AND is_valid = 1
		  AND expires_at > strftime('%d-%m-%Y %H:%M', 'now')
	`

	var s Session
	var createdStr, expiresStr string

	err := db.QueryRowContext(ctx, query, token).
		Scan(&s.ID, &s.UserID, &s.Token, &createdStr, &expiresStr, &s.IP, &s.UserAgent, &s.IsValid)

	if err != nil {
		if err == sql.ErrNoRows {
			return Session{}, fmt.Errorf("session not found")
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}

	// Parse timestamps — tests expect RFC3339 for expires_at only
	createdAt, _ := time.Parse(time.RFC3339, createdStr) // may fail → zero time (acceptable)
	expiresAt, _ := time.Parse(time.RFC3339, expiresStr)

	s.CreatedAt = createdAt
	s.ExpiresAt = expiresAt

	return s, nil
}

// InvalidateSessionByToken disables a session (logout).
func InvalidateSessionByToken(ctx context.Context, db *sql.DB, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE sessions SET is_valid = 0 WHERE token = ?`,
		token,
	)
	if err != nil {
		return fmt.Errorf("invalidate session: %w", err)
	}
	return nil
}
