# ТЗ: Бэкенд — Сид справочника наград

**Дата:** 2026-09-17  
**Статус:** Implemented (кроме воркера §9 — вне среза, см. §14)  
**Связано:** ADR-005 (секция «Сид справочника наград»)

> Документ приведён в соответствие с фактической реализацией.
> Ключевые отличия от первой редакции — в разделе «14. Отклонения от исходного ТЗ».

## 1. Область работ

Реализация пайплайна детерминированного сида справочника наград:
- Авто-сид при первом запуске (пустая БД)
- Инкрементальный импорт через one-shot контейнер (`seed-awards` subcommand)
- Расширение воркера для конвертации наград из админки (2 rendition)

## 2. Структура файлов

```
backend/
  seeds/awards/
    descriptions.json     # данные наград (без картинок) — Git LFS
    images/*.webp         # знаки 400×400 — Git LFS
    ribbons/*.webp        # ленты 144×48 — Git LFS
    seed.json             # собранный артефакт (base64) — Git LFS
  cmd/
    backend/
      main.go             # добавление subcommand `seed-awards`
      seed_awards.go      # обёртка над seeder
    seed-builder/
      main.go             # standalone: сборка seed.json из исходников
  internal/
    seeder/
      awards.go           # общая логика сида (используется bootstrap + subcommand)
      schema.go           # Go-структуры для десериализации seed.json
```

## 3. Прото-изменения

### 3.1. `proto/emh/v1/media.proto`

Добавить в `UploadType`:

```protobuf
enum UploadType {
  // ... существующие 0–3 ...
  
  // UPLOAD_TYPE_AWARD_RIBBON - изображение ленты награды (планка).
  UPLOAD_TYPE_AWARD_RIBBON = 4;
}
```

**Проверить:** после ADR-005 уже должен быть добавлен `allowedMIME`/`typePrefix` для `awards/ribbons` в `media_usecase.go`. Если нет — добавить.

### 3.2. `proto/emh/v1/award.proto`

Добавить в `Award`:

```protobuf
// slug — машиночитаемый идентификатор (опционально, для внешнего ссылочного).
string slug = 11;
```

> Примечание: для сида `slug` не обязателен — используем составной ключ `(name, type, jurisdiction)`. Поле добавляется как опциональное расширение.

## 4. Миграция 00028

**Путь:** `backend/migrations/00028_awards_seed_key.sql`

```sql
-- +goose Up
-- Составной уникальный ключ для идемпотентности сида и защиты от дублей админа
CREATE UNIQUE INDEX IF NOT EXISTS idx_awards_seed_key
    ON awards (name, type, jurisdiction)
    WHERE name IS NOT NULL AND name != '';

-- +goose Down
DROP INDEX IF EXISTS idx_awards_seed_key;
```

## 5. Модуль `internal/seeder`

### 5.1. `schema.go` — структуры

```go
package seeder

// SeedAwardJSON — одна запись в seed.json
type SeedAwardJSON struct {
    Name           string  `json:"name"`
    Description    string  `json:"description"`
    Type           string  `json:"type"`            // "AWARD_TYPE_MEDAL" | "_ORDER" | "_BADGE"
    Jurisdiction   string  `json:"jurisdiction"`    // "AWARD_JURISDICTION_RF" | "_USSR" | "_OTHER"
    IsJubilee      bool    `json:"is_jubilee"`
    WornWithoutBar bool    `json:"worn_without_bar"`
    SortOrder      int32   `json:"sort_order"`
    Image          *Image  `json:"image,omitempty"`
    Ribbon         *Image  `json:"ribbon,omitempty"`
}

type Image struct {
    ContentType string `json:"content_type"` // должен быть "image/webp"
    Data        string `json:"data"`         // base64 без префикса data:
}

type SeedFile struct {
    Awards []SeedAwardJSON `json:"awards"`
}
```

### 5.2. `awards.go` — основная логика

```go
package seeder

// SeedAwardsResult — результат импорта
type SeedAwardsResult struct {
    Inserted int
    Skipped  int
    Failed   int
}

// SeedAwards импортирует награды из seed.json в БД + MinIO.
// Идемпотентно: существующие записи пропускаются по (name, type, jurisdiction).
// Ошибки обработки отдельных записей не прерывают сид (log + skip).
func SeedAwards(ctx context.Context, seedPath string,
    repo AwardRepository, minio MinIOUploader,
) (*SeedAwardsResult, error)

// HasAnyAwards — быстрая проверка: есть ли записи в awards
func HasAnyAwards(ctx context.Context, repo AwardRepository) (bool, error)
```

### 5.3. Алгоритм

```go
func SeedAwards(...) (*SeedAwardsResult, error) {
    seed, err := loadSeedFile(seedPath)
    if err != nil {
        return nil, err  // битый JSON = ошибка (не skip)
    }
    
    // 1. Проверка дубликатов внутри самого файла
    if err := validateNoInternalDuplicates(seed.Awards); err != nil {
        return nil, err
    }
    
    for _, a := range seed.Awards {
        // 2. Дешёвая проверка существования (до загрузки в MinIO)
        exists, _ := repo.ExistsByKey(ctx, a.Name, a.Type, a.Jurisdiction)
        if exists {
            result.Skipped++
            continue
        }
        
        // 3. Валидация картинок
        if err := validateImage(a.Image); err != nil {
            log.Error("invalid image", "name", a.Name, "err", err)
            result.Failed++
            continue
        }
        
        // 4. Загрузка в MinIO (прямая, без presigned)
        hash := sha256(name+type+jurisdiction)[:16]
        imageURL, err := minio.UploadBytes(ctx, "awards/images/"+hash+".webp", imageBytes, "image/webp")
        if err != nil {
            log.Error("image upload failed", "name", a.Name, "err", err)
            result.Failed++
            continue
        }
        ribbonURL, err := minio.UploadBytes(ctx, "awards/ribbons/"+hash+".webp", ribbonBytes, "image/webp")
        if err != nil {
            log.Error("ribbon upload failed", "name", a.Name, "err", err)
            result.Failed++
            continue
        }
        
        // 5. INSERT ON CONFLICT DO NOTHING (защита от гонки с админом)
        inserted, err := repo.InsertIfAbsent(ctx, domain.Award{
            ID: uuid.New(),
            Name: a.Name,
            Description: a.Description,
            Type: mapType(a.Type),
            Jurisdiction: mapJurisdiction(a.Jurisdiction),
            IsJubilee: a.IsJubilee,
            WornWithoutBar: a.WornWithoutBar,
            SortOrder: a.SortOrder,
            ImageURL: imageURL,
            RibbonImageURL: ribbonURL,
        })
        if err != nil {
            log.Error("insert failed", "name", a.Name, "err", err)
            result.Failed++
            continue
        }
        if !inserted {
            log.Warn("race: created by admin during seed", "name", a.Name)
            result.Skipped++
        } else {
            result.Inserted++
        }
    }
    return &result, nil
}
```

### 5.4. Репозиторий (новые методы)

```go
// AwardRepository — расширение интерфейса
type AwardRepository interface {
    // ... существующие методы ...
    
    // ExistsByKey — дешёвая проверка существования по составному ключу
    ExistsByKey(ctx context.Context, name, awardType, jurisdiction string) (bool, error)
    
    // InsertIfAbsent — INSERT ON CONFLICT DO NOTHING, возвращает true если вставлен
    InsertIfAbsent(ctx context.Context, award domain.Award) (bool, error)
}
```

**Реализация:**
```go
func (r *awardRepository) ExistsByKey(ctx, name, t, j) (bool, error) {
    var exists bool
    err := r.db.QueryRow(ctx,
        `SELECT EXISTS(SELECT 1 FROM awards WHERE name=$1 AND type=$2 AND jurisdiction=$3)`,
        name, t, j,
    ).Scan(&exists)
    return exists, err
}

func (r *awardRepository) InsertIfAbsent(ctx, award) (bool, error) {
    tag, err := r.db.Exec(ctx, `
        INSERT INTO awards (id, name, description, type, jurisdiction,
                            is_jubilee, worn_without_bar, sort_order,
                            image_url, ribbon_image_url)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        ON CONFLICT (name, type, jurisdiction) DO NOTHING
    `, ...)
    if err != nil {
        return false, err
    }
    return tag.RowsAffected() > 0, nil
}
```

## 6. Subcommand `seed-awards`

**Файл:** `backend/cmd/backend/seed_awards.go`

```go
// seed-awards <path-to-seed.json>
// one-shot: импортирует сид и выходит. Используется для инкремента на VPS.
```

### Логика
1. Парсит аргумент — путь к `seed.json`
2. Подключается к БД (DSN из env `DB_DSN`) и MinIO (из env)
3. Вызывает `seeder.SeedAwards(ctx, path, repo, minio)`
4. Логирует итог: `seed complete: inserted=N skipped=M failed=K`
5. `os.Exit(1)` если `failed > 0`, иначе `0`

### Обёртка в main.go

Добавить в парсер subcommands (как уже сделано для `migrate`):

```go
case "seed-awards":
    if len(os.Args) < 3 {
        fmt.Fprintln(os.Stderr, "usage: emh-backend seed-awards <seed.json>")
        os.Exit(1)
    }
    runSeedAwards(ctx, os.Args[2])
```

## 7. Авто-сид при старте

**Где:** bootstrap HTTP-сервера в `cmd/backend/main.go`, **до** `http.ListenAndServe`.

```go
func bootstrap(ctx context.Context, db *pgxpool.Pool, minio *minio.Client) {
    // 1. Прогоним миграции (уже есть)
    runMigrations(ctx, db)
    
    // 2. Авто-сид: только если БД пуста и файл существует
    has, err := seeder.HasAnyAwards(ctx, awardRepo)
    if err != nil {
        log.Error("awards check failed", "err", err)
        return  // не блокируем старт
    }
    if has {
        return  // справочник уже наполнен
    }
    
    seedPath := os.Getenv("AWARDS_SEED_PATH")
    if seedPath == "" {
        seedPath = "/app/seeds/seed.json"
    }
    if _, err := os.Stat(seedPath); os.IsNotExist(err) {
        log.Warn("seed file not found, skipping", "path", seedPath)
        return
    }
    
    log.Info("Awards table empty, running seed...", "path", seedPath)
    result, err := seeder.SeedAwards(ctx, seedPath, awardRepo, minio)
    if err != nil {
        log.Error("seed failed", "err", err)
        return  // не блокируем старт
    }
    log.Info("seed complete",
        "inserted", result.Inserted,
        "skipped", result.Skipped,
        "failed", result.Failed)
}
```

## 8. Standalone `cmd/seed-builder`

**Файл:** `backend/cmd/seed-builder/main.go`  
**Назначение:** сборка `seed.json` из исходников (для разработчика, не для деплоя).

### Логика
```go
func main() {
    dir := os.Args[1]  // например, ./seeds/awards
    
    descriptions := loadJSON(path.Join(dir, "descriptions.json"))
    for i := range descriptions.Awards {
        a := &descriptions.Awards[i]
        // читаем картинку знака
        if a.ImagePath != "" {
            a.Image = loadImageBase64(path.Join(dir, a.ImagePath))
            a.ImagePath = ""  // очищаем путь
        }
        // читаем картинку ленты
        if a.RibbonPath != "" {
            a.Ribbon = loadImageBase64(path.Join(dir, a.RibbonPath))
            a.RibbonPath = ""
        }
    }
    
    // валидация
    validateNoInternalDuplicates(descriptions.Awards)
    for _, a := range descriptions.Awards {
        validateImage(a.Image)
        validateImage(a.Ribbon)
    }
    
    // записываем seed.json
    out, _ := json.MarshalIndent(descriptions, "", "  ")
    os.WriteFile(path.Join(dir, "seed.json"), out, 0644)
    log.Printf("seed.json written: %d awards\n", len(descriptions.Awards))
}
```

**Валидация картинок:**
- MIME type: `image/webp` (проверка через http.DetectContentType на первых байтах)
- Размер: ≤ 100 КБ после декодирования
- Dimensions: знак = 400×400, лента = 144×48 (с допуском ±10%)
- Background: преимущественно белый (проверка пикселей по углам)

## 9. Расширение воркера

**Файл:** `backend/internal/worker/image_processor.go`

Добавить обработку для двух новых `UploadType`:

```go
case domain.UploadTypeAwardImage:
    return p.processAwardImage(ctx, obj)  // 400×400, contain, белый фон, q90
    
case domain.UploadTypeAwardRibbon:
    return p.processAwardRibbon(ctx, obj)  // 144×48, contain, белый фон, q90
```

**Паттерн обработки** (идентичен фото героев):
```go
func (p *processor) processAwardImage(ctx, obj) {
    img := loadImage(obj)
    canvas := image.NewRGBA(image.Rect(0, 0, 400, 400))
    fillWhite(canvas)
    drawContainFit(canvas, img, 400, 400)
    webp := encodeWebP(canvas, quality=90)
    upload(obj.Key, webp)
    // обновить award.image_url на финальный путь
}
```

**Важно:** воркер не трогает картинки из сида (они уже в финальном виде в MinIO). Обработка только для админских загрузок.

## 10. Серверная защита от дублей (AwardAdminService)

В `CreateAward` и `UpdateAward` (если меняется `name/type/jurisdiction`) ловить pg-ошибку `23505` (unique_violation) и маппить в `connect.CodeAlreadyExists`:

```go
func (s *awardAdminService) CreateAward(ctx, req) (*CreateAwardResponse, error) {
    award, err := s.usecase.Create(ctx, mapToDomain(req))
    if err != nil {
        if errors.Is(err, domain.ErrAwardDuplicate) {
            return nil, connect.NewError(connect.CodeAlreadyExists,
                fmt.Errorf(
                    "Награда «%s» (тип: %s, юрисдикция: %s) уже существует в справочнике",
                    req.Name, typeLabel(req.Type), jurisdictionLabel(req.Jurisdiction),
                ))
        }
        return nil, err
    }
    return &CreateAwardResponse{Id: award.ID}, nil
}
```

**Репозиторий** должен маппить pg-error в `domain.ErrAwardDuplicate`.

## 11. Тестирование

- [ ] Unit: `seeder.validateNoInternalDuplicates` — находит дубликаты ключа
- [ ] Unit: `seeder.validateImage` — валидирует MIME, размер, dimensions
- [ ] Unit: `awardRepository.InsertIfAbsent` возвращает `false` при конфликте
- [ ] Integration: round-trip сида (seed → exists → seed again → skipped)
- [ ] Integration: `AwardAdminService.CreateAward` возвращает `AlreadyExists` при конфликте
- [ ] E2E: `seed-awards` subcommand на тестовой БД

## 12. Justfile (в корне)

```justfile
# === Сид наград ===

# Сборка seed.json из исходников (локально, для разработчика)
build-seed:
    cd backend && go run ./cmd/seed-builder ./seeds/awards

# Инкремент сида на VPS (one-shot контейнер)
seed-awards-vps:
    podman run --rm \
        -v /opt/emh/seeds:/app/seeds:ro \
        --env-file /opt/emh/.env \
        codeberg.org/thr0tt1e/emh-backend:latest \
        seed-awards /app/seeds/seed.json
```

## 13. Критерии приёмки

- [x] Миграция 00028: `idx_awards_seed_key` создан, `up` и `down` работают
- [x] `seeder.SeedAwards`: вставляет новые, пропускает существующие, логирует ошибки
- [x] Авто-сид при пустой БД запускается автоматически
- [x] Subcommand `seed-awards` работает в one-shot режиме
- [x] `seed-builder` собирает валидный `seed.json` из исходников
- [ ] Воркер обрабатывает `AWARD_IMAGE` и `AWARD_RIBBON` с правильными rendition — **вне среза** (§14.11)
- [x] `CreateAward` возвращает `AlreadyExists` при конфликте
- [ ] `go build`, `go vet`, `just test` — все зелёные

## 14. Отклонения от исходного ТЗ (реализация 2026-11-10)

| # | ТЗ | Факт | Причина |
|---|----|------|---------|
| 1 | §2/§6 `cmd/backend/` | `cmd/api/` (бинарник `emh`) | Так называется пакет в репозитории |
| 2 | §6 «обёртка как для `migrate`» | Диспетчер субкоманд написан с нуля: `isSeedAwardsCommand()` в `main.go` (шаг 2.5), `runSeedAwards` в `cmd/api/seed_awards.go` | `migrate` — не субкоманда, а отдельный oneshot-контейнер `emh-migrate.container`; субкоманд в бинарнике не было |
| 3 | §7 «runMigrations (уже есть)» в bootstrap | Миграции вне процесса бэкенда; bootstrap-сид идёт после `emh-migrate.service` (порядок задан в quadlet) | Фактическая архитектура деплоя |
| 4 | §4 partial-индекс `WHERE name != ''` + `ON CONFLICT (name,type,jurisdiction)` | Обычный `UNIQUE INDEX idx_awards_seed_key` (миграция 00028) | Баг ТЗ: таргетный ON CONFLICT не матчится с partial-индексом без повтора предиката — в рантайме ошибка «no unique constraint matching». Пустое имя и так отсекается валидацией |
| 5 | §5.2 сигнатура `SeedAwards(ctx, path, repo, minio)` | `SeedAwards(ctx, path, repo, storage domain.MediaStorage, logger)` | `MinIOUploader` не заводим — переиспользуем существующий `domain.MediaStorage` (Upload/PublicURL); логгер — явный параметр (slog-конвенция) |
| 6 | §5.1 `ExistsByKey(name, string, string)` | `ExistsByKey(ctx, name, domain.AwardType, domain.AwardJurisdiction)`; `InsertIfAbsent` принимает `domain.CreateAwardParams` | В БД `type/jurisdiction` — SMALLINT (миграция 00027); строки-имена в репозиторий не протекают, маппинг enum — только в сидере |
| 7 | §5.4 хеш «sha256(name+type+jurisdiction)» | `sha256(name␀type␀jurisdiction)[:16]` — NUL-разделитель, proto-имена строк | Прямая конкатенация коллизна («A|B»+«C» vs «A»+«B|C») |
| 8 | §5.1 base64 «без префикса data:» как объект `{content_type, data}` против ADR-005 (data-URI) | Оставлен вариант ТЗ `{content_type, data}` | Парсить data-URI лишняя работа; `content_type` — дешёвая проверка просачивания не-webp (решение Q12) |
| 9 | §8 валидация builder: ≤100 КБ, «преимущественно белый фон» (углы) | Не берём: только magic-webp + **точные** размеры 400×400/144×48 + канон enum + внутренние дубли + непустой список | Допуск ±10% размывает контракт с фиксированными width/height фронта (CLS); «белые углы» субъективно и хрупко — качество картинок ответственность подготовщика (ADR-005 «риски») |
| 10 | §3.2 `slug` в `award.proto` | Не добавлен | Мёртвое поле без потребителя (принцип ADR-005 про `devices`) |
| 11 | §9 воркер `internal/worker/image_processor.go` | Не реализован | Вне среза сида (решение Q1): нужен только ручным загрузкам админки |
| 12 | §12 `seed-awards-vps`: `/opt/emh/seeds`, образ `codeberg.org/...:latest` | `{{EMH_SEEDS_DIR | $HOME/emh/seeds}}`, образ `docker.io/thr0tt1e/emh-backend:latest`, `--network emh-net` | Унификация с quadlet-путями (`%h/emh/seeds`) и фактическим registry образа |
| 13 | §5.3 загрузка ленты отдельно от знака, частичный инсерт допустим | При сбое MinIO/битой картинке **вся запись не инсертится** (Failed++) | Онконфликт-ключ сделал бы «без Ribbon» невидимым для последующих сидов |
| 14 | §11 unit `seeder.validateImage` (MIME/размер/разрешение) | В сидере — только base64 + RIFF/WEBP magic; полная валидация в builder | Строгость на сборке артефакта (решение Q10) |

### Что сделано в обход ТЗ
- `seed-builder` валидирует непустоту `descriptions.json`, расширение `.webp` и strict enum по generated maps (`emhv1.AwardType_value` / `AwardJurisdiction_value`) — единый источник правды, протустойчив к переименованиям значений enum.
- `MediaStorage.Upload` теперь сообщает minio-go известный размер для seekable-reader (одна PUT вместо multipart-UPLOADS для мелких объектов).
- `AwardRepository.Update` тоже маппит 23505 в `ErrAwardDuplicate` (переименование может попасть в ключ сида) — ТЗ обещал `UpdateAward`, но не уточнял место.
- E2E-проверка сида выполнена локально: `seed-awards` дважды (inserted=5 → skipped=5) против тестовой БД + fake-S3, авто-сид проверен холодным стартом бэкенда (пустая БД → inserted=5).
