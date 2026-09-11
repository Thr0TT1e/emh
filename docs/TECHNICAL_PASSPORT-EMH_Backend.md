# CONTEXT SNAPSHOT — EMH Backend

**Технический паспорт проекта «Вечная память героям»**

---

## 0. Метаданные снимка

| Параметр | Значение |
|---|---|
| Проект | Вечная память героям (EMH) |
| Go-модуль | `codeberg.org/Thr0TT1e/emh/backend` |
| Назначение снимка | Полное восстановление контекста бэкенд-разработки в новой сессии |
| Фаза | Бэкенд production-ready. Техдолг P0/P1 устранён. Frontend завершён. Deploy через Podman Quadlet. Следующий этап: тестирование, LLM-пайплайн, MCP-сервер |
| Миграции | 00001–00016 применены |
| Production VPS | `82.202.139.132`, домены: `neverforgotten.ru`, `вечнаяпамятьгероям.рус`, `вежливые.рус`, `emh.su` |

---

## 1. Цель проекта

Мемориальный сайт о героях, погибших в глобальных и локальных военных конфликтах. Наполнение — через админ-панель и краудсорсинг (заявки посетителей). Контент: короткие справки и развёрнутые биографии.

---

## 2. Стек (зафиксированные версии из `go.mod`)

### Backend

| Компонент | Версия | Назначение |
|---|---|---|
| Go | 1.26.5 | Язык |
| Connect RPC (`connectrpc.com/connect`) | v1.20.0 | API-протокол (HTTP+JSON) |
| `connectrpc.com/validate` | v0.6.0 | Авто-валидация по `buf.validate` |
| `google.golang.org/protobuf` | v1.36.11 | Protobuf runtime |
| `jackc/pgx/v5` | v5.10.0 | PostgreSQL драйвер + пул |
| `Masterminds/squirrel` | v1.5.4 | SQL-билдер |
| `go-chi/chi/v5` + `cors` | v5.3.1 / v1.2.2 | HTTP-роутер, CORS |
| `minio/minio-go/v7` | v7.2.1 | S3-клиент |
| `golang-jwt/jwt/v5` | v5.3.1 | JWT токены |
| `google/uuid` | v1.6.0 | Генерация UUID (API-key) |
| `golang.org/x/sync` | v0.22.0 | `errgroup` (параллельная загрузка) |
| `golang.org/x/crypto` | v0.54.0 | bcrypt для паролей/API-key |
| `golang.org/x/time` | — | rate-limiting (token bucket) |
| `golang.org/x/image` | — | WebP-декодер для `image.Decode` |
| `disintegration/imaging` | — | resize/crop изображений |
| `HugoSmits86/nativewebp` | v1.3.0 | WebP encode (lossless, pure Go) |
| `go.yaml.in/yaml/v3` | v3.0.5 | Парсинг `config.yaml` (форк) |

### Инструменты

| Инструмент | Назначение |
|---|---|
| easyp | Линтинг и генерация proto (замена buf) |
| goose | SQL-миграции |
| just (Justfile) | Таск-раннер |
| air | Hot-reload Go в контейнере (dev) |
| mkcert | Генерация TLS-сертификатов (dev) |
| Podman + Podman Quadlet | Контейнеризация (rootless, systemd) |
| Caddy 2 | Reverse Proxy + Let's Encrypt + CORS |

---

## 3. Архитектура

**Clean Architecture:** `domain` ← `usecase` ← `delivery`; `repository`/`storage` — инфраструктурные реализации.

**CQRS для героя:** `HeroUseCase` (команды) и `HeroQueryUseCase` (чтение агрегата `HeroDetail` с параллельной загрузкой 6 связей через `errgroup`: photos, awards, conflicts, locations, sources, relations).

**Connect RPC:** Публичные и админские сервисы разделены. Интерцепторы подключаются через `connect.WithInterceptors(...)` при `Mount`.

**Порядок интерцепторов:** `authInterceptor` первым (внешний), затем `validateInterceptor`.

**Middleware-стек:** CORS → RateLimit → chi Router → Connect handlers.

**Фоновые воркеры** (все запускаются через `runWorkerWithRecovery` с автоперезапуском при panic):
- **ThumbnailWorker** — генерация превью и конвертация оригиналов в WebP. In-memory очередь, N воркеров, startup backfill.
- **OrphanCleanupWorker** — удаление сиротских файлов из S3. Периодический, с grace period и dry-run режимом.
- **AuthAuditWorker** — асинхронная запись аудита аутентификации в БД.

**Обработка ошибок:** Типизированные sentinel errors в domain → коды ошибок для логирования/метрик → локализованные сообщения в delivery.

---

## 4. Структура директорий

```
backend/
 ├── cmd/api/main.go                 # точка входа, wiring 14 сервисов + middleware + воркеры
 ├── cmd/hashpassword/main.go        # генерация bcrypt-хеша пароля
 ├── config/config.yaml              # app/db/s3/auth/rate_limit/thumbnails/orphan_cleanup/auth_audit/auth_alerts
 ├── internal/
 │   ├── auth/                       # jwt.go, password.go, apikey.go, token.go
 │   ├── config/config.go            # лоадер конфига + env override
 │   ├── delivery/
 │   │   ├── interceptor/auth.go     # AuthInterceptor (JWT/API-key, trusted proxies, аудит, алерты)
 │   │   ├── middleware/ratelimit.go # RateLimitMiddleware
 │   │   └── v1/                     # 14 серверов + errors.go (локализация + маппинг)
 │   ├── domain/                     # сущности + errors.go + error_codes.go + flexible_date.go
 │   │   └── repository/             # интерфейсы репозиториев (включая AuthAuditRepository)
 │   ├── gen/emh/v1/                 # сгенерированный код (+emhv1connect)
 │   ├── repository/pg/              # реализации на pgx+squirrel (14 репозиториев)
 │   ├── storage/s3/media_storage.go # MinIO presigned URL + retry + ListObjects
 │   └── usecase/
 │       ├── thumbnail_worker.go     # ThumbnailWorker
 │       ├── orphan_cleanup_worker.go# OrphanCleanupWorker
 │       ├── auth_audit_worker.go    # AuthAuditWorker
 │       ├── auth_alert_service.go   # AuthAlertService (brute-force detection)
 │       ├── media_usecase.go        # MediaUseCase
 │       └── ...                     # остальные usecase
 ├── migrations/                     # goose 00001..00016
 └── pkg/env/env.go                  # Get(key, def)
```

---

## 5. Proto-контракты — статус

Пакет `emh.v1`. Линтер `easyp` требует leading-комментарии и `buf.validate`. Все файлы синхронизированы и сгенерированы.

| Файл | Статус | Изменения |
|---|---|---|
| `enums_emh.proto` | ✅ | **НОВОЕ:** `DatePrecision` enum (8 значений) |
| `common.proto` | ✅ | **НОВОЕ:** `FlexibleDate` message |
| `hero.proto` | ✅ | **НОВОЕ:** `birth_date_info`, `death_date_info`, `service_start_date_info` (FlexibleDate), старые Timestamp поля помечены DEPRECATED |
| `hero_admin.proto` | ✅ | **НОВОЕ:** `birth_date_info`, `death_date_info`, `service_start_date_info` в Create/Update |
| `award.proto`, `award_admin.proto` | ✅ | — |
| `conflict.proto`, `conflict_admin.proto` | ✅ | — |
| `location.proto`, `location_admin.proto` | ✅ | — |
| `media.proto` | ✅ | `BatchGetUploadUrls`, `FileToUpload`, `UploadUrlInfo` |
| `submission.proto` | ✅ | — |
| `auth.proto` | ✅ | `LoginResponse.refresh_token/expires_in`, Refresh, Logout RPC |
| `api_key_admin.proto` | ✅ | — |
| `extraction.proto` | 🔵 | Заготовка для LLM-пайплайна (P0) |

### Гибкие даты (FlexibleDate)

Решают проблему неполных дат в исторических источниках: «28 июля», «Февраль 1994», «Лето 1989», «1994».

| Precision | Значение | Пример | Anchor |
|---|---|---|---|
| `UNSPECIFIED` | 0 | — (не используется) | — |
| `EXACT` | 1 | `1994-02-15` | Полная дата |
| `MONTH` | 2 | `Февраль 1994` | Первое число месяца |
| `YEAR` | 3 | `1994` | 1 января |
| `SEASON` | 4 | `Лето 1989` | Условное начало сезона |
| `DAY_MONTH` | 5 | `28 июля` | Пустой |
| `RANGE` | 6 | `1994-1995` | Зарезервировано |
| `UNKNOWN` | 7 | Дата неизвестна | Пустой |

**Приоритет в API:** Если переданы и старое строковое поле (`birth_date`), и новое (`birth_date_info`), новое имеет приоритет.

---

## 6. Схема БД

PostgreSQL 18+. Первичные ключи — UUIDv7 (монотонность для курсорной пагинации).

Все 16 миграций (00001..00016) успешно применены.

| Таблица | Примечания |
|---|---|
| `heroes` | Все 16 полей + **гибкие даты:** `birth_date_precision`, `birth_date_display`, `death_date_precision`, `death_date_display`, `service_start_date_precision`, `service_start_date_display` |
| `awards`, `conflicts`, `locations` | Справочники с иерархией |
| `hero_awards`, `hero_conflicts`, `hero_locations`, `hero_sources`, `hero_relations` | Связи M:N |
| `photos` | Триггеры `trg_single_main_photo` + `trg_reassign_main_photo`, `face_box` JSONB |
| `submissions` | JSONB payload |
| `api_keys` | bcrypt `secret_hash` |
| `refresh_tokens` | Хранение хешей refresh-токенов |
| `auth_audit_log` | **НОВОЕ:** Аудит аутентификации (все попытки входа) |

### Триггеры на `photos`

| Триггер | Событие | Поведение |
|---|---|---|
| `trg_single_main_photo` | BEFORE INSERT/UPDATE OF is_main | Сбрасывает is_main у остальных фото героя |
| `trg_reassign_main_photo` | AFTER DELETE | Если удалено главное — назначает следующее по sort_order |

### Колонки гибких дат в `heroes`

| Колонка | Тип | Назначение |
|---|---|---|
| `*_date` | DATE | Anchor-дата для сортировки/фильтрации |
| `*_date_precision` | SMALLINT | Точность (1-7, CHECK constraint) |
| `*_date_display` | TEXT | Человекочитаемый текст из источника |

---

## 7. Статус Go-слоёв

| Слой | Статус |
|---|---|
| domain | ✅ Все сущности, `FlexibleDate`, `errors.go` (sentinel errors), `error_codes.go`, `auth_audit.go` |
| repository/pg | ✅ 14 репозиториев (включая `AuthAuditRepository`, транзакции в `BatchAdd`/`DeleteByIDs`/`Reorder`, IDOR protection) |
| usecase | ✅ Бизнес-логика (CQRS, ThumbnailWorker, OrphanCleanupWorker, AuthAuditWorker, AuthAlertService, валидация гибких дат) |
| delivery/v1 | ✅ Все 14 серверов + `errors.go` (локализация + типизированный маппинг ошибок) |
| delivery/interceptor | ✅ AuthInterceptor (JWT/API-key/static, trusted proxies + CIDR, аудит, алерты, логирование) |
| delivery/middleware | ✅ RateLimitMiddleware |
| storage/s3 | ✅ MediaStorage с retry logic (exponential backoff), `ListObjects` |
| auth | ✅ token.go (GenerateRefreshToken, HashRefreshToken) |
| cmd/api | ✅ Полный wiring + `runWorkerWithRecovery` + graceful shutdown |

---

## 8. Безопасность

### Аутентификация

| Механизм | Детали |
|---|---|
| Access token | JWT HS256, 15 минут, claims `{username, role}` |
| Refresh token | Random 256-bit (base64url), 7 дней, SHA-256 hash в БД, ротация при Refresh |
| API-ключи | Формат `emh_<key_id>_<secret>`, bcrypt `secret_hash` |
| Статические ключи | Аварийный доступ, логируются как WARN |
| Разделение | Публичные сервисы без auth; админские — `authInterceptor` + `validateInterceptor` |

### Trusted Proxies и защита от подделки IP

`AuthInterceptor` поддерживает список доверенных прокси (точные IP + CIDR-диапазоны). Заголовки `X-Forwarded-For` / `X-Real-IP` используются только если запрос пришёл от доверенного прокси. Иначе используется реальный адрес соединения из `req.Peer().Addr`.

### Аудит аутентификации

Все попытки аутентификации (успешные и неудачные) записываются в `auth_audit_log` через асинхронный `AuthAuditWorker` (buffered channel, не блокирует запросы).

Поля: `ip_address`, `mechanism` (static_key/api_key/jwt), `result` (success/failure), `identity`, `procedure`, `failure_reason`, `user_agent`.

### Алерты brute-force

`AuthAlertService` отслеживает неудачные попытки по IP. При превышении порога (по умолчанию 10 за 5 минут) логируется `ERROR` с `alert_type=auth_brute_force` для подключения к мониторингу.

### Логирование аутентификации

| Механизм | Уровень | Что логируется |
|---|---|---|
| Статический ключ | `WARN` | Факт аварийного доступа |
| API-ключ из БД | `INFO` | Имя ключа, роль, IP |
| JWT | `DEBUG` | Username, роль |
| Неудача | `WARN` | Причина (без токена) |

### Rate-limiting

| Группа | Эндпоинты | Лимит | Окно |
|---|---|---|---|
| Auth | `AuthService/*` | 5 req | 1 мин |
| Submission | `SubmissionService/*` | 10 req | 1 час |
| Read | Все публичные Get/List | 120 req | 1 мин |
| Admin | Все `*AdminService/*` | Без лимита (JWT) | — |
| Media | `MediaService/*` | Без лимита (JWT) | — |

In-memory `sync.Map` по IP + группа. Фоновая очистка записей. Skip для OPTIONS, `/healthz`, admin-сервисов, MediaService.

### Обработка ошибок

Типизированные sentinel errors в `domain/errors.go` → коды ошибок в `domain/error_codes.go` → локализованные русские сообщения в `delivery/v1/errors.go`.

| Код ошибки | Connect Code |
|---|---|
| `NOT_FOUND` | `CodeNotFound` |
| `DEATH_BEFORE_BIRTH`, `REQUIRED_FIELD_EMPTY`, `DUPLICATE_FIELD_MASK`, и др. | `CodeInvalidArgument` |
| `INTERNAL` | `CodeInternal` |

Клиенту возвращается локализованное сообщение. Оригинальная ошибка логируется с атрибутом `error_code` для метрик.

### IDOR Protection

Все операции с фото проверяют принадлежность герою:
- `Delete(ctx, heroID, photoID)` — `WHERE id = $1 AND hero_id = $2`
- `DeleteByIDs` — транзакция с предварительной проверкой принадлежности всех фото
- `Reorder` — транзакция с проверкой принадлежности
- `SetMain`, `Update` — `WHERE id = $1 AND hero_id = $2`

---

## 9. Медиа (S3)

### Загрузка

Паттерн Presigned PUT URL: `GetUploadUrl` / `BatchGetUploadUrls` → фронт грузит напрямую в MinIO через `s3.neverforgotten.ru` → `public_url` привязывается к записи.

### Retry Logic

Все операции с MinIO обёрнуты в `withRetry` с exponential backoff:
- 3 повторные попытки
- Задержки: 100ms → 200ms → 400ms (максимум 2s)
- Повторяются: сетевые ошибки, 5xx, 429
- Не повторяются: 4xx (кроме 429), отмена контекста

Покрыты: `Delete`, `Download`, `ListObjects`, `EnsureBucket`.

### Thumbnail-пайплайн

```
AddHeroPhoto / BatchAddHeroPhotos / UpdateHeroPhoto(face_box)
    │
    ├── DB INSERT/UPDATE (транзакция для batch)
    └── ThumbnailWorker.Enqueue()
            │
            ▼
    Background goroutine (N=2):
    1. Download оригинал из MinIO
    2. Decode image (image.Decode + blank imports)
    3. Crop: face_box → центральный 4:5 (fallback)
    4. Resize 480×600 (Fill, Lanczos)
    5. Encode WebP (lossless) → thumbnails/{photo_id}.webp
    6. Конвертация оригинала в WebP (если не WebP) → photos/{photo_id}.webp
    7. UPDATE photos SET url, thumbnail_url
```

**StartupBackfill:** При старте бэкенда находит все фото с пустым `thumbnail_url` и ставит их в очередь. Запускается асинхронно.

### Orphan Cleanup

`OrphanCleanupWorker` периодически:
1. Собирает все URL медиафайлов из БД
2. Сканирует объекты в MinIO через `ListObjects`
3. Удаляет файлы, не привязанные ни к одной записи и старше grace period

Параметры: `interval: 24h`, `grace_period: 48h`, `dry_run: false`.

### Главное фото — полная логика

| Событие | Поведение |
|---|---|
| Первое фото героя | Автоматически `is_main=true` (usecase) |
| Пакет без главного | `photos[0].IsMain = true` (репозиторий, внутри транзакции) |
| Удаление главного | Следующее по sort_order → главное (триггер БД) |
| Явная смена | `SetMainHeroPhoto` + триггер |
| Чтение в списках | `COALESCE(is_main, первое фото)` в SQL |
| Thumbnail для списков | `main_thumbnail_url` в `HeroSummary` |

---

## 10. Инфраструктура (Podman Quadlet + Caddy)

### Production Deployment Model

Продакшн-деплой использует Podman Quadlet (systemd-юниты) вместо `podman-compose`. Registry-centric deployment: образы из Docker Hub, на VPS только Quadlet-файлы, `.env` и `Caddyfile`.

### Quadlet-файлы (`infra/podman/podman_quadlets/`)

| Файл | Назначение |
|---|---|
| `emh-net.network` | Podman bridge-сеть для всех сервисов |
| `emh-postgres.container` | PostgreSQL 18.4+ (volume: `pg_data`) |
| `emh-minio.container` | MinIO Object Storage (volume: `minio_data`) |
| `emh-migrate.container` | Init-контейнер: запускает `goose up` и завершается |
| `emh-backend.container` | Go API-сервер (image из Docker Hub) |
| `emh-frontend.container` | Nginx со статикой Nuxt SSG (image из Docker Hub) |
| `emh-caddy.container` | Reverse Proxy + Let's Encrypt + CORS |

### Доменная архитектура

| Поддомен | Цель | Проксируется на |
|---|---|---|
| `neverforgotten.ru`, `www.neverforgotten.ru` | Публичный сайт | `emh-frontend:80` |
| `вечнаяпамятьгероям.рус`, `вежливые.рус`, `emh.su` | Альтернативные домены | `emh-frontend:80` |
| `api.neverforgotten.ru` | Backend Connect RPC API | `emh-backend:3480` |
| `s3.neverforgotten.ru` | MinIO S3 API (presigned URL) | `emh-minio:9000` + CORS headers |
| `console.minio.neverforgotten.ru` | MinIO Web Console | `emh-minio:9001` |

---

## 11. Конфигурация

`config.yaml` + `overrideFromEnv()`. Кастомный тип `Duration` для парсинга `15m` из YAML.

Секреты (`DATABASE_URL`, `MINIO_ROOT_USER/PASSWORD`, `AUTH_JWT_SECRET`) передаются через файл `.env` в Quadlet (`EnvironmentFile=%h/emh/.env`).

### Секции конфига

| Секция | Ключевые параметры |
|---|---|
| `app` | name, addr `:3480`, log_level, cors_origins |
| `db` | url, max_conns 20, min_conns 2 |
| `s3` | endpoint, public_endpoint, public_base_url, bucket, presign_expiry `15m`, use_ssl |
| `auth` | jwt_secret, jwt_expiry `15m`, refresh_expiry `168h`, admins, **trusted_proxies** |
| `rate_limit` | enabled, auth_limit 5/1m, submission_limit 10/1h, media_limit 20/1m, read_limit 120/1m |
| `thumbnails` | width 480, height 600, quality 80, workers 2, queue_size 100 |
| `orphan_cleanup` | **НОВОЕ:** enabled, interval `24h`, grace_period `48h`, dry_run |
| `auth_audit` | **НОВОЕ:** enabled, queue_size 1000 |
| `auth_alerts` | **НОВОЕ:** enabled, threshold 10, window `5m` |

---

## 12. Фоновые воркеры и recovery

Все воркеры запускаются через `runWorkerWithRecovery`:

| Воркер | Назначение | restartDelay |
|---|---|---|
| `thumbnail` | Генерация превью и конвертация в WebP | 5s |
| `orphan_cleanup` | Удаление сиротских файлов из S3 | 60s |
| `auth_audit` | Асинхронная запись аудита аутентификации | 5s |

**Поведение при panic:** Логируется panic + стек, воркер перезапускается через `restartDelay`. При отмене контекста (graceful shutdown) перезапуск не происходит.

**Startup backfill** для thumbnail worker запускается асинхронно, не блокирует старт сервера.

---

## 13. Конвенции кода

| Конвенция | Детали |
|---|---|
| Маппинг ошибок | Типизированные sentinel errors → `errors.Is()` → локализованные сообщения |
| Даты | Гибкие даты `FlexibleDate` (anchor + precision + display). Старые строковые поля поддерживаются для обратной совместимости |
| Partial update | `field_mask` + защита от очистки обязательных полей в usecase |
| Репозитории | `squirrel` + `ToSql()` + `pool.Query`. Транзакции для batch-операций |
| Echo-паттерн | `Add*` / `Remove*` / `SetMain*` возвращают запрос |
| Response-паттерн | `Create` / `Update` / `Delete` / `Batch*` возвращают отдельный Response |
| FaceBox | Координаты в пикселях оригинала от левого верхнего угла |
| Воркеры | Запуск через `runWorkerWithRecovery`, остановка через `workerCancel()` |

---

## 14. Подтверждённо работающее (Final Checklist)

✅ `go build ./...` и `go vet ./...` проходят чисто
✅ `easyp lint` и `easyp generate` проходят без ошибок
✅ Все 16 SQL-миграций применены
✅ CRUD героя со всеми полями + **гибкие даты**
✅ Связи M:N: Awards, Conflicts, Locations, Sources, Relations
✅ Пакетная работа с Photo: Add, BatchAdd, Delete, Reorder, SetMain, UpdateHeroPhoto
✅ **IDOR protection** во всех операциях с фото
✅ **Транзакции** в BatchAdd, DeleteByIDs, Reorder
✅ Автоназначение/переназначение главного фото (usecase + триггеры)
✅ COALESCE fallback в `ListHeroes`
✅ `GetHeroDetail` загружает 6 связей параллельно
✅ `UpdateHero` возвращает полный `HeroDetail`
✅ MinIO TLS и Presigned URL (dialer-redirect + HTTPS rewrite)
✅ **Retry logic** для операций с MinIO
✅ BatchGetUploadUrls — пакетная генерация presigned URL
✅ AuthInterceptor (JWT + API-key + static keys)
✅ **Trusted proxies с CIDR** защитой от подделки IP
✅ **Аудит аутентификации** (auth_audit_log)
✅ **Алерты brute-force** по IP
✅ **Логирование аутентификации** (все механизмы)
✅ Refresh-токены: Login → пара, Refresh → ротация, Logout → отзыв
✅ Rate-limiting: 3 группы, graceful shutdown, skip OPTIONS/healthz/admin/media
✅ SubmissionAdminService (модерация заявок)
✅ Production deploy через Podman Quadlet + Caddy + Docker Hub
✅ ThumbnailWorker — генерация превью и конвертация в WebP
✅ StartupBackfill — восстановление thumbnails при перезапуске
✅ **OrphanCleanupWorker** — удаление сиротских файлов из S3
✅ **runWorkerWithRecovery** — автоперезапуск воркеров при panic
✅ **Типизированные ошибки** с кодами и локализацией
✅ Frontend полностью готов и интегрирован

---

## 15. ⚠️ ТЕХНИЧЕСКИЙ ДОЛГ

| # | Проблема | Влияние | Приоритет |
|---|---|---|---|
| 1 | `nativewebp` lossless-only — файлы больше, чем при lossy | Расход места в S3, скорость загрузки | 🟡 P2 |
| 2 | `ListHeroPhotos` без пагинации (все фото одним запросом) | Проблемы при >100 фото | 🟢 P2 |
| 3 | Auto face-detect не реализован | Только ручной face_box + центральный crop | 🟢 P2 |
| 4 | Гибкие даты для `Award` и `Conflict` не реализованы | Награды и конфликты не могут иметь неполные даты | 🟡 P2 |
| 5 | `total_count` в ListHeroes только для первой страницы | Фича, не баг (оптимизация) | ℹ️ Документировано |
| 6 | **Тесты не написаны** (есть поэтапный план) | Нет уверенности при рефакторинге | 🔴 P1 |
| 7 | Ротация `auth_audit_log` не реализована | Таблица растёт бесконечно | 🟢 P2 |
| 8 | Batch-запись аудита (по одной записи) | Нагрузка на БД при высоком трафике | 🟢 P2 |
| 9 | Метрики (Prometheus/OpenTelemetry) не подключены | Мониторинг только через логи | 🟡 P2 |

---

## 16. Следующие шаги (Приоритеты для новой сессии)

| Приоритет | Задача | Детали |
|---|---|---|
| 🔴 P0 | **Тестирование** | Поэтапный план готов. Начать с domain unit-тестов (`FlexibleDate`, `HeroUseCase`), затем repository integration, delivery API tests |
| 🔴 P0 | LLM-пайплайн | `ExtractionService` + Ollama, human-in-the-loop предзаполнение формы героя из текста |
| 🟡 P1 | MCP-сервер | Инструменты для внешних LLM-агентов поверх Connect API |
| 🟡 P1 | Гибкие даты для наград | `award_date` → `FlexibleDate` в `AddHeroAwardRequest` |
| 🟡 P1 | Метрики | Prometheus-счётчики на основе `error_code` и `auth_audit_log` |
| 🟢 P2 | Auto face-detect | Fallback при отсутствии face_box |
| 🟢 P2 | Пагинация ListHeroPhotos | Для галерей >100 фото |
| 🟢 P2 | Lossy WebP | Переход с nativewebp (lossless) на lossy-кодек |
| 🟢 P2 | Ротация auth_audit_log | Удаление записей старше 90 дней |
| 🟢 P2 | Гибкие даты для конфликтов | `start_date`/`end_date` → `FlexibleDate` |

### Поэтапный план тестирования

| Этап | Содержание | Тип |
|---|---|---|
| 0 | Тестовая инфраструктура (Justfile, test DB, fixtures) | — |
| 1 | Domain unit: `FlexibleDate`, `Hero`, статусы | Unit |
| 2 | UseCase unit: `HeroUseCase`, `PhotoUseCase` с фейками | Unit |
| 3 | Repository integration: SQL, триггеры, курсоры, транзакции | Integration |
| 4 | Delivery API: маппинг, Connect errors, flexible dates | Integration |
| 5 | Auth: JWT, refresh, API-key, trusted proxies | Integration |
| 6 | Media: presigned URL, retry, thumbnail worker | Integration |
| 7 | E2E smoke: полный сценарий создания героя с фото | E2E |
| 8 | CI: автоматизация, coverage | — |

---

## 17. Шпаргалка команд

```bash
# Proto
just gen-sdk  # -> easyp lint -r . && easyp generate

# Миграции (локально)
just up | just down | just status

# Тесты
just test          # unit-тесты
just test-int      # интеграционные тесты
just test-all      # всё вместе
just cover         # coverage

# Локальная инфраструктура (dev HMR)
podman-compose -f infra/podman/hmr/podman-compose-hmr.yml up
podman logs -f emh

# Хеш пароля админа
go run ./cmd/hashpassword '<password>'

# Логин → пара токенов
curl -s -X POST http://localhost:3480/emh.v1.AuthService/Login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<pass>"}' | jq

# Создание героя с гибкой датой
curl -s -X POST http://localhost:3480/emh.v1.HeroAdminService/CreateHero \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Иван",
    "last_name": "Иванов",
    "status": 1,
    "birth_date_info": {
      "precision": 2,
      "anchor_date": "1994-02-01T00:00:00Z",
      "display_text": "Февраль 1994"
    },
    "death_date_info": {
      "precision": 5,
      "display_text": "28 июля"
    }
  }' | jq

# Обновление героя с field_mask
curl -s -X POST http://localhost:3480/emh.v1.HeroAdminService/UpdateHero \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "<uuid>",
    "birth_date_info": {
      "precision": 3,
      "anchor_date": "1995-01-01T00:00:00Z",
      "display_text": "1995"
    },
    "field_mask": ["birth_date_info"]
  }' | jq

# Очистка даты
curl -s -X POST http://localhost:3480/emh.v1.HeroAdminService/UpdateHero \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "<uuid>",
    "death_date_info": {"precision": 7},
    "field_mask": ["death_date_info"]
  }' | jq

# BatchGetUploadUrls (пакетная генерация presigned URL)
curl -s -X POST http://localhost:3480/emh.v1.MediaService/BatchGetUploadUrls \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": 1,
    "files": [
      {"filename": "photo1.jpg", "content_type": "image/jpeg"},
      {"filename": "photo2.png", "content_type": "image/png"}
    ]
  }' | jq

# Аудит аутентификации (запрос к БД)
psql "$DATABASE_URL" -c "
  SELECT created_at, ip_address, mechanism, result, identity, procedure
  FROM auth_audit_log
  ORDER BY created_at DESC
  LIMIT 20;
"

# Production (на VPS)
systemctl --user daemon-reload
systemctl --user start emh-caddy.service
systemctl --user status emh-backend.service
journalctl --user -u emh-backend.service -f

# Build & Push (на dev-машине)
podman build -t docker.io/thr0tt1e/emh-backend:vX.X.X -f infra/podman/Podmanfile.back backend/
podman push docker.io/thr0tt1e/emh-backend:vX.X.X

# Update на VPS
podman pull docker.io/thr0tt1e/emh-backend:vX.X.X
systemctl --user restart emh-backend.service
```

---

## 18. Принятые ключевые решения (все сессии)

| # | Решение | Обоснование |
|---|---|---|
| 1 | Presigned URL вместо стриминга через RPC | Разгрузка бэкенда |
| 2 | Два MinIO-клиента + dialer-redirect (SigV4 host) | Внутренний/внешний доступ |
| 3 | Курсор `id DESC` валиден благодаря UUIDv7 | Монотонность |
| 4 | Best-effort удаление S3 + orphan cleanup воркер | Надёжность + очистка |
| 5 | LLM не пишет в БД напрямую (только Submission) | Безопасность |
| 6 | Refresh-токены в БД (не stateless JWT) | Возможность отзыва, ротация |
| 7 | Access 15 мин + Refresh 7 дней | Баланс безопасности и UX |
| 8 | In-memory rate-limit (без Redis) | MVP, нет внешних зависимостей |
| 9 | `nativewebp` lossless | Pure Go, без CGO, без внешних утилит |
| 10 | Face box ручной, без auto-detect | Качество > автоматизация для мемориального проекта |
| 11 | Thumbnail 480×600 (4:5) с crop | Формат карточек фронтенда |
| 12 | `main_thumbnail_url` отдельно от `main_photo_url` | Thumbnail для карточек, оригинал для lightbox |
| 13 | Конвертация оригиналов в WebP q90 | Экономия трафика, единый формат |
| 14 | EXIF игнорируется | Упрощение, приватность |
| 15 | Podman Quadlet вместо podman-compose в проде | Интеграция с systemd, автоперезапуск, чистый VPS |
| 16 | Caddy как Reverse Proxy + TLS-терминатор | Авто Let's Encrypt, HTTP/3, единая точка входа |
| 17 | CORS для S3 API на уровне Caddy | Решение Mixed Content + IDN (Punycode) |
| 18 | HTTPS rewrite в presigned URL | SigV4 не подписывает схему |
| 19 | Registry-centric deployment (Docker Hub) | Чистый VPS без исходников, быстрый откат |
| 20 | Выделенный поддомен `s3.neverforgotten.ru` | Разделение веб-консоли и S3 API |
| 21 | `BatchGetUploadUrls` вместо N параллельных | Один запрос вместо 22, нет 429 |
| 22 | MediaService исключён из rate-limiting | Админы защищены JWT |
| 23 | Blank imports для `image.Decode` | Без них Go не декодирует изображения |
| 24 | StartupBackfill для восстановления thumbnails | При перезапуске бэкенда |
| 25 | **Гибкие даты (FlexibleDate)** вместо строгих YYYY-MM-DD | Исторические источники часто содержат неполные даты |
| 26 | **Типизированные sentinel errors** вместо `strings.Contains` | Устойчивость к рефакторингу, локализация, метрики |
| 27 | **Trusted proxies + CIDR** в AuthInterceptor | Защита от подделки X-Forwarded-For |
| 28 | **Асинхронный аудит аутентификации** через buffered channel | Не блокирует запросы |
| 29 | **`runWorkerWithRecovery`** для всех фоновых воркеров | Автоперезапуск при panic |
| 30 | **Retry logic** для MinIO с exponential backoff | Устойчивость к сетевым сбоям |
| 31 | **Транзакции** для batch-операций с фото | Атомарность, защита от race conditions |
| 32 | **IDOR protection** через `hero_id` во всех операциях с фото | Защита от подмены ID |

---

## 19. Новые файлы, созданные в текущей сессии

| Файл | Назначение |
|---|---|
| `internal/domain/flexible_date.go` | `FlexibleDate`, `DatePrecision`, валидация |
| `internal/domain/errors.go` | Sentinel errors |
| `internal/domain/error_codes.go` | `ErrorCode`, `CodeFromError` |
| `internal/domain/auth_audit.go` | `AuthAuditEntry`, константы |
| `internal/domain/repository/auth_audit_repository.go` | Интерфейс `AuthAuditRepository` |
| `internal/repository/pg/auth_audit_repository.go` | Реализация записи аудита |
| `internal/usecase/orphan_cleanup_worker.go` | Orphan cleanup S3 |
| `internal/usecase/auth_audit_worker.go` | Асинхронная запись аудита |
| `internal/usecase/auth_alert_service.go` | Brute-force detection |
| `internal/delivery/v1/errors.go` | Локализация + маппинг на Connect |
| `internal/delivery/v1/flexible_date.go` | Мапперы proto ↔ domain для гибких дат |
| `migrations/00015_add_flexible_date_columns.sql` | Колонки гибких дат в `heroes` |
| `migrations/00016_auth_audit_log.sql` | Таблица аудита аутентификации |
