-- +goose Up
-- +goose StatementBegin

-- Составной уникальный ключ идемпотентности сида и защиты от дублей админки (ADR-005).
-- Обычный UNIQUE (не partial): таргетный ON CONFLICT (name, type, jurisdiction) DO NOTHING
-- матчится только с полным индексом без предиката. Пустое имя отсекается валидацией
-- (ErrAwardNameRequired в usecase, строгая проверка в seed-builder).
-- ВНИМАНИЕ: на БД с уже существующими дублями индекс не создастся — сначала дедупликация.
CREATE UNIQUE INDEX idx_awards_seed_key ON awards (name, type, jurisdiction);

COMMENT ON INDEX idx_awards_seed_key IS
    'Ключ идемпотентности сида наград: (name, type, jurisdiction)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_awards_seed_key;

-- +goose StatementEnd
