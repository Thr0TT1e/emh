package config

import (
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

// validTestConfig возвращает базовый валидный конфиг для тестов.
// Все проверки должны проходить.
func validTestConfig() *Config {
	return &Config{
		App: AppConfig{
			Addr: ":3480",
		},
		DB: DBConfig{
			URL:      "postgres://user:pass@localhost:5432/test",
			MaxConns: 10,
		},
		S3: S3Config{
			Endpoint:      "minio:9000",
			Bucket:        "test-bucket",
			AccessKey:     "minioadmin",
			SecretKey:     "minioadmin",
			PresignExpiry: Duration{15 * time.Minute},
		},
		Auth: AuthConfig{
			JWTSecret:     "test-jwt-secret-at-least-32-characters-long",
			JWTExpiry:     Duration{15 * time.Minute},
			RefreshExpiry: Duration{168 * time.Hour},
			Admins: []AdminConfig{
				{Username: "admin", PasswordHash: "$2a$10$fTNJjClTwWQZsTYHbCuqwuaiCPT7mout5UzU0kSt/xbtlsEsCnDoO", Role: "admin"},
			},
		},
		Thumbnails: ThumbnailConfig{
			Width:           480,
			Height:          600,
			Quality:         80,
			OriginalQuality: 90,
			Workers:         2,
			QueueSize:       100,
		},
		RateLimit: RateLimitConfig{
			Enabled:            true,
			AuthLimit:          5,
			AuthInterval:       Duration{1 * time.Minute},
			SubmissionLimit:    10,
			SubmissionInterval: Duration{1 * time.Hour},
			MediaLimit:         20,
			MediaInterval:      Duration{1 * time.Minute},
			ReadLimit:          120,
			ReadInterval:       Duration{1 * time.Minute},
			AdminLimit:         60,
			AdminInterval:      Duration{1 * time.Minute},
			EntryTTL:           Duration{30 * time.Minute},
			CleanupInterval:    Duration{10 * time.Minute},
		},
		Contact: ContactConfig{
			SMTPEnabled:     false,
			IPPepper:        "test-ip-pepper-16-chars-min",
			RateLimitHour:   5,
			RateLimitDay:    20,
			DedupeWindow:    Duration{5 * time.Minute},
			EntryTTL:        Duration{24 * time.Hour},
			CleanupInterval: Duration{10 * time.Minute},
		},
	}
}

// --- Базовая валидация ---

// TestConfig_Validate_Valid проверяет, что валидный конфиг проходит.
func TestConfig_Validate_Valid(t *testing.T) {
	cfg := validTestConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

// TestConfig_Validate_MissingAppAddr проверяет ошибку при отсутствии адреса.
func TestConfig_Validate_MissingAppAddr(t *testing.T) {
	cfg := validTestConfig()
	cfg.App.Addr = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing app.addr")
	}
	if !strings.Contains(err.Error(), "app.addr") {
		t.Errorf("expected 'app.addr' in error, got %v", err)
	}
}

// TestConfig_Validate_MissingDBURL проверяет ошибку при отсутствии URL БД.
func TestConfig_Validate_MissingDBURL(t *testing.T) {
	cfg := validTestConfig()
	cfg.DB.URL = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing db.url")
	}
}

// --- JWT_SECRET (#11) ---

// TestConfig_Validate_ShortJWTSecret проверяет ошибку при коротком секрете.
func TestConfig_Validate_ShortJWTSecret(t *testing.T) {
	cfg := validTestConfig()
	cfg.Auth.JWTSecret = "short"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for short JWT secret")
	}
	if !strings.Contains(err.Error(), "jwt_secret") {
		t.Errorf("expected 'jwt_secret' in error, got %v", err)
	}
}

// TestConfig_Validate_EmptyJWTSecret проверяет ошибку при пустом секрете.
func TestConfig_Validate_EmptyJWTSecret(t *testing.T) {
	cfg := validTestConfig()
	cfg.Auth.JWTSecret = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty JWT secret")
	}
}

// TestConfig_Validate_WeakJWTSecret проверяет ошибку при небезопасном секрете.
func TestConfig_Validate_WeakJWTSecret(t *testing.T) {
	// Слабые секреты должны быть ≥32 символов, чтобы пройти проверку длины
	// и попасть на проверку слабости. Это демонстрирует, что даже длинные
	// предсказуемые секреты отклоняются.
	weakSecrets := []string{
		"changeme-changeme-changeme-changeme",         // 36 символов
		"CHANGE-ME-CHANGE-ME-CHANGE-ME-CHANGE-ME",     // 39 символов
		"change_me_change_me_change_me_change_me",     // 39 символов
		"secret-secret-secret-secret-secret-secret",   // 39 символов
		"test-secret-test-secret-test-secret-test",    // 39 символов
		"dev-secret-dev-secret-dev-secret-dev-secret", // 42 символа
		"jwt-secret-jwt-secret-jwt-secret-jwt-secret", // 42 символа
	}
	for _, secret := range weakSecrets {
		t.Run(secret, func(t *testing.T) {
			cfg := validTestConfig()
			cfg.Auth.JWTSecret = secret
			err := cfg.Validate()
			if err == nil {
				t.Errorf("expected error for weak secret %q", secret)
			}
			if err != nil && !strings.Contains(err.Error(), "weak") {
				t.Errorf("expected 'weak' in error, got %v", err)
			}
		})
	}
}

// TestConfig_Validate_JWTSecretWithWhitespace проверяет ошибку при секрете с пробелами.
func TestConfig_Validate_JWTSecretWithWhitespace(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{"space", "test jwt secret with spaces here 32chars"},              // 42 символа
		{"tab", "test\tjwt\tsecret\twith\ttabs\there-32chars!!"},           // 38 символов
		{"newline", "test-jwt-secret-with-newline\nhere-32char"},           // 40 символов
		{"trailing_space", "test-jwt-secret-at-least-32-characters-long "}, // 45 символов
		{"leading_space", " test-jwt-secret-at-least-32-characters-long"},  // 45 символов
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validTestConfig()
			cfg.Auth.JWTSecret = tt.secret
			err := cfg.Validate()
			if err == nil {
				t.Errorf("expected error for JWT secret with %s", tt.name)
			}
			if err != nil && !strings.Contains(err.Error(), "whitespace") {
				t.Errorf("expected 'whitespace' in error, got %v", err)
			}
		})
	}
}

// --- IPPepper ---

// TestConfig_Validate_EmptyIPPepper проверяет ошибку при пустом IP pepper.
func TestConfig_Validate_EmptyIPPepper(t *testing.T) {
	cfg := validTestConfig()
	cfg.Contact.IPPepper = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty IP pepper")
	}
	if !strings.Contains(err.Error(), "ip_pepper") {
		t.Errorf("expected 'ip_pepper' in error, got %v", err)
	}
}

// TestConfig_Validate_WhitespaceIPPepper проверяет ошибку при пустом из пробелов.
func TestConfig_Validate_WhitespaceIPPepper(t *testing.T) {
	cfg := validTestConfig()
	cfg.Contact.IPPepper = "   "
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for whitespace-only IP pepper")
	}
}

// TestConfig_Validate_ShortIPPepper проверяет ошибку при коротком IP pepper.
func TestConfig_Validate_ShortIPPepper(t *testing.T) {
	cfg := validTestConfig()
	cfg.Contact.IPPepper = "short"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for short IP pepper")
	}
	if !strings.Contains(err.Error(), "at least 16") {
		t.Errorf("expected 'at least 16' in error, got %v", err)
	}
}

// TestConfig_Validate_IPPepperRequiredEvenWhenSMTPDisabled проверяет,
// что проверка выполняется независимо от состояния SMTP.
func TestConfig_Validate_IPPepperRequiredEvenWhenSMTPDisabled(t *testing.T) {
	cfg := validTestConfig()
	cfg.Contact.SMTPEnabled = false
	cfg.Contact.IPPepper = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty IP pepper even when SMTP disabled")
	}
}

// --- S3 ---

// TestConfig_Validate_MissingS3Endpoint проверяет ошибку при отсутствии эндпоинта.
func TestConfig_Validate_MissingS3Endpoint(t *testing.T) {
	cfg := validTestConfig()
	cfg.S3.Endpoint = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing S3 endpoint")
	}
}

// TestConfig_Validate_MissingS3Credentials проверяет ошибку при отсутствии ключей.
func TestConfig_Validate_MissingS3Credentials(t *testing.T) {
	cfg := validTestConfig()
	cfg.S3.AccessKey = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing S3 credentials")
	}
}

// --- Auth ---

// TestConfig_Validate_NoAdmins проверяет ошибку при отсутствии админов.
func TestConfig_Validate_NoAdmins(t *testing.T) {
	cfg := validTestConfig()
	cfg.Auth.Admins = nil
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing admins")
	}
}

// TestConfig_Validate_ZeroRefreshExpiry проверяет ошибку при нулевом сроке.
func TestConfig_Validate_ZeroRefreshExpiry(t *testing.T) {
	cfg := validTestConfig()
	cfg.Auth.RefreshExpiry = Duration{0}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for zero refresh_expiry")
	}
}

// --- Contact / SMTP ---

// TestConfig_Validate_ContactSMTP проверяет валидацию контактного блока.
func TestConfig_Validate_ContactSMTP(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "missing smtp_host",
			mutate: func(c *Config) {
				c.Contact.SMTPEnabled = true
				c.Contact.SMTPHost = ""
				c.Contact.SMTPPort = 587
				c.Contact.SMTPFromEmail = "test@example.com"
				c.Contact.SMTPTo = "admin@example.com"
				c.Contact.SMTPTimeout = Duration{10 * time.Second}
			},
			wantErr: "smtp_host",
		},
		{
			name: "missing smtp_from_email",
			mutate: func(c *Config) {
				c.Contact.SMTPEnabled = true
				c.Contact.SMTPHost = "smtp.example.com"
				c.Contact.SMTPPort = 587
				c.Contact.SMTPFromEmail = ""
				c.Contact.SMTPTo = "admin@example.com"
				c.Contact.SMTPTimeout = Duration{10 * time.Second}
			},
			wantErr: "smtp_from_email",
		},
		{
			name: "missing smtp_to",
			mutate: func(c *Config) {
				c.Contact.SMTPEnabled = true
				c.Contact.SMTPHost = "smtp.example.com"
				c.Contact.SMTPPort = 587
				c.Contact.SMTPFromEmail = "test@example.com"
				c.Contact.SMTPTo = ""
				c.Contact.SMTPTimeout = Duration{10 * time.Second}
			},
			wantErr: "smtp_to",
		},
		{
			name: "invalid smtp_port",
			mutate: func(c *Config) {
				c.Contact.SMTPEnabled = true
				c.Contact.SMTPHost = "smtp.example.com"
				c.Contact.SMTPPort = -1
				c.Contact.SMTPFromEmail = "test@example.com"
				c.Contact.SMTPTo = "admin@example.com"
				c.Contact.SMTPTimeout = Duration{10 * time.Second}
			},
			wantErr: "smtp_port",
		},
		{
			name: "zero smtp_timeout",
			mutate: func(c *Config) {
				c.Contact.SMTPEnabled = true
				c.Contact.SMTPHost = "smtp.example.com"
				c.Contact.SMTPPort = 587
				c.Contact.SMTPFromEmail = "test@example.com"
				c.Contact.SMTPTo = "admin@example.com"
				c.Contact.SMTPTimeout = Duration{0}
			},
			wantErr: "smtp_timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validTestConfig()
			tt.mutate(cfg)
			err := cfg.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected %q in error, got %v", tt.wantErr, err)
			}
		})
	}
}

// --- Thumbnails ---

// TestConfig_Validate_Thumbnails проверяет валидацию параметров превью.
func TestConfig_Validate_Thumbnails(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"zero width", func(c *Config) { c.Thumbnails.Width = 0 }},
		{"negative height", func(c *Config) { c.Thumbnails.Height = -100 }},
		{"quality out of range", func(c *Config) { c.Thumbnails.Quality = 150 }},
		{"original quality out of range", func(c *Config) { c.Thumbnails.OriginalQuality = 0 }},
		{"zero workers", func(c *Config) { c.Thumbnails.Workers = 0 }},
		{"zero queue size", func(c *Config) { c.Thumbnails.QueueSize = 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validTestConfig()
			tt.mutate(cfg)
			if err := cfg.Validate(); err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

// TestConfig_Validate_RateLimit проверяет валидацию полей rate_limit,
// включая admin_limit/admin_interval (M2, security.md).
func TestConfig_Validate_RateLimit(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"zero auth limit", func(c *Config) { c.RateLimit.AuthLimit = 0 }},
		{"zero auth interval", func(c *Config) { c.RateLimit.AuthInterval = Duration{} }},
		{"zero submission limit", func(c *Config) { c.RateLimit.SubmissionLimit = 0 }},
		{"zero media limit", func(c *Config) { c.RateLimit.MediaLimit = 0 }},
		{"zero read limit", func(c *Config) { c.RateLimit.ReadLimit = 0 }},
		{"zero admin limit", func(c *Config) { c.RateLimit.AdminLimit = 0 }},
		{"zero admin interval", func(c *Config) { c.RateLimit.AdminInterval = Duration{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validTestConfig()
			tt.mutate(cfg)
			if err := cfg.Validate(); err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

// --- Override from env ---

// TestConfig_OverrideFromEnv проверяет переопределение из окружения.
func TestConfig_OverrideFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://new:new@localhost:5432/newdb")
	t.Setenv("AUTH_JWT_SECRET", "env-jwt-secret-override-32-chars!")
	t.Setenv("APP_ADDR", ":9999")

	cfg := validTestConfig()
	cfg.overrideFromEnv()

	if cfg.DB.URL != "postgres://new:new@localhost:5432/newdb" {
		t.Errorf("DB.URL not overridden: %v", cfg.DB.URL)
	}
	if cfg.Auth.JWTSecret != "env-jwt-secret-override-32-chars!" {
		t.Errorf("JWTSecret not overridden: %v", cfg.Auth.JWTSecret)
	}
	if cfg.App.Addr != ":9999" {
		t.Errorf("App.Addr not overridden: %v", cfg.App.Addr)
	}
}

// TestConfig_OverrideFromEnv_APIKeys проверяет парсинг списка ключей.
func TestConfig_OverrideFromEnv_APIKeys(t *testing.T) {
	t.Setenv("AUTH_API_KEYS", "key1,key2,key3")

	cfg := validTestConfig()
	cfg.overrideFromEnv()

	if len(cfg.Auth.APIKeys) != 3 {
		t.Errorf("expected 3 API keys, got %d", len(cfg.Auth.APIKeys))
	}
	if cfg.Auth.APIKeys[0] != "key1" {
		t.Errorf("expected key1, got %q", cfg.Auth.APIKeys[0])
	}
}

// --- Duration ---

// TestDuration_UnmarshalYAML проверяет парсинг длительности из строки.
func TestDuration_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{"minutes", "15m", 15 * time.Minute},
		{"hours", "168h", 168 * time.Hour},
		{"seconds", "10s", 10 * time.Second},
		{"complex", "1h30m", 90 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			if err := d.UnmarshalYAML(makeYAMLNode(tt.input)); err != nil {
				t.Fatalf("UnmarshalYAML failed: %v", err)
			}
			if d.Duration != tt.expected {
				t.Errorf("duration = %v, want %v", d.Duration, tt.expected)
			}
		})
	}
}

// TestDuration_UnmarshalYAML_Invalid проверяет ошибку при невалидном значении.
func TestDuration_UnmarshalYAML_Invalid(t *testing.T) {
	var d Duration
	err := d.UnmarshalYAML(makeYAMLNode("not-a-duration"))
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

// makeYAMLNode создаёт скалярный узел для тестов.
func makeYAMLNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
		Tag:   "!!str",
	}
}

// TestConfig_OverrideFromEnv_TrustedProxies проверяет парсинг из окружения.
func TestConfig_OverrideFromEnv_TrustedProxies(t *testing.T) {
	t.Setenv("AUTH_TRUSTED_PROXIES", "127.0.0.1, 10.88.0.0/16, ::1")
	cfg := validTestConfig()
	cfg.overrideFromEnv()
	if len(cfg.Auth.TrustedProxies) != 3 {
		t.Fatalf("expected 3 proxies, got %d", len(cfg.Auth.TrustedProxies))
	}
}

// TestConfig_Validate_InvalidTrustedProxy проверяет ошибку при невалидном прокси.
func TestConfig_Validate_InvalidTrustedProxy(t *testing.T) {
	cfg := validTestConfig()
	cfg.Auth.TrustedProxies = []string{"not-an-ip"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid trusted proxy")
	}
}
