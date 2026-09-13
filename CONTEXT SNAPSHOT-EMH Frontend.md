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

### Раздел 5.1 — «Изменения в компонентах (Спринт 9)»

**`pages/heroes/[id].vue`**
- ✅ `loading="lazy"` и `decoding="async"` возвращены на `img.galleria__thumbnail` (закрытие техдолга #9)
- ✅ `alt` с fallback: `image.description || \`Фотография ${fullName}\`` (закрытие техдолга #11)
- ✅ `twitterCard` и все `twitter:*` метатеги полностью удалены из `useSeoMeta` (закрытие техдолга #13)
- ✅ Галерея работает через `IntersectionObserver` + сентинел (автоподгрузка), кнопка «Загрузить ещё» — fallback
- ⚠️ INP ~271ms — требует оптимизации в будущем спринте
- ✅ Оптимизация INP: `trackPhotoView` вынесен в `requestIdleCallback` (с fallback для Safari)
- ✅ Предзагрузка полного фото по `pointerenter` на превью (`preloadFullImage`)
- ✅ `content-visibility: auto` на `.hero__section` + `contain-intrinsic-size`
- ✅ `content-visibility: visible` на `.hero__opening` (first screen)
- ✅ `isolation: isolate; contain: layout paint` на `.p-galleria-mask`

**`layouts/default.vue`**
- ✅ Добавлен `preload` логотипа с `fetchpriority="high"` (частичное закрытие техдолга #14)
- ✅ Логотип имеет `fetchpriority="high"` в `<img>`

**`pages/index.vue`**
- ✅ Изображения карточек героев используют `loading="lazy"`
- ✅ CLS предотвращён через `aspect-ratio: 4 / 5` на `.hero-card__photo` (закрытие техдолга #15)


### Рефакторинг ссылок-кнопок (P2.1)

Паттерн `<NuxtLink><Button/></NuxtLink>` заменён на один семантический интерактивный элемент:

```vue
<Button as="router-link" :to="..." />
```

Изменения внесены в файлы:

- `app/components/admin/HeroForm.vue` — кнопка «Отмена»
- `app/pages/submit.vue` — кнопка «Вернуться к реестру»
- `app/pages/admin/index.vue` — «Новый герой», «Реестр героев», «Модерация заявок», «API-ключи»
- `app/pages/admin/heroes/[id].vue` — «Открыть на сайте»
- `app/pages/admin/heroes/index.vue` — «+ Новый герой», «Изменить»
- `app/pages/admin/submissions/index.vue` — «Открыть героя»

Результат:

- устранены вложенные интерактивные элементы `<a><button>`;
- улучшена HTML-семантика;
- улучшена доступность для клавиатуры и скринридеров;
- сохранены визуальные стили, иконки и поведение навигации.

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
| Ленивая загрузка изображений галереи (публичная + админка) | ✅ Завершено (Спринт 9) |
| Fallback для `alt` в галерее (`Фотография ${fullName}`) | ✅ Завершено (Спринт 9) |
| Удаление `twitterCard` / `twitter:*` из всех страниц | ✅ Завершено (Спринт 9) |
| CLS на главной (`aspect-ratio: 4/5` для карточек) | ✅ Завершено (Спринт 9) |
| `preload` + `fetchpriority="high"` для логотипа | ✅ Завершено (Спринт 9) |
| Рефакторинг ссылок-кнопок (`Button as="router-link"`) | ✅ Завершено (Спринт 9) |

---

## 10. Техдолг и известные ограничения

| # | Файл / область | Проблема | Действие |
|---|---|---|---|
| 9 | `[id].vue` (public) | `loading="lazy"` и `decoding="async"` были откатаны | ✅ Решено (Спринт 9) — возвращены, галерея работает корректно |
| 10 | `HeroPhotos.vue` (admin) | `loading="lazy"` был откатан | ✅ Решено (Спринт 9) — возвращён, PrimeVue `preview` работает |
| 11 | `[id].vue` (public) | `alt` без fallback | ✅ Решено (Спринт 9) — fallback `Фотография ${fullName}` |
| 13 | `twitterCard` | Оставался в `useSeoMeta` страницы `[id].vue` | ✅ Решено (Спринт 9) — полностью удалён |
| 14 | LCP на главной | 3.6 сек, логотип без приоритета | 🟡 Частично закрыт: `preload` + `fetchpriority="high"` добавлены |
| 15 | Изображения на главной | CLS из-за отсутствия размеров | ✅ Решено (Спринт 9) — `aspect-ratio: 4 / 5` |
| 3 | Дубли `robots`-метатегов | Глобальный `seoMeta` + локальный `useSeoMeta` | 🔵 Отложено — требует глубокого аудита `nuxt-seo-utils` |
| 16 | INP ~271ms | ✅ **Оптимизирован** — TBT упал с 20ms до 3ms. Требуется контрольный замер INP с реальным взаимодействием. |
| 17 | `[id].vue` (public) | `Numeric tagPriority (35)` в unhead | 🟡 Низкий приоритет — заменить на алиас `critical` |
| 18 | `[id].vue` (public) | Inline `<script>` 3.0KB (вероятно, Яндекс.Метрика или OG-secret) | 🟡 Низкий приоритет — вынести во внешний файл |
| 19 | `<NuxtLink><Button/>` паттерн | Вложенные интерактивные элементы `<a><button>` нарушали семантику HTML и ухудшали a11y | ✅ Решено (Спринт 9) — заменено на `Button as="router-link"` |
| **NEW** 20 | Image Delivery (score 0) | 🔴 Открыт — добавить `<picture>` с WebP/AVIF и responsive sizes |
| **NEW** 21 | Legacy JavaScript (133.8ms) | 🟡 Средний приоритет — поднять `vite.build.target` до `es2022` |
| **NEW** 22 | `/submit` | PrimeIcons SVG 347 КБ — загружается весь набор иконок | 🔴 Открыт — перейти на шрифтовую версию или tree-shaking |
| **NEW** 23 | `/submit` | JS-бандлы ~640 КБ | 🟡 Средний приоритет — включить code splitting |
| **NEW** 24 | `contacts.vue` | Контрастность ссылки в `.contact-form-head__text` — 1.25:1 | 🔴 Открыт — добавить `text-decoration: underline` и `color: var(--emh-crimson)` |
| **NEW** 25 | `contacts.vue` | Избыточные `v-reveal` на первом экране | 🟡 Средний — убрать с первого экрана, оставить только ниже fold |
| **NEW** 26 | `/contacts`, `/submit` | Нет `preconnect` к S3 и `preload` шрифтов | 🟡 Средний — добавить в `app.vue` |
| **NEW** 27 | `/about`, `/contacts`, `/submit` | Нет `preconnect` к S3 (`s3.neverforgotten.ru`) | 🔴 Открыт — добавить `preconnect` для ускорения загрузки шрифтов и изображений |
| **NEW** 28 | `/about` | Unused CSS 68 КиБ | 🟡 Низкий приоритет — микрооптимизация |
| **NEW** 29 | `/` (mobile) | TTI = 15.2 сек на мобилке | 🔴 Критично — применить code splitting, lazy loading изображений, убрать PrimeIcons SVG |
| **NEW** 30 | `/` (mobile) | TBT = 431 мс | 🔴 Критично — связано с тяжёлым JS бандлом и отсутствием code splitting |

---

### Раздел 11. Следующие шаги (приоритеты для новой сессии)

#### P0 — Критические

| # | Задача | Описание |
|---|---|---|
| P0.1 | Валидация Спринта 8 на проде | Проверить работу `@nuxt/fonts`, `colorMode`, `htmlValidator`, ISR с `allowQuery` |

> ⚠️ P0.2 закрыт в Спринте 9 — `loading="lazy"` возвращён, alt с fallback добавлен.

#### P1 — Высокий приоритет

| # | Задача | Описание |
|---|---|---|
| P1.1 | LCP-оптимизация главной (довести до < 2.5s) | `preload` и `fetchpriority` уже есть. Осталось: проверить Lighthouse, при необходимости вынести критические CSS |

> ⚠️ P1.2 (дубли robots) перенесён в P3.

#### P2 — Средний приоритет (обновлён)

| # | Задача | Описание |
|---|---|---|
| P2.1 | **Оптимизация INP в галерее** | Техдолг #16: debounce `imageClick`, убрать синхронные операции |
| P2.2 | Замена `Numeric tagPriority (35)` на алиас | Техдолг #17: unhead предупреждение |
| P2.3 | Мониторинг через метрики | Визуализация Prometheus метрик из бэкенда |
| P2.4 | Повышение покрытия тестами | `internal/smtp` (42.4%) и `internal/usecase` (60.5%) |

#### P3 — Отложено (обновлён)

| # | Задача |
|---|---|
| P3.1 | Множественные face box (групповые фото) |
| P3.2 | Кэширование `ListHeroes` на уровне Nitro |
| P3.3 | Service Worker для оффлайн-режима админки |
| P3.4 | Гибкие даты для всех сущностей (конфликты, локации) |
| P3.5 | Grafana-дашборд |
| P3.6 | PWA manifest |
| **NEW** P3.7 | Дубли `robots`-метатегов (требует аудита `nuxt-seo-utils`) |
| **NEW** P3.8 | Inline `<script>` 3.0KB — вынести во внешний файл |

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
| Конвертация `FlexibleDateJson` → protobuf | `app/components/admin/HeroForm.vue` → `prepareDateForApi` + `fromJson` ||---|---|
| SEO карточки героя (без Twitter) | `app/pages/heroes/[id].vue` |
| Логотип с preload + fetchpriority | `app/layouts/default.vue` |
| CLS-safe карточки на главной | `app/pages/index.vue` (`.hero-card__photo { aspect-ratio: 4/5 }`) |
| Галерея с lazy + alt fallback | `app/pages/heroes/[id].vue` (`img.galleria__thumbnail`) |
| Семантическая ссылка-кнопка | `app/pages/admin/index.vue` (`Button as="router-link"`) |

---

## 13. Changelog сессий

## Спринт 9 (2026-09-12) — Закрытие техдолга по lazy-loading и LCP

### ✅ Завершено

**Производительность (Core Web Vitals):**
- `loading="lazy"` и `decoding="async"` возвращены на `img.galleria__thumbnail` в `[id].vue` (техдолг #9)
- `loading="lazy"` возвращён на `<Image>` в `HeroPhotos.vue` (техдолг #10)
- `alt` с fallback `Фотография ${fullName}` в публичной галерее (техдолг #11)
- `aspect-ratio: 4 / 5` на `.hero-card__photo` в `index.vue` — устранён CLS (техдолг #15)
- `preload` + `fetchpriority="high"` для логотипа в `default.vue` (частично техдолг #14)

**SEO:**
- `twitterCard` и все `twitter:*` метатеги полностью удалены из `[id].vue` (техдолг #13)
- Open Graph полностью покрывает потребности всех соцсетей (включая Telegram/X)

**Регрессионное тестирование:**
- PrimeVue `<Image preview>` работает корректно с `loading="lazy"` в админке
- Galleria в `[id].vue` открывается по клику на превью (lazy не ломает preview)
- Главное фото героя (`<figure class="hero__portrait">`) загружается без `loading="lazy"` — LCP не страдает

**Качество кода и a11y:**
- Закрыт P2.1: паттерн `<NuxtLink><Button/></NuxtLink>` заменён на `Button as="router-link"`
- Устранены вложенные интерактивные элементы `<a><button>`
- Улучшена семантика ссылок-кнопок в админке и публичной части
- Навигация, стили, иконки и открытие в новой вкладке работают корректно

Оптимизация интерактивности (техдолг #16):
- `trackPhotoView` в `imageClick` перенесён в `requestIdleCallback` с fallback для Safari
- Предзагрузка полноразмерного фото при наведении на превью (`pointerenter` + `new Image()`)
- `content-visibility: auto` на секциях ниже первого экрана
- `isolation: isolate` + `contain: layout paint` на маске PrimeVue Galleria

Результаты оптимизации INP (контрольный замер):
- Total Blocking Time: 20ms → 3ms (-85%)
- Speed Index: улучшен на ~15%
- CLS: стабильно 0
- Image Delivery: требует доработки (score 0)
- Legacy JavaScript: 133.8ms (требует поднятия Vite target)

Результаты Lighthouse для /submit (2026-09-12):
- Performance: 80/100
- Accessibility: ~90/100
- Best Practices: 100/100
- SEO: ~95/100

Обнаруженные проблемы:
- PrimeIcons SVG: 347 КБ (требует перехода на шрифт или tree-shaking)
- JS-бандлы: ~640 КБ (требует code splitting)
- Total Blocking Time: 87.5ms (не критично)

Результаты Lighthouse для /contacts (2026-09-12):
- Performance: 66/100
- FCP: 1.9s, TTI: 4.6s — требуют оптимизации
- Accessibility: проблема контрастности ссылки в форме
- Best Practices: 100/100
- SEO: ок

Обнаруженные проблемы:
- `v-reveal` directive на первом экране замедляет FCP/TTI
- Ссылка на `/submit` в `.contact-form-head__text` имеет контраст 1.25:1

Результаты Lighthouse для /about (2026-09-12):
- Performance: ~85/100
- FCP: 1.7s, LCP: ~2.0s
- TBT: 50ms (отлично)
- CLS: 0 (отлично)
- Network Dependency Tree: score 0 (требует preconnect к S3)
- Unused CSS: 68 КиБ (микрооптимизация)

Результаты Lighthouse для главной (mobile, 2026-09-12):
- Performance: ~45/100 (критично)
- TTI: 15.2 сек (цель < 3.8 сек)
- TBT: 431 мс (цель < 200 мс)
- Image Delivery: score 0.5, экономия 315 КиБ

Требуется срочная оптимизация:
- Code splitting для уменьшения основного бандла
- Переход с PrimeIcons SVG на woff2
- Lazy loading изображений на главной
- Оптимизация размера превью (backend)

### 🟡 Перенесено в отложенные (P3)
- Дубли `robots`-метатегов (техдолг #3) — требует аудита `nuxt-seo-utils`
- `width`/`height` для изображений на главной — заменено на `aspect-ratio` (более гибкое решение)

### 🔴 Обнаружено (новый техдолг)
- INP ~271ms на странице героя (`needs-improvement`)
- `Numeric tagPriority (35)` в unhead (предупреждение)
- Inline `<script>` 3.0KB (предупреждение unhead)

---

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
5. **Директивы `Content-Usage` / `Content-Signal` в `robots.txt`** — не являются стандартом. Lighthouse помечает как `Unknown directive`. Если это неприемлемо, убрать из конфигурации `@nuxtjs/robots`.
6. **`useThemeSync.ts`** — использует `import.meta.client` для защиты от обращения к `document` на сервере. Без этой проверки будет `Cannot read properties of undefined (reading 'documentElement')`.
7. **`allowQuery` в `routeRules`** — критичен. Без него любой бот может сгенерировать миллионы уникальных URL и заполнить кэш (cache poisoning). Список параметров должен строго соответствовать полям `ListHeroesRequest` из `hero.proto`.
- ✅ `twitterCard` и все `twitter:*` метатеги **полностью удалены** в Спринте 9. Open Graph полностью заменяет их функциональность.
- ✅ INP на странице героя требует внимания: `imageClick` в `[id].vue` выполняется за ~271ms. Если INP останется высоким, применить `useDebounceFn` из `@vueuse/nuxt` или `requestIdleCallback` для тяжёлых операций.
- ✅ `aspect-ratio: 4 / 5` в `index.vue` — это замена `width`/`height`. Если дизайн изменится и карточки станут другого соотношения сторон, обновить именно этот CSS-класс.
- ✅ Не используй паттерн `<NuxtLink><Button/></NuxtLink>`. Для ссылок, выглядящих как кнопки, используй:

```vue
<Button as="router-link" :to="..." label="..." />
```

Это корректно с точки зрения HTML, клавиатурной навигации и скринридеров.

После полного отказа от вложенных `<a><button>` можно рассмотреть повторное включение правила `element-permitted-content` в `htmlValidator`, но только после отдельной проверки всех страниц на аналогичные паттерны.
