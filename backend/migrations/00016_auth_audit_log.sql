-- +goose Up

CREATE TABLE auth_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address      TEXT,
    mechanism       TEXT NOT NULL,
    result          TEXT NOT NULL,
    identity        TEXT,
    procedure       TEXT,
    failure_reason  TEXT,
    user_agent      TEXT
);

COMMENT ON TABLE auth_audit_log IS
    'Аудит-лог аутентификации: все попытки входа (успешные и неудачные)';

COMMENT ON COLUMN auth_audit_log.mechanism IS
    'Механизм аутентификации: static_key, api_key, jwt';

COMMENT ON COLUMN auth_audit_log.result IS
    'Результат: success, failure';

COMMENT ON COLUMN auth_audit_log.identity IS
    'Идентичность: username (JWT), key_name (API-key), пусто для static_key';

COMMENT ON COLUMN auth_audit_log.procedure IS
    'RPC-метод, для которого выполнялась аутентификация';

COMMENT ON COLUMN auth_audit_log.failure_reason IS
    'Причина неудачи (заполняется только при result=failure)';

CREATE INDEX idx_auth_audit_log_created_at ON auth_audit_log (created_at DESC);
CREATE INDEX idx_auth_audit_log_ip ON auth_audit_log (ip_address);
CREATE INDEX idx_auth_audit_log_result ON auth_audit_log (result);

-- +goose Down

DROP TABLE IF EXISTS auth_audit_log;
