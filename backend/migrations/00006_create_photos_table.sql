-- +goose Up
-- +goose StatementBegin

-- Медиа ресурсы
CREATE TABLE photos (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    url VARCHAR(1024) NOT NULL,
    thumbnail_url VARCHAR(1024),
    description VARCHAR(500),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_main BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_photos_hero ON photos (hero_id);
CREATE INDEX idx_photos_hero_main ON photos (hero_id) WHERE is_main = TRUE;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON photos
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- Триггер: при установке is_main=TRUE сбрасываем флаг у остальных фото героя
CREATE OR REPLACE FUNCTION ensure_single_main_photo()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_main = TRUE THEN
        UPDATE photos
        SET is_main = FALSE
        WHERE hero_id = NEW.hero_id AND id != NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_single_main_photo
BEFORE INSERT OR UPDATE OF is_main ON photos
FOR EACH ROW
EXECUTE FUNCTION ensure_single_main_photo();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS photos;
-- +goose StatementEnd
