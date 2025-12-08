package middleware

import (
	"context"
	"database/sql"
	"forum/internal/db"
	"forum/internal/handlers"
	"net/http"
)

func Auth(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				handlers.WriteError(w, handlers.NewError("UNAUTHORIZED", "no session cookie", http.StatusUnauthorized))
				return
			}
			session, err := db.GetSessionByToken(r.Context(), database, cookie.Value)
			if err != nil {
				handlers.WriteError(w, handlers.NewError("UNAUTHORIZED", "invalid session", http.StatusUnauthorized))
				return
			}
			ctx := context.WithValue(r.Context(), "userID", session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
