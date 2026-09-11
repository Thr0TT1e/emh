package domain

import "time"

// EmailStatus статус отправки email-уведомления по сообщению обратной связи.
type EmailStatus string

const (
	// EmailStatusPending — сообщение сохранено, отправка ещё не выполнялась.
	EmailStatusPending EmailStatus = "pending"
	// EmailStatusSent — письмо успешно отправлено через SMTP.
	EmailStatusSent EmailStatus = "sent"
	// EmailStatusFailed — отправка завершилась ошибкой (текст в email_error).
	EmailStatusFailed EmailStatus = "failed"
	// EmailStatusSkipped — отправка пропущена (SMTP выключен или honeypot).
	EmailStatusSkipped EmailStatus = "skipped"
)

// ContactMessage сообщение обратной связи с формы /contacts.
// Таблица append-only: после создания обновляются только поля отправки email.
type ContactMessage struct {
	// ID — UUIDv7 сообщения.
	ID string
	// Name — имя отправителя (2–200 символов).
	Name string
	// Email — email отправителя в нижнем регистре.
	Email string
	// Subject — тема обращения (может быть пустой).
	Subject string
	// Message — текст сообщения (10–5000 символов).
	Message string
	// MessageHash — SHA-256 нормализованного текста (для анти-дубликата).
	MessageHash string
	// PageURL — страница, с которой была отправлена форма.
	PageURL string
	// IPHash — HMAC-SHA256(IP, pepper). Сырой IP не хранится.
	IPHash string
	// UserAgent — User-Agent клиента (обрезается до 512 символов).
	UserAgent string
	// Consent — согласие на обработку персональных данных.
	Consent bool
	// IsHoneypot — true, если скрытое поле honeypot было заполнено.
	IsHoneypot bool
	// EmailStatus — текущий статус отправки уведомления.
	EmailStatus EmailStatus
	// EmailError — текст ошибки SMTP (заполняется при EmailStatusFailed).
	EmailError string
	// CreatedAt — время создания записи.
	CreatedAt time.Time
	// EmailSentAt — время успешной отправки письма (nil, если не отправлено).
	EmailSentAt *time.Time
}

// CreateContactMessageParams параметры для создания сообщения обратной связи.
// Передаются из usecase в репозиторий после нормализации и валидации.
type CreateContactMessageParams struct {
	// Name — нормализованное имя отправителя.
	Name string
	// Email — email в нижнем регистре, без управляющих символов.
	Email string
	// Subject — тема обращения.
	Subject string
	// Message — текст сообщения.
	Message string
	// MessageHash — SHA-256 хеш нормализованного сообщения.
	MessageHash string
	// PageURL — страница-источник.
	PageURL string
	// IPHash — хеш IP-адреса клиента.
	IPHash string
	// UserAgent — User-Agent клиента.
	UserAgent string
	// Consent — согласие на обработку данных.
	Consent bool
	// IsHoneypot — флаг honeypot-сообщения.
	IsHoneypot bool
}
