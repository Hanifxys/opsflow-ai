package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	DisplayName  string     `json:"display_name"`
	Status       UserStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
}

type RefreshToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}
