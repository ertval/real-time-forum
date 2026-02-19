// internal/auth/oauth.go
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func FindOrCreateOAuthUser(
	ctx context.Context,
	conn *sql.DB,
	provider string,
	providerUserID string,
	email string,
	name string,
) (int64, error) {

	if email == "" {
		return 0, fmt.Errorf("oauth provider returned empty email")
	}

	// Check existing oauth link
	var existingUserID int64
	err := conn.QueryRowContext(
		ctx,
		`SELECT user_id FROM oauth_users 
		 WHERE provider = ? AND provider_user_id = ? LIMIT 1`,
		provider,
		providerUserID,
	).Scan(&existingUserID)

	if err == nil {
		return existingUserID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("lookup oauth user: %w", err)
	}

	// Check if email already exists
	var userID int64
	err = conn.QueryRowContext(
		ctx,
		`SELECT id FROM users WHERE email = ? LIMIT 1`,
		email,
	).Scan(&userID)

	if err == nil {
		_, err = conn.ExecContext(
			ctx,
			`INSERT INTO oauth_users (user_id, provider, provider_user_id, created_at)
			 VALUES (?, ?, ?, ?)`,
			userID,
			provider,
			providerUserID,
			nowUTC(),
		)
		if err != nil {
			return 0, fmt.Errorf("link oauth user: %w", err)
		}
		return userID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("email lookup failed: %w", err)
	}

	// Create new user
	username := createUsernameFromName(name)

	for i := 0; i < 3; i++ {
		result, err := conn.ExecContext(
			ctx,
			`INSERT INTO users 
			 (username, email, password_hash, is_active, created_at, updated_at)
			 VALUES (?, ?, ?, 1, ?, ?)`,
			username,
			email,
			"OAUTH",
			nowUTC(),
			nowUTC(),
		)

		if err == nil {
			userID, _ = result.LastInsertId()
			break
		}

		username = username + "-" + uuid.New().String()[:6]
	}

	if userID == 0 {
		return 0, fmt.Errorf("failed to create user after retries")
	}

	// Insert oauth link
	_, err = conn.ExecContext(
		ctx,
		`INSERT INTO oauth_users 
		 (user_id, provider, provider_user_id, created_at)
		 VALUES (?, ?, ?, ?)`,
		userID,
		provider,
		providerUserID,
		nowUTC(),
	)
	if err != nil {
		return 0, fmt.Errorf("insert oauth user: %w", err)
	}

	return userID, nil
}

func createUsernameFromName(name string) string {
	name = strings.TrimSpace(name)

	if name == "" {
		return "user-" + uuid.New().String()[:8]
	}

	name = strings.ReplaceAll(name, " ", "-")
	name = sanitizeUsername(name)

	if len(name) < 3 {
		return "user-" + uuid.New().String()[:8]
	}

	return name
}

func sanitizeUsername(s string) string {
	allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	var cleaned strings.Builder

	for _, ch := range s {
		if strings.ContainsRune(allowed, ch) {
			cleaned.WriteRune(ch)
		}
	}

	out := cleaned.String()
	if len(out) > 30 {
		out = out[:30]
	}
	return out
}

func nowUTC() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}
