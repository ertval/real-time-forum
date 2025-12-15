package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	sessionTimeout  = 2 * time.Second
	sessionDuration = 12 * time.Hour
)

type Session struct {
	ID        int64
	UserID    int64
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	IP        string
	UserAgent string
	IsValid   bool
}

// ---------------------------------------------------------
// PUBLIC API
// ---------------------------------------------------------

func CreateSession(
	ctx context.Context,
	db *sql.DB,
	userID int64,
	ip,
	userAgent string,
) (Session, error) {

	ctx, cancel := context.WithTimeout(ctx, sessionTimeout)
	defer cancel()

	if err := invalidateUserSessions(ctx, db, userID); err != nil {
		return Session{}, err
	}

	token := generateSessionToken()
	expiresAt := time.Now().Add(sessionDuration)

	id, err := insertSession(ctx, db, userID, token, expiresAt, ip, userAgent)
	if err != nil {
		return Session{}, err
	}

	return Session{
		ID:        id,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		IP:        ip,
		UserAgent: userAgent,
		IsValid:   true,
	}, nil
}

func GetSessionByToken(
	ctx context.Context,
	db *sql.DB,
	token string,
) (Session, error) {

	ctx, cancel := context.WithTimeout(ctx, sessionTimeout)
	defer cancel()

	return fetchValidSession(ctx, db, token)
}

func InvalidateSessionByToken(
	ctx context.Context,
	db *sql.DB,
	token string,
) error {

	ctx, cancel := context.WithTimeout(ctx, sessionTimeout)
	defer cancel()

	if _, err := db.ExecContext(ctx,
		`UPDATE sessions SET is_valid = 0 WHERE token = ?`,
		token,
	); err != nil {
		return fmt.Errorf("invalidate session: %w", err)
	}

	return nil
}

// ---------------------------------------------------------
// HELPERS
// ---------------------------------------------------------

func invalidateUserSessions(
	ctx context.Context,
	db *sql.DB,
	userID int64,
) error {

	if _, err := db.ExecContext(ctx,
		`UPDATE sessions SET is_valid = 0 WHERE user_id = ? AND is_valid = 1`,
		userID,
	); err != nil {
		return fmt.Errorf("invalidate old sessions: %w", err)
	}

	return nil
}

func insertSession(
	ctx context.Context,
	db *sql.DB,
	userID int64,
	token string,
	expiresAt time.Time,
	ip,
	userAgent string,
) (int64, error) {

	result, err := db.ExecContext(ctx,
		`INSERT INTO sessions (user_id, token, expires_at, ip, user_agent)
		 VALUES (?, ?, ?, ?, ?)`,
		userID,
		token,
		expiresAt.UTC().Format(time.RFC3339),
		ip,
		userAgent,
	)
	if err != nil {
		return 0, fmt.Errorf("create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

func fetchValidSession(
	ctx context.Context,
	db *sql.DB,
	token string,
) (Session, error) {

	const query = `
		SELECT id, user_id, token, created_at, expires_at, ip, user_agent
		FROM sessions
		WHERE token = ?
		  AND is_valid = 1
		  AND expires_at > datetime('now')
	`

	var session Session
	var createdRaw, expiresRaw string

	err := db.QueryRowContext(ctx, query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&createdRaw,
		&expiresRaw,
		&session.IP,
		&session.UserAgent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Session{}, fmt.Errorf("session not found")
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}

	session.CreatedAt, _ = time.Parse(time.RFC3339, createdRaw)
	session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresRaw)
	session.IsValid = true

	return session, nil
}

func generateSessionToken() string {
	return uuid.New().String()
}
