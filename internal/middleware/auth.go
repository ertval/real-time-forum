package middleware

import (
	"context"
	"database/sql"
	"forum/internal/handlers"
	"net/http"
	"strings"

	db "forum/internal/db"
)

func Auth(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			session, err := db.GetSessionByToken(r.Context(), database, token)
			if err != nil {
				handlers.WriteError(w, handlers.NewError("UNAUTHORIZED", "invalid token", http.StatusUnauthorized))
				return
			}
			ctx := context.WithValue(r.Context(), "userID", session.UserID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
