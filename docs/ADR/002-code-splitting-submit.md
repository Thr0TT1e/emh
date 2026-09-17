# ADR-002: Code Splitting для `/submit`

## Статус
**Принято** (2026-09-15)

## Контекст
Страница `/submit` имеет JS-бандл **~640 КБ** (Lighthouse Performance score 80/100). Анализ показал:
- PrimeVue компоненты (~200-300 КБ) — autoImport работает, но страница использует много компонентов
- Protobuf runtime + generated code (~150-200 КБ)
- Connect RPC (~50-80 КБ)
- Свой код + composables (~100-150 КБ)

Критичные тяжёлые компоненты:
- `HeroSearchPicker` — поиск существующих героев для supplement mode
- `AttachmentUpload` — загрузка документов к заявке

## Решение

### 1. Вынести тяжёлые компоненты в отдельные файлы

Создать:
- `app/components/submit/HeroSearchPicker.vue`
- `app/components/submit/AttachmentUpload.vue`

### 2. Использовать `defineAsyncComponent` для lazy loading

В `app/pages/submit.vue`:
```vue
<script setup lang="ts">
const HeroSearchPicker = defineAsyncComponent(() => 
  import('~/components/submit/HeroSearchPicker.vue')
)

const AttachmentUpload = defineAsyncComponent(() => 
  import('~/components/submit/AttachmentUpload.vue')
)
</script>

<template>
  <div class="submit-form">
    <!-- Above fold: immediate -->
    <SubmitterInfo v-model:name="submitterName" v-model:email="submitterEmail" />
    
    <!-- Below fold: lazy -->
    <HeroSearchPicker
      v-if="mode === 'supplement'"
      v-model:hero-id="targetHeroId"
      v-model:hero-name="targetHeroName"
    />
    
    <HeroDetails v-if="mode === 'new'" v-model="newHero" />
    
    <AttachmentUpload v-model:attachments="attachments" />
  </div>
</template>
```

### 3. Проверить бандл через `nuxt build --analyze`

После рефакторинга запустить:
```bash
pnpm build --analyze
```

Целевые метрики:
- **До**: ~640 КБ (один chunk)
- **После**: ~300-400 КБ (основной chunk) + отдельные chunks для lazy компонентов

## Последствия

### Положительные
- ✅ Initial bundle size: 640 КБ → ~300-400 КБ (-35-45%)
- ✅ Faster First Contentful Paint (FCP)
- ✅ Lazy loading компонентов ниже fold
- ✅ Лучше code organization

### Отрицательные
- ❌ Небольшая задержка при первом рендере lazy компонентов (но они ниже fold)
- ❌ Дополнительная сложность в коде (но это стандартная практика)

## Метрики успеха
- Initial JS bundle: ≤400 КБ
- Lighthouse Performance score: ≥85
- FCP improvement: ≥200ms

## Ссылки
- `app/pages/submit.vue` — исходная страница
- `app/components/admin/FaceBoxEditor.vue` — пример использования `defineAsyncComponent` (не влияет на `/submit`)
- `app/components/admin/ExtractionPanel.vue` — пример использования `defineAsyncComponent` (не влияет на `/submit`)
