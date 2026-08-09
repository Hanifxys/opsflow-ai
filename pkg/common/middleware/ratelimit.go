package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/opsflow/common/httputil"
)

type clientLimiter struct {
	tokens     float64
	lastUpdate time.Time
}

// RateLimitMiddleware provides a sliding window token-bucket rate limiter per IP address.
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	clients := make(map[string]*clientLimiter)

	maxTokens := float64(requestsPerMinute)
	refillRate := maxTokens / 60.0 // tokens per second

	// Cleanup old entries periodically
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			now := time.Now()
			for ip, cl := range clients {
				if now.Sub(cl.lastUpdate) > 10*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = xff
			}

			mu.Lock()
			now := time.Now()
			cl, exists := clients[ip]
			if !exists {
				cl = &clientLimiter{
					tokens:     maxTokens - 1.0,
					lastUpdate: now,
				}
				clients[ip] = cl
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			elapsed := now.Sub(cl.lastUpdate).Seconds()
			cl.lastUpdate = now
			cl.tokens += elapsed * refillRate
			if cl.tokens > maxTokens {
				cl.tokens = maxTokens
			}

			if cl.tokens < 1.0 {
				mu.Unlock()
				httputil.WriteError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded. Please try again later.", httputil.RequestID(r.Context()))
				return
			}

			cl.tokens -= 1.0
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
