package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"forum/internal/auth"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const googleAuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
const googleTokenURL = "https://oauth2.googleapis.com/token"
const googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

/* ---------------- GOOGLE START ---------------- */

func (u *UsersHandler) GoogleStart(w http.ResponseWriter, r *http.Request) {

	state, err := generateState()
	if err != nil {
		log.Printf("[OAUTH][GOOGLE] state generation failed: %v", err)
		WriteError(w, r, NewError("SERVER_ERROR", "authentication unavailable", http.StatusInternalServerError))
		return
	}

	setOAuthStateCookie(w, state)

	params := url.Values{}
	params.Add("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	params.Add("redirect_uri", os.Getenv("GOOGLE_REDIRECT_URL"))
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)

	http.Redirect(w, r, googleAuthURL+"?"+params.Encode(), http.StatusFound)
}

/* ---------------- GOOGLE CALLBACK ---------------- */

func (u *UsersHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {

	if err := validateOAuthState(r); err != nil {
		log.Printf("[OAUTH][GOOGLE] invalid state: %v", err)
		WriteError(w, r, NewError("BAD_REQUEST", "invalid oauth state", http.StatusBadRequest))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		WriteError(w, r, NewError("BAD_REQUEST", "missing code", http.StatusBadRequest))
		return
	}

	tokenResp, err := exchangeGoogleCode(code)
	if err != nil {
		log.Printf("[OAUTH][GOOGLE] token exchange failed: %v", err)
		WriteError(w, r, NewError("UNAUTHORIZED", "google authentication failed", http.StatusUnauthorized))
		return
	}

	user, err := fetchGoogleUser(tokenResp.AccessToken)
	if err != nil {
		log.Printf("[OAUTH][GOOGLE] user fetch failed: %v", err)
		WriteError(w, r, NewError("UNAUTHORIZED", "google authentication failed", http.StatusUnauthorized))
		return
	}

	userID, err := auth.FindOrCreateOAuthUser(
		r.Context(),
		u.conn,
		"google",
		user.Sub,
		user.Email,
		user.Name,
	)
	if err != nil {
		log.Printf("[OAUTH][GOOGLE] upsert failed: %v", err)
		WriteError(w, r, NewError("SERVER_ERROR", "authentication failed", http.StatusInternalServerError))
		return
	}

	createSessionAndRedirect(w, r, u.conn, userID)
}

/* ---------------- TOKEN EXCHANGE ---------------- */

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func exchangeGoogleCode(code string) (*googleTokenResponse, error) {

	values := url.Values{}
	values.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	values.Set("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	values.Set("redirect_uri", os.Getenv("GOOGLE_REDIRECT_URL"))
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)

	req, _ := http.NewRequest("POST", googleTokenURL, bytes.NewBufferString(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "forum-app")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google token exchange failed: %s", string(body))
	}

	var token googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

/* ---------------- FETCH USER ---------------- */

type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func fetchGoogleUser(accessToken string) (*googleUserInfo, error) {

	req, _ := http.NewRequest("GET", googleUserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "forum-app")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google user API failed: %s", string(body))
	}

	var user googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
