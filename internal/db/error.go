package db

import (
	"errors"
	"fmt"
	"log"
)

// Common reusable database errors
var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicate     = errors.New("duplicate entry")
	ErrForeignKey    = errors.New("foreign key constraint failed")
	ErrInvalidInput  = errors.New("invalid input")
	ErrSchemaMissing = errors.New("database schema not applied")
	ErrConnection    = errors.New("database connection error")
)

// WrapError adds context to an existing error and logs it.
// Use this in DB operations to provide consistent error messages.
func WrapError(context string, err error) error {
	if err == nil {
		return nil
	}

	wrapped := fmt.Errorf("%s: %w", context, err)
	log.Printf("[DB ERROR] %s", wrapped.Error())
	return wrapped
}

// LogInfo provides a consistent log format for successful DB operations.
func LogInfo(message string, args ...any) {
	log.Printf("[DB INFO] "+message, args...)
}

// LogWarn provides a consistent log format for non-critical warnings.
func LogWarn(message string, args ...any) {
	log.Printf("[DB WARN] "+message, args...)
}

// MapSQLError attempts to classify raw SQLite errors into known types.
func MapSQLError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case containsAny(msg, "UNIQUE constraint failed"):
		return ErrDuplicate
	case containsAny(msg, "FOREIGN KEY constraint failed"):
		return ErrForeignKey
	case containsAny(msg, "no such table"):
		return ErrSchemaMissing
	default:
		return err
	}
}

// containsAny is a small helper used by MapSQLError.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if contains := stringContains(s, sub); contains {
			return true
		}
	}
	return false
}

// stringContains is a safe alias for strings.Contains (avoiding import cycles).
func stringContains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (len(s) == len(sub) && s == sub || len(s) > len(sub) && (indexOf(s, sub) >= 0))
}

// indexOf finds substring position or returns -1 (tiny helper to keep it standalone).
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// HandleInitError logs fatal startup errors (like DB init) and exits.
func HandleInitError(err error, context string) {
	if err != nil {
		log.Fatalf("[FATAL] %s: %v", context, err)
	}
}

// HandleRuntimeError logs recoverable runtime DB or logic errors.
func HandleRuntimeError(err error, context string) {
	if err != nil {
		log.Printf("[RUNTIME ERROR] %s: %v", context, err)
	}
}

// HandleFatalError logs unrecoverable errors (e.g. server crash) and exits.
func HandleFatalError(err error, context string) {
	if err != nil {
		log.Fatalf("[FATAL] %s: %v", context, err)
	}
}
