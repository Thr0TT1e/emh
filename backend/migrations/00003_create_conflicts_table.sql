-- +goose Up
-- +goose StatementBegin

-- Конфликты
CREATE TABLE conflicts (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type SMALLINT NOT NULL DEFAULT 0,
    start_date DATE,
    end_date DATE,
    parent_conflict_id UUID REFERENCES conflicts(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conflicts_type ON conflicts (type);
CREATE INDEX idx_conflicts_parent ON conflicts (parent_conflict_id) WHERE parent_conflict_id IS NOT NULL;
CREATE INDEX idx_conflicts_name_trgm ON conflicts USING gin (name gin_trgm_ops);

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON conflicts
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS conflicts;
-- +goose StatementEnd
