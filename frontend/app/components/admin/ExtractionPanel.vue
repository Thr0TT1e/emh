<script setup lang="ts">
    import { useToast } from 'primevue/usetoast';
    import { isPermissionDenied, isRateLimited } from '~/lib/errors';
    import type { ExtractHeroDataResponseJson } from '~/sdk/emh/v1/extraction_pb';
    import CheckIcon from '~icons/mdi/check?width=1em&height=1em';
    import LinkIcon from '~icons/mdi/link?width=1em&height=1em';
    import SparklesOutlineIcon from '~icons/mdi/sparkles-outline?width=1em&height=1em';

    const emit = defineEmits<{
        apply: [data: ExtractHeroDataResponseJson];
    }>();

    const { extraction } = useApi();
    const toast = useToast();

    // Лимит бэкенда на размер текста для LLM (см. справку, M5).
    const MAX_TEXT_SIZE = 50 * 1024; // 50 КБ

    // Человекочитаемый размер для сообщения об ошибке.
    const formatBytes = (bytes: number): string => {
        if (bytes < 1024) return `${bytes} Б`;
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`;
        return `${(bytes / (1024 * 1024)).toFixed(1)} МБ`;
    };

    const rawText = ref('');
    const sourceUrl = ref('');
    const extracting = ref(false);
    const result = ref<ExtractHeroDataResponseJson | null>(null);
    const showResult = ref(false);

    const hasInput = computed(() => rawText.value.trim().length > 0 || sourceUrl.value.trim().length > 0);

    // Размер текста в байтах — для счётчика лимита в шаблоне
    // (в шаблонах нет доступа к глобальным объектам браузера — выносим сюда)
    const textSize = computed(() => new Blob([rawText.value]).size);
    const isTextTooLarge = computed(() => textSize.value > MAX_TEXT_SIZE);

    const extract = async () => {
        if (!hasInput.value || extracting.value) return;

        // Проверка размера текста ДО отправки (бэкенд вернёт 400, но лучше отсеять на клиенте).
        const text = rawText.value.trim();
        const textSize = new Blob([text]).size;
        if (text && isTextTooLarge.value) {
            toast.add({
                severity: 'error',
                summary: 'Текст слишком длинный',
                detail: `Размер ${formatBytes(textSize)} превышает лимит 50 КБ. Сократите текст или разбейте на части.`,
                life: 7000,
            });
            return;
        }

        extracting.value = true;
        result.value = null;

        try {
            const res = await extraction.extractHeroData({
                rawText: text,
                sourceUrl: sourceUrl.value.trim(),
            });

            result.value = res;
            showResult.value = true;

            const heroName = [res.hero?.lastName, res.hero?.firstName, res.hero?.middleName]
                .filter(Boolean)
                .join(' ');

            toast.add({
                severity: 'success',
                summary: 'Данные извлечены',
                detail: heroName || 'Герой распознан',
                life: 3000,
            });
        } catch (err: any) {
            // Дифференцированная обработка ошибок бэкенда.
            let summary = 'Ошибка извлечения';
            let detail = err?.message ?? 'Не удалось извлечь данные';

            if (isPermissionDenied(err)) {
                summary = 'Нет прав доступа';
                detail = 'Извлечение данных требует роль администратора.';
            } else if (isRateLimited(err)) {
                summary = 'Превышен лимит запросов';
                detail = 'Слишком много запросов к LLM. Попробуйте через минуту.';
            }

            toast.add({
                severity: 'error',
                summary,
                detail,
                life: 5000,
            });
        } finally {
            extracting.value = false;
        }
    };

    const applyToForm = () => {
        if (!result.value) return;
        emit('apply', result.value);
        showResult.value = false;
    };

    const reset = () => {
        rawText.value = '';
        sourceUrl.value = '';
        result.value = null;
        showResult.value = false;
    };
</script>

<template>
    <Card class="extraction-panel">
        <template #title>
            <div class="extraction-panel__header">
                <SparklesOutlineIcon />
                <span>Извлечение данных через LLM</span>
            </div>
        </template>

        <template #content>
            <!-- Входные данные -->
            <div class="extraction-panel__inputs">
                <label class="afield">
                    <span class="afield__label">Сырой текст (статья, пост, письмо)</span>
                    <Textarea v-model="rawText" class="w-full" rows="6"
                        placeholder="Вставьте текст некролога, статьи или поста…" />
                    <label class="afield">
                        <span class="afield__label">Сырой текст (статья, пост, письмо)</span>
                        <Textarea v-model="rawText" class="w-full" rows="6"
                            placeholder="Вставьте текст некролога, статьи или поста…" />
                        <small :class="isTextTooLarge ? 'text-red-600' : 'text-muted'">
                            {{ textSize }} / {{ MAX_TEXT_SIZE }} байт (лимит 50 КБ)
                        </small>
                    </label>
                </label>

                <label class="afield">
                    <span class="afield__label">Или URL источника</span>
                    <InputText v-model="sourceUrl" class="w-full" placeholder="https://vk.com/wall-123456_789" />
                </label>
            </div>

            <!-- Кнопки -->
            <div class="extraction-panel__actions">
                <Button :label="extracting ? 'Извлекаем…' : 'Извлечь данные'" :loading="extracting"
                    :disabled="!hasInput" @click="extract" />
                <Button v-if="result" outlined label="Сбросить" @click="reset" />
            </div>

            <!-- Результат -->
            <div v-if="showResult && result" class="extraction-panel__result">
                <Divider />

                <!-- Предупреждения -->
                <Message v-if="result.warnings?.length" severity="warn" :closable="false" class="mb-3">
                    <ul class="extraction-panel__warnings">
                        <li v-for="(w, i) in result.warnings" :key="i">{{ w }}</li>
                    </ul>
                </Message>

                <!-- Дубликаты -->
                <Message v-if="result.duplicateHeroIds?.length" severity="info" :closable="false" class="mb-3">
                    <p>
                        ⚠️ Возможно, этот герой уже есть в базе
                        ({{ result.duplicateHeroIds.length }} совпадений).
                        Проверьте перед сохранением.
                    </p>
                </Message>

                <!-- Извлечённые данные -->
                <AdminExtractionResultPreview :data="result" />

                <!-- Информация об авто-привязке связей -->
                <div v-if="result && (result.conflicts?.length || result.awards?.length || result.locations?.length)"
                    class="extraction-panel__links-info">
                    <p class="extraction-panel__links-hint">
                        <LinkIcon /> После создания героя будут автоматически привязаны:
                        <template v-if="result.conflicts?.length">конфликты ({{ result.conflicts.length }})</template>
                        <template v-if="result.awards?.length"> · награды ({{ result.awards.length }})</template>
                        <template v-if="result.locations?.length"> · локации ({{ result.locations.length }})</template>
                    </p>
                </div>

                <!-- Кнопка применения -->
                <div class="extraction-panel__apply">
                    <Button label="Применить к форме" @click="applyToForm">
                        <template #icon>
                            <CheckIcon />
                        </template>
                    </Button>
                    <Button outlined label="Отмена" @click="showResult = false" />
                </div>
            </div>
        </template>
    </Card>
</template>

<style scoped>
    .extraction-panel {
        border: 1px dashed var(--emh-bronze);
    }

    .extraction-panel__header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        color: var(--emh-bronze);
    }

    .extraction-panel__inputs {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        margin-bottom: 1rem;
    }

    .extraction-panel__actions {
        display: flex;
        gap: 0.6rem;
    }

    .extraction-panel__warnings {
        margin: 0;
        padding-left: 1.2rem;
        font-size: 0.9rem;
    }

    .extraction-panel__preview-title {
        font-size: 1rem;
        font-weight: 600;
        margin-bottom: 0.8rem;
    }

    .extraction-panel__preview-grid {
        display: grid;
        grid-template-columns: 1fr 2fr;
        gap: 0.4rem 1rem;
        font-size: 0.9rem;
    }

    .extraction-panel__preview-grid dt {
        color: var(--emh-muted);
        font-weight: 500;
    }

    .extraction-panel__apply {
        display: flex;
        gap: 0.6rem;
        margin-top: 1.2rem;
    }

    .extraction-panel__links-hint {
        font-size: 0.85rem;
        color: var(--emh-muted);
        display: flex;
        align-items: center;
        gap: 0.4rem;
        margin-top: 0.8rem;
    }
</style>