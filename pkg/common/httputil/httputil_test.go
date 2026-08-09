package httputil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsflow/common/httputil"
)

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	httputil.WriteSuccess(w, http.StatusOK, map[string]string{"hello": "world"}, "req-001")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Fatalf("expected JSON content type, got %s", ct)
	}

	var resp httputil.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Meta["request_id"] != "req-001" {
		t.Fatalf("expected request_id req-001, got %v", resp.Meta["request_id"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", "req-002")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	var resp httputil.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode error: %v", err)
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected code NOT_FOUND, got %s", resp.Error.Code)
	}
	if resp.Error.RequestID != "req-002" {
		t.Fatalf("expected request_id req-002, got %s", resp.Error.RequestID)
	}
}

func TestHealthHandler(t *testing.T) {
	handler := httputil.HealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", body["status"])
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	handler := httputil.HealthHandler()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", w.Code)
	}
}

func TestReadyHandler(t *testing.T) {
	handler, setReady := httputil.ReadyHandler()

	// Default: ready.
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when ready, got %d", w.Code)
	}

	// Set not ready.
	setReady(false)
	req = httptest.NewRequest(http.MethodGet, "/ready", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when not ready, got %d", w.Code)
	}

	// Set ready again.
	setReady(true)
	req = httptest.NewRequest(http.MethodGet, "/ready", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when ready again, got %d", w.Code)
	}
}

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := httputil.RequestID(r.Context())
		if id == "" {
			t.Fatal("expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := httputil.RequestIDMiddleware(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header")
	}
}

func TestRequestIDMiddleware_ReusesClientID(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := httputil.RequestID(r.Context())
		if id != "client-req-123" {
			t.Fatalf("expected client-req-123, got %s", id)
		}
	})

	handler := httputil.RequestIDMiddleware(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "client-req-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "client-req-123" {
		t.Fatal("expected echoed client request ID")
	}
}
