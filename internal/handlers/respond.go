package handlers

import (
	"encoding/json"
	"net/http"
)

// APIResponse is the unified JSON envelope for success & errors.
type APIResponse struct {
	Data  any       `json:"data,omitempty"`
	Meta  any       `json:"meta,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

// APIError is a standardized error message for JSON responses.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// --------------------------
// SUCCESS RESPONSES
// --------------------------

func WriteOK(w http.ResponseWriter, data any, meta any) {
	writeJSON(w, http.StatusOK, &APIResponse{
		Data: data,
		Meta: meta,
	})
}

func WriteCreated(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, &APIResponse{
		Data: data,
	})
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// --------------------------
// ERROR RESPONSE
// --------------------------

func WriteError(w http.ResponseWriter, err *APIError) {
	writeJSON(w, err.Status, &APIResponse{
		Error: err,
	})
}

func NewError(code, message string, status int) *APIError {
	return &APIError{Code: code, Message: message, Status: status}
}

// --------------------------
// INTERNAL JSON WRITER
// --------------------------

func writeJSON(w http.ResponseWriter, status int, body *APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
