# CONTEXT SNAPSHOT — EMH Frontend (Спринт 9+)

**Технический паспорт проекта «Вечная память героям»**

---

## 1. Метаданные

| Параметр | Значение |
|---|---|
| Проект | Вечная память героям (EMH) |
| Scope | Только `frontend/` (Nuxt 4) |
| Фаза | Публичная часть, админка, модерация, SEO-стек завершены. Спринт 9: оптимизация производительности и закрытие техдолга |
| Дата снимка | 2026-09-14 |
| Production VPS | 82.202.139.132 |
| Канонический домен | вежливые.рус (IDN, кириллица) |
| Зеркала (301 → канонический) | вечнаяпамятьгероям.рус, emh.su, neverforgotten.ru |
| Технические поддомены | api.вежливые.рус → бэкенд, s3.вежливые.рус → MinIO |

---

## 2. Стек и зависимости

### Runtime

| Компонент | Версия | Назначение |
|---|---|---|
| Nuxt | ^4.5.2 | Фреймворк (структура `app/`) |
| Vue | ^3.5.47 | UI |
| PrimeVue | 4.5.5 (зафиксирована) | UI-кит + `@primeuix/themes` |
| @bufbuild/protobuf | ^2.16.0 | protobuf-es v2 (рантайм) |
| @connectrpc/connect-web | ^2.2.1 | Connect RPC транспорт |
| vue-advanced-cropper | ^2.8.9 | Face Box UI |
| OpenLayers (ol) | ^10.10.0 | Карты (локации героев) |
| Node.js | 22.23.1 | Runtime |
| Пакетный менеджер | pnpm@12.3.4 | type: module |

### SEO и оптимизация

| Модуль | Назначение |
|---|---|
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

---

## 3. Текущее состояние и приоритеты

### ✅ Завершено (Спринт 9)

- **Ленивая загрузка**: `loading="lazy"` + `decoding="async"` на превью галереи (`[id].vue`, `HeroPhotos.vue`)
- **Alt-фолбэк**: `image.description || \`Фотография ${fullName}\``
- **Twitter-метатеги**: полностью удалены (заменены на Open Graph)
- **CLS на главной**: `aspect-ratio: 4 / 5` на `.hero-card__photo`
- **Preload логотипа**: `fetchpriority="high"` + `<link rel="preload">`
- **Рефакторинг кнопок**: `<NuxtLink><Button/></NuxtLink>` → `Button as="router-link"`
- **Гибкие даты для конфликтов**: `Conflict` использует `start_date_info` / `end_date_info`
- **Оптимизация INP**: `trackPhotoView` в `requestIdleCallback`, предзагрузка по `pointerenter`, `content-visibility: auto`

### 🔴 Критичные задачи (следующий спринт)

| # | Задача | Описание |
|---|---|---|
| 1 | Мобильная производительность главной | TTI = 15.2 сек, TBT = 431 мс. Требуется: code splitting, lazy loading изображений, переход с PrimeIcons SVG на woff2 |
| 2 | Замена `Content-Usage` / `Content-Signal` в `robots.txt` | Директивы не являются стандартом. План многоуровневой замены готов, но не реализован |
| 3 | Image Delivery (score 0) | Добавить `<picture>` с WebP/AVIF и responsive sizes |

### 🟡 Средний приоритет

- **Дубли `robots`-метатегов**: требует аудита `nuxt-seo-utils`
- **Legacy JavaScript (133.8ms)**: поднять `vite.build.target` до `es2022`
- **`Numeric tagPriority (35)` в unhead**: заменить на алиас `critical`
- **Inline `<script>` 3.0KB**: вынести во внешний файл

### ⚪ Отложено / Не делается

- Множественные face box (групповые фото)
- Кэширование `ListHeroes` на уровне Nitro
- Service Worker для оффлайн-режима админки
- Grafana-дашборд
- PWA manifest
- **Гибкие даты для `HeroConflict`, `Location`, `Photo`, `HeroSource`** — НЕ ДЕЛАТЬ (архитектурное решение)
- **Контрастность ссылок в карточке героя** — не требуется (`#7a5c2e` = WCAG AA 4.5:1)
- **Публичный API + OpenAPI** — документация генерируется из `.proto` в `.md`/`.html` в `docs/`

---

## 4. Архитектурные решения (зафиксированы, не обсуждаются)

### Гибкие даты (FlexibleDate)

**Реализовано и работает:**
- `Hero` — даты рождения, гибели, начала службы (`birth_date_info`, `death_date_info`, `service_start_date_info`)
- `HeroAward` — дата награждения (`award_date_info`)
- `Conflict` — даты начала и окончания (`start_date_info`, `end_date_info`)

**НЕ делаем и не планируем:**
- `HeroConflict` — даты участия героя в конфликте
- `Location` — нет временных атрибутов у географии
- `Photo` — нет даты съёмки в контракте
- `HeroSource` — нет даты публикации источника

**Эталоны:** `AdminFlexibleDateInput.vue` + `prepareDateForApi()` из `HeroForm.vue`

### Конвертация дат

```typescript
// Паттерн: гибкая дата → protobuf Message
const prepareDateForApi = (fd?: FlexibleDateJson | null) => {
  const json = (!fd || fd.precision === 'DATE_PRECISION_UNSPECIFIED')
    ? { precision: 'DATE_PRECISION_UNKNOWN' as const, displayText: '' }
    : fd;
  return fromJson(FlexibleDateSchema, json);
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
'/heroes': {
  isr: {
    expiration: 300,
    passQuery: true,
    allowQuery: ['search_query', 'conflict_id', 'location_id', 'cursor', 'date_from', 'date_to'],
  },
},
'/heroes/**': { isr: { expiration: 3600 } },
```

**Критично:** `allowQuery` должен строго соответствовать полям `ListHeroesRequest` из `hero.proto`. Без него — cache poisoning.

---

## 5. Техдолг и открытые задачи

| # | Область | Проблема | Статус |
|---|---|---|---|
| 3 | Дубли `robots`-метатегов | Глобальный `seoMeta` + локальный `useSeoMeta` | 🔵 Отложено |
| 17 | `[id].vue` | `Numeric tagPriority (35)` в unhead | 🟡 Низкий приоритет |
| 18 | `[id].vue` | Inline `<script>` 3.0KB | 🟡 Низкий приоритет |
| 20 | Image Delivery | score 0, требуется `<picture>` с WebP/AVIF | 🔴 Открыт |
| 21 | Legacy JavaScript | 133.8ms, поднять `vite.build.target` до `es2022` | 🟡 Средний |
| 22 | `/submit` | PrimeIcons SVG 347 КБ | 🔴 Открыт |
| 23 | `/submit` | JS-бандлы ~640 КБ, требуется code splitting | 🟡 Средний |
| 24 | `contacts.vue` | Контрастность ссылки 1.25:1 | 🔴 Открыт |
| 25 | `contacts.vue` | Избыточные `v-reveal` на первом экране | 🟡 Средний |
| 26 | `/contacts`, `/submit` | Нет `preconnect` к S3 и `preload` шрифтов | 🟡 Средний |
| 27 | `/about`, `/contacts`, `/submit` | Нет `preconnect` к `s3.neverforgotten.ru` | 🔴 Открыт |
| 29 | `/` (mobile) | TTI = 15.2 сек | 🔴 Критично |
| 30 | `/` (mobile) | TBT = 431 мс | 🔴 Критично |

---

## 6. Точки входа в код (эталоны)

| Паттерн | Файл |
|---|---|
| Форма с `fieldMask` | `app/components/admin/HeroForm.vue` |
| Курсорная пагинация + фильтр статуса | `app/pages/admin/heroes/index.vue` |
| Presigned Upload + пагинация галерей | `app/components/admin/HeroPhotos.vue` |
| Face Box Editor | `app/components/admin/FaceBoxEditor.vue` |
| Auto-refresh при 401 | `app/plugins/connect.ts` |
| Обработка 403/429 | `app/lib/errors.ts` |
| SEO карточки героя | `app/pages/heroes/[id].vue` |
| OG-карточка | `app/components/OgImage/HeroCard.takumi.vue` |
| ISR invalidation | `app/composables/useCachePurge.ts` + `server/api/cache/purge.post.ts` |
| Гибкие даты для наград | `app/components/admin/HeroAwards.vue` + `AdminFlexibleDateInput.vue` |
| Гибкие даты для конфликтов | `app/pages/admin/conflicts/index.vue` |
| Синхронизация темы | `app/composables/useThemeSync.ts` + `app.vue` |
| Семантическая ссылка-кнопка | `app/pages/admin/index.vue` (`Button as="router-link"`) |

---

## 7. Критичные предупреждения для агента

### НЕ ДЕЛАТЬ

- ❌ Не подключай `@nuxt/fonts` повторно — уже настроен с `provider: 'local'`
- ❌ Не подключай `@nuxt/a11y` — рантайм-проверки замедляют сборку и дают ложные срабатывания на PrimeVue. Для статического анализа достаточно `eslint-plugin-vuejs-accessibility` (уже в devDependencies).
- ❌ Не используй паттерн `<NuxtLink><Button/></NuxtLink>` — только `Button as="router-link"`
- ❌ Не добавляй гибкие даты для `HeroConflict`, `Location`, `Photo`, `HeroSource` — архитектурное решение
- ❌ Не добавляй `twitterCard` или `twitter:*` метатеги — полностью удалены

### КРИТИЧНО ПРОВЕРЯТЬ

- ✅ `htmlValidator.ignore` — именно `ignore`, а не `exclude`
- ✅ `useThemeSync.ts` — использует `import.meta.client` для защиты от обращения к `document` на сервере
- ✅ `allowQuery` в `routeRules` — должен строго соответствовать полям `ListHeroesRequest` из `hero.proto`
- ✅ После полного отказа от вложенных `<a><button>` можно рассмотреть повторное включение правила `element-permitted-content` в `htmlValidator`

### КОНТРАСТНОСТЬ

Контрастность полностью соответствует WCAG AA. Не менять цвета без необходимости.

---

## 8. Деплой

**Registry-centric deployment:**
```bash
podman build → podman push → на VPS: podman pull → systemctl --user restart emh-frontend.service
```

**Multi-stage build:** `node:22.23.1-alpine` builder → `node:22.23.1-alpine` runtime. Порт 3000.

---

## 9. Результаты Lighthouse (контрольные замеры)

### `/heroes/[id]` (после оптимизации)

- **Total Blocking Time:** 20ms → 3ms (-85%)
- **Speed Index:** улучшен на ~15%
- **CLS:** стабильно 0
- **Image Delivery:** требует доработки (score 0)
- **Legacy JavaScript:** 133.8ms

### `/submit`

- **Performance:** 80/100
- **PrimeIcons SVG:** 347 КБ (требует перехода на шрифт или tree-shaking)
- **JS-бандлы:** ~640 КБ (требует code splitting)

### `/contacts`

- **Performance:** 66/100
- **FCP:** 1.9s, TTI: 4.6s — требуют оптимизации
- **Контрастность ссылки в форме:** 1.25:1

### `/` (mobile)

- **Performance:** ~45/100 (критично)
- **TTI:** 15.2 сек (цель < 3.8 сек)
- **TBT:** 431 мс (цель < 200 мс)
