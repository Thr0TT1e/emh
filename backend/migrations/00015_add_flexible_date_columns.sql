-- +goose Up

ALTER TABLE heroes
    ADD COLUMN birth_date_precision SMALLINT,
    ADD COLUMN birth_date_display TEXT,
    ADD COLUMN death_date_precision SMALLINT,
    ADD COLUMN death_date_display TEXT,
    ADD COLUMN service_start_date_precision SMALLINT,
    ADD COLUMN service_start_date_display TEXT;

-- Маппинг DatePrecision:
--
-- 0 = DATE_PRECISION_UNSPECIFIED
-- 1 = DATE_PRECISION_EXACT
-- 2 = DATE_PRECISION_MONTH
-- 3 = DATE_PRECISION_YEAR
-- 4 = DATE_PRECISION_SEASON
-- 5 = DATE_PRECISION_DAY_MONTH
-- 6 = DATE_PRECISION_RANGE
-- 7 = DATE_PRECISION_UNKNOWN

UPDATE heroes
SET
    birth_date_precision = CASE
        WHEN birth_date IS NULL THEN 7
        ELSE 1
    END,
    death_date_precision = CASE
        WHEN death_date IS NULL THEN 7
        ELSE 1
    END,
    service_start_date_precision = CASE
        WHEN service_start_date IS NULL THEN 7
        ELSE 1
    END;

ALTER TABLE heroes
    ALTER COLUMN birth_date_precision SET NOT NULL,
    ALTER COLUMN death_date_precision SET NOT NULL,
    ALTER COLUMN service_start_date_precision SET NOT NULL;

ALTER TABLE heroes
    ALTER COLUMN birth_date_precision SET DEFAULT 7,
    ALTER COLUMN death_date_precision SET DEFAULT 7,
    ALTER COLUMN service_start_date_precision SET DEFAULT 7;

ALTER TABLE heroes
    ADD CONSTRAINT heroes_birth_date_precision_check
        CHECK (birth_date_precision BETWEEN 1 AND 7),
    ADD CONSTRAINT heroes_death_date_precision_check
        CHECK (death_date_precision BETWEEN 1 AND 7),
    ADD CONSTRAINT heroes_service_start_date_precision_check
        CHECK (service_start_date_precision BETWEEN 1 AND 7);

COMMENT ON COLUMN heroes.birth_date_precision IS
    'Точность даты рождения: 1 exact, 2 month, 3 year, 4 season, 5 day_month, 6 range, 7 unknown';

COMMENT ON COLUMN heroes.birth_date_display IS
    'Человекочитаемое представление даты рождения, например: Февраль 1994, 1994, 28 июля';

COMMENT ON COLUMN heroes.death_date_precision IS
    'Точность даты гибели: 1 exact, 2 month, 3 year, 4 season, 5 day_month, 6 range, 7 unknown';

COMMENT ON COLUMN heroes.death_date_display IS
    'Человекочитаемое представление даты гибели, например: 28 июля, Сентябрь 1999';

COMMENT ON COLUMN heroes.service_start_date_precision IS
    'Точность даты начала службы: 1 exact, 2 month, 3 year, 4 season, 5 day_month, 6 range, 7 unknown';

COMMENT ON COLUMN heroes.service_start_date_display IS
    'Человекочитаемое представление даты начала службы, например: 1989, Лето 1989';

-- +goose Down

ALTER TABLE heroes
    DROP CONSTRAINT IF EXISTS heroes_birth_date_precision_check,
    DROP CONSTRAINT IF EXISTS heroes_death_date_precision_check,
    DROP CONSTRAINT IF EXISTS heroes_service_start_date_precision_check;

ALTER TABLE heroes
    DROP COLUMN IF EXISTS birth_date_display,
    DROP COLUMN IF EXISTS birth_date_precision,
    DROP COLUMN IF EXISTS death_date_display,
    DROP COLUMN IF EXISTS death_date_precision,
    DROP COLUMN IF EXISTS service_start_date_display,
    DROP COLUMN IF EXISTS service_start_date_precision;
