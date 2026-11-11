# CONTEXT SNAPSHOT — EMH Frontend (ADR-006 performance-контур, v0.0.9)

Технический паспорт проекта «Вечная память героям»

---

## 1. Метаданные

| Параметр | Значение |
|---|---|
| Проект | Вечная память героям (EMH) |
| Scope | Только `frontend/` (Nuxt 4) |
| Фаза | Спринт 11 (2026-11-11) — **ADR-006 performance-контур публичной части**: компрессия на Caddy, reveal убран с реестра, LCP-логотип нативным `<img>` + preload, preconnect, кэш S3 (релиз v0.0.9). ISR временно отключён для отладки прода |
| Дата снимка | 2026-11-11 |
| Production VPS | 82.202.139.132 |
| Канонический домен | вежливые.рус (IDN, кириллица) |
| Зеркала (301 → канонический) | вечнаяпамятьгероям.рус, emh.su, neverforgotten.ru |
| Технические поддомены | api.neverforgotten.ru → бэкенд, s3.neverforgotten.ru → MinIO (живут на neverforgotten.ru, редирект зеркал покрывает только апекс) |
| Версии образов | Quadlet'ы фронтенда/бэкенда/мигратора — `:latest` из релиза **v0.0.9** |
| CI | ✅ [run 42 / v0.0.8](https://codeberg.org/Thr0TT1e/emh/actions/runs/42/jobs/0/attempt/1), [run 44 / v0.0.8-1](https://codeberg.org/Thr0TT1e/emh/actions/runs/44/jobs/0/attempt/1), [run 45 / v0.0.9](https://codeberg.org/Thr0TT1e/emh/actions/runs/45) |
| Пакеты образов | [emh-frontend:latest](https://codeberg.org/Thr0TT1e/-/packages/container/emh-frontend/latest), [emh-backend:latest](https://codeberg.org/Thr0TT1e/-/packages/container/emh-backend/latest) |
| Зеркало/референс | origin — `codeberg.org/Thr0TT1e/emh`; проверка/актуализация — `github.com/Thr0TT1e/emh` |

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

### ✅ Завершено (сессия 2026-11-11 — админка наград, v0.0.8-1)

| # | Задача | Статус | Описание |
| --- | --- | --- | --- |
| 1 | Полный каталог наград для дубль-чека | ✅ | `useAwardCatalog.ts`: курсорный цикл `pageSize=100` до пустого `nextCursor`, ключ `awards-catalog-full`, `refresh()` после мутаций |
| 2 | Уровень 1 — клиентская проверка дублей | ✅ | Строгое равенство `(name, type, jurisdiction)` (зеркало `idx_awards_seed_key`), self-exclude при редактировании, inline `Message severity="warn"` + disabled «Создать» с title |
| 3 | Уровень 2 — серверный `AlreadyExists` | ✅ | `isAlreadyExists(e)` → toast с серверным русским сообщением; после конфликта — `refreshCatalog()` |
| 4 | Пагинация таблицы справочника | ✅ | `pageSize=50`, `cursorStack` + «← Назад / Вперёд →» (прецедент `admin/heroes`) |
| 5 | Keyset-пагинация на бэкенде (зависимость) | ✅ | `ListAwards` — кортеж `(sort_order,name,id)` вместо `id > cursor`; курсор непрозрачный (см. ADR-005) |
| 6 | Баг-фиксы, найденные браузерной проверкой | ✅ | numeric vs JSON-подписи enum в Message; бэкенд: COALESCE `description`/`image_url` (latent-500 на NULL) |

### ✅ Завершено (сессия 2026-11-11 — ADR-006 performance-контур, v0.0.9)

| # | Задача | Статус | Описание |
| --- | --- | --- | --- |
| 1 | HTTP-компрессия на edge | ✅ | Caddy `encode zstd gzip` на каноническом хосте. HTML −6.4× (364→57 КиБ), JS −3.6×, SVG −3.4×. Без пересборки образа |
| 2 | `Cache-Control` на S3 | ✅ | awards/photos/heroes — `immutable, 1 год` (write-once); thumbnails — `max-age=300, stale-while-revalidate=604800` (перезапись при FaceBox-правах) |
| 3 | Reveal убран с реестра | ✅ | `.reveal`/`v-reveal`/stagger removed from `index.vue` cards. Контент первого экрана не ждёт гидрации |
| 4 | Eager для первых 4 карточек | ✅ | `loading="eager"`, остальные `lazy`. Мёртвый `sizes` (без srcset) убран. `fetchpriority` карточкам НЕ даётся (конкуренция с LCP) |
| 5 | LCP-логотип | ✅ | PrimeVue `<Image>` → нативный `<img width height fetchpriority="high">` + `preload` в head. CSS `height: auto` |
| 6 | Preconnect | ✅ | `nuxt.config.ts`: s3 (no-cors, без crossorigin) + api (`crossorigin=''`). Все страницы |
| 7 | `prefers-reduced-motion` | ✅ | Глобальный a11y-fallback в `main.css` — reveal не скрывает контент |

### 🔴 Критичные задачи (следующий спринт)

Производительность публичной части закрыта (ADR-006). Критичных задач нет —
следующий фокус: функциональные (§🔵), начиная со «Связей между героями».

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
| 5 | ~~`v-reveal` на первом экране `contacts.vue`~~ — закрыто ADR-006 T7: глобальный `prefers-reduced-motion` fallback в `main.css` |
| 6 | Компонентные тесты для админки наград (нет test-harness вообще — см. §6) |

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
- Файл `public/_robots.txt` очищен от нестандартных директив (удалён `Content-Usage`) и больше не переопределяет/не конфликтует с модулем `@nuxtjs/robots`.

### Иконки (unplugin-icons)

PrimeIcons полностью удалён из-за отсутствия tree-shaking (бандл 347 КБ SVG загружался целиком).
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

### 🆕 Админка наград: защита от дублей (v0.0.8-1)

- **Уровень 1 (UX)** — клиентский каталог `useAwardCatalog` (ключ `awards-catalog-full`, курсорный цикл `pageSize=100`). Сравнение — **строгое равенство** тройки `(name, type, jurisdiction)`; нормализация (trim/case-fold) НЕ применяется: она ложно блокировала бы то, что сервер принимает. Собственная запись исключается по `id` при редактировании. Индикация — inline `Message severity="warn"` + disabled-кнопка с `title`; submit-тост уровня 1 не используется.
- **Уровень 2 (страховка)** — серверный `CodeAlreadyExists` → dedicated toast с серверным текстом (единый источник формулировок) + `refreshCatalog()`, чтобы индикатор уровня 1 подхватил конфликт.
- **Refresh-модель**: каталог обновляется после create/update/delete и после `AlreadyExists`; на открытии диалога не перезапрашивается (цена — гонка двух админов, её ловит уровень 2).
- **Таблица**: курсорная пагинация `pageSize=50`, `cursorStack` (паттерн `admin/heroes`). Курсор `ListAwards` — непрозрачный keyset-токен, фронт его не парсит.

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
| Клиентская проверка дубликатов (заявки) | `app/composables/useDuplicateCheck.ts` |
| 🆕 Полный каталог наград (дубль-чек админки) | `app/composables/useAwardCatalog.ts` (ключ `awards-catalog-full`) |
| 🆕 Админка наград: дубли + пагинация | `app/pages/admin/awards/index.vue` |
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
- ❌ Не переиспользуй ключ `useAsyncData('awards-catalog')` для полного каталога — он занят `HeroAwards.vue` (урезанные `{id, name}` первой страницы)
- ❌ Не нормализуй имя награды в дубль-чеке (trim/case-fold/ёлочки→дюймы) — сервер сравнивает побайтово, ложная блокировка хуже пропуска

### КРИТИЧНО ПРОВЕРЯТЬ

- ✅ `htmlValidator.ignore` — именно `ignore`, а не `exclude`
- ✅ `useThemeSync.ts` — использует `import.meta.client` для защиты от обращения к `document` на сервере
- ✅ `allowQuery` в `routeRules` — должен строго соответствовать полям `ListHeroesRequest` из `hero.proto`
- ✅ После полного отказа от вложенных `<a><button>` можно рассмотреть повторное включение правила `element-permitted-content` в `htmlValidator`
- ⚠️ `ISR` в `nitro.routeRules` временно отключён (`isr: false`). Не включать без согласования — идёт отладка прода
- ⚠️ PrimeIcons больше нет в проекте. Все иконки брать из `~icons/carbon/...` или `~icons/mdi/...` через `unplugin-icons`
- ⚠️ Карта локаций использует ESRI World Topo (`services.arcgisonline.com`), **не OSM** (заблокирован в РФ). Pre-flight проверка `probeTileService()` обязательна
- ⚠️ Плавающие проблемы Lighthouse (BFCache, Unused JS, TBT, Яндекс.Метрика) **не трогать** — зависят от окружения, не от кода приложения
- 🆕 ⚠️ `awardTypeLabel`/`awardJurisdictionLabel` (`app/lib/awards.ts`) принимают **JSON-строки** proto; форма наград держит **numeric** enum → для подписей формы использовать `formTypeLabel`/`formJurisdictionLabel` (из `TYPE_OPTIONS`/`JURISDICTION_OPTIONS`)
- 🆕 ⚠️ `ListAwards`-курсор — непрозрачный keyset-токен; не парсить, не конструировать вручную, не считать UUID
- 🆕 ⚠️ В таблице наград `pageSize=50` — не убирать `pagination` из запроса: без него usecase отдаёт первые 20 и «теряет» хвост справочника
- 🆕 ⚠️ `CreateAward` требует валидный `image_url` (protovalidate) — форма без загруженного знака отклоняется сервером до проверки дублей

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

Образы (обновлено): quadlet'ы `emh-frontend.container` / `emh-backend.container` / `emh-migrate.container` ссылаются на `:latest`, собранный из релиза **v0.0.9**.
Пакеты: [emh-frontend:latest](https://codeberg.org/Thr0TT1e/-/packages/container/emh-frontend/latest), [emh-backend:latest](https://codeberg.org/Thr0TT1e/-/packages/container/emh-backend/latest).
CI: [run 42 / v0.0.8](https://codeberg.org/Thr0TT1e/emh/actions/runs/42/jobs/0/attempt/1), [run 44 / v0.0.8-1](https://codeberg.org/Thr0TT1e/emh/actions/runs/44/jobs/0/attempt/1), run 45 / v0.0.9.

**Инфра (ADR-006):** Caddyfile на VPS (`~/emh/Caddyfile`, volume в `emh-caddy.container`) содержит `encode zstd gzip` + S3 `Cache-Control`-матчеры. Деплой без пересборки образов: sync → `caddy validate` → `podman exec emh-caddy caddy reload`. Откат — бэкап `~/emh/Caddyfile.bak-adr006`.
Тег `v0.0.9` — **lightweight** (annotated не триггерит CI — прецедент v0.0.8/run 42).

---

## 8. Результаты Lighthouse (сводка)

| Страница | Performance | Ключевые проблемы |
| --- | --- | --- |
| `/heroes/[id]` | **88/100** | TBT 12-70мс, CLS 0.00094, FCP 1.4с | ✅ Image Delivery закрыт, ESRI карта работает |
| `/submit` | 80/100 | TBT 0ms, TTI 1.7s | ✅ Code splitting + unplugin-icons применены |
| `/` (mobile) | **56–61/100** (v0.0.9) | TTI 6.7–7.0с, TBT 450–630мс, CLS 0, FCP 2.8с, LCP 6.7–7.0с | 🟡 ADR-006 закрыт: сеть/видимость исправлены; LCP упирается в Render Delay (декод SVG + гидрация, CPU throttling), не в сеть. Дальнейшее = отклонённые не-цели (Unused JS/TBT-джиттер) |
| `/contacts` | 66/100 | FCP 1.9s, TTI 4.6s | 🟡 Средний приоритет |
| `/about` | ~85/100 | Unused CSS 68 КиБ | 🟡 Средний приоритет |
| `/admin/awards` | не замерялся | — | админка за middleware, в Lighthouse-план не входит |

---

## 9. Примечания для следующего агента

- Производительность публичной части закрыта ADR-006 (v0.0.9): компрессия на Caddy, reveal убран с реестра, LCP-логотип, preconnect, кэш S3. Не возвращаться к «lazy loading» и «оптимизации LCP» главной — сделано.
- Реестр на `/` сознательно БЕЗ reveal — не возвращать `v-reveal` на карточки главной (решение G2: анимация ниже сгиба — прогрессивное улучшение, не условие видимости).
- Плагин `reveal.ts` и класс `.reveal` остаются для about/contacts; `prefers-reduced-motion` fallback обязателен.
- Для LCP-элементов нативный `<img>` с width/height/fetchpriority — PrimeVue `<Image>` запрещён.
- `preconnect` к s3/api — в глобальном `app.head.link` (`nuxt.config.ts`), на всех страницах.
- PrimeIcons полностью удалён (Спринт 10). Все иконки — через `~icons/carbon/...` или `~icons/mdi/...`.
- Гибкие даты закрыты — не трогать архитектуру, только использовать существующие компоненты (`AdminFlexibleDateInput.vue`).
- Кнопки-ссылки — только через `Button as="router-link"`, никаких вложенных элементов.
- Контрастность — не трогать, всё исправлено (WCAG AA).
- Image Delivery — не строить `srcset`/AVIF. Выбор `thumbnailUrl` vs `url` по контексту. См. ADR-001.
- Карта локаций — ESRI World Topo (не OSM!), pre-flight проверка обязательна. См. ADR-004.
- Плавающие проблемы Lighthouse (BFCache, Unused JS, TBT, Метрика) — **не трогать**, зависят от окружения.
- `requestIdleCallback` — проверить наличие fallback для Safari (`window.requestIdleCallback ?? setTimeout`).
- 🆕 Админка наград: дубль-чек — двухуровневый (каталог `awards-catalog-full` + серверный `AlreadyExists`); ключ `awards-catalog` занят `HeroAwards.vue`, не переиспользовать.
- 🆕 Админка наград: подписи формы — из OPTIONS-массивов (`formTypeLabel`/`formJurisdictionLabel`), не из JSON-хелперов.
- 🆕 Курсор `ListAwards` — непрозрачный keyset; не парсить. Пагинация таблицы наград — 50 записей.
- 🆕 Перед работой над дублями наград вспомнить про валидацию `image_url` (protovalidate): без картинки создать награду нельзя, значит сценарий «двух табов» требует загруженного знака.
- После закрытия производительности — фокус на функциональных задачах: связи героев, источники, краудсорсинг, LLM-извлечение.

---

## 10. Changelog сессий (краткий)

### Сессия 2026-11-11 (ADR-006 performance-контур, v0.0.9)

**Инфраструктура (без пересборки образа):**
- Caddy `encode zstd gzip` на каноническом хосте — HTML −6.4×, entry JS −3.6×, SVG −3.4× (замер на проде)
- S3 `Cache-Control` по мутабельности ключа: awards/photos/heroes `immutable, 1 год`; thumbnails `max-age=300, stale-while-revalidate=604800` (перезапись при FaceBox-правах — см. Reference - photo pipeline §1–2)
- Pre-flight: `caddy validate` проходил даже с placeholder bcrypt-хешем в console-блоке (fail-closed 401, не падает)

**Фронт:**
- Реестр на `/`: `.reveal`/`v-reveal`/stagger убраны полностью — контент виден до гидрации
- Первые 4 карточки `loading="eager"` (два ряда мобильной сетки), остальные `lazy`; мёртвый `sizes` убран; `fetchpriority` карточкам не даётся (конкуренция с LCP)
- Логотип: PrimeVue `<Image>` → нативный `<img width="300" height="378" fetchpriority="high">` (viewBox 3816.55×4804.25), `height:auto` вместо `height:100%/object-fit`; `preload as=image` в head
- `nuxt.config.ts` head: preconnect s3 (no-cors, без crossorigin) + api (`crossorigin=''`) + dns-prefetch s3 — на всех страницах
- `main.css`: `@media (prefers-reduced-motion: reduce)` → `.reveal{opacity:1;transform:none;transition:none}`

**Верификация:** prerender-HTML + прод (grep-чек §7 спеки: 0 reveal/stagger, eager=4, preload/preconnect на месте), `pnpm build` + `oxlint` — зелёные, curl-замеры компрессии/кэша на проде. Lighthouse mobile (два прогона): Perf 56–61, TTI 6.7–7.0 с (−54%), CLS 0, LCP-элемент зарегистрирован (логотип `<img>`).

**Мета:** тег `v0.0.9` lightweight; CI run 45 — success (4 джоба, включая Trigger VPS Deploy); образ `:latest` перетянут на VPS, `systemctl --user restart emh-frontend`. Исправлена неточность снапшота: технические поддомены — `api.neverforgotten.ru`/`s3.neverforgotten.ru` (не `*.вежливые.рус`).

### Сессия 2026-11-11 (админка наград, v0.0.8-1)

**Функциональность:**
- Полный каталог наград для клиентской проверки дублей: `useAwardCatalog.ts` (курсорный цикл `pageSize=100`, ключ `awards-catalog-full`, refresh после мутаций)
- Уровень 1: строгое равенство `(name, type, jurisdiction)`, self-exclude в edit, inline `Message` + disabled «Создать» с title
- Уровень 2: `isAlreadyExists` → toast с серверным сообщением + рефреш каталога
- Таблица справочника: курсорная пагинация `pageSize=50` (`cursorStack`, «← Назад / Вперёд →»)

**Зависимость на бэкенде:**
- Keyset-курсор `ListAwards` (кортеж `sort_order,name,id`, непрозрачный токен) вместо теряющего записи `id > cursor` — см. ADR-005 «Уточнения от 2026-11-11»

**Найденные и исправленные баги:**
- Подписи типа/юрисдикции в Message показывали «Не указан» (numeric-форма vs JSON-хелперы)
- Бэкенд: `ListAwards` падал в 500 на строках с NULL `image_url`/`description` (COALESCE + regression-тест)

**Верификация:**
- Браузерный стенд (бэкенд + fake-S3 + `.output`): обход 65 записей без потерь/повторов, дубль→Message+disabled, rename-в-дубль блокируется, гонка двух табов → `already_exists` → после рефреша ловится уровнем 1
- `pnpm build` + `oxlint` — зелёные; бэкенд: unit + integration (`-p 1`) — 85 passed

**Инфраструктура:**
- Теги `v0.0.8`, `v0.0.8-1`; CI run 42 и run 44 — success; образы `latest` опубликованы (backend + frontend), quadlet'ы переведены на `latest`

### Спринт 10 (2026-09-15/16)

**Производительность:**
- Code splitting `/submit` через `defineAsyncComponent` (TBT 0ms, TTI 1.7s)
- Полный отказ от PrimeIcons SVG в пользу `unplugin-icons` (tree-shaking, -347 КБ)
- `vite.build.target` поднят до `es2024` (устранён Legacy JS)
- Image Delivery: `width`/`height`/`sizes`/`fetchpriority` на всех изображениях. TBT -97%, CLS -69%
- Кэширование шрифтов и статики через Caddy (`Cache-Control: immutable`, 1 год)
- `font-display: swap` для всех шрифтов

**Карты:**
- Карта локаций на `/heroes/[id]` через OpenLayers + ESRI World Topo (ADR-004)
- OSM заменён на ESRI (OSM заблокирован в РФ)
- Pre-flight проверка доступности сервиса тайлов + graceful fallback

**SEO и Инфраструктура:**
- Страница `/ai-policy` + `/.well-known/ai-policy.json`
- ISR временно отключён (`isr: false`) для отладки продакшена
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
