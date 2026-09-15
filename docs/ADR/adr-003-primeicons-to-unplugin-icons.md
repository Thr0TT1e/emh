# ADR-003: Замена PrimeIcons на unplugin-icons + Material Design Icons

## Статус
**Принято** (2026-09-15)

## Контекст
PrimeIcons SVG весит **347 КБ** и тянет весь набор иконок, даже если используются только несколько. Это критично влияет на:
- TTI главной страницы (15.2 сек)
- Overall bundle size
- Performance score

Альтернативы:
- Font Awesome 6 Free (~400 КБ) — тяжёлый
- Iconify runtime — runtime overhead
- **unplugin-icons + @iconify-json/mdi** — compile-time, tree-shakable ✅

## Решение

### 1. Установить зависимости

```bash
pnpm add -D unplugin-icons @iconify-json/mdi
```

### 2. Настроить Vite plugin в `nuxt.config.ts`

```ts
import Icons from 'unplugin-icons/vite'

export default defineNuxtConfig({
  // ... другие настройки
  
  vite: {
    plugins: [
      tailwindcss(),
      Icons({
        compiler: 'vue3',
        autoInstall: true,
      }),
    ],
  },
})
```

### 3. Использовать иконки в компонентах

**Старый способ (PrimeIcons):**
```vue
<i class="pi pi-user" />
<Button icon="pi pi-plus" label="Добавить" />
```

**Новый способ (unplugin-icons):**
```vue
<i-mdi-account />
<Button label="Добавить">
  <template #icon>
    <i-mdi-plus />
  </template>
</Button>
```

### 4. Маппинг PrimeIcons → MDI

Полная таблица соответствий в `docs/migration/primeicons-to-mdi.md`.

### 5. Постепенная миграция

Порядок миграции (по приоритету):
1. Главная страница (`app/pages/index.vue`) — критично для TTI
2. `/submit` — большой бандл
3. `/heroes/[id]` — детальная страница
4. Админка — не критично для performance

### 6. Удалить PrimeIcons после полной миграции

```bash
pnpm remove primeicons
```

Удалить из `nuxt.config.ts`:
```ts
// УДАЛИТЬ эту строку
css: ['~/assets/css/main.css', 'primeicons/primeicons.css']
```

## Последствия

### Положительные
- ✅ Экономия ~300 КБ (347 КБ SVG → только используемые иконки ~10-50 КБ)
- ✅ Compile-time resolution (иконки как Vue компоненты)
- ✅ Tree-shakable (только используемые иконки в бандл)
- ✅ Ноль runtime overhead
- ✅ Поддержка 200K+ иконок через Iconify наборы
- ✅ Совместим с PrimeVue 4 (иконки как слоты)

### Отрицательные
- ❌ Требует настройки Vite plugin
- ❌ Нужно мигрировать все иконки вручную
- ❌ Небольшой learning curve для команды

## Метрики успеха
- Bundle size reduction: ≥300 КБ
- TTI improvement на главной: ≥2 сек
- Количество используемых иконок: <100

## Альтернативы (отклонены)

### A. Font Awesome 6 Free
**Вердикт**: отклонено
- Тяжёлый SVG (~400 КБ)
- Лицензионные ограничения

### B. Iconify runtime
**Вердикт**: отклонено
- Runtime overhead
- Дополнительные HTTP requests

### C. Material Design Icons (manual)
**Вердикт**: отклонено
- Нужно подключать вручную
- Нет tree-shaking без дополнительной настройки

## Ссылки
- [unplugin-icons documentation](https://github.com/unplugin/unplugin-icons)
- [Iconify MDI set](https://icon-sets.iconify.design/mdi/)
- [PrimeIcons to MDI mapping](docs/migration/primeicons-to-mdi.md)
