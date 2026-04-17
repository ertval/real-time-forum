package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSPARouting(t *testing.T) {
	// Setup test directories
	tmpDir, err := os.MkdirTemp("", "spa-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	spaDir := filepath.Join(tmpDir, "web/SPA")
	staticDir := filepath.Join(tmpDir, "web/static")
	os.MkdirAll(spaDir, 0755)
	os.MkdirAll(staticDir, 0755)

	indexContent := "<html>SPA Shell</html>"
	os.WriteFile(filepath.Join(spaDir, "index.html"), []byte(indexContent), 0644)
	jsContent := "console.log('hi');"
	os.WriteFile(filepath.Join(spaDir, "main.js"), []byte(jsContent), 0644)
	faviconContent := "fake-favicon"
	os.WriteFile(filepath.Join(staticDir, "favicon.ico"), []byte(faviconContent), 0644)

	// Override paths for testing (if possible, but let's just use the real structure)
	// For this test, I'll use a mocked mux that points to the temp dir.

	mux := http.NewServeMux()

	// Static
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "favicon.ico"))
	})

	// SPA Fallback
	spaFileServer := NewCustomFileServer(http.Dir(spaDir), filepath.Join(spaDir, "index.html"))
	mux.Handle("/", spaFileServer)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"Root", "/", http.StatusOK, indexContent},
		{"Existing JS", "/main.js", http.StatusOK, jsContent},
		{"Login route", "/login", http.StatusOK, indexContent},
		{"Register route", "/register", http.StatusOK, indexContent},
		{"Post detail route", "/post/1", http.StatusOK, indexContent},
		{"Create post route", "/create-post", http.StatusOK, indexContent},
		{"Edit post route", "/edit-post/1", http.StatusOK, indexContent},
		{"Activity route", "/activity", http.StatusOK, indexContent},
		{"Favicon", "/favicon.ico", http.StatusOK, faviconContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
