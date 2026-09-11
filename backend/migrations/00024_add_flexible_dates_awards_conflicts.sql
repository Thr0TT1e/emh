-- +goose Up
-- +goose StatementBegin

-- Гибкие даты для наград (hero_awards)
ALTER TABLE hero_awards
    ADD COLUMN award_date_precision SMALLINT,
    ADD COLUMN award_date_display_text VARCHAR(200),
    ADD COLUMN award_date_anchor DATE;

-- Гибкие даты для конфликтов (conflicts)
ALTER TABLE conflicts
    ADD COLUMN start_date_precision SMALLINT,
    ADD COLUMN start_date_display_text VARCHAR(200),
    ADD COLUMN start_date_anchor DATE,
    ADD COLUMN end_date_precision SMALLINT,
    ADD COLUMN end_date_display_text VARCHAR(200),
    ADD COLUMN end_date_anchor DATE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE hero_awards
    DROP COLUMN IF EXISTS award_date_precision,
    DROP COLUMN IF EXISTS award_date_display_text,
    DROP COLUMN IF EXISTS award_date_anchor;

ALTER TABLE conflicts
    DROP COLUMN IF EXISTS start_date_precision,
    DROP COLUMN IF EXISTS start_date_display_text,
    DROP COLUMN IF EXISTS start_date_anchor,
    DROP COLUMN IF EXISTS end_date_precision,
    DROP COLUMN IF EXISTS end_date_display_text,
    DROP COLUMN IF EXISTS end_date_anchor;

-- +goose StatementEnd
