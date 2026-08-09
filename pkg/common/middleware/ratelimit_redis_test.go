package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsflow/common/middleware"
)

func TestRedisRateLimitMiddleware_FallbackWhenNil(t *testing.T) {
	// rdb is nil -> should fallback to in-memory limiter gracefully without failing
	handler := middleware.RedisRateLimitMiddleware(nil, 5, middleware.RateLimitMiddleware(5))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	// First 5 requests should pass
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200 OK, got %d", i+1, w.Code)
		}
	}

	// 6th request should hit rate limit
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", w.Code)
	}
}
