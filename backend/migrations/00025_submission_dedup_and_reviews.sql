-- +goose Up

-- 1. Добавляем content_hash в submissions для анти-дубликата
ALTER TABLE submissions ADD COLUMN content_hash TEXT;

-- Частичный уникальный индекс: только для активных заявок (не ARCHIVED).
-- Отклонённые заявки можно переотправлять — хеш может совпадать.
-- NULL значения не индексируются (заявки до миграции).
--
-- ⚠️  статус хранится как SMALLINT (маппинг из PublicationStatus):
--     0=UNSPECIFIED, 1=DRAFT, 2=PUBLISHED, 3=ARCHIVED
CREATE UNIQUE INDEX idx_submissions_content_hash_active
    ON submissions (content_hash)
    WHERE status != 3  -- PUBLICATION_STATUS_ARCHIVED
      AND content_hash IS NOT NULL;

COMMENT ON COLUMN submissions.content_hash IS
    'SHA-256 хеш нормализованного payload_json для защиты от дубликатов';

-- 2. Таблица аудита модерации
CREATE TABLE submission_reviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id   UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    reviewer_name   TEXT NOT NULL,
    decision        TEXT NOT NULL,
    comment         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE submission_reviews IS
    'Аудит модерации заявок: кто, когда и какое решение принял';

CREATE INDEX idx_submission_reviews_submission_id
    ON submission_reviews (submission_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS submission_reviews;
DROP INDEX IF EXISTS idx_submissions_content_hash_active;
ALTER TABLE submissions DROP COLUMN IF EXISTS content_hash;
