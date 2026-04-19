package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIProxyPreservesPathAndCookie(t *testing.T) {
	originalBackendURL := backendBaseURL
	defer func() {
		backendBaseURL = originalBackendURL
	}()

	var receivedPath string
	var receivedCookie string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedCookie = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("proxied"))
	}))
	defer backend.Close()

	backendBaseURL = backend.URL
	mux := NewMux()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Cookie", "session=abc123")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, res.Code)
	}

	if receivedPath != "/api/v1/users/me" {
		t.Fatalf("expected backend path /api/v1/users/me, got %q", receivedPath)
	}

	if !strings.Contains(receivedCookie, "session=abc123") {
		t.Fatalf("expected session cookie to be proxied, got %q", receivedCookie)
	}
}
