// internal/handlers/users.go
package handlers

import (
	"encoding/json"
	"forum/internal/middleware"
	"log"
	"net/http"

	repository "forum/internal/db"
)

/*--------------------------------
  GET /api/v1/users/{id}
--------------------------------*/

func (u *UsersHandler) HandleUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w, r)
		return
	}

	userID, ok := resolveUserID(w, r)
	if !ok {
		return
	}

	user, err := repository.GetUser(r.Context(), u.conn, userID)
	if err != nil {
		log.Printf("[USERS][GET] failed: %v", err)
		WriteError(w, r, NewError(
			"NOT_FOUND",
			"user not found",
			http.StatusNotFound,
		))
		return
	}

	WriteOK(w, user, nil)
}

/*-----------------------
  GET /api/v1/users/me
-----------------------*/

func (u *UsersHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		WriteError(w, r, NewError(
			"UNAUTHORIZED",
			"login required",
			http.StatusUnauthorized,
		))
		return
	}

	user, err := repository.GetUser(r.Context(), u.conn, userID)
	if err != nil {
		log.Printf("[USERS][ME] failed: %v", err)
		writeHandlerError(w, r, err, "failed to load user")
		return
	}

	WriteOK(w, user, nil)
}

/*------------------------------
  POST /api/v1/users/register
------------------------------*/

func (u *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		MethodNotAllowed(w, r)
		return
	}

	var req repository.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, NewError(
			"BAD_REQUEST",
			"invalid json",
			http.StatusBadRequest,
		))
		return
	}

	userID, err := repository.CreateUser(r.Context(), u.conn, req)
	if err != nil {
		log.Printf("[USERS][REGISTER] failed: %v", err)
		WriteError(w, r, NewError(
			"BAD_REQUEST",
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	session, err := repository.CreateSession(
		r.Context(),
		u.conn,
		userID,
		r.RemoteAddr,
		r.UserAgent(),
	)
	if err != nil {
		log.Printf("[AUTH][REGISTER] session creation failed: %v", err)
		writeHandlerError(w, r, err, "failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	WriteCreated(w, map[string]any{
		"id":      userID,
		"message": "user registered and logged in",
	})
}

/*---------------------------
  POST /api/v1/users/login
---------------------------*/

func (u *UsersHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		MethodNotAllowed(w, r)
		return
	}

	var req repository.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, NewError(
			"BAD_REQUEST",
			"invalid login json",
			http.StatusBadRequest,
		))
		return
	}

	user, err := repository.LoginUser(r.Context(), u.conn, req)
	if err != nil {
		log.Printf("[AUTH][LOGIN] failed: %v", err)
		WriteError(w, r, NewError(
			"UNAUTHORIZED",
			"invalid credentials",
			http.StatusUnauthorized,
		))
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
		log.Printf("[AUTH][LOGIN] session creation failed: %v", err)
		writeHandlerError(w, r, err, "failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	WriteOK(w, map[string]any{
		"message": "Login successful",
	}, nil)
}

/*---------------------------
  POST /api/v1/users/logout
---------------------------*/

func (u *UsersHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		MethodNotAllowed(w, r)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		WriteError(w, r, NewError(
			"UNAUTHORIZED",
			"no active session",
			http.StatusUnauthorized,
		))
		return
	}

	if err := repository.InvalidateSessionByToken(
		r.Context(),
		u.conn,
		cookie.Value,
	); err != nil {
		log.Printf("[AUTH][LOGOUT] session invalidation failed: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	WriteOK(w, map[string]string{
		"message": "logout successful",
	}, nil)
}
