-- +goose Up
-- +goose StatementBegin

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    -- Публичный идентификатор для поиска (не секрет). UNIQUE даёт индекс.
    key_id VARCHAR(64) NOT NULL UNIQUE,
    -- bcrypt-хеш секретной части. Plaintext НЕ хранится.
    secret_hash VARCHAR(255) NOT NULL,
    -- Человекочитаемое имя и описание (для кого/зачем ключ)
    name VARCHAR(255) NOT NULL,
    description TEXT,
    -- Права ключа (по умолчанию admin)
    role VARCHAR(50) NOT NULL DEFAULT 'admin',
    -- Аудит: кто создал ключ
    created_by VARCHAR(255),
    -- Если не NULL — ключ отозван
    revoked_at TIMESTAMPTZ,
    -- Срок действия (NULL = бессрочный)
    expires_at TIMESTAMPTZ,
    -- Аудит использования
    last_used_at TIMESTAMPTZ,
    last_used_ip VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Частичный индекс для списка активных ключей
CREATE INDEX idx_api_keys_active ON api_keys (created_at DESC) WHERE revoked_at IS NULL;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON api_keys
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS api_keys;
-- +goose StatementEnd
