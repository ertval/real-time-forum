/* internal/handlers/oauth_google.go */

package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"forum/internal/db"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const googleAuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
const googleTokenURL = "https://oauth2.googleapis.com/token"
const googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

// ----------------------------------------
// STEP 1: Redirect user to Google login page
// GET /api/v1/auth/google
// ----------------------------------------
func (u *UsersHandler) GoogleStart(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		WriteError(w, r, NewError("SERVER_ERROR", "could not generate oauth state", 500))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes
	})

	params := url.Values{}
	params.Add("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	params.Add("redirect_uri", os.Getenv("GOOGLE_REDIRECT_URL"))
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)

	authURL := googleAuthURL + "?" + params.Encode()

	http.Redirect(w, r, authURL, http.StatusFound)
}

// ----------------------------------------
// STEP 2: Google redirects here with ?code=...
// GET /api/v1/auth/google/callback
// ----------------------------------------
func (u *UsersHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		WriteError(w, r, NewError("BAD_REQUEST", "missing oauth state", 400))
		return
	}

	if r.URL.Query().Get("state") != cookie.Value {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid oauth state", 400))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		WriteError(w, r, NewError("BAD_REQUEST", "missing code", 400))
		return
	}

	// Exchange code for token
	tokenResp, err := exchangeCodeForToken(code)
	if err != nil {
		log.Println("token exchange error:", err)
		WriteError(w, r, NewError("UNAUTHORIZED", "google token exchange failed", 401))
		return
	}

	// Get user info from Google
	googleUser, err := fetchGoogleUser(tokenResp.AccessToken)
	if err != nil {
		log.Println("userinfo error:", err)
		WriteError(w, r, NewError("UNAUTHORIZED", "google user info failed", 401))
		return
	}

	// Upsert user into DB
	userID, err := db.FindOrCreateOAuthUser(
		r.Context(),
		u.conn,
		"google",
		googleUser.Sub,
		googleUser.Email,
		googleUser.Name,
	)
	if err != nil {
		log.Println("upsert oauth error:", err)
		WriteError(w, r, NewError("SERVER_ERROR", "failed to create oauth user", 500))
		return
	}

	// 5️⃣ Create session
	session, err := db.CreateSession(
		r.Context(),
		u.conn,
		userID,
		r.RemoteAddr,
		r.UserAgent(),
	)
	if err != nil {
		log.Println("session create error:", err)
		WriteError(w, r, NewError("SERVER_ERROR", "failed to create session", 500))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "http://localhost:3000/", http.StatusFound)
}

// ---------------------------------------------------------------------
// HELPERS
// ---------------------------------------------------------------------

func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return base64.URLEncoding.EncodeToString(b), err
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func exchangeCodeForToken(code string) (*googleTokenResponse, error) {
	values := url.Values{}
	values.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	values.Set("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	values.Set("redirect_uri", os.Getenv("GOOGLE_REDIRECT_URL"))
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)

	req, _ := http.NewRequest("POST", googleTokenURL, bytes.NewBufferString(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func fetchGoogleUser(accessToken string) (*googleUserInfo, error) {
	req, _ := http.NewRequest("GET", googleUserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}
