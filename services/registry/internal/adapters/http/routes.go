package http

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/opsflow/common/httputil"
	"github.com/opsflow/common/middleware"
)

func RegisterRoutes(mux *http.ServeMux, handler *RegistryHandler, jwtSecret string) {
	authMw := middleware.AuthMiddleware(jwtSecret)

	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/services")
		path = strings.TrimPrefix(path, "/")

		// Base root route: GET / or POST /
		if path == "" {
			if r.Method == http.MethodPost {
				middleware.RequirePermission("service:write")(http.HandlerFunc(handler.Create)).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodGet {
				middleware.RequirePermission("service:read")(http.HandlerFunc(handler.List)).ServeHTTP(w, r)
				return
			}
			httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", httputil.RequestID(r.Context()))
			return
		}

		parts := strings.Split(path, "/")
		id, err := uuid.Parse(parts[0])
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid service ID", httputil.RequestID(r.Context()))
			return
		}

		if len(parts) == 1 {
			if r.Method == http.MethodGet {
				middleware.RequirePermission("service:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					handler.Get(w, r, id)
				})).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodDelete {
				middleware.RequirePermission("service:write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					handler.Delete(w, r, id)
				})).ServeHTTP(w, r)
				return
			}
			httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", httputil.RequestID(r.Context()))
			return
		}

		if len(parts) == 2 {
			sub := parts[1]
			switch sub {
			case "environments":
				if r.Method == http.MethodPost {
					middleware.RequirePermission("service:write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						handler.AddEnvironment(w, r, id)
					})).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodGet {
					middleware.RequirePermission("service:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						handler.ListEnvironments(w, r, id)
					})).ServeHTTP(w, r)
					return
				}
			case "dependencies":
				if r.Method == http.MethodPost {
					middleware.RequirePermission("service:write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						handler.AddDependency(w, r, id)
					})).ServeHTTP(w, r)
					return
				}
				if r.Method == http.MethodGet {
					middleware.RequirePermission("service:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						handler.ListDependencies(w, r, id)
					})).ServeHTTP(w, r)
					return
				}
			}
		}

		if len(parts) == 4 && parts[1] == "environments" && parts[3] == "health-checks" {
			envID, err := uuid.Parse(parts[2])
			if err != nil {
				httputil.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid environment ID", httputil.RequestID(r.Context()))
				return
			}
			if r.Method == http.MethodPost {
				middleware.RequirePermission("service:write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					handler.AddHealthCheck(w, r, envID)
				})).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodGet {
				middleware.RequirePermission("service:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					handler.ListHealthChecks(w, r, envID)
				})).ServeHTTP(w, r)
				return
			}
		}

		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Endpoint not found", httputil.RequestID(r.Context()))
	})

	mux.Handle("/", authMw(rootHandler))
}
