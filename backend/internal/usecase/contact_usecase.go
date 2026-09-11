package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
	"codeberg.org/Thr0TT1e/emh/backend/internal/smtp"
)

// ContactUseCaseConfig параметры ContactUseCase.
type ContactUseCaseConfig struct {
	// SMTPEnabled — включена ли отправка email через SMTP.
	// Если false, сообщения сохраняются со статусом skipped.
	SMTPEnabled bool
	// IPPepper — секрет для хеширования IP-адресов (HMAC-SHA256).
	IPPepper string
	// DedupeWindow — окно анти-дубликата (по умолчанию 5 минут).
	DedupeWindow time.Duration
}

// SubmitContactInput входные данные для отправки сообщения обратной связи.
type SubmitContactInput struct {
	// Name — имя отправителя.
	Name string
	// Email — email отправителя.
	Email string
	// Subject — тема обращения.
	Subject string
	// Message — текст сообщения.
	Message string
	// PageURL — страница, с которой была отправлена форма.
	PageURL string
	// Honeypot — скрытое поле для ботов. Должно быть пустым.
	Honeypot string
	// Consent — согласие на обработку данных.
	Consent bool
	// ClientIP — IP-адрес клиента (из delivery-слоя).
	ClientIP string
	// UserAgent — User-Agent клиента (из HTTP заголовков).
	UserAgent string
}

// ContactUseCase бизнес-логика обработки сообщений обратной связи.
type ContactUseCase struct {
	repo        repository.ContactRepository
	rateLimiter *ContactRateLimiter
	smtpSender  smtp.Sender
	cfg         ContactUseCaseConfig
	logger      *slog.Logger
}

// NewContactUseCase создаёт usecase для обработки сообщений обратной связи.
func NewContactUseCase(
	repo repository.ContactRepository,
	rateLimiter *ContactRateLimiter,
	smtpSender smtp.Sender,
	cfg ContactUseCaseConfig,
	logger *slog.Logger,
) *ContactUseCase {
	return &ContactUseCase{
		repo:        repo,
		rateLimiter: rateLimiter,
		smtpSender:  smtpSender,
		cfg:         cfg,
		logger:      logger,
	}
}

// Submit обрабатывает сообщение обратной связи:
// нормализация → rate limit → анти-дубликат → сохранение → отправка email.
// Возвращает UUID созданного (или существующего при дубликате) сообщения.
func (uc *ContactUseCase) Submit(ctx context.Context, input SubmitContactInput) (string, error) {
	// 1. Нормализация полей.
	name := normalizeHeaderField(input.Name)
	email := normalizeEmail(input.Email)
	subject := normalizeHeaderField(input.Subject)
	message := strings.TrimSpace(input.Message)
	pageURL := strings.TrimSpace(input.PageURL)
	userAgent := truncateRunes(input.UserAgent, 512)

	// 2. Rate limiting по IP и email.
	if err := uc.rateLimiter.Allow("ip:" + input.ClientIP); err != nil {
		// метрика rate-limit по IP
		metrics.ContactRateLimitedTotal.WithLabelValues("ip").Inc()
		metrics.ContactMessagesTotal.WithLabelValues("rate_limited").Inc()

		return "", err
	}
	if err := uc.rateLimiter.Allow("email:" + email); err != nil {
		// метрика rate-limit по email
		metrics.ContactRateLimitedTotal.WithLabelValues("email").Inc()
		metrics.ContactMessagesTotal.WithLabelValues("rate_limited").Inc()

		return "", err
	}

	// 3. Хеш нормализованного сообщения (для анти-дубликата).
	messageHash := hashMessage(message)

	// 4. Анти-дубликат: ищем (email, message_hash) за последние DedupeWindow.
	since := time.Now().Add(-uc.cfg.DedupeWindow)
	existingID, err := uc.repo.FindRecentDuplicate(ctx, email, messageHash, since)
	if err != nil {
		return "", fmt.Errorf("find duplicate: %w", err)
	}
	if existingID != "" {
		uc.logger.Info("обнаружен дубликат сообщения",
			"existing_id", existingID,
			"window", uc.cfg.DedupeWindow.String(),
		)
		// метрика дубликата
		metrics.ContactMessagesTotal.WithLabelValues("duplicate").Inc()

		return existingID, nil
	}

	// 5. Определяем honeypot.
	isHoneypot := strings.TrimSpace(input.Honeypot) != ""

	// 6. Хеш IP (HMAC-SHA256 с pepper).
	ipHash := hashIP(input.ClientIP, uc.cfg.IPPepper)

	// 7. Сохраняем сообщение в БД (email_status = pending по умолчанию).
	params := domain.CreateContactMessageParams{
		Name:        name,
		Email:       email,
		Subject:     subject,
		Message:     message,
		MessageHash: messageHash,
		PageURL:     pageURL,
		IPHash:      ipHash,
		UserAgent:   userAgent,
		Consent:     input.Consent,
		IsHoneypot:  isHoneypot,
	}

	id, err := uc.repo.Create(ctx, params)
	if err != nil {
		return "", fmt.Errorf("create contact message: %w", err)
	}

	// метрика созданного сообщения
	metrics.ContactMessagesTotal.WithLabelValues("created").Inc()

	uc.logger.Info("сообщение обратной связи сохранено",
		"id", id,
		"is_honeypot", isHoneypot,
	)

	// 8. Honeypot: email не отправляем, статус skipped.
	if isHoneypot {
		// метрика пропущенной отправки (honeypot)
		metrics.SMTPSendTotal.WithLabelValues("skipped").Inc()
		uc.updateEmailStatus(ctx, id, domain.EmailStatusSkipped, "")

		return id, nil
	}

	// 9. SMTP отключён или sender не инициализирован: статус skipped.
	if !uc.cfg.SMTPEnabled || uc.smtpSender == nil {
		uc.logger.Debug("SMTP отключён, сообщение сохранено без отправки", "id", id)
		// метрика пропущенной отправки (SMTP disabled)
		metrics.SMTPSendTotal.WithLabelValues("skipped").Inc()
		uc.updateEmailStatus(ctx, id, domain.EmailStatusSkipped, "")

		return id, nil
	}

	// 10. Отправляем email-уведомление.
	// Таймаут 30 секунд: защита от блокировки при долгих ретраях.
	// Если отправка с ретраями превысит таймаут — сохранится как failed,
	// клиент получит ошибку, но сообщение остаётся в БД.
	sendCtx, sendCancel := context.WithTimeout(ctx, 30*time.Second)
	defer sendCancel()

	emailMsg := smtp.Message{
		ReplyTo: email,
		Subject: formatSubject(subject),
		Body:    buildEmailBody(name, email, subject, pageURL, message),
	}

	if err := uc.smtpSender.Send(sendCtx, emailMsg); err != nil {
		uc.logger.Error("ошибка отправки SMTP-уведомления",
			"id", id,
			"error", err,
		)
		// метрика неудачной отправки
		metrics.SMTPSendTotal.WithLabelValues("failed").Inc()
		uc.updateEmailStatus(ctx, id, domain.EmailStatusFailed, err.Error())

		return "", fmt.Errorf("send contact email: %w", domain.ErrContactSMTP)
	}

	// 11. Успех: обновляем статус.
	// метрика успешной отправки
	metrics.SMTPSendTotal.WithLabelValues("success").Inc()
	uc.updateEmailStatus(ctx, id, domain.EmailStatusSent, "")

	uc.logger.Info("SMTP-уведомление отправлено", "id", id)

	return id, nil
}

// updateEmailStatus обновляет статус отправки email с retry.
//
// Использует 3 попытки с exponential backoff (100ms → 200ms → 400ms).
// Это защищает от кратковременных сбоев БД после успешной SMTP-отправки.
//
// Если все попытки исчерпаны — логирует ошибку и инкрементирует метрику.
// Запись остаётся со статусом pending, но это не критично:
// - Пользователь уже получил ответ
// - Админ может вручную проверить статус в БД
// - Orphan cleanup не затрагивает contact_messages
func (uc *ContactUseCase) updateEmailStatus(
	ctx context.Context,
	id string,
	status domain.EmailStatus,
	emailErr string,
) {
	const (
		maxRetries = 3
		baseDelay  = 100 * time.Millisecond
		maxDelay   = 400 * time.Millisecond
	)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Задержка перед повтором (кроме первой попытки)
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			if delay > maxDelay {
				delay = maxDelay
			}
			select {
			case <-ctx.Done():
				uc.logger.Error("updateEmailStatus cancelled",
					"id", id,
					"status", string(status),
					"error", ctx.Err(),
				)
				metrics.ContactEmailStatusUpdateFailuresTotal.Inc()
				return
			case <-time.After(delay):
			}
		}

		// Проверка отмены контекста
		if err := ctx.Err(); err != nil {
			uc.logger.Error("updateEmailStatus cancelled before attempt",
				"id", id,
				"status", string(status),
				"error", err,
			)
			metrics.ContactEmailStatusUpdateFailuresTotal.Inc()
			return
		}

		err := uc.repo.UpdateEmailStatus(ctx, id, status, emailErr)
		if err == nil {
			// Успех
			return
		}

		lastErr = err

		// Если запись не найдена — не повторяем (запись была удалена)
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.Warn("contact message not found for status update",
				"id", id,
				"status", string(status),
			)
			return
		}

		uc.logger.Debug("updateEmailStatus retry",
			"id", id,
			"status", string(status),
			"attempt", attempt+1,
			"error", err,
		)
	}

	// Все retry исчерпаны
	uc.logger.Error("failed to update email_status after retries",
		"id", id,
		"status", string(status),
		"email_error", emailErr,
		"last_error", lastErr,
		"max_retries", maxRetries,
	)
	metrics.ContactEmailStatusUpdateFailuresTotal.Inc()
}

// --- Вспомогательные функции ---

// normalizeHeaderField удаляет управляющие символы перевода строки и нулевые байты,
// обрезает пробелы по краям. Защита от CRLF injection в email-заголовках.
// Последовательности \r\n, \r, \n заменяются на пробел в один проход,
// поэтому "Иван\r\nПетров" превращается в "Иван Петров" (один пробел),
// а не в "Иван  Петров" (два пробела).
func normalizeHeaderField(s string) string {
	s = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "\x00", "").Replace(s)
	return strings.TrimSpace(s)
}

// normalizeEmail приводит email к нижнему регистру и удаляет управляющие символы.
func normalizeEmail(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(s)
}

// hashMessage вычисляет SHA-256 хеш нормализованного сообщения.
func hashMessage(message string) string {
	h := sha256.Sum256([]byte(message))
	return hex.EncodeToString(h[:])
}

// hashIP вычисляет HMAC-SHA256 хеш IP-адреса с pepper.
// Сырой IP не хранится в БД — только хеш.
func hashIP(ip, pepper string) string {
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

// truncateRunes обрезает строку до maxLen рун (безопасно для UTF-8).
func truncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

// formatSubject формирует тему письма с префиксом [EMH].
func formatSubject(subject string) string {
	if subject == "" {
		return "[EMH] Без темы"
	}
	return "[EMH] " + subject
}

// buildEmailBody формирует plain-text тело письма в рекомендуемом формате.
func buildEmailBody(name, email, subject, pageURL, message string) string {
	var b strings.Builder

	b.WriteString("Новое сообщение с формы обратной связи\n\n")
	b.WriteString("Имя: " + name + "\n")
	b.WriteString("Email: " + email + "\n")

	if subject != "" {
		b.WriteString("Тема: " + subject + "\n")
	}
	if pageURL != "" {
		b.WriteString("Страница: " + pageURL + "\n")
	}

	b.WriteString("\nСообщение:\n")
	b.WriteString(message)
	b.WriteString("\n")

	return b.String()
}
