//go:build integration

package pg

import (
	"context"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg/testutil"
)

// TestMain запускается один раз на пакет. Здесь очищаем БД.
// Тесты в пакете pg изолированы через CleanupAllTables.
func TestMain(m *testing.M) {
	// m.Run() запускается из testutil через init-паттерн ниже.
	// Для простоты используем стандартный m.Run(), а cleanup в каждом тесте.
	code := m.Run()
	if code != 0 {
		return
	}
}

// newContactRepo создает тестовый репозиторий с общим пулом.
// Cleanup выполняется в t.Cleanup() автоматически.
func newContactRepo(t *testing.T) (*contactMessageRepository, func()) {
	t.Helper()
	pool, cleanupPool := testutil.NewTestPool(t)

	// Очищаем таблицу перед каждым тестом.
	testutil.CleanupTable(t, pool, "contact_messages")

	repo := NewContactMessageRepository(pool)

	cleanup := func() {
		testutil.CleanupTable(t, pool, "contact_messages")
		cleanupPool()
	}

	return repo.(*contactMessageRepository), cleanup
}

// newContactParams возвращает валидные параметры для Create.
func newContactParams() domain.CreateContactMessageParams {
	return domain.CreateContactMessageParams{
		Name:        "Иван Тестовый",
		Email:       testutil.RandomEmail(),
		Subject:     "Тестовая тема",
		Message:     "Тестовое сообщение длиной больше 10 символов",
		MessageHash: "abc123def456",
		PageURL:     "https://neverforgotten.ru/contacts",
		IPHash:      "hashed-ip-value",
		UserAgent:   "Mozilla/5.0 (Test Browser)",
		Consent:     true,
		IsHoneypot:  false,
	}
}

// TestContactRepo_Create_Success проверяет успешное создание записи.
func TestContactRepo_Create_Success(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	params := newContactParams()
	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Fatal("Create returned empty id")
	}

	// Проверяем, что id — валидный UUID.
	if len(id) != 36 {
		t.Errorf("id should be UUID (36 chars), got %d chars", len(id))
	}
}

// TestContactRepo_Create_EmptyOptionalFields проверяет, что необязательные поля могут быть пустыми.
func TestContactRepo_Create_EmptyOptionalFields(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	params := newContactParams()
	params.Subject = ""
	params.PageURL = ""
	params.IPHash = ""
	params.UserAgent = ""

	id, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create with empty optional fields failed: %v", err)
	}
	if id == "" {
		t.Fatal("Create returned empty id")
	}
}

// TestContactRepo_UpdateEmailStatus_Sent проверяет обновление статуса на sent.
func TestContactRepo_UpdateEmailStatus_Sent(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	id, err := repo.Create(context.Background(), newContactParams())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.UpdateEmailStatus(context.Background(), id, domain.EmailStatusSent, "")
	if err != nil {
		t.Fatalf("UpdateEmailStatus(sent) failed: %v", err)
	}
}

// TestContactRepo_UpdateEmailStatus_Failed проверяет сохранение ошибки SMTP.
func TestContactRepo_UpdateEmailStatus_Failed(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	id, err := repo.Create(context.Background(), newContactParams())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.UpdateEmailStatus(context.Background(), id, domain.EmailStatusFailed, "dial tcp: timeout")
	if err != nil {
		t.Fatalf("UpdateEmailStatus(failed) failed: %v", err)
	}
}

// TestContactRepo_UpdateEmailStatus_NotFound проверяет, что обновление несуществующей записи возвращает ErrNotFound.
func TestContactRepo_UpdateEmailStatus_NotFound(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	err := repo.UpdateEmailStatus(context.Background(), testutil.NewUUID(), domain.EmailStatusSent, "")
	if err == nil {
		t.Fatal("UpdateEmailStatus should fail for non-existent id")
	}
	// TODO: после добавления errors.Is(err, domain.ErrNotFound) в репозиторий —
	// проверить конкретную ошибку. Сейчас репозиторий оборачивает ErrNotFound через fmt.Errorf.
}

// TestContactRepo_FindRecentDuplicate_Found проверяет поиск дубликата.
func TestContactRepo_FindRecentDuplicate_Found(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	email := testutil.RandomEmail()
	hash := "same-hash-12345"

	params := newContactParams()
	params.Email = email
	params.MessageHash = hash

	existingID, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Ищем дубликат за последние 5 минут — должен найти.
	since := time.Now().Add(-5 * time.Minute)
	foundID, err := repo.FindRecentDuplicate(context.Background(), email, hash, since)
	if err != nil {
		t.Fatalf("FindRecentDuplicate failed: %v", err)
	}
	if foundID != existingID {
		t.Errorf("expected existing id %q, got %q", existingID, foundID)
	}
}

// TestContactRepo_FindRecentDuplicate_NotFound проверяет, что дубликат не находится за пределами окна.
func TestContactRepo_FindRecentDuplicate_NotFound(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	email := testutil.RandomEmail()
	hash := "unique-hash-99999"

	params := newContactParams()
	params.Email = email
	params.MessageHash = hash

	_, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Ищем с since = будущее — не должно найти.
	since := time.Now().Add(1 * time.Hour)
	foundID, err := repo.FindRecentDuplicate(context.Background(), email, hash, since)
	if err != nil {
		t.Fatalf("FindRecentDuplicate failed: %v", err)
	}
	if foundID != "" {
		t.Errorf("expected empty id for future window, got %q", foundID)
	}
}

// TestContactRepo_FindRecentDuplicate_DifferentEmail проверяет, что дубликат не совпадает по email.
func TestContactRepo_FindRecentDuplicate_DifferentEmail(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	hash := "same-hash-77777"

	params1 := newContactParams()
	params1.MessageHash = hash
	if _, err := repo.Create(context.Background(), params1); err != nil {
		t.Fatalf("Create 1 failed: %v", err)
	}

	// Ищем с тем же хешем, но другим email.
	since := time.Now().Add(-5 * time.Minute)
	otherEmail := testutil.RandomEmail()
	foundID, err := repo.FindRecentDuplicate(context.Background(), otherEmail, hash, since)
	if err != nil {
		t.Fatalf("FindRecentDuplicate failed: %v", err)
	}
	if foundID != "" {
		t.Errorf("expected empty id for different email, got %q", foundID)
	}
}

// TestContactRepo_Create_MultipleRecords проверяет, что несколько записей создаются без конфликтов.
func TestContactRepo_Create_MultipleRecords(t *testing.T) {
	repo, cleanup := newContactRepo(t)
	defer cleanup()

	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id, err := repo.Create(context.Background(), newContactParams())
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
		if ids[id] {
			t.Errorf("duplicate id generated: %q", id)
		}
		ids[id] = true
	}

	if len(ids) != 10 {
		t.Errorf("expected 10 unique ids, got %d", len(ids))
	}
}
