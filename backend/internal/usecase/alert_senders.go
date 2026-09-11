package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/smtp"
)

// ─── MultiAlertSender ────────────────────────────────────────────────────────

// MultiAlertSender отправляет алерт во все настроенные каналы параллельно.
// Ошибки одного канала не влияют на другие.
type MultiAlertSender struct {
	senders []AlertSender
	logger  *slog.Logger
}

// NewMultiAlertSender создаёт составной отправитель.
func NewMultiAlertSender(logger *slog.Logger, senders ...AlertSender) *MultiAlertSender {
	return &MultiAlertSender{senders: senders, logger: logger}
}

func (s *MultiAlertSender) SendAlert(ctx context.Context, alert domain.Alert) error {
	for _, sender := range s.senders {
		if err := sender.SendAlert(ctx, alert); err != nil {
			s.logger.Warn("alert sender failed",
				"alert_type", string(alert.Type),
				"error", err,
			)
		}
	}
	return nil
}

// ─── SMTPAlertSender ─────────────────────────────────────────────────────────

// SMTPAlertSender отправляет алерты через существующий smtp.Sender.
type SMTPAlertSender struct {
	sender smtp.Sender
}

// NewSMTPAlertSender создаёт SMTP-отправитель алертов.
func NewSMTPAlertSender(sender smtp.Sender) *SMTPAlertSender {
	return &SMTPAlertSender{sender: sender}
}

func (s *SMTPAlertSender) SendAlert(ctx context.Context, alert domain.Alert) error {
	msg := smtp.Message{
		Subject: fmt.Sprintf("[EMH SECURITY] %s", alert.Subject),
		Body:    formatAlertBody(alert),
	}
	return s.sender.Send(ctx, msg)
}

// ─── TelegramAlertSender ─────────────────────────────────────────────────────

// TelegramAlertSender отправляет алерты через Telegram Bot API.
type TelegramAlertSender struct {
	botToken string
	chatID   string
	client   *http.Client
}

// NewTelegramAlertSender создаёт Telegram-отправитель алертов.
func NewTelegramAlertSender(botToken, chatID string, timeout time.Duration) *TelegramAlertSender {
	return &TelegramAlertSender{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: timeout},
	}
}

func (s *TelegramAlertSender) SendAlert(ctx context.Context, alert domain.Alert) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)

	text := formatTelegramAlertText(alert)
	payload := map[string]string{
		"chat_id":    s.chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

// ─── Форматирование ─────────────────────────────────────────────────────────

func formatAlertBody(alert domain.Alert) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Тип алерта: %s\n", alert.Type))
	b.WriteString(fmt.Sprintf("Время: %s\n", alert.Timestamp.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Описание: %s\n", alert.Subject))
	if len(alert.Metadata) > 0 {
		b.WriteString("\nДетали:\n")
		for k, v := range alert.Metadata {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}
	if alert.Body != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", alert.Body))
	}
	return b.String()
}

func formatTelegramAlertText(alert domain.Alert) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("🚨 <b>%s</b>\n", alert.Subject))
	b.WriteString(fmt.Sprintf("Тип: <code>%s</code>\n", alert.Type))
	b.WriteString(fmt.Sprintf("Время: <code>%s</code>\n", alert.Timestamp.Format(time.RFC3339)))
	for k, v := range alert.Metadata {
		b.WriteString(fmt.Sprintf("%s: <code>%s</code>\n", k, v))
	}
	return b.String()
}
