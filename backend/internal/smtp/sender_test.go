package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

// ============================================================
// Моки
// ============================================================

// mockConn мок сетевого соединения для тестирования без реального соединения.
type mockConn struct {
	closed bool
}

func (m *mockConn) Read(b []byte) (n int, err error)  { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error) { return len(b), nil }
func (m *mockConn) Close() error {
	m.closed = true
	return nil
}
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// newTestSender создаёт SMTP-отправитель с моком диалера для тестирования.
func newTestSender(dialer func(ctx context.Context, network, addr string) (net.Conn, error)) *smtpSender {
	return &smtpSender{
		cfg: Config{
			Host:       "smtp.test.com",
			Port:       587,
			Username:   "user@test.com",
			Password:   "password",
			FromName:   "Test Sender",
			FromEmail:  "sender@test.com",
			To:         "recipient@test.com",
			Timeout:    10 * time.Second,
			MaxRetries: 3,
			BaseDelay:  5 * time.Second,
			MaxDelay:   60 * time.Second,
		},
		logger: nil, // nil логгер для тестов
	}
}

// ============================================================
// Тесты isRetryableSMTPError
// ============================================================

// TestIsRetryableSMTPError_Nil проверяет, что nil не повторяется.
func TestIsRetryableSMTPError_Nil(t *testing.T) {
	if isRetryableSMTPError(nil) {
		t.Error("nil error should not be retryable")
	}
}

// TestIsRetryableSMTPError_Temporary4xx проверяет, что временные ошибки 421, 450-452 повторяются.
func TestIsRetryableSMTPError_Temporary4xx(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"421 Service not available", 421},
		{"450 Mailbox unavailable", 450},
		{"451 Local error in processing", 451},
		{"452 Insufficient system storage", 452},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &textproto.Error{Code: tt.code, Msg: "test"}
			if !isRetryableSMTPError(err) {
				t.Errorf("code %d should be retryable", tt.code)
			}
		})
	}
}

// TestIsRetryableSMTPError_Permanent5xx проверяет, что постоянные ошибки 500-504, 535, 550-554 НЕ повторяются.
func TestIsRetryableSMTPError_Permanent5xx(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"500 Syntax error", 500},
		{"501 Syntax error in parameters", 501},
		{"502 Command not implemented", 502},
		{"503 Bad sequence of commands", 503},
		{"504 Command parameter not implemented", 504},
		{"535 Authentication failed", 535},
		{"550 Mailbox unavailable", 550},
		{"551 User not local", 551},
		{"552 Exceeded storage allocation", 552},
		{"553 Mailbox name not allowed", 553},
		{"554 Transaction failed", 554},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &textproto.Error{Code: tt.code, Msg: "test"}
			if isRetryableSMTPError(err) {
				t.Errorf("code %d should NOT be retryable", tt.code)
			}
		})
	}
}

// TestIsRetryableSMTPError_ContextCanceled проверяет, что отмена контекста НЕ повторяется.
func TestIsRetryableSMTPError_ContextCanceled(t *testing.T) {
	if isRetryableSMTPError(context.Canceled) {
		t.Error("context.Canceled should not be retryable")
	}
}

// TestIsRetryableSMTPError_ContextDeadlineExceeded проверяет, что таймаут контекста НЕ повторяется.
func TestIsRetryableSMTPError_ContextDeadlineExceeded(t *testing.T) {
	if isRetryableSMTPError(context.DeadlineExceeded) {
		t.Error("context.DeadlineExceeded should not be retryable")
	}
}

// TestIsRetryableSMTPError_NetworkErrors проверяет, что сетевые ошибки повторяются.
func TestIsRetryableSMTPError_NetworkErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"connection reset", errors.New("connection reset by peer")},
		{"broken pipe", errors.New("broken pipe")},
		{"timeout", errors.New("timeout")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !isRetryableSMTPError(tt.err) {
				t.Errorf("network error %q should be retryable", tt.name)
			}
		})
	}
}

// TestIsRetryableSMTPError_TLSError проверяет, что TLS ошибки повторяются.
func TestIsRetryableSMTPError_TLSError(t *testing.T) {
	err := &tls.CertificateVerificationError{}
	if !isRetryableSMTPError(err) {
		t.Error("TLS certificate verification error should be retryable")
	}
}

// TestIsRetryableSMTPError_UnknownError проверяет, что неизвестные ошибки НЕ повторяются.
func TestIsRetryableSMTPError_UnknownError(t *testing.T) {
	err := errors.New("some unknown error")
	if isRetryableSMTPError(err) {
		t.Error("unknown error should not be retryable")
	}
}

// ============================================================
// Тесты backoffDelay
// ============================================================

// TestBackoffDelay_ExponentialGrowth проверяет экспоненциальный рост задержки.
func TestBackoffDelay_ExponentialGrowth(t *testing.T) {
	base := 5 * time.Second
	max := 60 * time.Second

	tests := []struct {
		name      string
		attempt   int
		expected  time.Duration
		minJitter float64 // -25%
		maxJitter float64 // +25%
	}{
		{"attempt 0", 0, 5 * time.Second, -0.25, 0.25},
		{"attempt 1", 1, 10 * time.Second, -0.25, 0.25},
		{"attempt 2", 2, 20 * time.Second, -0.25, 0.25},
		{"attempt 3", 3, 40 * time.Second, -0.25, 0.25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := backoffDelay(tt.attempt, base, max)
			minDelay := time.Duration(float64(tt.expected) * (1 + tt.minJitter))
			maxDelay := time.Duration(float64(tt.expected) * (1 + tt.maxJitter))
			if delay < minDelay || delay > maxDelay {
				t.Errorf("backoffDelay(%d) = %v, want between %v and %v",
					tt.attempt, delay, minDelay, maxDelay)
			}
		})
	}
}

// TestBackoffDelay_MaxCap проверяет, что задержка не превышает максимум.
func TestBackoffDelay_MaxCap(t *testing.T) {
	base := 5 * time.Second
	max := 60 * time.Second

	tests := []struct {
		name    string
		attempt int
	}{
		{"attempt 5", 5},   // 5 * 32 = 160s > 60s
		{"attempt 10", 10}, // 5 * 1024 = 5120s > 60s
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := backoffDelay(tt.attempt, base, max)
			if delay > max {
				t.Errorf("backoffDelay(%d) = %v, should not exceed max %v", tt.attempt, delay, max)
			}
		})
	}
}

// TestBackoffDelay_ZeroBase проверяет поведение с нулевой базой.
func TestBackoffDelay_ZeroBase(t *testing.T) {
	delay := backoffDelay(0, 0, time.Minute)
	if delay != 0 {
		t.Errorf("backoffDelay with zero base should be 0, got %v", delay)
	}
}

// ============================================================
// Тесты Send с моком
// ============================================================

// TestSend_ContextCanceled проверяет, что отмена контекста прерывает отправку.
func TestSend_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу

	sender := newTestSender(nil)
	msg := Message{
		ReplyTo: "reply@test.com",
		Subject: "Test Subject",
		Body:    "Test body",
	}

	err := sender.Send(ctx, msg)
	if err == nil {
		t.Error("Send should fail with canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ============================================================
// Тесты sanitizeHeader (дополнение)
// ============================================================

// TestSanitizeHeader_CRLFInjection проверяет защиту от CRLF injection.
func TestSanitizeHeader_CRLFInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"clean string", "normal-value", "normal-value"},
		{"CRLF injection", "value\r\nBcc: attacker@evil.com", "value Bcc: attacker@evil.com"},
		{"CR only", "value\rinjection", "value injection"},
		{"LF only", "value\ninjection", "value injection"},
		{"multiple CRLF", "a\r\nb\r\nc", "a b c"},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeHeader(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeHeader(%q) = %q, want %q", tt.input, result, tt.expected)
			}
			// Убеждаемся, что в результате нет переводов строк.
			if strings.Contains(result, "\r") || strings.Contains(result, "\n") {
				t.Errorf("sanitizeHeader result contains newline: %q", result)
			}
		})
	}
}

// ============================================================
// Тесты encodeSubject (дополнение)
// ============================================================

// TestEncodeSubject_Empty проверяет кодирование пустой строки.
func TestEncodeSubject_Empty(t *testing.T) {
	if encodeSubject("") != "" {
		t.Error("encodeSubject(\"\") should return empty string")
	}
}

// TestEncodeSubject_ASCII проверяет, что ASCII строка не кодируется.
func TestEncodeSubject_ASCII(t *testing.T) {
	subject := "Test Subject"
	encoded := encodeSubject(subject)
	if encoded != subject {
		t.Errorf("encodeSubject(%q) should return same string for ASCII, got %q", subject, encoded)
	}
}

// ============================================================
// Тесты formatAddress (дополнение)
// ============================================================

// TestFormatAddress_EmptyName проверяет форматирование без имени.
func TestFormatAddress_EmptyName(t *testing.T) {
	result := formatAddress("", "no-reply@example.ru")
	expected := "no-reply@example.ru"
	if result != expected {
		t.Errorf("formatAddress() = %q, want %q", result, expected)
	}
}

// ============================================================
// Тесты generateMessageID (дополнение)
// ============================================================

// TestGenerateMessageID_Unique проверяет уникальность генерации.
func TestGenerateMessageID_Unique(t *testing.T) {
	msgID1 := generateMessageID("no-reply@example.ru")
	msgID2 := generateMessageID("no-reply@example.ru")
	if msgID1 == msgID2 {
		t.Error("generateMessageID should produce unique IDs")
	}
}

// ============================================================
// Тесты buildLetter (дополнение)
// ============================================================

// TestBuildLetter_NoReplyTo проверяет письмо без Reply-To.
func TestBuildLetter_NoReplyTo(t *testing.T) {
	sender := &smtpSender{
		cfg: Config{
			FromName:  "Test",
			FromEmail: "test@example.ru",
			To:        "to@example.ru",
		},
	}
	msg := Message{
		ReplyTo: "",
		Subject: "Test",
		Body:    "Body",
	}
	letter := sender.buildLetter(msg)
	if strings.Contains(letter, "Reply-To:") {
		t.Error("letter should not have Reply-To header when ReplyTo is empty")
	}
}

// TestEncodeSubject_Cyrillic проверяет RFC 2047 кодирование кириллицы.
func TestEncodeSubject_Cyrillic(t *testing.T) {
	subject := "Сотрудничество"
	encoded := encodeSubject(subject)

	// Результат не должен содержать исходную кириллицу (закодировано).
	if encoded == subject {
		t.Error("encodeSubject should encode cyrillic subject")
	}

	// Результат должен быть ASCII-safe для email-заголовков.
	for _, r := range encoded {
		if r > 127 {
			t.Errorf("encodeSubject result contains non-ASCII character: %q", string(r))
			break
		}
	}

	// Пустая тема возвращает пустую строку.
	if encodeSubject("") != "" {
		t.Error("encodeSubject(\"\") should return empty string")
	}
}

// TestFormatAddress проверяет форматирование адреса.
func TestFormatAddress(t *testing.T) {
	tests := []struct {
		name     string
		fromName string
		email    string
		expected string
	}{
		{"with name", "Вечная память героям", "no-reply@example.ru", "Вечная память героям <no-reply@example.ru>"},
		{"without name", "", "no-reply@example.ru", "no-reply@example.ru"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAddress(tt.fromName, tt.email)
			if result != tt.expected {
				t.Errorf("formatAddress(%q, %q) = %q, want %q", tt.fromName, tt.email, result, tt.expected)
			}
		})
	}
}

// TestGenerateMessageID проверяет формат Message-ID.
func TestGenerateMessageID(t *testing.T) {
	// С доменом из email.
	msgID := generateMessageID("no-reply@example.ru")
	if !strings.HasPrefix(msgID, "<") || !strings.HasSuffix(msgID, ">") {
		t.Errorf("Message-ID should be wrapped in angle brackets: %q", msgID)
	}
	if !strings.Contains(msgID, "@example.ru") {
		t.Errorf("Message-ID should contain domain: %q", msgID)
	}

	// Без домена (пустой email).
	msgIDFallback := generateMessageID("")
	if !strings.Contains(msgIDFallback, "@localhost") {
		t.Errorf("Message-ID should fallback to localhost: %q", msgIDFallback)
	}

	// Уникальность: два вызова дают разные ID.
	msgID2 := generateMessageID("no-reply@example.ru")
	if msgID == msgID2 {
		t.Error("generateMessageID should produce unique IDs")
	}
}

// TestBuildLetter_Headers проверяет структуру письма.
func TestBuildLetter_Headers(t *testing.T) {
	sender := &smtpSender{
		cfg: Config{
			FromName:  "Вечная память героям",
			FromEmail: "no-reply@neverforgotten.ru",
			To:        "info@neverforgotten.ru",
		},
	}

	msg := Message{
		ReplyTo: "ivan@example.ru",
		Subject: "[EMH] Сотрудничество",
		Body:    "Текст сообщения",
	}

	letter := sender.buildLetter(msg)

	// Проверяем обязательные заголовки.
	if !strings.Contains(letter, "From: Вечная память героям <no-reply@neverforgotten.ru>") {
		t.Error("letter missing From header")
	}
	if !strings.Contains(letter, "To: info@neverforgotten.ru") {
		t.Error("letter missing To header")
	}
	if !strings.Contains(letter, "Reply-To: ivan@example.ru") {
		t.Error("letter missing Reply-To header")
	}
	if !strings.Contains(letter, "Subject:") {
		t.Error("letter missing Subject header")
	}
	if !strings.Contains(letter, "Date:") {
		t.Error("letter missing Date header")
	}
	if !strings.Contains(letter, "Message-ID:") {
		t.Error("letter missing Message-ID header")
	}
	if !strings.Contains(letter, "MIME-Version: 1.0") {
		t.Error("letter missing MIME-Version header")
	}
	if !strings.Contains(letter, "Content-Type: text/plain; charset=\"UTF-8\"") {
		t.Error("letter missing Content-Type header")
	}

	// Тело письма после пустой строки.
	parts := strings.SplitN(letter, "\r\n\r\n", 2)
	if len(parts) != 2 {
		t.Fatal("letter should have headers and body separated by empty line")
	}
	if !strings.Contains(parts[1], "Текст сообщения") {
		t.Error("letter body missing message text")
	}
}

// TestBuildLetter_CRLFInjectionInSubject проверяет, что CRLF в теме не попадает в заголовки.
func TestBuildLetter_CRLFInjectionInSubject(t *testing.T) {
	sender := &smtpSender{
		cfg: Config{
			FromName:  "Test",
			FromEmail: "test@example.ru",
			To:        "to@example.ru",
		},
	}

	msg := Message{
		ReplyTo: "user@example.ru",
		Subject: "Тема\r\nBcc: attacker@evil.com",
		Body:    "Текст",
	}

	letter := sender.buildLetter(msg)

	// Bcc не должен появиться как отдельный заголовок.
	lines := strings.Split(letter, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Bcc:") {
			t.Errorf("CRLF injection succeeded, found Bcc header: %q", line)
		}
	}
}
