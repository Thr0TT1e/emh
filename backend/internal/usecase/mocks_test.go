package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// ============================================================
// Общие моки для всех тестов пакета usecase
// ============================================================

// ─── mockHeroRepository ─────────────────────────────────────────────────────

// mockHeroRepository мок репозитория героев.
type mockHeroRepository struct {
	createFunc   func(ctx context.Context, p domain.CreateHeroParams) (string, error)
	updateFunc   func(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error)
	deleteFunc   func(ctx context.Context, id string, hardDelete bool) error
	getByIDFunc  func(ctx context.Context, id string) (*domain.Hero, error)
	listFunc     func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error)
	getByIDCalls atomic.Int32 // для тестов HeroQueryUseCase
}

func (m *mockHeroRepository) Create(ctx context.Context, p domain.CreateHeroParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "new-hero-id", nil
}

func (m *mockHeroRepository) Update(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Hero{ID: *p.ID}, nil
}

func (m *mockHeroRepository) Delete(ctx context.Context, id string, hardDelete bool) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, hardDelete)
	}
	return nil
}

func (m *mockHeroRepository) GetByID(ctx context.Context, id string) (*domain.Hero, error) {
	m.getByIDCalls.Add(1)
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Hero{ID: id}, nil
}

func (m *mockHeroRepository) List(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return nil, "", 0, nil
}

// Остальные методы HeroRepository (заглушки)
func (m *mockHeroRepository) ListByConflict(ctx context.Context, conflictID string) ([]*domain.Hero, error) {
	return nil, nil
}
func (m *mockHeroRepository) ListByLocation(ctx context.Context, locationID string) ([]*domain.Hero, error) {
	return nil, nil
}
func (m *mockHeroRepository) ListByAward(ctx context.Context, awardID string) ([]*domain.Hero, error) {
	return nil, nil
}
func (m *mockHeroRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Hero, error) {
	return nil, nil
}
func (m *mockHeroRepository) Count(ctx context.Context) (int64, error) { return 0, nil }

// ─── mockConflictRepository ─────────────────────────────────────────────────

// mockConflictRepository мок репозитория конфликтов.
type mockConflictRepository struct {
	createFn             func(ctx context.Context, p domain.CreateConflictParams) (string, error)
	updateFn             func(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error)
	deleteFn             func(ctx context.Context, id string) error
	getByIDFn            func(ctx context.Context, id string) (*domain.Conflict, error)
	listFn               func(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error)
	hasHeroesFn          func(ctx context.Context, id string) (bool, error)
	hasCyclicReferenceFn func(ctx context.Context, id, parentID string) (bool, error)
}

func (m *mockConflictRepository) Create(ctx context.Context, p domain.CreateConflictParams) (string, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return "new-conflict-id", nil
}

func (m *mockConflictRepository) Update(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return &domain.Conflict{ID: p.ID}, nil
}

func (m *mockConflictRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockConflictRepository) GetByID(ctx context.Context, id string) (*domain.Conflict, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockConflictRepository) List(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, f)
	}
	return nil, "", 0, nil
}

func (m *mockConflictRepository) HasHeroes(ctx context.Context, id string) (bool, error) {
	if m.hasHeroesFn != nil {
		return m.hasHeroesFn(ctx, id)
	}
	return false, nil
}

func (m *mockConflictRepository) HasCyclicReference(ctx context.Context, id, parentID string) (bool, error) {
	if m.hasCyclicReferenceFn != nil {
		return m.hasCyclicReferenceFn(ctx, id, parentID)
	}
	return false, nil
}

// ─── mockLLMProvider ────────────────────────────────────────────────────────

// mockLLMProvider мок LLM-провайдера.
type mockLLMProvider struct {
	nameFn     func() string
	typeFn     func() string
	modelFn    func() string
	generateFn func(ctx context.Context, prompt string) (string, error)
}

func (m *mockLLMProvider) Name() string {
	if m.nameFn != nil {
		return m.nameFn()
	}
	return "mock_provider"
}

func (m *mockLLMProvider) Type() string {
	if m.typeFn != nil {
		return m.typeFn()
	}
	return "mock"
}

func (m *mockLLMProvider) Model() string {
	if m.modelFn != nil {
		return m.modelFn()
	}
	return "mock-model"
}

func (m *mockLLMProvider) Generate(ctx context.Context, prompt string) (string, error) {
	if m.generateFn != nil {
		return m.generateFn(ctx, prompt)
	}
	return "", nil
}

func (m *mockLLMProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// ─── mockLLMExtractionLogRepo ───────────────────────────────────────────────

// mockLLMExtractionLogRepo мок репозитория логов экстракции.
type mockLLMExtractionLogRepo struct {
	createFn          func(ctx context.Context, log *domain.LLMExtractionLog) (string, error)
	getByIDFn         func(ctx context.Context, id string) (*domain.LLMExtractionLog, error)
	listFn            func(ctx context.Context, limit, offset int) ([]*domain.LLMExtractionLog, error)
	deleteOlderThanFn func(ctx context.Context, olderThan time.Time) (int64, error)
}

func (m *mockLLMExtractionLogRepo) Create(ctx context.Context, log *domain.LLMExtractionLog) (string, error) {
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return "log-id", nil
}

func (m *mockLLMExtractionLogRepo) GetByID(ctx context.Context, id string) (*domain.LLMExtractionLog, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockLLMExtractionLogRepo) List(ctx context.Context, limit, offset int) ([]*domain.LLMExtractionLog, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}
	return nil, nil
}

func (m *mockLLMExtractionLogRepo) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	if m.deleteOlderThanFn != nil {
		return m.deleteOlderThanFn(ctx, olderThan)
	}
	return 0, nil
}

// ─── mockLLMProviderRepo ────────────────────────────────────────────────────

// mockLLMProviderRepo мок репозитория провайдеров.
type mockLLMProviderRepo struct {
	createFn    func(ctx context.Context, record *domain.LLMProviderRecord) (string, error)
	getByIDFn   func(ctx context.Context, id string) (*domain.LLMProviderRecord, error)
	getByNameFn func(ctx context.Context, name string) (*domain.LLMProviderRecord, error)
	getActiveFn func(ctx context.Context) (*domain.LLMProviderRecord, error)
	listFn      func(ctx context.Context) ([]*domain.LLMProviderRecord, error)
	updateFn    func(ctx context.Context, record *domain.LLMProviderRecord) error
	setActiveFn func(ctx context.Context, id string) error
}

func (m *mockLLMProviderRepo) Create(ctx context.Context, record *domain.LLMProviderRecord) (string, error) {
	if m.createFn != nil {
		return m.createFn(ctx, record)
	}
	return "", nil
}

func (m *mockLLMProviderRepo) GetByID(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockLLMProviderRepo) GetByName(ctx context.Context, name string) (*domain.LLMProviderRecord, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(ctx, name)
	}
	return nil, domain.ErrNotFound
}

func (m *mockLLMProviderRepo) GetActive(ctx context.Context) (*domain.LLMProviderRecord, error) {
	if m.getActiveFn != nil {
		return m.getActiveFn(ctx)
	}
	return nil, domain.ErrNotFound
}

func (m *mockLLMProviderRepo) List(ctx context.Context) ([]*domain.LLMProviderRecord, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockLLMProviderRepo) Update(ctx context.Context, record *domain.LLMProviderRecord) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, record)
	}
	return nil
}

func (m *mockLLMProviderRepo) SetActive(ctx context.Context, id string) error {
	if m.setActiveFn != nil {
		return m.setActiveFn(ctx, id)
	}
	return nil
}

// ─── mockLLMProviderManager ─────────────────────────────────────────────────

// mockLLMProviderManager мок менеджера провайдеров.
type mockLLMProviderManager struct {
	getActiveFn         func(ctx context.Context) (domain.LLMProvider, error)
	getProviderByNameFn func(name string) (domain.LLMProvider, error)
	listProvidersFn     func() []domain.LLMProvider
}

func (m *mockLLMProviderManager) GetActiveProvider(ctx context.Context) (domain.LLMProvider, error) {
	if m.getActiveFn != nil {
		return m.getActiveFn(ctx)
	}
	return nil, domain.ErrNotFound
}

func (m *mockLLMProviderManager) GetProviderByName(name string) (domain.LLMProvider, error) {
	if m.getProviderByNameFn != nil {
		return m.getProviderByNameFn(name)
	}
	return nil, domain.ErrNotFound
}

func (m *mockLLMProviderManager) ListProviders() []domain.LLMProvider {
	if m.listProvidersFn != nil {
		return m.listProvidersFn()
	}
	return nil
}

// ─── mockAPIKeyRepository ─────────────────────────────────────────────────

// mockAPIKeyRepository мок репозитория API-ключей.
type mockAPIKeyRepository struct {
	mu                  sync.Mutex
	createFunc          func(ctx context.Context, key *domain.APIKey) error
	listFunc            func(ctx context.Context, includeRevoked bool) ([]*domain.APIKey, error)
	revokeFunc          func(ctx context.Context, id string) error
	getByKeyIDFunc      func(ctx context.Context, keyID string) (*domain.APIKey, error)
	updateLastUsedFunc  func(ctx context.Context, id string, ip string, t time.Time) error
	createCalls         []*domain.APIKey
	updateLastUsedCalls []updateLastUsedCall
}

type updateLastUsedCall struct {
	id string
	ip string
	t  time.Time
}

func (m *mockAPIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	m.mu.Lock()
	m.createCalls = append(m.createCalls, key)
	m.mu.Unlock()

	if m.createFunc != nil {
		return m.createFunc(ctx, key)
	}
	// Имитация присвоения ID и меток времени.
	key.ID = "api-key-id-001"
	key.CreatedAt = time.Now()
	key.UpdatedAt = time.Now()
	return nil
}

func (m *mockAPIKeyRepository) List(ctx context.Context, includeRevoked bool) ([]*domain.APIKey, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, includeRevoked)
	}
	return []*domain.APIKey{}, nil
}

func (m *mockAPIKeyRepository) Revoke(ctx context.Context, id string) error {
	if m.revokeFunc != nil {
		return m.revokeFunc(ctx, id)
	}
	return nil
}

func (m *mockAPIKeyRepository) GetByKeyID(ctx context.Context, keyID string) (*domain.APIKey, error) {
	if m.getByKeyIDFunc != nil {
		return m.getByKeyIDFunc(ctx, keyID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockAPIKeyRepository) UpdateLastUsed(ctx context.Context, id string, ip string, t time.Time) error {
	m.mu.Lock()
	m.updateLastUsedCalls = append(m.updateLastUsedCalls, updateLastUsedCall{id: id, ip: ip, t: t})
	m.mu.Unlock()

	if m.updateLastUsedFunc != nil {
		return m.updateLastUsedFunc(ctx, id, ip, t)
	}
	return nil
}

func (m *mockAPIKeyRepository) lastCreated() *domain.APIKey {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.createCalls) == 0 {
		return nil
	}
	return m.createCalls[len(m.createCalls)-1]
}

// ============================================================
// Моки для RefreshTokenRepository
// ============================================================
type mockRefreshTokenRepository struct {
	mu                            sync.Mutex
	createdTokens                 []domain.CreateRefreshTokenParams
	storedTokens                  map[string]*domain.RefreshToken
	revokedTokens                 map[string]*domain.RefreshToken // отозванные токены
	revokeCalls                   []string
	createFunc                    func(ctx context.Context, p domain.CreateRefreshTokenParams) (string, error)
	getByHashFunc                 func(ctx context.Context, hash string) (*domain.RefreshToken, error)
	revokeFunc                    func(ctx context.Context, id string) error
	deleteExpiredFunc             func(ctx context.Context) (int64, error)
	getByHashIncludingRevokedFunc func(ctx context.Context, hash string) (*domain.RefreshToken, error)
	revokeAllForUserFunc          func(ctx context.Context, username string) (int64, error)
}

func (m *mockRefreshTokenRepository) Create(ctx context.Context, p domain.CreateRefreshTokenParams) (string, error) {
	m.mu.Lock()
	m.createdTokens = append(m.createdTokens, p)
	m.mu.Unlock()
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "refresh-token-id", nil
}

func (m *mockRefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	if m.getByHashFunc != nil {
		return m.getByHashFunc(ctx, hash)
	}
	if token, ok := m.storedTokens[hash]; ok {
		return token, nil
	}
	return nil, errors.New("token not found")
}

// Revoke отзывает токен по ID.
// Дефолтное поведение: перемещает токен из storedTokens в revokedTokens,
// симулируя real-world UPDATE + soft-delete в БД.
// Сначала вызывает revokeFunc (если задан); при ошибке перемещение не выполняется.
func (m *mockRefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	// Сначала кастомная функция — если вернёт ошибку, мок не мутирует состояние
	// (в реальной БД SQL-транзакция откатится).
	if m.revokeFunc != nil {
		if err := m.revokeFunc(ctx, id); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.revokeCalls = append(m.revokeCalls, id)

	// Дефолтное поведение: перемещаем токен из активных в отозванные
	if m.storedTokens != nil {
		for hash, token := range m.storedTokens {
			if token.ID == id {
				if m.revokedTokens == nil {
					m.revokedTokens = make(map[string]*domain.RefreshToken)
				}
				m.revokedTokens[hash] = token
				delete(m.storedTokens, hash)
				break
			}
		}
	}

	m.mu.Unlock()
	return nil
}

func (m *mockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	if m.deleteExpiredFunc != nil {
		return m.deleteExpiredFunc(ctx)
	}
	return 0, nil
}

// GetByHashIncludingRevoked ищет среди отозванных токенов (для детекта кражи).
func (m *mockRefreshTokenRepository) GetByHashIncludingRevoked(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	if m.getByHashIncludingRevokedFunc != nil {
		return m.getByHashIncludingRevokedFunc(ctx, hash)
	}
	if m.revokedTokens != nil {
		if token, ok := m.revokedTokens[hash]; ok {
			return token, nil
		}
	}
	return nil, errors.New("refresh token not found")
}

// RevokeAllForUser отзывает все активные токены пользователя (при детекте кражи).
func (m *mockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, username string) (int64, error) {
	if m.revokeAllForUserFunc != nil {
		return m.revokeAllForUserFunc(ctx, username)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var count int64
	if m.storedTokens != nil {
		for hash, token := range m.storedTokens {
			if token.Username == username {
				if m.revokedTokens == nil {
					m.revokedTokens = make(map[string]*domain.RefreshToken)
				}
				m.revokedTokens[hash] = token
				delete(m.storedTokens, hash)
				count++
			}
		}
	}
	return count, nil
}

// --- contact_usecase_test.go ---

// updateEmailStatusCall зафиксированный вызов UpdateEmailStatus.
type updateEmailStatusCall struct {
	id       string
	status   domain.EmailStatus
	emailErr string
}

// mockContactRepository мок репозитория сообщений обратной связи.
type mockContactRepository struct {
	mu                      sync.Mutex
	createFunc              func(ctx context.Context, p domain.CreateContactMessageParams) (string, error)
	updateEmailStatusFunc   func(ctx context.Context, id string, status domain.EmailStatus, emailErr string) error
	findRecentDuplicateFunc func(ctx context.Context, email, messageHash string, since time.Time) (string, error)
	updateEmailStatusCalls  []updateEmailStatusCall
}

func (m *mockContactRepository) Create(ctx context.Context, p domain.CreateContactMessageParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "test-id-001", nil
}

func (m *mockContactRepository) UpdateEmailStatus(ctx context.Context, id string, status domain.EmailStatus, emailErr string) error {
	m.mu.Lock()
	m.updateEmailStatusCalls = append(m.updateEmailStatusCalls, updateEmailStatusCall{
		id:       id,
		status:   status,
		emailErr: emailErr,
	})
	m.mu.Unlock()

	if m.updateEmailStatusFunc != nil {
		return m.updateEmailStatusFunc(ctx, id, status, emailErr)
	}
	return nil
}

func (m *mockContactRepository) FindRecentDuplicate(ctx context.Context, email, messageHash string, since time.Time) (string, error) {
	if m.findRecentDuplicateFunc != nil {
		return m.findRecentDuplicateFunc(ctx, email, messageHash, since)
	}
	return "", nil
}

// lastUpdateEmailStatus возвращает последний вызов UpdateEmailStatus.
func (m *mockContactRepository) lastUpdateEmailStatus() *updateEmailStatusCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.updateEmailStatusCalls) == 0 {
		return nil
	}
	return &m.updateEmailStatusCalls[len(m.updateEmailStatusCalls)-1]
}

// ============================================================
// Моки для недостающих репозиториев HeroQueryUseCase
// mockHeroRepository и mockPhotoRepository уже определены
// в hero_usecase_test.go и photo_usecase_test.go.
// ============================================================

// --- mockHeroAwardRepository ---

type mockHeroAwardRepository struct {
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroAward, error)
	callCount      atomic.Int32
}

func (m *mockHeroAwardRepository) Add(ctx context.Context, heroID, awardID string, awardDate *string, decreeNumber string) error {
	return nil
}

func (m *mockHeroAwardRepository) Remove(ctx context.Context, heroID, awardID string) error {
	return nil
}

func (m *mockHeroAwardRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroAward, error) {
	m.callCount.Add(1)
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroAward{}, nil
}

// --- mockHeroConflictRepository ---

type mockHeroConflictRepository struct {
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroConflict, error)
	callCount      atomic.Int32
}

func (m *mockHeroConflictRepository) Add(ctx context.Context, heroID, conflictID, specificLocation, rankAtConflict string) error {
	return nil
}

func (m *mockHeroConflictRepository) Remove(ctx context.Context, heroID, conflictID string) error {
	return nil
}

func (m *mockHeroConflictRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroConflict, error) {
	m.callCount.Add(1)
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroConflict{}, nil
}

// --- mockHeroLocationRepository ---

type mockHeroLocationRepository struct {
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroLocation, error)
	callCount      atomic.Int32
}

func (m *mockHeroLocationRepository) Add(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error {
	return nil
}

func (m *mockHeroLocationRepository) Remove(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error {
	return nil
}

func (m *mockHeroLocationRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroLocation, error) {
	m.callCount.Add(1)
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroLocation{}, nil
}

// --- mockHeroSourceRepository ---

type mockHeroSourceRepository struct {
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroSource, error)
	callCount      atomic.Int32
}

func (m *mockHeroSourceRepository) Add(ctx context.Context, p domain.AddHeroSourceParams) (string, error) {
	return "source-id-mock", nil
}

func (m *mockHeroSourceRepository) Remove(ctx context.Context, heroID, sourceID string) error {
	return nil
}

func (m *mockHeroSourceRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroSource, error) {
	m.callCount.Add(1)
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroSource{}, nil
}

// --- mockHeroRelationRepository ---

type mockHeroRelationRepository struct {
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroRelation, error)
	callCount      atomic.Int32
}

func (m *mockHeroRelationRepository) Add(ctx context.Context, p domain.AddHeroRelationParams) (string, error) {
	return "relation-id-mock", nil
}

func (m *mockHeroRelationRepository) Remove(ctx context.Context, id string) error {
	return nil
}

func (m *mockHeroRelationRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroRelation, error) {
	m.callCount.Add(1)
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroRelation{}, nil
}

// --- location_usecase_test.go ---

// mockLocationRepository мок репозитория локаций.
type mockLocationRepository struct {
	createFunc             func(ctx context.Context, p domain.CreateLocationParams) (string, error)
	updateFunc             func(ctx context.Context, p domain.UpdateLocationParams) (*domain.Location, error)
	deleteFunc             func(ctx context.Context, id string) error
	getByIDFunc            func(ctx context.Context, id string) (*domain.Location, error)
	listFunc               func(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error)
	reparentFunc           func(ctx context.Context, oldParentID, newParentID string) (int, error)
	hasHeroesFunc          func(ctx context.Context, locationID string) (bool, error)
	hasCyclicReferenceFunc func(ctx context.Context, selfID, candidateParentID string) (bool, error)
}

func (m *mockLocationRepository) Create(ctx context.Context, p domain.CreateLocationParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "location-id-001", nil
}

func (m *mockLocationRepository) Update(ctx context.Context, p domain.UpdateLocationParams) (*domain.Location, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Location{ID: p.ID}, nil
}

func (m *mockLocationRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockLocationRepository) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Location{ID: id, Name: "Москва"}, nil
}

func (m *mockLocationRepository) List(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Location{}, "", 0, nil
}

func (m *mockLocationRepository) Reparent(ctx context.Context, oldParentID, newParentID string) (int, error) {
	if m.reparentFunc != nil {
		return m.reparentFunc(ctx, oldParentID, newParentID)
	}
	return 0, nil
}

func (m *mockLocationRepository) HasHeroes(ctx context.Context, locationID string) (bool, error) {
	if m.hasHeroesFunc != nil {
		return m.hasHeroesFunc(ctx, locationID)
	}
	return false, nil
}

func (m *mockLocationRepository) HasCyclicReference(ctx context.Context, selfID, candidateParentID string) (bool, error) {
	if m.hasCyclicReferenceFunc != nil {
		return m.hasCyclicReferenceFunc(ctx, selfID, candidateParentID)
	}
	return false, nil
}

// --- mockAwardRepository ---
type mockAwardRepository struct {
	createFunc  func(ctx context.Context, p domain.CreateAwardParams) (string, error)
	updateFunc  func(ctx context.Context, p domain.UpdateAwardParams) (*domain.Award, error)
	deleteFunc  func(ctx context.Context, id string) error
	getByIDFunc func(ctx context.Context, id string) (*domain.Award, error)
	listFunc    func(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error)
	hasHeroesFn func(ctx context.Context, awardID string) (bool, error)
}

func (m *mockAwardRepository) Create(ctx context.Context, p domain.CreateAwardParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "award-id-001", nil
}

func (m *mockAwardRepository) Update(ctx context.Context, p domain.UpdateAwardParams) (*domain.Award, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Award{ID: p.ID}, nil
}

func (m *mockAwardRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockAwardRepository) GetByID(ctx context.Context, id string) (*domain.Award, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Award{ID: id, Name: "Герой России"}, nil
}

func (m *mockAwardRepository) List(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Award{}, "", 0, nil
}

func (m *mockAwardRepository) HasHeroes(ctx context.Context, awardID string) (bool, error) {
	if m.hasHeroesFn != nil {
		return m.hasHeroesFn(ctx, awardID)
	}
	return false, nil
}

// --- media_usecase_test.go ---

// newMediaUseCase создаёт MediaUseCase с моками.
func newMediaUseCase(storage *mockMediaStorage) MediaUseCase {
	return NewMediaUseCase(storage, 15*time.Minute)
}

// --- photo_usecase_test.go ---

// mockPhotoRepository мок репозитория фотографий.
type mockPhotoRepository struct {
	mu              sync.Mutex
	addFunc         func(ctx context.Context, p domain.AddPhotoParams) (string, error)
	batchAddFunc    func(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error)
	deleteFunc      func(ctx context.Context, heroID, photoID string) (string, error)
	deleteByIDsFunc func(ctx context.Context, heroID string, photoIDs []string) ([]string, error)
	listByHeroFunc  func(ctx context.Context, heroID string) ([]*domain.Photo, error)
	reorderFunc     func(ctx context.Context, heroID string, photoIDs []string) error
	setMainFunc     func(ctx context.Context, heroID, photoID string) error
	updateFunc      func(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error)
	addCalls        []domain.AddPhotoParams
	enqueueCalls    []ThumbnailTask
	listByHeroCalls atomic.Int32 // для тестов HeroQueryUseCase
}

func (m *mockPhotoRepository) Add(ctx context.Context, p domain.AddPhotoParams) (string, error) {
	m.mu.Lock()
	m.addCalls = append(m.addCalls, p)
	m.mu.Unlock()

	if m.addFunc != nil {
		return m.addFunc(ctx, p)
	}
	return "photo-id-001", nil
}

func (m *mockPhotoRepository) BatchAdd(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error) {
	if m.batchAddFunc != nil {
		return m.batchAddFunc(ctx, heroID, photos)
	}
	result := make([]*domain.Photo, len(photos))
	for i, p := range photos {
		result[i] = &domain.Photo{ID: "photo-" + string(rune('a'+i)), URL: p.URL, IsMain: p.IsMain}
	}
	return result, nil
}

func (m *mockPhotoRepository) Delete(ctx context.Context, heroID, photoID string) (string, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, heroID, photoID)
	}
	return "https://s3.example.com/bucket/photo.jpg", nil
}

func (m *mockPhotoRepository) DeleteByIDs(ctx context.Context, heroID string, photoIDs []string) ([]string, error) {
	if m.deleteByIDsFunc != nil {
		return m.deleteByIDsFunc(ctx, heroID, photoIDs)
	}
	urls := make([]string, len(photoIDs))
	for i := range urls {
		urls[i] = "https://s3.example.com/bucket/photo" + string(rune('a'+i)) + ".jpg"
	}
	return urls, nil
}

func (m *mockPhotoRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.Photo, error) {
	m.listByHeroCalls.Add(1)
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.Photo{}, nil
}

func (m *mockPhotoRepository) Reorder(ctx context.Context, heroID string, photoIDs []string) error {
	if m.reorderFunc != nil {
		return m.reorderFunc(ctx, heroID, photoIDs)
	}
	return nil
}

func (m *mockPhotoRepository) GetMainPhotoURL(ctx context.Context, heroID string) (string, error) {
	return "", nil
}

func (m *mockPhotoRepository) SetMain(ctx context.Context, heroID, photoID string) error {
	if m.setMainFunc != nil {
		return m.setMainFunc(ctx, heroID, photoID)
	}
	return nil
}

func (m *mockPhotoRepository) Update(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Photo{ID: "photo-id", HeroID: p.HeroID, URL: "https://example.com/photo.jpg"}, nil
}

func (m *mockPhotoRepository) UpdateAfterProcessing(ctx context.Context, photoID, url, thumbnailURL string) error {
	return nil
}

func (m *mockPhotoRepository) ListWithoutThumbnails(ctx context.Context) ([]*domain.Photo, error) {
	return []*domain.Photo{}, nil
}

func (m *mockPhotoRepository) ListAllMediaURLs(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

// ListWithoutThumbnailsPaged — заглушка для удовлетворения интерфейса.
// StartupBackfill тестируется отдельно.
func (m *mockPhotoRepository) ListWithoutThumbnailsPaged(ctx context.Context, cursor string, limit int) ([]*domain.Photo, string, error) {
	// Возвращаем пустой результат — по умолчанию нет фото без thumbnail
	return nil, "", nil
}

// mockMediaStorage мок S3-хранилища.
type mockMediaStorage struct {
	mu             sync.Mutex
	deleteFunc     func(ctx context.Context, key string) error
	keyFromURLFunc func(rawURL string) (string, error)
	deleteCalls    []string
	presignError   error // для тестов MediaUseCase
}

func (m *mockMediaStorage) PresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	if m.presignError != nil {
		return "", m.presignError
	}
	return "https://presigned.example.com/" + key, nil
}

func (m *mockMediaStorage) PublicURL(key string) string {
	return "https://public.example.com/" + key
}

// Delete потокобезопасно записывает вызов.
// Реальный MediaStorage вызывается параллельно из deleteFilesFromStorage через errgroup.
func (m *mockMediaStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	m.deleteCalls = append(m.deleteCalls, key)
	m.mu.Unlock()

	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return nil
}

func (m *mockMediaStorage) KeyFromPublicURL(rawURL string) (string, error) {
	if m.keyFromURLFunc != nil {
		return m.keyFromURLFunc(rawURL)
	}
	return "extracted-key", nil
}

func (m *mockMediaStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockMediaStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	return nil
}

func (m *mockMediaStorage) ListObjects(ctx context.Context) ([]domain.ObjectInfo, error) {
	return []domain.ObjectInfo{}, nil
}

func (m *mockMediaStorage) EnsureBucket(ctx context.Context) error {
	return nil
}

// deleteCount потокобезопасно возвращает количество удалённых ключей.
func (m *mockMediaStorage) deleteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.deleteCalls)
}

func (m *mockMediaStorage) ListObjectsPaged(ctx context.Context, marker string, limit int) ([]domain.ObjectInfo, string, error) {
	// Для тестов возвращаем все объекты как одну страницу
	objects, err := m.ListObjects(ctx)
	if err != nil {
		return nil, "", err
	}
	return objects, "", nil
}

// mockThumbnailWorker мок воркера генерации превью.
type mockThumbnailWorker struct {
	mu       sync.Mutex
	enqueued []ThumbnailTask
}

func (m *mockThumbnailWorker) Enqueue(task ThumbnailTask) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enqueued = append(m.enqueued, task)
}

func (m *mockThumbnailWorker) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.enqueued)
}

// ============================================================
// submission_usecase_test.go
// ============================================================

// mockSubmissionRepository мок репозитория заявок.
type mockSubmissionRepository struct {
	mu                  sync.Mutex
	created             []domain.CreateSubmissionParams // для legacy-тестов
	createCalls         []domain.CreateSubmissionParams // для новых тестов
	createFunc          func(ctx context.Context, p domain.CreateSubmissionParams) (string, error)
	getByIDFunc         func(ctx context.Context, id string) (*domain.Submission, error)
	listFunc            func(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error)
	reviewFunc          func(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error)
	reviewed            []domain.ReviewSubmissionParams
	findByContentHashFn func(ctx context.Context, hash string) (*domain.Submission, error)
	createReviewFn      func(ctx context.Context, p domain.CreateSubmissionReviewParams) (string, error)
	listReviewsFn       func(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error)
	createReviewCalls   []domain.CreateSubmissionReviewParams
}

func (m *mockSubmissionRepository) Create(ctx context.Context, p domain.CreateSubmissionParams) (string, error) {
	m.mu.Lock()
	// Новое поле — для проверок с mutex (тесты анти-дубликата)
	m.createCalls = append(m.createCalls, p)
	// Старое поле — для обратной совместимости с legacy-тестами
	m.created = append(m.created, p)
	m.mu.Unlock()
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}

	return "submission-id-001", nil
}

func (m *mockSubmissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Submission{ID: id, Status: domain.StatusDraft}, nil
}

func (m *mockSubmissionRepository) List(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Submission{}, "", 0, nil
}

func (m *mockSubmissionRepository) Review(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error) {
	m.reviewed = append(m.reviewed, p)
	if m.reviewFunc != nil {
		return m.reviewFunc(ctx, p)
	}
	return &domain.Submission{ID: p.ID, Status: domain.StatusPublished}, nil
}

func (m *mockSubmissionRepository) FindByContentHash(ctx context.Context, hash string) (*domain.Submission, error) {
	if m.findByContentHashFn != nil {
		return m.findByContentHashFn(ctx, hash)
	}
	return nil, domain.ErrNotFound
}

func (m *mockSubmissionRepository) CreateReview(ctx context.Context, p domain.CreateSubmissionReviewParams) (string, error) {
	m.mu.Lock()
	m.createReviewCalls = append(m.createReviewCalls, p)
	m.mu.Unlock()
	if m.createReviewFn != nil {
		return m.createReviewFn(ctx, p)
	}
	return "review-id-001", nil
}

func (m *mockSubmissionRepository) ListReviews(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error) {
	if m.listReviewsFn != nil {
		return m.listReviewsFn(ctx, submissionID)
	}
	return []*domain.SubmissionReview{}, nil
}

// ============================================================
// thumbnail_worker_test.go
// ============================================================

// mockMediaStorageForWorker реализует domain.MediaStorage для тестов воркера.
// Позволяет подменять данные, возвращаемые при скачивании.
type mockMediaStorageForWorker struct {
	downloadData []byte
	downloadErr  error
	deleteCalled bool
	uploadCalls  []mockUploadCall
	uploadErr    error
}

// mockUploadCall сохраняет аргументы вызова Upload для проверок в тестах.
type mockUploadCall struct {
	Key         string
	ContentType string
	DataSize    int
}

func (m *mockMediaStorageForWorker) PresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	return "https://example.com/presigned/" + key, nil
}

func (m *mockMediaStorageForWorker) PublicURL(key string) string {
	return "https://example.com/" + key
}

func (m *mockMediaStorageForWorker) Delete(ctx context.Context, key string) error {
	m.deleteCalled = true
	return nil
}

func (m *mockMediaStorageForWorker) KeyFromPublicURL(rawURL string) (string, error) {
	return "photos/test.jpg", nil
}

func (m *mockMediaStorageForWorker) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	return io.NopCloser(bytes.NewReader(m.downloadData)), nil
}

func (m *mockMediaStorageForWorker) ListObjects(ctx context.Context) ([]domain.ObjectInfo, error) {
	return nil, nil
}

func (m *mockMediaStorageForWorker) EnsureBucket(ctx context.Context) error {
	return nil
}

func (m *mockMediaStorageForWorker) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	if m.uploadErr != nil {
		return m.uploadErr
	}
	// Считываем данные для проверки размера
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, data); err != nil {
		return err
	}
	m.uploadCalls = append(m.uploadCalls, mockUploadCall{
		Key:         key,
		ContentType: contentType,
		DataSize:    buf.Len(),
	})
	return nil
}

// ListObjectsPaged — заглушка для удовлетворения интерфейса domain.MediaStorage.
// Thumbnail worker не использует пагинацию S3 (работает только с ListWithoutThumbnails из БД).
func (m *mockMediaStorageForWorker) ListObjectsPaged(ctx context.Context, marker string, limit int) ([]domain.ObjectInfo, string, error) {
	return nil, "", nil
}

// mockPhotoRepositoryForWorker реализует repository.PhotoRepository для тестов воркера.
// Отслеживает, был ли вызван UpdateAfterProcessing (признак успешного завершения).
// Встраиваем интерфейс-заглушку, чтобы не реализовывать все методы вручную.
type mockPhotoRepositoryForWorker struct {
	repository.PhotoRepository  // nil-заглушка: непереопределённые методы не вызываются в тестах
	updateAfterProcessingCalled bool
}

func (m *mockPhotoRepositoryForWorker) UpdateAfterProcessing(ctx context.Context, photoID, originalURL, thumbnailURL string) error {
	m.updateAfterProcessingCalled = true
	return nil
} // ============================================================
// auth_audit_worker_test.go
// ============================================================

// mockAuthAuditRepository мок репозитория аудита аутентификации.
// Воркер пишет по одной записи через Add (не batch).
type mockAuthAuditRepository struct {
	mu                sync.Mutex
	addFn             func(ctx context.Context, entry domain.AuthAuditEntry) error
	entries           []domain.AuthAuditEntry
	deleteOlderThanFn func(ctx context.Context, olderThan time.Time) (int64, error)
}

func (m *mockAuthAuditRepository) Add(ctx context.Context, entry domain.AuthAuditEntry) error {
	m.mu.Lock()
	m.entries = append(m.entries, entry)
	m.mu.Unlock()
	if m.addFn != nil {
		return m.addFn(ctx, entry)
	}
	return nil
}

func (m *mockAuthAuditRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	if m.deleteOlderThanFn != nil {
		return m.deleteOlderThanFn(ctx, olderThan)
	}
	return 0, nil
}

// totalEntries возвращает общее число записанных записей.
func (m *mockAuthAuditRepository) totalEntries() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

// lastEntry возвращает последнюю записанную запись.
func (m *mockAuthAuditRepository) lastEntry() *domain.AuthAuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) == 0 {
		return nil
	}
	e := m.entries[len(m.entries)-1]
	return &e
}

// ============================================================
// orphan_cleanup_worker_test.go — дополнение mockPhotoRepository
// ============================================================

// mockPhotoRepositoryForOrphan специализированный мок для orphan cleanup.
type mockPhotoRepositoryForOrphan struct {
	listAllMediaURLsFn func(ctx context.Context) ([]string, error)
}

func (m *mockPhotoRepositoryForOrphan) ListAllMediaURLs(ctx context.Context) ([]string, error) {
	if m.listAllMediaURLsFn != nil {
		return m.listAllMediaURLsFn(ctx)
	}
	return nil, nil
}

// Остальные методы PhotoRepository — заглушки для удовлетворения интерфейса.
func (m *mockPhotoRepositoryForOrphan) Add(ctx context.Context, p domain.AddPhotoParams) (string, error) {
	return "", nil
}
func (m *mockPhotoRepositoryForOrphan) BatchAdd(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error) {
	return nil, nil
}
func (m *mockPhotoRepositoryForOrphan) Delete(ctx context.Context, heroID, photoID string) (string, error) {
	return "", nil
}
func (m *mockPhotoRepositoryForOrphan) DeleteByIDs(ctx context.Context, heroID string, photoIDs []string) ([]string, error) {
	return nil, nil
}
func (m *mockPhotoRepositoryForOrphan) ListByHero(ctx context.Context, heroID string) ([]*domain.Photo, error) {
	return nil, nil
}
func (m *mockPhotoRepositoryForOrphan) ListByHeroPaged(ctx context.Context, heroID string, cursor string, limit int) ([]*domain.Photo, string, int64, error) {
	return nil, "", 0, nil
}
func (m *mockPhotoRepositoryForOrphan) Reorder(ctx context.Context, heroID string, photoIDs []string) error {
	return nil
}
func (m *mockPhotoRepositoryForOrphan) SetMain(ctx context.Context, heroID, photoID string) error {
	return nil
}
func (m *mockPhotoRepositoryForOrphan) Update(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error) {
	return nil, nil
}
func (m *mockPhotoRepositoryForOrphan) UpdateAfterProcessing(ctx context.Context, photoID, url, thumbnailURL string) error {
	return nil
}
func (m *mockPhotoRepositoryForOrphan) ListWithoutThumbnails(ctx context.Context) ([]*domain.Photo, error) {
	return nil, nil
}
func (m *mockPhotoRepositoryForOrphan) ListWithoutThumbnailsPaged(ctx context.Context, cursor string, limit int) ([]*domain.Photo, string, error) {
	return nil, "", nil
}
func (m *mockPhotoRepositoryForOrphan) GetMainPhotoURL(ctx context.Context, heroID string) (string, error) {
	return "", nil
}

// mockAlertSender — мок AlertSender для проверки отправки алертов.
type mockAlertSender struct {
	mu     sync.Mutex
	alerts []domain.Alert
}

func (m *mockAlertSender) SendAlert(_ context.Context, alert domain.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = append(m.alerts, alert)
	return nil
}

func (m *mockAlertSender) Alerts() []domain.Alert {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]domain.Alert, len(m.alerts))
	copy(cp, m.alerts)
	return cp
}
