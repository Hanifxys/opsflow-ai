package jwt_test

import (
	"testing"
	"time"

	"github.com/opsflow/common/jwt"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "super-secret-key-for-testing"
	userID := "usr-12345"
	email := "test@opsflow.local"
	permissions := []string{"incident:read", "incident:create"}

	token, err := jwt.GenerateAccessToken(userID, email, permissions, secret, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := jwt.ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
	if len(claims.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(claims.Permissions))
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "super-secret-key-for-testing"
	userID := "usr-12345"
	email := "test@opsflow.local"

	token, err := jwt.GenerateAccessToken(userID, email, nil, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = jwt.ValidateToken(token, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	secret := "super-secret-key-for-testing"
	wrongSecret := "wrong-secret-key"
	userID := "usr-12345"

	token, err := jwt.GenerateAccessToken(userID, "test@opsflow.local", nil, secret, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = jwt.ValidateToken(token, wrongSecret)
	if err == nil {
		t.Fatal("expected error for token validated with wrong secret, got nil")
	}
}
