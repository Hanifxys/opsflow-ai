package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsflow/auth-service/internal/domain"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User, roleNames []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert user
	queryUser := `
		INSERT INTO users (id, email, password_hash, display_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.Exec(ctx, queryUser, user.ID, user.Email, user.PasswordHash, user.DisplayName, user.Status, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	// Attach roles
	for _, roleName := range roleNames {
		var roleID uuid.UUID
		err := tx.QueryRow(ctx, "SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
		if err != nil {
			return fmt.Errorf("role %s not found: %w", roleName, err)
		}

		_, err = tx.Exec(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)", user.ID, roleID)
		if err != nil {
			return fmt.Errorf("failed to attach role %s: %w", roleName, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	return u, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	return u, nil
}

func (r *PostgresUserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT p.code
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		perms = append(perms, code)
	}

	return perms, rows.Err()
}

func (r *PostgresUserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}

	return roles, rows.Err()
}

func (r *PostgresUserRepository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) ValidateRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	query := `
		SELECT user_id, expires_at, revoked
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var userID uuid.UUID
	var expiresAt time.Time
	var revoked bool

	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&userID, &expiresAt, &revoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrInvalidRefreshToken
		}
		return uuid.Nil, fmt.Errorf("failed to query refresh token: %w", err)
	}

	if revoked || time.Now().After(expiresAt) {
		return uuid.Nil, domain.ErrInvalidRefreshToken
	}

	return userID, nil
}

func (r *PostgresUserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`
	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}
