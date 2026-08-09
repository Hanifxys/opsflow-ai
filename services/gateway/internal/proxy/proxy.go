package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/opsflow/common/jwt"
)

type Route struct {
	Prefix   string
	Upstream string
	Public   bool // if true, bypasses JWT validation
}

// NewGatewayHandler sets up the reverse proxy routes with JWT validation and header injection.
func NewGatewayHandler(routes []Route, jwtSecret string, loggerHandler http.HandlerFunc) (http.Handler, error) {
	mux := http.NewServeMux()

	for _, route := range routes {
		target, err := url.Parse(route.Upstream)
		if err != nil {
			return nil, err
		}

		revProxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.SetURL(target)
				pr.Out.Header.Set("X-Forwarded-Host", pr.In.Host)
				if reqID := pr.In.Header.Get("X-Request-ID"); reqID != "" {
					pr.Out.Header.Set("X-Request-ID", reqID)
				}
				if userID := pr.In.Header.Get("X-User-ID"); userID != "" {
					pr.Out.Header.Set("X-User-ID", userID)
				}
				if userEmail := pr.In.Header.Get("X-User-Email"); userEmail != "" {
					pr.Out.Header.Set("X-User-Email", userEmail)
				}
				if userPerms := pr.In.Header.Get("X-User-Permissions"); userPerms != "" {
					pr.Out.Header.Set("X-User-Permissions", userPerms)
				}
			},
		}

		prefix := route.Prefix
		isPublic := route.Public

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Trim route prefix before forwarding if needed
			forwardPath := strings.TrimPrefix(r.URL.Path, prefix)
			if !strings.HasPrefix(forwardPath, "/") {
				forwardPath = "/" + forwardPath
			}
			r.URL.Path = forwardPath

			// Check public exemption (e.g. /login, /refresh)
			if !isPublic && !isPublicEndpoint(r.URL.Path, prefix) {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
					http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Missing Authorization header"}}`, http.StatusUnauthorized)
					return
				}

				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := jwt.ValidateToken(tokenStr, jwtSecret)
				if err != nil {
					http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid or expired token"}}`, http.StatusUnauthorized)
					return
				}

				// Inject identity headers into request before proxying
				r.Header.Set("X-User-ID", claims.UserID)
				r.Header.Set("X-User-Email", claims.Email)
				r.Header.Set("X-User-Permissions", strings.Join(claims.Permissions, ","))
			}

			revProxy.ServeHTTP(w, r)
		})

		mux.Handle(prefix+"/", handler)
		mux.Handle(prefix, handler)
	}

	return mux, nil
}

func isPublicEndpoint(path string, prefix string) bool {
	if prefix == "/api/v1/auth" {
		return path == "/login" || path == "/refresh"
	}
	return false
}
