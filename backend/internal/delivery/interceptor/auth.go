// Package interceptor содержит Connect-интерцепторы (аутентификация и др.).
package interceptor

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
	"codeberg.org/Thr0TT1e/emh/backend/internal/netutil"
)

type ctxKey int

const claimsCtxKey ctxKey = iota

// APIKeyAuthenticator минимальный интерфейс для валидации API-ключей.
// Позволяет interceptor не зависеть от полного APIKeyUseCase.
type APIKeyAuthenticator interface {
	Authenticate(ctx context.Context, fullKey string, ip string) (*domain.APIKey, error)
}

// AuthAuditor интерфейс для асинхронной записи аудита аутентификации.
type AuthAuditor interface {
	Enqueue(entry domain.AuthAuditEntry)
}

// AuthAlerter интерфейс для фиксации неудачных попыток аутентификации.
type AuthAlerter interface {
	RecordFailure(ip string)
}

// AuthInterceptor проверяет аутентификацию для защищённых сервисов.
// Механизмы: статические ключи (emergency), управляемые ключи из БД, JWT.
type AuthInterceptor struct {
	jwtSecret string
	// staticKeys: полный ключ → стабильный идентификатор для аудита.
	// Идентификатор — усечённый SHA-256 хеш ключа, не раскрывает секрет.
	staticKeys   map[string]string
	apiKeyAuth   APIKeyAuthenticator
	proxyChecker *netutil.ProxyChecker
	auditor      AuthAuditor
	alerter      AuthAlerter
	ipPepper     string
	logger       *slog.Logger
}

// NewAuthInterceptor создаёт интерцептор аутентификации.
//
// trustedProxies — список доверенных прокси в формате точных IP
// или CIDR-диапазонов. Примеры:
//
//	"127.0.0.1", "::1", "10.88.0.0/16", "192.168.1.100"
//
// Заголовки X-Forwarded-For / X-Real-IP используются только если запрос
// пришёл от одного из этих адресов.
func NewAuthInterceptor(
	jwtSecret string,
	staticKeys []string,
	apiKeyAuth APIKeyAuthenticator,
	trustedProxies []string,
	auditor AuthAuditor,
	alerter AuthAlerter,
	ipPepper string,
	logger *slog.Logger,
) *AuthInterceptor {
	keys := make(map[string]string, len(staticKeys))
	for _, k := range staticKeys {
		if k != "" {
			keys[k] = staticKeyIdentity(k)
		}
	}

	return &AuthInterceptor{
		jwtSecret:    jwtSecret,
		staticKeys:   keys,
		apiKeyAuth:   apiKeyAuth,
		proxyChecker: netutil.NewProxyChecker(trustedProxies),
		auditor:      auditor,
		alerter:      alerter,
		ipPepper:     ipPepper,
		logger:       logger,
	}
}

// WrapUnary реализует проверку аутентификации для unary-методов.
func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Peer() возвращает значение connect.Peer (не указатель).
		// Если адрес недоступен, Addr будет пустой строкой.
		remoteAddr := req.Peer().Addr

		userAgent := req.Header().Get("User-Agent")
		procedure := req.Spec().Procedure

		ctx, err := i.authenticate(ctx, req.Header(), remoteAddr, userAgent, procedure)
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient passthrough — стриминг в проекте не используется.
func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler passthrough — стриминг в проекте не используется.
func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// authenticate извлекает и валидирует credentials из заголовка Authorization.
func (i *AuthInterceptor) authenticate(
	ctx context.Context,
	header http.Header,
	remoteAddr string,
	userAgent string,
	procedure string,
) (context.Context, error) {
	// Унифицированное извлечение IP клиента для всех механизмов.
	// Учитывает trusted proxies (X-Forwarded-For) или использует реальный адрес.
	clientIP := i.extractClientIP(header, remoteAddr)

	token := extractBearerToken(header)
	if token == "" {
		i.recordFailure(clientIP, userAgent, procedure, domain.AuthMechanismNone, "missing token")
		return ctx, connect.NewError(connect.CodeUnauthenticated, errors.New("missing or malformed authorization header"))
	}

	// 1. Статические ключи (аварийный доступ, не требуют БД).
	// Логируем с WARN-уровнем: аварийный доступ должен быть заметен.
	// В аудит передаём стабильный идентификатор ключа, чтобы различать
	// разные статические ключи без раскрытия самого секрета.
	if keyID, ok := i.staticKeys[token]; ok {
		i.logger.WarnContext(ctx, "static api key authentication",
			"key_id", keyID,
			"ip", clientIP,
			"procedure", procedure,
		)
		i.recordSuccess(clientIP, userAgent, procedure, domain.AuthMechanismStaticKey, keyID)
		return ctx, nil
	}

	// 2. Управляемый API-ключ из БД (определяется по префиксу)
	if auth.IsAPIKey(token) {
		key, err := i.apiKeyAuth.Authenticate(ctx, token, clientIP)
		if err != nil {
			i.recordFailure(clientIP, userAgent, procedure, domain.AuthMechanismAPIKey, "invalid or revoked api key")
			return ctx, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid or revoked api key"))
		}
		if key.Role != "admin" {
			i.recordFailure(clientIP, userAgent, procedure, domain.AuthMechanismAPIKey, "insufficient role")
			return ctx, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
		}
		// Сохраняем идентичность ключа в контекст для аудита
		ctx = context.WithValue(ctx, claimsCtxKey, &auth.Claims{
			Username: key.Name,
			Role:     key.Role,
		})
		i.recordSuccess(clientIP, userAgent, procedure, domain.AuthMechanismAPIKey, key.Name)
		return ctx, nil
	}

	// 3. JWT (админы)
	claims, err := auth.ValidateToken(token, i.jwtSecret)
	if err != nil {
		i.recordFailure(clientIP, userAgent, procedure, domain.AuthMechanismJWT, "invalid or expired token")
		return ctx, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid or expired token"))
	}
	if claims.Role != "admin" {
		i.recordFailure(clientIP, userAgent, procedure, domain.AuthMechanismJWT, "insufficient role")
		return ctx, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	i.recordSuccess(clientIP, userAgent, procedure, domain.AuthMechanismJWT, claims.Username)
	return context.WithValue(ctx, claimsCtxKey, claims), nil
}

// extractClientIP извлекает IP клиента с учётом trusted proxies.
//
// Если соединение пришло от доверенного прокси (Caddy), используется
// X-Forwarded-For (метод rightmost-untrusted, см. netutil.ClientIP).
// Иначе возвращается реальный адрес соединения, что защищает от подделки.
func (i *AuthInterceptor) extractClientIP(header http.Header, remoteAddr string) string {
	return netutil.ClientIP(i.proxyChecker, header, remoteAddr)
}

// recordSuccess записывает успешную аутентификацию в аудит.
func (i *AuthInterceptor) recordSuccess(ip, userAgent, procedure string, mechanism domain.AuthAuditMechanism, identity string) {
	metrics.AuthAttemptsTotal.WithLabelValues(string(mechanism), "success").Inc()

	if i.auditor == nil {
		return
	}
	i.auditor.Enqueue(domain.AuthAuditEntry{
		IPAddress: hashIP(ip, i.ipPepper),
		Mechanism: mechanism,
		Result:    domain.AuthResultSuccess,
		Identity:  identity,
		Procedure: procedure,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	})
}

// recordFailure записывает неудачную аутентификацию в аудит и алерт-сервис.
func (i *AuthInterceptor) recordFailure(ip, userAgent, procedure string, mechanism domain.AuthAuditMechanism, reason string) {
	metrics.AuthAttemptsTotal.WithLabelValues(string(mechanism), "failure").Inc()

	if i.alerter != nil {
		i.alerter.RecordFailure(ip)
	}
	if i.auditor == nil {
		return
	}
	i.auditor.Enqueue(domain.AuthAuditEntry{
		IPAddress:     hashIP(ip, i.ipPepper),
		Mechanism:     mechanism,
		Result:        domain.AuthResultFailure,
		Procedure:     procedure,
		FailureReason: reason,
		UserAgent:     userAgent,
		Timestamp:     time.Now(),
	})
}

// extractBearerToken извлекает токен из "Authorization: Bearer <token>".
func extractBearerToken(header http.Header) string {
	value := header.Get("Authorization")
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// ClaimsFromContext возвращает JWT claims из контекста (для аудита в хендлерах).
func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsCtxKey).(*auth.Claims)
	return claims, ok
}

// staticKeyIdentity генерирует стабильный идентификатор статического ключа
// для аудита и логирования. Используется усечённый SHA-256 хеш ключа,
// который позволяет различать разные ключи в логах, но не раскрывает секрет.
//
// Формат: "static-<первые 12 hex-символов хеша>", например "static-a3f8b2c1d4e5".
//
// Свойства:
//   - Детерминированность: один и тот же ключ всегда даёт один идентификатор,
//     что позволяет коррелировать записи аудита между сессиями.
//   - Необратимость: по идентификатору невозможно восстановить ключ.
//   - Различимость: разные ключи дают разные идентификаторы
//     (12 символов = 48 бит, коллизия для небольшого числа ключей практически невозможна).
func staticKeyIdentity(key string) string {
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("static-%s", hex.EncodeToString(h[:6]))
}

func hashIP(ip, pepper string) string {
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

// ContextWithClaims добавляет JWT claims в контекст.
// Используется в тестах usecase-слоя, где нет реального AuthInterceptor.
func ContextWithClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsCtxKey, claims)
}
