package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	userTimeout       = 2 * time.Second
	minUsernameLength = 3
	maxUsernameLength = 30
	minPasswordLength = 8
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
// PUBLIC API
// ---------------------------------------------------------

func CreateUser(
	ctx context.Context,
	db *sql.DB,
	in CreateUserRequest,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	if err := validateCreateUser(in); err != nil {
		return 0, err
	}

	hash, err := hashPassword(in.Password)
	if err != nil {
		return 0, err
	}

	id, err := insertUser(ctx, db, in.Username, in.Email, hash)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func LoginUser(
	ctx context.Context,
	db *sql.DB,
	in LoginRequest,
) (User, error) {

	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	u, err := fetchUserForLogin(ctx, db, in)
	if err != nil {
		return User{}, err
	}

	if err := comparePassword(u.PasswordHash, in.Password); err != nil {
		return User{}, fmt.Errorf("invalid username/email or password")
	}

	return u, nil
}

func GetUser(
	ctx context.Context,
	db *sql.DB,
	id int64,
) (User, error) {

	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	var u User
	err := db.QueryRowContext(ctx,
		`SELECT id, username, email, is_active, created_at, updated_at
		 FROM users
		 WHERE id = ?`,
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

//
// ---------------------------------------------------------
// HELPERS
// ---------------------------------------------------------

func validateCreateUser(in CreateUserRequest) error {
	switch {
	case len(in.Username) < minUsernameLength:
		return fmt.Errorf("username must be at least %d characters long", minUsernameLength)
	case len(in.Username) > maxUsernameLength:
		return fmt.Errorf("username can't be longer than %d characters", maxUsernameLength)
	case len(in.Password) < minPasswordLength:
		return fmt.Errorf("password must be at least %d characters long", minPasswordLength)
	case !strings.Contains(in.Email, "@"):
		return fmt.Errorf("invalid email address")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}

func insertUser(
	ctx context.Context,
	db *sql.DB,
	username,
	email,
	passwordHash string,
) (int64, error) {

	res, err := db.ExecContext(ctx,
		`INSERT INTO users (username, email, password_hash)
		 VALUES (?, ?, ?)`,
		username,
		email,
		passwordHash,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			msg := err.Error()
			switch {
			case strings.Contains(msg, "username"):
				return 0, fmt.Errorf("username already exists")
			case strings.Contains(msg, "email"):
				return 0, fmt.Errorf("email already exists")
			}
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

func fetchUserForLogin(
	ctx context.Context,
	db *sql.DB,
	in LoginRequest,
) (User, error) {

	var row *sql.Row

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
	if err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.IsActive,
	); err != nil {

		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("invalid username/email or password")
		}
		return User{}, fmt.Errorf("fetch user: %w", err)
	}

	return u, nil
}
