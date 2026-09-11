package usecase

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// lastUsedThrottle — минимальный интервал обновления last_used_at.
// Снижает нагрузку на БД при частом использовании ключа.
const lastUsedThrottle = time.Minute

// APIKeyUseCase бизнес-логика управления API-ключами.
type APIKeyUseCase interface {
	Create(ctx context.Context, p domain.CreateAPIKeyParams) (key *domain.APIKey, fullKey string, err error)
	List(ctx context.Context, includeRevoked bool) ([]*domain.APIKey, error)
	Revoke(ctx context.Context, id string) error
	// Authenticate валидирует полный ключ и возвращает соответствующий APIKey.
	Authenticate(ctx context.Context, fullKey string, ip string) (*domain.APIKey, error)
}

type apiKeyUseCase struct {
	repo repository.APIKeyRepository
}

func NewAPIKeyUseCase(repo repository.APIKeyRepository) APIKeyUseCase {
	return &apiKeyUseCase{repo: repo}
}

func (uc *apiKeyUseCase) Create(ctx context.Context, p domain.CreateAPIKeyParams) (*domain.APIKey, string, error) {
	if p.Name == "" {
		return nil, "", fmt.Errorf("api key name is required")
	}
	if p.Role == "" {
		p.Role = "admin"
	}

	keyID, secret, fullKey, err := auth.GenerateAPIKey()
	if err != nil {
		return nil, "", err
	}

	// Хешируем секрет: plaintext никогда не сохраняется
	secretHash, err := auth.HashPassword(secret)
	if err != nil {
		return nil, "", fmt.Errorf("hash secret: %w", err)
	}

	key := &domain.APIKey{
		KeyID:       keyID,
		SecretHash:  secretHash,
		Name:        p.Name,
		Description: p.Description,
		Role:        p.Role,
		CreatedBy:   p.CreatedBy,
		ExpiresAt:   p.ExpiresAt,
	}

	if err := uc.repo.Create(ctx, key); err != nil {
		return nil, "", err
	}

	// fullKey возвращается только здесь — повторно получить секрет невозможно
	return key, fullKey, nil
}

func (uc *apiKeyUseCase) List(ctx context.Context, includeRevoked bool) ([]*domain.APIKey, error) {
	return uc.repo.List(ctx, includeRevoked)
}

func (uc *apiKeyUseCase) Revoke(ctx context.Context, id string) error {
	return uc.repo.Revoke(ctx, id)
}

func (uc *apiKeyUseCase) Authenticate(ctx context.Context, fullKey string, ip string) (*domain.APIKey, error) {
	keyID, secret, ok := auth.ParseAPIKey(fullKey)
	if !ok {
		return nil, fmt.Errorf("malformed api key")
	}

	key, err := uc.repo.GetByKeyID(ctx, keyID)
	if err != nil {
		// Не раскрываем, существует ли ключ
		return nil, fmt.Errorf("invalid api key")
	}

	if !key.IsActive(time.Now()) {
		return nil, fmt.Errorf("api key revoked or expired")
	}

	if !auth.CheckPassword(key.SecretHash, secret) {
		return nil, fmt.Errorf("invalid api key")
	}

	// Аудит использования: обновляем last_used с троттлингом, best-effort
	// (ошибка аудита не должна блокировать аутентификацию)
	now := time.Now()
	if key.LastUsedAt == nil || now.Sub(*key.LastUsedAt) > lastUsedThrottle {
		_ = uc.repo.UpdateLastUsed(ctx, key.ID, ip, now)
	}

	return key, nil
}
