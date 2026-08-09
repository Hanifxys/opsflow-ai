package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

type Claims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// GenerateAccessToken builds a signed JWT access token.
func GenerateAccessToken(userID, email string, permissions []string, secret string, duration time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:      userID,
		Email:       email,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "opsflow",
			Audience:  jwt.ClaimStrings{"opsflow-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken builds a signed JWT refresh token.
func GenerateRefreshToken(userID string, secret string, duration time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "opsflow",
		Audience:  jwt.ClaimStrings{"opsflow-refresh"},
		ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return signed, nil
}

// ValidateToken parses and validates a JWT access token string.
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
