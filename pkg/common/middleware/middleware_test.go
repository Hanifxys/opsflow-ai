package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opsflow/common/jwt"
	"github.com/opsflow/common/middleware"
)

func TestAuthAndRBACMiddleware(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateAccessToken("user-1", "user@opsflow.local", []string{"incident:read"}, secret, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := middleware.AuthMiddleware(secret)(
		middleware.RequirePermission("incident:read")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := middleware.UserID(r.Context())
				if userID != "user-1" {
					t.Errorf("expected user-1, got %s", userID)
				}
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	// Test valid request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	// Test missing permission
	handlerForbidden := middleware.AuthMiddleware(secret)(
		middleware.RequirePermission("incident:delete")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	handlerForbidden.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", w.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := middleware.CORSMiddleware([]string{"http://localhost:3000"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for preflight, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("unexpected allow origin header: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
