-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION reassign_main_photo()
RETURNS TRIGGER AS $$
BEGIN
    -- Если удалили главное фото и у героя остались другие фото —
    -- назначаем главным первое по sort_order
    IF OLD.is_main = TRUE THEN
        UPDATE photos
        SET is_main = TRUE
        WHERE id = (
            SELECT id FROM photos
            WHERE hero_id = OLD.hero_id
            ORDER BY sort_order ASC, id ASC
            LIMIT 1
        );
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_reassign_main_photo
AFTER DELETE ON photos
FOR EACH ROW
EXECUTE FUNCTION reassign_main_photo();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_reassign_main_photo ON photos;
DROP FUNCTION IF EXISTS reassign_main_photo();
-- +goose StatementEnd
