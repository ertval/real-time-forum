package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"forum/internal/db"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Auth ensures a valid session and injects userID into context.
func Auth(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := ""

			// Cookie first
			if cookie, err := r.Cookie("session_token"); err == nil {
				token = cookie.Value
			}

			// Authorization header fallback
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
				log.Printf("failed to get session by token: %v", err)
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
