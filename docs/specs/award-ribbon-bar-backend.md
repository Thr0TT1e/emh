# Техническое задание: Бэкенд — Орденская планка

**Дата:** 2026-09-17
**Статус:** Implemented (кроме `devices` — вне среза)
**Связано:** ADR-005

> Документ приведён в соответствие с фактической реализацией.
> Ключевые отличия от первой редакции — в разделе «9. Отклонения от исходного ТЗ».

## 1. Расширение `proto/emh/v1/enums_emh.proto`

### 1.1. enum `AwardType`

Приведён к единому стилю файла (leading-комментарий на enum и каждое значение):

```protobuf
// AwardType определяет тип награды. Используется для старшинства в орденской планке.
enum AwardType {
  // AWARD_TYPE_UNSPECIFIED - тип не указан.
  AWARD_TYPE_UNSPECIFIED = 0;
  // AWARD_TYPE_ORDER - орден.
  AWARD_TYPE_ORDER = 1;
  // AWARD_TYPE_MEDAL - медаль.
  AWARD_TYPE_MEDAL = 2;
  // AWARD_TYPE_BADGE - знак отличия или почётный знак.
  AWARD_TYPE_BADGE = 3;
}
```

### 1.2. enum `AwardJurisdiction` (новый)

Первичная ось старшинства (приказ МО РФ №1500). Нулевое значение — `_UNSPECIFIED`
(фактический стиль `enums_emh.proto`; правило `ENUM_ZERO_VALUE_SUFFIX` в `easyp.yaml` отключено).

```protobuf
// AwardJurisdiction определяет государственную принадлежность награды.
enum AwardJurisdiction {
  // AWARD_JURISDICTION_UNSPECIFIED - принадлежность не указана.
  AWARD_JURISDICTION_UNSPECIFIED = 0;
  // AWARD_JURISDICTION_RUSSIAN_FEDERATION - награда Российской Федерации.
  AWARD_JURISDICTION_RUSSIAN_FEDERATION = 1;
  // AWARD_JURISDICTION_USSR - награда СССР.
  AWARD_JURISDICTION_USSR = 2;
  // AWARD_JURISDICTION_DEPARTMENTAL - ведомственная награда.
  AWARD_JURISDICTION_DEPARTMENTAL = 3;
}
```

## 2. Расширение `proto/emh/v1/award.proto`

`message Award` дополнен `jurisdiction = 10` (помимо `ribbon_image_url = 6`,
`type = 7`, `worn_without_bar = 8`, `is_jubilee = 9`). Все поля — в конец, breaking-free.

## 3. Расширение `proto/emh/v1/hero.proto`

`HeroAward` **денормализует** поля справочника (JOIN заполняется в репозитории),
чтобы публичная страница не делала второй запрос к каталогу. Прецедент — `award_name`.

```protobuf
message HeroAward {
  // ... существующие поля 1-6 (включая devices = 6) ...
  string ribbon_image_url = 7;        // пусто, если лента не загружена
  AwardType type = 8;
  bool worn_without_bar = 9;
  bool is_jubilee = 10;
  AwardJurisdiction jurisdiction = 11;
  string image_url = 12;              // знак награды — для текстового fallback
}
```

`AwardDevice` и `devices` объявлены, но **не реализованы** (нет `UpdateHeroAwardRequest`,
бэкенд их не читает). Валидация `count > 0` намеренно не добавлялась — мёртвый путь
до реализации устройств.

## 4. Расширение `proto/emh/v1/award_admin.proto`

`CreateAwardRequest`: добавлены `jurisdiction = 9`, leading-комментарии на все новые поля.
`ribbon_image_url = 5` — с `(buf.validate.field).ignore = IGNORE_IF_ZERO_VALUE`
(лента опциональна: награду можно создать без ленты; без `ignore` `string.uri` отклонял бы пустую строку).

`UpdateAwardRequest`: добавлены `jurisdiction = 11`, leading-комментарии, и
`ignore = IGNORE_IF_ZERO_VALUE` на `ribbon_image_url`, `type`, `worn_without_bar`,
`is_jubilee` — без этого partial update (`field_mask: ['name']`) падал бы на `type = UNSPECIFIED`.

> `image_url = 3` в `CreateAwardRequest` оставлен без `ignore` (существующее поведение,
> не в этом срезе) — известная асимметрия зафиксирована в frontend-ТЗ.

## 5. Расширение `proto/emh/v1/media.proto`

`UPLOAD_TYPE_AWARD_RIBBON = 4` объявлен в enum.

## 6. Миграция БД — `backend/migrations/00027_add_award_ribbon_fields.sql`

Enum'ы хранятся как **SMALLINT** (маппинг из proto), по образцу `submissions.status`:

```sql
ALTER TABLE awards
    ADD COLUMN ribbon_image_url VARCHAR(1024),
    ADD COLUMN type SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN jurisdiction SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN worn_without_bar BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN is_jubilee BOOLEAN NOT NULL DEFAULT FALSE;
-- COMMENT ON COLUMN ... для каждого
CREATE INDEX idx_awards_seniority ON awards (jurisdiction, type, sort_order);
```

`-- +goose Down` — `DROP INDEX` + `DROP COLUMN IF EXISTS`. `ribbon_image_url` nullable
(пусто = нет ленты); `COALESCE(..., '')` при чтении гасит NULL у строк до миграции.

`hero_award_devices` **не создавалась**: FK на `hero_awards(id)` невозможен —
PK у `hero_awards` композитный `(hero_id, award_id)`, суррогатного `id` нет.
Devices вне среза (см. §3).

## 7. Домен и usecase

### 7.1. `backend/internal/domain/award.go`

```go
type AwardType int          // значения совпадают с proto enum (iota)
type AwardJurisdiction int  // значения совпадают с proto enum (iota)
```

`Award`, `HeroAward`, `CreateAwardParams`, `UpdateAwardParams` дополнены новыми полями.
`HeroAward` несёт денормализованные `RibbonImageURL`, `ImageURL`, `Type`, `Jurisdiction`,
`WornWithoutBar`, `IsJubilee`.

### 7.2. `backend/internal/domain/media.go`

`UploadTypeAwardRibbon` добавлен в iota-блок `UploadType` (значение 4 = proto).

### 7.3. `backend/internal/usecase/media_usecase.go`

Без этого загрузка ленты возвращала `ErrUnsupportedUploadType`:

```go
allowedMIME[domain.UploadTypeAwardRibbon] = {jpeg, png, webp}
typePrefix[domain.UploadTypeAwardRibbon]  = "awards/ribbons"
```

Delivery (`media.go`) кастует `domain.UploadType(req.Msg.Type)` напрямую — switch не нужен.

## 8. Репозитории и delivery

- `award_repository.go`: общий `awardColumns` (с `COALESCE(ribbon_image_url,'')`) +
  хелпер `scanAward(scannable)`. Create/GetByID/List/Update читают и пишут новые колонки.
  `Update` ветвится по `mask[...]` для каждого нового поля (snake_case-ключи).
- `hero_award_repository.go` `ListByHero`: JOIN `awards a` денормализует
  `COALESCE(a.ribbon_image_url,'')`, `COALESCE(a.image_url,'')`, `a.type`,
  `a.jurisdiction`, `a.worn_without_bar`, `a.is_jubilee`.
- `delivery/v1/award.go` `mapAwardToProto`: + новые поля (`emhv1.AwardType(a.Type)` и т.д.).
- `delivery/v1/hero.go` `mapHeroAwardToProto`: + денормализованные поля.
- `delivery/v1/award_admin.go`: маппинг запросов Create/Update → `domain.*Params`.

> Замечен существующий дефект (не в этом срезе): `ListByHero` не читает `award_date_*`,
> а `mapHeroAwardToProto` не маппит `AwardDateInfo`.

## 9. Отклонения от исходного ТЗ

- **`hero_award_devices` не создавалась** — FK на композитный PK `(hero_id, award_id)` невозможен.
  Devices полностью вне среза (нет `UpdateHeroAwardRequest`, бэкенд не читает).
- **`type`/`jurisdiction` как SMALLINT**, не `VARCHAR(20)` — по образцу `submissions.status`.
- **Добавлен `jurisdiction`** (новая ось сортировки) — без него приказ №1500 недостижим.
- **Денормализация в `HeroAward`** вместо клиентского join через `ListAwards`.
- **`ignore = IGNORE_IF_ZERO_VALUE`** на `ribbon_image_url` (Create) и на всех новых полях
  `UpdateAwardRequest` — иначе optional-лента и partial update падают на валидации.
- **Один rendition** ленты; thumbnail-воркер и два rendition отложены (ADR-001 покрывает только Photo).
- **`count > 0` не добавлялась** на `AwardDevice` (мёртвый путь).
- **`UploadTypeAwardRibbon` прописан в media-пайплайне** (`domain` + `allowedMIME` + `typePrefix`).

## 10. Критерии приёмки

1. ✅ `Award` имеет `ribbon_image_url`, `type`, `jurisdiction`, `worn_without_bar`, `is_jubilee`
2. ✅ `HeroAward` денормализует эти поля + `image_url`
3. ⛔ `HeroAward.devices[]` — объявлено, вне среза
4. ✅ `MediaService` поддерживает `UPLOAD_TYPE_AWARD_RIBBON` (domain + MIME + префикс)
5. ✅ Сортировка по `jurisdiction → type → isJubilee → sortOrder → name` (на фронтенде)
6. ✅ Миграция `00027` (SMALLINT enum'ы, индекс старшинства, round-trip Down/Up)

## 11. Верификация

- [x] `just gen-sdk` — `easyp lint` 0 issues (было 14), Go + TS + docs сгенерированы
- [x] `go build -C backend ./...` — exit 0
- [x] `go vet -C backend ./...` — новых предупреждений нет (2 существующих в `submission_repository.go`, `orphan_cleanup_worker.go` — не в срезе)
- [x] `just test` — все пакеты ok (delivery, domain, usecase и др.)
- [x] Интеграционные тесты `./internal/repository/pg/...` (с `TEST_DATABASE_URL`) — ok
- [x] Миграция `00027`: `up` → `down` → `up` round-trip — ok
- [x] Ручная проверка денормализации SQL на тестовой БД (COALESCE гасит NULL) — ok
- [ ] E2E загрузки ленты через `MediaService` — требует поднятого MinIO + данных (интеграционная проверка)
