// Package smtp реализует отправку email-уведомлений через SMTP.
// Поддерживает STARTTLS (порт 587), implicit TLS (порт 465) и plain (порт 25).
package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"mime"
	"net"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
	"github.com/google/uuid"
)

// Config параметры SMTP-подключения.
type Config struct {
	// Enabled — включена ли отправка. Если false, usecase пропускает отправку.
	Enabled bool
	// Host — адрес SMTP-сервера (например, smtp.yandex.ru).
	Host string
	// Port — порт SMTP-сервера (25, 465, 587).
	Port int
	// Username — логин для аутентификации.
	Username string
	// Password — пароль для аутентификации.
	Password string
	// FromName — отображаемое имя отправителя в заголовке From.
	FromName string
	// FromEmail — email отправителя (Envelope From и заголовок From).
	FromEmail string
	// To — email получателя (владелец проекта).
	To string
	// Timeout — таймаут на подключение и все SMTP-операции.
	Timeout time.Duration
	// MaxRetries — максимум повторных попыток отправки (по умолчанию 3).
	// Итого: 1 начальная + MaxRetries повторных.
	MaxRetries int
	// BaseDelay — начальная задержка между попытками (по умолчанию 5s).
	BaseDelay time.Duration
	// MaxDelay — максимальная задержка между попытками (по умолчанию 60s).
	MaxDelay time.Duration
}

// Message email-сообщение для отправки.
type Message struct {
	// ReplyTo — адрес для ответа (email отправителя формы).
	ReplyTo string
	// Subject — тема письма (без префикса [EMH], добавляется в usecase).
	Subject string
	// Body — текст письма (plain text, UTF-8).
	Body string
}

// Sender интерфейс отправки email-уведомлений.
type Sender interface {
	// Send отправляет email-сообщение.
	// Возвращает ошибку при сбое подключения, аутентификации или передачи.
	Send(ctx context.Context, msg Message) error
}

// smtpSender реализация Sender через стандартный net/smtp.
type smtpSender struct {
	cfg    Config
	logger *slog.Logger
}

// NewSender создаёт SMTP-отправитель с логгером.
// Заполняет параметры ретраев значениями по умолчанию, если не заданы.
func NewSender(cfg Config, logger *slog.Logger) Sender {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 5 * time.Second
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 60 * time.Second
	}
	return &smtpSender{cfg: cfg, logger: logger}
}

// Send отправляет email через SMTP с экспоненциальным ретраем.
//
// Повторяет только при временных ошибках (421, 450-452, сетевые сбои).
// Постоянные ошибки (535, 550-554) не повторяются.
//
// Длительность: максимум ~35 секунд (3 ретрая по 5+10+20 сек).
// Общее время контролируется контекстом.
func (s *smtpSender) Send(ctx context.Context, msg Message) error {
	start := time.Now()
	defer func() {
		metrics.SMTPDurationSeconds.Observe(time.Since(start).Seconds())
	}()

	var lastErr error
	for attempt := 0; attempt <= s.cfg.MaxRetries; attempt++ {
		// Задержка перед повтором (кроме первой попытки)
		if attempt > 0 {
			metrics.SMTPRetriesTotal.Inc()
			delay := backoffDelay(attempt-1, s.cfg.BaseDelay, s.cfg.MaxDelay)
			select {
			case <-ctx.Done():
				metrics.SMTPSendTotal.WithLabelValues("failed").Inc()
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		// Проверка отмены контекста перед каждой попыткой
		select {
		case <-ctx.Done():
			metrics.SMTPSendTotal.WithLabelValues("failed").Inc()
			return ctx.Err()
		default:
		}

		err := s.sendOnce(ctx, msg)
		if err == nil {
			metrics.SMTPSendTotal.WithLabelValues("success").Inc()
			return nil
		}

		lastErr = err

		// Постоянные ошибки не повторяются
		if !isRetryableSMTPError(err) {
			metrics.SMTPSendTotal.WithLabelValues("failed").Inc()
			return fmt.Errorf("smtp send (permanent error): %w", err)
		}

		s.logger.Debug("smtp send retryable error, will retry",
			"attempt", attempt+1,
			"max_retries", s.cfg.MaxRetries,
			"error", err,
		)
	}

	// Все ретраи исчерпаны
	metrics.SMTPSendTotal.WithLabelValues("failed").Inc()
	return fmt.Errorf("smtp send failed after %d retries: %w", s.cfg.MaxRetries, lastErr)
}

// sendOnce выполняет одну попытку отправки без ретраев.
// Возвращает ошибку при любом сбое.
func (s *smtpSender) sendOnce(ctx context.Context, msg Message) error {
	// Быстрая проверка отмены контекста до установки соединения.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))

	var conn net.Conn
	var err error

	// Порт 465 — implicit TLS: устанавливаем TLS-соединение сразу.
	if s.cfg.Port == 465 {
		dialer := &net.Dialer{Timeout: s.cfg.Timeout}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName: s.cfg.Host,
		})
	} else {
		conn, err = net.DialTimeout("tcp", addr, s.cfg.Timeout)
	}
	if err != nil {
		return fmt.Errorf("smtp dial %s: %w", addr, err)
	}
	defer conn.Close()

	// Единый deadline на все SMTP-операции.
	if err := conn.SetDeadline(time.Now().Add(s.cfg.Timeout)); err != nil {
		return fmt.Errorf("smtp set deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	// STARTTLS для портов без implicit TLS.
	if s.cfg.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsCfg := &tls.Config{ServerName: s.cfg.Host}
			if err := client.StartTLS(tlsCfg); err != nil {
				return fmt.Errorf("smtp starttls: %w", err)
			}
		} else {
			s.logger.Warn("SMTP сервер не поддерживает STARTTLS, отправка без шифрования",
				"host", s.cfg.Host, "port", s.cfg.Port)
		}
	}

	// Аутентификация (если задан username).
	if s.cfg.Username != "" {
		auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	// MAIL FROM.
	if err := client.Mail(s.cfg.FromEmail); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	// RCPT TO.
	if err := client.Rcpt(s.cfg.To); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	// DATA.
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	letter := s.buildLetter(msg)
	if _, err := wc.Write([]byte(letter)); err != nil {
		wc.Close()
		return fmt.Errorf("smtp write data: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	// QUIT.
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}

	return nil
}

// buildLetter формирует MIME-сообщение (RFC 5322) с защитой от CRLF injection.
func (s *smtpSender) buildLetter(msg Message) string {
	var b strings.Builder

	b.WriteString("From: " + sanitizeHeader(formatAddress(s.cfg.FromName, s.cfg.FromEmail)) + "\r\n")
	b.WriteString("To: " + sanitizeHeader(s.cfg.To) + "\r\n")

	if msg.ReplyTo != "" {
		b.WriteString("Reply-To: " + sanitizeHeader(msg.ReplyTo) + "\r\n")
	}

	b.WriteString("Subject: " + sanitizeHeader(encodeSubject(msg.Subject)) + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("Message-ID: " + generateMessageID(s.cfg.FromEmail) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)

	return b.String()
}

// sanitizeHeader удаляет символы перевода строки из значения заголовка.
// Защита от CRLF injection: злоумышленник не может добавить собственные заголовки.
// Последовательности \r\n, \r, \n заменяются на пробел в один проход.
func sanitizeHeader(s string) string {
	return strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ").Replace(s)
}

// encodeSubject кодирует тему письма в RFC 2047 для поддержки UTF-8 (кириллица).
func encodeSubject(subject string) string {
	if subject == "" {
		return ""
	}
	return mime.QEncoding.Encode("utf-8", subject)
}

// formatAddress форматирует адрес в формате "Имя <email>".
func formatAddress(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

// generateMessageID генерирует уникальный Message-ID для письма.
func generateMessageID(fromEmail string) string {
	domain := "localhost"
	if parts := strings.SplitN(fromEmail, "@", 2); len(parts) == 2 {
		domain = parts[1]
	}
	return fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
}

// isRetryableSMTPError определяет, стоит ли повторять отправку при данной ошибке.
//
// Классификация:
//   - Временные (повторять): 421, 450-452, сетевые ошибки, таймауты
//   - Постоянные (НЕ повторять): 500-504, 535 (auth), 550-554
//   - Контекст отменён: НЕ повторять
func isRetryableSMTPError(err error) bool {
	if err == nil {
		return false
	}

	// Контекст отменён — не повторять
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// SMTP код ошибки (net/textproto.Error содержит метод Code() int).
	// net/smtp возвращает *textproto.Error для SMTP-ответов сервера.
	var textprotoErr *textproto.Error
	if errors.As(err, &textprotoErr) {
		code := textprotoErr.Code
		// 4xx — временные ошибки, повторять
		if code >= 400 && code < 500 {
			return true
		}
		// 5xx — постоянные ошибки, не повторять
		return false
	}

	// Сетевые ошибки (таймаут, connection reset) — повторять
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// TLS ошибки — повторять (могут быть временные сбои рукопожатия)
	var tlsErr *tls.CertificateVerificationError
	if errors.As(err, &tlsErr) {
		return true
	}

	// Ошибки "соединение сброшено" без конкретного типа
	errStr := err.Error()
	if strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "timeout") {
		return true
	}

	// По умолчанию не повторять (неизвестные ошибки)
	return false
}

// backoffDelay вычисляет задержку для экспоненциального отката с джиттером.
//
// Формула: base * 2^attempt, джиттер ±25%, максимум maxDelay.
// Потолок (max) применяется ПОСЛЕ джиттера, чтобы гарантировать,
// что итоговая задержка никогда не превысит max.
func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	delay := base * time.Duration(1<<uint(attempt))

	// Джиттер ±25% для предотвращения "грома" при массовых сбоях
	jitter := time.Duration((rand.Float64()*0.5 - 0.25) * float64(delay))
	delay += jitter

	// Применяем потолок после джиттера
	if delay > max {
		delay = max
	}
	// Страховка от отрицательных значений (если base очень маленькая)
	if delay < 0 {
		delay = 0
	}

	return delay
}
