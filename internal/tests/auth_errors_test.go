// internal/tests/auth_errors_test.go
//
// C08 — auth endpoint error-path coverage: malformed request bodies and the
// POST-only method contract on register/login.
package tests

import (
	"net/http"
	"testing"
)

func TestRegister_MalformedJSONRejected(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(`{not valid json`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed register body, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_MalformedJSONRejected(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(`{not valid json`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed login body, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegister_WrongMethodRejected(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	rec, _ := doRequest(t, h, http.MethodGet, "/api/v1/users/register", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET on register, got %d", rec.Code)
	}
}

func TestLogin_WrongMethodRejected(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	rec, _ := doRequest(t, h, http.MethodGet, "/api/v1/users/login", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET on login, got %d", rec.Code)
	}
}
