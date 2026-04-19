// Internal/db/users.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	userTimeout       = 2 * time.Second
	minUsernameLength = 3
	maxUsernameLength = 30
	minPasswordLength = 8
	maxPasswordLength = 64
)

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Age          int    `json:"age"`
	Gender       string `json:"gender"`
	PasswordHash string `json:"-"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type CreateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

/*------------
  PUBLIC API
------------*/

func CreateUser(
	ctx context.Context,
	db *sql.DB,
	req CreateUserRequest,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	if err := validateCreateUser(req); err != nil {
		return 0, err
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return 0, err
	}

	id, err := insertUser(
		ctx,
		db,
		req.Username,
		req.Email,
		hash,
		req.FirstName,
		req.LastName,
		req.Age,
		req.Gender,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func LoginUser(
	ctx context.Context,
	db *sql.DB,
	req LoginRequest,
) (User, error) {

	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	user, err := fetchUserForLogin(ctx, db, req)
	if err != nil {
		return User{}, err
	}

	if err := comparePassword(user.PasswordHash, req.Password); err != nil {
		return User{}, ErrInvalidCredentials
	}

	return user, nil
}

func GetUser(
	ctx context.Context,
	db *sql.DB,
	id int64,
) (User, error) {

	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	var user User
	err := db.QueryRowContext(ctx,
		`SELECT id, username, email, first_name, last_name, age, gender, is_active, created_at, updated_at
		 FROM users
		 WHERE id = ?`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Age,
		&user.Gender,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}
