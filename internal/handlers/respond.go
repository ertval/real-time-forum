package handlers

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]interface{}

// APIError represents an API error.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// WriteOK sends a 200 OK response with a data payload.
func WriteOK(w http.ResponseWriter, data interface{}, meta interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(envelope{"data": data, "meta": meta})
}

func WriteCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(envelope{"data": data})
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
}

// NewError creates a new APIError.
func NewError(code, message string, status int) *APIError {
	return &APIError{Code: code, Message: message, Status: status}
}
