package db

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

// ------------------------------------------------------------
// COMMON DATABASE ERROR SENTINELS
// ------------------------------------------------------------

var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicate     = errors.New("duplicate entry")
	ErrForeignKey    = errors.New("foreign key constraint failed")
	ErrInvalidInput  = errors.New("invalid input")
	ErrSchemaMissing = errors.New("database schema not applied")
	ErrConnection    = errors.New("database connection error")
)

// ------------------------------------------------------------
// ERROR WRAPPING + LOGGING
// ------------------------------------------------------------

// WrapError attaches context and logs it.
// Example: return WrapError("create user", err)
func WrapError(context string, err error) error {
	if err == nil {
		return nil
	}
	wrapped := fmt.Errorf("%s: %w", context, err)
	log.Printf("[DB ERROR] %s", wrapped)
	return wrapped
}

// ------------------------------------------------------------
// GENERIC LOG HELPERS
// ------------------------------------------------------------

func LogInfo(format string, args ...any) {
	log.Printf("[DB INFO] "+format, args...)
}

func LogWarn(format string, args ...any) {
	log.Printf("[DB WARN] "+format, args...)
}

// ------------------------------------------------------------
// SQLITE ERROR CLASSIFICATION
// ------------------------------------------------------------

// MapSQLError converts raw SQLite errors into friendly Go errors.
func MapSQLError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "UNIQUE constraint failed"):
		return ErrDuplicate
	case strings.Contains(msg, "FOREIGN KEY constraint failed"):
		return ErrForeignKey
	case strings.Contains(msg, "no such table"):
		return ErrSchemaMissing
	default:
		return err
	}
}

// ------------------------------------------------------------
// PROCESS LIFECYCLE ERROR HELPERS
// ------------------------------------------------------------

func HandleInitError(err error, context string) {
	if err != nil {
		log.Fatalf("[FATAL] %s: %v", context, err)
	}
}

func HandleRuntimeError(err error, context string) {
	if err != nil {
		log.Printf("[RUNTIME ERROR] %s: %v", context, err)
	}
}

func HandleFatalError(err error, context string) {
	if err != nil {
		log.Fatalf("[FATAL] %s: %v", context, err)
	}
}
