package domain

import "time"

// AdminUser представляет администратора системы.
type AdminUser struct {
	Username     string
	PasswordHash string
	Role         string
}

// RefreshToken представляет сессию администратора (refresh-токен).
type RefreshToken struct {
	ID        string
	Username  string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

// CreateRefreshTokenParams параметры создания refresh-токена.
type CreateRefreshTokenParams struct {
	Username  string
	TokenHash string
	ExpiresAt time.Time
}
