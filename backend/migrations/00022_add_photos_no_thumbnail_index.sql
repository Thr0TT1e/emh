-- +goose Up
-- +goose StatementBegin

-- Частичный индекс для быстрого поиска фото без thumbnail.
-- Используется в ListWithoutThumbnailsPaged для StartupBackfill и RescanLoop.
-- WHERE-условие соответствует фильтру в репозитории: COALESCE(thumbnail_url, '') = ''
CREATE INDEX IF NOT EXISTS idx_photos_no_thumbnail
    ON photos (id)
    WHERE COALESCE(thumbnail_url, '') = '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_photos_no_thumbnail;
-- +goose StatementEnd
