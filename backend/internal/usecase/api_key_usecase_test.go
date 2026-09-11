package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// --- Хелперы ---

// newTestAPIKeyUseCase создаёт APIKeyUseCase с тестовым репозиторием.
func newTestAPIKeyUseCase(repo *mockAPIKeyRepository) APIKeyUseCase {
	return NewAPIKeyUseCase(repo)
}

// validCreateParams возвращает валидные параметры создания.
func validCreateParams() domain.CreateAPIKeyParams {
	return domain.CreateAPIKeyParams{
		Name:        "CI/CD Pipeline",
		Description: "Автоматизация развёртывания",
		Role:        "admin",
		CreatedBy:   "threet",
	}
}

// generateAndStoreKey создаёт ключ через UC и возвращает связку (key, fullKey).
// Полезно для тестов Authenticate.
func generateAndStoreKey(t *testing.T, uc APIKeyUseCase, params domain.CreateAPIKeyParams) (*domain.APIKey, string) {
	t.Helper()
	key, fullKey, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	return key, fullKey
}

// --- Тесты Create ---

// TestAPIKeyUseCase_Create_EmptyName проверяет валидацию имени.
func TestAPIKeyUseCase_Create_EmptyName(t *testing.T) {
	uc := newTestAPIKeyUseCase(&mockAPIKeyRepository{})

	params := validCreateParams()
	params.Name = ""

	_, _, err := uc.Create(context.Background(), params)
	if err == nil {
		t.Fatal("Create should fail for empty name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestAPIKeyUseCase_Create_DefaultRole проверяет дефолтную роль admin.
func TestAPIKeyUseCase_Create_DefaultRole(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	params := validCreateParams()
	params.Role = "" // не задана

	key, _, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if key.Role != "admin" {
		t.Errorf("Role = %q, want %q (default)", key.Role, "admin")
	}
}

// TestAPIKeyUseCase_Create_Success проверяет полный цикл создания.
func TestAPIKeyUseCase_Create_Success(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	params := validCreateParams()
	key, fullKey, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Проверяем возвращённый ключ.
	if key == nil {
		t.Fatal("returned key is nil")
	}
	if fullKey == "" {
		t.Error("fullKey is empty")
	}
	if !strings.HasPrefix(fullKey, "emh_") {
		t.Errorf("fullKey %q does not have prefix emh_", fullKey)
	}

	// Проверяем, что ключ был передан в репозиторий.
	created := repo.lastCreated()
	if created == nil {
		t.Fatal("repo.Create was not called")
	}
	if created.Name != "CI/CD Pipeline" {
		t.Errorf("created.Name = %q", created.Name)
	}

	// Секрет должен быть захеширован, а не в plaintext.
	if created.SecretHash == "" {
		t.Error("SecretHash is empty")
	}
	// Хешированный секрет не должен совпадать с оригиналом из fullKey.
	_, secret, ok := auth.ParseAPIKey(fullKey)
	if !ok {
		t.Fatalf("cannot parse generated fullKey: %q", fullKey)
	}
	if created.SecretHash == secret {
		t.Error("SecretHash equals plaintext secret (not hashed)")
	}
	// Но хеш должен валидировать оригинальный секрет.
	if !auth.CheckPassword(created.SecretHash, secret) {
		t.Error("SecretHash does not validate the original secret")
	}
}

// TestAPIKeyUseCase_Create_UniqueKeys проверяет уникальность каждой генерации.
func TestAPIKeyUseCase_Create_UniqueKeys(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	fullKeys := make(map[string]bool)
	for i := 0; i < 10; i++ {
		_, fullKey, err := uc.Create(context.Background(), validCreateParams())
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
		if fullKeys[fullKey] {
			t.Errorf("duplicate fullKey generated: %q", fullKey)
		}
		fullKeys[fullKey] = true
	}
}

// --- Тесты List ---

// TestAPIKeyUseCase_List_DelegatesToRepo проверяет делегирование в репозиторий.
func TestAPIKeyUseCase_List_DelegatesToRepo(t *testing.T) {
	capturedIncludeRevoked := false
	repo := &mockAPIKeyRepository{
		listFunc: func(ctx context.Context, includeRevoked bool) ([]*domain.APIKey, error) {
			capturedIncludeRevoked = includeRevoked
			return []*domain.APIKey{
				{ID: "key-1", Name: "key-a"},
				{ID: "key-2", Name: "key-b"},
			}, nil
		},
	}
	uc := newTestAPIKeyUseCase(repo)

	keys, err := uc.List(context.Background(), true)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if !capturedIncludeRevoked {
		t.Error("includeRevoked=true was not passed to repo")
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

// TestAPIKeyUseCase_List_ExcludeRevoked проверяет, что флаг false передаётся в репозиторий.
func TestAPIKeyUseCase_List_ExcludeRevoked(t *testing.T) {
	capturedIncludeRevoked := true
	repo := &mockAPIKeyRepository{
		listFunc: func(ctx context.Context, includeRevoked bool) ([]*domain.APIKey, error) {
			capturedIncludeRevoked = includeRevoked
			return []*domain.APIKey{}, nil
		},
	}
	uc := newTestAPIKeyUseCase(repo)

	_, _ = uc.List(context.Background(), false)
	if capturedIncludeRevoked {
		t.Error("includeRevoked=false was not passed to repo")
	}
}

// --- Тесты Revoke ---

// TestAPIKeyUseCase_Revoke_DelegatesToRepo проверяет делегирование в репозиторий.
func TestAPIKeyUseCase_Revoke_DelegatesToRepo(t *testing.T) {
	var capturedID string
	repo := &mockAPIKeyRepository{
		revokeFunc: func(ctx context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	uc := newTestAPIKeyUseCase(repo)

	err := uc.Revoke(context.Background(), "key-to-revoke")
	if err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}
	if capturedID != "key-to-revoke" {
		t.Errorf("captured ID = %q, want %q", capturedID, "key-to-revoke")
	}
}

// --- Тесты Authenticate ---

// newStoredKey создаёт "сохранённый" ключ для тестов Authenticate.
// Принимает plaintext secret для возможности проверки хеша.
func newStoredKey(secret string) *domain.APIKey {
	hash, _ := auth.HashPassword(secret)
	return &domain.APIKey{
		ID:         "key-id-001",
		KeyID:      "test-key-id",
		SecretHash: hash,
		Name:       "test-key",
		Role:       "admin",
		CreatedAt:  time.Now(),
	}
}

// TestAPIKeyUseCase_Authenticate_Success проверяет полный цикл успешной аутентификации.
func TestAPIKeyUseCase_Authenticate_Success(t *testing.T) {
	// Сначала генерируем ключ, чтобы получить валидный fullKey.
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	_, fullKey, err := uc.Create(context.Background(), validCreateParams())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Настраиваем repo.GetByKeyID на возврат созданного ключа.
	keyID, _, ok := auth.ParseAPIKey(fullKey)
	if !ok {
		t.Fatalf("cannot parse generated fullKey: %q", fullKey)
	}

	created := repo.lastCreated()
	repo.getByKeyIDFunc = func(ctx context.Context, kid string) (*domain.APIKey, error) {
		if kid == keyID {
			return created, nil
		}
		return nil, domain.ErrNotFound
	}

	// Authenticate должен вернуть тот же ключ.
	got, err := uc.Authenticate(context.Background(), fullKey, "192.168.1.1")
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("got.ID = %q, want %q", got.ID, created.ID)
	}
	if got.Name != created.Name {
		t.Errorf("got.Name = %q, want %q", got.Name, created.Name)
	}
}

// TestAPIKeyUseCase_Authenticate_MalformedKey проверяет ошибку при неверном формате.
func TestAPIKeyUseCase_Authenticate_MalformedKey(t *testing.T) {
	uc := newTestAPIKeyUseCase(&mockAPIKeyRepository{})

	tests := []struct {
		name    string
		fullKey string
	}{
		{"no prefix", "wrongprefix_abc123_secret456"},
		{"empty", ""},
		{"only prefix", "emh_"},
		{"no separator", "emh_noseparator"},
		{"jwt token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Authenticate(context.Background(), tt.fullKey, "192.168.1.1")
			if err == nil {
				t.Errorf("expected error for malformed key %q", tt.fullKey)
			}
			if !strings.Contains(err.Error(), "malformed") {
				t.Errorf("expected 'malformed' in error, got %q", err.Error())
			}
		})
	}
}

// TestAPIKeyUseCase_Authenticate_KeyNotFound проверяет, что несуществующий ключ не раскрывает информацию.
func TestAPIKeyUseCase_Authenticate_KeyNotFound(t *testing.T) {
	uc := newTestAPIKeyUseCase(&mockAPIKeyRepository{})

	// Сгенерируем "незарегистрированный" ключ.
	keyID, secret, fullKey, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}
	_ = keyID
	_ = secret

	_, err = uc.Authenticate(context.Background(), fullKey, "192.168.1.1")
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
	// Сообщение должно быть общим (не "key not found", чтобы не раскрывать существование).
	if !strings.Contains(err.Error(), "invalid api key") {
		t.Errorf("expected generic 'invalid api key' error, got %q", err.Error())
	}
}

// TestAPIKeyUseCase_Authenticate_RevokedKey проверяет ошибку для отозванного ключа.
func TestAPIKeyUseCase_Authenticate_RevokedKey(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	// Генерируем ключ и помечаем его отозванным.
	keyID, secret, fullKey, _ := auth.GenerateAPIKey()
	storedKey := newStoredKey(secret)
	storedKey.KeyID = keyID
	now := time.Now()
	storedKey.RevokedAt = &now

	repo.getByKeyIDFunc = func(ctx context.Context, kid string) (*domain.APIKey, error) {
		if kid == keyID {
			return storedKey, nil
		}
		return nil, domain.ErrNotFound
	}

	_, err := uc.Authenticate(context.Background(), fullKey, "192.168.1.1")
	if err == nil {
		t.Fatal("expected error for revoked key")
	}
	if !strings.Contains(err.Error(), "revoked or expired") {
		t.Errorf("expected 'revoked or expired' error, got %q", err.Error())
	}
}

// TestAPIKeyUseCase_Authenticate_ExpiredKey проверяет ошибку для истёкшего ключа.
func TestAPIKeyUseCase_Authenticate_ExpiredKey(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	keyID, secret, fullKey, _ := auth.GenerateAPIKey()
	storedKey := newStoredKey(secret)
	storedKey.KeyID = keyID
	pastTime := time.Now().Add(-1 * time.Hour)
	storedKey.ExpiresAt = &pastTime

	repo.getByKeyIDFunc = func(ctx context.Context, kid string) (*domain.APIKey, error) {
		if kid == keyID {
			return storedKey, nil
		}
		return nil, domain.ErrNotFound
	}

	_, err := uc.Authenticate(context.Background(), fullKey, "192.168.1.1")
	if err == nil {
		t.Fatal("expected error for expired key")
	}
	if !strings.Contains(err.Error(), "revoked or expired") {
		t.Errorf("expected 'revoked or expired' error, got %q", err.Error())
	}
}

// TestAPIKeyUseCase_Authenticate_WrongSecret проверяет ошибку при неверном пароле.
func TestAPIKeyUseCase_Authenticate_WrongSecret(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	// Генерируем ключ, но подменяем secret в fullKey.
	keyID, _, _, _ := auth.GenerateAPIKey()
	storedKey := newStoredKey("correct-secret")
	storedKey.KeyID = keyID

	// Создаём fullKey с тем же keyID, но неверным secret.
	wrongFullKey := "emh_" + keyID + "_wrong_secret"

	repo.getByKeyIDFunc = func(ctx context.Context, kid string) (*domain.APIKey, error) {
		if kid == keyID {
			return storedKey, nil
		}
		return nil, domain.ErrNotFound
	}

	_, err := uc.Authenticate(context.Background(), wrongFullKey, "192.168.1.1")
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
	if !strings.Contains(err.Error(), "invalid api key") {
		t.Errorf("expected generic 'invalid api key' error, got %q", err.Error())
	}
}

// TestAPIKeyUseCase_Authenticate_UpdatesLastUsed проверяет аудит использования.
func TestAPIKeyUseCase_Authenticate_UpdatesLastUsed(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	_, fullKey, _ := uc.Create(context.Background(), validCreateParams())
	keyID, _, _ := auth.ParseAPIKey(fullKey)

	created := repo.lastCreated()
	repo.getByKeyIDFunc = func(ctx context.Context, kid string) (*domain.APIKey, error) {
		if kid == keyID {
			return created, nil
		}
		return nil, domain.ErrNotFound
	}

	clientIP := "203.0.113.42"
	_, err := uc.Authenticate(context.Background(), fullKey, clientIP)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	// Проверяем, что UpdateLastUsed был вызван.
	repo.mu.Lock()
	calls := repo.updateLastUsedCalls
	repo.mu.Unlock()

	if len(calls) != 1 {
		t.Fatalf("expected 1 UpdateLastUsed call, got %d", len(calls))
	}
	if calls[0].id != created.ID {
		t.Errorf("UpdateLastUsed id = %q, want %q", calls[0].id, created.ID)
	}
	if calls[0].ip != clientIP {
		t.Errorf("UpdateLastUsed ip = %q, want %q", calls[0].ip, clientIP)
	}
}

// TestAPIKeyUseCase_Authenticate_LastUsedThrottle проверяет троттлинг обновлений.
func TestAPIKeyUseCase_Authenticate_LastUsedThrottle(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	_, fullKey, _ := uc.Create(context.Background(), validCreateParams())
	keyID, _, _ := auth.ParseAPIKey(fullKey)

	created := repo.lastCreated()
	// Устанавливаем LastUsedAt в текущее время — следующая аутентификация
	// должна пропустить UpdateLastUsed (в пределах lastUsedThrottle = 1 минута).
	now := time.Now()
	created.LastUsedAt = &now

	repo.getByKeyIDFunc = func(ctx context.Context, kid string) (*domain.APIKey, error) {
		if kid == keyID {
			return created, nil
		}
		return nil, domain.ErrNotFound
	}

	_, err := uc.Authenticate(context.Background(), fullKey, "192.168.1.1")
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	// UpdateLastUsed НЕ должен быть вызван (в пределах throttle).
	repo.mu.Lock()
	calls := repo.updateLastUsedCalls
	repo.mu.Unlock()

	if len(calls) != 0 {
		t.Errorf("expected 0 UpdateLastUsed calls (throttled), got %d", len(calls))
	}
}

// TestAPIKeyUseCase_Authenticate_LastUsedExpiredTriggerUpdate проверяет, что
// после истечения throttle обновления вызывается.
func TestAPIKeyUseCase_Authenticate_LastUsedExpiredTriggerUpdate(t *testing.T) {
	repo := &mockAPIKeyRepository{}
	uc := newTestAPIKeyUseCase(repo)

	_, fullKey, _ := uc.Create(context.Background(), validCreateParams())
	keyID, _, _ := auth.ParseAPIKey(fullKey)

	created := repo.lastCreated()
	// Устанавливаем LastUsedAt в прошлое — за пределами throttle.
	pastTime := time.Now().Add(-2 * time.Minute) // lastUsedThrottle = 1 мин
	created.LastUsedAt = &pastTime

	repo.getByKeyIDFunc = func(ctx context.Context, kid string) (*domain.APIKey, error) {
		if kid == keyID {
			return created, nil
		}
		return nil, domain.ErrNotFound
	}

	_, err := uc.Authenticate(context.Background(), fullKey, "192.168.1.1")
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	// UpdateLastUsed должен быть вызван.
	repo.mu.Lock()
	calls := repo.updateLastUsedCalls
	repo.mu.Unlock()

	if len(calls) != 1 {
		t.Errorf("expected 1 UpdateLastUsed call (throttle expired), got %d", len(calls))
	}
}

// --- Тесты APIKey.IsActive ---

// TestAPIKey_IsActive проверяет логику активности ключа.
func TestAPIKey_IsActive(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-1 * time.Hour)
	futureTime := now.Add(1 * time.Hour)

	tests := []struct {
		name     string
		key      domain.APIKey
		expected bool
	}{
		{"active without expiry", domain.APIKey{}, true},
		{"active with future expiry", domain.APIKey{ExpiresAt: &futureTime}, true},
		{"expired", domain.APIKey{ExpiresAt: &pastTime}, false},
		{"revoked", domain.APIKey{RevokedAt: &pastTime}, false},
		{"revoked and expired", domain.APIKey{RevokedAt: &pastTime, ExpiresAt: &pastTime}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.IsActive(now); got != tt.expected {
				t.Errorf("IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}
