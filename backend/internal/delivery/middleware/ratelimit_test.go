package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
)

// --- Хелперы ---

// discardLogger возвращает логгер, который ничего не пишет.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestConfig создаёт тестовую конфигурацию с низкими лимитами.
func newTestConfig() config.RateLimitConfig {
	return config.RateLimitConfig{
		Enabled:            true,
		AuthLimit:          2,
		AuthInterval:       config.Duration{Duration: 1 * time.Minute},
		SubmissionLimit:    3,
		SubmissionInterval: config.Duration{Duration: 1 * time.Minute},
		MediaLimit:         4,
		MediaInterval:      config.Duration{Duration: 1 * time.Minute},
		ReadLimit:          5,
		ReadInterval:       config.Duration{Duration: 1 * time.Minute},
		AdminLimit:         6,
		AdminInterval:      config.Duration{Duration: 1 * time.Minute},
		EntryTTL:           config.Duration{Duration: 100 * time.Millisecond},
		CleanupInterval:    config.Duration{Duration: 50 * time.Millisecond},
	}
}

// newDisabledConfig создаёт конфигурацию с отключённым rate limiting.
func newDisabledConfig() config.RateLimitConfig {
	cfg := newTestConfig()
	cfg.Enabled = false
	return cfg
}

// dummyHandler возвращает простой handler, который возвращает 200 OK.
func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

// makeRequest создаёт HTTP-запрос с заданными заголовками.
func makeRequest(method, path string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.RemoteAddr = "192.168.1.1:12345"
	return req
}

// --- Тесты: базовое поведение ---

// TestRateLimitMiddleware_Disabled проверяет, что отключённый rate limiting пропускает все запросы.
func TestRateLimitMiddleware_Disabled(t *testing.T) {
	mw := NewRateLimitMiddleware(newDisabledConfig(), nil, discardLogger())
	handler := mw.Handler(dummyHandler())

	// Делаем много запросов — все должны пройти.
	for i := 0; i < 20; i++ {
		req := makeRequest("GET", "/emh.v1.HeroService/GetHero", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

// TestRateLimitMiddleware_OptionsPassthrough проверяет, что OPTIONS не лимитируется.
func TestRateLimitMiddleware_OptionsPassthrough(t *testing.T) {
	mw := NewRateLimitMiddleware(newTestConfig(), nil, discardLogger())
	handler := mw.Handler(dummyHandler())

	// Делаем много OPTIONS-запросов — все должны пройти.
	for i := 0; i < 20; i++ {
		req := makeRequest("OPTIONS", "/emh.v1.AuthService/Login", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("OPTIONS request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

// TestRateLimitMiddleware_HealthzPassthrough проверяет, что /healthz не лимитируется.
func TestRateLimitMiddleware_HealthzPassthrough(t *testing.T) {
	mw := NewRateLimitMiddleware(newTestConfig(), nil, discardLogger())
	handler := mw.Handler(dummyHandler())

	for i := 0; i < 20; i++ {
		req := makeRequest("GET", "/healthz", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("healthz request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

// TestRateLimitMiddleware_AdminGroup проверяет, что admin-сервисы лимитируются
// по собственной группе (M2, security.md) — не bypass, но и не наравне с read.
func TestRateLimitMiddleware_AdminGroup(t *testing.T) {
	cfg := newTestConfig()
	cfg.AdminLimit = 3 // низкий лимит для теста
	mw := NewRateLimitMiddleware(cfg, nil, discardLogger())
	handler := mw.Handler(dummyHandler())

	// Первые 3 запроса должны пройти.
	for i := 0; i < 3; i++ {
		req := makeRequest("POST", "/emh.v1.HeroAdminService/CreateHero", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("admin request %d: expected 200, got %d", i, rec.Code)
		}
	}

	// 4-й запрос должен быть отклонён лимитером.
	req := makeRequest("POST", "/emh.v1.HeroAdminService/CreateHero", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("admin request over limit: expected 429, got %d", rec.Code)
	}
}

// --- Тесты: rate limiting по группам ---

// TestRateLimitMiddleware_AuthGroup проверяет лимиты для AuthService.
func TestRateLimitMiddleware_AuthGroup(t *testing.T) {
	cfg := newTestConfig()
	cfg.AuthLimit = 2 // низкий лимит для теста
	mw := NewRateLimitMiddleware(cfg, nil, discardLogger())
	handler := mw.Handler(dummyHandler())

	// Первые 2 запроса должны пройти.
	for i := 0; i < 2; i++ {
		req := makeRequest("POST", "/emh.v1.AuthService/Login", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
	}

	// Третий запрос должен вернуть 429.
	req := makeRequest("POST", "/emh.v1.AuthService/Login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}

	// Проверяем Retry-After header.
	if rec.Header().Get("Retry-After") != "60" {
		t.Errorf("Retry-After = %q, want 60", rec.Header().Get("Retry-After"))
	}

	// Проверяем тело ответа.
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["code"] != "resource_exhausted" {
		t.Errorf("code = %q, want resource_exhausted", resp["code"])
	}
}

// TestRateLimitMiddleware_ReadGroup проверяет лимиты для публичных read-сервисов.
func TestRateLimitMiddleware_ReadGroup(t *testing.T) {
	cfg := newTestConfig()
	cfg.ReadLimit = 3
	mw := NewRateLimitMiddleware(cfg, nil, discardLogger())
	handler := mw.Handler(dummyHandler())

	// Первые 3 запроса проходят.
	for i := 0; i < 3; i++ {
		req := makeRequest("GET", "/emh.v1.HeroService/ListHeroes", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
	}

	// Четвёртый запрос отклоняется.
	req := makeRequest("GET", "/emh.v1.HeroService/ListHeroes", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}

// TestRateLimitMiddleware_DifferentIPsAreIndependent проверяет независимость лимитов для разных IP.
func TestRateLimitMiddleware_DifferentIPsAreIndependent(t *testing.T) {
	cfg := newTestConfig()
	cfg.AuthLimit = 1
	mw := NewRateLimitMiddleware(cfg, nil, discardLogger())
	handler := mw.Handler(dummyHandler())

	// IP1: первый запрос проходит.
	req1 := makeRequest("POST", "/emh.v1.AuthService/Login", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("IP1 first request: expected 200, got %d", rec1.Code)
	}

	// IP1: второй запрос отклоняется.
	req1b := makeRequest("POST", "/emh.v1.AuthService/Login", nil)
	req1b.RemoteAddr = "192.168.1.1:12345"
	rec1b := httptest.NewRecorder()
	handler.ServeHTTP(rec1b, req1b)
	if rec1b.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 second request: expected 429, got %d", rec1b.Code)
	}

	// IP2: первый запрос проходит (независимый лимит).
	req2 := makeRequest("POST", "/emh.v1.AuthService/Login", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("IP2 first request: expected 200, got %d", rec2.Code)
	}
}

// TestRateLimitMiddleware_DifferentGroupsAreIndependent проверяет независимость групп.
func TestRateLimitMiddleware_DifferentGroupsAreIndependent(t *testing.T) {
	cfg := newTestConfig()
	cfg.AuthLimit = 1
	cfg.ReadLimit = 1
	mw := NewRateLimitMiddleware(cfg, nil, discardLogger())
	handler := mw.Handler(dummyHandler())

	// Auth: первый запрос проходит.
	req1 := makeRequest("POST", "/emh.v1.AuthService/Login", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("auth first: expected 200, got %d", rec1.Code)
	}

	// Read: первый запрос тоже проходит (другая группа).
	req2 := makeRequest("GET", "/emh.v1.HeroService/ListHeroes", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("read first: expected 200, got %d", rec2.Code)
	}

	// Auth: второй запрос отклоняется.
	req3 := makeRequest("POST", "/emh.v1.AuthService/Login", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("auth second: expected 429, got %d", rec3.Code)
	}

	// Read: второй запрос тоже отклоняется.
	req4 := makeRequest("GET", "/emh.v1.HeroService/ListHeroes", nil)
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusTooManyRequests {
		t.Errorf("read second: expected 429, got %d", rec4.Code)
	}
}

// --- Тесты: X-Forwarded-For ---

// TestRateLimitMiddleware_TrustedProxy_UsesXFF проверяет, что для доверенного прокси
// IP клиента берётся из XFF (rightmost-untrusted).
func TestRateLimitMiddleware_TrustedProxy_UsesXFF(t *testing.T) {
	cfg := newTestConfig()
	cfg.AuthLimit = 1
	mw := NewRateLimitMiddleware(cfg, []string{"10.0.0.5"}, discardLogger())
	handler := mw.Handler(dummyHandler())

	// Запрос от клиента 203.0.113.42 через прокси 10.0.0.5
	// (XFF: клиентский спуфинг слева, реальный IP дописан прокси последним).
	makeXFFRequest := func(xff string) *httptest.ResponseRecorder {
		req := makeRequest("POST", "/emh.v1.AuthService/Login", map[string]string{
			"X-Forwarded-For": xff,
		})
		req.RemoteAddr = "10.0.0.5:443"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	if rec := makeXFFRequest("6.6.6.6, 203.0.113.42"); rec.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec.Code)
	}
	if rec := makeXFFRequest("7.7.7.7, 203.0.113.42"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429 (same real client), got %d", rec.Code)
	}
	if rec := makeXFFRequest("8.8.8.8, 198.51.100.1"); rec.Code != http.StatusOK {
		t.Errorf("different client: expected 200, got %d", rec.Code)
	}
}

// TestRateLimitMiddleware_UntrustedProxy_IgnoresXFF проверяет, что подделка XFF
// от недоверенного адреса не даёт новые token bucket (обход лимитов).
func TestRateLimitMiddleware_UntrustedProxy_IgnoresXFF(t *testing.T) {
	cfg := newTestConfig()
	cfg.AuthLimit = 1
	mw := NewRateLimitMiddleware(cfg, []string{"10.0.0.5"}, discardLogger())
	handler := mw.Handler(dummyHandler())

	// Прямое соединение с атакующего IP, каждый запрос — с новым поддельным XFF.
	for i := 0; i < 5; i++ {
		req := makeRequest("POST", "/emh.v1.AuthService/Login", map[string]string{
			"X-Forwarded-For": fmt.Sprintf("203.0.11%d.42", i),
		})
		req.RemoteAddr = "198.51.100.7:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if i == 0 && rec.Code != http.StatusOK {
			t.Fatalf("first request: expected 200, got %d", rec.Code)
		}
		if i > 0 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("request %d: expected 429 despite rotating XFF, got %d", i, rec.Code)
		}
	}
}

// TestRateLimitMiddleware_XRealIpFromTrustedProxy проверяет fallback на X-Real-IP
// от доверенного прокси.
func TestRateLimitMiddleware_XRealIpFromTrustedProxy(t *testing.T) {
	cfg := newTestConfig()
	cfg.AuthLimit = 1
	mw := NewRateLimitMiddleware(cfg, []string{"10.0.0.5"}, discardLogger())
	handler := mw.Handler(dummyHandler())

	makeReq := func() *httptest.ResponseRecorder {
		req := makeRequest("POST", "/emh.v1.AuthService/Login", map[string]string{
			"X-Real-IP": "203.0.113.42",
		})
		req.RemoteAddr = "10.0.0.5:443"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	if rec := makeReq(); rec.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec.Code)
	}
	if rec := makeReq(); rec.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rec.Code)
	}
}

// --- Тесты: classifyService ---

// TestClassifyService проверяет определение группы по пути.
func TestClassifyService(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/emh.v1.AuthService/Login", "auth"},
		{"/emh.v1.AuthService/Refresh", "auth"},
		{"/emh.v1.SubmissionService/CreateSubmission", "submission"},
		{"/emh.v1.MediaService/GetUploadUrl", "media"},
		{"/emh.v1.HeroAdminService/CreateHero", "admin"},
		{"/emh.v1.ApiKeyAdminService/CreateApiKey", "admin"},
		{"/emh.v1.SubmissionAdminService/ApproveSubmission", "admin"},
		{"/emh.v1.HeroService/GetHero", "read"},
		{"/emh.v1.HeroService/ListHeroes", "read"},
		{"/emh.v1.ConflictService/ListConflicts", "read"},
		{"/emh.v1.LocationService/GetLocation", "read"},
		{"/unknown", "read"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := classifyService(tt.path); got != tt.expected {
				t.Errorf("classifyService(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

// --- Тесты: getLimiter ---

// TestGetLimiter_CreatesNew проверяет создание нового лимитера.
func TestGetLimiter_CreatesNew(t *testing.T) {
	mw := NewRateLimitMiddleware(newTestConfig(), nil, discardLogger())

	limiter := mw.getLimiter("test-key", 10, time.Minute)
	if limiter == nil {
		t.Fatal("getLimiter returned nil")
	}

	// Проверяем, что лимитер сохранён.
	mw.mu.Lock()
	_, exists := mw.limiters["test-key"]
	mw.mu.Unlock()

	if !exists {
		t.Error("limiter not stored in map")
	}
}

// TestGetLimiter_ReusesExisting проверяет переиспользование существующего лимитера.
func TestGetLimiter_ReusesExisting(t *testing.T) {
	mw := NewRateLimitMiddleware(newTestConfig(), nil, discardLogger())

	limiter1 := mw.getLimiter("test-key", 10, time.Minute)
	limiter2 := mw.getLimiter("test-key", 10, time.Minute)

	if limiter1 != limiter2 {
		t.Error("getLimiter should return same limiter for same key")
	}
}

// TestGetLimiter_UpdatesLastSeen проверяет обновление lastSeen при доступе.
func TestGetLimiter_UpdatesLastSeen(t *testing.T) {
	mw := NewRateLimitMiddleware(newTestConfig(), nil, discardLogger())

	mw.getLimiter("test-key", 10, time.Minute)

	mw.mu.Lock()
	rl := mw.limiters["test-key"]
	firstSeen := rl.lastSeen
	mw.mu.Unlock()

	time.Sleep(10 * time.Millisecond)

	mw.getLimiter("test-key", 10, time.Minute)

	mw.mu.Lock()
	secondSeen := rl.lastSeen
	mw.mu.Unlock()

	if !secondSeen.After(firstSeen) {
		t.Error("lastSeen should be updated on access")
	}
}

// --- Тесты: cleanupLoop ---

// TestCleanupLoop_RemovesExpiredEntries проверяет удаление записей старше TTL.
func TestCleanupLoop_RemovesExpiredEntries(t *testing.T) {
	cfg := newTestConfig()
	cfg.EntryTTL = config.Duration{Duration: 50 * time.Millisecond}
	cfg.CleanupInterval = config.Duration{Duration: 25 * time.Millisecond}

	mw := NewRateLimitMiddleware(cfg, nil, discardLogger())

	// Создаём лимитер.
	mw.getLimiter("test-key", 10, time.Minute)

	// Проверяем, что запись создана.
	mw.mu.Lock()
	if len(mw.limiters) != 1 {
		t.Fatalf("expected 1 limiter, got %d", len(mw.limiters))
	}
	mw.mu.Unlock()

	// Запускаем cleanup loop.
	ctx, cancel := context.WithCancel(context.Background())
	go mw.cleanupLoop(ctx)

	// Ждём истечения TTL + cleanup interval.
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что запись удалена.
	mw.mu.Lock()
	count := len(mw.limiters)
	mw.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 limiters after cleanup, got %d", count)
	}

	cancel() // останавливаем cleanup loop
}

// TestCleanupLoop_StopsOnContextCancel проверяет остановку при отмене контекста.
func TestCleanupLoop_StopsOnContextCancel(t *testing.T) {
	cfg := newTestConfig()
	mw := NewRateLimitMiddleware(cfg, nil, discardLogger())

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		mw.cleanupLoop(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// OK: cleanup loop остановился
	case <-time.After(1 * time.Second):
		t.Fatal("cleanupLoop did not stop after context cancel")
	}
}

// --- Тесты: Start ---

// TestStart_DisabledDoesNotLaunchGoroutine проверяет, что отключённый middleware не запускает горутину.
func TestStart_DisabledDoesNotLaunchGoroutine(t *testing.T) {
	mw := NewRateLimitMiddleware(newDisabledConfig(), nil, discardLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start не должен запустить cleanupLoop.
	mw.Start(ctx)

	// Даём время на возможный запуск горутины.
	time.Sleep(50 * time.Millisecond)

	// Проверяем, что никаких побочных эффектов нет.
	// (cleanupLoop не запущен, поэтому ничего не происходит)
}

// TestStart_EnabledLaunchesGoroutine проверяет, что включённый middleware запускает cleanup.
func TestStart_EnabledLaunchesGoroutine(t *testing.T) {
	cfg := newTestConfig()
	cfg.EntryTTL = config.Duration{Duration: 50 * time.Millisecond}
	cfg.CleanupInterval = config.Duration{Duration: 25 * time.Millisecond}

	mw := NewRateLimitMiddleware(cfg, nil, discardLogger())

	// Создаём лимитер.
	mw.getLimiter("test-key", 10, time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw.Start(ctx)

	// Ждём истечения TTL + cleanup.
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что запись удалена (cleanup loop работает).
	mw.mu.Lock()
	count := len(mw.limiters)
	mw.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 limiters after cleanup, got %d", count)
	}
}
