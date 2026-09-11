-- +goose Up
-- +goose StatementBegin

CREATE TABLE hero_sources (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    url VARCHAR(1024) NOT NULL,
    title VARCHAR(500),
    -- vk, website, archive, book
    source_type VARCHAR(50) NOT NULL DEFAULT 'website',
    -- Сохранённый фрагмент текста (защита от удаления источника)
    excerpt TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hero_sources_hero ON hero_sources (hero_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS hero_sources;
-- +goose StatementEnd
