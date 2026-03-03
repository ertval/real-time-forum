package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type errorEnvelope struct {
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestPostPageRoute_InvalidID_BadRequest(t *testing.T) {
	mux := NewMux()

	invalidPaths := []string{
		"/view-post",
		"/view-post/",
		"/view-post/abc",
		"/view-post/0",
		"/post",
		"/post/",
		"/post/abc",
		"/post/0",
	}

	for _, path := range invalidPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Accept", "application/json")

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
			}

			var env errorEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal error envelope: %v", err)
			}

			if env.Error == nil {
				t.Fatalf("expected error envelope, got body=%s", rec.Body.String())
			}

			if env.Error.Code != "BAD_REQUEST" {
				t.Fatalf("expected BAD_REQUEST, got %q", env.Error.Code)
			}
		})
	}
}
