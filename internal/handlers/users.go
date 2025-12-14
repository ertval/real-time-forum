package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	repository "forum/internal/db"
	"forum/internal/middleware"
)

type UsersHandler struct {
	conn *sql.DB
}

func NewUsersHandler(database *sql.DB) *UsersHandler {
	return &UsersHandler{conn: database}
}

// ============================================================
// COLLECTION / SINGLE ROUTES
// ============================================================

// HandleUser → GET /api/v1/users/{id} (auth required via router)
func (u *UsersHandler) HandleUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	userID, err := parseID(r.URL.Path, "/api/v1/users/")
	if err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid user ID", http.StatusBadRequest))
		return
	}

	user, err := repository.GetUser(r.Context(), u.conn, userID)
	if err != nil {
		WriteError(w, NewError("NOT_FOUND", "user not found", http.StatusNotFound))
		return
	}

	WriteOK(w, user, nil)
}

// Register → POST /api/v1/users/register
func (u *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
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

// ============================================================
// AUTH
// ============================================================

// Login → POST /api/v1/users/login
func (u *UsersHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req repository.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, NewError("BAD_REQUEST", "invalid login json", http.StatusBadRequest))
		return
	}

	user, err := repository.LoginUser(r.Context(), u.conn, req)
	if err != nil {
		WriteError(w, NewError("UNAUTHORIZED", "invalid credentials", http.StatusUnauthorized))
		return
	}

	session, err := repository.CreateSession(
		r.Context(),
		u.conn,
		user.ID,
		r.RemoteAddr,
		r.UserAgent(),
	)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to create session", http.StatusInternalServerError))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // important for tests
	})

	WriteOK(w, map[string]any{"message": "Login successful"}, nil)
}

// Me → GET /api/v1/users/me
func (u *UsersHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		WriteError(w, NewError("UNAUTHORIZED", "login required", http.StatusUnauthorized))
		return
	}

	user, err := repository.GetUser(r.Context(), u.conn, userID)
	if err != nil {
		WriteError(w, NewError("INTERNAL_SERVER_ERROR", "failed to get user", http.StatusInternalServerError))
		return
	}

	WriteOK(w, user, nil)
}

// Logout → POST /api/v1/users/logout
func (u *UsersHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		WriteError(w, NewError("UNAUTHORIZED", "no session", http.StatusUnauthorized))
		return
	}

	_ = repository.InvalidateSessionByToken(r.Context(), u.conn, cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	WriteOK(w, map[string]string{"message": "Logout successful"}, nil)
}
