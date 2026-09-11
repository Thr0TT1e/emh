<script setup lang="ts">
  import type {
    ExtractFromSubmissionResponseJson,
    ExtractHeroDataResponseJson,
  } from '~/sdk/emh/v1/extraction_pb';

  type ExtractionData = ExtractHeroDataResponseJson | ExtractFromSubmissionResponseJson;

  defineProps<{ data: ExtractionData }>();

  const fullName = (data: ExtractionData) =>
    [data.hero?.lastName, data.hero?.firstName, data.hero?.middleName]
      .filter(Boolean)
      .join(' ');
</script>

<template>
  <div class="extraction-preview">
    <!-- Предупреждения LLM -->
    <Message v-if="data.warnings?.length" severity="warn" :closable="false" class="mb-3">
      <ul class="extraction-preview__warnings">
        <li v-for="(w, i) in data.warnings" :key="i">{{ w }}</li>
      </ul>
    </Message>

    <!-- Дубликаты (только для ExtractHeroData; у submission-ответа поля нет) -->
    <Message v-if="'duplicateHeroIds' in data && data.duplicateHeroIds?.length" severity="info" :closable="false"
      class="mb-3">
      ⚠️ Возможно, этот герой уже есть в базе
      ({{ data.duplicateHeroIds!.length }} совпадений). Проверьте перед сохранением.
    </Message>

    <!-- Извлечённые данные -->
    <h4 class="extraction-preview__title">Извлечённые данные:</h4>
    <dl class="extraction-preview__grid">
      <template v-if="fullName(data)">
        <dt>ФИО</dt>
        <dd>{{ fullName(data) }}</dd>
      </template>
      <template v-if="data.hero?.rank">
        <dt>Звание</dt>
        <dd>{{ data.hero.rank }}</dd>
      </template>
      <template v-if="data.hero?.unit">
        <dt>Подразделение</dt>
        <dd>{{ data.hero.unit }}</dd>
      </template>
      <template v-if="data.hero?.birthDate">
        <dt>Дата рождения</dt>
        <dd>{{ data.hero.birthDate }}</dd>
      </template>
      <template v-if="data.hero?.deathDate">
        <dt>Дата гибели</dt>
        <dd>{{ data.hero.deathDate }}</dd>
      </template>
      <template v-if="data.hero?.causeOfDeath">
        <dt>Причина гибели</dt>
        <dd>{{ data.hero.causeOfDeath }}</dd>
      </template>
      <template v-if="data.conflicts?.length">
        <dt>Конфликты</dt>
        <dd>{{data.conflicts.map((c) => c.name).join(', ')}}</dd>
      </template>
      <template v-if="data.awards?.length">
        <dt>Награды</dt>
        <dd>{{data.awards.map((a) => a.name).join(', ')}}</dd>
      </template>
      <template v-if="data.locations?.length">
        <dt>Локации</dt>
        <dd>{{data.locations.map((l) => l.name).join(', ')}}</dd>
      </template>
    </dl>

    <!-- Подсказка об авто-привязке связей -->
    <p v-if="data.conflicts?.length || data.awards?.length || data.locations?.length"
      class="extraction-preview__links-hint">
      <i class="pi pi-link" /> После создания героя будут автоматически привязаны:
      <template v-if="data.conflicts?.length">конфликты ({{ data.conflicts.length }})</template>
      <template v-if="data.awards?.length"> · награды ({{ data.awards.length }})</template>
      <template v-if="data.locations?.length"> · локации ({{ data.locations.length }})</template>
    </p>
  </div>
</template>

<style scoped>
  .extraction-preview__warnings {
    margin: 0;
    padding-left: 1.2rem;
    font-size: 0.9rem;
  }

  .extraction-preview__title {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 0.8rem;
  }

  .extraction-preview__grid {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 0.4rem 1rem;
    font-size: 0.9rem;
  }

  .extraction-preview__grid dt {
    color: var(--emh-muted);
    font-weight: 500;
  }

  .extraction-preview__links-hint {
    font-size: 0.85rem;
    color: var(--emh-muted);
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin-top: 0.8rem;
  }
</style>