package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opsflow/common/telemetry"
)

func TestTelemetryMiddleware_TraceparentPropagation(t *testing.T) {
	handler := telemetry.TelemetryMiddleware("test-service")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := telemetry.TraceID(r.Context())
			if traceID == "" {
				t.Error("expected traceID in context, got empty string")
			}
			w.WriteHeader(http.StatusOK)
		}),
	)

	// 1. Without incoming traceparent header -> generates new trace ID
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	respTraceparent := w.Header().Get("traceparent")
	if respTraceparent == "" || !strings.HasPrefix(respTraceparent, "00-") {
		t.Errorf("expected valid traceparent response header, got %s", respTraceparent)
	}

	// 2. With incoming traceparent header -> preserves incoming trace ID
	incomingTraceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	incomingHeader := "00-" + incomingTraceID + "-00f067aa0ba902b7-01"

	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("traceparent", incomingHeader)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	respTraceparent = w.Header().Get("traceparent")
	if !strings.Contains(respTraceparent, incomingTraceID) {
		t.Errorf("expected traceparent header to preserve trace_id %s, got %s", incomingTraceID, respTraceparent)
	}
}
