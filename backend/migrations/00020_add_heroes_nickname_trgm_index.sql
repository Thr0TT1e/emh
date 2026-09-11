-- +goose Up
-- +goose StatementBegin

-- 1. GIN trgm индекс на nickname для ILIKE-поиска (публичный API).
-- Без CONCURRENTLY: таблица героев небольшая, блокировка на миллисекунды.
CREATE INDEX IF NOT EXISTS idx_heroes_nickname_trgm
    ON heroes USING gin (nickname gin_trgm_ops);

-- 2. Добавляем колонку search_vector для полнотекстового поиска.
-- Весовые категории: фамилия/имя/позывной = A (высший приоритет),
-- отчество = B (вторичный).
ALTER TABLE heroes ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- 3. Заполняем search_vector для существующих записей.
UPDATE heroes SET search_vector =
    setweight(to_tsvector('russian', COALESCE(last_name, '')), 'A') ||
    setweight(to_tsvector('russian', COALESCE(first_name, '')), 'A') ||
    setweight(to_tsvector('russian', COALESCE(middle_name, '')), 'B') ||
    setweight(to_tsvector('russian', COALESCE(nickname, '')), 'A');

-- 4. GIN-индекс на search_vector для быстрого полнотекстового поиска.
CREATE INDEX IF NOT EXISTS idx_heroes_search_vector
    ON heroes USING gin (search_vector);

-- 5. Триггер: автообновление search_vector при изменении ФИО/позывного.
CREATE OR REPLACE FUNCTION heroes_search_vector_update()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('russian', COALESCE(NEW.last_name, '')), 'A') ||
        setweight(to_tsvector('russian', COALESCE(NEW.first_name, '')), 'A') ||
        setweight(to_tsvector('russian', COALESCE(NEW.middle_name, '')), 'B') ||
        setweight(to_tsvector('russian', COALESCE(NEW.nickname, '')), 'A');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_heroes_search_vector
    BEFORE INSERT OR UPDATE OF first_name, last_name, middle_name, nickname
    ON heroes
    FOR EACH ROW
    EXECUTE FUNCTION heroes_search_vector_update();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_heroes_search_vector ON heroes;
DROP FUNCTION IF EXISTS heroes_search_vector_update();
DROP INDEX IF EXISTS idx_heroes_search_vector;
ALTER TABLE heroes DROP COLUMN IF EXISTS search_vector;
DROP INDEX IF EXISTS idx_heroes_nickname_trgm;
-- +goose StatementEnd
