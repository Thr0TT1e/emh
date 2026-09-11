package domain

import "time"

// APIKey представляет управляемый API-ключ для доступа к админским сервисам.
type APIKey struct {
	ID          string
	KeyID       string
	SecretHash  string
	Name        string
	Description string
	Role        string
	CreatedBy   string
	RevokedAt   *time.Time
	ExpiresAt   *time.Time
	LastUsedAt  *time.Time
	LastUsedIP  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsActive проверяет, что ключ действующий (не отозван и не истёк).
func (k *APIKey) IsActive(now time.Time) bool {
	if k.RevokedAt != nil {
		return false
	}
	if k.ExpiresAt != nil && now.After(*k.ExpiresAt) {
		return false
	}
	return true
}

// CreateAPIKeyParams параметры создания API-ключа.
type CreateAPIKeyParams struct {
	Name        string
	Description string
	Role        string
	CreatedBy   string
	ExpiresAt   *time.Time
}
