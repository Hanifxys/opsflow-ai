package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opsflow/auth-service/internal/domain"
	"github.com/opsflow/auth-service/internal/ports"
	"github.com/opsflow/common/jwt"
)

type AuthService struct {
	repo       ports.UserRepository
	hasher     ports.PasswordHasher
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(repo ports.UserRepository, hasher ports.PasswordHasher, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		repo:       repo,
		hasher:     hasher,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string, roleNames []string) (*domain.User, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if len(roleNames) == 0 {
		roleNames = []string{"ENGINEER"}
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		DisplayName:  displayName,
		Status:       domain.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, user, roleNames); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, *domain.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", "", nil, domain.ErrInvalidCredentials
		}
		return "", "", nil, err
	}

	if user.Status != domain.UserStatusActive {
		return "", "", nil, domain.ErrUserInactive
	}

	if !s.hasher.Compare(password, user.PasswordHash) {
		return "", "", nil, domain.ErrInvalidCredentials
	}

	perms, err := s.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return "", "", nil, err
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID.String(), user.Email, perms, s.jwtSecret, s.accessTTL)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID.String(), s.jwtSecret, s.refreshTTL)
	if err != nil {
		return "", "", nil, err
	}

	refreshTokenHash := hashToken(refreshToken)
	expiresAt := time.Now().UTC().Add(s.refreshTTL)
	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshTokenHash, expiresAt); err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	tokenHash := hashToken(refreshTokenStr)
	userID, err := s.repo.ValidateRefreshToken(ctx, tokenHash)
	if err != nil {
		return "", "", err
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	if user.Status != domain.UserStatusActive {
		return "", "", domain.ErrUserInactive
	}

	perms, err := s.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	// Revoke old refresh token (rotation)
	_ = s.repo.RevokeRefreshToken(ctx, tokenHash)

	newAccessToken, err := jwt.GenerateAccessToken(user.ID.String(), user.Email, perms, s.jwtSecret, s.accessTTL)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := jwt.GenerateRefreshToken(user.ID.String(), s.jwtSecret, s.refreshTTL)
	if err != nil {
		return "", "", err
	}

	newTokenHash := hashToken(newRefreshToken)
	expiresAt := time.Now().UTC().Add(s.refreshTTL)
	if err := s.repo.SaveRefreshToken(ctx, user.ID, newTokenHash, expiresAt); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*domain.User, []string, []string, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	roles, err := s.repo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	perms, err := s.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	return user, roles, perms, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
