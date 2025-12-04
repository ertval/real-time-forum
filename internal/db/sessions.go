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
	CreatedAt string
	ExpiresAt string
	IP        string
	UserAgent string
	IsValid   int
}

func CreateSession(ctx context.Context, db *sql.DB, userID int64, ip, userAgent string) (Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	token := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour * 12)
	expiresAtStr := expiresAt.Format("02-01-2006 15:04")
	result, err := db.ExecContext(ctx, "INSERT INTO sessions (user_id, token, expires_at, ip, user_agent) VALUES (?, ?, ?, ?, ?)", userID, token, expiresAtStr, ip, userAgent)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Session{}, fmt.Errorf("last insert id: %w", err)
	}
	return Session{ID: id, UserID: userID, Token: token, ExpiresAt: expiresAtStr, IP: ip, UserAgent: userAgent}, nil
}

func GetSessionByToken(ctx context.Context, db *sql.DB, token string) (Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	q := `SELECT id, user_id, token, created_at, expires_at, ip, user_agent, is_valid FROM sessions WHERE token = ? AND is_valid = 1 AND expires_at > strftime('%d-%m-%Y %H:%M', 'now')`
	var s Session
	row := db.QueryRowContext(ctx, q, token)
	if err := row.Scan(&s.ID, &s.UserID, &s.Token, &s.CreatedAt, &s.ExpiresAt, &s.IP, &s.UserAgent, &s.IsValid); err != nil {
		if err == sql.ErrNoRows {
			return Session{}, fmt.Errorf("session not found")
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}
	return s, nil

}
