# Техническое задание: Фронтенд — Орденская планка

**Дата:** 2026-09-17
**Статус:** Implemented
**Связано:** ADR-005

> Документ приведён в соответствие с фактической реализацией.
> Ключевые отличия от первой редакции — в разделе «9. Отклонения от исходного ТЗ».

## 1. Доступ к данным

Публичная страница героя **не** выполняет второй запрос к каталогу. Поля справочника
денормализованы в `HeroAward` на бэкенде (JOIN в `hero_award_repository.ListByHero`):
`ribbonImageUrl`, `imageUrl`, `type`, `jurisdiction`, `wornWithoutBar`, `isJubilee`.
Данные приходят одним `GetHero`-запросом, поэтому `useAsyncData('awards-catalog', …)`
и коллизия этого ключа с `admin/HeroAwards.vue` устранены.

## 2. Логика сортировки — `app/lib/awards.ts`

Чистый модуль, переиспользуется планкой и текстовым списком.

```typescript
export const RIBBON_BAR_THRESHOLD = 3;

// enum'ы приходят строками — ранги через таблицы соответствия (не арифметикой).
export const compareAwards = (a: SortableAward, b: SortableAward): number => {
  // jurisdiction → type → isJubilee → sortOrder → name(localeCompare 'ru')
};

// В планку: лента загружена И награда носится на колодке.
export const isRibbonBearing = (a: SortableAward): boolean =>
  Boolean(a.ribbonImageUrl) && !a.wornWithoutBar;

export const ribbonAwardsOf = <T extends SortableAward>(awards: readonly T[]): T[] =>
  awards.filter(isRibbonBearing).sort(compareAwards);

export const sortedAwards = <T extends SortableAward>(awards: readonly T[]): T[] =>
  [...awards].sort(compareAwards);

// Лейблы для Select/Tag в админке
export const awardTypeLabel = (t?: string): string => /* Орден | Медаль | Знак отличия | Не указан */;
export const awardJurisdictionLabel = (j?: string): string => /* РФ | СССР | Ведомственная | Не указана */;
```

Порядок старшинства (приказ МО РФ №1500): `jurisdiction` (РФ → СССР → ведомственные →
UNSPECIFIED в конец) → `type` (ORDER → MEDAL → BADGE) → `isJubilee` (false раньше) →
`sortOrder` (asc) → `name`. `HeroAward` не несёт `sortOrder`, поэтому для привязок
ось пропускается; порядок внутри равных рангов сохраняет серверный
`ORDER BY a.sort_order ASC` (стабильность `Array.prototype.sort`).

## 3. Компонент `app/components/heroes/AwardRibbonBar.vue`

Авто-имя в Nuxt — `<HeroesAwardRibbonBar>` (pathPrefix не отключён).

### 3.1. Props

```typescript
// Награды уже отфильтрованы (ribbon-bearing) и отсортированы вызывающей страницей.
defineProps<{ awards: HeroAwardJson[] }>()
```

Каталог в props **не передаётся** — всё денормализовано.

### 3.2. Поведение

- `<ul class="arb__bar">` из `<li>`: лента `ribbonImageUrl`, фиксированные `width="36" height="12"`,
  `loading="lazy" decoding="async"`, `:alt="awardName"`.
- Кнопка-триггер «Все награды (N)» — `<button type="button" :aria-expanded aria-haspopup="dialog">`.
  Открывает **один** `<Popover>` (PrimeVue 4, автоимпорт) по клику/тапу — одинаково
  на desktop и mobile. Компонента `<Tooltip>` в PrimeVue нет.
- Внутри `<Popover>` — **знак награды** `imageUrl` (планка — лишь «образец» награды),
  название и номер приказа. `Popover` рендерится через Portal в `<body>` (`role="dialog"`),
  поэтому вложенность безопасна для htmlValidator.
- `aria-expanded` синхронизируется через события `@show`/`@hide` Popover.
- Типизация ref на Popover — локальный минимальный интерфейс `{ toggle(event: Event): void }`,
  т.к. ссылаться на тип автоимпортируемого компонента в type-позиции нельзя.

### 3.3. Стили

Scoped CSS на токенах `--emh-*` (`--emh-surface`, `--emh-line`, `--emh-ink`,
`--emh-muted`, `--emh-crimson`, `--emh-crimson-dark`). BEM-префикс `arb__`.
Фиксированные размеры `<img>` против CLS. `content-visibility` **не** применяется
(блок в первом экране). Глобальный dark-mode не трогаем (токены `--emh-*` не имеют
`.dark-mode`-переопределений — это отдельная задача).

## 4. Интеграция в `app/pages/heroes/[id].vue`

```typescript
import { RIBBON_BAR_THRESHOLD, ribbonAwardsOf, sortedAwards } from '~/lib/awards';

const ribbonAwards = computed(() => ribbonAwardsOf(data.value?.awards ?? []));
const showRibbonBar = computed(() => ribbonAwards.value.length > RIBBON_BAR_THRESHOLD);
const listAwards = computed(() => sortedAwards(data.value?.awards ?? []));
```

```vue
<HeroesAwardRibbonBar v-if="showRibbonBar" :awards="ribbonAwards" />

<ul v-else-if="listAwards.length" class="hero__awards">
  <li v-for="a in listAwards" :key="a.awardId" class="hero__award">
    <img v-if="a.imageUrl" class="hero__award-sign" :src="a.imageUrl" alt=""
      width="48" height="48" loading="lazy" decoding="async" />
    <span class="hero__award-text">
      <span class="hero__award-name">{{ a.awardName }}</span>
      <span v-if="a.decreeNumber" class="hero__award-decree">{{ a.decreeNumber }}</span>
    </span>
  </li>
</ul>
```

Текстовый fallback показывает **все** награды героя (включая `wornWithoutBar`
и награды без ленты), каждая со знаком `imageUrl`. `alt=""` — знак декоративен,
название рядом текстом. `.hero__award` перешёл с `align-items: baseline` на `center`,
текст обёрнут в `.hero__award-text` (baseline внутри строки с картинкой).

## 5. Админка: `app/pages/admin/awards/index.vue`

Страница **уже существовала** (вопреки формулировке первой редакции ТЗ «новая»).
Расширена формой и таблицей:

| Поле | Компонент | Примечание |
|------|-----------|------------|
| `name` | InputText | обязательно (валидация в `save()`) |
| `description` | Textarea | |
| `imageUrl` | ImageUpload (`AWARD_IMAGE`) | знак награды |
| `ribbonImageUrl` | ImageUpload (`AWARD_RIBBON`) | лента для планки |
| `type` | Select | обязательно, не `UNSPECIFIED` (валидация в `save()`) |
| `jurisdiction` | Select | РФ / СССР / ведомственная / не указана |
| `wornWithoutBar` | Checkbox (`:binary`) | |
| `isJubilee` | Checkbox (`:binary`) | |
| `sortOrder` | InputNumber | ≥ 0 |

- Enum'ы в форме держатся как numeric (`AwardType` / `AwardJurisdiction`),
  чтение из JSON — через мапы `TYPE_JSON_TO_NUM` / `JURISDICTION_JSON_TO_NUM`
  (паттерн `admin/conflicts/index.vue`).
- `fieldMask` для `updateAward` — полный snake_case-набор:
  `name, description, image_url, sort_order, ribbon_image_url, type, jurisdiction, worn_without_bar, is_jubilee`.
- DataTable: добавлены колонки «Лента» (36×12, empty-state) и «Тип» (Tag с типом
  и, при указанной, принадлежностью).
- Иконки — явный импорт `~icons/...` (ADR-003), не PrimeIcons.

## 6. Производительность

- `loading="lazy"` + `decoding="async"` на всех `<img>`.
- Явные `width`/`height` против CLS.
- Один rendition ленты + CSS-масштаб; thumbnail-воркер отложен (см. ADR-001, backend-ТЗ).

## 7. Тестирование

Frontend-тесты в проекте не настроены (нет vitest/playwright). Новые тесты **не**
добавлялись (Q18а/Q29а). Верификация среза:

- [x] `pnpm exec oxlint` — 0 warnings, 0 errors
- [x] `pnpm build` — exit 0 (`✨ Build complete!`)
- [x] typecheck `app/lib/awards.ts` через `tsc -p` — exit 0
- [ ] Полная HTML-валидация блока наград (`htmlValidator`) — только в dev/preview
      с бэкендом и данными (при production-`build` валидатор отключён,
      `/heroes/[id]` не пререндерится). Интеграционная проверка.
- [ ] Визуальная/ручная проверка планки, Popover, fallback-списка, mobile.

## 8. Критерии приёмки

1. ✅ `app/lib/awards.ts` с компаратором по `jurisdiction → type → isJubilee → sortOrder → name`
2. ✅ `AwardRibbonBar.vue` создан, рендерит планку + единый Popover со знаком награды
3. ✅ Страница героя: `v-if="showRibbonBar"` (ribbon-bearing > 3) / `v-else` текстовый список
4. ✅ Порог считается по ribbon-bearing наградам
5. ✅ `loading="lazy"` + фиксированные размеры на всех изображениях
6. ✅ `/admin/awards` ведёт новые поля (type, jurisdiction, worn_without_bar, is_jubilee, ribbon_image_url)
7. ⛔ Редактор `devices[]` — вне среза (нет `UpdateHeroAwardRequest`)

## 9. Отклонения от исходного ТЗ

- Убран `catalog: AwardJson[]` из props и `useAsyncData('awards-catalog')` — данные денормализованы.
- Убраны `useMediaQuery` и связка Popover/Tooltip — единый Popover по клику.
- Компаратор вынесен в `app/lib/awards.ts`; enum'ы сравниваются через таблицы рангов, не арифметикой.
- Порог `> 3` — по ribbon-bearing наградам, логика на уровне страницы (`v-if/v-else`).
- В Popover — знак награды (`imageUrl`), не увеличенная лента.
- Раздел про редактор `devices[]` в `HeroAwards.vue` снят (вне среза).
- `/admin/awards` — расширение существующей страницы, а не новая.
- Один rendition ленты + CSS-масштаб вместо двух rendition.
