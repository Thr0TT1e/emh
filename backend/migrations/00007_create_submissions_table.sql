-- +goose Up
-- +goose StatementBegin

-- Пользовательских заявок (краудсорсинг)
CREATE TABLE submissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    submitter_name VARCHAR(255) NOT NULL,
    submitter_email VARCHAR(255) NOT NULL,
    -- При удалении героя заявка сохраняется (историческая ценность), ссылка обнуляется
    target_hero_id UUID REFERENCES heroes(id) ON DELETE SET NULL,
    -- JSONB позволяет индексировать и запрашивать гибкие данные заявок
    payload_json JSONB NOT NULL,
    attachment_urls TEXT[] NOT NULL DEFAULT '{}',
    -- 1=DRAFT (на модерации), 2=PUBLISHED (одобрена), 3=ARCHIVED (отклонена)
    status SMALLINT NOT NULL DEFAULT 1,
    moderator_comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submissions_status ON submissions (status);
CREATE INDEX idx_submissions_target_hero ON submissions (target_hero_id) WHERE target_hero_id IS NOT NULL;
-- uuidv7 монотонен, поэтому индекс по id даёт хронологический порядок
CREATE INDEX idx_submissions_created ON submissions (id DESC);
-- GIN-индекс для поиска по содержимому payload
CREATE INDEX idx_submissions_payload ON submissions USING gin (payload_json);

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON submissions
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS submissions;
-- +goose StatementEnd
