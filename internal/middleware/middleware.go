package middleware

import (
	"forum/internal/handlers"
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

		// Nicely formatted aligned output
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

				// Log panic + stack trace (only in backend logs)
				log.Printf("PANIC: %v\n%s", rec, debug.Stack())

				// Send standard JSON error envelope
				handlers.WriteError(
					w,
					handlers.NewError(
						"SERVER_ERROR",
						"internal server error",
						http.StatusInternalServerError,
					),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ------------------------------------------------------------
// CORS MIDDLEWARE
// Enables frontend ↔ backend communication on different ports.
// ------------------------------------------------------------
func CORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Allow frontend domain
			w.Header().Set("Access-Control-Allow-Origin", origin)

			// Required to send cookies (sessions)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Which headers are accepted
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Which HTTP methods are allowed
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

			// Handle preflight request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
