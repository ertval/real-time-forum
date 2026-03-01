// internal/handlers/oauth_helpers.go
package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"forum/internal/db"
)

/*
generateState creates secure random state
*/
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func setOAuthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   300,
	})
}

func validateOAuthState(r *http.Request) error {
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		return fmt.Errorf("missing oauth state cookie")
	}

	state := r.URL.Query().Get("state")
	if state == "" || state != cookie.Value {
		return fmt.Errorf("invalid oauth state")
	}

	return nil
}

func createSessionAndRedirect(
	w http.ResponseWriter,
	r *http.Request,
	conn *sql.DB,
	userID int64,
) {
	session, err := db.CreateSession(
		r.Context(),
		conn,
		userID,
		r.RemoteAddr,
		r.UserAgent(),
	)
	if err != nil {
		WriteError(w, r, NewError("SERVER_ERROR", "failed to create session", 500))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Expires:  session.ExpiresAt,
	})

	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		frontend = "http://localhost:3000"
	}

	http.Redirect(w, r, frontend+"/", http.StatusFound)
}
