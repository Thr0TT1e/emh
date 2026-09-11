<script setup lang="ts">
  import type {
    ExtractFromSubmissionResponseJson,
    ExtractHeroDataResponseJson,
  } from '~/sdk/emh/v1/extraction_pb';

  definePageMeta({ layout: "admin", middleware: "admin" });
  useHead({ title: "Новый герой — канцелярия" });

  // Извлечение из модерации заявок (если админ пришёл по кнопке "Создать героя")
  type ExtractionData = ExtractHeroDataResponseJson | ExtractFromSubmissionResponseJson;
  const initialExtraction = ref<ExtractionData | undefined>(undefined);

  // Читаем синхронно и только на клиенте (админка не индексируется,
  // данные должны быть доступны до монтирования HeroForm).
  if (import.meta.client) {
    const raw = sessionStorage.getItem('emh-submission-extraction');
    if (raw) {
      try {
        initialExtraction.value = JSON.parse(raw);
      } catch {
        initialExtraction.value = undefined;
      }
      sessionStorage.removeItem('emh-submission-extraction'); // чистим сразу
    }
  }
</script>

<template>
  <Card>
    <template #title>Новый герой</template>
    <template #subtitle>После создания можно добавить награды и фотографии</template>

    <template #content>
      <AdminHeroForm mode="create" :initial-extraction="initialExtraction" />
    </template>
  </Card>
</template>