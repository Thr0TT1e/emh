# CONTRIBUTING.md — Правила участия в проекте «Вечная память героям»

<p align="center">
  <h1 align="center">Вклад в проект EMH</h1>
  <p align="center">
    <em>Руководство для разработчиков, историков и всех, кто хочет помочь сохранить память о героях</em>
  </p>
</p>

Спасибо, что рассматриваете возможность участия! Этот проект объединяет разработчиков, историков, родственников героев и неравнодушных людей. Каждый вклад важен — от исправления опечатки до реализации новой функциональности.

---

## 📋 Содержание

1. [Кодекс поведения](#-кодекс-поведения)
2. [Типы участия](#-типы-участия)
3. [Настройка окружения](#-настройка-окружения)
4. [Архитектурные принципы](#-архитектурные-принципы)
5. [Стиль кода](#-стиль-кода)
6. [Работа с Proto-контрактами](#-работа-с-proto-контрактами)
7. [Миграции БД](#-миграции-бд)
8. [Тестирование](#-тестирование)
9. [Git-конвенции](#-git-конвенции)
10. [Процесс Pull Request](#-процесс-pull-request)
11. [Специфика мемориального проекта](#-специфика-мемориального-проекта)
12. [Связь с командой](#-связь-с-командой)

---

## 🤝 Кодекс поведения

Проект посвящён памяти павших героев. Мы ожидаем от всех участников:

- **Уважения** к памяти погибших и их родственникам
- **Достоверности** — никаких непроверенных фактов в контенте
- **Конструктивности** в обсуждениях и ревью
- **Конфиденциальности** при работе с персональными данными (личные фото, документы родственников)

Недопустимо:
- Политизирование памяти героев
- Размещение недостоверной или провокационной информации
- Использование проекта в коммерческих или политических целях

---

## 🎯 Типы участия

### Разработка (Backend / Frontend / DevOps)
- Исправление багов
- Реализация новых фич
- Оптимизация производительности
- Написание тестов
- Улучшение документации

### Контент и верификация
- Добавление биографий героев (через админ-панель или заявки)
- Проверка фактов из открытых источников
- Оцифровка архивных материалов
- Перевод материалов на другие языки

### Дизайн и UX
- Улучшение интерфейса
- Адаптация под мобильные устройства
- Доступность (a11y) для людей с ограниченными возможностями

### Инфраструктура
- Помощь с деплоем и мониторингом
- Настройка CI/CD
- Оптимизация работы с S3 и PostgreSQL

---

## 🛠 Настройка окружения

### Требования

| Инструмент | Минимальная версия | Назначение |
|---|---|---|
| Go | 1.26.5 | Язык бэкенда |
| Podman или Docker | latest | Контейнеризация |
| podman-compose | latest | Локальная инфраструктура |
| Just | latest | Task-runner |
| easyp | latest | Proto-линтинг и генерация |
| goose | latest | SQL-миграции |
| Node.js | 20+ | Frontend (Nuxt 4) |

### Быстрый старт (Backend)

```bash
# 1. Клонирование
git clone https://codeberg.org/Thr0TT1e/emh.git
cd emh/backend

# 2. Создание .env
vim infra/podman/.env
# Отредактируйте секреты: DATABASE_URL, MINIO_ROOT_USER/PASSWORD, AUTH_JWT_SECRET

# 3. Запуск инфраструктуры (PostgreSQL + MinIO)
podman-compose -f infra/podman/hmr/podman-compose-hmr.yml up

# 4. Применение миграций
just up

# 5. Генерация proto-кода
just gen-sdk

# 6. Запуск сервера
go run ./cmd/api
```

### Быстрый старт (Frontend)

```bash
cd ../frontend
pnpm i
pnpm dev
```

Frontend будет доступен на `http://localhost:3000`, backend на `http://localhost:3480`.

---

## 🏗 Архитектурные принципы

### Clean Architecture

```
domain ← usecase ← delivery
   ↑         ↑
repository  storage
```

**Правила зависимостей:**
- `domain` не зависит ни от кого (чистый Go)
- `usecase` зависит только от `domain` и интерфейсов `repository`
- `delivery` зависит от `usecase` и proto-контрактов
- `repository` и `storage` — инфраструктурные реализации интерфейсов

### CQRS для героя

- `HeroUseCase` — команды (Create/Update/Delete)
- `HeroQueryUseCase` — чтение агрегата `HeroDetail` с параллельной загрузкой 6 связей через `errgroup`

### Именование слоёв

| Слой | Паттерн |
|---|---|
| domain | `Hero`, `FlexibleDate`, `PublicationStatus` |
| repository | `HeroRepository` (интерфейс), `heroRepository` (реализация) |
| usecase | `HeroUseCase` (интерфейс), `heroUseCase` (реализация) |
| delivery | `HeroServer`, `HeroAdminServer` |

---

## 💻 Стиль кода

### Go

- Стандартный `gofmt` + `goimports`
- Ошибки: **всегда** оборачивать через `fmt.Errorf("...: %w", err)`
- Контекст: передавать первым параметром `(ctx context.Context, ...)`

### Типизированные ошибки

**НЕ** использовать `strings.Contains(err.Error(), "...")` для определения типа ошибки.

**ИСПОЛЬЗОВАТЬ** sentinel errors из `domain/errors.go`:

```go
// domain/errors.go
var (
    ErrNotFound         = errors.New("not found")
    ErrDeathBeforeBirth = errors.New("death date cannot be before birth date")
    // ...
)

// usecase
return fmt.Errorf("hero %w", domain.ErrNotFound)

// delivery
if errors.Is(err, domain.ErrNotFound) {
    return connect.NewError(connect.CodeNotFound, err)
}
```

### Обработка ошибок в delivery

Используйте `mapDomainError` из `delivery/v1/errors.go`:

```go
if err := s.heroUC.DeleteHero(ctx, id, hardDelete); err != nil {
    return nil, mapDomainError(ctx, s.logger, err, "delete hero failed")
}
```

Это обеспечивает:
- Типизированный маппинг на Connect-коды
- Локализованные русские сообщения для клиента
- Логирование с кодом ошибки для метрик

### Partial Update через field_mask

Все Update-методы используют `field_mask` для явного указания обновляемых полей:

```go
type UpdateHeroParams struct {
    ID        *string
    FirstName *string  // nil = не менять
    // ...
    FieldMask []string
}
```

Защита обязательных полей реализуется в usecase:

```go
if mask["first_name"] && isNilOrBlank(p.FirstName) {
    return nil, domain.ErrRequiredFieldEmpty
}
```

---

## 📝 Работа с Proto-контрактами

### Структура

```
proto/emh/v1/
├── enums_emh.proto      # общие перечисления
├── common.proto         # Pagination, AuditInfo, FlexibleDate
├── hero.proto           # публичный HeroService
├── hero_admin.proto     # админский HeroAdminService
├── award.proto          # справочник наград
└── ...
```

### Правила

1. **Все поля** должны иметь leading-комментарий:
   ```proto
   // first_name - имя героя (обязательно).
   string first_name = 1 [(buf.validate.field).string.min_len = 1];
   ```

2. **Валидация** через `buf.validate`:
   ```proto
   string id = 1 [(buf.validate.field).string.uuid = true];
   int32 page_size = 2 [(buf.validate.field).int32.lte = 100];
   ```

3. **Гибкие даты** используйте `FlexibleDate` вместо `Timestamp`:
   ```proto
   FlexibleDate birth_date_info = 15;
   ```

4. **Обратная совместимость**: новые поля добавляйте в конец, старые помечайте `DEPRECATED`.

### Генерация кода

```bash
just gen-sdk
# эквивалентно:
easyp lint -r .
easyp generate
```

Сгенерированный код попадает в `internal/gen/emh/v1/`. **Никогда не редактируйте его вручную.**

---

## 🗄 Миграции БД

### Создание миграции

```bash
cd backend/migrations
goose create add_new_feature sql
```

### Правила

1. **Именование**: `NNNNN_описание.sql` (например, `00017_add_hero_tags.sql`)
2. **Обратимость**: всегда пишите `-- +goose Down` блок
3. **UUIDv7** для PK (монотонность для курсорной пагинации)
4. **Индексы** для часто используемых фильтров
5. **Комментарии** на таблицы и колонки через `COMMENT ON`

### Пример миграции

```sql
-- +goose Up

CREATE TABLE hero_tags (
    hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    tag     TEXT NOT NULL,
    PRIMARY KEY (hero_id, tag)
);

CREATE INDEX idx_hero_tags_tag ON hero_tags (tag);

COMMENT ON TABLE hero_tags IS 'Теги для классификации героев';

-- +goose Down

DROP TABLE IF EXISTS hero_tags;
```

### Применение

```bash
just up       # применить все неприменённые
just down     # откатить последнюю
just status   # текущая версия
```

---

## 🧪 Тестирование (в разработке)

### Уровни тестов

| Уровень | Расположение | Запуск |
|---|---|---|
| Unit | `internal/domain/*_test.go` | `just test` |
| UseCase | `internal/usecase/*_test.go` с фейками | `just test` |
| Integration | `internal/repository/pg/*_test.go` | `just test-int` |
| API | `internal/delivery/v1/*_test.go` | `just test-int` |

### Запуск

```bash
just test          # unit-тесты (-short)
just test-int      # интеграционные (требуют PostgreSQL + MinIO)
just test-all      # все тесты
just cover         # coverage report
```

### Правила написания тестов

1. **Table-driven tests** для валидации и маппинга
2. **Фикстуры** в `internal/test/fixtures.go`
3. **Очистка БД** между тестами через `TRUNCATE ... CASCADE`
4. **Не тестируйте** сгенерированный proto-код
5. **Проверяйте коды ошибок** через `connect.CodeOf(err)`, а не текст

### Пример unit-теста

```go
func TestFlexibleDate_IsValid(t *testing.T) {
    tests := []struct {
        name    string
        date    domain.FlexibleDate
        wantErr error
    }{
        {
            name: "exact with anchor",
            date: domain.FlexibleDate{
                Precision: domain.PrecisionExact,
                Anchor:    ptrTime(time.Date(1994, 2, 15, 0, 0, 0, 0, time.UTC)),
            },
            wantErr: nil,
        },
        {
            name: "exact without anchor",
            date: domain.FlexibleDate{
                Precision: domain.PrecisionExact,
            },
            wantErr: domain.ErrExactDateRequiresAnchor,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.date.IsValid()
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Покрытие

Минимальные цели:

| Слой | Покрытие |
|---|---|
| domain | 85%+ |
| usecase | 70%+ |
| repository | 60%+ |
| delivery | 50%+ |

---

## 📦 Git-конвенции

### Ветвление

```
master          ← production-ready код
├── develop     ← интеграционная ветка
├── feature/    ← новые фичи
├── fix/        ← исправления багов
├── refactor/   ← рефакторинг без изменения поведения
└── docs/       ← документация
```

### Формат коммитов

Используем **Conventional Commits**:

```
<type>(<scope>): <subject>

[body]

[footer]
```

**Типы:**

| Тип | Назначение |
|---|---|
| `feat` | Новая функциональность |
| `fix` | Исправление бага |
| `refactor` | Рефакторинг без изменения поведения |
| `docs` | Документация |
| `test` | Тесты |
| `chore` | Инфраструктура, зависимости |
| `perf` | Оптимизация производительности |
| `security` | Безопасность |

**Области (scope):**

- `hero`, `photo`, `auth`, `media`, `db`, `proto`, `worker`, `api`

**Примеры:**

```bash
feat(hero): add flexible dates support

- Add FlexibleDate message to common.proto
- Implement DatePrecision enum with 8 values
- Add birth_date_info, death_date_info, service_start_date_info fields
- Preserve backward compatibility with legacy string fields

Closes #42
```

```bash
fix(photo): prevent IDOR in Delete operation

Add hero_id check to WHERE clause in photo_repository.Delete.
Previously, attacker could delete any photo by knowing its UUID.

Fixes: P0-1
```

```bash
refactor(delivery): replace strings.Contains with typed errors

Use errors.Is() with sentinel errors from domain/errors.go.
Add localized Russian messages in delivery/v1/errors.go.
```

### Pull Request

1. **Один PR — одна задача**
2. **Описание**: что сделано, почему, как проверить
3. **Скриншоты** для UI-изменений
4. **Ссылка на issue** (если есть)
5. **Self-review** перед отправкой

---

## 🔍 Процесс Pull Request

### Чеклист перед отправкой

- [ ] Код проходит `go build ./...`
- [ ] Код проходит `go vet ./...`
- [ ] Proto проходит `easyp lint`
- [ ] Все тесты проходят (`just test`)
- [ ] Добавлены тесты для новой функциональности
- [ ] Документация обновлена (если нужно)
- [ ] Миграции протестированы на чистой БД
- [ ] Commit messages следуют Conventional Commits

### Review процесс (в разработке)

1. **Автоматические проверки** (CI):
   - `easyp lint` + `easyp generate`
   - `go build` + `go vet`
   - `go test -short`

2. **Ручное ревью**:
   - Архитектурная корректность
   - Безопасность (IDOR, SQL injection, XSS)
   - Производительность (N+1 queries, индексы)
   - Совместимость API

3. **После merge**:
   - Squash-merge в `develop`
   - Автоматический деплой в staging
   - Ручной деплой в production после тестирования

---

## ⚠️ Специфика мемориального проекта

### Качество данных

- **Верификация**: каждая биография должна иметь минимум один источник
- **Источники**: ссылки на архивы, книги, публикации, интервью
- **Фото**: только с подтверждённым авторством или из общественных источников
- **Даты**: используйте `FlexibleDate` для неполных дат (например, «Лето 1943»)

### Чувствительность контента

- **Личные данные**: не публикуйте персональные данные живых родственников без согласия
- **Фото погибших**: уважительное отношение, без Сенсационали́зма
- **Причины гибели**: формулировки должны быть нейтральными и фактическими
- **Политика**: избегайте политических оценок, фокус на личности героя

### Локализация

- Основной язык: **русский**
- Сообщения об ошибках: локализованные через `delivery/v1/errors.go`
- UI: поддержка русского + английского (???)

---

## 📞 Связь с командой

### Каналы коммуникации

| Канал | Назначение |
|---|---|
| [Codeberg Issues](https://codeberg.org/Thr0TT1e/emh/issues) | Баги, фичи, вопросы |
| [Codeberg Discussions](https://codeberg.org/Thr0TT1e/emh/discussions) | Общие обсуждения |
| Telegram (по запросу) | Оперативная связь с мейнтейнерами |
| info@noble24.ru | Общие обсуждения |

### Как задать вопрос

1. **Поиск**: проверьте, не задавался ли вопрос ранее
2. **Контекст**: укажите версию, ОС, шаги воспроизведения
3. **Логи**: приложите релевантные логи (без секретов!)
4. **Минимальный пример**: для багов — минимальный код для воспроизведения

### Как предложить фичу

1. Откройте issue с тегом `enhancement`
2. Опишите проблему, которую решает фича
3. Предложите решение (опционально)
4. Дождитесь обсуждения с мейнтейнерами
5. После согласования — начинайте реализацию

---

## 🎁 Признание вклада

Все участники добавляются в `CONTRIBUTORS.md` (с разрешения).

---

## 📚 Полезные ссылки

- [README.md](./README.md) — обзор проекта
- [TECHNICAL PASSPORT - EMH Backend.md](../docs/TECHNICAL_PASSPORT-EMH_Backend.md) — технический паспорт
- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) — детальная архитектура

---

## ❓ FAQ

### Q: Можно ли использовать Docker вместо Podman?
**A:** Да, все `podman` команды работают с `docker`. Quadlet-файлы специфичны для Podman, но локальная разработка через `podman-compose` (или `docker-compose`) работает одинаково.

### Q: Почему Connect RPC, а не REST или gRPC?
**A:** Connect сочетает простоту HTTP+JSON (удобно для браузера) с типобезопасностью gRPC. Не требует gRPC-Web прокси, работает из любого HTTP-клиента.

### Q: Можно ли добавлять новые поля в Hero?
**A:** Да, но следуйте правилам обратной совместимости:
1. Добавляйте поля в конец proto-сообщения
2. Помечайте старые поля как `DEPRECATED`
3. Обновляйте миграции и репозитории
4. Сохраняйте поддержку старых полей в API

### Q: Как работает rate-limiting?
**A:** In-memory `sync.Map` по IP + группа (auth/submission/read). Фоновая очистка записей. Admin и Media сервисы исключены (защищены JWT).

---

<p align="center">
  <sub>Спасибо за ваш вклад в сохранение памяти!</sub>
</p>

<p align="center">
  <em>«Никто не забыт, ничто не забыто»</em>
</p>
