package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSPARouting(t *testing.T) {
	// Setup test directories and files in a temporary workspace
	tmpDir, err := os.MkdirTemp("", "frontend-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create real structure expected by NewMux
	spaDir := filepath.Join(tmpDir, "web/SPA")
	staticDir := filepath.Join(tmpDir, "web/static")
	errorDir := filepath.Join(tmpDir, "web/errors")
	os.MkdirAll(spaDir, 0755)
	os.MkdirAll(staticDir, 0755)
	os.MkdirAll(errorDir, 0755)

	indexContent := "<html>SPA Shell</html>"
	os.WriteFile(filepath.Join(spaDir, "index.html"), []byte(indexContent), 0644)
	jsContent := "console.log('hi');"
	os.WriteFile(filepath.Join(spaDir, "main.js"), []byte(jsContent), 0644)
	faviconContent := "fake-favicon"
	os.WriteFile(filepath.Join(staticDir, "favicon.ico"), []byte(faviconContent), 0644)

	// Mock backend for proxy testing
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer backend.Close()

	// Redirect NewMux to use our mock backend
	oldBackendURL := backendBaseURL
	backendBaseURL = backend.URL
	defer func() { backendBaseURL = oldBackendURL }()

	// Change working directory so NewMux can find ./web/...
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)

	// Now use the real Mux!
	mux := NewMux()

	tests := []struct {
		name                string
		path                string
		accept              string
		expectedStatus      int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "Root",
			path:                "/",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/html; charset=utf-8",
			expectedBody:        indexContent,
		},
		{
			name:                "Existing JS",
			path:                "/main.js",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/javascript; charset=utf-8",
			expectedBody:        jsContent,
		},
		{
			name:                "SPA Deep Link",
			path:                "/login",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/html; charset=utf-8",
			expectedBody:        indexContent,
		},
		{
			name:                "SPA Deep Link with HTML accept",
			path:                "/posts/123",
			accept:              "text/html,application/xhtml+xml",
			expectedStatus:      http.StatusOK,
			expectedContentType: "text/html; charset=utf-8",
			expectedBody:        indexContent,
		},
		{
			name:                "Missing Asset (P2 fix)",
			path:                "/assets/missing.css",
			expectedStatus:      http.StatusNotFound,
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "API Proxy (Proxy gate check)",
			path:                "/api/v1/posts",
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/json",
			expectedBody:        `{"status":"ok"}`,
		},
		{
			name:                "Favicon",
			path:                "/favicon.ico",
			expectedStatus:      http.StatusOK,
			expectedContentType: "icon",
			expectedBody:        faviconContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			actualCT := rr.Header().Get("Content-Type")
			if tt.expectedContentType != "" && !strings.Contains(actualCT, tt.expectedContentType) {
				t.Errorf("expected Content-Type to contain %q, got %q", tt.expectedContentType, actualCT)
			}

			if tt.expectedBody != "" && rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
