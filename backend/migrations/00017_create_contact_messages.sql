-- +goose Up
-- +goose StatementBegin

-- Таблица сообщений обратной связи (форма /contacts).
-- Append-only: запись создаётся при отправке формы,
-- затем обновляется только email_status / email_error / email_sent_at.
CREATE TABLE IF NOT EXISTS contact_messages (
    -- UUIDv7 для монотонности и эффективности B-Tree индексов
    -- (используем uuidv7(), а не gen_random_uuid(), как в heroes)
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    -- Данные формы (CHECK дублирует proto-валидацию на уровне БД)
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 2 AND 200),
    email TEXT NOT NULL CHECK (char_length(email) <= 254),
    subject TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL CHECK (char_length(message) BETWEEN 10 AND 5000),

    -- SHA-256 хеш нормализованного сообщения (для анти-дубликата)
    message_hash TEXT NOT NULL,

    -- Контекст отправки
    page_url TEXT NOT NULL DEFAULT '',
    ip_hash TEXT,            -- HMAC-SHA256(IP, pepper); сырой IP не храним
    user_agent TEXT,

    -- Метаданные
    consent BOOLEAN NOT NULL DEFAULT TRUE,
    is_honeypot BOOLEAN NOT NULL DEFAULT FALSE,

    -- Статус отправки email-уведомления
    email_status TEXT NOT NULL DEFAULT 'pending'
        CHECK (email_status IN ('pending', 'sent', 'failed', 'skipped')),
    email_error TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    email_sent_at TIMESTAMPTZ
);

-- Админ-панель / отладка: последние сообщения первыми
CREATE INDEX IF NOT EXISTS contact_messages_created_at_idx
    ON contact_messages (created_at DESC);

-- Partial index: воркер ретрая или ручной запрос
-- «дай все неотправленные и упавшие»
CREATE INDEX IF NOT EXISTS contact_messages_email_status_idx
    ON contact_messages (email_status)
    WHERE email_status IN ('pending', 'failed');

-- Анти-дубликат: поиск (email, message_hash) за последние N минут.
-- created_at DESC позволяет эффективно выбрать свежую запись.
CREATE INDEX IF NOT EXISTS contact_messages_dedupe_idx
    ON contact_messages (email, message_hash, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS contact_messages;
-- +goose StatementEnd
