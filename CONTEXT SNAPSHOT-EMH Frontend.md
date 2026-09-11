# CONTEXT SNAPSHOT — EMH Frontend
Технический паспорт проекта «Вечная память героям»

## 0. Метаданные снимка

| Параметр | Значение |
|---|---|
| Проект | Вечная память героям (EMH) |
| Scope | Только `frontend/` (Nuxt 4) |
| Назначение снимка | Полное восстановление контекста фронтенд-разработки в новой сессии |
| Фаза | Публичная часть, админка, модерация, SEO-стек, форма обратной связи **завершены**. Переход с SSG на ISR **завершён**. Техдолг закрыт. **Критичные изменения из справки + фичи Спринта 7 реализованы** |
| Дата снимка | 2026-08-30 (обновлённый) |
| Предыдущий снимок | 2026-08-30 |
| Production VPS | 82.202.139.132 |
| Канонический домен | вежливые.рус (IDN, кириллица) |
| Зеркала (301 → канонический) | вечнаяпамятьгероям.рус, emh.su, neverforgotten.ru |
| Технические поддомены | api.вежливые.рус → бэкенд, s3.вежливые.рус → MinIO, console.minio.вежливые.рус |

---

## 1. Стек и зависимости (зафиксировано в `package.json`)

### Runtime

| Компонент | Версия | Назначение |
|---|---|---|
| Nuxt | ^4.5.2 | Фреймворк (структура `app/`) |
| Vue | ^3.5.42 | UI |
| PrimeVue | ^4.5.5 | UI-кит + `@primeuix/themes` |
| @bufbuild/protobuf | ^2.14.1 | protobuf-es v2 (рантайм) |
| @connectrpc/connect-web | ^2.1.2 | Connect RPC транспорт |
| vue-advanced-cropper | ^2.8.9 | Face Box UI |
| OpenLayers (ol) | ^10.10.0 | Карты (локации героев) |
| Node.js | 22.23.1 | Runtime |
| Пакетный менеджер | pnpm@11.21.0 | `type: module` |
| TypeScript | ^7.0.2 | Типизация |
| oxlint | ^1.81.0 | Линтер |
| oxfmt | ^0.62.0 | Форматтер |

### SEO-стек (без изменений)

| Компонент | Версия | Назначение |
|---|---|---|
| nuxt-seo-utils | 8.4.2 | Глобальный site config |
| @nuxtjs/sitemap | 8.5.0 | Динамический sitemap |
| @nuxtjs/robots | 6.2.0 | Автогенерация `robots.txt` |
| nuxt-schema-org | 6.3.0 | JSON-LD (Person, BreadcrumbList) |
| nuxt-og-image | 6.7.8 | OG-карточки через Satori |
| @nuxt/scripts | 1.3.8 | Ленивая загрузка Яндекс.Метрики (счётчик 111894071) |
| @nuxt/hints | 1.1.4 | Подсказки для разработчиков |
| @takumi-rs/core | ^2.13.4 | Ядро для OG-изображений |

### Удалённые модули

| Модуль | Причина удаления |
|---|---|
| @nuxt/fonts | Body Timeout Error при сборке без интернета. Заменён на ручной `@font-face` в `main.css` |

---

## 2. Режим сборки: переход с SSG на ISR (завершён)

### Конфигурация (`nuxt.config.ts`)

```typescript
nitro: {
  storage: {
    'nitro:routes': {
      driver: 'fs',
      base: './.data/nitro-routes',
    },
  },
  prerender: {
    failOnError: false,
    ignore: ['/api/**'],
    crawlLinks: false,
  },
  routeRules: {
    '/': { swr: 3600 },
    '/heroes': { swr: 300 },      // список
    '/heroes/': { swr: 300 },     // карточки (5 минут)
    '/conflicts': { swr: 3600 },
    '/locations': { swr: 3600 },
    '/contacts': { prerender: true },
    '/about': { prerender: true },
    '/admin/': { robots: false },
  },
},
experimental: {
  typedPages: true,
},
```

⚠️ **Наблюдение:** `/heroes/` в `routeRules` не покрывает вложенные пути `/heroes/<id>`. Если нужен кэш на карточки, следует использовать `/heroes/**`. Проверить поведение на проде.

### Invalidation ISR-кэша

```
Правка в админке → useCachePurge().purgeHero(id)
   ↓
POST /api/cache/purge → useStorage('cache').removeItem()
   ↓
Следующий запрос → свежий рендер
```

- **Эндпоинт:** `app/server/api/cache/purge.post.ts` (на уровне проекта, не внутри `app/`)
- **Composable:** `app/composables/useCachePurge.ts`
- **Точки вызова:** `HeroForm.save()` (create/edit), `admin/heroes/[id].vue.onChanged()` (M:N)

### Обработка ошибок в интерцепторе (`connect.ts`)

| Код ошибки | Поведение |
|---|---|
| 401 Unauthenticated | Авто-рефреш токена + повтор запроса. Если рефреш не удался — редирект на логин |
| 403 PermissionDenied | Пробрасывается дальше (пользователь авторизован, но нет роли `admin`). НЕ делаем редирект на логин |
| 429 ResourceExhausted | Пробрасывается дальше (не делаем ретрай). Пусть UI решает |
| Прочие | Пробрасываются дальше |

**Методы без ретрая:** `/Login`, `/Refresh`, `/Logout` (чтобы избежать бесконечного цикла)

### Хелперы для обработки ошибок (`errors.ts`)

| Функция | Что проверяет |
|---|---|
| `isConnectError(err)` | Является ли ошибка `ConnectError` |
| `isRateLimited(err)` | 429 Too Many Requests |
| `isUnauthenticated(err)` | 401 Unauthenticated |
| `isPermissionDenied(err)` | 403 PermissionDenied |
| `isNotFound(err)` | 404 NotFound |
| `isAlreadyExists(err)` | 409 AlreadyExists |
| `retryAfterSeconds(err)` | Читает заголовок Retry-After |

---

## 3. Структура каталогов (актуальная)

```
frontend/
├── app/
│   ├── assets/css/
│   │   ├── main.css
│   │   └── admin.css
│   ├── components/
│   │   ├── HeroRecord.vue
│   │   ├── OgImage/
│   │   │   ├── HeroCard.takumi.vue
│   │   │   └── HomeCard.takumi.vue
│   │   └── admin/
│   │       ├── HeroForm.vue                # +ExtractionPanel, +useExtractionLinks, +parseExtractedDate
│   │       ├── HeroPhotos.vue              # пагинация через useHeroPhotos
│   │       ├── HeroAwards.vue              # гибкие даты для наград (Вариант C)
│   │       ├── HeroConflicts.vue           # боевой путь героя
│   │       ├── ExtractionPanel.vue         # извлечение из текста/URL + проверка размера 50КБ
│   │       ├── ExtractionResultPreview.vue # переиспользуемое превью (модерация)
│   │       ├── SubmissionReviewHistory.vue # история модерации заявки (Спринт 7)
│   │       ├── FaceBoxEditor.vue
│   │       ├── FlexibleDateInput.vue       # компонент ввода гибких дат
│   │       ├── AdminSidebarContent.vue     # содержимое сайдбара админки
│   │       └── ... (ImageUpload, LocationMap, AdminFlexibleDateInput)
│   ├── composables/
│   │   ├── useApi.ts
│   │   ├── useAuth.ts
│   │   ├── useContactForm.ts
│   │   ├── useCountUp.ts
│   │   ├── useHeroRegistry.ts
│   │   ├── useHeroPhotos.ts          # пагинация галерей
│   │   ├── useCachePurge.ts          # invalidation ISR
│   │   ├── useExtractionLinks.ts     # авто-привязка связей
│   │   └── useDuplicateCheck.ts      # клиентская проверка дубликатов заявок
│   ├── constants/auth.ts
│   ├── layouts/
│   │   ├── default.vue
│   │   └── admin.vue
│   ├── lib/
│   │   ├── api.ts                    # +extraction, +llmAdmin клиенты
│   │   ├── pb.ts
│   │   ├── format.ts
│   │   └── errors.ts
│   ├── middleware/admin.ts
│   ├── pages/
│   │   ├── index.vue
│   │   ├── about.vue
│   │   ├── contacts.vue
│   │   ├── submit.vue
│   │   ├── heroes/[id].vue
│   │   └── admin/
│   │       ├── heroes/
│   │       │   ├── index.vue         # фильтр статуса + кликабельный тег
│   │       │   ├── new.vue           # приём извлечения через проп
│   │       │   └── [id].vue
│   │       ├── submissions/
│   │       │   └── index.vue         # извлечение через LLM + sessionStorage + история модерации
│   │       ├── llm/
│   │       │   └── index.vue         # управление LLM-провайдерами (Спринт 7)
│   │       └── ... (keys, awards, conflicts, locations, login)
│   ├── plugins/
│   │   ├── connect.ts
│   │   └── reveal.ts
│   └── sdk/emh/v1/                   # Сгенерированный код (не править!)
│       ├── llm_admin_pb.ts           # LlmAdminService (Спринт 7)
│       ├── submission_pb.ts          # +ListSubmissionReviews, +content_hash
│       ├── hero_admin_pb.ts          # +award_date_info
│       └── ...
├── server/
│   └── api/cache/
│       └── purge.post.ts             # ISR cache purge endpoint
└── nuxt.config.ts
```

---

## 4. Production Docker-образ (ISR) — без изменений

Multi-stage build: `node:22.23.1-alpine` builder → `node:22.23.1-alpine` runtime. Порт 3000. `Podmanfile.front`, `emh-frontend.container`, `Caddyfile` соответствуют целевому состоянию.

---

## 5. Взаимодействие с бэкендом и типизация

**Протокол:** Connect Protocol (HTTP/1.1 + JSON, `useBinaryFormat: false`).
**Хелпер:** `toPlain` (извлекает строгий тип `XxxJson`).
**Enum:** в ответах — строки, для записи — `toEnum(Schema, strValue)`.

### Реестр клиентов (`api.ts`)

| Группа | Сервисы |
|---|---|
| Публичные | `hero`, `award`, `conflict`, `location`, `submission`, `media`, `auth`, `contact` |
| Админские | `heroAdmin`, `awardAdmin`, `conflictAdmin`, `locationAdmin`, `apiKeyAdmin`, `submissionAdmin`, `extraction`, `llmAdmin` |

### ExtractionService — извлечение данных через LLM

**RPC:**
- `ExtractHeroData({rawText, sourceUrl})` → извлечение из текста/URL
- `ExtractFromSubmission({submissionId})` → извлечение из `payload_json` заявки

**Возвращает:** `{hero, conflicts, awards, locations, warnings, duplicateHeroIds, sourceUrls}` (для `ExtractFromSubmission` — без `sourceUrls` и `duplicateHeroIds`).

**Валидация на клиенте:**
- Проверка размера текста перед отправкой: максимум 50 КБ (`MAX_TEXT_SIZE = 50 * 1024`)
- Используется `new Blob([text]).size` для определения размера

**UI-интеграция:**
1. **Создание героя:** `ExtractionPanel.vue` в `HeroForm` (режим `create`) → `applyExtractedData` → заполнение формы → `CreateHero` → `applyExtractionLinks` (авто-привязка).
2. **Модерация заявок:** `submissions/index.vue` → кнопка «Извлечь через LLM» → превью → «Создать героя» → `sessionStorage['emh-submission-extraction']` → `new.vue` → проп `initial-extraction` → `HeroForm`.

### LlmAdminService — управление LLM-провайдерами (Спринт 7)

**Назначение:** Управление провайдерами для извлечения данных героя. Позволяет администратору активировать/деактивировать провайдеров, тестировать их доступность, обновлять приоритеты и заметки.

**Провайдеры:**
- `ollama` — локальный провайдер (для разработки)
- `openai_compatible` — совместимый с OpenAI API (для продакшена)
- `anthropic` — Anthropic API (для продакшена)

**Таблица провайдеров:**
| Название | Тип | Модель | Активен |
|---|---|---|---|
| dev-ollama | ollama | llama3.2 | ✅ (по умолчанию) |
| prod-openai | openai_compatible | gpt-4o | ❌ |
| prod-anthropic | anthropic | claude-3-5-sonnet | ❌ |

**RPC:**
- `ListLlmProviders({})` → список провайдеров
- `SetActiveLlmProvider({id})` → активация провайдера
- `TestLlmProvider({id, testPrompt})` → тестирование доступности
- `UpdateLlmProvider({id, priority, notes, fieldMask})` → обновление приоритета/заметок

**Возвращает (`LlmProviderInfo`):** `{id, name, type, model, isActive, priority, notes, createdAt, updatedAt}`

**UI-интеграция:**
- Страница `/admin/llm` — таблица провайдеров с действиями:
  - **Активировать** — делает провайдера активным (остальные деактивируются триггером БД)
  - **Тестировать** — отправляет тестовый запрос к провайдеру для проверки доступности
  - **Изменить** — открывает диалог редактирования приоритета и заметок
- Диалог тестирования:
  - Ввод тестового промпта
  - Результат: успех/ошибка + время выполнения + ответ модели
- Диалог редактирования:
  - Приоритет (число)
  - Заметки (текст)
  - Поле `field_mask` для указания обновляемых полей

**Конфигурация провайдеров:**
- Конфигурация провайдеров хранится в таблице `llm_providers` в БД
- При активации провайдера через `SetActiveLlmProvider` триггер БД автоматически деактивирует остальных
- Приоритет используется для fallback-логики (если активный провайдер недоступен, используется следующий по приоритету)

**Тестирование провайдера:**
- Тестовый промпт отправляется к провайдеру
- Результат содержит:
  - `success` — флаг успешного получения ответа
  - `raw_response` — сырой текст ответа от модели
  - `processing_time_ms` — время выполнения запроса в миллисекундах (int64, конвертируется в number через `Number()`)
  - `error_message` — текст ошибки, если `success = false`

**Примечание:** `processing_time_ms` в `TestLlmProviderResponse` имеет тип `int64`, который в JSON приходит как `string`. Для отображения конвертируется через `Number()`.

### SubmissionAdminService — модерация заявок (обновлено в Спринт 7)

**Новые поля в `Submission`:**
- `content_hash` — SHA-256 хеш нормализованного `payload_json` (read-only)

**Новый RPC:**
- `ListSubmissionReviews({submissionId})` → история модерации заявки (аудит-трейл)

**Возвращает (`SubmissionReview`):** `{id, submissionId, reviewerName, decision, comment, createdAt}`

**История модерации:**
- Записи аудита неизменяемы: только добавляются, не редактируются
- История возвращается в обратном хронологическом порядке (новые записи первыми)
- `reviewerName` извлекается из JWT токена модератора

**Уникальные заявки:**
- Бэкенд проверяет уникальность заявки на этапе `CreateSubmission` через SHA-256 хеш нормализованного `payload_json`
- При совпадении с активной заявкой возвращается `CodeAlreadyExists` (HTTP 409)
- Отклонённые (REJECTED) заявки не считаются дубликатами — их можно переотправлять

**Уведомления заявителям:**
- После успешного `ReviewSubmission` бэкенд ставит уведомление в очередь `SubmissionEmailWorker`
- Воркер асинхронно отправляет письмо заявителю (100-500мс)
- Шаблон зависит от решения (Approve/Reject)
- Уведомление отправляется best-effort — если SMTP недоступен, письмо не придёт, но модерация всё равно завершится успешно

### Гибкие даты для наград (Вариант C, Спринт 7)

**Новые поля:**
- `AddHeroAwardRequest.award_date_info` — гибкая дата награждения (`FlexibleDate`)
- `HeroAward.award_date_info` — гибкая дата награждения (`FlexibleDate`)

**Приоритет:**
- Если `award_date_info` заполнено, оно имеет приоритет над `award_date`
- `award_date` остаётся для обратной совместимости

**Формат `FlexibleDate`:**
- `anchor_date` — опорная дата для сортировки и фильтрации (формат зависит от `precision`)
- `precision` — уровень точности даты (`DATE_PRECISION_*`)
- `display_text` — человекочитаемое представление даты из источника

**Уровни точности (`DatePrecision`):**
| Значение | Описание | Формат `anchor_date` |
|---|---|---|
| `DATE_PRECISION_UNSPECIFIED` | Точность не указана | Должно быть пустым |
| `DATE_PRECISION_EXACT` | Точная дата | `YYYY-MM-DD` |
| `DATE_PRECISION_MONTH` | Месяц и год | Первое число месяца |
| `DATE_PRECISION_YEAR` | Только год | 1 января |
| `DATE_PRECISION_SEASON` | Сезон и год | Условная дата начала сезона |
| `DATE_PRECISION_DAY_MONTH` | День и месяц без года | Должно быть пустым |
| `DATE_PRECISION_RANGE` | Диапазон дат | Зарезервировано |
| `DATE_PRECISION_UNKNOWN` | Дата неизвестна | Должно быть пустым |

**Конвертация в `HeroAwards.vue`:**
- Используется `fromJson(FlexibleDateSchema, json)` для преобразования `FlexibleDateJson` в `FlexibleDate` (protobuf Message)
- Если дата пустая или `UNSPECIFIED`, отправляется `DATE_PRECISION_UNKNOWN`, чтобы бэкенд её очистил

### Проверка роли `admin`

**Где проверяется:**
- В `useAuth.ts` через `claims` из JWT токена
- `claims` — computed, который автоматически пересчитывается при рефреше токена

**Как использовать:**
```typescript
const { claims } = useAuth();
const isAdmin = computed(() => claims.value?.role === 'admin');
```

**Примечание:** Бэкенд проверяет роль `admin` на своей стороне. Фронтенд может использовать `claims` для отображения роли в сайдбаре (`AdminSidebarContent.vue`).

---

## 6. SEO-стек и аналитика — без изменений

Яндекс.Метрика (счётчик 111894071) через `@nuxt/scripts` с `defer`.

---

## 7. Дизайн-система (`emhPreset`) — без изменений

---

## 8. Деплой (Production Workflow) — без изменений

Registry-centric deployment: `podman build` → `podman push` → на VPS `podman pull` → `systemctl --user restart emh-frontend.service`.

---

## 9. Состояние реализации (Final Checklist)

| Экран/модуль | Статус |
|---|---|
| Публичный список и карточка героя | ✅ Завершено |
| Страницы `/about`, `/contacts` | ✅ Завершено |
| Форма обратной связи (ContactService) | ✅ Завершено |
| Краудсорсинг `/submit` + модерация | ✅ Завершено |
| Админка (герои, справочники, ключи) | ✅ Завершено |
| Загрузка медиа (Presigned URL → MinIO) | ✅ Завершено |
| SEO-стек (метатеги, sitemap, robots) | ✅ Завершено |
| Schema.org + OG-карточки | ✅ Завершено |
| Яндекс.Метрика | ✅ Завершено |
| Локальные шрифты | ✅ Завершено |
| Переход с SSG на ISR | ✅ Завершено |
| Invalidation ISR-кэша после редактирования | ✅ Завершено |
| Пагинация галерей (кнопка «Загрузить ещё») | ✅ Завершено |
| Фильтр статуса в админке героев | ✅ Завершено |
| Авто-привязка связей из извлечения | ✅ Завершено |
| Модерация заявок через LLM | ✅ Завершено |
| Обработка ошибок 401, 403, 429 | ✅ Завершено (Спринт 7) |
| Проверка размера текста для LLM (50 КБ) | ✅ Завершено (Спринт 7) |
| Проверка роли `admin` | ✅ Завершено (Спринт 7) |
| `LlmAdminService` — управление провайдерами | ✅ Завершено (Спринт 7) |
| История модерации заявок | ✅ Завершено (Спринт 7) |
| Гибкие даты для наград | ✅ Завершено (Спринт 7) |
| Уникальные заявки (анти-дубликат) | ✅ Завершено (Спринт 7) |
| Уведомления заявителям | ✅ Завершено (Спринт 7) |
| Техдолг (план 2026-08-30) | ✅ Закрыт |

---

## 10. Техдолг и известные ограничения

| # | Файл | Проблема | Действие |
|---|---|---|---|
| 1 | `HeroForm.vue` | fieldMask захардкожен | ✅ Решено |
| 2 | Thumbnail-генерация | Асинхронная (1–5 сек), `thumbnail_url` может быть пустым | Использовать `url` как fallback + polling |
| 3 | Дубли `robots`-метатегов | Глобальный `seoMeta` + локальный `useSeoMeta` | Проверить приоритет после деплоя |
| 4 | ISR-кэш | Жил до 24 ч после правки | ✅ Решено через `/api/cache/purge` |
| 5 | `@nuxt/fonts` | Body Timeout Error при сборке без интернета | ✅ Решено — модуль удалён, заменён на ручной `@font-face` |
| 6 | `routeRules` | `/heroes/` не покрывает `/heroes/<id>` | Проверить, возможно нужно `/heroes/**` |
| 7 | `HeroPhotos.vue` | `move()` / `setMain()` использовали `props.photos` | ✅ Решено (реактивный `photos` из composable) |
| 8 | `processing_time_ms` в `TestLlmProviderResponse` | Тип `int64`, в JSON приходит как `string` | ✅ Решено — конвертируется через `Number()` |

---

## 11. Следующие шаги (приоритеты для новой сессии)

### P0 — Критические

| # | Задача | Описание |
|---|---|---|
| P0.1 | Проверить `routeRules` `/heroes/` | Убедиться, что карточки героев попадают под кэш (возможно нужно `/heroes/**`) |
| P0.2 | Валидация критичных изменений на проде | Проверить обработку 401/403/429, проверку размера текста, проверку роли `admin` |

### P1 — Высокий приоритет

| # | Задача | Описание |
|---|---|---|
| P1.1 | Проверить дубли `robots`-метатегов | После деплоя: `curl -s https://вежливые.рус/heroes/<id> \| grep robots` — должен быть ровно один тег |
| P1.2 | Thumbnail fallback + polling | Использовать `url` как fallback, polling до появления `thumbnailUrl` |

### P2 — Средний приоритет

| # | Задача | Описание |
|---|---|---|
| P2.1 | Рефакторинг `ExtractionPanel` на `ExtractionResultPreview` | Убрать дублирование разметки превью |
| P2.2 | Мониторинг через метрики | Визуализация Prometheus метрик из бэкенда |
| P2.3 | Повышение покрытия тестами | `internal/smtp` (42.4%) и `internal/usecase` (60.5%) |

### P3 — Отложено

| # | Задача |
|---|---|
| P3.1 | Множественные face box (групповые фото) |
| P3.2 | Кэширование `ListHeroes` на уровне Nitro |
| P3.3 | Preload критических фото (LCP) через `<link rel="preload">` |
| P3.4 | Service Worker для оффлайн-режима админки |
| P3.5 | Гибкие даты для всех сущностей (конфликты, локации) |
| P3.6 | Grafana-дашборд |

---

## 12. Точки восстановления контекста (эталоны кода)

| Паттерн | Эталон |
|---|---|
| Форма с `fieldMask` | `app/components/admin/HeroForm.vue` |
| Курсорная пагинация + фильтр статуса | `app/pages/admin/heroes/index.vue` |
| Presigned Upload + пагинация галерей | `app/components/admin/HeroPhotos.vue` |
| Face Box Editor | `app/components/admin/FaceBoxEditor.vue` |
| Auto-refresh при 401 | `app/plugins/connect.ts` |
| Обработка 403/429 | `app/lib/errors.ts` |
| Краудсорсинг | `app/pages/submit.vue` |
| SEO карточки героя | `app/pages/heroes/[id].vue` |
| Форма обратной связи | `app/pages/contacts.vue` + `app/composables/useContactForm.ts` |
| OG-карточка | `app/components/OgImage/HeroCard.takumi.vue` |
| Пагинация галерей | `app/composables/useHeroPhotos.ts` |
| ISR invalidation | `app/composables/useCachePurge.ts` + `server/api/cache/purge.post.ts` |
| LLM извлечение | `app/components/admin/ExtractionPanel.vue` |
| Авто-привязка связей | `app/composables/useExtractionLinks.ts` |
| Модерация через LLM | `app/pages/admin/submissions/index.vue` |
| Передача данных между страницами | `sessionStorage` + проп `initial-extraction` |
| Проверка размера текста для LLM | `app/components/admin/ExtractionPanel.vue` (`MAX_TEXT_SIZE`) |
| Управление LLM-провайдерами | `app/pages/admin/llm/index.vue` |
| История модерации заявок | `app/components/admin/SubmissionReviewHistory.vue` |
| Гибкие даты для наград | `app/components/admin/HeroAwards.vue` + `app/components/admin/FlexibleDateInput.vue` |
| Клиентская проверка дубликатов | `app/composables/useDuplicateCheck.ts` |

---

## 13. Changelog сессии (что было сделано 2026-08-30)

### ✅ Завершено

1. **Валидация на проде**
   - Пользователь подтвердил, что всё работает (тосты, фильтрация, история модерации)

2. **Критичные изменения из справки**
   - Обработка ошибок 401, 403, 429 в `connect.ts` и `errors.ts`
   - Проверка размера текста перед отправкой на `ExtractHeroData` (максимум 50 КБ)
   - Проверка роли `admin` через `claims` из JWT токена

3. **Фичи Спринта 7**
   - **Вариант A: `LlmAdminService`**
     - Страница `/admin/llm` — управление LLM-провайдерами
     - Активация/деактивация провайдеров
     - Тестирование доступности провайдеров
     - Обновление приоритетов и заметок
   - **Вариант B: Заявки с обратной связью**
     - История модерации заявок (`ListSubmissionReviews`)
     - Колонки "Рассмотрена" и "Комментарий" в списке заявок
     - Компонент `SubmissionReviewHistory.vue`
     - Уникальные заявки (анти-дубликат через `content_hash`)
     - Уведомления заявителям (тост после `ReviewSubmission`)
   - **Вариант C: Гибкие даты**
     - Гибкие даты для наград (`award_date_info`)
     - Компонент `FlexibleDateInput.vue` для ввода гибких дат
     - Конвертация `FlexibleDateJson` → `FlexibleDate` через `fromJson`

### 🔧 Исправленные баги

1. **Обработка 403 в `connect.ts`**
   - Добавлена явная обработка `Code.PermissionDenied`
   - НЕ делаем редирект на логин (сессия валидна)

2. **Проверка размера текста в `ExtractionPanel.vue`**
   - Добавлена проверка `MAX_TEXT_SIZE = 50 * 1024`
   - Используется `new Blob([text]).size` для определения размера

3. **Конвертация `processing_time_ms` в `llm/index.vue`**
   - Тип `int64` в JSON приходит как `string`
   - Конвертируется через `Number()`

### ⚠️ Наблюдения

1. **`routeRules` в `nuxt.config.ts`**
   - `/heroes/` не покрывает вложенные пути `/heroes/<id>`
   - Возможно, нужно использовать `/heroes/**` для кэширования карточек

2. **`@nuxt/fonts` удалён**
   - Модуль вызывал Body Timeout Error при сборке без интернета
   - Заменён на ручной `@font-face` в `main.css`

3. **`processing_time_ms` в `TestLlmProviderResponse`**
   - Тип `int64`, в JSON приходит как `string`
   - Нужно конвертировать через `Number()` для отображения
