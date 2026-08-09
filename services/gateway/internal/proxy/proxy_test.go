package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opsflow/common/jwt"
	"github.com/opsflow/gateway/internal/proxy"
)

func TestGatewayProxy_JWTValidationAndHeaderInjection(t *testing.T) {
	jwtSecret := "test-gateway-secret"

	// Mock upstream service
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			// Public route — no headers expected
			w.WriteHeader(http.StatusOK)
			return
		}

		userID := r.Header.Get("X-User-ID")
		userEmail := r.Header.Get("X-User-Email")
		userPerms := r.Header.Get("X-User-Permissions")

		if userID != "usr-100" || userEmail != "user@opsflow.local" || userPerms != "incident:read" {
			t.Errorf("unexpected headers: userID=%s email=%s perms=%s", userID, userEmail, userPerms)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer upstream.Close()

	routes := []proxy.Route{
		{Prefix: "/api/v1/incidents", Upstream: upstream.URL, Public: false},
		{Prefix: "/api/v1/auth", Upstream: upstream.URL, Public: false},
	}

	handler, err := proxy.NewGatewayHandler(routes, jwtSecret, nil)
	if err != nil {
		t.Fatalf("failed to create gateway handler: %v", err)
	}

	// 1. Request without token -> 401 Unauthorized
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", w.Code)
	}

	// 2. Request to public route /login without token -> 200 OK
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for public route, got %d", w.Code)
	}

	// 3. Request to protected route with valid token -> 200 OK & Header Injection
	token, err := jwt.GenerateAccessToken("usr-100", "user@opsflow.local", []string{"incident:read"}, jwtSecret, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for valid token request, got %d", w.Code)
	}
}
