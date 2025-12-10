package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

//
// ---------------------------------------------------------
// CREATE USER
// ---------------------------------------------------------
//

func CreateUser(ctx context.Context, db *sql.DB, in CreateUserRequest) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Basic validation
	switch {
	case len(in.Username) < 3:
		return 0, fmt.Errorf("username must be at least 3 characters long")
	case len(in.Username) > 30:
		return 0, fmt.Errorf("username can't be longer than 30 characters")
	case len(in.Password) < 8:
		return 0, fmt.Errorf("password must be at least 8 characters long")
	case !strings.Contains(in.Email, "@"):
		return 0, fmt.Errorf("invalid email address")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	stmt, err := db.PrepareContext(ctx,
		`INSERT INTO users (username, email, password_hash)
         VALUES (?, ?, ?)`,
	)
	if err != nil {
		return 0, fmt.Errorf("prepare insert user: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, in.Username, in.Email, string(hash))
	if err != nil {
		// Unique constraint violations
		msg := err.Error()
		switch {
		case strings.Contains(msg, "UNIQUE") && strings.Contains(msg, "username"):
			return 0, fmt.Errorf("username already exists")
		case strings.Contains(msg, "UNIQUE") && strings.Contains(msg, "email"):
			return 0, fmt.Errorf("email already exists")
		default:
			return 0, fmt.Errorf("insert user: %w", err)
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

//
// ---------------------------------------------------------
// LOGIN USER
// ---------------------------------------------------------
//

func LoginUser(ctx context.Context, db *sql.DB, in LoginRequest) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var row *sql.Row

	// Prefer username when both given
	if strings.TrimSpace(in.Username) != "" {
		row = db.QueryRowContext(ctx,
			`SELECT id, username, email, password_hash, is_active
             FROM users
             WHERE username = ?`,
			in.Username,
		)
	} else {
		row = db.QueryRowContext(ctx,
			`SELECT id, username, email, password_hash, is_active
             FROM users
             WHERE email = ?`,
			in.Email,
		)
	}

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("invalid username/email or password")
		}
		return User{}, fmt.Errorf("login user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return User{}, fmt.Errorf("invalid username/email or password")
	}

	return u, nil
}

//
// ---------------------------------------------------------
// GET USER BY ID
// ---------------------------------------------------------
//

func GetUser(ctx context.Context, db *sql.DB, id int64) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var u User
	err := db.QueryRowContext(ctx,
		`SELECT id, username, email, is_active, created_at, updated_at
         FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}

	return u, nil
}
