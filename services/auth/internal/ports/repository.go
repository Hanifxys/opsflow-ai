package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/opsflow/auth-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User, roleNames []string) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)

	SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}
