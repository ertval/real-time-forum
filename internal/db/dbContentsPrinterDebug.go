package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// NOTE:
// This file is a DEBUG / DEVELOPMENT utility ONLY.
// It is NOT used by handlers, middleware, or production flow.

// PrintDBContents prints basic database contents for debugging purposes.
// It performs READ-ONLY queries and must not contain business logic.
func PrintDBContents(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := printUsers(ctx, db); err != nil {
		return err
	}
	if err := printPosts(ctx, db); err != nil {
		return err
	}
	if err := printCategories(ctx, db); err != nil {
		return err
	}
	if err := printComments(ctx, db); err != nil {
		return err
	}
	if err := printReactions(ctx, db); err != nil {
		return err
	}
	if err := printSessions(ctx, db); err != nil {
		return err
	}

	return nil
}

/*--------
  USERS
--------*/

func printUsers(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx,
		`SELECT id, username, email, is_active, created_at FROM users`,
	)
	if err != nil {
		return fmt.Errorf("print users: %w", err)
	}
	defer rows.Close()

	fmt.Println("---- USERS ----")
	for rows.Next() {
		var id int64
		var username, email, createdAt string
		var isActive bool

		if err := rows.Scan(&id, &username, &email, &isActive, &createdAt); err != nil {
			return err
		}

		fmt.Printf("ID=%d USERNAME=%s EMAIL=%s ACTIVE=%v CREATED=%s\n",
			id, username, email, isActive, createdAt)
	}

	return rows.Err()
}

/*--------
  POSTS
--------*/

func printPosts(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx,
		`SELECT id, author_id, title, created_at FROM posts`,
	)
	if err != nil {
		return fmt.Errorf("print posts: %w", err)
	}
	defer rows.Close()

	fmt.Println("---- POSTS ----")
	for rows.Next() {
		var id, authorID int64
		var title, createdAt string

		if err := rows.Scan(&id, &authorID, &title, &createdAt); err != nil {
			return err
		}

		fmt.Printf("ID=%d AUTHOR=%d TITLE=%q CREATED=%s\n",
			id, authorID, title, createdAt)
	}

	return rows.Err()
}

/*------------
  CATEGORIES
------------*/

func printCategories(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, slug FROM categories`,
	)
	if err != nil {
		return fmt.Errorf("print categories: %w", err)
	}
	defer rows.Close()

	fmt.Println("---- CATEGORIES ----")
	for rows.Next() {
		var id int64
		var name, slug string

		if err := rows.Scan(&id, &name, &slug); err != nil {
			return err
		}

		fmt.Printf("ID=%d NAME=%s SLUG=%s\n", id, name, slug)
	}

	return rows.Err()
}

/*----------
  COMMENTS
----------*/

func printComments(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx,
		`SELECT id, post_id, user_id, body, created_at FROM comments`,
	)
	if err != nil {
		return fmt.Errorf("print comments: %w", err)
	}
	defer rows.Close()

	fmt.Println("---- COMMENTS ----")
	for rows.Next() {
		var id, postID, userID int64
		var body, createdAt string

		if err := rows.Scan(&id, &postID, &userID, &body, &createdAt); err != nil {
			return err
		}

		fmt.Printf("ID=%d POST=%d USER=%d BODY=%q CREATED=%s\n",
			id, postID, userID, body, createdAt)
	}

	return rows.Err()
}

/*-----------
  REACTIONS
-----------*/

func printReactions(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx,
		`SELECT user_id, post_id, value FROM reactions`,
	)
	if err != nil {
		return fmt.Errorf("print reactions: %w", err)
	}
	defer rows.Close()

	fmt.Println("---- REACTIONS ----")
	for rows.Next() {
		var userID, postID int64
		var value int

		if err := rows.Scan(&userID, &postID, &value); err != nil {
			return err
		}

		fmt.Printf("USER=%d POST=%d VALUE=%d\n",
			userID, postID, value)
	}

	return rows.Err()
}

/*----------
  SESSIONS
----------*/

func printSessions(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx,
		`SELECT id, user_id, token, is_valid, expires_at FROM sessions`,
	)
	if err != nil {
		return fmt.Errorf("print sessions: %w", err)
	}
	defer rows.Close()

	fmt.Println("---- SESSIONS ----")
	for rows.Next() {
		var id, userID int64
		var token, expiresAt string
		var isValid int

		if err := rows.Scan(&id, &userID, &token, &isValid, &expiresAt); err != nil {
			return err
		}

		fmt.Printf("ID=%d USER=%d VALID=%d EXPIRES=%s TOKEN=%s\n",
			id, userID, isValid, expiresAt, token)
	}

	return rows.Err()
}
