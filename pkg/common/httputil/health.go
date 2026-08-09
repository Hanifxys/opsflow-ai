package httputil

import (
	"net/http"
	"sync/atomic"
)

// HealthHandler returns a handler that responds to GET /health.
// It always returns 200 with {"status": "ok"}.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", RequestID(r.Context()))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// ReadyHandler returns a handler that responds to GET /ready.
// The ready state is controlled externally via the returned setter function.
func ReadyHandler() (http.HandlerFunc, func(bool)) {
	var ready atomic.Bool
	ready.Store(true)

	setter := func(isReady bool) {
		ready.Store(isReady)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", RequestID(r.Context()))
			return
		}
		if !ready.Load() {
			WriteError(w, http.StatusServiceUnavailable, "NOT_READY", "Service is not ready", RequestID(r.Context()))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}

	return handler, setter
}
