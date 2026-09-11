-- +goose Up
-- +goose StatementBegin

CREATE TABLE llm_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    priority INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_llm_providers_active ON llm_providers(is_active) WHERE is_active = true;

-- Триггер: только один активный провайдер
CREATE OR REPLACE FUNCTION trg_single_active_provider()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_active = true THEN
        UPDATE llm_providers SET is_active = false WHERE id != NEW.id AND is_active = true;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_single_active_provider
    BEFORE INSERT OR UPDATE OF is_active ON llm_providers
    FOR EACH ROW
    EXECUTE FUNCTION trg_single_active_provider();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_single_active_provider ON llm_providers;
DROP FUNCTION IF EXISTS trg_single_active_provider();
DROP TABLE IF EXISTS llm_providers;
-- +goose StatementEnd
