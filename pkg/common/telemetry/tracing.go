package telemetry

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type traceContextKey string

const (
	TraceIDHeader                = "traceparent"
	traceIDKey   traceContextKey = "trace_id"
)

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func TraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// ExtractOrGenerateTraceID extracts W3C trace_id from traceparent header or generates a new 16-byte hex string.
func ExtractOrGenerateTraceID(r *http.Request) string {
	tp := r.Header.Get(TraceIDHeader)
	if tp != "" {
		parts := strings.Split(tp, "-")
		if len(parts) >= 4 && len(parts[1]) == 32 {
			return parts[1]
		}
	}

	return generateRandomHex(16)
}

// BuildTraceparent returns a valid W3C traceparent header string.
func BuildTraceparent(traceID string) string {
	if len(traceID) != 32 {
		traceID = generateRandomHex(16)
	}
	parentID := generateRandomHex(8)
	return fmt.Sprintf("00-%s-%s-01", traceID, parentID)
}

func generateRandomHex(bytesCount int) string {
	b := make([]byte, bytesCount)
	_, err := rand.Read(b)
	if err != nil {
		return "0123456789abcdef0123456789abcdef"[:bytesCount*2]
	}
	return hex.EncodeToString(b)
}

// TelemetryMiddleware injects trace_id and correlation context into HTTP requests and response headers.
func TelemetryMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := ExtractOrGenerateTraceID(r)
			traceparent := BuildTraceparent(traceID)

			w.Header().Set(TraceIDHeader, traceparent)

			ctx := ContextWithTraceID(r.Context(), traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
