-- +goose Up
-- +goose StatementBegin

CREATE TABLE hero_relations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    from_hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    to_hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    -- father_son, comrade, commander и т.д.
    relation_type VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Защита от связи героя с самим собой (вторая линия обороны после usecase)
    CHECK (from_hero_id <> to_hero_id),
    -- Одна связь одного типа между парой героев (на этот ключ опирается ON CONFLICT в репозитории)
    UNIQUE (from_hero_id, to_hero_id, relation_type)
);

-- UNIQUE покрывает только from_hero_id; для поиска по to_hero_id нужен отдельный индекс
CREATE INDEX idx_hero_relations_to ON hero_relations (to_hero_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS hero_relations;
-- +goose StatementEnd
