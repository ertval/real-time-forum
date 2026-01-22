// internal/middleware/middleware.go

package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// ------------------------------------------------------------
// LOGGER MIDDLEWARE
// Logs method, path and request duration.
// ------------------------------------------------------------
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%-6s %-40s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// ------------------------------------------------------------
// RECOVERER MIDDLEWARE
// Prevents server crash on panic and returns a safe JSON error.
// ------------------------------------------------------------
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if rec := recover(); rec != nil {

				// Log panic + stack trace (backend only)
				log.Printf("PANIC: %v\n%s", rec, debug.Stack())

				// Minimal, framework-agnostic response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(
					`{"error":{"code":"SERVER_ERROR","message":"internal server error"}}`,
				))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ------------------------------------------------------------
// CORS MIDDLEWARE
// Enables frontend ↔ backend communication on different ports.
// ------------------------------------------------------------
func EnableCORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
