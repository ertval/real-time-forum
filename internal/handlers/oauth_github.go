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

const githubAuthURL = "https://github.com/login/oauth/authorize"
const githubTokenURL = "https://github.com/login/oauth/access_token"
const githubUserURL = "https://api.github.com/user"
const githubEmailsURL = "https://api.github.com/user/emails"

/* ---------------- GITHUB START ---------------- */

func (u *UsersHandler) GithubStart(w http.ResponseWriter, r *http.Request) {

	state, err := generateState()
	if err != nil {
		log.Printf("[OAUTH][GITHUB] state generation failed: %v", err)
		WriteError(w, r, NewError("SERVER_ERROR", "authentication unavailable", http.StatusInternalServerError))
		return
	}

	setOAuthStateCookie(w, state)

	params := url.Values{}
	params.Add("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	params.Add("redirect_uri", os.Getenv("GITHUB_REDIRECT_URL"))
	params.Add("scope", "user:email")
	params.Add("state", state)

	http.Redirect(w, r, githubAuthURL+"?"+params.Encode(), http.StatusFound)
}

/* ---------------- GITHUB CALLBACK ---------------- */

func (u *UsersHandler) GithubCallback(w http.ResponseWriter, r *http.Request) {

	if err := validateOAuthState(r); err != nil {
		log.Printf("[OAUTH][GITHUB] invalid state: %v", err)
		WriteError(w, r, NewError("BAD_REQUEST", "invalid oauth state", http.StatusBadRequest))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		WriteError(w, r, NewError("BAD_REQUEST", "missing code", http.StatusBadRequest))
		return
	}

	token, err := exchangeGitHubCode(code)
	if err != nil {
		log.Printf("[OAUTH][GITHUB] token exchange failed: %v", err)
		WriteError(w, r, NewError("UNAUTHORIZED", "github authentication failed", http.StatusUnauthorized))
		return
	}

	user, err := fetchGitHubUser(token)
	if err != nil {
		log.Printf("[OAUTH][GITHUB] user fetch failed: %v", err)
		WriteError(w, r, NewError("UNAUTHORIZED", "github authentication failed", http.StatusUnauthorized))
		return
	}

	userID, err := auth.FindOrCreateOAuthUser(
		r.Context(),
		u.conn,
		"github",
		fmt.Sprintf("%d", user.ID),
		user.Email,
		user.Login,
	)
	if err != nil {
		log.Printf("[OAUTH][GITHUB] upsert failed: %v", err)
		WriteError(w, r, NewError("SERVER_ERROR", "authentication failed", http.StatusInternalServerError))
		return
	}

	createSessionAndRedirect(w, r, u.conn, userID)
}

/* ---------------- TOKEN EXCHANGE ---------------- */

func exchangeGitHubCode(code string) (string, error) {

	values := url.Values{}
	values.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	values.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	values.Set("code", code)
	values.Set("redirect_uri", os.Getenv("GITHUB_REDIRECT_URL"))

	req, _ := http.NewRequest("POST", githubTokenURL, bytes.NewBufferString(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "forum-app")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed: %s", string(body))
	}

	var token struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

/* ---------------- FETCH USER ---------------- */

type githubUserInfo struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

func fetchGitHubUser(accessToken string) (*githubUserInfo, error) {

	req, _ := http.NewRequest("GET", githubUserURL, nil)
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
		return nil, fmt.Errorf("github user API failed: %s", string(body))
	}

	var user githubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	if user.Email == "" {
		email, err := fetchPrimaryGitHubEmail(accessToken)
		if err != nil {
			return nil, err
		}
		user.Email = email
	}

	return &user, nil
}

/* ---------------- FETCH EMAIL ---------------- */

func fetchPrimaryGitHubEmail(accessToken string) (string, error) {

	req, _ := http.NewRequest("GET", githubEmailsURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "forum-app")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github email API failed: %s", string(body))
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no primary email found")
}
