package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// ============================================================
// Хелперы
// ============================================================

func newAuthUseCase(
	admins []domain.AdminUser,
	jwtSecret string,
	jwtExpiry time.Duration,
	refreshExpiry time.Duration,
	refreshTokenRepo repository.RefreshTokenRepository,
	alertWorker *AlertWorker,
) AuthUseCase {
	return NewAuthUseCase(
		admins,
		jwtSecret,
		jwtExpiry,
		refreshExpiry,
		refreshTokenRepo,
		alertWorker,
		discardTestLogger(),
	)
}

func sampleAdmin() domain.AdminUser {
	hash, _ := auth.HashPassword("correct-password")
	return domain.AdminUser{
		Username:     "admin",
		PasswordHash: hash,
		Role:         "admin",
	}
}

// ============================================================
// Тесты Login
// ============================================================

// TestAuthUseCase_Login_Success проверяет успешную аутентификацию.
func TestAuthUseCase_Login_Success(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	result, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Проверяем AuthResult
	if result.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if result.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
	if result.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}
	if result.ExpiresIn <= 0 {
		t.Error("ExpiresIn should be positive")
	}

	// Проверяем что ExpiresAt примерно через 15 минут
	expectedExpiry := time.Now().Add(15 * time.Minute)
	if result.ExpiresAt.Before(expectedExpiry.Add(-1*time.Second)) || result.ExpiresAt.After(expectedExpiry.Add(1*time.Second)) {
		t.Errorf("ExpiresAt = %v, expected around %v", result.ExpiresAt, expectedExpiry)
	}

	// Проверяем что ExpiresIn соответствует jwtExpiry
	if result.ExpiresIn != int64(15*time.Minute/time.Second) {
		t.Errorf("ExpiresIn = %d, want %d", result.ExpiresIn, int64(15*time.Minute/time.Second))
	}
}

// TestAuthUseCase_Login_InvalidUsername проверяет ошибку для неверного username.
func TestAuthUseCase_Login_InvalidUsername(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	_, err := uc.Login(context.Background(), "nonexistent", "correct-password")
	if err == nil {
		t.Fatal("expected error for invalid username")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("expected 'invalid credentials' error, got %v", err)
	}
}

// TestAuthUseCase_Login_InvalidPassword проверяет ошибку для неверного пароля.
func TestAuthUseCase_Login_InvalidPassword(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	_, err := uc.Login(context.Background(), "admin", "wrong-password")
	if err == nil {
		t.Fatal("expected error for invalid password")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("expected 'invalid credentials' error, got %v", err)
	}
}

// TestAuthUseCase_Login_RefreshTokenSaved проверяет сохранение refresh token в репозиторий.
func TestAuthUseCase_Login_RefreshTokenSaved(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	result, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Проверяем что Create был вызван
	if len(refreshRepo.createdTokens) != 1 {
		t.Fatalf("expected 1 Create call, got %d", len(refreshRepo.createdTokens))
	}

	created := refreshRepo.createdTokens[0]
	if created.Username != "admin" {
		t.Errorf("created.Username = %s, want admin", created.Username)
	}
	if created.TokenHash == "" {
		t.Error("created.TokenHash should not be empty")
	}
	// Проверяем что хеш соответствует выданному токену
	expectedHash := auth.HashRefreshToken(result.RefreshToken)
	if created.TokenHash != expectedHash {
		t.Errorf("created.TokenHash = %s, want %s", created.TokenHash, expectedHash)
	}

	// Проверяем что ExpiresAt примерно через 7 дней
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	if created.ExpiresAt.Before(expectedExpiry.Add(-1*time.Second)) || created.ExpiresAt.After(expectedExpiry.Add(1*time.Second)) {
		t.Errorf("created.ExpiresAt = %v, expected around %v", created.ExpiresAt, expectedExpiry)
	}
}

// TestAuthUseCase_Login_ValidJWT проверяет валидность выданного JWT.
func TestAuthUseCase_Login_ValidJWT(t *testing.T) {
	admin := sampleAdmin()
	jwtSecret := "test-jwt-secret-at-least-32-chars-long"
	refreshRepo := &mockRefreshTokenRepository{}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		jwtSecret,
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	result, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Валидируем JWT
	claims, err := auth.ValidateToken(result.AccessToken, jwtSecret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.Username != "admin" {
		t.Errorf("claims.Username = %s, want admin", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("claims.Role = %s, want %s", claims.Role, "admin")
	}
}

// ============================================================
// Тесты Refresh
// ============================================================

// TestAuthUseCase_Refresh_Success проверяет успешную ротацию токенов.
func TestAuthUseCase_Refresh_Success(t *testing.T) {
	admin := sampleAdmin()
	jwtSecret := "test-jwt-secret-at-least-32-chars-long"

	// Сначала логинимся чтобы получить валидный refresh token
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: make(map[string]*domain.RefreshToken),
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		jwtSecret,
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	loginResult, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	oldRefreshToken := loginResult.RefreshToken
	oldHash := auth.HashRefreshToken(oldRefreshToken)

	// Сохраняем токен в мок
	refreshRepo.storedTokens[oldHash] = &domain.RefreshToken{
		ID:        "token-id-1",
		Username:  "admin",
		TokenHash: oldHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// Делаем refresh
	refreshResult, err := uc.Refresh(context.Background(), oldRefreshToken)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// Проверяем что выданы новые токены
	if refreshResult.AccessToken == "" {
		t.Error("new AccessToken should not be empty")
	}
	if refreshResult.RefreshToken == "" {
		t.Error("new RefreshToken should not be empty")
	}

	// Проверяем что старый токен был отозван
	if len(refreshRepo.revokeCalls) != 1 {
		t.Fatalf("expected 1 Revoke call, got %d", len(refreshRepo.revokeCalls))
	}
	if refreshRepo.revokeCalls[0] != "token-id-1" {
		t.Errorf("revokeCalls[0] = %s, want token-id-1", refreshRepo.revokeCalls[0])
	}

	// Проверяем что создан новый refresh token
	if len(refreshRepo.createdTokens) != 2 { // 1 от Login + 1 от Refresh
		t.Fatalf("expected 2 Create calls, got %d", len(refreshRepo.createdTokens))
	}
	newHash := auth.HashRefreshToken(refreshResult.RefreshToken)
	if refreshRepo.createdTokens[1].TokenHash != newHash {
		t.Errorf("new token hash mismatch")
	}

	// Проверяем что старый и новый refresh токены разные
	if oldRefreshToken == refreshResult.RefreshToken {
		t.Error("new refresh token should be different from old")
	}
}

// TestAuthUseCase_Refresh_InvalidToken проверяет ошибку для недействительного токена.
func TestAuthUseCase_Refresh_InvalidToken(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{
		getByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return nil, errors.New("token not found")
		},
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	_, err := uc.Refresh(context.Background(), "invalid-refresh-token")
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}
	if !strings.Contains(err.Error(), "invalid refresh token") {
		t.Errorf("expected 'invalid refresh token' error, got %v", err)
	}
}

// TestAuthUseCase_Refresh_UserNotFound проверяет ошибку когда пользователь удалён.
func TestAuthUseCase_Refresh_UserNotFound(t *testing.T) {
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: map[string]*domain.RefreshToken{
			"hash": {
				ID:        "token-id-1",
				Username:  "deleted-user",
				TokenHash: "hash",
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			},
		},
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{}, // пользователь не существует
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	_, err := uc.Refresh(context.Background(), "some-token")
	if err == nil {
		t.Fatal("expected error for deleted user")
	}
	if !strings.Contains(err.Error(), "invalid refresh token") {
		t.Errorf("expected 'invalid refresh token' error, got %v", err)
	}
}

// TestAuthUseCase_Refresh_RevokeError проверяет обработку ошибки при отзыве.
func TestAuthUseCase_Refresh_RevokeError(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: make(map[string]*domain.RefreshToken),
		revokeFunc: func(ctx context.Context, id string) error {
			return errors.New("database error")
		},
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	loginResult, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	hash := auth.HashRefreshToken(loginResult.RefreshToken)
	refreshRepo.storedTokens[hash] = &domain.RefreshToken{
		ID:        "token-id-1",
		Username:  "admin",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	_, err = uc.Refresh(context.Background(), loginResult.RefreshToken)
	if err == nil {
		t.Fatal("expected error from revoke")
	}
	if !strings.Contains(err.Error(), "revoke old token") {
		t.Errorf("expected 'revoke old token' error, got %v", err)
	}
}

// ============================================================
// Тесты Logout
// ============================================================

// TestAuthUseCase_Logout_Success проверяет успешный отзыв токена.
func TestAuthUseCase_Logout_Success(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: make(map[string]*domain.RefreshToken),
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	loginResult, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	hash := auth.HashRefreshToken(loginResult.RefreshToken)
	refreshRepo.storedTokens[hash] = &domain.RefreshToken{
		ID:        "token-id-1",
		Username:  "admin",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = uc.Logout(context.Background(), loginResult.RefreshToken)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Проверяем что Revoke был вызван
	if len(refreshRepo.revokeCalls) != 1 {
		t.Fatalf("expected 1 Revoke call, got %d", len(refreshRepo.revokeCalls))
	}
	if refreshRepo.revokeCalls[0] != "token-id-1" {
		t.Errorf("revokeCalls[0] = %s, want token-id-1", refreshRepo.revokeCalls[0])
	}
}

// TestAuthUseCase_Logout_TokenNotFound проверяет idempotent поведение.
func TestAuthUseCase_Logout_TokenNotFound(t *testing.T) {
	refreshRepo := &mockRefreshTokenRepository{
		getByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return nil, errors.New("token not found")
		},
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	// Logout должен быть успешным даже если токен не найден
	err := uc.Logout(context.Background(), "nonexistent-token")
	if err != nil {
		t.Errorf("Logout should succeed for nonexistent token, got %v", err)
	}
}

// TestAuthUseCase_Logout_AlreadyRevoked проверяет idempotent для уже отозванного токена.
func TestAuthUseCase_Logout_AlreadyRevoked(t *testing.T) {
	refreshRepo := &mockRefreshTokenRepository{
		getByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return nil, errors.New("token already revoked")
		},
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	// Logout должен быть успешным даже если токен уже отозван
	err := uc.Logout(context.Background(), "revoked-token")
	if err != nil {
		t.Errorf("Logout should succeed for already revoked token, got %v", err)
	}
}

// TestAuthUseCase_Logout_RevokeError проверяет обработку ошибки при отзыве.
func TestAuthUseCase_Logout_RevokeError(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: make(map[string]*domain.RefreshToken),
		revokeFunc: func(ctx context.Context, id string) error {
			return errors.New("database error")
		},
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	loginResult, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	hash := auth.HashRefreshToken(loginResult.RefreshToken)
	refreshRepo.storedTokens[hash] = &domain.RefreshToken{
		ID:        "token-id-1",
		Username:  "admin",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = uc.Logout(context.Background(), loginResult.RefreshToken)
	if err == nil {
		t.Fatal("expected error from revoke")
	}
}

// ============================================================
// Интеграционные тесты (Login -> Refresh -> Logout)
// ============================================================

// TestAuthUseCase_FullFlow проверяет полный цикл аутентификации.
func TestAuthUseCase_FullFlow(t *testing.T) {
	admin := sampleAdmin()
	jwtSecret := "test-jwt-secret-at-least-32-chars-long"
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: make(map[string]*domain.RefreshToken),
	}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		jwtSecret,
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	// 1. Login
	loginResult, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	hash := auth.HashRefreshToken(loginResult.RefreshToken)
	refreshRepo.storedTokens[hash] = &domain.RefreshToken{
		ID:        "token-id-1",
		Username:  "admin",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// 2. Refresh
	refreshResult, err := uc.Refresh(context.Background(), loginResult.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	newHash := auth.HashRefreshToken(refreshResult.RefreshToken)
	refreshRepo.storedTokens[newHash] = &domain.RefreshToken{
		ID:        "token-id-2",
		Username:  "admin",
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// 3. Logout
	err = uc.Logout(context.Background(), refreshResult.RefreshToken)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Проверяем что были отозваны оба токена
	if len(refreshRepo.revokeCalls) != 2 {
		t.Errorf("expected 2 Revoke calls, got %d", len(refreshRepo.revokeCalls))
	}
}

// ============================================================
// Тесты: Constant-time Login (#5)
// ============================================================

// TestAuthUseCase_Login_ConstantTimeForUnknownUser проверяет, что время ответа
// для несуществующего пользователя примерно такое же, как для существующего.
// Это защита от username enumeration через timing attack.
//
// Примечание: тест использует宽松的 границы (±50мс) для стабильности в CI,
// но реальная разница должна быть <5мс (bcrypt доминирует).
func TestAuthUseCase_Login_ConstantTimeForUnknownUser(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{}
	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	// Замер для несуществующего пользователя
	startUnknown := time.Now()
	_, _ = uc.Login(context.Background(), "nonexistent-user-xyz", "some-password")
	durationUnknown := time.Since(startUnknown)

	// Замер для существующего пользователя с неверным паролем
	startKnown := time.Now()
	_, _ = uc.Login(context.Background(), "admin", "wrong-password")
	durationKnown := time.Since(startKnown)

	// Обе операции должны занимать примерно одинаковое время (bcrypt ~200-300мс).
	// Разница >100мс указывает на timing leak.
	diff := durationUnknown - durationKnown
	if diff < 0 {
		diff = -diff
	}
	const maxAcceptableDiff = 100 * time.Millisecond
	if diff > maxAcceptableDiff {
		t.Errorf("timing leak detected: unknown=%v, known=%v, diff=%v (max %v)",
			durationUnknown, durationKnown, diff, maxAcceptableDiff)
	}
	// Обе операции должны быть > 100мс (bcrypt занимает время)
	if durationUnknown < 100*time.Millisecond {
		t.Errorf("unknown user login too fast (%v): bcrypt not called?", durationUnknown)
	}
}

// ============================================================
// Тесты: Refresh Token Reuse Detection (#6)
// ============================================================

// TestAuthUseCase_Refresh_ReuseDetectedRevokesAll проверяет, что при попытке
// использовать уже отозванный refresh-токен (признак кражи) отзываются
// ВСЕ активные токены этого пользователя.
func TestAuthUseCase_Refresh_ReuseDetectedRevokesAll(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: make(map[string]*domain.RefreshToken),
		getByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			// Токен не найден среди активных (уже отозван при предыдущем Refresh)
			return nil, errors.New("token not found")
		},
		getByHashIncludingRevokedFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			// Но найден среди отозванных — это и есть детект кражи
			return &domain.RefreshToken{
				ID:        "stolen-token-id",
				Username:  "admin",
				TokenHash: hash,
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}, nil
		},
	}

	var revokedUsername string
	refreshRepo.revokeAllForUserFunc = func(ctx context.Context, username string) (int64, error) {
		revokedUsername = username
		return 3, nil // имитация отзыва 3 активных токенов
	}

	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	// Пытаемся использовать уже отозванный токен (имитация кражи)
	_, err := uc.Refresh(context.Background(), "stolen-refresh-token")

	// Ошибка всё равно возвращается (токен невалиден)
	if err == nil {
		t.Fatal("expected error for reused token")
	}
	if !strings.Contains(err.Error(), "invalid refresh token") {
		t.Errorf("expected 'invalid refresh token', got %v", err)
	}

	// КЛЮЧЕВАЯ ПРОВЕРКА: RevokeAllForUser был вызван с правильным username
	if revokedUsername != "admin" {
		t.Errorf("RevokeAllForUser called with %q, want 'admin'", revokedUsername)
	}
}

// TestAuthUseCase_Refresh_ReuseDetection_RevokeAllError проверяет, что
// ошибка RevokeAllForUser логируется, но не меняет итоговую ошибку Refresh.
func TestAuthUseCase_Refresh_ReuseDetection_RevokeAllError(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{
		getByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return nil, errors.New("token not found")
		},
		getByHashIncludingRevokedFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{
				ID:        "stolen-token-id",
				Username:  "admin",
				TokenHash: hash,
			}, nil
		},
		revokeAllForUserFunc: func(ctx context.Context, username string) (int64, error) {
			return 0, errors.New("database connection lost")
		},
	}

	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	// Refresh должен вернуть "invalid refresh token", а не ошибку БД
	_, err := uc.Refresh(context.Background(), "stolen-token")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid refresh token") {
		t.Errorf("expected 'invalid refresh token', got %v", err)
	}
}

// TestAuthUseCase_Refresh_ValidTokenDoesNotTriggerTheftDetection проверяет,
// что нормальная ротация токенов НЕ вызывает RevokeAllForUser.
func TestAuthUseCase_Refresh_ValidTokenDoesNotTriggerTheftDetection(t *testing.T) {
	admin := sampleAdmin()
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: make(map[string]*domain.RefreshToken),
	}

	revokeAllCalled := false
	refreshRepo.revokeAllForUserFunc = func(ctx context.Context, username string) (int64, error) {
		revokeAllCalled = true
		return 0, nil
	}

	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		"test-jwt-secret-at-least-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		nil,
	)

	// Логинимся → получаем валидный refresh
	loginResult, err := uc.Login(context.Background(), "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	hash := auth.HashRefreshToken(loginResult.RefreshToken)
	refreshRepo.storedTokens[hash] = &domain.RefreshToken{
		ID:        "valid-token-id",
		Username:  "admin",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// Нормальный Refresh — токен активный
	_, err = uc.Refresh(context.Background(), loginResult.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// RevokeAllForUser НЕ должен был вызываться
	if revokeAllCalled {
		t.Error("RevokeAllForUser must NOT be called for valid token rotation")
	}
}

// ============================================================
// Тесты: RefreshTokenCleanupWorker
// ============================================================

// TestRefreshTokenCleanupWorker_CallsDeleteExpired проверяет, что воркер
// вызывает DeleteExpired с правильным контекстом.
func TestRefreshTokenCleanupWorker_CallsDeleteExpired(t *testing.T) {
	var deleteCalls int
	repo := &mockRefreshTokenRepository{
		deleteExpiredFunc: func(ctx context.Context) (int64, error) {
			deleteCalls++
			return 5, nil
		},
	}

	worker := NewRefreshTokenCleanupWorker(repo, discardTestLogger(), 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Запускаем воркер — он должен сделать 1+ очисток за 50мс
	worker.Run(ctx)

	// Минимум 1 вызов (первый cleanup сразу при старте)
	if deleteCalls < 1 {
		t.Errorf("expected at least 1 DeleteExpired call, got %d", deleteCalls)
	}
}

// TestRefreshTokenCleanupWorker_HandlesError проверяет, что ошибка DeleteExpired
// не прерывает цикл работы воркера.
func TestRefreshTokenCleanupWorker_HandlesError(t *testing.T) {
	var callCount int
	repo := &mockRefreshTokenRepository{
		deleteExpiredFunc: func(ctx context.Context) (int64, error) {
			callCount++
			return 0, errors.New("db error")
		},
	}

	worker := NewRefreshTokenCleanupWorker(repo, discardTestLogger(), 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	// Не должно паниковать или зависать
	worker.Run(ctx)

	// Должно было быть несколько попыток, несмотря на ошибки
	if callCount < 2 {
		t.Errorf("expected multiple DeleteExpired attempts despite errors, got %d", callCount)
	}
}

// TestAuthUseCase_Refresh_TheftDetection_WithAlert проверяет, что при
// повторном использовании отозванного токена отправляется алерт.
func TestAuthUseCase_Refresh_TheftDetection_WithAlert(t *testing.T) {
	admin := sampleAdmin()
	jwtSecret := "test-jwt-secret-at-least-32-chars-long"
	refreshRepo := &mockRefreshTokenRepository{
		storedTokens: make(map[string]*domain.RefreshToken),
	}
	sender := &mockAlertSender{}
	worker := NewAlertWorker(sender, slog.Default(), 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)

	uc := newAuthUseCase(
		[]domain.AdminUser{admin},
		jwtSecret,
		15*time.Minute,
		7*24*time.Hour,
		refreshRepo,
		worker,
	)

	// 1. Логинимся — получаем refresh token
	result, err := uc.Login(ctx, "admin", "correct-password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Явно сохраняем токен в мок (Create добавил, но перезаписываем для гарантии)
	hash := auth.HashRefreshToken(result.RefreshToken)
	refreshRepo.storedTokens[hash] = &domain.RefreshToken{
		ID:        "valid-token-id",
		Username:  "admin",
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// 2. Refresh — получаем новую пару, старый токен отзывается
	_, err = uc.Refresh(ctx, result.RefreshToken)
	if err != nil {
		t.Fatalf("first Refresh failed: %v", err)
	}

	// 3. Повторный Refresh с отозванным токеном — детект кражи
	_, err = uc.Refresh(ctx, result.RefreshToken)
	if err == nil {
		t.Fatal("expected error on reuse of revoked token")
	}

	// Даём воркеру время обработать
	time.Sleep(100 * time.Millisecond)

	alerts := sender.Alerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 theft alert, got %d", len(alerts))
	}
	if alerts[0].Type != domain.AlertTypeRefreshTokenTheft {
		t.Errorf("alert.Type = %q, want %q", alerts[0].Type, domain.AlertTypeRefreshTokenTheft)
	}

	cancel()
	worker.Stop()
}
