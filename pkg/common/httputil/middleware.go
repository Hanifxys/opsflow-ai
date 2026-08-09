package httputil

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// RequestID extracts the request ID from context.
func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// RequestIDMiddleware injects a unique request/correlation ID into every request.
// If the client sends X-Request-ID, it is reused; otherwise a new UUID is generated.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs every request with method, path, status, duration, request_id, and trace_id.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			traceID := ""
			if v, ok := r.Context().Value("trace_id").(string); ok {
				traceID = v
			}

			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.String("request_id", RequestID(r.Context())),
				slog.String("trace_id", traceID),
			)
		})
	}
}

// RecoverMiddleware recovers from panics and returns 500.
func RecoverMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						slog.Any("error", err),
						slog.String("request_id", RequestID(r.Context())),
					)
					WriteError(w, http.StatusInternalServerError,
						"INTERNAL_ERROR", "An internal error occurred", RequestID(r.Context()))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// statusWriter wraps ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
