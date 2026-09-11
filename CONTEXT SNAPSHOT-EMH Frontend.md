# CONTEXT SNAPSHOT — EMH Frontend (Спринт 8)

## Технический паспорт проекта «Вечная память героям»

---

## 0. Метаданные снимка

| Параметр | Значение |
|---|---|
| Проект | Вечная память героям (EMH) |
| Scope | Только `frontend/` (Nuxt 4) |
| Назначение снимка | Полное восстановление контекста фронтенд-разработки в новой сессии |
| Фаза | Публичная часть, админка, модерация, SEO-стек, форма обратной связи **завершены**. Переход с SSG на ISR **завершён**. Техдолг закрыт. Критичные изменения из справки + фичи Спринта 7 реализованы. **Спринт 8 (оптимизация фронтенда) завершён** |
| Дата снимка | 2026-09-11 |
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
| PrimeVue | 4.5.5 | UI-кит + `@primeuix/themes` (зафиксирована, не `^`) |
| @bufbuild/protobuf | ^2.14.1 | protobuf-es v2 (рантайм) |
| @connectrpc/connect-web | ^2.1.2 | Connect RPC транспорт |
| vue-advanced-cropper | ^2.8.9 | Face Box UI |
| OpenLayers (ol) | ^10.10.0 | Карты (локации героев) |
| Node.js | 22.23.1 | Runtime |
| Пакетный менеджер | pnpm@12.3.4 | `type: module` |
| TypeScript | ^7.0.2 | Типизация |
| oxlint | ^1.81.0 | Линтер |
| oxfmt | ^0.66.0 | Форматтер |

### SEO-стек

| Компонент | Версия | Назначение |
|---|---|---|
| nuxt-seo-utils | ^8.5.0 | Глобальный site config |
| @nuxtjs/sitemap | ^8.5.0 | Динамический sitemap |
| @nuxtjs/robots | ^6.2.0 | Автогенерация `robots.txt` |
| nuxt-schema-org | ^6.3.1 | JSON-LD (Person, BreadcrumbList) |
| nuxt-og-image | ^6.7.8 | OG-карточки через Satori |
| @nuxt/scripts | ^1.3.8 | Ленивая загрузка Яндекс.Метрики (счётчик 111894071) |
| @nuxt/hints | ^1.1.4 | Подсказки для разработчиков |
| @takumi-rs/core | ^2.13.6 | Ядро для OG-изображений |

### Модули оптимизации и DX (добавлены в Спринте 8)

| Модуль | Версия | Назначение |
|---|---|---|
| @nuxt/fonts | ^0.14.0 | Локальные шрифты (`provider: 'local'`, woff2) |
| @nuxtjs/html-validator | ^2.1.0 | Валидация HTML в dev/CI |
| @vueuse/nuxt | ^14.4.0 | Composables (`useIntersectionObserver` и др.) |
| @nuxtjs/color-mode | ^4.0.1 | Dark/Light mode с `classSuffix: '-mode'` |
| eslint-plugin-vuejs-accessibility | ^2.6.0 | A11y-правила для Vue-компонентов |

### Удалённые модули

| Модуль | Причина |
|---|---|
| *(пусто)* | Все ранее удалённые модули возвращены или не применялись |

> ⚠️ **Историческая справка по `@nuxt/fonts`:** В предыдущем спринте модуль вызывал `Body Timeout Error` при сборке без интернета и был заменён на ручной `@font-face`. В Спринте 8 проблема решена: шрифты лежат локально в `public/fonts/*.woff2`, модуль настроен с `provider: 'local'`.

---

## 2. Режим сборки: переход с SSG на ISR (завершён, обновлён в Спринте 8)

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
    // ISR для списка героев с защитой от cache poisoning
    '/heroes': {
      isr: {
        expiration: 300,
        passQuery: true,
        allowQuery: [
          'search_query', 'conflict_id', 'location_id',
          'cursor', 'date_from', 'date_to',
        ],
      },
    },
    '/heroes/': {
      isr: {
        expiration: 300,
        passQuery: true,
        allowQuery: [
          'search_query', 'conflict_id', 'location_id',
          'cursor', 'date_from', 'date_to',
        ],
      },
    },
    // ISR для детальной страницы героя (решение техдолга #6)
    '/heroes/**': { isr: { expiration: 3600 } },
    // Справочники
    '/conflicts': { isr: 3600 },
    '/locations': { isr: 3600 },
    // Статические страницы
    '/': { prerender: true },
    '/contacts': { prerender: true },
    '/about': { prerender: true },
    // Никогда не индексировать
    '/admin/': { robots: false },
  },
},
```

### Ключевые отличия от предыдущего снимка

| Параметр | Было (2026-08-30) | Стало (2026-09-11) |
|---|---|---|
| `/` | `swr: 3600` | `prerender: true` |
| `/heroes`, `/heroes/` | `swr: 300` | `isr` с `allowQuery` (защита от cache poisoning) |
| `/heroes/**` | ❌ отсутствовало | ✅ `isr: { expiration: 3600 }` |
| `/conflicts`, `/locations` | `swr: 3600` | `isr: 3600` |
| `passQuery` / `allowQuery` | ❌ | ✅ |

> ⚠️ **Решён техдолг #6:** `/heroes/` не покрывал `/heroes/<id>`. Добавлено правило `/heroes/**` с `isr: 3600`.

### Invalidation ISR-кэша

```
Правка в админке → useCachePurge().purgeHero(id)
  ↓
POST /api/cache/purge → useStorage('cache').removeItem()
  ↓
Следующий запрос → свежий рендер
```

**Точки вызова (обновлено в Спринте 8):**
- `HeroForm.save()` — при `create` и `edit`
- `admin/heroes/[id].vue.onChanged()` — после `refresh()` для M:N-панелей

### Обработка ошибок в интерцепторе (`connect.ts`) — без изменений

### Хелперы для обработки ошибок (`errors.ts`) — без изменений

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
 │   │       ├── HeroForm.vue                # +useCachePurge, +useExtractionLinks, +parseExtractedDate, +fromJson
 │   │       ├── HeroPhotos.vue              # пагинация через useHeroPhotos
 │   │       ├── HeroAwards.vue
 │   │       ├── HeroConflicts.vue
 │   │       ├── ExtractionPanel.vue
 │   │       ├── ExtractionResultPreview.vue
 │   │       ├── SubmissionReviewHistory.vue
 │   │       ├── FaceBoxEditor.vue
 │   │       ├── FlexibleDateInput.vue
 │   │       ├── AdminSidebarContent.vue
 │   │       └── ... (ImageUpload, LocationMap, AdminFlexibleDateInput)
 │   ├── composables/
 │   │   ├── useApi.ts
 │   │   ├── useAuth.ts
 │   │   ├── useContactForm.ts
 │   │   ├── useCountUp.ts
 │   │   ├── useHeroRegistry.ts
 │   │   ├── useHeroPhotos.ts
 │   │   ├── useCachePurge.ts
 │   │   ├── useExtractionLinks.ts
 │   │   ├── useThemeSync.ts               # 🆕 Спринт 8: синхронизация color-mode + PrimeVue
 │   │   └── useDuplicateCheck.ts
 │   ├── constants/auth.ts
 │   ├── layouts/
 │   │   ├── default.vue
 │   │   └── admin.vue
 │   ├── lib/
 │   │   ├── api.ts
 │   │   ├── pb.ts
 │   │   ├── format.ts
 │   │   └── errors.ts
 │   ├── middleware/admin.ts
 │   ├── pages/
 │   │   ├── index.vue
 │   │   ├── about.vue
 │   │   ├── contacts.vue
 │   │   ├── submit.vue
 │   │   ├── heroes/[id].vue              # 🔄 Обновлён в Спринте 8
 │   │   └── admin/
 │   │       ├── heroes/
 │   │       │   ├── index.vue
 │   │       │   ├── new.vue
 │   │       │   └── [id].vue             # 🔄 +useCachePurge
 │   │       ├── submissions/
 │   │       │   └── index.vue
 │   │       ├── llm/
 │   │       │   └── index.vue
 │   │       └── ... (keys, awards, conflicts, locations, login)
 │   ├── plugins/
 │   │   ├── connect.ts
 │   │   └── reveal.ts
 │   └── sdk/emh/v1/                       # Сгенерированный код (не править!)
 │       └── ...
 ├── public/
 │   ├── fonts/                            # 🆕 Спринт 8: локальные шрифты
 │   │   ├── GolosText-Regular.woff2
 │   │   ├── PlayfairDisplay-Italic.woff2
 │   │   └── PlayfairDisplay-Regular.woff2
 │   ├── logo_favicon.svg
 │   ├── logo_gor_x190.png
 │   ├── logo_v5_full_gor_rwb.svg
 │   ├── logo_v5_full_gor_rwb.webp
 │   ├── logo_v5_full_gor_wrb.svg
 │   ├── logo_v5_full_gor_wrb.webp
 │   ├── logo_v5_full_vert_wbr.svg
 │   ├── logo_v5_full_X6.svg
 │   └── opensearch.xml                   # 🆕 Спринт 8: OpenSearch
 ├── server/
 │   └── api/cache/
 │       └── purge.post.ts
 └── nuxt.config.ts
```

---

## 4. Production Docker-образ (ISR) — без изменений

Multi-stage build: `node:22.23.1-alpine` builder → `node:22.23.1-alpine` runtime. Порт 3000. `Podmanfile.front`, `emh-frontend.container`, `Caddyfile` соответствуют целевому состоянию.

---

## 5. Взаимодействие с бэкендом и типизация

Протокол: Connect Protocol (HTTP/1.1 + JSON, `useBinaryFormat: false`).

Хелперы:
- `toPlain` (извлекает строгий тип `XxxJson`)
- `toEnum` (преобразование строки в числовой enum)
- `fromJson` из `@bufbuild/protobuf` — для конвертации `FlexibleDateJson` → `FlexibleDate` (используется в `HeroForm.prepareDateForApi`)

Enum: в ответах — строки, для записи — `toEnum(Schema, strValue)`.

### Реестр клиентов (`api.ts`) — без изменений

### Изменения в компонентах (Спринт 8)

#### `HeroForm.vue`
- Добавлен `useCachePurge` — вызов `purgeHero(id)` после `createHero` и `updateHero`
- Добавлен `useExtractionLinks` — авто-привязка связей после `CreateHero`
- `prepareDateForApi` использует `fromJson(FlexibleDateSchema, json)` для конвертации в protobuf Message
- `parseExtractedDate` — парсинг дат из LLM-извлечения (`YYYY`, `YYYY-MM`, `YYYY-MM-DD`)

#### `admin/heroes/[id].vue`
- Добавлен `useCachePurge` — вызов `purgeHero(id)` в `onChanged()` после `refresh()`

#### `pages/heroes/[id].vue`
- Убран `useIntersectionObserver` и сентинел-элемент (эксперимент Спринта 8, откатан)
- `twitterCard` остался в `useSeoMeta` на уровне страницы

---

## 6. Конфигурация оптимизации фронтенда (Спринт 8)

### 6.1 Локальные шрифты (`@nuxt/fonts`)

```typescript
fonts: {
  provider: 'local',
  families: [
    {
      name: 'Playfair Display',
      weights: [500, 600, 700, 800],
      styles: ['normal', 'italic'],
      subsets: ['cyrillic', 'latin'],
    },
    {
      name: 'Golos Text',
      weights: [400, 500, 600, 700],
      subsets: ['cyrillic', 'latin'],
    },
  ],
},
```

Шрифты лежат в `public/fonts/*.woff2` (конвертированы из `.ttf`). Google Fonts полностью удалены.

### 6.2 Темизация (`@nuxtjs/color-mode` + PrimeVue)

```typescript
colorMode: {
  preference: 'system',
  fallback: 'light',
  classSuffix: '-mode', // Генерирует: light-mode, dark-mode
},

primevue: {
  options: {
    theme: { options: { darkModeSelector: '.dark-mode' } },
  },
},
```

Синхронизация через `useThemeSync.ts` в `app.vue`:

```typescript
export function useThemeSync() {
  const colorMode = useColorMode();
  const syncTheme = (newMode: string) => {
    if (!import.meta.client) return;
    if (newMode === 'dark') {
      document.documentElement.classList.add('dark-mode');
      document.documentElement.classList.remove('light-mode');
    } else {
      document.documentElement.classList.add('light-mode');
      document.documentElement.classList.remove('dark-mode');
    }
  };
  watch(() => colorMode.value, syncTheme, { immediate: true });
}
```

### 6.3 HTML Validator

```typescript
htmlValidator: {
  enabled: process.env.NODE_ENV !== 'production' || process.env.VALIDATE_HTML === 'true',
  failOnError: process.env.CI === 'true',
  ignore: [/^\/admin/],  // Админка исключена
  options: {
    rules: {
      'meta-refresh': 'off',
      'element-case': 'off',
      'element-permitted-content': 'off',
      'no-redundant-role': 'off',
      'wcag/h63': 'off',
      'no-missing-references': 'off',
    },
  },
},
```

### 6.4 Доступность (a11y)

- `eslint-plugin-vuejs-accessibility` установлен как devDependency
- Контрастность: `#b08d57` → `#7a5c2e` (глобальная замена, решена ошибка WCAG AA)

### 6.5 Производительность сборки

```typescript
vite: {
  css: { transformer: 'lightningcss' },
},
```

### 6.6 OpenSearch

Файл `public/opensearch.xml` + ссылка в `app.head.link`. Позволяет добавить поиск героев в браузер.

### 6.7 Защита OG-картинок

```typescript
ogImage: {
  security: {
    strict: !!process.env.NUXT_OG_IMAGE_SECRET,
    secret: process.env.NUXT_OG_IMAGE_SECRET,
  },
},
```

### 6.8 Убрано из `seoMeta`

- `twitterCard: 'summary_large_image'` — удалён из глобального `app.seoMeta` (deprecated по рекомендации unhead)
- На странице героя (`[id].vue`) `twitterCard` остался в `useSeoMeta` на уровне страницы

### 6.9 Прочее

- `devtools.enabled`: `true` → `process.env.NODE_ENV === 'development'`
- `router.options.scrollBehaviorType: 'smooth'`
- `vite.optimizeDeps.include`: добавлены `vue-advanced-cropper` и модули `ol/*`

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
| Локальные шрифты | ✅ Завершено (Спринт 8, `@nuxt/fonts` + woff2) |
| Переход с SSG на ISR | ✅ Завершено |
| ISR с `allowQuery` (защита от cache poisoning) | ✅ Завершено (Спринт 8) |
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
| **Локальные шрифты вместо Google Fonts** | ✅ Завершено (Спринт 8) |
| **`@nuxtjs/color-mode` + синхронизация с PrimeVue** | ✅ Завершено (Спринт 8) |
| **`@nuxtjs/html-validator`** | ✅ Завершено (Спринт 8) |
| **`@vueuse/nuxt`** | ✅ Завершено (Спринт 8) |
| **`eslint-plugin-vuejs-accessibility`** | ✅ Завершено (Спринт 8) |
| **Контрастность `#b08d57` → `#7a5c2e`** | ✅ Завершено (Спринт 8) |
| **Убран `twitterCard` из глобального `seoMeta`** | ✅ Завершено (Спринт 8) |
| **OpenSearch** | ✅ Завершено (Спринт 8) |
| **`ogImage.security` (HMAC)** | ✅ Завершено (Спринт 8) |
| **LightningCSS** | ✅ Завершено (Спринт 8) |
| **`scrollBehaviorType: 'smooth'`** | ✅ Завершено (Спринт 8) |
| **`/heroes/**` в `routeRules`** | ✅ Завершено (Спринт 8) |
| **Ленивая загрузка изображений галереи** | ⚠️ Откатано (см. техдолг #9) |

---

## 10. Техдолг и известные ограничения

| # | Файл / область | Проблема | Действие |
|---|---|---|---|
| 1 | `HeroForm.vue` | `fieldMask` захардкожен | ✅ Решено |
| 2 | Thumbnail-генерация | Асинхронная (1–5 сек), `thumbnail_url` может быть пустым | Использовать `url` как fallback + polling |
| 3 | Дубли `robots`-метатегов | Глобальный `seoMeta` + локальный `useSeoMeta` | Проверить приоритет после деплоя |
| 4 | ISR-кэш | Жил до 24 ч после правки | ✅ Решено через `/api/cache/purge` |
| 5 | `@nuxt/fonts` | Body Timeout Error при сборке без интернета | ✅ Решено — модуль возвращён с `provider: 'local'`, шрифты в `public/fonts/*.woff2` |
| 6 | `routeRules` | `/heroes/` не покрывал `/heroes/<id>` | ✅ Решено — добавлено `/heroes/**` |
| 7 | `HeroPhotos.vue` | `move()` / `setMain()` использовали `props.photos` | ✅ Решено |
| 8 | `processing_time_ms` в `TestLlmProviderResponse` | Тип `int64`, в JSON приходит как `string` | ✅ Решено — `Number()` |
| 9 | `[id].vue` (public) | `loading="lazy"` и `decoding="async"` на превью галереи были добавлены, но **откатаны** | 🔴 **Открыт**: вернуть `loading="lazy"` + `decoding="async"` на `img.galleria__thumbnail` в `[id].vue` |
| 10 | `HeroPhotos.vue` (admin) | `loading="lazy"` на `<Image>` был добавлен, но **откатан** | 🔴 **Открыт**: вернуть `loading="lazy"` на `<Image>` в `HeroPhotos.vue` |
| 11 | `[id].vue` (public) | `alt` у превью галереи = `image.description` без fallback | 🟡 Рекомендация: `image.description \|\| \`Фотография ${fullName}\`` |
| 12 | `robots.txt` | Директивы `Content-Usage` и `Content-Signal` не являются стандартом | 🟡 Известно: `@nuxtjs/robots` генерирует их как есть, валидатор Lighthouse помечает как `Unknown directive` |
| 13 | `twitterCard` | Убран из глобального `seoMeta`, но остался в `useSeoMeta` страницы `[id].vue` | 🟢 Низкий приоритет: `twitterCard` deprecated, но пока работает |
| 14 | LCP на главной | 3.6 сек (цель < 2.5 сек). Причины: шрифты, логотип без `fetchpriority="high"` | 🟡 **Открыт**: добавить `fetchpriority="high"` и `preload` для логотипа на главной |
| 15 | Изображения на главной | Нет `width`/`height` у изображений карточек героев | 🟡 **Открыт**: добавить размеры для уменьшения CLS |

---

## 11. Следующие шаги (приоритеты для новой сессии)

### P0 — Критические

| # | Задача | Описание |
|---|---|---|
| P0.1 | Валидация Спринта 8 на проде | Проверить работу `@nuxt/fonts`, `colorMode`, `htmlValidator`, ISR с `allowQuery` |
| P0.2 | Вернуть `loading="lazy"` | Техдолг #9, #10: добавить `loading="lazy"` на изображения галереи в `[id].vue` и `HeroPhotos.vue` |

### P1 — Высокий приоритет

| # | Задача | Описание |
|---|---|---|
| P1.1 | LCP-оптимизация главной | Техдолг #14: `fetchpriority="high"` на логотип, `preload` в `<head>`, `width`/`height` на изображения |
| P1.2 | Проверить дубли `robots`-метатегов | `curl -s https://вежливые.рус/heroes/<id> \| grep robots` |

### P2 — Средний приоритет

| # | Задача | Описание |
|---|---|---|
| P2.1 | Рефакторинг `<NuxtLink><Button/></NuxtLink>` → `Button as="router-link"` | Семантика HTML + a11y |
| P2.2 | Мониторинг через метрики | Визуализация Prometheus метрик из бэкенда |
| P2.3 | Повышение покрытия тестами | `internal/smtp` (42.4%) и `internal/usecase` (60.5%) |

### P3 — Отложено

| # | Задача |
|---|---|
| P3.1 | Множественные face box (групповые фото) |
| P3.2 | Кэширование `ListHeroes` на уровне Nitro |
| P3.3 | Service Worker для оффлайн-режима админки |
| P3.4 | Гибкие даты для всех сущностей (конфликты, локации) |
| P3.5 | Grafana-дашборд |
| P3.6 | PWA manifest |

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
| Синхронизация темы (color-mode + PrimeVue) | `app/composables/useThemeSync.ts` + `app.vue` |
| Парсинг дат из извлечения | `app/components/admin/HeroForm.vue` → `parseExtractedDate` |
| Конвертация `FlexibleDateJson` → protobuf | `app/components/admin/HeroForm.vue` → `prepareDateForApi` + `fromJson` |

---

## 13. Changelog сессий

### Спринт 8 (2026-09-11) — Оптимизация фронтенда

#### ✅ Завершено

**Производительность и приватность:**
- `@nuxt/fonts` с `provider: 'local'` — замена Google Fonts на локальные `.woff2`
- Конвертация шрифтов из `.ttf` → `.woff2` (экономия ~500 КБ)
- `vite.css.transformer: 'lightningcss'`
- ISR `routeRules` с `allowQuery` и `passQuery` (защита от cache poisoning)
- `/heroes/**` добавлено в `routeRules` (решение техдолга #6)
- `scrollBehaviorType: 'smooth'`

**Доступность (a11y):**
- `eslint-plugin-vuejs-accessibility` установлен
- Контрастность `#b08d57` → `#7a5c2e` (WCAG AA 4.5:1)

**Темизация:**
- `@nuxtjs/color-mode` с `classSuffix: '-mode'`
- `useThemeSync.ts` — синхронизация с PrimeVue (`darkModeSelector: '.dark-mode'`)
- `useThemeSync()` вызывается в `app.vue`

**Качество кода:**
- `@nuxtjs/html-validator` с исключением админки (`ignore: [/^\/admin/]`) и отключением ложных срабатываний PrimeVue
- `@vueuse/nuxt` добавлен в модули

**Безопасность и SEO:**
- `ogImage.security` с HMAC-подписью
- `twitterCard` убран из глобального `seoMeta`
- OpenSearch (`public/opensearch.xml` + `app.head.link`)
- `devtools.enabled` зависит от `NODE_ENV`

**Инфраструктура:**
- `pnpm` обновлён: `11.21.0` → `12.3.4`
- Версия проекта: `0.0.3` → `0.0.4`

#### ⚠️ Откатано (зафиксировано в техдолге)
- `useIntersectionObserver` + сентинел для автоподгрузки галереи в `[id].vue` — эксперимент не прижился, откатан
- `loading="lazy"` и `decoding="async"` на превью галереи — откатаны (техдолг #9, #10)

---

### Спринт 7 (2026-08-30) — Без изменений (см. предыдущий снимок)

---

## 14. Примечания для следующего агента

1. **Не подключай `@nuxt/fonts` повторно** — он уже настроен с `provider: 'local'`, шрифты в `public/fonts/*.woff2`.
2. **Не используй `@nuxt/a11y`** — модуль нестабилен для Nuxt 3. Используй `eslint-plugin-vuejs-accessibility` вместо него.
3. **`htmlValidator.ignore`** — именно `ignore`, а не `exclude`. Свойство `exclude` не существует в `@nuxtjs/html-validator`.
4. **`twitterCard`** на странице `[id].vue` в `useSeoMeta` — оставлен намеренно. Убирать только если будет принято решение полностью отказаться от Twitter Cards.
5. **Директивы `Content-Usage` / `Content-Signal` в `robots.txt`** — не являются стандартом. Lighthouse помечает как `Unknown directive`. Если это неприемлемо, убрать из конфигурации `@nuxtjs/robots`.
6. **`useThemeSync.ts`** — использует `import.meta.client` для защиты от обращения к `document` на сервере. Без этой проверки будет `Cannot read properties of undefined (reading 'documentElement')`.
7. **`allowQuery` в `routeRules`** — критичен. Без него любой бот может сгенерировать миллионы уникальных URL и заполнить кэш (cache poisoning). Список параметров должен строго соответствовать полям `ListHeroesRequest` из `hero.proto`.
