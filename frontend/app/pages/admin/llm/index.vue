<script setup lang="ts">
    import { useConfirm } from 'primevue/useconfirm';
    import { useToast } from 'primevue/usetoast';
    import { isPermissionDenied, isRateLimited } from '~/lib/errors';
    import { formatDate } from '~/lib/format';
    import { toPlain } from '~/lib/pb';
    import type { LlmProviderInfoJson } from '~/sdk/emh/v1/llm_admin_pb';
    import { ListLlmProvidersResponseSchema } from '~/sdk/emh/v1/llm_admin_pb';

    definePageMeta({ layout: 'admin', middleware: 'admin' });
    useHead({ title: 'LLM-провайдеры — канцелярия' });

    const { llmAdmin } = useApi();
    const toast = useToast();
    const confirm = useConfirm();

    // ── Список ────────────────────────────────────────────────────
    const providers = ref<LlmProviderInfoJson[]>([]);
    const loading = ref(false);

    async function load() {
        loading.value = true;
        try {
            const res = await llmAdmin.listLlmProviders({});
            providers.value = toPlain(ListLlmProvidersResponseSchema, res).providers ?? [];
        } catch (e) {
            showError(e);
        } finally {
            loading.value = false;
        }
    }
    onMounted(() => load());

    // ── Активация ────────────────────────────────────────────────
    function activateProvider(p: LlmProviderInfoJson) {
        confirm.require({
            message: `Сделать «${p.name}» активным провайдером? Остальные будут деактивированы.`,
            header: 'Активация провайдера',
            rejectLabel: 'Отмена',
            acceptLabel: 'Активировать',
            acceptClass: 'p-button-success',
            accept: async () => {
                try {
                    await llmAdmin.setActiveLlmProvider({ id: p.id });
                    toast.add({ severity: 'success', summary: 'Активирован', detail: `Провайдер «${p.name}» теперь активен`, life: 3000 });
                    await load();
                } catch (e) {
                    showError(e);
                }
            },
        });
    }

    // ── Тестирование ─────────────────────────────────────────────
    // UI-представление результата теста: bigint → number для отображения.
    // llmAdmin.testLlmProvider возвращает Message, где processingTimeMs: bigint (int64).
    interface TestResultUi {
        success: boolean;
        rawResponse: string;
        processingTimeMs: number;
        errorMessage: string;
    }

    const testVisible = ref(false);
    const testing = ref(false);
    const testTarget = ref<LlmProviderInfoJson | null>(null);
    const testPrompt = ref('Ответь одним словом: ты работаешь?');
    const testResult = ref<TestResultUi | null>(null);

    function openTest(p: LlmProviderInfoJson) {
        testTarget.value = p;
        testResult.value = null;
        testVisible.value = true;
    }

    async function runTest() {
        if (!testTarget.value || testing.value || !testPrompt.value.trim()) return;
        testing.value = true;
        testResult.value = null;
        try {
            const res = await llmAdmin.testLlmProvider({
                id: testTarget.value.id,
                testPrompt: testPrompt.value.trim(),
            });
            testResult.value = {
                success: res.success,
                rawResponse: res.rawResponse,
                // Number(bigint) — безопасная конвертация int64 → number
                processingTimeMs: Number(res.processingTimeMs ?? 0n),
                errorMessage: res.errorMessage,
            };
        } catch (e) {
            showError(e);
        } finally {
            testing.value = false;
        }
    }

    // ── Редактирование (приоритет + заметки) ─────────────────────
    const editVisible = ref(false);
    const saving = ref(false);
    const editForm = reactive({ id: '', name: '', priority: 0, notes: '' });

    function openEdit(p: LlmProviderInfoJson) {
        editForm.id = p.id!;
        editForm.name = p.name!;
        editForm.priority = p.priority ?? 0;
        editForm.notes = p.notes ?? '';
        editVisible.value = true;
    }

    async function saveEdit() {
        if (saving.value) return;
        saving.value = true;
        try {
            await llmAdmin.updateLlmProvider({
                id: editForm.id,
                priority: editForm.priority,
                notes: editForm.notes,
                fieldMask: ['priority', 'notes'],
            });
            toast.add({ severity: 'success', summary: 'Сохранено', detail: `Провайдер «${editForm.name}» обновлён`, life: 3000 });
            editVisible.value = false;
            await load();
        } catch (e) {
            showError(e);
        } finally {
            saving.value = false;
        }
    }

    // ── Единая обработка ошибок (403 / 429 / прочее) ─────────────
    function showError(e: unknown) {
        if (isPermissionDenied(e)) {
            toast.add({ severity: 'error', summary: 'Нет прав доступа', detail: 'Требуется роль администратора.', life: 5000 });
            return;
        }
        if (isRateLimited(e)) {
            toast.add({ severity: 'error', summary: 'Превышен лимит запросов', detail: 'Попробуйте через минуту.', life: 5000 });
            return;
        }
        const msg = e instanceof Error ? e.message : 'Не удалось выполнить операцию';
        toast.add({ severity: 'error', summary: 'Ошибка', detail: msg, life: 5000 });
    }
</script>

<template>
    <Card class="mb-4">
        <template #title>LLM-провайдеры</template>
        <template #subtitle>Управление провайдерами для извлечения данных: активация, приоритет, тестирование</template>
        <template #content>
            <Button icon="pi pi-refresh" severity="secondary" text label="Обновить" @click="load" />
        </template>
    </Card>

    <DataTable :value="providers" :loading="loading" striped-rows class="atable" table-style="min-width: 60rem">
        <Column header="Провайдер">
            <template #body="{ data }">
                <div class="flex items-center gap-2">
                    <span class="font-semibold">{{ data.name }}</span>
                    <Tag v-if="data.isActive" value="активен" severity="success" />
                </div>
            </template>
        </Column>
        <Column field="type" header="Тип" />
        <Column field="model" header="Модель" />
        <Column header="Приоритет" style="width: 100px">
            <template #body="{ data }">{{ data.priority }}</template>
        </Column>
        <Column header="Заметки">
            <template #body="{ data }">
                <span v-if="data.notes">{{ data.notes }}</span>
                <span v-else class="text-muted">—</span>
            </template>
        </Column>
        <Column header="Обновлён" style="width: 140px">
            <template #body="{ data }">{{ formatDate(data.updatedAt) }}</template>
        </Column>
        <Column header="" style="width: 280px">
            <template #body="{ data }">
                <div class="flex justify-end gap-1 whitespace-nowrap">
                    <Button v-if="!data.isActive" size="small" severity="success" outlined label="Активировать"
                        @click="activateProvider(data)" />
                    <Button size="small" outlined icon="pi pi-play" label="Тест" @click="openTest(data)" />
                    <Button size="small" text icon="pi pi-pencil" label="Изменить" @click="openEdit(data)" />
                </div>
            </template>
        </Column>
        <template #empty>
            <div class="p-4 text-center">Провайдеры не настроены</div>
        </template>
    </DataTable>

    <!-- Диалог тестирования -->
    <Dialog v-model:visible="testVisible" modal :header="`Тест: ${testTarget?.name ?? ''}`" style="width: 640px">
        <div class="flex flex-col gap-3">
            <label>
                <span class="afield__label">Тестовый промпт</span>
                <Textarea v-model="testPrompt" rows="3" class="w-full"
                    placeholder="Введите запрос для проверки модели…" />
            </label>
            <Button :loading="testing" :disabled="!testPrompt.trim()" icon="pi pi-play"
                :label="testing ? 'Отправляем…' : 'Запустить тест'" @click="runTest" />

            <template v-if="testResult">
                <Message v-if="testResult.success" severity="success" :closable="false">
                    Успех · {{ testResult.processingTimeMs }} мс
                </Message>
                <Message v-else severity="error" :closable="false">
                    {{ testResult.errorMessage || 'Модель не ответила' }}
                </Message>
                <div v-if="testResult.success && testResult.rawResponse" class="test-response">
                    <span class="afield__label">Ответ модели:</span>
                    <pre>{{ testResult.rawResponse }}</pre>
                </div>
            </template>
        </div>
        <template #footer>
            <Button outlined label="Закрыть" @click="testVisible = false" />
        </template>
    </Dialog>

    <!-- Диалог редактирования -->
    <Dialog v-model:visible="editVisible" modal :header="`Провайдер: ${editForm.name}`" style="width: 480px">
        <div class="flex flex-col gap-3">
            <label>
                <span class="afield__label">Приоритет (для fallback-логики)</span>
                <InputNumber v-model="editForm.priority" class="w-full" :min="0" :max="100" />
            </label>
            <label>
                <span class="afield__label">Заметки</span>
                <Textarea v-model="editForm.notes" rows="3" class="w-full" placeholder="Комментарий администратора…" />
            </label>
        </div>
        <template #footer>
            <Button outlined label="Отмена" @click="editVisible = false" />
            <Button :loading="saving" label="Сохранить" @click="saveEdit" />
        </template>
    </Dialog>
</template>

<style scoped>
    .test-response {
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }

    .test-response pre {
        margin: 0;
        padding: 0.8rem;
        border: 1px solid var(--emh-line);
        background: var(--emh-bg);
        font-size: 0.85rem;
        white-space: pre-wrap;
        word-break: break-word;
        max-height: 240px;
        overflow: auto;
    }
</style>