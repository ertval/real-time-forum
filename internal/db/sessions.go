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

// ---------------------------------------------------------------------------
// CreateSession: invalidate any previous active session and create a new one.
// ---------------------------------------------------------------------------
func CreateSession(ctx context.Context, database *sql.DB, userID int64, ip, userAgent string) (Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Disable previous sessions for this user.
	_, err := database.ExecContext(ctx,
		`UPDATE sessions SET is_valid = 0 WHERE user_id = ? AND is_valid = 1`,
		userID,
	)
	if err != nil {
		return Session{}, fmt.Errorf("invalidate old sessions: %w", err)
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(12 * time.Hour)
	expiresAtUTC := expiresAt.UTC().Format(time.RFC3339)

	// created_at uses SQLite default timestamp
	result, err := database.ExecContext(ctx,
		`INSERT INTO sessions (user_id, token, expires_at, ip, user_agent)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, token, expiresAtUTC, ip, userAgent,
	)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}

	sessionID, err := result.LastInsertId()
	if err != nil {
		return Session{}, fmt.Errorf("last insert id: %w", err)
	}

	return Session{
		ID:        sessionID,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		IP:        ip,
		UserAgent: userAgent,
		IsValid:   1,
	}, nil
}

// ---------------------------------------------------------------------------
// GetSessionByToken: return active session if token is valid & not expired.
// ---------------------------------------------------------------------------
func GetSessionByToken(ctx context.Context, database *sql.DB, token string) (Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, token, created_at, expires_at, ip, user_agent, is_valid
		FROM sessions
		WHERE token = ?
		  AND is_valid = 1
		  AND expires_at > strftime('%d-%m-%Y %H:%M', 'now')
	`

	var sess Session
	var createdAtRaw, expiresAtRaw string

	err := database.QueryRowContext(ctx, query, token).Scan(
		&sess.ID,
		&sess.UserID,
		&sess.Token,
		&createdAtRaw,
		&expiresAtRaw,
		&sess.IP,
		&sess.UserAgent,
		&sess.IsValid,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Session{}, fmt.Errorf("session not found")
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}

	// Attempt to parse created_at (may not always be strict RFC3339)
	if parsedTime, parseErr := time.Parse(time.RFC3339, createdAtRaw); parseErr == nil {
		sess.CreatedAt = parsedTime
	}

	// expires_at is always RFC3339
	if parsedTime, parseErr := time.Parse(time.RFC3339, expiresAtRaw); parseErr == nil {
		sess.ExpiresAt = parsedTime
	}

	return sess, nil
}

// ---------------------------------------------------------------------------
// InvalidateSessionByToken: mark a session as invalid (logout).
// ---------------------------------------------------------------------------
func InvalidateSessionByToken(ctx context.Context, database *sql.DB, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := database.ExecContext(ctx,
		`UPDATE sessions SET is_valid = 0 WHERE token = ?`,
		token,
	)
	if err != nil {
		return fmt.Errorf("invalidate session: %w", err)
	}

	return nil
}
