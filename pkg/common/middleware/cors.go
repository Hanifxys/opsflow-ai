package middleware

import (
	"net/http"
	"strings"
)

// CORSMiddleware handles CORS preflight requests and response headers.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				for _, o := range allowedOrigins {
					if strings.EqualFold(o, origin) {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Idempotency-Key")
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
