# CONTEXT SNAPSHOT — EMH Backend (после Спринта 6)

## 0. Метаданные снимка

| Параметр | Значение |
|---|---|
| Проект | Вечная память героям (EMH) |
| Go-модуль | `codeberg.org/Thr0TT1e/emh/backend` |
| Назначение снимка | Полное восстановление контекста бэкенд-разработки в новой сессии |
| Фаза | 🟢 **Спринт 7 закрыт** (Submission Pipeline + Performance). Следующий этап: Спринт 8 (LLM Admin + MCP интеграция) |
| Миграции | 00001–00024 применены + **00025** (submission dedup + reviews) |
| Production VPS | 82.202.139.132, домены: neverforgotten.ru, вечнаяпамятьгероям.рус, вежливые.рус, emh.su |
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
| deepteams/webp | v0.4.1 | WebP encode/decode (lossy) |
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
│   ├── api/                          # Разделён на 5 файлов
│   │   ├── main.go                   # точка входа + graceful shutdown
│   │   ├── infra.go                  # connectDatabase, connectStorage, newLogger
│   │   ├── deps.go                   # wireDependencies (wiring 17 сервисов)
│   │   ├── workers.go                # startWorkers, runWorkerWithRecovery
│   │   └── server.go                 # buildRouter, mountServices + gzip
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
│   │   ├── errors.go                 # 42 sentinel error (+ ErrSubmissionDuplicate)
│   │   ├── error_codes.go            # + ErrCodeSubmissionDuplicate
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
│   │       └── award_repository.go       # List → cursor-based
│   ├── gen/emh/v1/                   # сгенерированный код
│   ├── llm/
│   ├── mcp/                          # 6 инструментов (обновлены для новых сигнатур List)
│   ├── metrics/                      # 10 файлов, 35+ метрик
│   ├── repository/pg/                # 16 репозиториев (обновлены для cursor)
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
│   └── migrations/                   # goose 00001..00025
└── test/e2e/
```

---

## 5. Proto-контракты — статус

| Файл | Статус | Изменения |
|---|---|---|
| `enums_emh.proto` | ✅ | DatePrecision enum |
| `common.proto` | ✅ | FlexibleDate, PaginationRequest/Response |
| `hero.proto` | ✅ | FlexibleDate поля, DEPRECATED Timestamp |
| `hero_admin.proto` | ✅ | FlexibleDate в Create/Update |
| `award.proto`, `award_admin.proto` | 🆕 | + PaginationRequest/Response в List |
| `conflict.proto`, `conflict_admin.proto` | 🆕 | + PaginationRequest/Response в List |
| `location.proto`, `location_admin.proto` | 🆕 | + PaginationRequest/Response в List |
| `media.proto` | ✅ | BatchGetUploadUrls |
| `submission.proto` | 🆕 | + SubmissionReview, SubmissionReviewDecision |
| `auth.proto` | ✅ | Refresh, Logout RPC |
| `api_key_admin.proto` | ✅ | — |
| `contact.proto` | ✅ | — |
| `extraction.proto` | ✅ | ExtractHeroData |
| `llm_admin.proto` | 🔵 Заготовка | Управление провайдерами |

---

## 6. Схема БД (обновлено)

PostgreSQL 18+. Первичные ключи — UUIDv7. **Все 25 миграций успешно применены.**

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
| **00025** | 🆕 `submissions.content_hash` + `submission_reviews` + частичный unique index |

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

-- Курсорная пагинация справочников (уже были PRIMARY KEY)
-- WHERE id > cursor использует PK индекс
```

---

## 7. Статус Go-слоёв (обновлено)

| Слой | Статус |
|---|---|
| domain | ✅ 42 sentinel error, `SubmissionReview`, `Alert`, фильтры для справочников |
| repository/pg | ✅ 16 репозиториев, cursor-based List для Conflict/Location/Award |
| usecase | ✅ + `SubmissionEmailWorker`, + `AlertWorker`, + `AuthAuditCleanupWorker` |
| delivery/v1 | ✅ 17 серверов + `mapDomainError` с метриками + gzip compression |
| delivery/interceptor | ✅ AuthInterceptor + `AuthAttemptsTotal` + `ContextWithClaims` |
| storage/s3 | ✅ `ListObjectsPaged`, countingReader, S3 метрики |
| smtp | ✅ `Send` с retry (4 попытки, экспоненциальный откат, джиттер) |
| llm/ | ✅ Ollama, OpenAI-compatible, Anthropic + ProviderManager |
| mcp/ | ✅ 6 инструментов (обновлены для cursor-based List) |
| metrics/ | ✅ 35+ метрик в 9 категориях |
| cmd/api/ | ✅ Разделён на 5 файлов |
| cmd/mcp/ | ✅ Отдельный бинарник |

---

## 8. Тестирование

Текущее покрытие (актуально после Спринта 7):

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
| `internal/repository/pg` | Integration | 42 теста |
| `test/e2e` | 8 тестов | Полный workflow |

Всего: ~365 теста, все проходят ✅

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

---

## 10. Медиа (S3) — без изменений

### Метрики

| Метрика | Описание |
|---------|----------|
| `emh_s3_operations_total{operation, status}` | Количество S3-операций |
| `emh_s3_duration_seconds{operation}` | Длительность S3-операций |
| `emh_s3_retries_total{operation}` | Повторные попытки |
| `emh_s3_bytes_transferred_total{operation, direction}` | Переданные байты (in/out) |

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

- `emh-backend.container`
- `emh-postgres.container`
- `emh-minio.container`
- Caddy: TLS терминирование, проксирование

---

## 12. Конфигурация

### 🆕 Новая секция `submission_emails`

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

### 🆕 Секция `auth_audit` (для cleanup)

```yaml
auth_audit:
  enabled: true
  queue_size: 1000
  retention: 2160h        # 90 дней
  cleanup_interval: 24h   # раз в сутки
```

---

## 13. Фоновые воркеры и метрики (обновлено)

| Воркер | Назначение | restartDelay | Метрики |
|---|---|---|---|
| thumbnail | Генерация превью + WebP | 5s | queue_size, processed_total, ... |
| thumbnail_rescan | Восстановление потерянных задач | 60s | rescan_triggered_total |
| orphan_cleanup | Удаление сиротских файлов (streaming) | 60s | scanned_total, deleted_total, pages_scanned_total |
| auth_audit | Аудит аутентификации | 5s | — |
| **auth_audit_cleanup** | 🆕 Ротация auth_audit_log | 60s | `auth_audit_deleted_total`, `auth_audit_cleanup_duration_seconds` |
| **alert** | 🆕 Алерты (SMTP + Telegram) | 5s | `alert_sent_total{type, channel, status}` |
| **submission_email** | 🆕 Уведомления заявителям | 5s | `submission_email_sent_total{status}` |
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
| Курсорная пагинация | `LIMIT + 1`, composite cursor для не-уникальных полей |
| `withRetry` в S3 | Принимает `operation string` для метрик |
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
  - `WHERE id > cursor ORDER BY ... LIMIT limit+1`
  - Возврат `limit` записей + `nextCursor` (последний ID)
✅ **UseCase**: валидация `limit` (default 20, max 100)
✅ **Delivery**: маппинг pagination из proto в domain filter, возврат `PaginationResponse`
✅ **MCP**: `tools_conflicts.go` обновлён для 4 return values
✅ **Производительность**: O(1) пагинация (WHERE id > cursor использует PRIMARY KEY)

---

## 16. ⚠️ ТЕХНИЧЕСКИЙ ДОЛГ (обновлено)

### 🆕 Закрыто в этой сессии (Спринт 7)

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

### Осталось (актуально) — P1/P2

| # | Проблема | Ожидаемый эффект |
|---|---|---|
| #9 | Покрытие тестами до 80%+ | Долгосрочная стабильность (сейчас 74.4%) |
| #1 | LLMAdminService (из `llm_admin.proto`) | Управление провайдерами через админку |
| #2 | Интеграция MCP с Claude Desktop | Реальный end-to-end UX-тест |
| #3 | Тестовые сценарии для агентов | «Найди героев из Чечни», «Создай заявку» |
| — | `internal/smtp` покрытие до 70%+ | Сейчас 51.2%, осталось ~18.8% |
| — | Интеграционные тесты для LLM-пайплайна | С реальным провайдером |

---

## 17. Следующие шаги

### ✅ Спринт 7 — ЗАКРЫТ

**Результат**: 6 задач закрыто (#4, #5, #6, #7, #8 + баг-фиксы)

### 🟡 Спринт 8 — варианты

**Вариант A: LLMAdminService + интеграция с Claude Desktop**

| # | Задача | Ожидаемый эффект |
|---|---|---|
| 1 | Реализация `LlmAdminService` (из `llm_admin.proto`) | Управление провайдерами через админку |
| 2 | Интеграция MCP с Claude Desktop | Реальный end-to-end UX-тест |
| 3 | Тестовые сценарии для агентов | «Найди героев из Чечни», «Создай заявку» |

**Вариант B: Покрытие тестами до 80%+**

| # | Задача | Ожидаемый эффект |
|---|---|---|
| 1 | Поднять `internal/usecase` с 74.4% → 80%+ | Долгосрочная стабильность |
| 2 | Поднять `internal/smtp` с 51.2% → 70%+ | Надёжность SMTP-слоя |
| 3 | Интеграционные тесты для LLM-пайплайна | Проверка end-to-end с реальным провайдером |

**Вариант C: Гибкие даты для оставшихся сущностей**

| # | Задача | Ожидаемый эффект |
|---|---|---|
| 21 | Гибкие даты для локаций (`birth_date_info`, `death_date_info`) | Полная историчность |
| 22 | Гибкие даты для источников (`publication_date_info`) | Точность источников |

---

## 18. Шпаргалка команд (без изменений)

```fish
# Proto
just gen-sdk  # easyp lint -r . && easyp generate

# Миграции
just up | just down | just status

# Тесты
just test          # unit-тесты
just test-int      # интеграционные
just test-all      # всё вместе
just cover         # coverage

# Dev
podman-compose -f infra/podman/hmr/podman-compose-hmr.yml up
podman logs -f emh

# MCP-сервер
go build -o emh-mcp-server ./cmd/mcp
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./emh-mcp-server 2>/dev/null | jq

# Метрики
curl -s http://localhost:3480/metrics | grep emh_

# Smoke test новых метрик Спринта 7
curl -s http://localhost:3480/metrics | grep -E "(auth_audit|alert|submission_email)" | sort

# Gzip smoke test
curl -s -H "Accept-Encoding: gzip" http://localhost:3480/emh.v1.ConflictService/ListConflicts -v
```

---

## 19. Принятые ключевые решения (обновлено)

| # | Решение | Обоснование |
|---|---|---|
| 1-87 | Из предыдущих снапшотов | — |
| 🆕 88 | SHA-256 нормализованного JSON для анти-дубликата | `json.Marshal` сортирует ключи → детерминированный хеш |
| 🆕 89 | Частичный unique index для активных заявок | `WHERE status != 3` — отклонённые можно переотправлять |
| 🆕 90 | Отдельная таблица `submission_reviews` вместо расширения `auth_audit_log` | Разные домены: security vs business workflow |
| 🆕 91 | Best-effort аудит/уведомления в `Review()` | Не блокирует HTTP-ответ модератора при ошибках |
| 🆕 92 | Fallback `reviewer_name="unknown"` при отсутствии JWT | Защита от потери аудита при нештатных вызовах |
| 🆕 93 | `ContextWithClaims()` хелпер для тестов | Публичная функция для unit-тестов без реального interceptor |
| 🆕 94 | `drainQueue()` для graceful shutdown воркеров | Обработка in-flight после `ctx.Done()`, не теряем сообщения |
| 🆕 95 | `context.Background()` для отправки email в воркерах | Отмена parent-ctx не обрывает in-flight отправки |
| 🆕 96 | Polling с deadline вместо `time.Sleep` в тестах | Детерминированные тесты без flaky sleep |
| 🆕 97 | Поиск по subject вместо индекса в email-тестах | Порядок недетерминирован из-за параллельных горутин |
| 🆕 98 | `connect.WithCompression("gzip", nil, nil)` | Прозрачное gzip-сжатие для всех Connect-хендлеров |
| 🆕 99 | Cursor-based пагинация для справочников | Стабильна при INSERT/DELETE, быстрый `WHERE id > cursor` |
| 🆕 100 | `LIMIT + 1` для определения `has_next_page` | Без дополнительного запроса, O(1) проверка |
| 🆕 101 | `COUNT(*)` для `total_count` в справочниках | Справочники небольшие (<1000), быстрый запрос |
| 🆕 102 | `status != 3` (число) вместо `'ARCHIVED'` (строка) | `status` хранится как smallint, не text |

---

## 20. Новые/изменённые файлы (Спринт 7)

### 🆕 Новые файлы

| Файл | Назначение |
|---|---|
| `backend/migrations/00025_submission_dedup_and_reviews.sql` | content_hash + submission_reviews |
| `backend/internal/domain/submission_email.go` | `SubmissionEmailNotification` |
| `backend/internal/domain/alert.go` | `Alert`, `AlertType`, `AlertSender` |
| `backend/internal/usecase/submission_email_worker.go` | Email-уведомления заявителям |
| `backend/internal/usecase/alert_worker.go` | Асинхронная отправка алертов |
| `backend/internal/usecase/auth_alert_service.go` | Brute-force detection + AlertWorker интеграция |
| `backend/internal/usecase/auth_audit_cleanup_worker.go` | Ротация auth_audit_log |
| `backend/internal/metrics/submission_email.go` | `SubmissionEmailSentTotal` |
| `backend/internal/metrics/alert.go` | `AlertSentTotal` |
| `backend/internal/metrics/auth_audit_cleanup.go` | `AuthAuditDeletedTotal`, `AuthAuditCleanupDurationSeconds` |
| `backend/internal/usecase/submission_email_worker_test.go` | 2 теста (approve/reject + drain) |
| `backend/internal/usecase/alert_worker_test.go` | Characterization tests |
| `backend/internal/usecase/auth_audit_cleanup_worker_test.go` | Characterization tests |
| `backend/internal/usecase/auth_alert_service_test.go` | Brute-force + cleanup tests |

### 🆕 Изменённые файлы

| Файл | Изменение |
|---|---|
| `backend/proto/emh/v1/submission.proto` | + `SubmissionReview`, `SubmissionReviewDecision`, `ListSubmissionReviews` RPC |
| `backend/proto/emh/v1/conflict.proto` | + `PaginationRequest`/`PaginationResponse` в List |
| `backend/proto/emh/v1/location.proto` | + `PaginationRequest`/`PaginationResponse` в List |
| `backend/proto/emh/v1/award.proto` | + `PaginationRequest`/`PaginationResponse` в List |
| `backend/internal/domain/submission.go` | + `ContentHash`, `SubmissionReview`, `CreateSubmissionReviewParams` |
| `backend/internal/domain/errors.go` | + `ErrSubmissionDuplicate` |
| `backend/internal/domain/error_codes.go` | + `ErrCodeSubmissionDuplicate` |
| `backend/internal/domain/repository/submission_repository.go` | + `FindByContentHash`, `CreateReview`, `ListReviews` |
| `backend/internal/domain/repository/auth_audit_repository.go` | + `DeleteOlderThan` |
| `backend/internal/domain/repository/conflict_repository.go` | `List` → cursor-based |
| `backend/internal/domain/repository/location_repository.go` | `List` → cursor-based |
| `backend/internal/domain/repository/award_repository.go` | `List` → cursor-based |
| `backend/internal/repository/pg/submission_repository.go` | Реализация новых методов |
| `backend/internal/repository/pg/auth_audit_repository.go` | Реализация `DeleteOlderThan` |
| `backend/internal/repository/pg/conflict_repository.go` | Cursor-based List + COUNT |
| `backend/internal/repository/pg/location_repository.go` | Cursor-based List + COUNT |
| `backend/internal/repository/pg/award_repository.go` | Cursor-based List + COUNT |
| `backend/internal/usecase/submission_usecase.go` | Нормализация + хеш + дубликат + аудит |
| `backend/internal/usecase/conflict_usecase.go` | `List` → `ConflictFilter` |
| `backend/internal/usecase/location_usecase.go` | `List` → `LocationFilter` |
| `backend/internal/usecase/award_usecase.go` | `List` → `AwardFilter` |
| `backend/internal/usecase/auth_usecase.go` | + `alertWorker` для theft detection |
| `backend/internal/delivery/interceptor/auth.go` | + `ContextWithClaims()` |
| `backend/internal/delivery/v1/submission_admin.go` | + `ListSubmissionReviews` |
| `backend/internal/delivery/v1/conflict.go` | + pagination mapping |
| `backend/internal/delivery/v1/location.go` | + pagination mapping |
| `backend/internal/delivery/v1/award.go` | + pagination mapping |
| `backend/internal/delivery/v1/errors.go` | + `ErrCodeSubmissionDuplicate → CodeAlreadyExists` |
| `backend/internal/config/config.go` | + `SubmissionEmailsConfig` + валидация |
| `backend/internal/mcp/tools_conflicts.go` | Обновлён для 4 return values |
| `backend/cmd/api/deps.go` | + wiring новых воркеров |
| `backend/cmd/api/workers.go` | + запуск новых воркеров |
| `backend/cmd/api/main.go` | + graceful shutdown для новых воркеров |
| `backend/cmd/api/server.go` | + `connect.WithCompression("gzip", nil, nil)` |
| `backend/config.example.yaml` | + `submission_emails` секция |

---

## 21. Статистика

| Метрика | Значение |
|---|---|
| Security уязвимостей закрыто | **10** (H1-H4, M1-M6) |
| Багов исправлено | **4** (SubmissionUseCase.Review, OrphanCleanupWorker data race, migration ARCHIVED, MCP tools) |
| Новых архитектурных решений | **15** (#88-102) |
| Миграций применено | **25** (было 22, добавлены 00023, 00024, 00025) |
| Всего тестов | **~365** (было ~303, добавлено ~62) |
| Покрытие `internal/usecase` | **74.4%** (было 60.5%, +13.9%) |
| Покрытие `internal/smtp` | **51.2%** (было 42.4%, +8.8%) |
| Задач закрыто в Спринте 7 | **6** (#4, #5, #6, #7, #8 + баг-фиксы) |
| Новых воркеров | **4** (AlertWorker, AuthAuditCleanupWorker, SubmissionEmailWorker, AuthAlertService) |
| Новых метрик | **5** (AlertSentTotal, SubmissionEmailSentTotal, AuthAuditDeletedTotal, AuthAuditCleanupDurationSeconds, ...) |

---

## 22. Рекомендации для следующей сессии

### Спринт 8 (Вариант A): LLMAdminService + интеграция

1. Реализация `LlmAdminService` (из `llm_admin.proto`)
2. Интеграция MCP с Claude Desktop
3. Тестовые сценарии для агентов

### Спринт 8 (Вариант B): Покрытие тестами

1. Поднять `internal/usecase` с 74.4% → 80%+
2. Поднять `internal/smtp` с 51.2% → 70%+
3. Интеграционные тесты для LLM-пайплайна

### Спринт 8 (Вариант C): Гибкие даты

1. Локации (`birth_date_info`, `death_date_info`)
2. Источники (`publication_date_info`)

### Опционально

- Выполнить `git filter-repo` для очистки коммита `f4a38be` из публичной истории Codeberg
- Создать Grafana-дашборд для визуализации 45+ метрик
- Добавить rate limiting для SubmissionService (анти-спам)

---

🎯 **Спринт 7 закрыт. 6 задач + 4 бага исправлено. Submission pipeline hardened. Gzip compression enabled. Coverage 74.4%. Готов к Спринту 8.** 🚀
