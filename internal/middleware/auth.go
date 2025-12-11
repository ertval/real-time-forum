package middleware

import (
	"context"
	"database/sql"
	"forum/internal/db"
	"forum/internal/handlers"
	"net/http"
)

// Context key for authenticated user ID (must match handlers).
const userIDKey = "userID"

// Auth middleware ensures that a valid session cookie is present.
// If valid, it attaches the user ID to the request context.
func Auth(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Read session_token cookie
			cookie, err := r.Cookie("session_token")
			if err != nil {
				handlers.WriteError(
					w,
					handlers.NewError("UNAUTHORIZED", "missing session cookie", http.StatusUnauthorized),
				)
				return
			}

			// Validate token in the database
			session, err := db.GetSessionByToken(r.Context(), database, cookie.Value)
			if err != nil {
				handlers.WriteError(
					w,
					handlers.NewError("UNAUTHORIZED", "invalid or expired session", http.StatusUnauthorized),
				)
				return
			}

			// Store the authenticated user's ID in the request context
			ctxWithUser := context.WithValue(r.Context(), userIDKey, session.UserID)

			// Continue request with updated context
			next.ServeHTTP(w, r.WithContext(ctxWithUser))
		})
	}
}
