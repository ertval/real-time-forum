package handlers

import (
	"database/sql"
	"encoding/json"
	db "forum/internal/db"
	"net/http"
	"strconv"
	"strings"
)

type Users struct{ db *sql.DB }

func NewUsers(database *sql.DB) *Users {
	return &Users{db: database}
}

// -----------------------------
//
//	USER REGISTER
//
// -----------------------------
func (u *Users) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}

	var req db.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	id, err := db.CreateUser(r.Context(), u.db, req)
	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", err.Error(), http.StatusBadRequest))
		return
	}

	WriteCreated(w, map[string]any{
		"id":      id,
		"message": "user created",
	})
}

// -----------------------------
//
//	GET USER BY ID
//
// -----------------------------
func (u *Users) Item(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	if idStr == "" {
		WriteError(w, NewError("BAD_REQUEST", "missing user id", http.StatusBadRequest))
		return
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid user ID", http.StatusBadRequest))
		return
	}

	user, err := db.GetUser(r.Context(), u.db, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteError(w, NewError("NOT_FOUND", "user not found", http.StatusNotFound))
		} else {
			WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to get user", http.StatusInternalServerError))
		}
		return
	}

	WriteOK(w, user, nil)
}

// -----------------------------
//
//	LOGIN
//
// -----------------------------
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}

	var req db.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid login json", http.StatusBadRequest))
		return
	}

	user, err := db.LoginUser(r.Context(), u.db, req)
	if err != nil {
		WriteError(w, NewError("UNAUTHORIZED", "invalid username/email or password", http.StatusUnauthorized))
		return
	}

	session, err := db.CreateSession(r.Context(), u.db, user.ID, r.RemoteAddr, r.UserAgent())
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to create session", http.StatusInternalServerError))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	WriteOK(w, map[string]any{"message": "Login successful"}, nil)
}

// -----------------------------
//
//	/me — CURRENT USER
//
// -----------------------------
func (u *Users) Me(w http.ResponseWriter, r *http.Request) {
	// Read userID from context (middleware)
	val := r.Context().Value("userID")
	if val == nil {
		WriteError(w, NewError("UNAUTHORIZED", "login required", http.StatusUnauthorized))
		return
	}

	userID, ok := val.(int64)
	if !ok {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "invalid user context", http.StatusInternalServerError))
		return
	}

	user, err := db.GetUser(r.Context(), u.db, userID)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to get user", http.StatusInternalServerError))
		return
	}

	WriteOK(w, user, nil)
}

// -----------------------------
//
//	LOGOUT
//
// -----------------------------
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		WriteError(w, NewError("UNAUTHORIZED", "no session cookie", http.StatusUnauthorized))
		return
	}

	if err := db.InvalidateSessionByToken(r.Context(), u.db, cookie.Value); err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to logout", http.StatusInternalServerError))
		return
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	WriteOK(w, map[string]string{"message": "Logout successful"}, nil)
}
