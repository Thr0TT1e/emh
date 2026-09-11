package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// AuthResult результат успешной аутентификации.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	ExpiresIn    int64
}

// AuthUseCase бизнес-логика аутентификации администраторов.
type AuthUseCase interface {
	Login(ctx context.Context, username, password string) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authUseCase struct {
	admins           map[string]domain.AdminUser
	jwtSecret        string
	jwtExpiry        time.Duration
	refreshExpiry    time.Duration
	refreshTokenRepo repository.RefreshTokenRepository
	alertWorker      *AlertWorker
	logger           *slog.Logger
}

func NewAuthUseCase(
	admins []domain.AdminUser,
	jwtSecret string,
	jwtExpiry time.Duration,
	refreshExpiry time.Duration,
	refreshTokenRepo repository.RefreshTokenRepository,
	alertWorker *AlertWorker,
	logger *slog.Logger,
) AuthUseCase {
	adminMap := make(map[string]domain.AdminUser, len(admins))
	for _, a := range admins {
		adminMap[a.Username] = a
	}
	return &authUseCase{
		admins:           adminMap,
		jwtSecret:        jwtSecret,
		jwtExpiry:        jwtExpiry,
		refreshExpiry:    refreshExpiry,
		refreshTokenRepo: refreshTokenRepo,
		alertWorker:      alertWorker,
		logger:           logger,
	}
}

// Фиктивный хеш — всегда одинаковый, чтобы время было стабильным
// Это валидный bcrypt-хеш для пароля "dummy", но мы его ни с чем не сравниваем
const dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// Login проверяет учётные данные и выдаёт пару access/refresh токенов.
func (uc *authUseCase) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	admin, ok := uc.admins[username]
	var hashToCheck string
	if !ok {
		// Юзер не найден, но ВСЁ РАВНО вызываем bcrypt для constant-time
		hashToCheck = dummyPasswordHash
	} else {
		hashToCheck = admin.PasswordHash
	}

	// bcrypt вызывается ВСЕГДА — timing attack невозможен
	if !auth.CheckPassword(hashToCheck, password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Если юзер не существовал — ошибка (но время уже "сожжено" на bcrypt)
	if !ok {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Метрики инкрементируются только ПОСЛЕ успешной проверки пароля —
	// иначе анонимный Login-спам раздувает cardinality gauge по username
	// и искажает счётчик выданных токенов (H2, security.md).
	metrics.AuthTokensIssuedTotal.WithLabelValues("access").Inc()
	metrics.AuthTokensIssuedTotal.WithLabelValues("refresh").Inc()
	metrics.AuthActiveSessions.WithLabelValues(username).Inc()

	return uc.issueTokens(ctx, admin)
}

// Refresh обменивает действующий refresh-токен на новую пару.
// Реализует ротацию: старый токен отзывается, новый создаётся.
func (uc *authUseCase) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	hash := auth.HashRefreshToken(refreshToken)

	// 1. Ищем среди активных токенов
	token, err := uc.refreshTokenRepo.GetByHash(ctx, hash)
	if err == nil {
		// Активный токен — стандартная ротация
		admin, ok := uc.admins[token.Username]
		if !ok {
			return nil, fmt.Errorf("invalid refresh token")
		}
		if err := uc.refreshTokenRepo.Revoke(ctx, token.ID); err != nil {
			return nil, fmt.Errorf("revoke old token: %w", err)
		}

		// Метрика — только после подтверждения валидности токена (H2, security.md).
		metrics.AuthTokensIssuedTotal.WithLabelValues("access").Inc()
		metrics.AuthTokensIssuedTotal.WithLabelValues("refresh").Inc()

		return uc.issueTokens(ctx, admin)
	}

	// 2. Токен не найден среди активных — проверяем, не отозван ли он
	revokedToken, err := uc.refreshTokenRepo.GetByHashIncludingRevoked(ctx, hash)
	if err == nil && revokedToken != nil {
		// 🚨 ДЕТЕКТ КРАЖИ: токен был отозван, но его пытаются использовать снова.
		// Это значит, что у атакующего есть копия refresh-токена.
		// Инвалидируем ВСЕ активные токены этого пользователя.
		revokedCount, revokeErr := uc.refreshTokenRepo.RevokeAllForUser(ctx, revokedToken.Username)
		if revokeErr != nil {
			// Логируем, но не теряем основную ошибку
			uc.logger.Error("failed to revoke all tokens after theft detection",
				"username", revokedToken.Username,
				"error", revokeErr,
			)
		} else {
			uc.logger.Warn("refresh token reuse detected — all user sessions invalidated",
				"username", revokedToken.Username,
				"revoked_count", revokedCount,
				"alert_type", "refresh_token_theft",
			)
			// Отправка асинхронного алерта
			if uc.alertWorker != nil {
				uc.alertWorker.Enqueue(domain.Alert{
					Type:      domain.AlertTypeRefreshTokenTheft,
					Subject:   "Обнаружено повторное использование отозванного токена",
					Body:      "Все активные сессии пользователя инвалидированы",
					Timestamp: time.Now(),
					Metadata: map[string]string{
						"username":      revokedToken.Username,
						"revoked_count": fmt.Sprintf("%d", revokedCount),
					},
				})
			}
		}
	}

	return nil, fmt.Errorf("invalid refresh token")
}

// Logout отзывает refresh-токен.
func (uc *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	hash := auth.HashRefreshToken(refreshToken)
	stored, err := uc.refreshTokenRepo.GetByHash(ctx, hash)
	if err == nil && stored != nil {
		metrics.AuthActiveSessions.WithLabelValues(stored.Username).Dec()
	}

	token, err := uc.refreshTokenRepo.GetByHash(ctx, hash)
	if err != nil {
		// Токен не найден или уже отозван — считаем операцию успешной
		return nil
	}

	return uc.refreshTokenRepo.Revoke(ctx, token.ID)
}

// issueTokens генерирует пару access/refresh и сохраняет refresh в БД.
func (uc *authUseCase) issueTokens(ctx context.Context, admin domain.AdminUser) (*AuthResult, error) {
	// Access token (JWT, короткоживущий)
	accessToken, err := auth.GenerateToken(uc.jwtSecret, admin.Username, admin.Role, uc.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// Refresh token (случайная строка, долгоживущий)
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Сохраняем хеш refresh-токена в БД
	tokenHash := auth.HashRefreshToken(refreshToken)
	_, err = uc.refreshTokenRepo.Create(ctx, domain.CreateRefreshTokenParams{
		Username:  admin.Username,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(uc.refreshExpiry),
	})
	if err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	expiresAt := time.Now().Add(uc.jwtExpiry)
	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		ExpiresIn:    int64(uc.jwtExpiry.Seconds()),
	}, nil
}
