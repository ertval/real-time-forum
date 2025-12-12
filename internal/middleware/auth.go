package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"forum/internal/db"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Auth ensures a valid session and injects userID into context.
// Supports BOTH:
//   - Cookie: session_token
//   - Header: Authorization: Bearer <token>
func Auth(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := ""

			// 1️⃣ Try cookie first
			if cookie, err := r.Cookie("session_token"); err == nil {
				token = cookie.Value
			}

			// 2️⃣ Fallback to Authorization header
			if token == "" {
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(auth, "Bearer ") {
					token = strings.TrimPrefix(auth, "Bearer ")
				}
			}

			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			session, err := db.GetSessionByToken(r.Context(), database, token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts authenticated user ID from context.
func GetUserID(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(UserIDKey).(int64)
	if !ok || id <= 0 {
		return 0, errors.New("unauthenticated")
	}
	return id, nil
}
