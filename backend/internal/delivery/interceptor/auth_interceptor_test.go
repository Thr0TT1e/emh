package interceptor

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// testIPPepper — тестовый секрет для хеширования IP в аудите.
const testIPPepper = "test-ip-pepper-16-chars-min"

// --- Моки ---

// mockAPIKeyAuthenticator мок API-key аутентификатора.
type mockAPIKeyAuthenticator struct {
	mu       sync.Mutex
	authFunc func(ctx context.Context, fullKey string, ip string) (*domain.APIKey, error)
	calls    []apiKeyAuthCall
}

type apiKeyAuthCall struct {
	fullKey string
	ip      string
}

func (m *mockAPIKeyAuthenticator) Authenticate(ctx context.Context, fullKey, ip string) (*domain.APIKey, error) {
	m.mu.Lock()
	m.calls = append(m.calls, apiKeyAuthCall{fullKey: fullKey, ip: ip})
	m.mu.Unlock()

	if m.authFunc != nil {
		return m.authFunc(ctx, fullKey, ip)
	}
	return nil, nil
}

// mockAuthAuditor мок аудитора с возможностью перехвата через enqueueFunc.
type mockAuthAuditor struct {
	mu          sync.Mutex
	entries     []domain.AuthAuditEntry
	enqueueFunc func(entry domain.AuthAuditEntry)
}

func (m *mockAuthAuditor) Enqueue(entry domain.AuthAuditEntry) {
	if m.enqueueFunc != nil {
		m.enqueueFunc(entry)
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
}

func (m *mockAuthAuditor) lastEntry() *domain.AuthAuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) == 0 {
		return nil
	}
	return &m.entries[len(m.entries)-1]
}

// mockAuthAlerter мок алерт-сервиса.
type mockAuthAlerter struct {
	mu       sync.Mutex
	failures []string
}

func (m *mockAuthAlerter) RecordFailure(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failures = append(m.failures, ip)
}

// mockAnyRequest реализует connect.AnyRequest для тестов.
type mockAnyRequest struct {
	connect.AnyRequest
	header http.Header
	peer   connect.Peer
	spec   connect.Spec
	ctx    context.Context
}

func (r *mockAnyRequest) Header() http.Header      { return r.header }
func (r *mockAnyRequest) Peer() connect.Peer       { return r.peer }
func (r *mockAnyRequest) Spec() connect.Spec       { return r.spec }
func (r *mockAnyRequest) Context() context.Context { return r.ctx }

// --- Хелперы ---

// discardLogger возвращает логгер, который ничего не пишет.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestRequest создаёт мок Connect-запрос с заданными заголовками.
func newTestRequest(headers map[string]string, remoteAddr string) *mockAnyRequest {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &mockAnyRequest{
		header: h,
		peer:   connect.Peer{Addr: remoteAddr},
		spec:   connect.Spec{Procedure: "/emh.v1.HeroAdminService/CreateHero"},
		ctx:    context.Background(),
	}
}

// dummyNext возвращает успешный connect.AnyResponse для прохода через интерцептор.
func dummyNext(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
	return nil, nil
}

// newTestInterceptor создаёт интерцептор с тестовыми параметрами.
func newTestInterceptor(
	staticKeys []string,
	apiKeyAuth *mockAPIKeyAuthenticator,
	trustedProxies []string,
	auditor *mockAuthAuditor,
	alerter *mockAuthAlerter,
) *AuthInterceptor {
	// JWT-секрет минимальной длины для прохождения валидации в auth.ValidateToken.
	return NewAuthInterceptor(
		"test-jwt-secret-at-least-32-characters-long",
		staticKeys,
		apiKeyAuth,
		trustedProxies,
		auditor,
		alerter,
		testIPPepper,
		discardLogger(),
	)
}

// --- Тесты: базовые сценарии ---

// TestAuthInterceptor_MissingToken возвращает CodeUnauthenticated при отсутствии токена.
func TestAuthInterceptor_MissingToken(t *testing.T) {
	interceptor := newTestInterceptor(nil, &mockAPIKeyAuthenticator{}, nil, &mockAuthAuditor{}, &mockAuthAlerter{})
	req := newTestRequest(map[string]string{}, "192.168.1.1:12345")

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing token")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeUnauthenticated)
	}
}

// TestAuthInterceptor_MalformedAuthHeader возвращает ошибку при некорректном заголовке.
func TestAuthInterceptor_MalformedAuthHeader(t *testing.T) {
	interceptor := newTestInterceptor(nil, &mockAPIKeyAuthenticator{}, nil, &mockAuthAuditor{}, &mockAuthAlerter{})
	req := newTestRequest(map[string]string{
		"Authorization": "NotBearer some-token",
	}, "192.168.1.1:12345")

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for malformed header")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeUnauthenticated)
	}
}

// TestAuthInterceptor_StaticKey_Success проверяет успешную аутентификацию статическим ключом.
func TestAuthInterceptor_StaticKey_Success(t *testing.T) {
	auditor := &mockAuthAuditor{}
	interceptor := newTestInterceptor(
		[]string{"emergency-key-123"},
		&mockAPIKeyAuthenticator{},
		nil,
		auditor,
		&mockAuthAlerter{},
	)
	req := newTestRequest(map[string]string{
		"Authorization": "Bearer emergency-key-123",
	}, "192.168.1.1:12345")

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("static key auth failed: %v", err)
	}

	// Проверяем аудит: success с mechanism=static_key.
	entry := auditor.lastEntry()
	if entry == nil {
		t.Fatal("audit entry not recorded")
	}
	if entry.Result != domain.AuthResultSuccess {
		t.Errorf("result = %q, want %q", entry.Result, domain.AuthResultSuccess)
	}
	if entry.Mechanism != domain.AuthMechanismStaticKey {
		t.Errorf("mechanism = %q, want %q", entry.Mechanism, domain.AuthMechanismStaticKey)
	}
}

// TestAuthInterceptor_StaticKey_NotMatched_FallsThrough проверяет, что не-матчащийся
// статический ключ не пропускает и передаёт управление следующему механизму.
func TestAuthInterceptor_StaticKey_NotMatched_FallsThrough(t *testing.T) {
	interceptor := newTestInterceptor(
		[]string{"emergency-key-123"},
		&mockAPIKeyAuthenticator{},
		nil,
		&mockAuthAuditor{},
		&mockAuthAlerter{},
	)
	req := newTestRequest(map[string]string{
		"Authorization": "Bearer wrong-static-key",
	}, "192.168.1.1:12345")

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)

	// Должна быть ошибка, т.к. ключ не прошёл ни один механизм (JWT валидация тоже провалит).
	if err == nil {
		t.Fatal("expected error for wrong static key")
	}
}

// --- Тесты: API-key ---

// TestAuthInterceptor_APIKey_Success проверяет успешную аутентификацию через API-key.
func TestAuthInterceptor_APIKey_Success(t *testing.T) {
	auditor := &mockAuthAuditor{}
	apiKeyAuth := &mockAPIKeyAuthenticator{
		authFunc: func(ctx context.Context, fullKey, ip string) (*domain.APIKey, error) {
			return &domain.APIKey{
				Name: "ci-key",
				Role: "admin",
			}, nil
		},
	}

	interceptor := newTestInterceptor(nil, apiKeyAuth, nil, auditor, &mockAuthAlerter{})
	// Формат API-ключа: emh_<key_id>_<secret> (функция auth.IsAPIKey проверяет префикс).
	req := newTestRequest(map[string]string{
		"Authorization": "Bearer emh_abc123_secret456",
	}, "192.168.1.1:12345")

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("api key auth failed: %v", err)
	}

	// Проверяем аудит.
	entry := auditor.lastEntry()
	if entry == nil {
		t.Fatal("audit entry not recorded")
	}
	if entry.Mechanism != domain.AuthMechanismAPIKey {
		t.Errorf("mechanism = %q, want %q", entry.Mechanism, domain.AuthMechanismAPIKey)
	}
	if entry.Identity != "ci-key" {
		t.Errorf("identity = %q, want %q", entry.Identity, "ci-key")
	}
}

// TestAuthInterceptor_APIKey_Invalid проверяет ошибку при невалидном API-ключе.
func TestAuthInterceptor_APIKey_Invalid(t *testing.T) {
	alerter := &mockAuthAlerter{}
	auditor := &mockAuthAuditor{}
	apiKeyAuth := &mockAPIKeyAuthenticator{
		authFunc: func(ctx context.Context, fullKey, ip string) (*domain.APIKey, error) {
			return nil, domain.ErrNotFound // имитация "ключа не существует"
		},
	}

	interceptor := newTestInterceptor(nil, apiKeyAuth, nil, auditor, alerter)
	req := newTestRequest(map[string]string{
		"Authorization": "Bearer emh_abc123_secret456",
	}, "192.168.1.1:12345")

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid api key")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeUnauthenticated)
	}

	// Проверяем алерт и аудит failure.
	alerter.mu.Lock()
	failureCount := len(alerter.failures)
	alerter.mu.Unlock()
	if failureCount != 1 {
		t.Errorf("expected 1 failure record, got %d", failureCount)
	}

	entry := auditor.lastEntry()
	if entry == nil {
		t.Fatal("failure audit entry not recorded")
	}
	if entry.Result != domain.AuthResultFailure {
		t.Errorf("result = %q, want %q", entry.Result, domain.AuthResultFailure)
	}
}

// --- Тесты: Trusted Proxies ---

// TestAuthInterceptor_TrustedProxy_ExactIP использует X-Forwarded-For от доверенного прокси.
func TestAuthInterceptor_TrustedProxy_ExactIP(t *testing.T) {
	auditor := &mockAuthAuditor{}
	// API-key auth, чтобы провалидировать IP, переданный в Authenticate.
	var capturedIP string
	apiKeyAuth := &mockAPIKeyAuthenticator{
		authFunc: func(ctx context.Context, fullKey, ip string) (*domain.APIKey, error) {
			capturedIP = ip
			return &domain.APIKey{Name: "test-key", Role: "admin"}, nil
		},
	}

	// Caddy с IP 10.0.0.5 — доверенный прокси.
	interceptor := newTestInterceptor(nil, apiKeyAuth, []string{"10.0.0.5"}, auditor, &mockAuthAlerter{})

	// Запрос от Caddy (10.0.0.5) с X-Forwarded-For оригинального клиента.
	req := newTestRequest(map[string]string{
		"Authorization":   "Bearer emh_abc123_secret456",
		"X-Forwarded-For": "203.0.113.42, 10.0.0.5",
	}, "10.0.0.5:54321")

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("auth from trusted proxy failed: %v", err)
	}

	// IP, переданный в apiKeyAuth.Authenticate, должен быть оригинальным клиентом.
	if capturedIP != "203.0.113.42" {
		t.Errorf("captured IP = %q, want %q (original client)", capturedIP, "203.0.113.42")
	}

	// Аудит должен содержать ХЕШ оригинального IP (защита приватности).
	// Сырой IP передаётся только в алерт-сервис (см. capturedIP выше).
	entry := auditor.lastEntry()
	if entry == nil {
		t.Fatal("audit entry not recorded")
	}
	wantIPHash := expectedIPHash("203.0.113.42", testIPPepper)
	if entry.IPAddress != wantIPHash {
		t.Errorf("audit IP = %q, want hash of 203.0.113.42 = %q", entry.IPAddress, wantIPHash)
	}
}

// TestAuthInterceptor_TrustedProxy_CIDR использует X-Forwarded-For от CIDR-диапазона.
func TestAuthInterceptor_TrustedProxy_CIDR(t *testing.T) {
	var capturedIP string
	apiKeyAuth := &mockAPIKeyAuthenticator{
		authFunc: func(ctx context.Context, fullKey, ip string) (*domain.APIKey, error) {
			capturedIP = ip
			return &domain.APIKey{Name: "test-key", Role: "admin"}, nil
		},
	}

	// Podman-сеть 10.88.0.0/16 — доверенный диапазон.
	interceptor := newTestInterceptor(nil, apiKeyAuth, []string{"10.88.0.0/16"}, &mockAuthAuditor{}, &mockAuthAlerter{})

	req := newTestRequest(map[string]string{
		"Authorization":   "Bearer emh_abc123_secret456",
		"X-Forwarded-For": "203.0.113.42",
	}, "10.88.1.25:54321") // внутри 10.88.0.0/16

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("auth from trusted CIDR failed: %v", err)
	}

	if capturedIP != "203.0.113.42" {
		t.Errorf("captured IP = %q, want %q", capturedIP, "203.0.113.42")
	}
}

// TestAuthInterceptor_UntrustedProxy_IgnoresXFF игнорирует X-Forwarded-For от недоверенного IP.
func TestAuthInterceptor_UntrustedProxy_IgnoresXFF(t *testing.T) {
	var capturedIP string
	apiKeyAuth := &mockAPIKeyAuthenticator{
		authFunc: func(ctx context.Context, fullKey, ip string) (*domain.APIKey, error) {
			capturedIP = ip
			return &domain.APIKey{Name: "test-key", Role: "admin"}, nil
		},
	}

	// Доверенный прокси: только 10.0.0.5.
	interceptor := newTestInterceptor(nil, apiKeyAuth, []string{"10.0.0.5"}, &mockAuthAuditor{}, &mockAuthAlerter{})

	// Атакующий с IP 198.51.100.1 пытается подделать X-Forwarded-For.
	req := newTestRequest(map[string]string{
		"Authorization":   "Bearer emh_abc123_secret456",
		"X-Forwarded-For": "127.0.0.1", // попытка обойти rate-limit/audit
	}, "198.51.100.1:54321") // не в списке доверенных

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	// Должен использоваться реальный IP соединения, а не поддельный XFF.
	if capturedIP != "198.51.100.1" {
		t.Errorf("captured IP = %q, want %q (real addr, not spoofed XFF)", capturedIP, "198.51.100.1")
	}
}

// TestAuthInterceptor_ExtractClientIP_NoXFF_XRealIP использует X-Real-IP как fallback.
func TestAuthInterceptor_ExtractClientIP_NoXFF_XRealIP(t *testing.T) {
	var capturedIP string
	apiKeyAuth := &mockAPIKeyAuthenticator{
		authFunc: func(ctx context.Context, fullKey, ip string) (*domain.APIKey, error) {
			capturedIP = ip
			return &domain.APIKey{Name: "test-key", Role: "admin"}, nil
		},
	}

	interceptor := newTestInterceptor(nil, apiKeyAuth, []string{"10.0.0.5"}, &mockAuthAuditor{}, &mockAuthAlerter{})
	req := newTestRequest(map[string]string{
		"Authorization": "Bearer emh_abc123_secret456",
		"X-Real-IP":     "203.0.113.42", // без X-Forwarded-For
	}, "10.0.0.5:54321")

	handler := interceptor.WrapUnary(dummyNext)
	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	if capturedIP != "203.0.113.42" {
		t.Errorf("captured IP = %q, want %q (X-Real-IP)", capturedIP, "203.0.113.42")
	}
}

// --- Тесты: вспомогательные функции ---

// TestExtractBearerToken проверяет парсинг Authorization заголовка.
func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"valid", "Bearer my-token", "my-token"},
		{"valid lowercase", "bearer my-token", "my-token"},
		{"valid mixed case", "BeArEr my-token", "my-token"},
		{"with extra spaces", "Bearer   my-token  ", "my-token"},
		{"no bearer prefix", "my-token", ""},
		{"basic auth", "Basic dXNlcjpwYXNz", ""},
		{"empty", "", ""},
		{"only bearer", "Bearer", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			if tt.value != "" {
				h.Set("Authorization", tt.value)
			}
			got := extractBearerToken(h)
			if got != tt.expected {
				t.Errorf("extractBearerToken(%q) = %q, want %q", tt.value, got, tt.expected)
			}
		})
	}
}

// TestAuthInterceptor_IPConsistency проверяет, что IP одинаковый для всех механизмов.
// Это важно для аудита: один клиент должен иметь один IP во всех записях auth_audit_log.
func TestAuthInterceptor_IPConsistency(t *testing.T) {
	proxyIP := "192.168.1.1"
	clientIP := "203.0.113.50"
	jwtSecret := "test-jwt-secret-at-least-32-characters-long"

	tests := []struct {
		name          string
		setupAuth     func() string // возвращает Authorization header
		wantMechanism domain.AuthAuditMechanism
	}{
		{
			name: "static key",
			setupAuth: func() string {
				return "Bearer static-key-1"
			},
			wantMechanism: domain.AuthMechanismStaticKey,
		},
		{
			name: "JWT",
			setupAuth: func() string {
				token, _ := auth.GenerateToken(jwtSecret, "admin", "admin", 15*time.Minute)
				return "Bearer " + token
			},
			wantMechanism: domain.AuthMechanismJWT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём mock auditor для захвата записей
			var capturedEntries []domain.AuthAuditEntry
			mockAuditor := &mockAuthAuditor{
				enqueueFunc: func(entry domain.AuthAuditEntry) {
					capturedEntries = append(capturedEntries, entry)
				},
			}

			interceptor := NewAuthInterceptor(
				jwtSecret,
				[]string{"static-key-1"},
				nil, // apiKeyAuth не нужен для этих тестов
				[]string{proxyIP},
				mockAuditor,
				nil, // alerter
				testIPPepper,
				discardLogger(),
			)

			header := http.Header{}
			header.Set("X-Forwarded-For", clientIP)
			header.Set("Authorization", tt.setupAuth())

			ctx, err := interceptor.authenticate(context.Background(), header, proxyIP+":12345", "test-agent", "test-procedure")
			if err != nil {
				t.Fatalf("authenticate failed: %v", err)
			}

			// Проверяем что аудит был записан
			if len(capturedEntries) != 1 {
				t.Fatalf("expected 1 audit entry, got %d", len(capturedEntries))
			}

			// В аудите хранится хеш IP (защита приватности), а не сырой адрес.
			entry := capturedEntries[0]
			wantIPHash := expectedIPHash(clientIP, testIPPepper)
			if entry.IPAddress != wantIPHash {
				t.Errorf("expected IP hash %q, got %q", wantIPHash, entry.IPAddress)
			}
			// Проверяем, что сырой ИП НЕ попал в аудит
			if entry.IPAddress == clientIP {
				t.Error("raw IP must not be stored in audit — only hash")
			}
			if entry.Mechanism != tt.wantMechanism {
				t.Errorf("expected mechanism %v, got %v", tt.wantMechanism, entry.Mechanism)
			}
			if entry.Result != domain.AuthResultSuccess {
				t.Errorf("expected success, got %v", entry.Result)
			}

			// Для JWT проверяем что claims сохранены в контексте
			if tt.wantMechanism == domain.AuthMechanismJWT {
				claims, ok := ClaimsFromContext(ctx)
				if !ok {
					t.Error("expected claims in context for JWT")
				}
				if claims.Username != "admin" {
					t.Errorf("expected username 'admin', got %q", claims.Username)
				}
			}
		})
	}
}

// expectedIPHash вычисляет ожидаемый хеш IP для проверок в тестах.
// Дублирует логику AuthInterceptor.hashIP, чтобы тесты были независимыми.
func expectedIPHash(ip, pepper string) string {
	if ip == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}
