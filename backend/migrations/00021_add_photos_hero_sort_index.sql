-- +goose Up
-- +goose StatementBegin

-- Индекс для ускорения LEFT JOIN LATERAL в ListHeroes.
-- Покрывает: WHERE hero_id = ? ORDER BY is_main DESC, sort_order ASC, id ASC
-- Частичный индекс: только герои с фото (исключаем пустые строки).
CREATE INDEX IF NOT EXISTS idx_photos_hero_main_sort
    ON photos (hero_id, is_main DESC, sort_order ASC, id ASC)
    INCLUDE (url, thumbnail_url);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_photos_hero_main_sort;
-- +goose StatementEnd
