<p align="center">
  <h1 align="center">Вечная память героям — Backend</h1>
  <p align="center">
    <em>API-сервер мемориального проекта о героях, не вернувшихся с поля боя</em>
  </p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Connect-RPC-4285F4?logo=googlescholar&logoColor=white" alt="Connect RPC">
  <img src="https://img.shields.io/badge/PostgreSQL-18-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/MinIO-S3-C72C48?logo=minio&logoColor=white" alt="MinIO">
  <img src="https://img.shields.io/badge/Podman-Quadlet-892CA0?logo=podman&logoColor=white" alt="Podman">
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
</p>

---

## 📖 О проекте

Бэкенд-сервер проекта **«Вечная память героям»** — API для мемориального сайта, собирающего и хранящего сведения о героях, погибших в глобальных и локальных военных конфликтах: от Великой Отечественной войны до современных спецопераций.

Каждая запись — это судьба, собранная по крупицам из открытых источников, архивов, интервью с родственниками и сослуживцами.

> 🌐 **Live:** [neverforgotten.ru](https://neverforgotten.ru) • [вечнаяпамятьгероям.рус](https://вечнаяпамятьгероям.рус) • [вежливые.рус](https://вежливые.рус) • [emh.su](https://emh.su)

---

## ✨ Возможности

### Архитектурные
- **Clean Architecture** — изоляция домена от инфраструктуры
- **CQRS** — разделение команд (`HeroUseCase`) и запросов (`HeroQueryUseCase`)
- **Connect RPC** — HTTP+JSON API с авто-валидацией через `buf.validate`
- **Partial Update** — обновление отдельных полей через `field_mask`
- **Cursor-based пагинация** — на основе UUIDv7 для монотонной сортировки

### Функциональные
- **Гибкие даты** (`FlexibleDate`) — исторические даты с уровнем точности:
  - `EXACT` — «15 февраля 1994»
  - `MONTH` — «Февраль 1994»
  - `YEAR` — «1994»
  - `SEASON` — «Лето 1989»
  - `DAY_MONTH` — «28 июля» (день памяти без года)
  - `UNKNOWN` — дата неизвестна
- **Thumbnail Worker** — фоновая генерация превью 480×600 (4:5) с учётом `face_box` и конвертация в WebP
- **Orphan Cleanup** — автоматическое удаление сиротских файлов из S3
- **Presigned PUT URL** — прямая загрузка фото клиентами в MinIO, минуя бэкенд
- **Главное фото** — автоматическое назначение через триггеры БД + явное управление

### Безопасность
- **JWT + Refresh tokens** с ротацией и хранением хешей в БД
- **API-ключи** для внешних интеграций (LLM-пайплайн, MCP-сервер)
- **Rate limiting** — три группы (auth, submission, read)
- **Trusted proxies + CIDR** — защита от подделки `X-Forwarded-For`
- **IDOR protection** — все операции с фото проверяют принадлежность герою
- **Аудит аутентификации** — все попытки входа в `auth_audit_log`
- **Brute-force alerts** — алерты при превышении порога неудачных попыток

### Надёжность
- **Retry logic** для MinIO с exponential backoff
- **Транзакции** для batch-операций (`BatchAdd`, `DeleteByIDs`, `Reorder`)
- **`runWorkerWithRecovery`** — автоперезапуск фоновых воркеров при panic
- **Graceful shutdown** с ожиданием завершения текущих задач
- **Startup backfill** — восстановление thumbnails при перезапуске

---

## 🏗 Архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                         Caddy (TLS + CORS)                      │
│                         :443 (production)                       │
└──────────────────────────────┬──────────────────────────────────┘
                               │
            ┌──────────────────┼──────────────────┐
            ▼                  ▼                  ▼
   ┌────────────────┐  ┌───────────────┐  ┌─────────────────┐
   │   Frontend     │  │    Backend    │  │   MinIO (S3)    │
   │   Nuxt SSG     │  │  :3480 (RPC)  │  │  :9000 (S3 API) │
   │   (nginx)      │  │               │  │                 │
   └────────────────┘  └───────┬───────┘  └─────────────────┘
                               │
                    ┌──────────┼──────────┐
                    ▼          ▼          ▼
               ┌────────┐ ┌───────┐ ┌───────────┐
               │  Auth  │ │ Rate  │ │ Thumbnail │
               │ Inter. │ │ Limit │ │  Worker   │
               └────┬───┘ └───┬───┘ └───────────┘
                    │         │
                    ▼         ▼
              ┌──────────────────────┐
              │  PostgreSQL 18+      │
              │  (pgx pool)          │
              └──────────────────────┘
```

### Слои приложения

```
cmd/api/main.go                 # точка входа, wiring
internal/
├── domain/                     # сущности, sentinel errors, интерфейсы
│   └── repository/             # интерфейсы репозиториев
├── usecase/                    # бизнес-логика (CQRS)
├── delivery/
│   ├── v1/                     # Connect RPC серверы (14 сервисов)
│   ├── interceptor/            # AuthInterceptor
│   └── middleware/             # RateLimitMiddleware
├── repository/pg/              # PostgreSQL реализации
├── storage/s3/                 # MinIO MediaStorage
├── auth/                       # JWT, API-key, password
├── config/                     # YAML-лоадер
└── gen/emh/v1/                 # сгенерированный proto-код
migrations/                     # goose SQL-миграции
```

---

## 📦 Стек

| Компонент | Версия | Назначение |
|---|---|---|
| **Go** | 1.26.5 | Язык реализации |
| **Connect RPC** | v1.20.0 | HTTP+JSON API с типобезопасностью |
| **Protobuf** | v1.36.11 | Контракты API |
| **pgx/v5** | v5.10.0 | PostgreSQL драйвер с пулом |
| **Squirrel** | v1.5.4 | SQL-билдер |
| **chi + cors** | v5.3.1 | HTTP-роутер |
| **MinIO Go** | v7.2.1 | S3-совместимое хранилище |
| **JWT v5** | v5.3.1 | Access tokens |
| **nativewebp** | v1.3.0 | WebP encode (pure Go, lossless) |
| **imaging** | — | Crop/resize изображений |
| **goose** | — | SQL-миграции |
| **easyp** | — | Proto-линтинг и генерация |

---

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.26+
- Podman (или Docker) + podman-compose
- Just (task runner)

### 1. Клонирование

```bash
git clone https://codeberg.org/Thr0TT1e/emh.git
cd emh/backend
```

### 2. Конфигурация

Создайте файл `.env` в корне репозитория:

```env
DATABASE_URL=postgres://emh:emh@localhost:5432/emh?sslmode=disable
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
AUTH_JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

### 3. Запуск инфраструктуры (dev)

```bash
# PostgreSQL + MinIO + Backend + Frontend в HMR-режиме
just dev
```

Или через podman-compose напрямую:

```bash
podman-compose -f infra/podman/hmr/podman-compose-hmr.yml up
```

### 4. Применение миграций

```bash
just up    # goose up
```

### 5. Генерация API-клиентов

```bash
just gen-sdk    # easyp lint + easyp generate
```

Backend будет доступен на `http://localhost:3480`, frontend — на `http://localhost:3000`.

---

## ⚙️ Конфигурация

Основные параметры в `config/config.yaml`:

```yaml
app:
  name: emh-backend
  addr: :3480
  log_level: info
  cors_origins: ["http://localhost:3000"]

db:
  url: ${DATABASE_URL}
  max_conns: 20
  min_conns: 2

s3:
  endpoint: localhost:9000           # внутренний адрес
  public_endpoint: localhost:9000    # для presigned URL
  public_base_url: http://localhost:9000
  bucket: emh
  presign_expiry: 15m

auth:
  jwt_secret: ${AUTH_JWT_SECRET}
  jwt_expiry: 15m
  refresh_expiry: 168h
  admins:
    - username: admin
      password_hash: "$2a$12$..."    # bcrypt
      role: admin
  trusted_proxies: ["127.0.0.1", "::1"]

rate_limit:
  auth_limit: 5        # /мин
  submission_limit: 10 # /час
  read_limit: 120      # /мин

thumbnails:
  width: 480
  height: 600
  quality: 80
  workers: 2
  queue_size: 100

orphan_cleanup:
  enabled: true
  interval: 24h
  grace_period: 48h
  dry_run: false

auth_audit:
  enabled: true
  queue_size: 1000

auth_alerts:
  enabled: true
  threshold: 10
  window: 5m
```

Секреты переопределяются через переменные окружения (см. `pkg/env/env.go`).

---

## 🛠 Команды (Justfile)

```bash
just dev           # запустить dev-окружение (podman-compose hmr)
just gen-sdk       # proto: lint + generate
just up            # миграции: goose up
just down          # миграции: goose down
just status        # статус миграций
just build         # go build
just lint          # go vet + easyp lint
just hash-pass     # сгенерировать bcrypt-хеш пароля
```

### Хеш пароля админа

```bash
go run ./cmd/hashpassword 'your-password'
```

Скопируйте вывод в `config.yaml` в поле `admins[].password_hash`.

---

## 📝 Proto-контракты

Контракты API описаны в `proto/emh/v1/` и используют синтаксис proto3 с расширениями:

- **`buf.validate`** — декларативная валидация (min_len, max_len, uuid, pattern, enum)
- **easyp** — линтер и генератор (замена buf CLI)

### Основные сервисы

| Сервис | Назначение | Защита |
|---|---|---|
| `HeroService` | Публичное чтение карточек | — |
| `HeroAdminService` | CRUD героев и связей | JWT / API-key |
| `MediaService` | Presigned URL для загрузки | JWT |
| `AuthService` | Login / Refresh / Logout | частично |
| `AwardService` | Справочник наград | — |
| `ConflictService` | Справочник конфликтов | — |
| `LocationService` | Справочник локаций | — |
| `SubmissionService` | Заявки посетителей | — |
| `ApiKeyAdminService` | Управление API-ключами | JWT |

### Генерация

```bash
just gen-sdk
# эквивалентно:
easyp lint -r .
easyp generate
```

Сгенерированный код попадает в `internal/gen/emh/v1/` и `internal/gen/emh/v1/emhv1connect/`.

---

### Ключевые особенности схемы

- **UUIDv7** в качестве PK — монотонность для курсорной пагинации
- **JSONB** для `memberships`, `face_box`, `submission.payload`
- **Триггеры** на `photos`:
  - `trg_single_main_photo` — гарантирует одно главное фото
  - `trg_reassign_main_photo` — переназначает главное после удаления
- **CHECK-констрейнты** на `*_date_precision` (1..7)
- **Функциональные индексы** для быстрого поиска

---

## 🔌 API

Backend предоставляет Connect RPC API на порту `:3480`. Connect — это HTTP+JSON/gRPC-совместимый протокол, который работает из браузера без gRPC-Web прокси.

### Пример: создание героя

```bash
TOKEN=$(curl -s -X POST http://localhost:3480/emh.v1.AuthService/Login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<pass>"}' | jq -r '.accessToken')

curl -X POST http://localhost:3480/emh.v1.HeroAdminService/CreateHero \
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
  }'
```

### Пример: batch-загрузка фото

```bash
curl -X POST http://localhost:3480/emh.v1.MediaService/BatchGetUploadUrls \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": 1,
    "files": [
      {"filename": "photo1.jpg", "content_type": "image/jpeg"},
      {"filename": "photo2.png", "content_type": "image/png"}
    ]
  }'
```

Клиенты получают presigned URL и загружают файлы **напрямую в MinIO**, не нагружая бэкенд.

### Health check

```bash
curl http://localhost:3480/healthz
# {"status":"ok"}
```

---

## 🚢 Production Deploy

Используется **registry-centric deployment** через **Podman Quadlet** (systemd-юниты):

```
infra/podman/podman_quadlets/
├── emh-net.network              # bridge-сеть
├── emh-postgres.container       # PostgreSQL 18+
├── emh-minio.container          # MinIO
├── emh-migrate.container        # init: goose up
├── emh-backend.container        # Go API
├── emh-frontend.container       # Nginx (Nuxt SSG)
└── emh-caddy.container          # Reverse Proxy + Let's Encrypt
```

### Build & Push (dev-машина)

```bash
podman build -t docker.io/thr0tt1e/emh-backend:v1.0.0 \
  -f infra/podman/Podmanfile.back backend/
podman push docker.io/thr0tt1e/emh-backend:v1.0.0
```

### Deploy (VPS)

```bash
# На VPS только Quadlet-файлы + .env + Caddyfile
systemctl --user daemon-reload
systemctl --user start emh-caddy.service

# Обновление
podman pull docker.io/thr0tt1e/emh-backend:v1.1.0
systemctl --user restart emh-backend.service
```

---

## 📂 Структура директорий

```
backend/
├── cmd/
│   ├── api/main.go              # entrypoint
│   └── hashpassword/main.go     # CLI для bcrypt
├── config/
│   └── config.yaml              # конфигурация
├── internal/
│   ├── auth/                    # JWT, API-key, password
│   ├── config/                  # YAML-лоадер + env override
│   ├── delivery/                # Connect RPC + middleware
│   ├── domain/                  # сущности + интерфейсы
│   ├── gen/emh/v1/              # сгенерированный proto-код
│   ├── repository/pg/           # PostgreSQL реализации
│   ├── storage/s3/              # MinIO MediaStorage
│   └── usecase/                 # бизнес-логика + воркеры
├── migrations/                  # SQL-миграции (goose)
├── pkg/env/                     # env helpers
├── go.mod
└── go.sum
```

---

## 📜 Лицензия

MIT — подробности в [LICENSE](../LICENSE).

---

## 🔗 Связанные репозитории

- **[emh/frontend](https://codeberg.org/Thr0TT1e/emh/src/branch/master/frontend)** — Nuxt 4 + PrimeVue 4
- **[emh/infra](https://codeberg.org/Thr0TT1e/emh/src/branch/master/infra)** — Podman Quadlet, Caddy, compose-файлы
- **[emh/proto](https://codeberg.org/Thr0TT1e/emh/src/branch/master/proto)** — API-контракты

---

## 🤝 Содействие

Проект открыт для участия:
- **Источники данных** — архивные материалы, документы, фотографии
- **Верификация** — подтверждение фактов от родственников, историков, сослуживцев
- **Разработка** — см. [CONTRIBUTING.md](../CONTRIBUTING.md)

> *«Никто не забыт, ничто не забыто»*

---

<p align="center">
  <sub>Сделано с уважением к памяти павших героев</sub>
</p>
