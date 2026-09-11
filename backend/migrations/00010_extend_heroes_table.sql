-- +goose Up
-- +goose StatementBegin

ALTER TABLE heroes
    ADD COLUMN nickname VARCHAR(100),          -- позывной
    ADD COLUMN unit VARCHAR(255),              -- подразделение/часть
    ADD COLUMN position VARCHAR(255),          -- должность
    ADD COLUMN service_branch VARCHAR(100),    -- ведомство/род войск
    ADD COLUMN cause_of_death TEXT,            -- причина гибели
    ADD COLUMN service_start_date DATE,        -- начало службы
    ADD COLUMN memberships JSONB NOT NULL DEFAULT '[]';  -- членство в организациях

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE heroes
    DROP COLUMN IF EXISTS nickname,
    DROP COLUMN IF EXISTS unit,
    DROP COLUMN IF EXISTS position,
    DROP COLUMN IF EXISTS service_branch,
    DROP COLUMN IF EXISTS cause_of_death,
    DROP COLUMN IF EXISTS service_start_date,
    DROP COLUMN IF EXISTS memberships;

-- +goose StatementEnd
