package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/opsflow/common/httputil"
	"github.com/redis/go-redis/v9"
)

// RedisRateLimitMiddleware provides a Redis-backed rate limiter with sliding window counter per IP.
// Key naming convention: ratelimit:{ip} (TTL 60s)
// If client is nil or Redis returns an error, it gracefully falls back to the provided fallback Handler.
func RedisRateLimitMiddleware(rdb *redis.Client, requestsPerMinute int, fallbackMiddleware func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	inMemLimiter := fallbackMiddleware
	if inMemLimiter == nil {
		inMemLimiter = RateLimitMiddleware(requestsPerMinute)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rdb == nil {
				// Graceful degradation: fallback to in-memory limiter
				inMemLimiter(next).ServeHTTP(w, r)
				return
			}

			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = xff
			}

			key := fmt.Sprintf("ratelimit:%s", ip)
			ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
			defer cancel()

			// Execute Redis INCR transaction with 60s TTL
			pipe := rdb.Pipeline()
			incrCmd := pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, 60*time.Second)
			_, err := pipe.Exec(ctx)

			if err != nil {
				// Redis error / unavailable -> fallback to in-memory limiter without failing request
				inMemLimiter(next).ServeHTTP(w, r)
				return
			}

			count := incrCmd.Val()
			if count > int64(requestsPerMinute) {
				httputil.WriteError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded. Please try again later.", httputil.RequestID(r.Context()))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
