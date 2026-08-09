package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opsflow/common/jwt"
	"github.com/opsflow/common/telemetry"
)

func TestOpsFlow_IntegrationValidation(t *testing.T) {
	jwtSecret := "test-secret-key"

	// 1. Validate JWT Token Generation & Verification
	token, err := jwt.GenerateAccessToken("u-123", "operator@opsflow.local", []string{"incident:write", "ai:use"}, jwtSecret, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	claims, err := jwt.ValidateToken(token, jwtSecret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.Email != "operator@opsflow.local" {
		t.Errorf("expected email operator@opsflow.local, got %s", claims.Email)
	}

	// 2. Validate W3C Telemetry Context Propagation
	traceMiddleware := telemetry.TelemetryMiddleware("integration-test")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := telemetry.TraceID(r.Context())
			if traceID == "" {
				t.Error("expected traceID in context")
			}
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	w := httptest.NewRecorder()
	traceMiddleware.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
