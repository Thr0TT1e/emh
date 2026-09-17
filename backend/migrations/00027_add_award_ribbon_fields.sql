-- +goose Up
-- +goose StatementBegin

-- Поля орденской планки в справочнике наград (ADR-005).
-- type/jurisdiction хранятся как SMALLINT — маппинг значений proto enum:
--   AwardType:         0=UNSPECIFIED, 1=ORDER, 2=MEDAL, 3=BADGE
--   AwardJurisdiction: 0=UNSPECIFIED, 1=RUSSIAN_FEDERATION, 2=USSR, 3=DEPARTMENTAL
ALTER TABLE awards
    ADD COLUMN ribbon_image_url VARCHAR(1024),
    ADD COLUMN type SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN jurisdiction SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN worn_without_bar BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN is_jubilee BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN awards.ribbon_image_url IS
    'URL изображения ленты награды для орденской планки. NULL, если лента не загружена';
COMMENT ON COLUMN awards.type IS
    'Тип награды (AwardType): 0=UNSPECIFIED, 1=ORDER, 2=MEDAL, 3=BADGE';
COMMENT ON COLUMN awards.jurisdiction IS
    'Государственная принадлежность (AwardJurisdiction): 0=UNSPECIFIED, 1=RF, 2=USSR, 3=DEPARTMENTAL';
COMMENT ON COLUMN awards.worn_without_bar IS
    'Награда носится без колодки (звёзды орденов, знаки) и не входит в блок планок';
COMMENT ON COLUMN awards.is_jubilee IS
    'Юбилейная награда: при равных прочих уступает боевой в старшинстве';

-- Индекс под сортировку старшинства справочника (ADR-005 §2)
CREATE INDEX idx_awards_seniority ON awards (jurisdiction, type, sort_order);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_awards_seniority;

ALTER TABLE awards
    DROP COLUMN IF EXISTS ribbon_image_url,
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS jurisdiction,
    DROP COLUMN IF EXISTS worn_without_bar,
    DROP COLUMN IF EXISTS is_jubilee;

-- +goose StatementEnd
