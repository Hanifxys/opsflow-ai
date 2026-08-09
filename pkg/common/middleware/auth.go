package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/opsflow/common/httputil"
	"github.com/opsflow/common/jwt"
)

type authContextKey string

const (
	userIDKey      authContextKey = "user_id"
	userEmailKey   authContextKey = "user_email"
	userPermsKey   authContextKey = "user_permissions"
)

func UserID(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

func UserEmail(ctx context.Context) string {
	if v, ok := ctx.Value(userEmailKey).(string); ok {
		return v
	}
	return ""
}

func UserPermissions(ctx context.Context) []string {
	if v, ok := ctx.Value(userPermsKey).([]string); ok {
		return v
	}
	return nil
}

// AuthMiddleware validates Bearer JWT token from Authorization header and populates context.
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid Authorization header", httputil.RequestID(r.Context()))
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwt.ValidateToken(tokenStr, secret)
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token", httputil.RequestID(r.Context()))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, userEmailKey, claims.Email)
			ctx = context.WithValue(ctx, userPermsKey, claims.Permissions)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
