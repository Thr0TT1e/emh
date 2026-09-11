-- +goose Up

-- ============================================================
-- Исправление миграции 00024: замена трёх колонок на JSONB
-- ============================================================

-- 1. conflicts: start_date_info
ALTER TABLE conflicts ADD COLUMN start_date_info JSONB;

-- Миграция данных (если есть)
UPDATE conflicts
SET start_date_info = jsonb_build_object(
    'precision', start_date_precision,
    'display_text', start_date_display_text,
    'anchor', start_date_anchor
)
WHERE start_date_precision IS NOT NULL
   OR start_date_display_text IS NOT NULL
   OR start_date_anchor IS NOT NULL;

ALTER TABLE conflicts
DROP COLUMN start_date_precision,
DROP COLUMN start_date_display_text,
DROP COLUMN start_date_anchor;

-- 2. conflicts: end_date_info
ALTER TABLE conflicts ADD COLUMN end_date_info JSONB;

UPDATE conflicts
SET end_date_info = jsonb_build_object(
    'precision', end_date_precision,
    'display_text', end_date_display_text,
    'anchor', end_date_anchor
)
WHERE end_date_precision IS NOT NULL
   OR end_date_display_text IS NOT NULL
   OR end_date_anchor IS NOT NULL;

ALTER TABLE conflicts
DROP COLUMN end_date_precision,
DROP COLUMN end_date_display_text,
DROP COLUMN end_date_anchor;

-- 3. hero_awards: award_date_info
ALTER TABLE hero_awards ADD COLUMN award_date_info JSONB;

UPDATE hero_awards
SET award_date_info = jsonb_build_object(
    'precision', award_date_precision,
    'display_text', award_date_display_text,
    'anchor', award_date_anchor
)
WHERE award_date_precision IS NOT NULL
   OR award_date_display_text IS NOT NULL
   OR award_date_anchor IS NOT NULL;

ALTER TABLE hero_awards
DROP COLUMN award_date_precision,
DROP COLUMN award_date_display_text,
DROP COLUMN award_date_anchor;

COMMENT ON COLUMN conflicts.start_date_info IS 'Гибкая дата начала (JSONB: precision, display_text, anchor)';
COMMENT ON COLUMN conflicts.end_date_info IS 'Гибкая дата окончания (JSONB: precision, display_text, anchor)';
COMMENT ON COLUMN hero_awards.award_date_info IS 'Гибкая дата награждения (JSONB: precision, display_text, anchor)';

-- +goose Down

-- Откат: возвращаем три колонки
ALTER TABLE conflicts
ADD COLUMN start_date_precision SMALLINT,
ADD COLUMN start_date_display_text VARCHAR(200),
ADD COLUMN start_date_anchor DATE;

UPDATE conflicts
SET start_date_precision = (start_date_info->>'precision')::smallint,
    start_date_display_text = start_date_info->>'display_text',
    start_date_anchor = (start_date_info->>'anchor')::date
WHERE start_date_info IS NOT NULL;

ALTER TABLE conflicts DROP COLUMN start_date_info;

ALTER TABLE conflicts
ADD COLUMN end_date_precision SMALLINT,
ADD COLUMN end_date_display_text VARCHAR(200),
ADD COLUMN end_date_anchor DATE;

UPDATE conflicts
SET end_date_precision = (end_date_info->>'precision')::smallint,
    end_date_display_text = end_date_info->>'display_text',
    end_date_anchor = (end_date_info->>'anchor')::date
WHERE end_date_info IS NOT NULL;

ALTER TABLE conflicts DROP COLUMN end_date_info;

ALTER TABLE hero_awards
ADD COLUMN award_date_precision SMALLINT,
ADD COLUMN award_date_display_text VARCHAR(200),
ADD COLUMN award_date_anchor DATE;

UPDATE hero_awards
SET award_date_precision = (award_date_info->>'precision')::smallint,
    award_date_display_text = award_date_info->>'display_text',
    award_date_anchor = (award_date_info->>'anchor')::date
WHERE award_date_info IS NOT NULL;

ALTER TABLE hero_awards DROP COLUMN award_date_info;
