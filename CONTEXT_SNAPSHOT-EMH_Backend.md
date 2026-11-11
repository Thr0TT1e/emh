# CONTEXT SNAPSHOT — EMH Backend (Спринт 7 + сид наград/дубли, v0.0.9)

## 0. Метаданные снимка

| Параметр | Значение |
|---|---|
| Проект | Вечная память героям (EMH) |
| Go-модуль | `codeberg.org/Thr0TT1e/emh/backend` |
| Назначение снимка | Полное восстановление контекста бэкенд-разработки в новой сессии |
| Фаза | 🟢 **Спринт 7 закрыт**; сессия 2026-11-10/11: **сид справочника наград (v0.0.8)** и **защита дублей + keyset-пагинация (v0.0.8-1)** закрыты. Следующий этап: воркер rendition для админских загрузок наград, наполнение справочника (~400) |
| Миграции | 00001–00025 применены + **00026** (fix flexible dates jsonb), **00027** (award ribbon fields, ADR-005), **00028** (idx_awards_seed_key) |
| Версии образов | Quadlet'ы backend/migrate (и frontend) переведены на `:latest`, собранный из релиза **v0.0.9**. Теги: `v0.0.8`, `v0.0.8-1`, `v0.0.9` (ADR-006 — фронтовый/инфра-релиз, бэкенд без изменений) |
| CI | ✅ [run 42 / v0.0.8](https://codeberg.org/Thr0TT1e/emh/actions/runs/42/jobs/0/attempt/1), [run 44 / v0.0.8-1](https://codeberg.org/Thr0TT1e/emh/actions/runs/44/jobs/0/attempt/1), run 45 / v0.0.9 |
| Пакеты образов | [emh-backend:latest](https://codeberg.org/Thr0TT1e/-/packages/container/emh-backend/latest), [emh-frontend:latest](https://codeberg.org/Thr0TT1e/-/packages/container/emh-frontend/latest) |
| Зеркало/референс | origin — `codeberg.org/Thr0TT1e/emh`; проверка/актуализация — `github.com/Thr0TT1e/emh` |
| Production VPS | 82.202.139.132, домены: neverforgotten.ru, вечнаяпамятьгероям.рус, вежливые.рус, emh.su. Технические поддомены (api→backend, s3→MinIO) — на neverforgotten.ru; на каноническом IDN `encode zstd gzip` (ADR-006) |
| Test coverage | ✅ `internal/usecase` стабильно **74.4%**. `internal/smtp` **51.2%** |
| Security audit | ✅ H1-H4, M1-M6 закрыты. Все критические уязвимости устранены |

---

## 1. Цель проекта

Мемориальный сайт о героях, погибших в глобальных и локальных военных конфликтах. Наполнение — через админ-панель, краудсорсинг (заявки посетителей) и **LLM-пайплайн** (автоматическое извлечение данных из текста с модерацией).

---

## 2. Стек

### Backend

| Компонент | Версия | Назначение |
|---|---|---|
| Go | 1.26.5 | Язык |
| Connect RPC (`connectrpc.com/connect`) | v1.20.0 | API-протокол (HTTP+JSON) |
| connectrpc.com/validate | v0.6.0 | Авто-валидация по `buf.validate` |
| google.golang.org/protobuf | v1.36.11 | Protobuf runtime |
| jackc/pgx/v5 | v5.10.0 | PostgreSQL драйвер + пул |
| Masterminds/squirrel | v1.5.4 | SQL-билдер |
| go-chi/chi/v5 + cors | v5.3.1 / v1.2.2 | HTTP-роутер, CORS |
| minio/minio-go/v7 | v7.2.1 | S3-клиент |
| golang-jwt/jwt/v5 | v5.3.1 | JWT токены |
| google/uuid | v1.6.0 | Генерация UUID |
| golang.org/x/sync | v0.22.0 | errgroup (параллельная загрузка) |
| golang.org/x/crypto | v0.54.0 | bcrypt для паролей/API-key |
| golang.org/x/time | — | rate-limiting (token bucket) |
| golang.org/x/image | — | WebP-декодер |
| disintegration/imaging | — | resize/crop изображений |
| deepteams/webp | v1.2.7 | WebP encode/decode (lossy); регистрирует image-декодер для `DecodeConfig` |
| go.yaml.in/yaml/v3 | v3.0.5 | Парсинг `config.yaml` |
| github.com/prometheus/client_golang | latest | Prometheus-метрики |
| github.com/mark3labs/mcp-go | latest | MCP-сервер (stdio/HTTP+SSE) |

---

## 3. Архитектура

Clean Architecture: `domain` ← `usecase` ← `delivery`; `repository`/`storage` — инфраструктурные реализации.

### Ключевые архитектурные решения

- **CQRS для героя**: `HeroUseCase` (команды) и `HeroQueryUseCase` (чтение агрегата `HeroDetail` с параллельной загрузкой 6 связей через `errgroup`)
- **LLM-пайплайн**: `ExtractionUseCase` → Ollama/OpenAI-compatible → промпт-инжиниринг → парсинг JSON → поиск дубликатов (tsvector) → аудит в `llm_extraction_logs`
- **MCP-сервер**: отдельный бинарник `cmd/mcp` с 6 инструментами, переиспользует usecase'ы через общий пакет
- **Prometheus метрики**: отдельный пакет `internal/metrics` с 35+ метриками в 9 категориях
- **Порядок Connect-интерцепторов**: `rpcMetricsInterceptor` (самый внешний) → `authInterceptor` → `validateInterceptor`
- **Human-in-the-loop**: LLM не пишет в БД напрямую — только через `Submission` (модерация)
- 🆕 **Курсорная пагинация**: `ListHeroes` (LEFT JOIN LATERAL), `ListHeroPhotos` (composite cursor `sort_order:id`), `ListObjectsPaged` (marker-based S3), `ListWithoutThumbnailsPaged` (cursor by `id`), **+ ListConflicts/ListLocations/ListAwards** (cursor-based)
- 🆕 **Gzip compression**: `connect.WithCompression("gzip", nil, nil)` для всех Connect-хендлеров (public/admin/extraction)
- 🆕 **Backpressure**: `waitForQueueCapacity` в `StartupBackfill` — защита от переполнения очереди при массовом импорте
- 🆕 **Retry с экспоненциальным откатом**: SMTP (4xx), Enqueue (переполнение очереди), `UpdateEmailStatus`, S3 операции
- 🆕 **Async multi-sender alerter**: `AlertWorker` + `MultiAlertSender` (SMTP + Telegram), `AuthAlertService` для brute-force detection
- 🆕 **Submission pipeline hardening**:
  - Анти-дубликат через SHA-256 нормализованного JSON
  - Аудит модерации в `submission_reviews` (reviewer, decision, comment)
  - Email-уведомления заявителей о статусе модерации
  - Частичный уникальный индекс для активных заявок
- 🆕 **Сид справочника наград (ADR-005, v0.0.8)** — детерминированный пайплайн:
  - Артефакт `backend/seeds/awards/seed.json` (Git LFS; картинка — объект `{content_type, data}`, base64 StdEncoding без префикса)
  - Сборка — `cmd/seed-builder` со строгой валидацией: webp-магия, **точные** размеры 400×400 / 144×48, proto-канон enum (`emhv1.*_value`), внутренние дубли, непустой список; побайтовый детерминизм (порядок записей = входной файл, `Encoder.SetIndent("", "  ")`, `SetEscapeHTML(false)`)
  - `internal/seeder`: идемпотентный импорт `ON CONFLICT (name,type,jurisdiction) DO NOTHING`; MinIO недоступен / битая картинка → запись НЕ инсертится целиком (`Failed++`)
  - Входы: bootstrap авто-сид при пустом справочнике (синхронно до `ListenAndServe`, ошибки не блокируют старт) + one-shot субкоманда `emh seed-awards <path>`; путь `AWARDS_SEED_PATH` (дефолт `/app/seeds/seed.json`, volume `%h/emh/seeds:/app/seeds:ro`)
  - Ключи MinIO: `awards/images|ribbons/<sha256(name␀type␀jurisdiction)[:16]>.webp` (proto-имена, NUL-разделитель)
- 🆕 **Keyset-пагинация ListAwards (v0.0.8-1)**: курсор — непрозрачный base64url-JSON `(sort_order, name, id)`, условие `WHERE (sort_order,name,id) > (?,?,?)` вместо сломанного `id > cursor` (терял и дублировал записи при сортировке не по id). Неразбираемый курсор → `ErrInvalidCursor` → `CodeInvalidArgument`
- 🆕 **Защита дублей справочника**: `idx_awards_seed_key` UNIQUE `(name,type,jurisdiction)` (миграция 00028); 23505 в Create/Update → `ErrAwardDuplicate` → `CodeAlreadyExists` с русским сообщением

### Фоновые воркеры (все через `runWorkerWithRecovery`)

| Воркер | Назначение | Метрики |
|---|---|---|
| ThumbnailWorker | Генерация превью + WebP-конвертация | `emh_thumbnail_*` |
| ThumbnailWorker.RescanLoop | Периодическое восстановление потерянных задач | `emh_thumbnail_rescan_triggered_total` |
| OrphanCleanupWorker | Удаление сиротских файлов из S3 (streaming) | `emh_orphan_cleanup_*` |
| AuthAuditWorker | Асинхронная запись аудита | `emh_auth_audit_*` |
| **AuthAuditCleanupWorker** | Ротация `auth_audit_log` (90 дней) | `emh_auth_audit_deleted_total`, `emh_auth_audit_cleanup_duration_seconds` |
| **AlertWorker** | Асинхронная отправка алертов (SMTP + Telegram) | `emh_alert_sent_total{type, channel, status}` |
| **SubmissionEmailWorker** | Email-уведомления заявителям о статусе модерации | `emh_submission_email_sent_total{status}` |
| RefreshTokenCleanupWorker | Удаление истёкших refresh-токенов | `emh_refresh_tokens_expired_total` |
| LLMLogCleanupWorker | Очистка LLM-логов (90 дней) | `emh_llm_log_deleted_total` |

---

## 4. Структура директорий (обновлено)

```
backend/
├── cmd/
│   ├── api/                          # Разделён на 6 файлов
│   │   ├── main.go                   # точка входа + graceful shutdown + dispatch seed-awards (шаг 2.5) + bootstrap авто-сида (шаг 2.6)
│   │   ├── infra.go                  # connectDatabase, connectStorage, newLogger
│   │   ├── deps.go                   # wireDependencies (wiring 17 сервисов)
│   │   ├── workers.go                # startWorkers, runWorkerWithRecovery
│   │   ├── server.go                 # buildRouter, mountServices + gzip
│   │   └── seed_awards.go            # 🆕 runSeedAwards (one-shot) + bootstrapAwardsSeed
│   ├── seed-builder/main.go          # 🆕 сборка seed.json из seeds/awards (строгая валидация)
│   ├── mcp/main.go                   # MCP-сервер (stdio/HTTP)
│   └── hashpassword/main.go          # генерация bcrypt-хеша пароля
├── config/config.yaml                # +submission_emails секция
├── internal/
│   ├── auth/
│   ├── config/config.go              # +SubmissionEmailsConfig, +validate
│   ├── delivery/
│   │   ├── interceptor/auth.go       # AuthInterceptor + ContextWithClaims
│   │   ├── middleware/ratelimit.go
│   │   └── v1/                       # 17 серверов + errors.go
│   ├── domain/
│   │   ├── errors.go                 # 44 sentinel error (+ ErrSubmissionDuplicate, + ErrAwardDuplicate, + ErrInvalidCursor)
│   │   ├── error_codes.go            # + ErrCodeSubmissionDuplicate, + ErrCodeAwardDuplicate, + ErrCodeInvalidCursor
│   │   ├── submission.go             # + SubmissionReview, ContentHash
│   │   ├── submission_email.go       # SubmissionEmailNotification
│   │   ├── alert.go                  # Alert, AlertType
│   │   ├── conflict.go               # + ConflictFilter
│   │   ├── location.go               # + LocationFilter
│   │   ├── award.go                  # + AwardFilter
│   │   └── repository/
│   │       ├── submission_repository.go  # + FindByContentHash, CreateReview, ListReviews
│   │       ├── auth_audit_repository.go  # + DeleteOlderThan
│   │       ├── conflict_repository.go    # List → cursor-based
│   │       ├── location_repository.go    # List → cursor-based
│   │       └── award_repository.go       # List → cursor-based; 🆕 + ExistsByKey/InsertIfAbsent
│   ├── gen/emh/v1/                   # сгенерированный код
│   ├── llm/
│   ├── mcp/                          # 6 инструментов (обновлены для новых сигнатур List)
│   ├── metrics/                      # 10 файлов, 35+ метрик
│   ├── repository/pg/                # 16 репозиториев (обновлены для cursor; award — keyset)
│   ├── seeder/                       # 🆕 schema.go + awards.go (SeedAwards, HasAnyAwards, objectHash)
│   ├── storage/s3/
│   ├── smtp/sender.go                # + Send с retry
│   ├── usecase/
│   │   ├── submission_usecase.go     # + анти-дубликат, аудит модерации
│   │   ├── submission_email_worker.go # 🆕
│   │   ├── alert_worker.go           # 🆕
│   │   ├── auth_alert_service.go     # 🆕
│   │   ├── auth_audit_cleanup_worker.go # 🆕
│   │   ├── conflict_usecase.go       # List → ConflictFilter
│   │   ├── location_usecase.go       # List → LocationFilter
│   │   ├── award_usecase.go          # List → AwardFilter
│   │   └── ...
│   └── migrations/                   # goose 00001..00028
├── seeds/awards/                     # 🆕 описания + картинки + seed.json (LFS)
└── test/e2e/
```

---

## 5. Proto-контракты — статус

| Файл | Статус | Изменения |
|---|---|---|
| `enums_emh.proto` | ✅ | DatePrecision, AwardType, AwardJurisdiction |
| `common.proto` | ✅ | FlexibleDate, PaginationRequest/Response |
| `hero.proto` | ✅ | FlexibleDate поля, DEPRECATED Timestamp |
| `hero_admin.proto` | ✅ | FlexibleDate в Create/Update |
| `award.proto`, `award_admin.proto` | ✅ | + PaginationRequest/Response в List |
| `conflict.proto`, `conflict_admin.proto` | ✅ | + PaginationRequest/Response в List |
| `location.proto`, `location_admin.proto` | ✅ | + PaginationRequest/Response в List |
| `media.proto` | ✅ | BatchGetUploadUrls, UploadType AWARD_IMAGE(2)/AWARD_RIBBON(4) |
| `submission.proto` | ✅ | + SubmissionReview, SubmissionReviewDecision |
| `auth.proto` | ✅ | Refresh, Logout RPC |
| `api_key_admin.proto` | ✅ | — |
| `contact.proto` | ✅ | — |
| `extraction.proto` | ✅ | ExtractHeroData |
| `llm_admin.proto` | 🔵 Заготовка | Управление провайдерами |

> `slug` для `Award` в рамках среза сида **не добавлялся** (мёртвое поле без потребителя — зафиксировано в `awards-seed-backend.md` §14).

---

## 6. Схема БД (обновлено)

PostgreSQL 18+. Первичные ключи — UUIDv7. **Все 28 миграций успешно применены.**

| Миграция | Описание |
|---|---|
| 00001–00017 | Базовые таблицы, связи, триггеры, индексы |
| 00018 | `llm_extraction_logs` — аудит LLM-экстракций |
| 00019 | `llm_providers` — управление LLM-провайдерами + триггер |
| 00020 | `search_vector` (tsvector) + GIN + nickname trgm индекс |
| 00021 | Индексы фото: `idx_photos_hero_main_sort` + `idx_photos_hero_sort` |
| 00022 | Частичный индекс `idx_photos_no_thumbnail` |
| 00023 | `contact_messages` + `email_status` + `message_hash` |
| 00024 | Гибкие даты (FlexibleDate JSONB) для наград/конфликтов |
| 00025 | `submissions.content_hash` + `submission_reviews` + частичный unique index |
| 00026 | Фикс flexible dates jsonb |
| **00027** | 🆕 Поля орденской планки: `ribbon_image_url`, `type`, `jurisdiction` (SMALLINT), `worn_without_bar`, `is_jubilee` + `idx_awards_seniority` (ADR-005) |
| **00028** | 🆕 `idx_awards_seed_key` UNIQUE `(name, type, jurisdiction)` — ключ сида и защиты дублей |

### Ключевые индексы (актуально)

```sql
-- Анти-дубликат заявок (частичный)
CREATE UNIQUE INDEX idx_submissions_content_hash_active
    ON submissions (content_hash)
    WHERE status != 3  -- ARCHIVED
      AND content_hash IS NOT NULL;

-- Аудит модерации
CREATE INDEX idx_submission_reviews_submission_id
    ON submission_reviews (submission_id, created_at DESC);

-- Ключ сида/дублей справочника наград (ОБЫЧНЫЙ unique, не partial:
-- таргетный ON CONFLICT (name,type,jurisdiction) не матчится с partial-индексом)
CREATE UNIQUE INDEX idx_awards_seed_key ON awards (name, type, jurisdiction);

-- Старшинство планки (ADR-005 §2)
CREATE INDEX idx_awards_seniority ON awards (jurisdiction, type, sort_order);
```

⚠️ **Ловушка курсорной пагинации (закрыта в v0.0.8-1 для ListAwards)**: `WHERE id > cursor` корректен ТОЛЬКО при `ORDER BY id`. ListAwards сортирует по `(sort_order, name, id)` — там теперь keyset-кортеж. ListConflicts/ListLocations по-прежнему `id > cursor` — проверять их ORDER BY при расширении.

---

## 7. Статус Go-слоёв (обновлено)

| Слой | Статус |
|---|---|
| domain | ✅ 44 sentinel error (+`ErrAwardDuplicate`, +`ErrInvalidCursor`), `SubmissionReview`, `Alert`, фильтры справочников; `AwardRepository` + `ExistsByKey`/`InsertIfAbsent` |
| repository/pg | ✅ 16 репозиториев; award List — keyset-курсор `(sort_order,name,id)` (base64url-JSON токен); `awardColumns` с COALESCE для NULL-колонок |
| usecase | ✅ + `SubmissionEmailWorker`, + `AlertWorker`, + `AuthAuditCleanupWorker` |
| delivery/v1 | ✅ 17 серверов + `mapDomainError` с метриками + gzip; `ListAwards` ошибки через mapDomainError (`ErrInvalidCursor` → InvalidArgument) |
| delivery/interceptor | ✅ AuthInterceptor + `AuthAttemptsTotal` + `ContextWithClaims` |
| storage/s3 | ✅ `ListObjectsPaged`, countingReader, S3 метрики; `Upload` передаёт известный size для seekable-ридеров (один PUT вместо multipart) |
| smtp | ✅ `Send` с retry (4 попытки, экспоненциальный откат, джиттер) |
| llm/ | ✅ Ollama, OpenAI-compatible, Anthropic + ProviderManager |
| mcp/ | ✅ 6 инструментов (обновлены для cursor-based List) |
| metrics | ✅ 35+ метрик в 9 категориях |
| seeder | 🆕 `internal/seeder`: SeedAwards/HasAnyAwards, валидация base64+RIFF/WEBP, objectHash sha256[:16] |
| cmd/api/ | ✅ 6 файлов (+`seed_awards.go`: dispatch `seed-awards`, bootstrap авто-сида) |
| cmd/seed-builder/ | 🆕 standalone-сборщик seed.json (строгая валидация данных) |
| cmd/mcp/ | ✅ Отдельный бинарник |

---

## 8. Тестирование

Текущее покрытие:

| Пакет | Coverage | Комментарий |
|---|---|---|
| `internal/auth` | 91.7% | JWT, API-key, password, refresh token |
| `internal/config` | 65.2% | +LLM валидация, +SubmissionEmails |
| `internal/delivery/interceptor` | 89.7% | +метрики, +ContextWithClaims |
| `internal/delivery/middleware` | 97.5% | — |
| `internal/delivery/v1` | 61.2% | +LLM + Prometheus + анти-дубликат + аудит модерации |
| `internal/domain` | 65.4% | +ExtractionResult, +SubmissionReview |
| `internal/smtp` | 51.2% | CRLF-защита, retry |
| `internal/usecase` | **74.4%** | +6 новых тестов для submission pipeline |
| `internal/netutil` | 100% | 7 тестов |
| `internal/repository/pg` | Integration | ~90 тестов (seed-key, keyset-обход без потерь, NULL-скан, дубликаты Create/Update) |
| `internal/seeder` | Unit + Integration | 14 unit (фейки repo/storage) + round-trip seed→exists→seed→skipped |
| `cmd/seed-builder` | Unit | побайтовый детерминизм; отказ enum-алиасов/PNG/не-размеров/дублей |
| `test/e2e` | 8 тестов | ⚠️ **не компилируется на HEAD** (сигнатуры `NewSubmissionUseCase`/`NewAuthUseCase` разошлись с тестами) — pre-existing, чинить отдельным срезом |

Всего: ~400 тестов проходят ✅. Единственный известный flake: `TestAuthUseCase_Login_ConstantTimeForUnknownUser` (bcrypt-тайминг; падает и на чистом HEAD — окружение, не код).

⚠️ **Интеграционные pg-тесты делят одну БД** (`just test-db-up`, порт 5444): запускать с `-p 1` (конвенция `just test-int`), параллельные пакеты трункейтят таблицы друг друга.

---

## 9. Безопасность

*(H1-H4, M1-M6 без изменений — все закрыты в предыдущих сессиях)*

### 🆕 Новые механизмы в Спринте 7

| # | Механизм | Статус | Детали |
|---|---|---|---|
| 🆕 Анти-дубликат заявок | ✅ | SHA-256 нормализованного JSON + частичный unique index |
| 🆕 Аудит модерации | ✅ | `submission_reviews` с `reviewer_name`, `decision`, `comment` |
| 🆕 JWT claims в контексте | ✅ | `ContextWithClaims()` + `ClaimsFromContext()` для аудита |
| 🆕 Fallback reviewer="unknown" | ✅ | Защита от потери аудита при отсутствии JWT |

### 🆕 v0.0.8 / v0.0.8-1

| Механизм | Детали |
|---|---|
| Сид без секретов в образе | `seed.json` монтируется volume'ом (`:ro`), в image не копируется |
| Ключи MinIO без кириллицы | Детерминированный sha256-хеш от составного ключа — без проблем с Unicode-ключами S3 |
| Строгая валидация артефакта | Вся проверка данных на сборке (builder), рантайм проверяет только целостность base64/RIFF-WEBP |

---

## 10. Медиа (S3)

### Метрики

| Метрика | Описание |
|---------|----------|
| `emh_s3_operations_total{operation, status}` | Количество S3-операций |
| `emh_s3_duration_seconds{operation}` | Длительность S3-операций |
| `emh_s3_retries_total{operation}` | Повторные попытки |
| `emh_s3_bytes_transferred_total{operation, direction}` | Переданные байты (in/out) |

### 🆕 Изменения (v0.0.8-1)

- `MediaStorage.Upload` определяет размер для seekable-ридеров (`Seek(0, io.SeekEnd)` + rewind) и передаёт его в `PutObject` → мелкие объекты (ленты, превью, знаки наград) уходят **одним PUT** вместо multipart-карго (`POST ?uploads=` + parts + complete).
- Дерево бакета: `heroes/photos`, `submissions`, `awards/images`, `awards/ribbons` (последние два — из сида напрямую, мимо воркера).

---

## 11. Инфраструктура

### Dev (podman-compose)

```yaml
# podman-compose-hmr.yml
services:
  emh-backend:     # :3480 (API), :3443 (HTTPS)
  emh-postgres:    # :5443 → 5432
  emh-minio:       # :4443 (API), :4401 (Console)
```

### Production (Quadlet)

| Юнит | Образ | Особенности |
|---|---|---|
| `emh-backend.container` | `codeberg.org/thr0tt1e/emh-backend:latest` (**из v0.0.9**) | + `Volume=%h/emh/seeds:/app/seeds:ro` (авто-сид) |
| `emh-migrate.container` | `codeberg.org/thr0tt1e/emh-backend:latest` | goose up, oneshot, `After=emh-postgres` |
| `emh-frontend.container` | `codeberg.org/thr0tt1e/emh-frontend:latest` (**из v0.0.9**) | ADR-006 performance-контур |
| `emh-postgres.container` | — | — |
| `emh-minio.container` | — | — |
| `emh-caddy.container` | — | TLS терминирование, проксирование; `~/emh/Caddyfile` → `/etc/caddy/Caddyfile:ro` (volume). ADR-006: `encode zstd gzip` + S3 `Cache-Control`-матчеры. Деплой — validate + reload без пересборки |

- Теги `latest` собраны из релиза **v0.0.9**; CI: run 42 (v0.0.8), run 44 (v0.0.8-1), run 45 (v0.0.9).
- Пакеты: [emh-backend](https://codeberg.org/Thr0TT1e/-/packages/container/emh-backend/latest), [emh-frontend](https://codeberg.org/Thr0TT1e/-/packages/container/emh-frontend/latest).

---

## 12. Конфигурация

### Секция `submission_emails`

```yaml
submission_emails:
  enabled: false
  from_name: "Вечная память героям"
  approve_subject: "Ваша заявка одобрена"
  reject_subject: "Ваша заявка отклонена"
  queue_size: 100
  approve_template: |
    Здравствуйте, {{.SubmitterName}}!
    Ваша заявка на добавление данных о герое одобрена модератором.
    ID заявки: {{.SubmissionID}}
    ...
  reject_template: |
    Здравствуйте, {{.SubmitterName}}!
    К сожалению, ваша заявка отклонена.
    Причина: {{.ModeratorComment}}
    ...
```

Env overrides: `EMH_SUBMISSION_EMAILS_ENABLED`, `EMH_SUBMISSION_EMAILS_QUEUE_SIZE`

### Секция `auth_audit` (для cleanup)

```yaml
auth_audit:
  enabled: true
  queue_size: 1000
  retention: 2160h        # 90 дней
  cleanup_interval: 24h   # раз в сутки
```

### 🆕 Сид наград

| Env | Дефолт | Назначение |
|---|---|---|
| `AWARDS_SEED_PATH` | `/app/seeds/seed.json` | Путь к артефакту (bootstrap; субкоманда берёт путь из argv) |

---

## 13. Фоновые воркеры и метрики (обновлено)

| Воркер | Назначение | restartDelay | Метрики |
|---|---|---|---|
| thumbnail | Генерация превью + WebP | 5s | queue_size, processed_total, ... |
| thumbnail_rescan | Восстановление потерянных задач | 60s | rescan_triggered_total |
| orphan_cleanup | Удаление сиротских файлов (streaming) | 60s | scanned_total, deleted_total, pages_scanned_total |
| auth_audit | Аудит аутентификации | 5s | — |
| **auth_audit_cleanup** | Ротация auth_audit_log | 60s | `auth_audit_deleted_total`, `auth_audit_cleanup_duration_seconds` |
| **alert** | Алерты (SMTP + Telegram) | 5s | `alert_sent_total{type, channel, status}` |
| **submission_email** | Уведомления заявителям | 5s | `submission_email_sent_total{status}` |
| refresh_token_cleanup | Удаление истёкших токенов | 60s | expired_total |
| llm_log_cleanup | Очистка LLM-логов | 60s | llm_log_deleted_total |

---

## 14. Конвенции кода (обновлено)

| Конвенция | Детали |
|---|---|
| Маппинг ошибок | `errors.Is` через `mapDomainError` + `metrics.ErrorsTotal` |
| Sentinel errors | В `domain/errors.go`, новые через `NewAppError(code, message)` |
| Защита от циклов | `HasCyclicReference` через `WITH RECURSIVE` |
| Prometheus-метрики | Отдельный пакет `internal/metrics`, глобальный Registry |
| LLM провайдеры | Фабрика `NewLLMProvider(cfg)`, типизированный switch по Type |
| Retry паттерн | Экспоненциальный откат с джиттером для внешних сервисов |
| Курсорная пагинация | `LIMIT + 1`; для не-id-сортировки — keyset-кортеж в том же порядке, что ORDER BY (прецедент ListAwards: base64url-JSON `(sort_order,name,id)`) |
| `withRetry` в S3 | Принимает `operation string` для метрик |
| 🆕 Сид-строгость на сборке | Вся валидация данных (enum-канон, webp-магия, точные размеры, дубли) — в `cmd/seed-builder`; seeder проверяет только base64 + `content_type` + RIFF/WEBP (артефакт доверенный) |
| 🆕 Enum-канон сида | Строки proto (`AWARD_TYPE_ORDER`, `AWARD_JURISDICTION_RUSSIAN_FEDERATION`) через `emhv1.*_value` — единый источник правды, алиасы запрещены |
| 🆕 Частичный инсерт сида запрещён | MinIO down / битая картинка → запись целиком не инсертится (`Failed++`), иначе ON CONFLICT сделает её невидимой для повторного сида |
| 🆕 Ошибки справочника | `ErrAwardDuplicate` (23505 Create/Update) и `ErrInvalidCursor` → `mapDomainError` → AlreadyExists/InvalidArgument с русским сообщением |
| 🆕 Анти-дубликат | SHA-256 от нормализованного JSON (`json.Marshal` сортирует ключи) |
| 🆕 Best-effort операции | Аудит и email-уведомления не блокируют основной поток |
| 🆕 Worker graceful shutdown | `drainQueue()` для обработки in-flight после `ctx.Done()` |
| 🆕 Gzip compression | `connect.WithCompression("gzip", nil, nil)` — прозрачно для клиентов |

---

## 15. Подтверждённо работающее (Финальный чек-лист)

### Существующее (Спринты 1-6)

✅ Все пункты из предыдущих снапшотов

### 🆕 Новое в Спринте 7

#### #4: Ротация `auth_audit_log`

✅ `AuthAuditCleanupWorker` (ticker-based, паттерн `LLMLogCleanupWorker`)
✅ `DeleteOlderThan(ctx, olderThan)` в `AuthAuditRepository`
✅ Метрики: `AuthAuditDeletedTotal`, `AuthAuditCleanupDurationSeconds`
✅ Конфигурация: `auth_audit.retention` (90 дней), `auth_audit.cleanup_interval` (24h)
✅ Валидация: `retention >= 24h`, `cleanup_interval > 0`

#### #5: Async Alerter (Multi-Sender)

✅ `AlertWorker` (buffered channel, non-blocking `Enqueue`, graceful `Stop()`)
✅ `AlertSender` интерфейс + `SMTPAlertSender` + `TelegramAlertSender`
✅ `MultiAlertSender` (fan-out по всем настроенным каналам)
✅ `AuthAlertService` с интеграцией AlertWorker (brute-force detection)
✅ `Alert` структура: Type, Subject, Body, Timestamp, Metadata
✅ Метрика: `AlertSentTotal{type, channel, status}`
✅ `AuthUseCase.Refresh` → `AlertWorker.Enqueue` при детекте кражи токена

#### #6: Анти-дубликат + аудит модерации заявок

✅ Миграция 00025: `content_hash` + `submission_reviews` + частичный unique index
✅ `ErrSubmissionDuplicate` + `ErrCodeSubmissionDuplicate`
✅ `SubmissionReview`, `CreateSubmissionReviewParams` структуры
✅ `FindByContentHash` в `SubmissionRepository` (фильтр `status != REJECTED`)
✅ `CreateReview`, `ListReviews` в `SubmissionRepository`
✅ Нормализация JSON: `json.Marshal` сортирует ключи (детерминированный SHA-256)
✅ Best-effort аудит в `Review()` с fallback на `reviewer_name="unknown"`
✅ `ListSubmissionReviews` RPC в `SubmissionAdminService`
✅ Маппинг `ErrCodeSubmissionDuplicate → CodeAlreadyExists`
✅ 6 новых тестов: дубликат, нормализация, rejected can be resubmitted, аудит с/без JWT

#### #7: Email-уведомления заявителей

✅ `SubmissionEmailWorker` (отдельный воркер, buffered channel)
✅ `SubmissionEmailsConfig` в `config.go` (enabled, templates, queue_size)
✅ `SubmissionEmailNotification` структура (SubmitterName, Email, Decision, Comment)
✅ `renderEmailTemplate` с подстановкой `{{.SubmitterName}}`, `{{.SubmissionID}}`, etc.
✅ Интеграция в `SubmissionUseCase.Review()` (best-effort Enqueue)
✅ Graceful shutdown через `drainQueue()` + `wg.Wait()`
✅ Метрика: `SubmissionEmailSentTotal{status=approved|rejected|failed}`
✅ Env: `EMH_SUBMISSION_EMAILS_ENABLED`, `EMH_SUBMISSION_EMAILS_QUEUE_SIZE`

#### #8: Пагинация справочников + gzip compression

✅ **Gzip**: `connect.WithCompression("gzip", nil, nil)` в public/admin/extraction opts
✅ **Proto**: `PaginationRequest`/`PaginationResponse` в `ListConflictsRequest/Response`, `ListLocationsRequest/Response`, `ListAwardsRequest/Response`
✅ **Domain**: `ConflictFilter`, `LocationFilter`, `AwardFilter` с `Cursor`/`Limit`
✅ **Repository**: cursor-based `List(ctx, filter) → ([]*X, nextCursor, total, error)`
  - `COUNT(*)` для `total_count`
  - `LIMIT limit+1`, возврат `limit` записей + `nextCursor`
✅ **UseCase**: валидация `limit` (default 20, max 100)
✅ **Delivery**: маппинг pagination из proto в domain filter, возврат `PaginationResponse`
✅ **MCP**: `tools_conflicts.go` обновлён для 4 return values

### 🆕 Новое в v0.0.8 / v0.0.8-1 (сессия 2026-11-10/11, ADR-005 seed + дубли)

#### Сид справочника наград (v0.0.8, коммит `0aef72d`, тег `v0.0.8`)

✅ Артефакт `backend/seeds/awards/seed.json` (Git LFS, 5 пилотных наград), картинка — `{content_type, data}` (base64 StdEncoding, без префикса)
✅ `cmd/seed-builder`: строгая валидация — webp-магия, **точные** размеры 400×400/144×48, proto-канон enum, внутренние дубли, непустой список; побайтовый детерминизм
✅ `internal/seeder`: `SeedAwards` (skip существующих по ключу, skip всей записи при MinIO-down/битой картинке, race с админкой → Skipped+warn), `HasAnyAwards`
✅ Входы: dispatch `seed-awards` в `cmd/api/main.go` (шаг 2.5, exit 1 при failed>0) + `bootstrapAwardsSeed` (шаг 2.6, синхронно до `ListenAndServe`, ошибки не блокируют старт)
✅ Миграция 00028: `idx_awards_seed_key` UNIQUE `(name,type,jurisdiction)` (обычный, не partial)
✅ `ExistsByKey`/`InsertIfAbsent` в `AwardRepository`; 23505 → `ErrAwardDuplicate` → `CodeAlreadyExists` + русское сообщение (Create и Update)
✅ Деплой: `Volume=%h/emh/seeds:/app/seeds:ro`; `just build-seed`, `just seed-awards-vps` (one-shot контейнер, сеть emh-net)
✅ Данные: `AWARD_JURISDICTION_RF` → `RUSSIAN_FEDERATION`; пилотные картинки сконвертированы в канон (contain, белый фон, q85); phantom-ribbon у звезды Героя РФ удалён (`worn_without_bar=true`)
✅ E2E: seed-awards ×2 (inserted=5 → skipped=5); bootstrap на пустой БД (inserted=5); 9 PUT-объектов с детерминированными ключами
✅ CI: [run 42](https://codeberg.org/Thr0TT1e/emh/actions/runs/42/jobs/0/attempt/1) — success

#### Защита дублей в админке + keyset (v0.0.8-1, коммит `a172e1a`, тег `v0.0.8-1`)

✅ Keyset-курсор ListAwards: base64url-JSON `(sort_order,name,id)`, `WHERE (sort_order,name,id) > (?,?,?)`; старый id-курсор терял/дублировал записи
✅ `ErrInvalidCursor` → `CodeInvalidArgument` («Некорректный курсор пагинации»); ListAwards ошибки через `mapDomainError`
✅ `awardColumns`: COALESCE `description`/`image_url` (latent-500 «cannot scan NULL into *string» на nullable-колонках миграции 00002)
✅ `MediaStorage.Upload`: известный size для seekable-ридеров (один PUT вместо multipart)
✅ Тесты: 4 unit кодека курсора; интеграционные — обход 7 записей pageSize 2 без потерь/повторов, NULL-скан, дубликаты Create/Update
✅ CI: [run 44](https://codeberg.org/Thr0TT1e/emh/actions/runs/44/jobs/0/attempt/1) — success; образы `latest` опубликованы (backend + frontend)

---

## 16. ⚠️ ТЕХНИЧЕСКИЙ ДОЛГ (обновлено)

### 🆕 Закрыто в Спринте 7

| # | Проблема | Статус |
|---|---|---|
| #34 | Email-уведомления при модерации | ✅ Реализовано (`SubmissionEmailWorker`) |
| #35 | Анти-дубликат заявок | ✅ Реализовано (SHA-256 + partial unique index) |
| #36 | Аудит модерации | ✅ Реализовано (`submission_reviews`) |
| #12 | Async alerter | ✅ Реализовано (`AlertWorker` + MultiSender) |
| #14 | Ротация `auth_audit_log` | ✅ Реализовано (`AuthAuditCleanupWorker`) |
| #8 | Пагинация справочников + gzip | ✅ Реализовано (cursor-based + `connect.WithCompression`) |
| Migration bug | `status != 'ARCHIVED'` для smallint | ✅ Исправлено на `status != 3` |
| Submission bug | `p.Status` вместо `p.Decision` в Review | ✅ Исправлено (подтверждено) |

### 🆕 Закрыто в сессии v0.0.8 / v0.0.8-1

| # | Проблема | Статус |
|---|---|---|
| Сид справочника | Ручной ввод 400 наград непрактичен | ✅ Детерминированный пайплайн seed.json + bootstrap + one-shot |
| Дубли справочника | Админ мог создать дубль ключа сида | ✅ idx_awards_seed_key + AlreadyExists (сервер) + клиентский уровень на фронте |
| Курсор ListAwards | `id > cursor` при ORDER BY (sort_order,name,id) терял/дублировал записи | ✅ Keyset-кортеж (v0.0.8-1) |
| NULL-скан awards | ListAwards падал в 500 на строках с NULL image_url/description | ✅ COALESCE в awardColumns |
| Multipart для мелких объектов | Upload всегда size=-1 → multipart-карго | ✅ size из seekable-ридера |
| CI на аннотированный тег | v0.0.8 (annotated) не породил run | ✅ v0.0.8-1 — lightweight-тег; workflow отработал (run 44) |

### Осталось (актуально) — P1/P2

| # | Проблема | Ожидаемый эффект |
|---|---|---|
| #9 | Покрытие тестами до 80%+ | Долгосрочная стабильность (сейчас 74.4%) |
| #1 | LLMAdminService (из `llm_admin.proto`) | Управление провайдерами через админку |
| #2 | Интеграция MCP с Claude Desktop | Реальный end-to-end UX-тест |
| #3 | Тестовые сценарии для агентов | «Найди героев из Чечни», «Создай заявку» |
| — | `test/e2e` не компилируется на HEAD | Сигнатуры `NewSubmissionUseCase`/`NewAuthUseCase` разошлись с тестами — чинить отдельным срезом |
| — | Воркер rendition для `AWARD_IMAGE`/`AWARD_RIBBON` (ADR-005 §9, спека awards-seed-backend §9) | Админские загрузки наград в канон 400×400/144×48; сид грузит мимо воркера |
| — | Данные справочника ~400 записей | Пилот = 5; наполнение — задача историка/дизайнера (`docs/specs/awards-images-guide.md`) |
| — | E2E загрузки ленты через MediaService (открытый чекбокс спеки) | Требует MinIO + данных; воркер вне среза |
| — | `internal/smtp` покрытие до 70%+ | Сейчас 51.2% |

---

## 17. Следующие шаги

### ✅ Спринт 7 — ЗАКРЫТ

**Результат**: 6 задач закрыто (#4, #5, #6, #7, #8 + баг-фиксы)

### ✅ Сессия 2026-11-10/11 — ЗАКРЫТА

**Результат**: сид справочника наград (ADR-005) + защита дублей и keyset-пагинация (спеки `awards-seed-*`); теги `v0.0.8`, `v0.0.8-1`; образы `latest` опубликованы

### 🟡 Варианты следующего среза

**Вариант A: Воркер rendition для админских загрузок наград (ADR-005 §9)**

| # | Задача | Ожидаемый эффект |
|---|---|---|
| 1 | `processAwardImage` / `processAwardRibbon` в image-processor (400×400 / 144×48, contain, белый фон, q90) | Ручные загрузки наград в канон |
| 2 | E2E загрузки ленты через `MediaService` (закрыть чекбокс спеки) | Полный пайплайн медиа наград |

**Вариант B: LLMAdminService + интеграция с Claude Desktop**

**Вариант C: Покрытие тестами до 80%+ / `test/e2e` починка**

---

## 18. Шпаргалка команд (обновлено)

```fish
# Proto
just gen-sdk  # easyp lint -r . && easyp generate

# Миграции
just up | just down | just status

# Тесты
just test          # unit-тесты
just test-int      # интеграционные (⚠️ -p 1 внутри рецепта)
just test-all      # всё вместе
just cover         # coverage

# Dev
podman-compose -f infra/podman/hmr/podman-compose-hmr.yml up
podman logs -f emh

# 🆕 Сид наград
just build-seed        # собрать seed.json из seeds/awards (строгая валидация)
just seed-awards-vps   # one-shot импорт на VPS (seed-awards /app/seeds/seed.json)

# 🆕 Smoke: keyset-пагинация наград
curl -s http://localhost:3480/emh.v1.AwardService/ListAwards \
  -H 'Content-Type: application/json' \
  -d '{"pagination":{"pageSize":50,"cursor":""}}'

# MCP-сервер
go build -o emh-mcp-server ./cmd/mcp
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./emh-mcp-server 2>/dev/null | jq

# Метрики
curl -s http://localhost:3480/metrics | grep emh_

# Gzip smoke test
curl -s -H "Accept-Encoding: gzip" http://localhost:3480/emh.v1.ConflictService/ListConflicts -v
```

---

## 19. Принятые ключевые решения (обновлено)

| # | Решение | Обоснование |
|---|---|---|
| 1-87 | Из предыдущих снапшотов | — |
| 88-102 | Из Спринта 7 | см. предыдущий снимок |
| 🆕 103 | Seed-артефакт в Git LFS + standalone builder | Бинарные данные не в SQL; пересборка артефакта без пересборки образа |
| 🆕 104 | Ключ идемпотентности `(name,type,jurisdiction)`; **обычный** UNIQUE-индекс | Таргетный `ON CONFLICT` не матчится с partial-индексом (баг ТЗ) |
| 🆕 105 | MinIO down → запись не инсертится целиком | Частичный инсерт «без ленты» стал бы невидим для повторного сида (ON CONFLICT) |
| 🆕 106 | Объектный ключ `sha256(name␀type␀jurisdiction)[:16]` | Детерминизм, идемпотентность, без кириллицы в ключах S3 |
| 🆕 107 | Bootstrap-сид синхронно до `ListenAndServe` | Каталог готов к первому запросу; идемпотентность переживает рестарт посреди сида |
| 🆕 108 | `seed-awards` — субкоманда `cmd/api` (не отдельный бинарник) | Один образ, переиспользование config/db/storage wiring |
| 🆕 109 | Keyset-курсор для не-id сортировки (base64url-JSON кортеж) | `id > cursor` при `ORDER BY sort_order,name,id` теряет и дублирует записи |
| 🆕 110 | Курсор непрозрачный; `ErrInvalidCursor` → `InvalidArgument` | Клиент не парсит формат; ошибки клиента не маскируются под 500 |
| 🆕 111 | `Upload` передаёт известный size для seekable-ридеров | Один PUT вместо multipart-карго для мелких объектов |
| 🆕 112 | COALESCE для nullable-колонок в read-путях | Прецедент `ribbon_image_url`; NULL в `image_url`/`description` ронял весь List |

---

## 20. Новые/изменённые файлы

### Спринт 7

(см. предыдущий снимок — без изменений)

### 🆕 Сессия v0.0.8 / v0.0.8-1

**Новые файлы**

| Файл | Назначение |
|---|---|
| `backend/migrations/00028_awards_seed_key.sql` | UNIQUE `(name,type,jurisdiction)` |
| `backend/cmd/api/seed_awards.go` | `runSeedAwards` (one-shot) + `bootstrapAwardsSeed` |
| `backend/cmd/seed-builder/main.go` (+`main_test.go`) | Сборка seed.json, строгая валидация, детерминизм |
| `backend/internal/seeder/schema.go`, `awards.go` | SeedFile/SeedAward, SeedAwards, HasAnyAwards, objectHash |
| `backend/internal/seeder/awards_test.go`, `roundtrip_integration_test.go` | 14 unit + round-trip интеграционный |
| `backend/internal/repository/pg/award_cursor_test.go` | Unit кодека курсора (4 теста) |
| `backend/internal/repository/pg/award_seed_repository_test.go` | Интеграционные: ExistsByKey/InsertIfAbsent/дубликаты |
| `backend/internal/repository/pg/award_list_integration_test.go` | Keyset-обход без потерь, NULL-скан, InvalidCursor |
| `backend/seeds/awards/descriptions.json` | Исходные данные (5 пилотных) |
| `backend/seeds/awards/images/*.webp`, `ribbons/*.webp`, `seed.json` | LFS-ассеты + артефакт |
| `frontend/app/composables/useAwardCatalog.ts` | Полный каталог наград для дубль-чека |
| `docs/specs/awards-seed-backend.md`, `awards-seed-frontend.md` | ТЗ среза (Implemented + отклонения) |
| `docs/specs/awards-images-guide.md` | Гайд подготовки картинок |

**Изменённые файлы**

| Файл | Изменение |
|---|---|
| `backend/internal/domain/errors.go` | + `ErrAwardDuplicate`, `ErrInvalidCursor` |
| `backend/internal/domain/error_codes.go` | + `ErrCodeAwardDuplicate`, `ErrCodeInvalidCursor` |
| `backend/internal/domain/repository/award_repository.go` | + `ExistsByKey`, `InsertIfAbsent` |
| `backend/internal/repository/pg/award_repository.go` | Keyset-курсор, COALESCE-колонки, 23505-маппинг |
| `backend/internal/delivery/v1/award.go` | ListAwards → `mapDomainError` |
| `backend/internal/delivery/v1/errors.go` | + сообщения/коды AlreadyExists, InvalidArgument |
| `backend/internal/storage/s3/media_storage.go` | size для seekable-ридеров |
| `backend/cmd/api/main.go` | dispatch seed-awards (2.5) + bootstrap (2.6) |
| `backend/internal/usecase/mocks_test.go` | мок AwardRepository под новые методы |
| `backend/seeds/awards/descriptions.json` | enum-канон, phantom-ribbon удалён |
| `Justfile` | `build-seed`, `seed-awards-vps` |
| `infra/podman/podman_quadlets/emh-backend.container` | `:latest`, volume seeds |
| `infra/podman/podman_quadlets/emh-migrate.container` | `:latest` |
| `infra/podman/podman_quadlets/emh-frontend.container` | `:latest` |
| `docs/ADR/005-award-ribbon-bar.md` | уточнения seed + keyset |
| `docs/glossary.md` | звания = ORDER, seed-термины |

---

## 21. Статистика

| Метрика | Значение |
|---|---|
| Security уязвимостей закрыто | **10** (H1-H4, M1-M6) |
| Багов исправлено | **4** (Спринт 7) + **3** (keyset-курсор, NULL-скан, multipart) |
| Новых архитектурных решений | **25** (#88–112) |
| Миграций применено | **28** (добавлены 00026, 00027, 00028) |
| Всего тестов | **~400** |
| Покрытие `internal/usecase` | **74.4%** |
| Покрытие `internal/smtp` | **51.2%** |
| Новых воркеров | **4** (AlertWorker, AuthAuditCleanupWorker, SubmissionEmailWorker, AuthAlertService) |
| Новых метрик | **5** |

---

## 22. Рекомендации для следующей сессии

1. **Воркер rendition для наград** — единственный незакрытый чекбокс ADR-005 §9 (ручные загрузки через админку). Сид грузит мимо воркера, поэтому E2E загрузки ленты до сих пор открыт.
2. **Наполнение справочника ~400 записей** — данные готовит историк/дизайнер по `docs/specs/awards-images-guide.md`; пайплайн (`just build-seed` → `just seed-awards-vps`) уже готов и идемпотентен.
3. **Починка `test/e2e`** — сигнатуры usecase-конструкторов разошлись; отдельный маленький срез.
4. **Каверзные места курсоров**: если у ListConflicts/ListLocations появится сортировка не по `id` — keyset обязателен; ListAwards уже исправлен.
5. **Проверять коллацию**: `(name, type, jurisdiction)` — collation-зависимый unique; для регистро-независимых дублей понадобится отдельная нормализация.
