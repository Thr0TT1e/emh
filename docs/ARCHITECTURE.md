# ARCHITECTURE.md — Детальное описание слоёв EMH Backend

<p align="center">
  <h1 align="center">Архитектура EMH Backend</h1>
  <p align="center">
    <em>Детальное описание слоёв, паттернов и потоков данных</em>
  </p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Architecture-Clean_Architecture-blue" alt="Architecture">
  <img src="https://img.shields.io/badge/Pattern-CQRS-orange" alt="Pattern">
  <img src="https://img.shields.io/badge/Protocol-Connect_RPC-green" alt="Protocol">
</p>

---

## 📋 Содержание

1. [Обзор архитектуры](#-обзор-архитектуры)
2. [Принципы проектирования](#-принципы-проектирования)
3. [Слой Domain](#-слой-domain)
4. [Слой UseCase](#-слой-usecase)
5. [Слой Delivery](#-слой-delivery)
6. [Слой Repository](#-слой-repository)
7. [Слой Storage](#-слой-storage)
8. [Слой Auth](#-слой-auth)
9. [Слой Config](#-слой-config)
10. [Ключевые паттерны](#-ключевые-паттерны)
11. [Потоки данных](#-потоки-данных)
12. [Фоновые воркеры](#-фоновые-воркеры)
13. [Безопасность](#-безопасность)
14. [Производительность](#-производительность)

---

## 🏗 Обзор архитектуры

### Высокоуровневая диаграмма

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Клиенты                                        │
│                     (Frontend, MCP, LLM-пайплайн)                           │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Caddy (Reverse Proxy)                               │
│                     TLS + CORS + Routing по Host                            │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          chi Router                                         │
│                     CORS → RateLimit → Connect Handlers                     │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    ▼                    ▼
┌─────────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   AuthInterceptor   │  │ ValidateInter.  │  │  RateLimitMW    │
│  (JWT/API-key)      │  │  (buf.validate) │  │  (3 группы)     │
└─────────┬───────────┘  └─────────────────┘  └─────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Delivery Layer (v1)                                  │
│   14 Connect RPC серверов + мапперы domain ↔ proto + errors.go              │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         UseCase Layer                                       │
│   HeroUseCase (CQRS) + ThumbnailWorker + OrphanCleanup + AuthAudit          │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    ▼                    ▼
┌─────────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   Repository Layer  │  │  Storage Layer  │  │   Auth Layer    │
│   (PostgreSQL)      │  │    (MinIO)      │  │  (JWT/API-key)  │
└─────────────────────┘  └─────────────────┘  └─────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Domain Layer                                     │
│   Сущности + интерфейсы + errors + FlexibleDate (чистый Go)                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Зависимости слоёв

```
domain (не зависит ни от кого)
   ▲
   │
usecase (зависит от domain + интерфейсов repository)
   ▲
   │
delivery (зависит от usecase + proto-контрактов)
   ▲
   │
infrastructure (реализации: repository/pg, storage/s3)
```

**Правило зависимостей:** стрелки всегда направлены внутрь. Внешние слои зависят от внутренних, но не наоборот.

---

## 🎯 Принципы проектирования

| Принцип | Реализация в EMH |
|---|---|
| **Single Responsibility** | Каждый слой решает одну задачу |
| **Dependency Inversion** | UseCase зависит от интерфейсов, не от реализаций |
| **Interface Segregation** | Узкие интерфейсы (например, `APIKeyAuthenticator`) |
| **Clean Architecture** | Domain изолирован от инфраструктуры |
| **CQRS** | Разделение команд и запросов для героя |
| **Fail Fast** | Валидация на границах слоёв |
| **Best Effort** | Удаление S3 не блокирует удаление из БД |
| **Graceful Degradation** | Воркеры перезапускаются при panic |

---

## 📦 Слой Domain

**Расположение:** `internal/domain/`

**Назначение:** Чистая бизнес-логика без зависимостей от инфраструктуры.

Слой **Domain** (слой предметной области) — это сердце приложения на Go, где сосредоточена вся основная бизнес-логика. Он отвечает за правила и сущности, которые описывают саму суть решаемой задачи, независимо от внешних деталей.

### Структура

```
internal/domain/
├── hero.go              # Hero, HeroDetail, CreateHeroParams, UpdateHeroParams
├── photo.go             # Photo, AddPhotoParams, UpdatePhotoParams, FaceBox
├── award.go             # Award, HeroAward
├── conflict.go          # Conflict, HeroConflict
├── location.go          # Location, HeroLocation
├── source.go            # HeroSource, AddHeroSourceParams
├── relation.go          # HeroRelation, AddHeroRelationParams
├── submission.go        # Submission
├── auth.go              # AdminUser, APIKey, Claims
├── refresh_token.go     # RefreshToken
├── flexible_date.go     # FlexibleDate, DatePrecision
├── errors.go            # Sentinel errors
├── error_codes.go       # ErrorCode, CodeFromError
├── auth_audit.go        # AuthAuditEntry
├── media_storage.go     # Интерфейс MediaStorage, ObjectInfo
└── repository/          # Интерфейсы репозиториев
    ├── hero_repository.go
    ├── photo_repository.go
    ├── award_repository.go
    ├── conflict_repository.go
    ├── location_repository.go
    ├── source_repository.go
    ├── relation_repository.go
    ├── submission_repository.go
    ├── refresh_token_repository.go
    ├── api_key_repository.go
    └── auth_audit_repository.go
```

### Ключевые сущности

#### Hero

```go
type Hero struct {
    ID               string
    FirstName        string
    LastName         string
    MiddleName       string
    ShortBio         string
    FullBio          string
    Rank             string
    BirthDate        FlexibleDate
    DeathDate        FlexibleDate
    ServiceStartDate FlexibleDate
    Status           PublicationStatus
    CreatedAt        time.Time
    UpdatedAt        time.Time
    Nickname         string
    Unit             string
    Position         string
    ServiceBranch    string
    CauseOfDeath     string
    Memberships      []string

    // Денормализация для списков
    MainPhotoURL     *string
    MainThumbnailURL *string
    AwardNames       []string
}
```

#### FlexibleDate

Решает проблему неполных исторических дат:

```go
type FlexibleDate struct {
    Anchor      *time.Time    // Опорная дата для сортировки
    Precision   DatePrecision // Уровень точности
    DisplayText string        // Текст из источника
}

type DatePrecision int

const (
    PrecisionUnspecified DatePrecision = iota // 0 - не используется
    PrecisionExact                            // 1 - YYYY-MM-DD
    PrecisionMonth                            // 2 - Месяц YYYY
    PrecisionYear                             // 3 - YYYY
    PrecisionSeason                           // 4 - Сезон YYYY
    PrecisionDayMonth                         // 5 - DD Месяц (без года)
    PrecisionRange                            // 6 - Зарезервировано
    PrecisionUnknown                          // 7 - Дата неизвестна
)
```

**Правила валидации:**
- `EXACT` требует `Anchor`
- `DAY_MONTH` и `UNKNOWN` не должны иметь `Anchor`
- `MONTH`, `YEAR`, `SEASON` используют `Anchor` для сортировки

#### HeroDetail (Read-модель)

```go
type HeroDetail struct {
    Hero      *Hero
    Photos    []*Photo
    Awards    []*HeroAward
    Conflicts []*HeroConflict
    Locations []*HeroLocation
    Sources   []*HeroSource
    Relations []*HeroRelation
}
```

### Типизированные ошибки

**Файл:** `domain/errors.go`

```go
var (
    ErrNotFound                      = errors.New("not found")
    ErrDeathBeforeBirth              = errors.New("death date cannot be before birth date")
    ErrServiceBeforeBirth            = errors.New("service start date cannot be before birth date")
    ErrServiceAfterDeath             = errors.New("service start date cannot be after death date")
    ErrRequiredFieldEmpty            = errors.New("required field cannot be empty")
    ErrStatusUnspecified             = errors.New("status cannot be unspecified")
    ErrStatusRequired                = errors.New("status is required when field_mask contains status")
    ErrDuplicateFieldMask            = errors.New("field_mask contains duplicate field")
    ErrPhotosNotBelongToHero         = errors.New("some photos do not belong to hero")
    ErrSelfRelation                  = errors.New("cannot create relation to itself")
    ErrInvalidDateFormat             = errors.New("invalid date format")
    ErrInvalidDatePrecision          = errors.New("invalid date precision")
    ErrExactDateRequiresAnchor       = errors.New("exact date requires anchor")
    ErrUnknownDateMustNotHaveAnchor  = errors.New("unknown date must not have anchor")
    ErrDayMonthMustNotHaveAnchor     = errors.New("day_month date must not have anchor")
    ErrHeroIDRequired                = errors.New("hero id is required")
    ErrPhotoIDsEmpty                 = errors.New("photo_ids must not be empty")
    ErrPhotosEmpty                   = errors.New("photos must not be empty")
    ErrPhotoURLRequired              = errors.New("photo url is required")
)
```

**Файл:** `domain/error_codes.go`

```go
type ErrorCode string

const (
    ErrCodeInternal             ErrorCode = "INTERNAL"
    ErrCodeNotFound             ErrorCode = "NOT_FOUND"
    ErrCodeDeathBeforeBirth     ErrorCode = "DEATH_BEFORE_BIRTH"
    // ... остальные коды
)

// CodeFromError возвращает доменный код ошибки для логирования и метрик.
func CodeFromError(err error) ErrorCode {
    switch {
    case errors.Is(err, ErrNotFound):
        return ErrCodeNotFound
    case errors.Is(err, ErrDeathBeforeBirth):
        return ErrCodeDeathBeforeBirth
    // ...
    default:
        return ErrCodeInternal
    }
}
```

### Интерфейсы репозиториев

**Принцип:** UseCase зависит от интерфейсов, не от реализаций.

```go
// domain/repository/hero_repository.go
type HeroRepository interface {
    Create(ctx context.Context, p domain.CreateHeroParams) (string, error)
    Update(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error)
    Delete(ctx context.Context, id string, hardDelete bool) error
    GetByID(ctx context.Context, id string) (*domain.Hero, error)
    List(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error)
}
```

**Преимущества:**
- UseCase тестируется с фейками
- Реализация БД может быть заменена без изменения бизнес-логики
- Чёткий контракт между слоями

---

## ⚙️ Слой UseCase

**Расположение:** `internal/usecase/`

**Назначение:** Бизнес-логика, оркестрация репозиториев, фоновые воркеры.

Слой **UseCase** (бизнес-логики) — это место, где реализуется вся бизнес-логика. Его задача — описать высокоуровневые бизнес-процессы, не привязываясь к техническим деталям (например, к HTTP-запросам или базе данных).

### Структура

```
internal/usecase/
├── hero_usecase.go            # HeroUseCase (команды)
├── hero_query_usecase.go      # HeroQueryUseCase (чтение)
├── photo_usecase.go           # PhotoUseCase
├── award_usecase.go           # AwardUseCase
├── hero_award_usecase.go      # HeroAwardUseCase
├── conflict_usecase.go        # ConflictUseCase
├── hero_conflict_usecase.go   # HeroConflictUseCase
├── location_usecase.go        # LocationUseCase
├── hero_location_usecase.go   # HeroLocationUseCase
├── hero_source_usecase.go     # HeroSourceUseCase
├── hero_relation_usecase.go   # HeroRelationUseCase
├── submission_usecase.go      # SubmissionUseCase
├── auth_usecase.go            # AuthUseCase
├── api_key_usecase.go         # APIKeyUseCase
├── media_usecase.go           # MediaUseCase
├── thumbnail_worker.go        # ThumbnailWorker
├── orphan_cleanup_worker.go   # OrphanCleanupWorker
├── auth_audit_worker.go       # AuthAuditWorker
└── auth_alert_service.go      # AuthAlertService
```

### CQRS для героя

**Команды (HeroUseCase):**

```go
type HeroUseCase interface {
    CreateHero(ctx context.Context, p domain.CreateHeroParams) (string, error)
    UpdateHero(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error)
    DeleteHero(ctx context.Context, id string, hardDelete bool) error
    GetByID(ctx context.Context, id string) (*domain.Hero, error)
    ListHeroes(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error)
}
```

**Запросы (HeroQueryUseCase):**

```go
type HeroQueryUseCase interface {
    GetHeroDetail(ctx context.Context, id string) (*domain.HeroDetail, error)
    ListHeroPhotos(ctx context.Context, heroID string) ([]*domain.Photo, error)
}
```

**Преимущества CQRS:**
- Чтение и запись оптимизируются независимо
- `GetHeroDetail` загружает 6 связей параллельно через `errgroup`
- Команды валидируют бизнес-правила, запросы просто читают

### Валидация гибких дат

```go
func validateHeroDates(birth, death, serviceStart domain.FlexibleDate) error {
    b := exactAnchor(birth)
    d := exactAnchor(death)
    s := exactAnchor(serviceStart)

    if b != nil && d != nil && d.Before(*b) {
        return domain.ErrDeathBeforeBirth
    }
    if s != nil && b != nil && s.Before(*b) {
        return domain.ErrServiceBeforeBirth
    }
    if s != nil && d != nil && s.After(*d) {
        return domain.ErrServiceAfterDeath
    }
    return nil
}

func exactAnchor(d domain.FlexibleDate) *time.Time {
    if d.Precision != domain.PrecisionExact {
        return nil
    }
    return d.Anchor
}
```

**Важно:** Сравниваются только `EXACT` даты. Для неполных дат сравнение может дать ложный результат.

### Partial Update через field_mask

```go
func (uc *heroUseCase) UpdateHero(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
    if p.ID == nil || strings.TrimSpace(*p.ID) == "" {
        return nil, domain.ErrHeroIDRequired
    }

    normalizedMask, err := normalizeHeroFieldMask(p.FieldMask)
    if err != nil {
        return nil, err
    }
    p.FieldMask = normalizedMask

    mask := make(map[string]bool, len(p.FieldMask))
    for _, f := range p.FieldMask {
        mask[f] = true
    }

    // Защита от очистки обязательных полей
    if mask["first_name"] && isNilOrBlank(p.FirstName) {
        return nil, fmt.Errorf("first_name: %w", domain.ErrRequiredFieldEmpty)
    }

    // ... остальная валидация ...

    return uc.repo.Update(ctx, p)
}
```

**Нормализация field_mask:**

```go
func normalizeHeroFieldMask(mask []string) ([]string, error) {
    aliases := map[string]string{
        "birth_date_info":         "birth_date",
        "death_date_info":         "death_date",
        "service_start_date_info": "service_start_date",
    }
    // ... проверка дубликатов ...
}
```

### ThumbnailWorker

**Назначение:** Фоновая генерация превью и конвертация оригиналов в WebP.

```go
type ThumbnailWorker struct {
    repo         repository.PhotoRepository
    mediaStorage domain.MediaStorage
    logger       *slog.Logger
    config       config.ThumbnailsConfig
    queue        chan ThumbnailTask
}

type ThumbnailTask struct {
    PhotoID string
    HeroID  string
    URL     string
    FaceBox *domain.FaceBox
}
```

**Пайплайн обработки:**

```
Enqueue(task)
    │
    ▼
Background goroutine (N=workers):
1. Download оригинал из MinIO
2. Decode image (image.Decode + blank imports)
3. Crop: face_box → центральный 4:5 (fallback)
4. Resize 480×600 (Fill, Lanczos)
5. Encode WebP (lossless) → thumbnails/{photo_id}.webp
6. Конвертация оригинала в WebP (если не WebP) → photos/{photo_id}.webp
7. UPDATE photos SET url, thumbnail_url
```

**StartupBackfill:** При старте бэкенда находит все фото с пустым `thumbnail_url` и ставит их в очередь.

### OrphanCleanupWorker

**Назначение:** Удаление сиротских файлов из S3.

```go
type OrphanCleanupWorker struct {
    photoRepo    repository.PhotoRepository
    mediaStorage domain.MediaStorage
    logger       *slog.Logger
    interval     time.Duration
    gracePeriod  time.Duration
    dryRun       bool
}
```

**Алгоритм:**
1. Собирает все URL медиафайлов из БД
2. Сканирует объекты в MinIO через `ListObjects`
3. Удаляет файлы, не привязанные ни к одной записи и старше grace period

### AuthAuditWorker

**Назначение:** Асинхронная запись аудита аутентификации в БД.

```go
type AuthAuditWorker struct {
    repo    repository.AuthAuditRepository
    logger  *slog.Logger
    queue   chan domain.AuthAuditEntry
}
```

**Принцип:** Не блокирует обработку запросов. Если очередь переполнена, запись теряется (логируется WARN).

### AuthAlertService

**Назначение:** Отслеживание неудачных попыток аутентификации по IP.

```go
type AuthAlertService struct {
    mu        sync.Mutex
    attempts  map[string]*attemptWindow
    threshold int
    window    time.Duration
    logger    *slog.Logger
}
```

При превышении порога (по умолчанию 10 за 5 минут) логируется `ERROR` с `alert_type=auth_brute_force`.

---

## 🌐 Слой Delivery

**Расположение:** `internal/delivery/`

**Назначение:** Connect RPC серверы, мапперы domain ↔ proto, обработка ошибок.

Слой **Delivery** в Golang — это внешний слой приложения, который отвечает за получение запросов от клиентов (например, через HTTP, gRPC или CLI) и отправку ответов. Его главная задача — конвертировать данные из внешнего формата (JSON, Protobuf) во внутренние доменные структуры, координировать работу с бизнес-логикой (слой Use Case) и преобразовывать результат обратно в формат, понятный клиенту.

### Структура

```
internal/delivery/
├── v1/
│   ├── hero.go                 # HeroServer (публичный)
│   ├── hero_admin.go           # HeroAdminServer
│   ├── award.go                # AwardServer
│   ├── award_admin.go          # AwardAdminServer
│   ├── conflict.go             # ConflictServer
│   ├── conflict_admin.go       # ConflictAdminServer
│   ├── location.go             # LocationServer
│   ├── location_admin.go       # LocationAdminServer
│   ├── media.go                # MediaServer
│   ├── submission.go           # SubmissionServer
│   ├── submission_admin.go     # SubmissionAdminServer
│   ├── auth.go                 # AuthServer
│   ├── api_key_admin.go        # APIKeyAdminServer
│   ├── errors.go               # Локализация + маппинг ошибок
│   └── flexible_date.go        # Мапперы FlexibleDate
├── interceptor/
│   └── auth.go                 # AuthInterceptor
└── middleware/
    └── ratelimit.go            # RateLimitMiddleware
```

### Connect RPC серверы

**Пример:** `HeroServer`

```go
type HeroServer struct {
    heroUC         usecase.HeroUseCase
    heroQueryUC    usecase.HeroQueryUseCase
    heroSourceUC   usecase.HeroSourceUseCase
    heroRelationUC usecase.HeroRelationUseCase
    logger         *slog.Logger
}

func (s *HeroServer) GetHero(
    ctx context.Context,
    req *connect.Request[emhv1.GetHeroRequest],
) (*connect.Response[emhv1.GetHeroResponse], error) {
    detail, err := s.heroQueryUC.GetHeroDetail(ctx, req.Msg.Id)
    if err != nil {
        return nil, mapDomainError(ctx, s.logger, err, "get hero detail failed")
    }
    return connect.NewResponse(&emhv1.GetHeroResponse{
        Hero: mapHeroDetailToProto(detail),
    }), nil
}
```

### Мапперы domain ↔ proto

**Пример:** `mapHeroToSummary`

```go
func mapHeroToSummary(h *domain.Hero) *emhv1.HeroSummary {
    summary := &emhv1.HeroSummary{
        Id:            h.ID,
        FirstName:     h.FirstName,
        LastName:      h.LastName,
        MiddleName:    h.MiddleName,
        Rank:          h.Rank,
        ShortBio:      h.ShortBio,
        AwardNames:    h.AwardNames,
        Nickname:      h.Nickname,
        Unit:          h.Unit,
        ServiceBranch: h.ServiceBranch,
    }

    if h.MainPhotoURL != nil {
        summary.MainPhotoUrl = *h.MainPhotoURL
    }
    if h.MainThumbnailURL != nil {
        summary.MainThumbnailUrl = *h.MainThumbnailURL
    }
    if h.BirthDate.Anchor != nil {
        summary.BirthDate = timestamppb.New(*h.BirthDate.Anchor)
    }
    if h.DeathDate.Anchor != nil {
        summary.DeathDate = timestamppb.New(*h.DeathDate.Anchor)
    }

    // Гибкие даты
    summary.BirthDateInfo = mapFlexibleDateToProto(h.BirthDate)
    summary.DeathDateInfo = mapFlexibleDateToProto(h.DeathDate)

    return summary
}
```

### Обработка ошибок

**Файл:** `delivery/v1/errors.go`

```go
// errorMessages содержит русские сообщения для доменных ошибок.
var errorMessages = map[domain.ErrorCode]string{
    domain.ErrCodeNotFound:             "Запись не найдена",
    domain.ErrCodeDeathBeforeBirth:     "Дата гибели не может быть раньше даты рождения",
    domain.ErrCodeRequiredFieldEmpty:   "Обязательное поле не может быть пустым",
    // ...
}

// mapDomainError маппит доменную ошибку в Connect-ошибку с локализованным сообщением.
func mapDomainError(ctx context.Context, logger *slog.Logger, err error, action string) error {
    code := domain.CodeFromError(err)

    if code == domain.ErrCodeInternal {
        logger.ErrorContext(ctx, action, "error", err, "error_code", string(code))
        return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
    }

    logger.WarnContext(ctx, action, "error", err, "error_code", string(code))

    connectCode := connectCodeFromDomainCode(code)
    msg := localizedMessage(code)

    return connect.NewError(connectCode, errors.New(msg))
}
```

**Преимущества:**
- Типизированный маппинг через `errors.Is()`
- Локализованные сообщения для клиента
- Логирование с кодом ошибки для метрик
- Оригинальная ошибка не передаётся клиенту (безопасность)

### AuthInterceptor

**Файл:** `delivery/interceptor/auth.go`

```go
type AuthInterceptor struct {
    jwtSecret        string
    staticKeys       map[string]struct{}
    apiKeyAuth       APIKeyAuthenticator
    logger           *slog.Logger
    auditor          AuthAuditor
    alerter          AuthAlerter
    trustedProxyIPs  []net.IP
    trustedProxyNets []*net.IPNet
}
```

**Механизмы аутентификации:**
1. Статические ключи (аварийный доступ)
2. Управляемые API-ключи из БД
3. JWT (админы)

**Trusted Proxies:**
- Заголовки `X-Forwarded-For` / `X-Real-IP` используются только если запрос пришёл от доверенного прокси
- Поддержка точных IP и CIDR-диапазонов
- Иначе используется реальный адрес соединения из `req.Peer().Addr`

### RateLimitMiddleware

**Файл:** `delivery/middleware/ratelimit.go`

| Группа | Эндпоинты | Лимит | Окно |
|---|---|---|---|
| Auth | `AuthService/*` | 5 req | 1 мин |
| Submission | `SubmissionService/*` | 10 req | 1 час |
| Read | Все публичные Get/List | 120 req | 1 мин |
| Admin | Все `*AdminService/*` | Без лимита (JWT) | — |
| Media | `MediaService/*` | Без лимита (JWT) | — |

**Реализация:** In-memory `sync.Map` по IP + группа. Фоновая очистка записей.

---

## 🗄 Слой Repository

**Расположение:** `internal/repository/pg/`

**Назначение:** PostgreSQL реализации интерфейсов репозиториев.

Слой **Repository** в Go — это архитектурный паттерн, который создаёт прослойку между бизнес-логикой приложения и конкретным хранилищем данных (например, базой данных, файлом или внешним API). Его главная задача — абстрагировать детали работы с данными, чтобы сервисный слой (или слой вариантов использования) не зависел от деталей реализации. 

### Структура

```
internal/repository/pg/
├── hero_repository.go            # HeroRepository
├── photo_repository.go           # PhotoRepository
├── award_repository.go           # AwardRepository
├── hero_award_repository.go      # HeroAwardRepository
├── conflict_repository.go        # ConflictRepository
├── hero_conflict_repository.go   # HeroConflictRepository
├── location_repository.go        # LocationRepository
├── hero_location_repository.go   # HeroLocationRepository
├── hero_source_repository.go     # HeroSourceRepository
├── hero_relation_repository.go   # HeroRelationRepository
├── submission_repository.go      # SubmissionRepository
├── refresh_token_repository.go   # RefreshTokenRepository
├── api_key_repository.go         # APIKeyRepository
├── auth_audit_repository.go      # AuthAuditRepository
└── scannable.go                  # Интерфейс для pgx.Row / pgx.Rows
```

### Паттерны реализации

**SQL-билдер:** Squirrel с `PlaceholderFormat(squirrel.Dollar)`

**Пример:** `HeroRepository.Create`

```go
func (r *heroRepository) Create(ctx context.Context, p domain.CreateHeroParams) (string, error) {
    query := r.sb.Insert("heroes").
        Columns("first_name", "last_name", "middle_name", "short_bio", "full_bio", "rank",
            "birth_date", "birth_date_precision", "birth_date_display",
            "death_date", "death_date_precision", "death_date_display",
            "status", "nickname", "unit", "position",
            "service_branch", "cause_of_death",
            "service_start_date", "service_start_date_precision", "service_start_date_display",
            "memberships").
        Values(
            p.FirstName, p.LastName, p.MiddleName, p.ShortBio, p.FullBio, p.Rank,
            p.BirthDate.Anchor, int(p.BirthDate.Precision), nullableString(p.BirthDate.DisplayText),
            p.DeathDate.Anchor, int(p.DeathDate.Precision), nullableString(p.DeathDate.DisplayText),
            int(p.Status), p.Nickname, p.Unit, p.Position,
            p.ServiceBranch, p.CauseOfDeath,
            p.ServiceStartDate.Anchor, int(p.ServiceStartDate.Precision), nullableString(p.ServiceStartDate.DisplayText),
            squirrel.Expr("?::jsonb", marshalMemberships(p.Memberships)),
        ).
        Suffix("RETURNING id")

    sql, args, err := query.ToSql()
    if err != nil {
        return "", fmt.Errorf("build create query: %w", err)
    }

    var id string
    if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
        return "", fmt.Errorf("exec create: %w", err)
    }

    return id, nil
}
```

### Транзакции для batch-операций

**Пример:** `PhotoRepository.BatchAdd`

```go
func (r *photoRepository) BatchAdd(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error) {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    defer func() { _ = tx.Rollback(ctx) }()

    // Проверяем, есть ли уже главное фото у героя (внутри транзакции)
    checkMainQuery := r.sb.Select("COUNT(*)").
        From("photos").
        Where(squirrel.Eq{"hero_id": heroID, "is_main": true})
    // ...

    result := make([]*domain.Photo, 0, len(photos))
    mainAssigned := hasMainInDB

    for _, p := range photos {
        isMain := p.IsMain && !mainAssigned
        if isMain {
            mainAssigned = true
        }
        // ... INSERT ...
    }

    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("commit tx: %w", err)
    }

    return result, nil
}
```

### IDOR Protection

Все операции с фото проверяют принадлежность герою:

```go
func (r *photoRepository) Delete(ctx context.Context, heroID, photoID string) (string, error) {
    query := r.sb.Delete("photos").
        Where(squirrel.Eq{"id": photoID, "hero_id": heroID}).
        Suffix("RETURNING url")
    // ...
}
```

### Денормализация в ListHeroes

```sql
SELECT
    h.id, h.first_name, h.last_name, h.middle_name,
    h.short_bio, h.rank, h.birth_date, h.death_date, h.status,
    h.nickname, h.unit, h.service_branch,
    -- Главное фото (оригинал)
    COALESCE(
        (SELECT p.url FROM photos p WHERE p.hero_id = h.id AND p.is_main = true LIMIT 1),
        (SELECT p.url FROM photos p WHERE p.hero_id = h.id ORDER BY p.sort_order ASC, p.id ASC LIMIT 1)
    ) AS main_photo_url,
    -- Превью главного фото
    COALESCE(
        (SELECT p.thumbnail_url FROM photos p WHERE p.hero_id = h.id AND p.is_main = true AND p.thumbnail_url IS NOT NULL AND p.thumbnail_url != '' LIMIT 1),
        (SELECT p.thumbnail_url FROM photos p WHERE p.hero_id = h.id AND p.thumbnail_url IS NOT NULL AND p.thumbnail_url != '' ORDER BY p.sort_order ASC, p.id ASC LIMIT 1),
        (SELECT p.url FROM photos p WHERE p.hero_id = h.id AND p.is_main = true LIMIT 1),
        (SELECT p.url FROM photos p WHERE p.hero_id = h.id ORDER BY p.sort_order ASC, p.id ASC LIMIT 1)
    ) AS main_thumbnail_url,
    -- Награды
    (SELECT ARRAY_AGG(a.name ORDER BY a.sort_order) FROM hero_awards ha JOIN awards a ON ha.award_id = a.id WHERE ha.hero_id = h.id) AS award_names
FROM heroes h
```

---

## 💾 Слой Storage

**Расположение:** `internal/storage/s3/`

**Назначение:** MinIO S3-совместимое хранилище.

### Структура

```
internal/storage/s3/
└── media_storage.go    # MediaStorage
```

### Два клиента MinIO

```go
type MediaStorage struct {
    internal      *minio.Client // серверные операции (bucket, delete) — внутренний адрес
    public        *minio.Client // генерация presigned URL — внешний host, внутренний TCP
    bucket        string
    publicBaseURL string
}
```

**Проблема:** Presigned URL в SigV4 подписывается с учётом host. Внутренний адрес (`emh-minio:9000`) недоступен браузеру.

**Решение:** Public-клиент сконфигурирован с внешним endpoint (host для подписи), но TCP-соединения через кастомный `DialContext` перенаправляются на внутренний адрес.

### Retry Logic

Все операции обёрнуты в `withRetry` с exponential backoff:

```go
var defaultRetryConfig = retryConfig{
    maxRetries: 3,
    baseDelay:  100 * time.Millisecond,
    maxDelay:   2 * time.Second,
}
```

**Повторяются:** сетевые ошибки, 5xx, 429
**Не повторяются:** 4xx (кроме 429), отмена контекста

### HTTPS Rewrite

```go
func (s *MediaStorage) PresignedPutURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
    u, err := s.public.PresignedPutObject(ctx, s.bucket, key, expiry)
    if err != nil {
        return "", fmt.Errorf("presign put object: %w", err)
    }
    // AWS SigV4 не подписывает схему — безопасно меняем http → https
    u.Scheme = "https"
    return u.String(), nil
}
```

---

## 🔐 Слой Auth

**Расположение:** `internal/auth/`

**Назначение:** JWT, API-key, password hashing.

### Структура

```
internal/auth/
├── jwt.go        # GenerateToken, ValidateToken
├── password.go   # HashPassword, VerifyPassword
├── apikey.go     # GenerateAPIKey, HashAPIKeySecret, IsAPIKey
└── token.go      # GenerateRefreshToken, HashRefreshToken
```

### JWT

```go
type Claims struct {
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(username, role string, secret string, expiry time.Duration) (string, error)
func ValidateToken(tokenString string, secret string) (*Claims, error)
```

**Параметры:**
- Алгоритм: HS256
- Access token: 15 минут
- Claims: `{username, role}`

### Refresh Tokens

```go
func GenerateRefreshToken() (string, error)          // Random 256-bit (base64url)
func HashRefreshToken(token string) string           // SHA-256
```

**Параметры:**
- Длина: 256 бит
- Срок жизни: 7 дней
- Хранение: SHA-256 hash в БД
- Ротация: при Refresh старый токен отзывается

### API-ключи

```go
func GenerateAPIKey() (keyID, secret, fullKey string, err error)
func HashAPIKeySecret(secret string) (string, error)  // bcrypt
func IsAPIKey(token string) bool                      // Проверка префикса "emh_"
```

**Формат:** `emh_<key_id>_<secret>`

---

## ⚙️ Слой Config

**Расположение:** `internal/config/`

**Назначение:** Загрузка и валидация конфигурации.

### Структура

```
internal/config/
└── config.go    # Load, overrideFromEnv, Duration
```

### Кастомный тип Duration

```go
type Duration struct {
    time.Duration
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
    var s string
    if err := unmarshal(&s); err != nil {
        return err
    }
    parsed, err := time.ParseDuration(s)
    if err != nil {
        return err
    }
    d.Duration = parsed
    return nil
}
```

**Пример использования:**

```yaml
auth:
  jwt_expiry: 15m
  refresh_expiry: 168h
```

### Override из окружения

```go
func (c *Config) overrideFromEnv() {
    if v := os.Getenv("DATABASE_URL"); v != "" {
        c.DB.URL = v
    }
    if v := os.Getenv("AUTH_JWT_SECRET"); v != "" {
        c.Auth.JWTSecret = v
    }
    // ...
}
```

---

## 🎨 Ключевые паттерны

### Echo-паттерн

**Назначение:** `Add*` / `Remove*` / `SetMain*` возвращают запрос для подтверждения.

```proto
rpc AddHeroAward(AddHeroAwardRequest) returns (AddHeroAwardRequest);
rpc RemoveHeroAward(RemoveHeroAwardRequest) returns (RemoveHeroAwardRequest);
rpc SetMainHeroPhoto(SetMainHeroPhotoRequest) returns (SetMainHeroPhotoRequest);
```

**Преимущества:**
- Клиент получает подтверждение с теми же данными
- Не нужно создавать отдельный Response message
- Идемпотентность

### Response-паттерн

**Назначение:** `Create` / `Update` / `Delete` / `Batch*` возвращают отдельный Response.

```proto
rpc CreateHero(CreateHeroRequest) returns (CreateHeroResponse);
rpc UpdateHero(UpdateHeroRequest) returns (UpdateHeroResponse);
rpc DeleteHero(DeleteHeroRequest) returns (DeleteHeroResponse);
rpc BatchAddHeroPhotos(BatchAddHeroPhotosRequest) returns (BatchAddHeroPhotosResponse);
```

**Преимущества:**
- Явная структура ответа
- Можно добавить мета-поля (`added_count`, `success`)
- Не смешивается с запросом

### Partial Update через field_mask

```proto
message UpdateHeroRequest {
    string id = 1;
    string first_name = 2 [(buf.validate.field).ignore = IGNORE_IF_ZERO_VALUE];
    // ...
    repeated string field_mask = 11;
}
```

**Логика:**
- Обновляются только поля, указанные в `field_mask`
- Защита от очистки обязательных полей в usecase
- Нормализация имён (`birth_date_info` → `birth_date`)

### Best Effort

**Назначение:** Удаление S3 не блокирует удаление из БД.

```go
func (uc *photoUseCase) deleteFilesFromStorage(ctx context.Context, urls []string) {
    // Ошибки хранилища логируются, но не влияют на результат операции
    // Неудалённые файлы становятся сиротами и убираются orphan cleanup
}
```

---

## 🔄 Потоки данных

### Создание героя

```
Client
    │
    ▼
HeroAdminService.CreateHero (Connect RPC)
    │
    ├── ValidateInterceptor (buf.validate)
    │
    ▼
HeroAdminServer.CreateHero (delivery)
    │
    ├── resolveCreateFlexibleDate (маппинг дат)
    │
    ▼
HeroUseCase.CreateHero (usecase)
    │
    ├── validateHeroDates (бизнес-правила)
    │
    ▼
HeroRepository.Create (repository)
    │
    ├── INSERT INTO heroes ...
    │
    ▼
PostgreSQL
    │
    ▼
Response: {id: "uuid"}
```

### Загрузка фото

```
Client
    │
    ▼
MediaService.BatchGetUploadUrls (Connect RPC)
    │
    ▼
MediaUseCase.BatchGetUploadURLs (usecase)
    │
    ▼
MediaStorage.PresignedPutURL (storage)
    │
    ├── SigV4 подпись с внешним host
    ├── HTTPS rewrite
    │
    ▼
Response: [{upload_url, public_url, expires_at}]

Client загружает напрямую в MinIO через presigned URL

Client
    │
    ▼
HeroAdminService.BatchAddHeroPhotos (Connect RPC)
    │
    ▼
PhotoUseCase.BatchAdd (usecase)
    │
    ├── PhotoRepository.BatchAdd (транзакция)
    ├── ThumbnailWorker.Enqueue (для каждого фото)
    │
    ▼
Response: {added: [...], added_count: N}

ThumbnailWorker (фон):
    │
    ├── Download из MinIO
    ├── Crop/Resize
    ├── Encode WebP
    ├── Upload в MinIO
    ├── UPDATE photos SET url, thumbnail_url
    │
    ▼
Фото доступно на сайте
```

### Аутентификация

```
Client
    │
    ▼
AuthService.Login (Connect RPC)
    │
    ▼
AuthUseCase.Login (usecase)
    │
    ├── Проверка пароля (bcrypt)
    ├── GenerateToken (JWT, 15 мин)
    ├── GenerateRefreshToken (256-bit)
    ├── HashRefreshToken (SHA-256)
    ├── INSERT INTO refresh_tokens
    │
    ▼
Response: {access_token, refresh_token, expires_in}

Client
    │
    ▼
HeroAdminService.CreateHero (Connect RPC)
    │
    ├── AuthInterceptor (JWT validation)
    ├── ValidateInterceptor (buf.validate)
    │
    ▼
... обработка запроса ...
```

---

## ⚙️ Фоновые воркеры

### Обзор

| Воркер | Назначение | Запуск | Остановка |
|---|---|---|---|
| ThumbnailWorker | Генерация превью и конвертация в WebP | `runWorkerWithRecovery` | `workerCancel()` |
| OrphanCleanupWorker | Удаление сиротских файлов из S3 | `runWorkerWithRecovery` | `workerCancel()` |
| AuthAuditWorker | Асинхронная запись аудита аутентификации | `runWorkerWithRecovery` | `workerCancel()` + `Stop()` |

### runWorkerWithRecovery

```go
func runWorkerWithRecovery(
    ctx context.Context,
    logger *slog.Logger,
    name string,
    run func(context.Context),
    restartDelay time.Duration,
) {
    go func() {
        for {
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        logger.Error("worker panicked, will restart",
                            "worker", name,
                            "panic", r,
                            "stack", string(debug.Stack()),
                        )
                    }
                }()
                run(ctx)
            }()

            select {
            case <-ctx.Done():
                logger.Info("worker stopped", "worker", name)
                return
            case <-time.After(restartDelay):
                logger.Info("restarting worker", "worker", name)
            }
        }
    }()
}
```

**Поведение:**
- Panic → логирование + перезапуск через `restartDelay`
- Отмена контекста → остановка без перезапуска
- Graceful shutdown → `workerCancel()` останавливает все воркеры

---

## 🔐 Безопасность

### Аутентификация

| Механизм | Детали |
|---|---|
| Access token | JWT HS256, 15 минут, claims `{username, role}` |
| Refresh token | Random 256-bit (base64url), 7 дней, SHA-256 hash в БД, ротация |
| API-ключи | Формат `emh_<key_id>_<secret>`, bcrypt `secret_hash` |
| Статические ключи | Аварийный доступ, логируются как WARN |

### Авторизация

- Публичные сервисы без auth
- Админские сервисы: `authInterceptor` + `validateInterceptor`
- Role-based: `admin` для всех админских операций

### Rate Limiting

| Группа | Лимит | Окно |
|---|---|---|
| Auth | 5 req | 1 мин |
| Submission | 10 req | 1 час |
| Read | 120 req | 1 мин |
| Admin | Без лимита (JWT) | — |
| Media | Без лимита (JWT) | — |

### IDOR Protection

Все операции с фото проверяют принадлежность герою через `hero_id` в WHERE.

### Trusted Proxies

- Заголовки `X-Forwarded-For` / `X-Real-IP` используются только от доверенных прокси
- Поддержка CIDR-диапазонов
- Иначе используется реальный адрес соединения

### Аудит аутентификации

Все попытки входа записываются в `auth_audit_log`:
- `ip_address`, `mechanism`, `result`, `identity`, `procedure`, `failure_reason`, `user_agent`

### Алерты brute-force

При превышении порога (10 за 5 минут) логируется `ERROR` с `alert_type=auth_brute_force`.

---

## ⚡ Производительность

### PostgreSQL

- **UUIDv7** для PK — монотонность для курсорной пагинации
- **Индексы** на часто используемые фильтры
- **Пул соединений** (pgxpool): max 20, min 2
- **Денормализация** в ListHeroes (COALESCE для фото и наград)

### CQRS

- Чтение и запись оптимизируются независимо
- `GetHeroDetail` загружает 6 связей параллельно через `errgroup`

### Thumbnail Worker

- In-memory очередь (buffered channel)
- N воркеров (по умолчанию 2)
- Startup backfill при старте

### Orphan Cleanup

- Периодический (по умолчанию 24h)
- Grace period (по умолчанию 48h) — защита от удаления свежих файлов

### Retry Logic

- Exponential backoff для MinIO
- 3 повторные попытки, максимум 2s

### Rate Limiting

- In-memory `sync.Map` по IP + группа
- Фоновая очистка записей

---

## 📚 Дополнительные материалы

- [README.md](../backend/README.md) — обзор проекта
- [CONTRIBUTING.md](../CONTRIBUTING.md) — правила участия

---

<p align="center">
  <sub>Архитектура EMH Backend — чистая, тестируемая, масштабируемая</sub>
</p>
