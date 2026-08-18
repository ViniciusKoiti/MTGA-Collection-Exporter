package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// apiError is the stable wire shape of every error response.
type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteError emits the stable error envelope; codes are a closed
// vocabulary so clients never parse free-form text.
func WriteError(w http.ResponseWriter, r *http.Request, status int,
	code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]apiError{"error": {
		Code: code, Message: message, RequestID: RequestIDFrom(r.Context()),
	}})
}

// WriteJSON emits a success payload with an explicit status.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// DecodeStrict parses exactly one JSON value, refusing unknown fields
// and trailing garbage; the body budget is enforced upstream.
func DecodeStrict(r *http.Request, target any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("httpapi: invalid json body: %w", err)
	}
	if dec.More() {
		return errors.New("httpapi: trailing data after json body")
	}
	if _, err := dec.Token(); err != io.EOF {
		return errors.New("httpapi: trailing data after json body")
	}
	return nil
}
