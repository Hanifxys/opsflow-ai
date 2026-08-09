package middleware

import (
	"net/http"

	"github.com/opsflow/common/httputil"
)

// RequirePermission checks if the authenticated user possesses the required permission.
func RequirePermission(requiredPerm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			perms := UserPermissions(r.Context())
			hasPerm := false
			for _, p := range perms {
				if p == requiredPerm || p == "admin:all" {
					hasPerm = true
					break
				}
			}

			if !hasPerm {
				httputil.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions for this resource", httputil.RequestID(r.Context()))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
