# QWEN.md — правила работы с репозиторием EMH

Монорепозиторий мемориального проекта «Вечная память героям».
**Рабочий язык проекта — русский**: комментарии, документация, сообщения об ошибках,
commit messages и UI-тексты пишутся по-русски.

## Карта репозитория

| Путь | Что это |
|---|---|
| `proto/emh/v1/` | **Источник истины для API.** 17 `.proto`-файлов (Connect RPC) |
| `backend/` | Go 1.26.5, модуль `codeberg.org/Thr0TT1e/emh/backend` |
| `frontend/` | Nuxt 4 + PrimeVue 4 + Tailwind 4, пакетный менеджер **pnpm** |
| `backend/migrations/` | goose-миграции `NNNNN_описание.sql` |
| `infra/` | Podman Quadlet, compose-файлы, Caddy, Prometheus |
| `docs/` | `ARCHITECTURE.md`, `glossary.md`, `ADR/`, `specs/`, `emh/v1/` |
| `EMH/` | Bruno/OpenAPI-коллекция для ручных запросов |
| `Justfile` | **Единственный** task runner, лежит в корне (не в `backend/`) |

## Никогда не редактировать вручную (сгенерировано)

- `backend/internal/gen/` — Go-код из proto (`just gen-sdk`)
- `frontend/app/sdk/` — TS-код из proto (`just gen-sdk`). Закоммичен в git, не в `.gitignore`
- `docs/emh/v1/*.md` — API-референс от `easydoc`
- `frontend/app/components.d.ts` — декларации unplugin-vue-components
- `backend/vendor/` — `go mod vendor`

Правка в proto → всегда `just gen-sdk` → потом правки в backend и frontend.

## Команды

Все `just`-цели запускаются **из корня репозитория**.

```bash
# Proto
just gen-sdk                # easyp lint -r proto && easyp generate (Go + TS + docs)

# Миграции (goose)
just up / just down / just status / just reset
just create add_hero_tags   # создать новую миграцию
just psql

# Тесты backend
just test                   # unit: -race -count=1, без БД
just test-int               # интеграционные, -tags=integration, -p 1
just test-all
just cover                  # coverage.out + coverage.html

# Тестовая БД (podman)
just test-db-up / test-db-down / test-db-migrate / test-db-reset

# Проверки Go
go build -C backend ./...
go vet -C backend ./...

# Frontend (из frontend/)
pnpm install
pnpm dev                    # nuxt dev --host, http://localhost:3000
pnpm build                  # nuxt build
pnpm generate               # SSG
pnpm exec oxlint            # линтер (скрипта в package.json НЕТ)
pnpm exec oxfmt             # форматтер (скрипта в package.json НЕТ)
```

⚠️ `just dev`, `just build`, `just lint`, `just hash-pass` упоминаются в `backend/README.md`,
но **в Justfile их нет** — не вызывать. Реальные цели: `just --list`.

Backend запускается напрямую: `go run -C backend ./cmd/api` (порт `:3480`),
hot reload — `air` (`.air.toml`, бинарь в `backend/tmp/`).

Интеграционные тесты требуют поднятой тестовой БД (`just test-db-up`, порт 5444).
`just test` БД не требует.

**Frontend-тестов нет** — ни vitest, ни playwright не настроены. Не выдумывать
`pnpm test`; если нужны тесты, сначала согласовать фреймворк.

## Backend: архитектура и конвенции

Clean Architecture, направление зависимостей строго внутрь:

```
domain  ←  usecase  ←  delivery
   ↑          ↑
repository   storage (реализации интерфейсов domain)
```

- `internal/domain/` — чистый Go, без внешних зависимостей
- `internal/usecase/` — бизнес-логика; CQRS: `HeroUseCase` (команды) + `HeroQueryUseCase` (чтение агрегата, параллельная загрузка связей через `errgroup`)
- `internal/delivery/v1/` — Connect RPC серверы; порядок интерцепторов: `rpcMetrics` (внешний) → `auth` → `validate`
- `internal/repository/pg/` — Squirrel + pgx; `internal/storage/s3/` — MinIO
- `internal/{llm,mcp,metrics,netutil,smtp,auth,config}` — инфраструктурные пакеты
- `cmd/{api,hashpassword,mcp}` — три бинарника

### Ошибки

Никаких `strings.Contains(err.Error(), ...)`. Два сосуществующих механизма:

1. Sentinel-ошибки в `internal/domain/errors.go` + `errors.Is(err, domain.ErrNotFound)`
2. Типизированные коды `domain.ErrorCode` (`internal/domain/error_codes.go`), создаются через `domain.NewAppError(code, message)`

Маппинг на Connect-коды и логирование — только через `mapDomainError(ctx, logger, err, "…")`
из `internal/delivery/v1/errors.go`. Русские сообщения клиентам лежат там же в `errorMessages map[domain.ErrorCode]string`.
**Добавил новый код ошибки → обязательно добавил запись в `errorMessages`.**

Оборачивание: `fmt.Errorf("hero: %w", err)`. `ctx context.Context` — всегда первым параметром.

### Прочее

- Все Update-методы — partial update через `FieldMask []string`; защита обязательных полей проверяется в usecase
- Даты — `FlexibleDate` (`common.proto`), не `Timestamp`
- PK — UUIDv7; пагинация курсорная (композитные курсоры)
- Миграции: обязательный блок `-- +goose Down`, `COMMENT ON` для таблиц/колонок, индексы под фильтры
- `backend/config/config.yaml` в **`.gitignore`** — не создавать и не коммитить; секреты через env (`infra/podman/.env.example`)

## Proto: требования easyp-линта

Линт строгий и падает без комментариев. Обязательно:

- leading-комментарий на **каждый** `enum`, значение enum, `message`, поле, `oneof`, `rpc`, `service`
- только unary RPC (клиентский и серверный стриминг запрещены)
- нулевое значение enum — с суффиксом `_NONE`; сервисы — с суффиксом `Service`
- `FIELD_LOWER_SNAKE_CASE`, `MESSAGE_PASCAL_CASE`, `ENUM_VALUE_UPPER_SNAKE_CASE`
- валидация через `buf.validate` (`(buf.validate.field).string.uuid = true` и т.п.)

Новые поля — в конец сообщения; устаревшие помечать `DEPRECATED`, не удалять номера.
Breaking-проверка идёт против `master`.

## Frontend: конвенции

- Nuxt 4 app-директория: `app/{pages,components,composables,layouts,plugins,middleware,lib,constants,types,themes,assets,sdk}` + `server/` (Nitro)
- SFC: `<script setup lang="ts">`, содержимое `<script>`/`<style>` с отступом 2 пробела (`vueIndentScriptAndStyle: true`)
- Форматтер `oxfmt`: printWidth 100, одинарные кавычки, сортировка импортов; **`.vue` и `app/sdk/*` исключены** из форматирования
- Линтер `oxlint`: плагины typescript/unicorn/oxc/eslint/import/vue, `correctness: error`, `no-alert: error`
- PrimeVue — `autoImport: true`, компоненты не импортировать руками
- Иконки: **unplugin-icons + MDI/Carbon**, не PrimeIcons. `<i-mdi-account />` вместо `<i class="pi pi-user" />` (ADR-003)
- Тяжёлые компоненты ниже первого экрана — через `defineAsyncComponent` (ADR-002)
- Tailwind 4 через `@tailwindcss/vite` + `lightningcss`, `css: ['~/assets/css/main.css']`
- Dark mode: `@nuxtjs/color-mode`, `classSuffix: '-mode'` → классы `light-mode`/`dark-mode`; PrimeVue синхронизирован через `darkModeSelector: '.dark-mode'`

### Доступ к API

Только через провайдер, не создавать транспорты заново:

```ts
const api = useApi();                       // app/composables/useApi.ts → $api из app/plugins/connect.ts
const res = await api.hero.getHeroDetail({ id });
```

`app/plugins/connect.ts` содержит auth-интерцептор: сам делает refresh на `401`,
не ретраит `Login/Refresh/Logout`, а `429` и `403` пробрасывает в UI.
Проверки кодов — хелперами из `app/lib/errors.ts` (`isRateLimited`, `isUnauthenticated`, `isPermissionDenied`, `isNotFound`).

Protobuf ↔ plain JSON: `toPlain(Schema, msg)` из `app/lib/pb.ts` (выводит точный `XxxJson`-тип),
enum из JSON-строки — `toEnum()`. Типы вида `FlexibleDateJson`, `DatePrecisionJson` берутся из `~/sdk/emh/v1/*_pb`.

Runtime config: `apiInternalBaseUrl` (server-only, `NUXT_API_INTERNAL_BASE_URL`) для SSR внутри
контейнера и `public.apiBaseUrl` (`NUXT_PUBLIC_API_BASE_URL`) для браузера. Читать через
`useRuntimeConfig()`, не через `process.env` в компонентах.

### Кэширование и ISR

- ISR-кэш Nitro на `fs`-драйвере в `.data/nitro-routes`
- `server/api/cache/purge.post.ts` — инвалидация по списку путей, `server/api/cache/debug.get.ts` — отладка;
  `server/plugins/ensure-cache-dir.ts` создаёт каталог
- В `nuxt.config.ts` ISR **намеренно отключён** (`isr: false` на `/heroes`, `/heroes/**`, `/conflicts`, `/locations`) —
  закомментированные блоки с `expiration`/`allowQuery` оставлены для отладки прода.
  Не включать и не «раскомментировать» без явного запроса
- Пререндерятся `/`, `/contacts`, `/about`; `/admin/**` — `robots: false`
- `sitemap.zeroRuntime: true` + динамический сбор `/heroes/*` через `ListHeroes` на этапе генерации

## Документация и процесс

Перед изменением поведения заглянуть в:

- `docs/ARCHITECTURE.md` — слои, потоки данных, воркеры, безопасность
- `docs/glossary.md` — доменная терминология (Award/Ribbon/Bar, HeroLocation, FlexibleDate, FaceBox…)
- `docs/ADR/001..005` — принятые решения (image delivery, code splitting, иконки, OpenLayers, орденские планки)
- `docs/specs/` — активные спецификации
- `CONTRIBUTING.md` — стиль кода, чеклист PR
- `CONTEXT_SNAPSHOT-*.md` — снимки состояния по спринтам. **Могут отставать от кода**:
  при расхождении верить коду и git-истории, а не снимку
- `docs/skills/` — сторонняя библиотека скиллов (Matt Pocock), не рантайм проекта

Новое архитектурное решение оформлять ADR в `docs/ADR/` по образцу существующих
(Статус / Дата / Контекст / Решение / Альтернативы).

### Git

Conventional Commits, тема на русском:

```
<type>(<scope>): <описание>
```

- типы: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `security`
- области: `hero`, `photo`, `auth`, `media`, `db`, `proto`, `worker`, `api`, `map`, `images`
- ветки: `master` (production), `develop`, `feature/`, `fix/`, `refactor/`, `docs/`
- remote — **Codeberg**, не GitHub

## Специфика мемориального проекта

- Каждая биография должна иметь минимум один подтверждённый источник (`HeroSource`)
- Неполные исторические даты («лето 1943», «28 июля») — только через `FlexibleDate` с корректной `precision`
- Формулировки нейтральные и фактические, без политизации и сенсационализма
- Персональные данные живых родственников не публиковать
- Фото — только с подтверждённым авторством или из открытых источников

## Рабочий процесс агента

1. Изменения в API начинать с `proto/emh/v1/`, затем `just gen-sdk`, затем backend и frontend
2. После правок Go: `go build -C backend ./... && go vet -C backend ./...` и `just test`
3. После правок proto: `just gen-sdk` (включает `easyp lint`)
4. После правок frontend: `pnpm exec oxlint` и, где применимо, `pnpm build`
5. В репозитории есть незакоммиченная работа пользователя (изменённые `proto/emh/v1/*.proto`, переименованные ADR, новые `docs/specs/`). Не откатывать, не коммитить и не «причёсывать» её
