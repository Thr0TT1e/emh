# CONTEXT SNAPSHOT — EMH Frontend (Спринт 9, финальная)

Технический паспорт проекта «Вечная память героям»

---

## 1. Метаданные

| Параметр | Значение |
|---|---|
| Проект | Вечная память героям (EMH) |
| Scope | Только `frontend/` (Nuxt 4) |
| Фаза | Спринт 10 завершен. Code splitting `/submit` завершен, ISR временно отключен для отладки прода, проблема с `robots.txt` решена |
| Дата снимка | 2026-09-16 |
| Production VPS | 82.202.139.132 |
| Канонический домен | вежливые.рус (IDN, кириллица) |
| Зеркала (301 → канонический) | вечнаяпамятьгероям.рус, emh.su, neverforgotten.ru |
| Технические поддомены | api.вежливые.рус → бэкенд, s3.вежливые.рус → MinIO |

---

## 2. Стек и зависимости

### Runtime

| Компонент | Версия | Назначение |
| --- | --- | --- |
| Nuxt | ^4.5.2 | Фреймворк (структура `app/`) |
| Vue | ^3.5.47 | UI |
| PrimeVue | 4.5.5 (зафиксирована) | UI-кит + `@primeuix/themes` |
| unplugin-icons | ^24.0.0 | Tree-shaking иконок (полная замена PrimeIcons SVG) |
| @iconify-json/carbon, mdi | ^1.2.x | Наборы иконок для unplugin-icons |
| @bufbuild/protobuf | ^2.16.0 | protobuf-es v2 (рантайм) |
| @connectrpc/connect-web | ^2.2.1 | Connect RPC транспорт |
| vue-advanced-cropper | ^2.8.9 | Face Box UI |
| OpenLayers (ol) | ^10.10.0 | Карты (локации героев) |
| Node.js | 22.23.1 | Runtime |
| Пакетный менеджер | pnpm@12.4.1 | type: module |

### SEO и модули

| Модуль | Назначение |
| --- | --- |
| nuxt-seo-utils | Глобальный site config |
| @nuxtjs/sitemap | Динамический sitemap |
| @nuxtjs/robots | Автогенерация `robots.txt` |
| nuxt-schema-org | JSON-LD (Person, BreadcrumbList) |
| nuxt-og-image | OG-карточки через Satori |
| @nuxt/scripts | Ленивая загрузка Яндекс.Метрики (счётчик 111894071) |
| @nuxt/fonts | Локальные шрифты (`provider: 'local'`, woff2) |
| @nuxtjs/html-validator | Валидация HTML в dev/CI |
| @vueuse/nuxt | Composables |
| @nuxtjs/color-mode | Dark/Light mode с `classSuffix: '-mode'` |
| eslint-plugin-vuejs-accessibility | A11y-правила для Vue-компонентов |

---

## 3. Текущее состояние

### ✅ Завершено (Спринт 10 (2026-09-15/16))

| # | Задача | Статус | Описание |
| --- | --- | --- | --- |
| 1 | Code splitting `/submit` | ✅ Завершено | `AttachmentUpload.vue` и `HeroSearchPicker.vue` вынесены в `defineAsyncComponent`. TBT 87.5ms → 0ms, TTI 2.5s → 1.7s |
| 2 | Замена PrimeIcons → unplugin-icons | ✅ Завершено | Полный отказ от SVG-бандла (347 КБ). Миграция на `~icons/carbon/...` и `~icons/mdi/...` с tree-shaking |
| 3 | Legacy JavaScript | ✅ Завершено | `vite.build.target` поднят до `es2024` |
| 4 | AI Policy | ✅ Завершено | Страница `/ai-policy` + `/.well-known/ai-policy.json` |
| 5 | Локальные шрифты | ✅ Завершено | TTF-файлы (Golos Text, Playfair Display) в `public/fonts/` |
| 6 | Очистка `public/_robots.txt` | ✅ Завершено | Удалена устаревшая директива `Content-Usage` |
| 7 | Image Delivery | ✅ Завершено | `width`/`height`/`sizes`/`fetchpriority` на всех изображениях. TBT -97%, CLS -69% (см. ADR-001) |
| 8 | Карта локаций на публичной странице | ✅ Завершено | `HeroLocationMap.vue` с OpenLayers. ESRI World Topo вместо OSM (заблокирован в РФ). Pre-flight проверка + graceful fallback (см. ADR-004) |
| 9 | Кэширование шрифтов и статики | ✅ Завершено | Caddy отдаёт `Cache-Control: public, max-age=31536000, immutable` для `/fonts/*` и `/_nuxt/*` |
| 10 | `font-display: swap` | ✅ Завершено | Явно указан в `@nuxt/fonts` конфигурации |

### 🔴 Критичные задачи (следующий спринт)

| # | Задача | Описание |
| --- | --- | --- |
| 1 | Мобильная производительность главной | TTI = 15.2 сек, TBT = 431 мс. Требуется: lazy loading изображений, оптимизация LCP |
| 2 | `preconnect` к S3 | Нет `preconnect` к `s3.neverforgotten.ru` на `/about`, `/contacts`, `/submit` |

### ⚠️ Плавающие проблемы (не трогать — зависят от окружения)
Следующие проблемы Lighthouse исключены из плана работ — они непостоянны и зависят от расширений браузера, нагрузки на VPS и состояния кэша:
- Проксирование Яндекс.Метрики (`mc.yandex.ru` — блокировки провайдерами)
- BFCache "Внутренняя ошибка" (`IgnoreEventAndEvict` — зависит от расширений Chrome)
- Unused JavaScript (~300-400 КБ) — общий chunk Nuxt, не связан с кодом приложения
- TBT 70мс — скачет между прогонами (12ms ↔ 70ms) в пределах нормы

### 🟡 Средний приоритет

| # | Задача |
| --- | --- |
| 1 | Дубли `robots`-метатегов — аудит `nuxt-seo-utils` |
| 2 | Numeric `tagPriority` (35) в unhead — заменить на алиас `critical` |
| 3 | Inline `<script>` 3.0KB — вынести во внешний файл |
| 4 | Контрастность ссылки в `contacts.vue` (1.25:1) — добавить `text-decoration: underline` |
| 5 | `v-reveal` на первом экране `contacts.vue` — убрать с первого экрана |

### 🔵 Функциональные задачи (приоритет после производительности)

| # | Задача | Компоненты/Контракты | Описание |
| --- | --- | --- | --- |
| 1 | Связи между героями | `HeroRelations` (hero.proto) | Публичный рендеринг блока «Связи» с переходами по `toHeroId`. Админская форма `HeroRelations.vue` |
| 2 | Источники героя | `HeroSource` (hero.proto) | Публичный рендеринг блока «Источники» с `sourceType`, `url`, `excerpt`. Админская форма `HeroSources.vue` |
| 3 | Краудсорсинг модерация | `SubmissionService` (hero_admin.proto) | Review workflow в админке: `SubmissionReviewHistory.vue`, статусы заявок |
| 4 | LLM-извлечение данных | `ExtractionService` (hero_admin.proto) | `ExtractionPanel.vue` → авто-заполнение форм, парсинг дат через `parseExtractedDate` |
| 5 | Конфликты героя | `HeroConflict` (hero.proto) | Привязка героя к конфликтам с `rankAtConflict`, `specificLocation`. `HeroConflicts.vue` |
| 6 | Награды с гибкими датами | `HeroAward` (hero.proto) | `HeroAwards.vue` + `AdminFlexibleDateInput.vue` для `award_date_info` |
| 7 | Фото с FaceBox | `Photo` (hero.proto) | Админская разметка `FaceBoxEditor.vue`, превью через `thumbnailUrl \|\| url` (см. ADR-001) |
| 8 | Управление LLM-провайдерами | `LlmAdminService` (hero_admin.proto) | `app/pages/admin/llm/index.vue` — выбор Ollama/OpenAI/Anthropic |

### ⚪ Отложено / Не делается

- Множественные face box (групповые фото)
- Кэширование `ListHeroes` на уровне Nitro
- Service Worker для оффлайн-режима админки
- Grafana-дашборд
- PWA manifest

---

## 4. Архитектурные решения (зафиксированы, не обсуждаются)

### Гибкие даты (FlexibleDate)

**Реализовано и работает:**
- `Hero` — `birth_date_info`, `death_date_info`, `service_start_date_info`
- `HeroAward` — `award_date_info`
- `Conflict` — `start_date_info`, `end_date_info`

**НЕ делаем и не планируем (не возвращаться):**
- `HeroConflict` — даты участия героя в конфликте
- `Location` — нет временных атрибутов у географии
- `Photo` — нет даты съёмки в контракте
- `HeroSource` — нет даты публикации источника

**Эталоны:** `AdminFlexibleDateInput.vue` + `prepareDateForApi()` из `HeroForm.vue`

### Конвертация дат

```typescript
const prepareDateForApi = (fd?: FlexibleDateJson | null, allowEmpty = false) => {
  if (!fd || fd.precision === 'DATE_PRECISION_UNSPECIFIED') {
    return allowEmpty
      ? undefined
      : fromJson(FlexibleDateSchema, { precision: 'DATE_PRECISION_UNKNOWN', displayText: '' });
  }
  return fromJson(FlexibleDateSchema, fd);
};
```

### Семантические ссылки-кнопки

```vue
<!-- ПРАВИЛЬНО -->
<Button as="router-link" :to="..." label="..." />

<!-- НЕПРАВИЛЬНО (не использовать) -->
<NuxtLink :to="..."><Button label="..." /></NuxtLink>
```

### ISR и защита от cache poisoning

```typescript
// nuxt.config.ts → nitro.routeRules

// ВРЕМЕННО ОТКЛЮЧЕНО ДЛЯ ОТЛАДКИ ПРОДА (15.09.2026)
// После стабилизации вернуть ISR с expiration и allowQuery.
'/heroes': { isr: false },
'/heroes/': { isr: false },
'/heroes/**': { isr: false },
'/conflicts': { isr: false },
'/locations': { isr: false },
```

**Критично:** `allowQuery` должен строго соответствовать полям `ListHeroesRequest` из `hero.proto`. Без него — cache poisoning.

### Robots.txt (замена Content-Usage / Content-Signal)

```ts
// nuxt.config.ts → robots.groups
{
  userAgent: ['GPTBot', 'OAI-SearchBot', 'ChatGPT-User', 'Google-Extended', ...],
  disallow: ['/'],
}
```

**Дополнительно:**
- `public/.well-known/ai-policy.json` — машиночитаемая политика
- `app/pages/ai-policy.vue` — человекочитаемая страница (`noindex`)
- Файл `public/_robots.txt` очищен от нестандартных директив (удален `Content-Usage`) и больше не переопределяет/не конфликтует с модулем `@nuxtjs/robots`.

Дополнительно:
- `public/.well-known/ai-policy.json` — машиночитаемая политика
- `app/pages/ai-policy.vue` — человекочитаемая страница (`noindex`)

### Иконки (unplugin-icons)

PrimeIcons полностью удален из-за отсутствия tree-shaking (бандл 347 КБ SVG загружался целиком).
Используется `unplugin-icons` + `unplugin-vue-components` с коллекциями `carbon` и `mdi`.

```vue
<!-- ✅ ПРАВИЛЬНО (unplugin-icons) -->
<IconCarbonSearch />
<Button>
    <IconCarbonSearch />
    <span>Добавить</span>
</Button>

<!-- ❌ НЕПРАВИЛЬНО (PrimeIcons — больше не работает) -->
<i class="pi pi-user" />
<Button icon="pi pi-plus" label="Добавить" />
```

### Убрано из плана навсегда

- Контрастность ссылок в карточке героя — не требуется (`#7a5c2e` = WCAG AA 4.5:1)
- Публичный API + OpenAPI — документация генерируется из `.proto` в `.md`/`.html` в `docs/`

### Image Delivery (ADR-001)

- **Бэкенд**: два фиксированных rendition (`thumbnail_url` 480×600 q80, `url` оригинал q90), оба WebP после обработки воркером.
- **Фронтенд**: выбирать rendition по контексту, не строить `srcset`. `<picture>` только на асинхронном окне (1–5 с).
- **Не делать**: AVIF, imgproxy, Nuxt Image с S3-провайдером, `srcset` из двух rendition.
- **Контракт**: `thumbnail_url` в `GetUploadUrlResponse` не добавлять.

---

## 5. Точки входа в код (эталоны)

| Паттерн | Файл |
| --- | --- |
| Форма с `fieldMask` | `app/components/admin/HeroForm.vue` |
| Курсорная пагинация + фильтр статуса | `app/pages/admin/heroes/index.vue` |
| Presigned Upload + пагинация галерей | `app/components/admin/HeroPhotos.vue` |
| Face Box Editor | `app/components/admin/FaceBoxEditor.vue` |
| Auto-refresh при 401 | `app/plugins/connect.ts` |
| Обработка 403/429 | `app/lib/errors.ts` |
| Краудсорсинг (заявки от пользователей) | `app/pages/submit.vue` |
| SEO карточки героя (без Twitter) | `app/pages/heroes/[id].vue` |
| Форма обратной связи | `app/pages/contacts.vue` + `app/composables/useContactForm.ts` |
| OG-карточка | `app/components/OgImage/HeroCard.takumi.vue` |
| Пагинация галерей (composable) | `app/composables/useHeroPhotos.ts` |
| ISR invalidation | `app/composables/useCachePurge.ts` + `server/api/cache/purge.post.ts` |
| LLM извлечение данных | `app/components/admin/ExtractionPanel.vue` |
| Авто-привязка связей из извлечения | `app/composables/useExtractionLinks.ts` |
| Модерация через LLM | `app/pages/admin/submissions/index.vue` |
| Передача данных между страницами | `sessionStorage` + проп `initial-extraction` |
| Проверка размера текста для LLM | `app/components/admin/ExtractionPanel.vue` (`MAX_TEXT_SIZE`) |
| Управление LLM-провайдерами | `app/pages/admin/llm/index.vue` |
| История модерации заявок | `app/components/admin/SubmissionReviewHistory.vue` |
| Гибкие даты для наград | `app/components/admin/HeroAwards.vue` + `AdminFlexibleDateInput.vue` |
| Гибкие даты для конфликтов | `app/pages/admin/conflicts/index.vue` + `AdminFlexibleDateInput.vue` |
| Клиентская проверка дубликатов | `app/composables/useDuplicateCheck.ts` |
| Синхронизация темы (color-mode + PrimeVue) | `app/composables/useThemeSync.ts` + `app.vue` |
| Парсинг дат из извлечения | `app/components/admin/HeroForm.vue` → `parseExtractedDate` |
| Конвертация `FlexibleDateJson` → protobuf | `app/components/admin/HeroForm.vue` → `prepareDateForApi` + `fromJson` |
| Семантическая ссылка-кнопка | `app/pages/admin/index.vue` (`Button as="router-link"`) |
| Оптимизация галереи (INP) | `app/pages/heroes/[id].vue` (`requestIdleCallback`, `preloadFullImage`, `content-visibility`) |
| Политика AI / robots | `nuxt.config.ts` (robots) + `public/.well-known/ai-policy.json` + `app/pages/ai-policy.vue` |
| Code splitting `/submit` | `app/pages/submit.vue` + `app/components/submit/*.vue` (async) |
| AI Policy | `app/pages/ai-policy.vue` + `public/.well-known/ai-policy.json` |
| Локальные шрифты | `public/fonts/*.ttf` (Golos Text, Playfair Display) |
| Карта локаций (OpenLayers) | `app/components/heroes/HeroLocationMap.vue` — ESRI World Topo, pre-flight проверка, async import через `<ClientOnly>` |
| Связи героев | `app/components/admin/HeroRelations.vue` + блок в `app/pages/heroes/[id].vue` |
| Источники героя | `app/components/admin/HeroSources.vue` + блок в `app/pages/heroes/[id].vue` |

---

## 6. Критичные предупреждения для агента

### НЕ ДЕЛАТЬ

- ❌ Не подключай `@nuxt/fonts` повторно — уже настроен с `provider: 'local'`
- ❌ Не подключай `@nuxt/a11y` — рантайм-проверки замедляют сборку и дают ложные срабатывания на PrimeVue. Для статического анализа достаточно `eslint-plugin-vuejs-accessibility` (уже в devDependencies)
- ❌ Не используй паттерн `<NuxtLink><Button/></NuxtLink>` — только `Button as="router-link"`
- ❌ Не добавляй гибкие даты для `HeroConflict`, `Location`, `Photo`, `HeroSource` — архитектурное решение
- ❌ Не добавляй `twitterCard` или `twitter:*` метатеги — полностью удалены
- ❌ Не добавляй `manualChunks` в `vite.build.rollupOptions` — ломает сборку Nitro

### КРИТИЧНО ПРОВЕРЯТЬ

- ✅ `htmlValidator.ignore` — именно `ignore`, а не `exclude`
- ✅ `useThemeSync.ts` — использует `import.meta.client` для защиты от обращения к `document` на сервере
- ✅ `allowQuery` в `routeRules` — должен строго соответствовать полям `ListHeroesRequest` из `hero.proto`
- ✅ После полного отказа от вложенных `<a><button>` можно рассмотреть повторное включение правила `element-permitted-content` в `htmlValidator`
- ⚠️ `ISR` в `nitro.routeRules` временно отключен (`isr: false`). Не включать без согласования — идет отладка прода.
- ⚠️ PrimeIcons больше нет в проекте. Все иконки брать из `~icons/carbon/...` или `~icons/mdi/...` через `unplugin-icons`.
- ⚠️ Карта локаций использует ESRI World Topo (`services.arcgisonline.com`), **не OSM** (заблокирован в РФ). Не возвращать OSM без веской причины. Pre-flight проверка `probeTileService()` обязательна.
- ⚠️ Плавающие проблемы Lighthouse (BFCache, Unused JS, TBT, Яндекс.Метрика) **не трогать** — зависят от окружения, не от кода приложения.

### КОНТРАСТНОСТЬ (закрыто, не трогать)

- `#b08d57` → `#7a5c2e` (глобальная замена, Спринт 8)
- `#8a6a3c` → `#7a5c2e` в карточке героя (Спринт 9)
- Контрастность ссылок в карточке героя — НЕ требуется исправлять

---

## 7. Деплой

Registry-centric deployment:

```
podman build → podman push → на VPS: podman pull → systemctl --user restart emh-frontend.service
```

Multi-stage build: `node:22.23.1-alpine` builder → `node:22.23.1-alpine` runtime. Порт 3000.

---

## 8. Результаты Lighthouse (сводка)

| Страница | Performance | Ключевые проблемы |
| --- | --- | --- |
| `/heroes/[id]` | **88/100** | TBT 12-70мс, CLS 0.00094, FCP 1.4с | ✅ Image Delivery закрыт, ESRI карта работает |
| `/submit` | 80/100 | TBT 0ms, TTI 1.7s | ✅ Code splitting + unplugin-icons применены |
| `/` (mobile) | ~45/100 | TTI 15.2с, TBT 431мс | 🔴 Требует оптимизации |
| `/contacts` | 66/100 | FCP 1.9s, TTI 4.6s | 🟡 Средний приоритет |
| `/about` | ~85/100 | Unused CSS 68 КиБ | 🟡 Средний приоритет |

---

## 9. Примечания для следующего агента

- Главная проблема — мобильная производительность главной страницы (TTI 15.2 сек). Начинать с lazy loading изображений и оптимизации LCP.
- PrimeIcons полностью удален (Спринт 10). Все иконки — через `~icons/carbon/...` или `~icons/mdi/...`.
- Гибкие даты закрыты — не трогать архитектуру, только использовать существующие компоненты (`AdminFlexibleDateInput.vue`).
- Кнопки-ссылки — только через `Button as="router-link"`, никаких вложенных элементов.
- Контрастность — не трогать, всё исправлено (WCAG AA).
- Image Delivery — не строить `srcset`/AVIF. Выбор `thumbnailUrl` vs `url` по контексту. См. ADR-001.
- Карта локаций — ESRI World Topo (не OSM!), pre-flight проверка обязательна. См. ADR-004.
- Плавающие проблемы Lighthouse (BFCache, Unused JS, TBT, Метрика) — **не трогать**, зависят от окружения.
- `requestIdleCallback` — проверить наличие fallback для Safari (`window.requestIdleCallback ?? setTimeout`).
- После закрытия производительности — фокус на функциональных задачах: связи героев, источники, краудсорсинг, LLM-извлечение.

---

## 10. Changelog сессий (краткий)

### Спринт 10 (2026-09-15/16)

**Производительность:**
- Code splitting `/submit` через `defineAsyncComponent` (TBT 0ms, TTI 1.7s)
- Полный отказ от PrimeIcons SVG в пользу `unplugin-icons` (tree-shaking, -347 КБ)
- `vite.build.target` поднят до `es2024` (устранен Legacy JS)
- Image Delivery: `width`/`height`/`sizes`/`fetchpriority` на всех изображениях. TBT -97%, CLS -69%
- Кэширование шрифтов и статики через Caddy (`Cache-Control: immutable`, 1 год)
- `font-display: swap` для всех шрифтов

**Карты:**
- Карта локаций на `/heroes/[id]` через OpenLayers + ESRI World Topo (ADR-004)
- OSM заменён на ESRI (OSM заблокирован в РФ)
- Pre-flight проверка доступности сервиса тайлов + graceful fallback

**SEO и Инфраструктура:**
- Страница `/ai-policy` + `/.well-known/ai-policy.json`
- ISR временно отключен (`isr: false`) для отладки продакшена
- Очищен `public/_robots.txt` от `Content-Usage`

### Спринт 9 (2026-09-12/14)

**Производительность:**
- `loading="lazy"` + `decoding="async"` на превью галереи
- `aspect-ratio: 4 / 5` на карточках главной
- `preload` + `fetchpriority="high"` для логотипа
- `requestIdleCallback` для аналитики, `preloadFullImage` по `pointerenter`, `content-visibility` на секциях
- TBT: 20ms → 3ms (-85%)

**SEO:**
- `twitterCard` и `twitter:*` полностью удалены
- `Content-Usage` / `Content-Signal` заменены на блокировку AI-ботов

**Качество кода:**
- Рефакторинг `<NuxtLink><Button/>` → `Button as="router-link"`
- Контрастность `#8a6a3c` → `#7a5c2e`
- Гибкие даты для конфликтов

**Закрыто навсегда:**
- Контрастность ссылок в карточке героя
- Публичный API + OpenAPI
- Гибкие даты для `HeroConflict`, `Location`, `Photo`, `HeroSource`

### Спринт 8 (2026-09-11)

- `@nuxt/fonts` с `provider: 'local'`
- `@nuxtjs/color-mode` + `useThemeSync.ts`
- `@nuxtjs/html-validator`
- `@vueuse/nuxt`
- `eslint-plugin-vuejs-accessibility`
- Контрастность `#b08d57` → `#7a5c2e`
- ISR с `allowQuery`
- OpenSearch
- `ogImage.security` (HMAC)
- LightningCSS

### Спринт 7 (2026-08-30)

- Обработка ошибок 401/403/429
- Проверка размера текста для LLM
- `LlmAdminService`
- История модерации
- Гибкие даты для наград
- Уникальные заявки
- Уведомления заявителям
