// Package config отвечает за загрузку и валидацию конфигурации приложения.
package config

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"

	"codeberg.org/Thr0TT1e/emh/backend/pkg/env"
)

// Config корневая структура конфигурации приложения.
type Config struct {
	App              AppConfig              `yaml:"app"`
	DB               DBConfig               `yaml:"db"`
	S3               S3Config               `yaml:"s3"`
	Auth             AuthConfig             `yaml:"auth"`
	Thumbnails       ThumbnailConfig        `yaml:"thumbnails"`
	RateLimit        RateLimitConfig        `yaml:"rate_limit"`
	OrphanCleanup    OrphanCleanupConfig    `yaml:"orphan_cleanup"`
	AuthAudit        AuthAuditConfig        `yaml:"auth_audit"`
	AuthAlerts       AuthAlertsConfig       `yaml:"auth_alerts"`
	Contact          ContactConfig          `yaml:"contact"`
	Telegram         TelegramConfig         `yaml:"telegram"`
	LLM              LLMConfig              `yaml:"llm"`
	Metrics          MetricsConfig          `yaml:"metrics"`
	SubmissionEmails SubmissionEmailsConfig `yaml:"submission_emails"`
}

// AppConfig параметры самого приложения.
type AppConfig struct {
	// Name - имя приложения (используется в логах и метриках).
	Name string `yaml:"name"`
	// Addr - адрес HTTP-сервера, например ":3480".
	Addr string `yaml:"addr"`
	// LogLevel - уровень логирования: debug, info, warn, error.
	LogLevel string `yaml:"log_level"`
	// CORSOrigins - разрешённые источники для CORS.
	CORSOrigins []string `yaml:"cors_origins"`
}

// DBConfig параметры подключения к PostgreSQL.
type DBConfig struct {
	// URL - строка подключения (переопределяется через DATABASE_URL).
	URL string `yaml:"url"`
	// MaxConns - максимальный размер пула соединений.
	MaxConns int32 `yaml:"max_conns"`
	// MinConns - минимальный размер пула соединений.
	MinConns int32 `yaml:"min_conns"`
}

// S3Config параметры объектного хранилища (MinIO/S3).
type S3Config struct {
	// Endpoint - внутренний адрес для серверных операций.
	Endpoint string `yaml:"endpoint"`
	// PublicEndpoint - внешний адрес для генерации presigned URL.
	PublicEndpoint string `yaml:"public_endpoint"`
	// PublicBaseURL - базовый URL для публичных ссылок.
	PublicBaseURL string `yaml:"public_base_url"`
	// AccessKey - ключ доступа (переопределяется через MINIO_ROOT_USER).
	AccessKey string `yaml:"access_key"`
	// SecretKey - секретный ключ (переопределяется через MINIO_ROOT_PASSWORD).
	SecretKey string `yaml:"secret_key"`
	// Bucket - имя бакета для медиа-контента.
	Bucket string `yaml:"bucket"`
	// UseSSL - использовать TLS для подключения.
	UseSSL bool `yaml:"use_ssl"`
	// PresignExpiry - время жизни presigned URL.
	PresignExpiry Duration `yaml:"presign_expiry"`
	// CACertPath — путь к CA-сертификату для самоподписанного TLS.
	CACertPath string `yaml:"ca_cert"`
}

// Duration обёртка над time.Duration для парсинга из YAML-строки вида "15m".
// Стандартный yaml.v3 не умеет парсить time.Duration, поэтому нужен кастомный тип.
type Duration struct {
	time.Duration
}

// AuthConfig параметры аутентификации и авторизации.
type AuthConfig struct {
	// JWTSecret - секрет для подписи JWT (минимум 32 символа). Из env AUTH_JWT_SECRET.
	JWTSecret string `yaml:"jwt_secret"`
	// JWTExpiry - время жизни токена.
	JWTExpiry Duration `yaml:"jwt_expiry"`
	// APIKeys - статические ключи для сервисных вызовов. Из env AUTH_API_KEYS (через запятую).
	RefreshExpiry Duration `yaml:"refresh_expiry"`
	// APIKeys - статические ключи для сервисных вызовов.
	APIKeys []string `yaml:"api_keys"`
	// Admins - администраторы (для MVP в конфиге, для продакшена — в БД).
	Admins []AdminConfig `yaml:"admins"`
	// TrustedProxies — IP-адреса доверенных прокси (Caddy).
	// Заголовки X-Forwarded-For используются только для запросов от этих IP.
	TrustedProxies []string `yaml:"trusted_proxies"`
}

// AdminConfig учётная запись администратора.
type AdminConfig struct {
	// Username - имя пользователя.
	Username string `yaml:"username"`
	// PasswordHash - bcrypt-хеш пароля (генерируется через cmd/hashpassword).
	PasswordHash string `yaml:"password_hash"`
	// Role - роль (admin).
	Role string `yaml:"role"`
}

// ThumbnailConfig параметры генерации превью.
type ThumbnailConfig struct {
	// Width - ширина thumbnail в пикселях.
	Width int `yaml:"width"`
	// Height - высота thumbnail в пикселях.
	Height int `yaml:"height"`
	// Quality - качество WebP (1-100).
	Quality int `yaml:"quality"`
	// OriginalQuality - качество WebP для конвертации оригиналов (1-100).
	OriginalQuality int `yaml:"original_quality"`
	// Workers - количество параллельных воркеров.
	Workers int `yaml:"workers"`
	// QueueSize - размер буфера очереди задач.
	QueueSize int `yaml:"queue_size"`
}

// RateLimitConfig параметры ограничения запросов (token bucket per IP).
type RateLimitConfig struct {
	// Enabled - включает/отключает rate-limiting глобально.
	Enabled bool `yaml:"enabled"`

	// --- Лимиты для AuthService (Login, Refresh, Logout) ---
	// AuthLimit - максимальное число запросов в окне AuthInterval.
	AuthLimit int `yaml:"auth_limit"`
	// AuthInterval - временное окно для Auth-группы.
	AuthInterval Duration `yaml:"auth_interval"`

	// --- Лимиты для SubmissionService (CreateSubmission) ---
	// SubmissionLimit - максимальное число запросов в окне SubmissionInterval.
	SubmissionLimit int `yaml:"submission_limit"`
	// SubmissionInterval - временное окно для Submission-группы.
	SubmissionInterval Duration `yaml:"submission_interval"`

	// --- Лимиты для MediaService (GetUploadUrl) ---
	// MediaLimit - максимальное число запросов в окне MediaInterval.
	MediaLimit int `yaml:"media_limit"`
	// MediaInterval - временное окно для Media-группы.
	MediaInterval Duration `yaml:"media_interval"`

	// --- Лимиты для публичных read-сервисов (Get*, List*) ---
	// ReadLimit - максимальное число запросов в окне ReadInterval.
	ReadLimit int `yaml:"read_limit"`
	// ReadInterval - временное окно для Read-группы.
	ReadInterval Duration `yaml:"read_interval"`

	// --- Лимиты для *AdminService ---
	// Раньше admin-сервисы полностью исключались из rate-limiting в расчёте
	// на защиту JWT/API-ключом, но это позволяло жечь CPU на bcrypt при
	// известном keyID/токене без ограничений и брутфорсить секрет без lockout.
	// AdminLimit - максимальное число запросов в окне AdminInterval.
	AdminLimit int `yaml:"admin_limit"`
	// AdminInterval - временное окно для Admin-группы.
	AdminInterval Duration `yaml:"admin_interval"`

	// --- Параметры in-memory хранилища лимитеров ---
	// EntryTTL - время жизни записи лимитера без активности (защита от утечки памяти).
	EntryTTL Duration `yaml:"entry_ttl"`
	// CleanupInterval - интервал фоновой горутины очистки записей.
	CleanupInterval Duration `yaml:"cleanup_interval"`
}

type OrphanCleanupConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Interval    Duration `yaml:"interval"`
	GracePeriod Duration `yaml:"grace_period"`
	DryRun      bool     `yaml:"dry_run"`
}

type AuthAuditConfig struct {
	Enabled         bool     `yaml:"enabled"`
	QueueSize       int      `yaml:"queue_size"`
	Retention       Duration `yaml:"retention"`
	CleanupInterval Duration `yaml:"cleanup_interval"`
}

type AuthAlertsConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Threshold int      `yaml:"threshold"`
	Window    Duration `yaml:"window"`
	QueueSize int      `yaml:"queue_size"`
}

// ContactConfig параметры сервиса обратной связи и SMTP-уведомлений.
type ContactConfig struct {
	// SMTPEnabled — включена ли отправка email через SMTP.
	SMTPEnabled bool `yaml:"smtp_enabled"`
	// SMTPHost — адрес SMTP-сервера. Из env EMH_CONTACT_SMTP_HOST.
	SMTPHost string `yaml:"smtp_host"`
	// SMTPPort — порт SMTP-сервера (25, 465, 587). Из env EMH_CONTACT_SMTP_PORT.
	SMTPPort int `yaml:"smtp_port"`
	// SMTPUsername — логин для аутентификации. Из env EMH_CONTACT_SMTP_USERNAME.
	SMTPUsername string `yaml:"smtp_username"`
	// SMTPPassword — пароль для аутентификации. Из env EMH_CONTACT_SMTP_PASSWORD.
	SMTPPassword string `yaml:"smtp_password"`
	// SMTPFromName — отображаемое имя отправителя. Из env EMH_CONTACT_SMTP_FROM_NAME.
	SMTPFromName string `yaml:"smtp_from_name"`
	// SMTPFromEmail — email отправителя. Из env EMH_CONTACT_SMTP_FROM_EMAIL.
	SMTPFromEmail string `yaml:"smtp_from_email"`
	// SMTPTo — email получателя (владелец проекта). Из env EMH_CONTACT_SMTP_TO.
	SMTPTo string `yaml:"smtp_to"`
	// SMTPTimeout — таймаут на подключение и SMTP-операции.
	SMTPTimeout Duration `yaml:"smtp_timeout"`
	// IPPepper — секрет для хеширования IP (HMAC-SHA256). Из env EMH_CONTACT_IP_PEPPER.
	IPPepper string `yaml:"ip_pepper"`
	// RateLimitHour — максимум сообщений в час (на один IP или email).
	RateLimitHour int `yaml:"rate_limit_hour"`
	// RateLimitDay — максимум сообщений в сутки (на один IP или email).
	RateLimitDay int `yaml:"rate_limit_day"`
	// DedupeWindow — окно анти-дубликата (по умолчанию 5 минут).
	DedupeWindow Duration `yaml:"dedupe_window"`
	// EntryTTL — время жизни записи rate limiter'а без активности.
	EntryTTL Duration `yaml:"entry_ttl"`
	// CleanupInterval — интервал фоновой очистки rate limiter'а.
	CleanupInterval Duration `yaml:"cleanup_interval"`
	// SMTPMaxRetries — максимум повторных попыток отправки (по умолчанию 3).
	SMTPMaxRetries int `yaml:"smtp_max_retries"`
	// SMTPBaseDelay — начальная задержка между попытками (по умолчанию 5s).
	SMTPBaseDelay Duration `yaml:"smtp_base_delay"`
	// SMTPMaxDelay — максимальная задержка между попытками (по умолчанию 60s).
	SMTPMaxDelay Duration `yaml:"smtp_max_delay"`
}

// SubmissionEmailsConfig параметры email-уведомлений заявителей.
type SubmissionEmailsConfig struct {
	Enabled         bool   `yaml:"enabled"`
	FromName        string `yaml:"from_name"`
	ApproveSubject  string `yaml:"approve_subject"`
	RejectSubject   string `yaml:"reject_subject"`
	ApproveTemplate string `yaml:"approve_template"`
	RejectTemplate  string `yaml:"reject_template"`
	QueueSize       int    `yaml:"queue_size"`
}

// TelegramConfig параметры отправки алертов через Telegram.
type TelegramConfig struct {
	Enabled  bool     `yaml:"enabled"`
	BotToken string   `yaml:"bot_token"` // из env EMH_TELEGRAM_BOT_TOKEN
	ChatID   string   `yaml:"chat_id"`   // из env EMH_TELEGRAM_CHAT_ID
	Timeout  Duration `yaml:"timeout"`
}

type LLMConfig struct {
	Enabled         bool                `yaml:"enabled"`
	Providers       []LLMProviderConfig `yaml:"providers"`        // Список доступных провайдеров
	DefaultProvider string              `yaml:"default_provider"` // Fallback если в БД нет активного
	MaxInputChars   int                 `yaml:"max_input_chars"`
}

type LLMProviderConfig struct {
	Name        string   `yaml:"name"`        // "ollama_local", "openrouter", "gptunnel", "anthropic"
	Type        string   `yaml:"type"`        // "ollama", "openai_compatible", "anthropic"
	Endpoint    string   `yaml:"endpoint"`    // URL API
	Model       string   `yaml:"model"`       // Название модели
	Timeout     Duration `yaml:"timeout"`     // Таймаут запроса
	MaxTokens   int      `yaml:"max_tokens"`  // Лимит токенов
	Temperature float64  `yaml:"temperature"` // Креативность
	APIKeyEnv   string   `yaml:"api_key_env"` // Имя env-переменной с ключом (например, "OPENROUTER_API_KEY")
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// UnmarshalYAML реализует интерфейс yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = parsed
	return nil
}

// MarshalYAML реализует интерфейс yaml.Marshaler (для симметрии).
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

// Load загружает конфигурацию из YAML-файла, применяет значения по умолчанию,
// переопределяет секреты из окружения и валидирует результат.
func Load(path string) (*Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", path, err)
	}

	cfg.overrideFromEnv()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// defaultConfig возвращает конфигурацию со значениями по умолчанию.
// Значения из YAML-файла накладываются поверх этих.
func defaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "emh",
			Addr:        ":3480",
			LogLevel:    "info",
			CORSOrigins: []string{"http://localhost:3000"},
		},
		DB: DBConfig{
			URL:      "postgres://hero_db:secret@postgres_emh:5432/heroes?sslmode=disable",
			MaxConns: 20,
			MinConns: 2,
		},
		S3: S3Config{
			Endpoint:       "minio_emh:9000",
			PublicEndpoint: "localhost:4443",
			PublicBaseURL:  "https://localhost:4443",
			Bucket:         "emh-media",
			UseSSL:         true,
			PresignExpiry:  Duration{15 * time.Minute},
		},
		Auth: AuthConfig{
			JWTExpiry:     Duration{15 * time.Minute},
			RefreshExpiry: Duration{168 * time.Hour},
		},
		AuthAlerts: AuthAlertsConfig{
			Enabled:   true,
			Threshold: 10,
			Window:    Duration{5 * time.Minute},
			QueueSize: 100,
		},
		AuthAudit: AuthAuditConfig{
			Enabled:         true,
			QueueSize:       1000,
			Retention:       Duration{90 * 24 * time.Hour},
			CleanupInterval: Duration{24 * time.Hour},
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
			SMTPPort:        587,
			SMTPFromName:    "Вечная память героям",
			SMTPTimeout:     Duration{10 * time.Second},
			RateLimitHour:   5,
			RateLimitDay:    20,
			DedupeWindow:    Duration{5 * time.Minute},
			EntryTTL:        Duration{24 * time.Hour},
			CleanupInterval: Duration{10 * time.Minute},
			SMTPMaxRetries:  3,
			SMTPBaseDelay:   Duration{5 * time.Second},
			SMTPMaxDelay:    Duration{60 * time.Second},
		},
		Telegram: TelegramConfig{
			Enabled: false,
			Timeout: Duration{10 * time.Second},
		},
		SubmissionEmails: SubmissionEmailsConfig{
			Enabled:        false,
			FromName:       "Вечная память героям",
			ApproveSubject: "Ваша заявка одобрена",
			RejectSubject:  "Ваша заявка отклонена",
			ApproveTemplate: `Здравствуйте, {{.SubmitterName}}!

		Ваша заявка на добавление данных о герое одобрена модератором.

		ID заявки: {{.SubmissionID}}
		Дата: {{.Timestamp}}

		Спасибо за ваш вклад в сохранение памяти о героях!

		С уважением,
		Команда проекта "Вечная память героям"`,
			RejectTemplate: `Здравствуйте, {{.SubmitterName}}!

		К сожалению, ваша заявка на добавление данных о герое отклонена.

		Причина: {{.ModeratorComment}}

		ID заявки: {{.SubmissionID}}
		Дата: {{.Timestamp}}

		Если вы считаете, что это ошибка, свяжитесь с нами через форму обратной связи.

		С уважением,
		Команда проекта "Вечная память героям"`,
			QueueSize: 100,
		},
		LLM: LLMConfig{
			Enabled:         false,
			DefaultProvider: "ollama_local",
			Providers: []LLMProviderConfig{
				{
					Name:        "ollama_local",
					Type:        "ollama",
					Endpoint:    "http://localhost:11434",
					Model:       "llama3.1:8b",
					Timeout:     Duration{60 * time.Second},
					MaxTokens:   4096,
					Temperature: 0.1,
					APIKeyEnv:   "",
				},
			},
			MaxInputChars: 150_000,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
}

// overrideFromEnv переопределяет чувствительные и критичные значения из окружения.
// Секреты никогда не должны храниться в YAML — только в env.
func (c *Config) overrideFromEnv() {
	c.App.Addr = env.Get("APP_ADDR", c.App.Addr)
	c.App.LogLevel = env.Get("LOG_LEVEL", c.App.LogLevel)

	c.DB.URL = env.Get("DATABASE_URL", c.DB.URL)

	c.S3.AccessKey = env.Get("MINIO_ROOT_USER", c.S3.AccessKey)
	c.S3.SecretKey = env.Get("MINIO_ROOT_PASSWORD", c.S3.SecretKey)
	c.S3.PublicBaseURL = env.Get("MINIO_PUBLIC_BASE_URL", c.S3.PublicBaseURL)
	c.S3.CACertPath = env.Get("S3_CA_CERT", c.S3.CACertPath)
	c.Auth.JWTSecret = env.Get("AUTH_JWT_SECRET", c.Auth.JWTSecret)
	if keys := env.Get("AUTH_API_KEYS", ""); keys != "" {
		c.Auth.APIKeys = strings.Split(keys, ",")
	}
	// Trusted proxies (через запятую, например "127.0.0.1,10.88.0.0/16")
	if proxies := env.Get("AUTH_TRUSTED_PROXIES", ""); proxies != "" {
		parts := strings.Split(proxies, ",")
		c.Auth.TrustedProxies = make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				c.Auth.TrustedProxies = append(c.Auth.TrustedProxies, p)
			}
		}
	}
	if hash := env.Get("AUTH_ADMIN_THREET_HASH", ""); hash != "" {
		for i := range c.Auth.Admins {
			if c.Auth.Admins[i].Username == "threet" && c.Auth.Admins[i].PasswordHash == "" {
				c.Auth.Admins[i].PasswordHash = hash
			}
		}
	}
	if v := env.GetInt("EMH_AUTH_ALERTS_QUEUE_SIZE", 0); v > 0 {
		c.AuthAlerts.QueueSize = v
	}

	// Auth audit cleanup
	if v := env.Get("EMH_AUTH_AUDIT_RETENTION", ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.AuthAudit.Retention = Duration{d}
		}
	}
	if v := env.Get("EMH_AUTH_AUDIT_CLEANUP_INTERVAL", ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.AuthAudit.CleanupInterval = Duration{d}
		}
	}

	// Contact / SMTP
	c.Contact.SMTPEnabled = env.GetBool("EMH_CONTACT_SMTP_ENABLED", c.Contact.SMTPEnabled)
	c.Contact.SMTPHost = env.Get("EMH_CONTACT_SMTP_HOST", c.Contact.SMTPHost)
	c.Contact.SMTPPort = env.GetInt("EMH_CONTACT_SMTP_PORT", c.Contact.SMTPPort)
	c.Contact.SMTPUsername = env.Get("EMH_CONTACT_SMTP_USERNAME", c.Contact.SMTPUsername)
	c.Contact.SMTPPassword = env.Get("EMH_CONTACT_SMTP_PASSWORD", c.Contact.SMTPPassword)
	c.Contact.SMTPFromName = env.Get("EMH_CONTACT_SMTP_FROM_NAME", c.Contact.SMTPFromName)
	c.Contact.SMTPFromEmail = env.Get("EMH_CONTACT_SMTP_FROM_EMAIL", c.Contact.SMTPFromEmail)
	c.Contact.SMTPTo = env.Get("EMH_CONTACT_SMTP_TO", c.Contact.SMTPTo)
	c.Contact.IPPepper = env.Get("EMH_CONTACT_IP_PEPPER", c.Contact.IPPepper)
	// SMTP ретраи
	if retries := env.Get("EMH_CONTACT_SMTP_MAX_RETRIES", ""); retries != "" {
		c.Contact.SMTPMaxRetries = env.GetInt("EMH_CONTACT_SMTP_MAX_RETRIES", c.Contact.SMTPMaxRetries)
	}
	if baseDelay := env.Get("EMH_CONTACT_SMTP_BASE_DELAY", ""); baseDelay != "" {
		if d, err := time.ParseDuration(baseDelay); err == nil {
			c.Contact.SMTPBaseDelay = Duration{d}
		}
	}
	if maxDelay := env.Get("EMH_CONTACT_SMTP_MAX_DELAY", ""); maxDelay != "" {
		if d, err := time.ParseDuration(maxDelay); err == nil {
			c.Contact.SMTPMaxDelay = Duration{d}
		}
	}

	// Submission emails
	c.SubmissionEmails.Enabled = env.GetBool("EMH_SUBMISSION_EMAILS_ENABLED", c.SubmissionEmails.Enabled)
	if v := env.GetInt("EMH_SUBMISSION_EMAILS_QUEUE_SIZE", 0); v > 0 {
		c.SubmissionEmails.QueueSize = v
	}

	// Telegram
	c.Telegram.Enabled = env.GetBool("EMH_TELEGRAM_ENABLED", c.Telegram.Enabled)
	c.Telegram.BotToken = env.Get("EMH_TELEGRAM_BOT_TOKEN", c.Telegram.BotToken)
	c.Telegram.ChatID = env.Get("EMH_TELEGRAM_CHAT_ID", c.Telegram.ChatID)

	// LLM
	c.LLM.Enabled = env.GetBool("EMH_LLM_ENABLED", c.LLM.Enabled)
	c.LLM.DefaultProvider = env.Get("EMH_LLM_DEFAULT_PROVIDER", c.LLM.DefaultProvider)
	if v := env.GetInt("EMH_LLM_MAX_INPUT_CHARS", 0); v > 0 {
		c.LLM.MaxInputChars = v
	}
}

// Validate проверяет обязательные поля конфигурации.
func (c *Config) Validate() error {
	if c.App.Addr == "" {
		return fmt.Errorf("app.addr is required")
	}
	if c.DB.URL == "" {
		return fmt.Errorf("db.url is required")
	}
	if c.DB.MaxConns <= 0 {
		return fmt.Errorf("db.max_conns must be positive")
	}
	if c.S3.Endpoint == "" {
		return fmt.Errorf("s3.endpoint is required")
	}
	if c.S3.Bucket == "" {
		return fmt.Errorf("s3.bucket is required")
	}
	if c.S3.AccessKey == "" || c.S3.SecretKey == "" {
		return fmt.Errorf("s3 credentials are required (set MINIO_ROOT_USER and MINIO_ROOT_PASSWORD)")
	}
	if c.S3.PresignExpiry.Duration <= 0 {
		return fmt.Errorf("s3.presign_expiry must be positive")
	}
	// --- Валидация JWT_SECRET (задача #11) ---
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("auth.jwt_secret must be at least 32 characters (set AUTH_JWT_SECRET)")
	}
	// --- Валидация TrustedProxies ---
	// Невалидные записи (ни точный IP, ни CIDR) отклоняются при старте,
	// чтобы не получить тихо неработающий прокси.
	for _, p := range c.Auth.TrustedProxies {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "/") {
			if _, _, err := net.ParseCIDR(p); err != nil {
				return fmt.Errorf("auth.trusted_proxies %q: invalid CIDR: %w", p, err)
			}
			continue
		}
		if net.ParseIP(p) == nil {
			return fmt.Errorf("auth.trusted_proxies %q: not a valid IP or CIDR", p)
		}
	}
	// Пробелы в секрете — частая ошибка копирования из .env
	if strings.ContainsAny(c.Auth.JWTSecret, " \t\n\r") {
		return fmt.Errorf("auth.jwt_secret must not contain whitespace")
	}
	for _, admin := range c.Auth.Admins {
		if admin.PasswordHash == "" {
			return fmt.Errorf("auth.admins[%s].password_hash is empty (set via env)", admin.Username)
		}
	}
	// Защита от дефолтных небезопасных значений
	weakJWTSecrets := map[string]bool{
		// Короткие слабые секреты
		"changeme":    true,
		"change-me":   true,
		"change_me":   true,
		"secret":      true,
		"test-secret": true,
		"dev-secret":  true,
		"jwt-secret":  true,
		"jwt_secret":  true,
		"jwtsecret":   true,
		// Длинные варианты — защита от попыток "удлинить" слабый секрет повторением.
		// Все они >= 32 символов, поэтому проходят проверку длины и отклоняются здесь.
		"changeme-changeme-changeme-changeme":         true,
		"change-me-change-me-change-me-change-me":     true,
		"change_me_change_me_change_me_change_me":     true,
		"secret-secret-secret-secret-secret-secret":   true,
		"test-secret-test-secret-test-secret-test":    true,
		"dev-secret-dev-secret-dev-secret-dev-secret": true,
		"jwt-secret-jwt-secret-jwt-secret-jwt-secret": true,
	}
	lowerSecret := strings.ToLower(c.Auth.JWTSecret)
	if weakJWTSecrets[lowerSecret] {
		return fmt.Errorf("auth.jwt_secret is a known weak value, generate with `just jwt-secret`")
	}
	// Проверка на повторяющиеся паттерны (например, "changeme-changeme-changeme")
	// Это защищает от попыток "удлинить" слабый секрет простым повторением.
	// Проверяем, состоит ли секрет ТОЛЬКО из повторений слабого слова (с разделителями).
	// Это не отклонит валидные секреты типа "test-jwt-secret-at-least-32-chars",
	// которые содержат "secret" как часть более длинной случайной строки.
	for base := range weakJWTSecrets {
		// Только короткие базы имеют смысл (длинные уже покрыты выше)
		if len(base) >= 32 {
			continue
		}
		stripped := strings.ReplaceAll(lowerSecret, "-", "")
		stripped = strings.ReplaceAll(stripped, "_", "")
		if len(stripped) > 0 && len(base) > 0 && len(stripped)%len(base) == 0 {
			repetitions := len(stripped) / len(base)
			if repetitions > 1 {
				expected := strings.Repeat(base, repetitions)
				if stripped == expected {
					return fmt.Errorf("auth.jwt_secret is composed of repeated weak pattern '%s', generate with `just jwt-secret`", base)
				}
			}
		}
	}
	if len(c.Auth.Admins) == 0 {
		return fmt.Errorf("auth.admins must contain at least one admin")
	}
	if c.Auth.RefreshExpiry.Duration <= 0 {
		return fmt.Errorf("auth.refresh_expiry must be positive")
	}
	if c.Thumbnails.Width <= 0 || c.Thumbnails.Height <= 0 {
		return fmt.Errorf("thumbnails.width and thumbnails.height must be positive")
	}
	if c.Thumbnails.Quality < 1 || c.Thumbnails.Quality > 100 {
		return fmt.Errorf("thumbnails.quality must be between 1 and 100")
	}
	if c.Thumbnails.OriginalQuality < 1 || c.Thumbnails.OriginalQuality > 100 {
		return fmt.Errorf("thumbnails.original_quality must be between 1 and 100")
	}
	if c.Thumbnails.Workers <= 0 {
		return fmt.Errorf("thumbnails.workers must be positive")
	}
	if c.Thumbnails.QueueSize <= 0 {
		return fmt.Errorf("thumbnails.queue_size must be positive")
	}
	// Rate-limit
	if c.RateLimit.Enabled {
		if c.RateLimit.AuthLimit <= 0 {
			return fmt.Errorf("rate_limit.auth_limit must be positive")
		}
		if c.RateLimit.AuthInterval.Duration <= 0 {
			return fmt.Errorf("rate_limit.auth_interval must be positive")
		}
		if c.RateLimit.SubmissionLimit <= 0 {
			return fmt.Errorf("rate_limit.submission_limit must be positive")
		}
		if c.RateLimit.SubmissionInterval.Duration <= 0 {
			return fmt.Errorf("rate_limit.submission_interval must be positive")
		}
		if c.RateLimit.MediaLimit <= 0 {
			return fmt.Errorf("rate_limit.media_limit must be positive")
		}
		if c.RateLimit.MediaInterval.Duration <= 0 {
			return fmt.Errorf("rate_limit.media_interval must be positive")
		}
		if c.RateLimit.ReadLimit <= 0 {
			return fmt.Errorf("rate_limit.read_limit must be positive")
		}
		if c.RateLimit.ReadInterval.Duration <= 0 {
			return fmt.Errorf("rate_limit.read_interval must be positive")
		}
		if c.RateLimit.AdminLimit <= 0 {
			return fmt.Errorf("rate_limit.admin_limit must be positive")
		}
		if c.RateLimit.AdminInterval.Duration <= 0 {
			return fmt.Errorf("rate_limit.admin_interval must be positive")
		}
		if c.RateLimit.EntryTTL.Duration <= 0 {
			return fmt.Errorf("rate_limit.entry_ttl must be positive")
		}
		if c.RateLimit.CleanupInterval.Duration <= 0 {
			return fmt.Errorf("rate_limit.cleanup_interval must be positive")
		}
	}

	// --- Валидация AuthAlerts ---
	if c.AuthAlerts.Enabled {
		if c.AuthAlerts.Threshold <= 0 {
			return fmt.Errorf("auth_alerts.threshold must be positive")
		}
		if c.AuthAlerts.Window.Duration <= 0 {
			return fmt.Errorf("auth_alerts.window must be positive")
		}
		if c.AuthAlerts.QueueSize <= 0 {
			return fmt.Errorf("auth_alerts.queue_size must be positive")
		}
	}

	// --- Валидация AuthAudit ---
	if c.AuthAudit.Enabled {
		if c.AuthAudit.QueueSize <= 0 {
			return fmt.Errorf("auth_audit.queue_size must be positive")
		}
		if c.AuthAudit.Retention.Duration <= 0 {
			return fmt.Errorf("auth_audit.retention must be positive")
		}
		if c.AuthAudit.CleanupInterval.Duration <= 0 {
			return fmt.Errorf("auth_audit.cleanup_interval must be positive")
		}
		if c.AuthAudit.Retention.Duration < 24*time.Hour {
			return fmt.Errorf("auth_audit.retention must be at least 24h (got %v)", c.AuthAudit.Retention.Duration)
		}
	}

	// --- Валидация Telegram ---
	if c.Telegram.Enabled {
		if c.Telegram.BotToken == "" {
			return fmt.Errorf("telegram.bot_token is required when telegram enabled (set EMH_TELEGRAM_BOT_TOKEN)")
		}
		if c.Telegram.ChatID == "" {
			return fmt.Errorf("telegram.chat_id is required when telegram enabled (set EMH_TELEGRAM_CHAT_ID)")
		}
		if c.Telegram.Timeout.Duration <= 0 {
			return fmt.Errorf("telegram.timeout must be positive")
		}
	}

	// --- Валидация SubmissionEmails ---
	if c.SubmissionEmails.Enabled {
		if !c.Contact.SMTPEnabled {
			return fmt.Errorf("submission_emails.enabled requires contact.smtp_enabled=true")
		}
		if c.SubmissionEmails.QueueSize <= 0 {
			return fmt.Errorf("submission_emails.queue_size must be positive")
		}
	}

	// Contact / SMTP
	if c.Contact.SMTPEnabled {
		if c.Contact.SMTPHost == "" {
			return fmt.Errorf("contact.smtp_host is required when smtp enabled")
		}
		if c.Contact.SMTPPort <= 0 {
			return fmt.Errorf("contact.smtp_port must be positive")
		}
		if c.Contact.SMTPFromEmail == "" {
			return fmt.Errorf("contact.smtp_from_email is required when smtp enabled")
		}
		if c.Contact.SMTPTo == "" {
			return fmt.Errorf("contact.smtp_to is required when smtp enabled")
		}
		if c.Contact.SMTPTimeout.Duration <= 0 {
			return fmt.Errorf("contact.smtp_timeout must be positive")
		}
	}
	// --- Валидация IPPepper ---
	// hashIP вызывается всегда в ContactUseCase.Submit, независимо от SMTPEnabled.
	// Пустой или короткий ключ ослабляет защиту: хеши можно обратить через
	// rainbow table, что раскрывает реальные клиентские адреса при утечке БД.
	if strings.TrimSpace(c.Contact.IPPepper) == "" {
		return fmt.Errorf("contact.ip_pepper is required (generate with `just ip-pepper`)")
	}
	if len(c.Contact.IPPepper) < 16 {
		return fmt.Errorf("contact.ip_pepper must be at least 16 characters (got %d)", len(c.Contact.IPPepper))
	}
	if c.Contact.RateLimitHour <= 0 {
		return fmt.Errorf("contact.rate_limit_hour must be positive")
	}
	if c.Contact.RateLimitDay <= 0 {
		return fmt.Errorf("contact.rate_limit_day must be positive")
	}
	if c.Contact.DedupeWindow.Duration <= 0 {
		return fmt.Errorf("contact.dedupe_window must be positive")
	}
	if c.LLM.Enabled {
		if c.LLM.MaxInputChars <= 0 {
			return fmt.Errorf("llm.max_input_chars must be positive when LLM is enabled")
		}
		if c.LLM.MaxInputChars > 1_000_000 {
			return fmt.Errorf("llm.max_input_chars is unreasonably large (>1M): got %d", c.LLM.MaxInputChars)
		}
	}

	return nil
}

// LoadMCP загружает конфигурацию для MCP-сервера с минимальной валидацией.
// MCP-сервер использует только DB (PostgreSQL) и App (log_level).
// S3, RateLimit, Thumbnails, Auth, Contact — не валидируются.
func LoadMCP(path string) (*Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", path, err)
	}
	cfg.overrideFromEnv()

	// Минимальная валидация: только то, что нужно MCP
	if err := cfg.validateMCP(); err != nil {
		return nil, fmt.Errorf("validate config (mcp): %w", err)
	}
	return cfg, nil
}

// validateMCP проверяет только обязательные поля для MCP-сервера.
func (c *Config) validateMCP() error {
	if c.DB.URL == "" {
		return fmt.Errorf("db.url is required")
	}
	return nil
}
