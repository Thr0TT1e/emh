-- +goose Up
-- +goose StatementBegin

-- Награды
CREATE TABLE awards (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image_url VARCHAR(1024),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_awards_sort_order ON awards (sort_order);
CREATE INDEX idx_awards_name_trgm ON awards USING gin (name gin_trgm_ops);

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON awards
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS awards;
-- +goose StatementEnd
