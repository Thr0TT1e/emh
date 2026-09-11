<script setup lang="ts">
    /**
     * Диалог «История модерации» — аудит-трейл по заявке.
     * Использует RPC ListSubmissionReviews (Спринт 7, задача #36).
     */
    import { useToast } from 'primevue/usetoast';
    import { formatDate } from '~/lib/format';
    import { toPlain } from '~/lib/pb';
    import type { SubmissionReviewJson } from '~/sdk/emh/v1/submission_pb';
    import { SubmissionReviewSchema } from '~/sdk/emh/v1/submission_pb';

    const props = defineProps<{ submissionId: string }>();
    const visible = defineModel<boolean>('visible', { default: false });

    const { submissionAdmin } = useApi();
    const toast = useToast();

    const reviews = ref<SubmissionReviewJson[]>([]);
    const loading = ref(false);

    const DECISION_META: Record<string, { label: string; severity: 'success' | 'danger' }> = {
        SUBMISSION_REVIEW_DECISION_APPROVE: { label: 'Одобрено', severity: 'success' },
        SUBMISSION_REVIEW_DECISION_REJECT: { label: 'Отклонено', severity: 'danger' },
    };

    async function loadHistory() {
        loading.value = true;
        reviews.value = [];
        try {
            const res = await submissionAdmin.listSubmissionReviews({ submissionId: props.submissionId });
            // toPlain конвертирует Message → Json: createdAt становится строкой,
            // решение — строковым enum. Это убирает необходимость в `as string`.
            reviews.value = (res.reviews ?? []).map((r) => toPlain(SubmissionReviewSchema, r));
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось загрузить историю', life: 5000 });
        } finally {
            loading.value = false;
        }
    }

    // Загружаем историю при открытии диалога
    watch(visible, (v) => {
        if (v) loadHistory();
    });
</script>

<template>
    <Dialog v-model:visible="visible" modal header="История модерации" style="width: 560px">
        <div v-if="loading" class="flex justify-center py-6">
            <ProgressSpinner style="width: 36px; height: 36px" />
        </div>
        <div v-else-if="reviews.length === 0" class="p-4 text-center text-muted">
            Заявка ещё не рассматривалась
        </div>
        <Timeline v-else :value="reviews" layout="vertical">
            <template #marker="slotProps">
                <i class="pi" :class="slotProps.item.decision === 'SUBMISSION_REVIEW_DECISION_APPROVE'
                    ? 'pi-check-circle text-green-600'
                    : 'pi-times-circle text-red-600'" style="font-size: 1.4rem" />
            </template>
            <template #content="slotProps">
                <div class="review-item">
                    <div class="review-item__head">
                        <Tag :value="DECISION_META[slotProps.item.decision ?? '']?.label ?? slotProps.item.decision"
                            :severity="DECISION_META[slotProps.item.decision ?? '']?.severity ?? 'secondary'" />
                        <span class="review-item__reviewer">{{ slotProps.item.reviewerName }}</span>
                        <span class="review-item__date">{{ formatDate(slotProps.item.createdAt) }}</span>
                    </div>
                    <p v-if="slotProps.item.comment" class="review-item__comment">
                        {{ slotProps.item.comment }}
                    </p>
                </div>
            </template>
        </Timeline>
        <template #footer>
            <Button outlined label="Закрыть" @click="visible = false" />
        </template>
    </Dialog>
</template>

<style scoped>
    .review-item {
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }

    .review-item__head {
        display: flex;
        align-items: center;
        gap: 0.6rem;
        flex-wrap: wrap;
    }

    .review-item__reviewer {
        font-weight: 600;
    }

    .review-item__date {
        color: var(--emh-muted);
        font-size: 0.85rem;
    }

    .review-item__comment {
        margin: 0;
        font-size: 0.9rem;
        color: var(--emh-muted);
        white-space: pre-line;
    }
</style>