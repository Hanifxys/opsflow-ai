package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/opsflow/auth-service/internal/adapters/bcrypt"
	"github.com/opsflow/auth-service/internal/application"
	"github.com/opsflow/auth-service/internal/domain"
)

type mockUserRepository struct {
	users         map[string]*domain.User
	userByID      map[uuid.UUID]*domain.User
	refreshTokens map[string]uuid.UUID
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:         make(map[string]*domain.User),
		userByID:      make(map[uuid.UUID]*domain.User),
		refreshTokens: make(map[string]uuid.UUID),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User, roleNames []string) error {
	if _, exists := m.users[user.Email]; exists {
		return domain.ErrEmailAlreadyExists
	}
	m.users[user.Email] = user
	m.userByID[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := m.userByID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return []string{"incident:read", "incident:create"}, nil
}

func (m *mockUserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return []string{"ENGINEER"}, nil
}

func (m *mockUserRepository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	m.refreshTokens[tokenHash] = userID
	return nil
}

func (m *mockUserRepository) ValidateRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	uid, ok := m.refreshTokens[tokenHash]
	if !ok {
		return uuid.Nil, domain.ErrInvalidRefreshToken
	}
	return uid, nil
}

func (m *mockUserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	delete(m.refreshTokens, tokenHash)
	return nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepository()
	hasher := bcrypt.New(4)
	svc := application.NewAuthService(repo, hasher, "test-secret", 15*time.Minute, 7*24*time.Hour)

	ctx := context.Background()

	// 1. Register
	user, err := svc.Register(ctx, "dev@opsflow.local", "securepassword", "Dev Engineer", []string{"ENGINEER"})
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	if user.Email != "dev@opsflow.local" {
		t.Errorf("expected email dev@opsflow.local, got %s", user.Email)
	}

	// 2. Duplicate registration
	_, err = svc.Register(ctx, "dev@opsflow.local", "securepassword", "Dev Engineer", nil)
	if err != domain.ErrEmailAlreadyExists {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}

	// 3. Login success
	access, refresh, loggedInUser, err := svc.Login(ctx, "dev@opsflow.local", "securepassword")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("expected non-empty tokens")
	}
	if loggedInUser.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, loggedInUser.ID)
	}

	// 4. Login wrong password
	_, _, _, err = svc.Login(ctx, "dev@opsflow.local", "wrongpassword")
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	// 5. Refresh token
	newAccess, newRefresh, err := svc.RefreshToken(ctx, refresh)
	if err != nil {
		t.Fatalf("failed to refresh token: %v", err)
	}
	if newAccess == "" || newRefresh == "" {
		t.Fatal("expected new non-empty tokens")
	}
}
