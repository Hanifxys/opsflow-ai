// Package httputil provides shared HTTP response envelopes, error formatting,
// and utility functions used across all OpsFlow services.
package httputil

import (
	"encoding/json"
	"net/http"
)

// Response is the standard success envelope.
//
//	{ "data": ..., "meta": { "request_id": "..." } }
type Response struct {
	Data any            `json:"data"`
	Meta map[string]any `json:"meta,omitempty"`
}

// ErrorBody is the standard error envelope.
//
//	{ "error": { "code": "...", "message": "...", "request_id": "..." } }
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail holds error specifics.
type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteSuccess writes a standard success response.
func WriteSuccess(w http.ResponseWriter, status int, data any, requestID string) {
	resp := Response{
		Data: data,
	}
	if requestID != "" {
		resp.Meta = map[string]any{"request_id": requestID}
	}
	WriteJSON(w, status, resp)
}

// WriteError writes a standard error response.
func WriteError(w http.ResponseWriter, status int, code, message, requestID string) {
	WriteJSON(w, status, ErrorBody{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	})
}
