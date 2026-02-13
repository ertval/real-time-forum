/* internal/db/oauth.go */

package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

/*
FindOrCreateOAuthUser

Steps:
 1. Check if oauth_users entry exists for (provider, provider_user_id)
 2. If exists → return linked user_id
 3. If not exists:
    a) Create user in USERS table
    b) Create entry in OAUTH_USERS table
    c) Return new user_id
*/
func FindOrCreateOAuthUser(
	ctx context.Context,
	conn *sql.DB,
	provider string,
	providerUserID string,
	email string,
	name string,
) (int64, error) {

	// -----------------------------------------
	// STEP 1: Check if OAuth user already exists
	// -----------------------------------------
	var existingUserID int64
	err := conn.QueryRowContext(
		ctx,
		`SELECT user_id FROM oauth_users WHERE provider = ? AND provider_user_id = ? LIMIT 1`,
		provider,
		providerUserID,
	).Scan(&existingUserID)

	if err == nil {
		// User already exists → return user_id
		return existingUserID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to lookup oauth user: %w", err)
	}

	// -----------------------------------------
	// STEP 2: Create a new user since it doesn't exist
	// -----------------------------------------

	// Generate fallback username (required by schema)
	username := createUsernameFromName(name)

	// Password_hash required by schema → we store ""
	// OAuth users DO NOT use password login
	result, err := conn.ExecContext(
		ctx,
		`INSERT INTO users (username, email, password_hash, is_active, created_at, updated_at)
         VALUES (?, ?, ?, 1, ?, ?)`,
		username,
		email,
		"", // empty password for OAuth users
		nowUTC(),
		nowUTC(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch new user id: %w", err)
	}

	// -----------------------------------------
	// STEP 3: Insert into oauth_users
	// -----------------------------------------
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
		log.Println("oauth insert error:", err)
		return 0, fmt.Errorf("failed to insert oauth user: %w", err)
	}

	return userID, nil
}

//
// HELPERS
//

// createUsernameFromName generates a sanitized username or fallback uuid.
func createUsernameFromName(name string) string {
	name = strings.TrimSpace(name)

	if name == "" {
		return "user-" + uuid.New().String()
	}

	// Convert spaces → dashes
	name = strings.ReplaceAll(name, " ", "-")

	// Remove characters not allowed by your username rules
	name = sanitizeUsername(name)

	if len(name) < 3 {
		return "user-" + uuid.New().String()
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

	// Username length limit (your schema requires 3–30 chars)
	if len(out) > 30 {
		out = out[:30]
	}
	return out
}

func nowUTC() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}
