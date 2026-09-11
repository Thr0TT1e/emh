-- +goose Up
-- +goose StatementBegin
ALTER TABLE photos ADD COLUMN face_box JSONB DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE photos DROP COLUMN IF EXISTS face_box;
-- +goose StatementEnd
