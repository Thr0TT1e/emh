<script setup lang="ts">
    /**
     * /admin/submissions — модерация пользовательских заявок.
     * Конвенции: useApi() (как в HeroForm.vue), числа enum в запросы,
     * JSON-строки enum в UI-мапах, toPlain для ответов.
     */
    import { useToast } from 'primevue/usetoast'
    import { formatDate } from '~/lib/format'
    import { toPlain } from '~/lib/pb'
    import type { PublicationStatusJson } from '~/sdk/emh/v1/enums_emh_pb'
    import { PublicationStatus } from '~/sdk/emh/v1/enums_emh_pb'
    import type { ExtractFromSubmissionResponseJson } from '~/sdk/emh/v1/extraction_pb'
    import type { SubmissionJson } from '~/sdk/emh/v1/submission_pb'
    import { SubmissionReviewDecision, SubmissionSchema } from '~/sdk/emh/v1/submission_pb'

    definePageMeta({ layout: 'admin', middleware: 'admin' })
    useHead({ title: 'Модерация заявок — Вечная память героям' })

    const { submissionAdmin, extraction } = useApi()
    const toast = useToast()
    const router = useRouter();

    // ── История модерации (диалог) ─────────────────────────────────
    const historyVisible = ref(false)
    const historySubmissionId = ref('')

    function openHistory(s: SubmissionJson) {
        historySubmissionId.value = s.id ?? ''
        historyVisible.value = true
    }

    // ---------------------------------------------------------------------------
    // Список + курсорная пагинация (паттерн cursorStack из admin/heroes/index.vue)
    // ---------------------------------------------------------------------------
    const PAGE_SIZE = 20

    const submissions = ref<SubmissionJson[]>([])
    const loading = ref(false)

    const currentCursor = ref('')
    const cursorStack = ref<string[]>([])
    const nextCursor = ref('')
    const hasPrev = computed(() => cursorStack.value.length > 0)

    // Фильтр держим числом — в запрос уходит без конвертации.
    // Дефолт — DRAFT («на модерации»): это очередь модератора.
    const statusFilter = ref<PublicationStatus>(PublicationStatus.DRAFT)

    // UNSPECIFIED (0) в ListSubmissionsRequest валидация не запрещает → «все»
    const STATUS_OPTIONS = [
        { label: 'Все статусы', value: PublicationStatus.UNSPECIFIED },
        { label: 'На модерации', value: PublicationStatus.DRAFT },
        { label: 'Одобренные', value: PublicationStatus.PUBLISHED },
        { label: 'Отклонённые', value: PublicationStatus.ARCHIVED },
    ]

    // toPlain отдаёт статус строкой PublicationStatusJson — мапа для тегов
    const STATUS_META: Record<PublicationStatusJson, { label: string; severity: 'warn' | 'success' | 'danger' | 'secondary' }> = {
        PUBLICATION_STATUS_UNSPECIFIED: { label: '—', severity: 'secondary' },
        PUBLICATION_STATUS_DRAFT: { label: 'На модерации', severity: 'warn' },
        PUBLICATION_STATUS_PUBLISHED: { label: 'Одобрена', severity: 'success' },
        PUBLICATION_STATUS_ARCHIVED: { label: 'Отклонена', severity: 'danger' },
    }

    // ── LLM-извлечение из заявки ──────────────────────────────────
    const extracting = ref(false)
    const extractionResult = ref<ExtractFromSubmissionResponseJson | null>(null)

    async function extractFromSubmission() {
        if (!active.value || extracting.value) return
        extracting.value = true
        extractionResult.value = null
        try {
            const res = await extraction.extractFromSubmission({ submissionId: active.value.id })
            extractionResult.value = res
            toast.add({ severity: 'success', summary: 'Данные извлечены', life: 3000 })
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка извлечения', detail: e?.message ?? 'Не удалось извлечь данные', life: 5000 })
        } finally {
            extracting.value = false
        }
    }

    // Переход на форму создания героя с передачей извлечения через sessionStorage
    function createHeroFromExtraction() {
        if (!extractionResult.value) return
        sessionStorage.setItem('emh-submission-extraction', JSON.stringify(extractionResult.value))
        router.push('/admin/heroes/new')
    }

    async function load(cursor = '') {
        loading.value = true
        try {
            const res = await submissionAdmin.listSubmissions({
                pagination: { pageSize: PAGE_SIZE, cursor },
                status: statusFilter.value, // число, 0 = все статусы
            })
            submissions.value = res.submissions.map((s) => toPlain(SubmissionSchema, s))
            nextCursor.value = res.pagination?.nextCursor ?? ''
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось загрузить заявки', life: 5000 })
        } finally {
            loading.value = false
        }
    }

    function goNext() {
        if (!nextCursor.value) return
        cursorStack.value.push(currentCursor.value)
        currentCursor.value = nextCursor.value
        load(currentCursor.value)
    }

    function goPrev() {
        if (!cursorStack.value.length) return
        currentCursor.value = cursorStack.value.pop()!
        load(currentCursor.value)
    }

    watch(statusFilter, () => {
        cursorStack.value = []
        currentCursor.value = ''
        nextCursor.value = ''
        load()
    })

    onMounted(() => load())

    // ---------------------------------------------------------------------------
    // Диалог рассмотрения
    // ---------------------------------------------------------------------------
    const reviewVisible = ref(false)
    const reviewing = ref(false)
    const active = ref<SubmissionJson | null>(null)
    const moderatorComment = ref('')

    function openReview(s: SubmissionJson) {
        active.value = s
        moderatorComment.value = ''
        extractionResult.value = null
        reviewVisible.value = true
    }

    // payload_json — строка. Показываем pretty-JSON, при битом JSON — raw.
    const prettyPayload = computed(() => {
        if (!active.value?.payloadJson) return ''
        try {
            return JSON.stringify(JSON.parse(active.value.payloadJson), null, 2)
        } catch {
            return active.value.payloadJson
        }
    })

    async function review(decision: SubmissionReviewDecision) {
        if (!active.value || reviewing.value) return
        reviewing.value = true
        try {
            const res = await submissionAdmin.reviewSubmission({
                id: active.value.id,
                decision, // числовой enum — как в HeroForm.vue
                moderatorComment: moderatorComment.value,
            })
            const updated = toPlain(SubmissionSchema, res.submission!)
            const idx = submissions.value.findIndex((s) => s.id === updated.id)
            if (idx !== -1) submissions.value[idx] = updated
            reviewVisible.value = false
            // После успешного Review бэкенд ставит письмо в очередь.
            // Уведомление асинхронное (100–500мс) — тост показываем сразу.
            toast.add({
                severity: 'success',
                summary: decision === SubmissionReviewDecision.APPROVE ? 'Одобрено' : 'Отклонено',
                detail: decision === SubmissionReviewDecision.APPROVE
                    ? 'Заявитель получит уведомление на email'
                    : 'Заявитель получит уведомление с причиной отклонения',
                life: 6000,
            })
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось применить решение', life: 5000 })
        } finally {
            reviewing.value = false
        }
    }

    async function copyPayload() {
        if (!active.value?.payloadJson) return
        try {
            await navigator.clipboard.writeText(active.value.payloadJson)
            toast.add({ severity: 'info', summary: 'Скопировано', detail: 'JSON заявки в буфере обмена', life: 2500 })
        } catch {
            toast.add({ severity: 'warn', summary: 'Не удалось скопировать', detail: 'Выделите JSON вручную', life: 3000 })
        }
    }
</script>

<template>
    <Card class="mb-4">
        <template #title>Модерация заявок</template>

        <template #content>
            <div class="apanel_toolbar">
                <Select v-model="statusFilter" :options="STATUS_OPTIONS" option-label="label" option-value="value"
                    placeholder="Статус" class="w-50" />
                <Button icon="pi pi-refresh" severity="secondary" text @click="load(currentCursor)" />
            </div>

            <Dialog v-model:visible="reviewVisible" modal header="Заявка" style="width: 720px">
                <template v-if="active">
                    <div class="flex flex-col gap-3">
                        <div>
                            <strong>Отправитель:</strong>
                            {{ active.submitterName }} ({{ active.submitterEmail }})
                        </div>

                        <div v-if="active.targetHeroId">
                            <strong>Целевой герой (UUID):</strong> {{ active.targetHeroId }}
                        </div>

                        <div class="flex items-center gap-2">
                            <strong>Данные заявки (payload):</strong>
                            <Button type="button" size="small" text icon="pi pi-copy" label="Копировать"
                                @click="copyPayload" />
                        </div>

                        <div v-if="active.attachmentUrls?.length">
                            <strong>Вложения:</strong>
                            <ul>
                                <li v-for="(url, i) in active.attachmentUrls" :key="i">
                                    <a :href="url" target="_blank" rel="noopener">{{ url }}</a>
                                </li>
                            </ul>
                        </div>

                        <!-- LLM-извлечение: только для заявок на нового героя (нет целевого героя) -->
                        <div v-if="!active.targetHeroId" class="flex flex-col gap-2">
                            <Button icon="pi pi-sparkles" :label="extracting ? 'Извлекаем…' : 'Извлечь через LLM'"
                                :loading="extracting" outlined @click="extractFromSubmission" />
                            <AdminExtractionResultPreview v-if="extractionResult" :data="extractionResult" />
                        </div>

                        <Textarea v-model="moderatorComment" placeholder="Комментарий модератора (виден отправителю)"
                            rows="3" class="w-full" />
                    </div>
                </template>

                <template #footer>
                    <Button v-if="extractionResult" label="Создать героя" icon="pi pi-user-plus" severity="success"
                        @click="createHeroFromExtraction" />

                    <Button v-if="active?.targetHeroId" as="router-link" :to="`/admin/heroes/${active.targetHeroId}`"
                        label="Открыть героя" outlined icon="pi pi-external-link" />

                    <Button label="Отклонить" severity="danger" icon="pi pi-times" :loading="reviewing"
                        @click="review(SubmissionReviewDecision.REJECT)" />
                    <Button label="Одобрить" severity="success" icon="pi pi-check" :loading="reviewing"
                        @click="review(SubmissionReviewDecision.APPROVE)" />
                </template>
            </Dialog>
        </template>
    </Card>

    <DataTable :value="submissions" :loading="loading" striped-rows class="atable" table-style="min-width: 50rem">
        <Column field="submitterName" header="Отправитель">
            <template #body="{ data }">
                <div>{{ data.submitterName }}</div>
                <small class="text-muted">{{ data.submitterEmail }}</small>
            </template>
        </Column>

        <Column header="Цель заявки">
            <template #body="{ data }">
                <NuxtLink v-if="data.targetHeroId" :to="`/admin/heroes/${data.targetHeroId}`">
                    Исправление героя
                </NuxtLink>
                <Tag v-else value="Новый герой" severity="info" />
            </template>
        </Column>

        <Column header="Статус">
            <template #body="{ data }">
                <Tag :value="STATUS_META[data.status as PublicationStatusJson]?.label ?? data.status"
                    :severity="STATUS_META[data.status as PublicationStatusJson]?.severity ?? 'secondary'" />
            </template>
        </Column>

        <Column header="Подана">
            <template #body="{ data }">
                {{ formatDate(data.audit?.createdAt as string | undefined) }}
            </template>
        </Column>

        <Column header="Рассмотрена">
            <template #body="{ data }">
                <span v-if="data.status !== 'PUBLICATION_STATUS_DRAFT'">
                    {{ formatDate(data.audit?.updatedAt as string | undefined) }}
                </span>
                <span v-else class="text-muted">—</span>
            </template>
        </Column>

        <Column header="Комментарий">
            <template #body="{ data }">
                <span v-if="data.moderatorComment" v-tooltip.bottom="data.moderatorComment" class="comment-truncate">
                    {{ data.moderatorComment }}
                </span>
                <span v-else class="text-muted">—</span>
            </template>
        </Column>

        <Column header="" style="width: 240px">
            <template #body="{ data }">
                <div class="flex gap-1">
                    <Button label="Рассмотреть" size="small" @click="openReview(data)" />
                    <Button label="История" size="small" severity="secondary" outlined icon="pi pi-clock"
                        @click="openHistory(data)" />
                </div>
            </template>
        </Column>

        <template #empty>
            <div class="p-4 text-center">Заявок по выбранному статусу нет</div>
        </template>
    </DataTable>

    <footer class="mt-3 flex gap-3">
        <Button outlined label="Назад" icon="pi pi-chevron-left" severity="secondary" :disabled="!hasPrev || loading"
            @click="goPrev" />
        <Button outlined label="Вперёд" icon="pi pi-chevron-right" icon-pos="right" severity="secondary"
            :disabled="!nextCursor || loading" @click="goNext" />
    </footer>

    <AdminSubmissionReviewHistory v-if="historySubmissionId" v-model:visible="historyVisible"
        :submission-id="historySubmissionId" />
</template>

<style scoped>
    .apanel_toolbar {
        display: grid;
        grid-template-columns: repeat(2, max-content);
    }

    .comment-truncate {
        display: inline-block;
        max-width: 220px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        cursor: help;
    }
</style>