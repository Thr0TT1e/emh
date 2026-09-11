-- +goose Up
-- +goose StatementBegin

-- Включаем расширение для триграммного поиска (нужно для быстрого ILIKE)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE heroes (
    -- UUIDv7 обеспечивает монотонность и идеален для B-Tree индексов
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),

    short_bio VARCHAR(1000),
    full_bio TEXT,
    rank VARCHAR(100),

    -- Для исторических дат лучше использовать DATE, а не TIMESTAMPTZ,
    -- чтобы избежать сдвигов из-за часовых поясов при сериализации.
    birth_date DATE,
    death_date DATE,

    -- Статус публикации (1=Draft, 2=Published, 3=Archived)
    status SMALLINT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы для фильтрации и сортировки
CREATE INDEX idx_heroes_status ON heroes (status);
CREATE INDEX idx_heroes_last_name ON heroes (last_name);
CREATE INDEX idx_heroes_death_date ON heroes (death_date) WHERE death_date IS NOT NULL;

-- GIN индексы для быстрого нечеткого поиска (ILIKE '%иван%')
CREATE INDEX idx_heroes_first_name_trgm ON heroes USING gin (first_name gin_trgm_ops);
CREATE INDEX idx_heroes_last_name_trgm ON heroes USING gin (last_name gin_trgm_ops);
CREATE INDEX idx_heroes_middle_name_trgm ON heroes USING gin (middle_name gin_trgm_ops);

-- Функция и триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON heroes
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS heroes;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd
