# ТЗ: Фронтенд — Защита от дублей в админке наград

**Дата:** 2026-09-17  
**Статус:** Implemented (см. «Отклонения» в §8)  
**Связано:** ADR-005 (секция «Сид справочника наград»), awards-seed-backend.md

## 1. Область работ

Добавить двухуровневую защиту от дублирования наград при создании через админку `/admin/awards`:
- **Уровень 1 (UX):** клиентская проверка по кэшу `ListAwards` → toast-предупреждение
- **Уровень 2 (страховка):** обработка `AlreadyExists` от сервера

## 2. Уровень 1: клиентская проверка

### 2.1. Форма создания награды

Перед вызовом `awardAdmin.createAward` проверять уникальность `(name, type, jurisdiction)`:

```typescript
// В app/pages/admin/awards/index.vue (или соответствующая форма)

const catalog = useAwardCatalog()  // composable с кэшем ListAwards

const isDuplicate = computed(() => {
    if (!form.name || !form.type || !form.jurisdiction) return false
    return catalog.value.some(a =>
        a.name === form.name &&
        a.type === form.type &&
        a.jurisdiction === form.jurisdiction,
    )
})

const handleSubmit = async () => {
    if (isDuplicate.value) {
        toast.add({
            severity: 'warn',
            summary: 'Возможный дубликат',
            detail: `Награда «${form.name}» с типом ${typeLabel(form.type)} и юрисдикцией ${jurLabel(form.jurisdiction)} уже существует в справочнике.`,
            life: 5000,
        })
        return  // не отправляем
    }
    await awardAdmin.createAward({ ...form })
}
```

### 2.2. UI-индикатор

В форме, пока валиден `name+type+jurisdiction` и совпадает с существующей:

```vue
<Message v-if="isDuplicate" severity="warn" icon="~icons/carbon/warning-alt">
    Такая награда уже есть в справочнике
</Message>
```

Или inline под полем `name`:

```vue
<div v-if="isDuplicate" class="afield__error">
    <IconCarbonWarningAlt class="text-orange-500" />
    <span>Награда уже существует в справочнике</span>
</div>
```

### 2.3. Composable `useAwardCatalog`

Использует `useAsyncData` с кэшированием (паттерн из существующего `HeroAwards.vue`):

```typescript
// app/composables/useAwardCatalog.ts
export const useAwardCatalog = () => {
    return useAsyncData('awards-catalog', async () => {
        const res = await useApi().award.listAwards({})
        return res.awards
    })
}
```

## 3. Уровень 2: обработка `AlreadyExists`

### 3.1. Обёртка в `app/lib/errors.ts`

Добавить обработку `connect.CodeAlreadyExists`:

```typescript
// app/lib/errors.ts

export function handleConnectError(err: ConnectError, ctx: {
    duplicateMessage?: string
}) {
    if (err.code === Code.AlreadyExists) {
        toast.add({
            severity: 'error',
            summary: 'Уже существует',
            detail: ctx.duplicateMessage ?? err.message,
            life: 5000,
        })
        return
    }
    // ... существующая обработка 403/429
}
```

### 3.2. В форме создания

```typescript
try {
    await awardAdmin.createAward({ ...form })
    toast.add({ severity: 'success', summary: 'Награда создана', life: 2500 })
} catch (err) {
    handleConnectError(err, {
        duplicateMessage: 'Награда с таким именем, типом и юрисдикцией уже существует',
    })
}
```

## 4. UX-детали

- Toast-предупреждение (уровень 1) — `severity: 'warn'`, жизнь 5 сек
- Toast-ошибка (уровень 2) — `severity: 'error'`, жизнь 5 сек
- Форма должна быть блокирована к отправке при `isDuplicate` (disabled кнопка)
- Иконки: `~icons/carbon/warning-alt` и `~icons/carbon/warning` (unplugin-icons, не PrimeIcons!)

## 5. Что НЕ делать

- ❌ Не добавлять `useDuplicateCheck` composable из `app/composables/useDuplicateCheck.ts` как есть — там паттерн для героев, не для наград. Лучше сделать inline для простоты.
- ❌ Не использовать PrimeIcons (`<i class="pi pi-...">`) — они удалены из проекта.
- ❌ Не делать серверную валидацию на клиенте через RPC — только локальная проверка по кэшу + обработка `AlreadyExists`.

## 6. Тестирование

- [ ] Component-тест: toast появляется при совпадении `name+type+jurisdiction`
- [ ] Component-тест: кнопка disabled при дубликате
- [ ] E2E: серверный `AlreadyExists` корректно обрабатывается

> Решением Q6 component-тесты не заводятся: во фронте нет тестовой инфраструктуры
> (vitest/playwright), harness ради двух тестов непропорционален. Верификация:
> `pnpm build` + `oxlint` + ручная браузерная проверка на dev-сервере.

## 7. Критерии приёмки

- [x] Клиентская проверка блокирует отправку дубликата
- [x] Серверный `AlreadyExists` показывает человекочитаемое сообщение
- [x] Используются `~icons/carbon/*` (unplugin-icons), не PrimeIcons
- [ ] `pnpm build` и `oxlint` — зелёные

## 8. Отклонения от исходного ТЗ (реализация 2026-11-11)

| # | ТЗ | Факт | Причина |
|---|----|------|---------|
| 1 | §3.1 новая обёртка `handleConnectError` в `lib/errors.ts` | Использован существующий предикат `isAlreadyExists` (паттерн submit.vue), toast с **серверным** текстом сообщения | Конвенция репо — предикаты + локальные toast'ы на странице; формулировка ошибки живёт на бэкенде (единый источник) |
| 2 | §2.1 toast-предупреждение при submit (уровень 1) | §4 disabled-кнопка + inline `Message severity="warn"` (Q3=a); submit-тост уровня 1 не нужен | disabled+Message объясняют причину до клика; submit-guard оставлен как тихая страховка |
| 3 | §2.3 `useAwardCatalog` = `useAsyncData('awards-catalog', ...)` одной страницей | Ключ `'awards-catalog-full'`, цикл курсорной пагинации pageSize=100 до пустого nextCursor | `'awards-catalog'` занят HeroAwards.vue (урезанные id/name первой страницы) — конфликт ключа отдал бы неполный каталог; без цикла при 400+ проверка дырявая (дефолт usecase — 20) |
| 4 | Спецификация не упоминает | Проверка дублей и в режиме редактирования с исключением собственной записи (`a.id !== editingId`) | Q4=b: rename в существующий ключ тоже конфликтует; серверный маппинг 23505 для Update добавлен в бэкенд-срезе |
| 5 | Спецификация не упоминает | Сравнение — строгое равенство строк, без нормализации | Q5=a: зеркалит `idx_awards_seed_key`; нормализация ложно блокировала бы то, что сервер принимает |
| 6 | Спецификация не упоминает | `refreshCatalog()` после create/update/delete и при `AlreadyExists` | Освобождение/занимание ключа должно отражаться на индикаторе без перезагрузки страницы |
| 7 | §1 scope | Добавлена курсорная пагинация таблицы (Q8=b, pageSize=50) + **бэкенд-фикс keyset-курсора ListAwards** (Q9=a) | Без них при сиде 400+ каталог и таблица неполные: старый курсор `WHERE id > cursor` при сортировке по sort_order терял записи и повторял их. См. ADR-005 «Уточнения от 2026-11-11» |
| 8 | — | Подписи типа/юрисдикции в Message — `formTypeLabel`/`formJurisdictionLabel` (из TYPE_OPTIONS/JURISDICTION_OPTIONS), а не `awardTypeLabel`/`awardJurisdictionLabel` | Те принимают JSON-строки proto, форма держит numeric — иначе текст показывал «Не указан» при выбранном типе (поймано браузерной проверкой) |
| 9 | — | Бэкенд: `awardColumns` обёрнут в COALESCE `description`/`image_url` | Latent-баг: nullable-колонки (миграция 00002) роняли весь ListAwards в 500 («cannot scan NULL into *string») на строках без полей; найден браузерной проверкой пагинации, regression-тест `TestAwardList_ScansNullColumns` |
