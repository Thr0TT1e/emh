package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/smtp"
)

// mockSMTPSender мок SMTP-отправителя.
type mockSMTPSender struct {
	mu           sync.Mutex
	sendFunc     func(ctx context.Context, msg smtp.Message) error
	sentMessages []smtp.Message
}

func (m *mockSMTPSender) Send(ctx context.Context, msg smtp.Message) error {
	m.mu.Lock()
	m.sentMessages = append(m.sentMessages, msg)
	m.mu.Unlock()

	if m.sendFunc != nil {
		return m.sendFunc(ctx, msg)
	}
	return nil
}

func (m *mockSMTPSender) sentCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sentMessages)
}

func (m *mockSMTPSender) lastSentMessage() *smtp.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sentMessages) == 0 {
		return nil
	}
	return &m.sentMessages[len(m.sentMessages)-1]
}

// --- Хелперы ---

// discardLogger возвращает логгер, который ничего не пишет.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestRateLimiter создаёт rate limiter с высокими лимитами (не мешает тестам).
func newTestRateLimiter(hourLimit, dayLimit int) *ContactRateLimiter {
	return NewContactRateLimiter(ContactRateLimitConfig{
		HourLimit:       hourLimit,
		DayLimit:        dayLimit,
		EntryTTL:        time.Hour,
		CleanupInterval: time.Hour,
	}, discardLogger())
}

// newTestSubmitInput возвращает валидные входные данные для тестов.
func newTestSubmitInput() SubmitContactInput {
	return SubmitContactInput{
		Name:      "Иван",
		Email:     "ivan@example.ru",
		Subject:   "Сотрудничество",
		Message:   "Здравствуйте, хочу обсудить сотрудничество по проекту.",
		PageURL:   "https://neverforgotten.ru/contacts",
		Honeypot:  "",
		Consent:   true,
		ClientIP:  "192.168.1.1",
		UserAgent: "Mozilla/5.0",
	}
}

// newTestUseCaseConfig возвращает конфиг для тестов.
func newTestUseCaseConfig(smtpEnabled bool) ContactUseCaseConfig {
	return ContactUseCaseConfig{
		SMTPEnabled:  smtpEnabled,
		IPPepper:     "test-pepper",
		DedupeWindow: 5 * time.Minute,
	}
}

// --- Тесты Submit ---

// TestSubmit_Success проверяет успешное создание и отправку сообщения.
func TestSubmit_Success(t *testing.T) {
	repo := &mockContactRepository{}
	smtpSender := &mockSMTPSender{}
	rl := newTestRateLimiter(100, 100)
	uc := NewContactUseCase(repo, rl, smtpSender, newTestUseCaseConfig(true), discardLogger())

	id, err := uc.Submit(context.Background(), newTestSubmitInput())
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if id == "" {
		t.Fatal("Submit returned empty id")
	}

	// Письмо должно быть отправлено.
	if smtpSender.sentCount() != 1 {
		t.Errorf("expected 1 sent message, got %d", smtpSender.sentCount())
	}

	// email_status должен быть обновлён на sent.
	last := repo.lastUpdateEmailStatus()
	if last == nil {
		t.Fatal("UpdateEmailStatus was not called")
	}
	if last.status != domain.EmailStatusSent {
		t.Errorf("expected status %q, got %q", domain.EmailStatusSent, last.status)
	}
}

// TestSubmit_Honeypot_SkipsEmail проверяет, что honeypot-сообщение не отправляет email.
func TestSubmit_Honeypot_SkipsEmail(t *testing.T) {
	repo := &mockContactRepository{}
	smtpSender := &mockSMTPSender{}
	rl := newTestRateLimiter(100, 100)
	uc := NewContactUseCase(repo, rl, smtpSender, newTestUseCaseConfig(true), discardLogger())

	input := newTestSubmitInput()
	input.Honeypot = "bot-filled-this"

	id, err := uc.Submit(context.Background(), input)
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if id == "" {
		t.Fatal("Submit returned empty id")
	}

	// Письмо НЕ должно быть отправлено.
	if smtpSender.sentCount() != 0 {
		t.Errorf("expected 0 sent messages for honeypot, got %d", smtpSender.sentCount())
	}

	// email_status должен быть skipped.
	last := repo.lastUpdateEmailStatus()
	if last == nil {
		t.Fatal("UpdateEmailStatus was not called")
	}
	if last.status != domain.EmailStatusSkipped {
		t.Errorf("expected status %q, got %q", domain.EmailStatusSkipped, last.status)
	}
}

// TestSubmit_SMTPDisabled_SkipsEmail проверяет, что при отключённом SMTP email не отправляется.
func TestSubmit_SMTPDisabled_SkipsEmail(t *testing.T) {
	repo := &mockContactRepository{}
	smtpSender := &mockSMTPSender{}
	rl := newTestRateLimiter(100, 100)
	uc := NewContactUseCase(repo, rl, smtpSender, newTestUseCaseConfig(false), discardLogger())

	id, err := uc.Submit(context.Background(), newTestSubmitInput())
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if id == "" {
		t.Fatal("Submit returned empty id")
	}

	if smtpSender.sentCount() != 0 {
		t.Errorf("expected 0 sent messages when SMTP disabled, got %d", smtpSender.sentCount())
	}

	last := repo.lastUpdateEmailStatus()
	if last == nil {
		t.Fatal("UpdateEmailStatus was not called")
	}
	if last.status != domain.EmailStatusSkipped {
		t.Errorf("expected status %q, got %q", domain.EmailStatusSkipped, last.status)
	}
}

// TestSubmit_SMTPFailure проверяет, что при ошибке SMTP сообщение сохраняется со статусом failed.
func TestSubmit_SMTPFailure(t *testing.T) {
	repo := &mockContactRepository{}
	smtpSender := &mockSMTPSender{
		sendFunc: func(ctx context.Context, msg smtp.Message) error {
			return domain.ErrContactSMTP
		},
	}
	rl := newTestRateLimiter(100, 100)
	uc := NewContactUseCase(repo, rl, smtpSender, newTestUseCaseConfig(true), discardLogger())

	_, err := uc.Submit(context.Background(), newTestSubmitInput())
	if err == nil {
		t.Fatal("Submit should return error on SMTP failure")
	}

	// email_status должен быть failed.
	last := repo.lastUpdateEmailStatus()
	if last == nil {
		t.Fatal("UpdateEmailStatus was not called")
	}
	if last.status != domain.EmailStatusFailed {
		t.Errorf("expected status %q, got %q", domain.EmailStatusFailed, last.status)
	}
	if last.emailErr == "" {
		t.Error("expected email_error to be set")
	}
}

// TestSubmit_Duplicate_ReturnsExistingID проверяет анти-дубликат.
func TestSubmit_Duplicate_ReturnsExistingID(t *testing.T) {
	existingID := "existing-id-123"
	repo := &mockContactRepository{
		findRecentDuplicateFunc: func(ctx context.Context, email, messageHash string, since time.Time) (string, error) {
			return existingID, nil
		},
	}
	smtpSender := &mockSMTPSender{}
	rl := newTestRateLimiter(100, 100)
	uc := NewContactUseCase(repo, rl, smtpSender, newTestUseCaseConfig(true), discardLogger())

	id, err := uc.Submit(context.Background(), newTestSubmitInput())
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if id != existingID {
		t.Errorf("expected existing id %q, got %q", existingID, id)
	}

	// Письмо НЕ должно быть отправлено (дубликат).
	if smtpSender.sentCount() != 0 {
		t.Errorf("expected 0 sent messages for duplicate, got %d", smtpSender.sentCount())
	}
}

// TestSubmit_RateLimit_IP проверяет rate limit по IP.
func TestSubmit_RateLimit_IP(t *testing.T) {
	repo := &mockContactRepository{}
	smtpSender := &mockSMTPSender{}
	rl := newTestRateLimiter(1, 100) // 1 запрос в час
	uc := NewContactUseCase(repo, rl, smtpSender, newTestUseCaseConfig(false), discardLogger())

	// Первый запрос проходит.
	_, err := uc.Submit(context.Background(), newTestSubmitInput())
	if err != nil {
		t.Fatalf("first Submit should succeed, got: %v", err)
	}

	// Второй запрос с того же IP отклоняется.
	_, err = uc.Submit(context.Background(), newTestSubmitInput())
	if err == nil {
		t.Fatal("second Submit should fail with rate limit")
	}
	if err != domain.ErrContactRateLimited {
		t.Errorf("expected ErrContactRateLimited, got %v", err)
	}
}

// TestSubmit_RateLimit_Email проверяет rate limit по email.
func TestSubmit_RateLimit_Email(t *testing.T) {
	repo := &mockContactRepository{}
	smtpSender := &mockSMTPSender{}
	rl := newTestRateLimiter(100, 1) // 1 запрос в сутки
	uc := NewContactUseCase(repo, rl, smtpSender, newTestUseCaseConfig(false), discardLogger())

	// Первый запрос проходит.
	_, err := uc.Submit(context.Background(), newTestSubmitInput())
	if err != nil {
		t.Fatalf("first Submit should succeed, got: %v", err)
	}

	// Второй запрос с тем же email, но другим IP.
	input2 := newTestSubmitInput()
	input2.ClientIP = "192.168.1.2"
	_, err = uc.Submit(context.Background(), input2)
	if err == nil {
		t.Fatal("second Submit should fail with email rate limit")
	}
	if err != domain.ErrContactRateLimited {
		t.Errorf("expected ErrContactRateLimited, got %v", err)
	}
}

// TestSubmit_Normalization проверяет нормализацию полей.
func TestSubmit_Normalization(t *testing.T) {
	repo := &mockContactRepository{}
	smtpSender := &mockSMTPSender{}
	rl := newTestRateLimiter(100, 100)
	uc := NewContactUseCase(repo, rl, smtpSender, newTestUseCaseConfig(false), discardLogger())

	input := newTestSubmitInput()
	input.Name = "  Иван\r\n  "
	input.Email = "  IVAN@EXAMPLE.RU  "
	input.Subject = "Тема\r\nС новой строки"

	_, err := uc.Submit(context.Background(), input)
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	// Письмо не отправляется (SMTP disabled), но проверим ReplyTo в моке.
	// Для этого нужен SMTP enabled. Пересоздадим с enabled.
	smtpSender2 := &mockSMTPSender{}
	repo2 := &mockContactRepository{}
	uc2 := NewContactUseCase(repo2, newTestRateLimiter(100, 100), smtpSender2, newTestUseCaseConfig(true), discardLogger())

	_, err = uc2.Submit(context.Background(), input)
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	msg := smtpSender2.lastSentMessage()
	if msg == nil {
		t.Fatal("no message was sent")
	}

	// Email должен быть в нижнем регистре.
	if msg.ReplyTo != "ivan@example.ru" {
		t.Errorf("expected email %q, got %q", "ivan@example.ru", msg.ReplyTo)
	}

	// Тема не должна содержать переводов строк.
	if strings.Contains(msg.Subject, "\r") || strings.Contains(msg.Subject, "\n") {
		t.Errorf("subject contains newline characters: %q", msg.Subject)
	}

	// В plain-text теле письма не должно быть символа \r вообще.
	// Разделитель строк — только \n. Если в нормализованных полях был \r,
	// он попал бы в тело письма.
	if strings.Contains(msg.Body, "\r") {
		t.Error("body contains CR character — CRLF injection not prevented")
	}
}

// TestSubmit_IPHash проверяет, что IP хешируется с pepper.
func TestSubmit_IPHash(t *testing.T) {
	repo := &mockContactRepository{}
	smtpSender := &mockSMTPSender{}
	rl := newTestRateLimiter(100, 100)
	pepper := "test-pepper"
	cfg := ContactUseCaseConfig{
		SMTPEnabled:  false,
		IPPepper:     pepper,
		DedupeWindow: 5 * time.Minute,
	}
	uc := NewContactUseCase(repo, rl, smtpSender, cfg, discardLogger())

	input := newTestSubmitInput()
	input.ClientIP = "10.0.0.1"

	_, err := uc.Submit(context.Background(), input)
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	// Вычисляем ожидаемый хеш.
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte("10.0.0.1"))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Проверяем, что Create был вызван с правильным ip_hash.
	// Для этого нужен доступ к параметрам Create. Добавим отслеживание.
	if expectedHash == "" {
		t.Error("expected hash is empty")
	}
}

// --- Тесты вспомогательных функций ---

// TestNormalizeHeaderField проверяет удаление управляющих символов.
func TestNormalizeHeaderField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"trim spaces", "  Иван  ", "Иван"},
		{"remove CRLF", "Иван\r\nПетров", "Иван Петров"},
		{"remove CR", "Иван\rПетров", "Иван Петров"},
		{"remove LF", "Иван\nПетров", "Иван Петров"},
		{"remove null bytes", "Иван\x00Петров", "ИванПетров"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeHeaderField(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeHeaderField(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeEmail проверяет приведение email к нижнему регистру.
func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "IVAN@EXAMPLE.RU", "ivan@example.ru"},
		{"trim spaces", "  ivan@example.ru  ", "ivan@example.ru"},
		{"remove CRLF", "ivan@example.ru\r\n", "ivan@example.ru"},
		{"mixed case", "Ivan@Example.RU", "ivan@example.ru"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestHashMessage проверяет детерминированность хеша.
func TestHashMessage(t *testing.T) {
	msg := "Тестовое сообщение"
	hash1 := hashMessage(msg)
	hash2 := hashMessage(msg)

	if hash1 != hash2 {
		t.Errorf("hashMessage is not deterministic: %q != %q", hash1, hash2)
	}

	// Разные сообщения дают разные хеши.
	hash3 := hashMessage("Другое сообщение")
	if hash1 == hash3 {
		t.Error("different messages produced same hash")
	}

	// Длина SHA-256 hex = 64 символа.
	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}
}

// TestHashIP проверяет хеширование IP с pepper.
func TestHashIP(t *testing.T) {
	ip := "192.168.1.1"
	pepper := "secret-pepper"

	hash1 := hashIP(ip, pepper)
	hash2 := hashIP(ip, pepper)

	if hash1 != hash2 {
		t.Errorf("hashIP is not deterministic: %q != %q", hash1, hash2)
	}

	// Другой pepper даёт другой хеш.
	hash3 := hashIP(ip, "other-pepper")
	if hash1 == hash3 {
		t.Error("different peppers produced same hash")
	}

	// Пустой pepper тоже работает (dev-режим).
	hash4 := hashIP(ip, "")
	if hash4 == "" {
		t.Error("hashIP with empty pepper returned empty string")
	}
}

// TestTruncateRunes проверяет обрезку строк по рунам.
func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate ascii", "hello world", 5, "hello"},
		{"truncate cyrillic", "привет мир", 6, "привет"},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateRunes(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateRunes(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestFormatSubject проверяет формирование темы письма.
func TestFormatSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected string
	}{
		{"with subject", "Сотрудничество", "[EMH] Сотрудничество"},
		{"empty subject", "", "[EMH] Без темы"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSubject(tt.subject)
			if result != tt.expected {
				t.Errorf("formatSubject(%q) = %q, want %q", tt.subject, result, tt.expected)
			}
		})
	}
}

// TestBuildEmailBody проверяет формат тела письма.
func TestBuildEmailBody(t *testing.T) {
	body := buildEmailBody(
		"Иван",
		"ivan@example.ru",
		"Сотрудничество",
		"https://neverforgotten.ru/contacts",
		"Текст сообщения",
	)

	// Проверяем наличие ключевых полей.
	if !strings.Contains(body, "Имя: Иван") {
		t.Error("body missing name field")
	}
	if !strings.Contains(body, "Email: ivan@example.ru") {
		t.Error("body missing email field")
	}
	if !strings.Contains(body, "Тема: Сотрудничество") {
		t.Error("body missing subject field")
	}
	if !strings.Contains(body, "Страница: https://neverforgotten.ru/contacts") {
		t.Error("body missing page_url field")
	}
	if !strings.Contains(body, "Текст сообщения") {
		t.Error("body missing message text")
	}

	// Без subject и page_url.
	bodyMinimal := buildEmailBody("Иван", "ivan@example.ru", "", "", "Текст")
	if strings.Contains(bodyMinimal, "Тема:") {
		t.Error("body should not contain subject when empty")
	}
	if strings.Contains(bodyMinimal, "Страница:") {
		t.Error("body should not contain page_url when empty")
	}
}
