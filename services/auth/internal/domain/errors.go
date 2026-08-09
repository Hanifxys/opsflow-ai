package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyExists  = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserInactive        = errors.New("user account is inactive")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)
