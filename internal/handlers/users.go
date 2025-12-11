package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	repository "forum/internal/db"
)

type UsersHandler struct {
	conn *sql.DB
}

func NewUsersHandler(database *sql.DB) *UsersHandler {
	return &UsersHandler{conn: database}
}

//
// ─────────────────────────────────────────────────────────────
//  USER REGISTER
// ─────────────────────────────────────────────────────────────
//

func (u *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}

	var req repository.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid json", http.StatusBadRequest))
		return
	}

	id, err := repository.CreateUser(r.Context(), u.conn, req)
	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", err.Error(), http.StatusBadRequest))
		return
	}

	WriteCreated(w, map[string]any{
		"id":      id,
		"message": "user created",
	})
}

//
// ─────────────────────────────────────────────────────────────
//  GET USER BY ID
// ─────────────────────────────────────────────────────────────
//

func (u *UsersHandler) Item(w http.ResponseWriter, r *http.Request) {
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

	user, err := repository.GetUser(r.Context(), u.conn, userID)
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

//
// ─────────────────────────────────────────────────────────────
//  LOGIN
// ─────────────────────────────────────────────────────────────
//

func (u *UsersHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}

	var req repository.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid login json", http.StatusBadRequest))
		return
	}

	user, err := repository.LoginUser(r.Context(), u.conn, req)
	if err != nil {
		WriteError(w, NewError("UNAUTHORIZED", "invalid username/email or password", http.StatusUnauthorized))
		return
	}

	session, err := repository.CreateSession(r.Context(), u.conn, user.ID, r.RemoteAddr, r.UserAgent())
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

//
// ─────────────────────────────────────────────────────────────
//  /me — CURRENT USER
// ─────────────────────────────────────────────────────────────
//

func (u *UsersHandler) Me(w http.ResponseWriter, r *http.Request) {
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

	user, err := repository.GetUser(r.Context(), u.conn, userID)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to get user", http.StatusInternalServerError))
		return
	}

	WriteOK(w, user, nil)
}

//
// ─────────────────────────────────────────────────────────────
//  LOGOUT
// ─────────────────────────────────────────────────────────────
//

func (u *UsersHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		WriteError(w, NewError("UNAUTHORIZED", "no session cookie", http.StatusUnauthorized))
		return
	}

	if err := repository.InvalidateSessionByToken(r.Context(), u.conn, cookie.Value); err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to logout", http.StatusInternalServerError))
		return
	}

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
