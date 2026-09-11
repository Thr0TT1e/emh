<script setup lang="ts">
    /**
     * /admin/conflicts — CRUD справочника конфликтов.
     * Чтение — публичный ConflictService (conflict), мутации — ConflictAdminService (conflictAdmin).
     * Особенности:
     *  - в мутациях даты — строки YYYY-MM-DD (не Timestamp);
     *  - search_query в контракте нет → текстовый фильтр клиентский;
     *  - endDate пустая = конфликт продолжается («н.в.»).
     */
    import { useConfirm } from 'primevue/useconfirm'
    import { useToast } from 'primevue/usetoast'
    import { conflictTypeLabel, formatYear } from '~/lib/format'
    import { toPlain } from '~/lib/pb'
    import type { ConflictJson } from '~/sdk/emh/v1/conflict_pb'
    import { ConflictSchema } from '~/sdk/emh/v1/conflict_pb'
    import type { ConflictTypeJson } from '~/sdk/emh/v1/enums_emh_pb'
    import { ConflictType } from '~/sdk/emh/v1/enums_emh_pb'

    definePageMeta({ layout: 'admin', middleware: 'admin' })
    useHead({ title: 'Конфликты — Вечная память героям' })

    const { conflict, conflictAdmin } = useApi()
    const toast = useToast()
    const confirm = useConfirm()

    // ---------------------------------------------------------------------------
    // Справочные мапы
    // ---------------------------------------------------------------------------
    const TYPE_OPTIONS = [
        { label: 'Все типы', value: ConflictType.UNSPECIFIED },
        { label: 'Мировая война', value: ConflictType.GLOBAL },
        { label: 'Локальный конфликт', value: ConflictType.LOCAL },
        { label: 'Миротворческая операция', value: ConflictType.PEACEKEEPING },
        { label: 'КТО', value: ConflictType.COUNTER_TERRORISM },
        { label: 'СВО', value: ConflictType.SPECIAL_OPERATION },
    ]

    const TYPE_SEVERITY: Record<ConflictTypeJson, 'info' | 'success' | 'warn' | 'danger' | 'secondary'> = {
        CONFLICT_TYPE_UNSPECIFIED: 'secondary',
        CONFLICT_TYPE_GLOBAL: 'danger',
        CONFLICT_TYPE_LOCAL: 'warn',
        CONFLICT_TYPE_PEACEKEEPING: 'info',
        CONFLICT_TYPE_COUNTER_TERRORISM: 'warn',
        CONFLICT_TYPE_SPECIAL_OPERATION: 'success',
    }

    // JSON-имя enum → число (toEnum полные имена не резолвит)
    const TYPE_JSON_TO_NUM: Record<ConflictTypeJson, ConflictType> = {
        CONFLICT_TYPE_UNSPECIFIED: ConflictType.UNSPECIFIED,
        CONFLICT_TYPE_GLOBAL: ConflictType.GLOBAL,
        CONFLICT_TYPE_LOCAL: ConflictType.LOCAL,
        CONFLICT_TYPE_PEACEKEEPING: ConflictType.PEACEKEEPING,
        CONFLICT_TYPE_COUNTER_TERRORISM: ConflictType.COUNTER_TERRORISM,
        CONFLICT_TYPE_SPECIAL_OPERATION: ConflictType.SPECIAL_OPERATION,
    }

    const typeLabel = (t?: ConflictTypeJson) =>
        !t || t === 'CONFLICT_TYPE_UNSPECIFIED' ? '—' : conflictTypeLabel(t)

    // ---------------------------------------------------------------------------
    // Список + фильтры (ЧТЕНИЕ — публичный сервис)
    // ---------------------------------------------------------------------------
    const conflicts = ref<ConflictJson[]>([])      // серверный список (фильтр по типу)
    const allConflicts = ref<ConflictJson[]>([])   // полный список: lookup родителей + селекты
    const loading = ref(false)
    const typeFilter = ref<ConflictType>(ConflictType.UNSPECIFIED)
    const searchQuery = ref('')                    // клиентский поиск (в RPC его нет)

    async function load() {
        loading.value = true
        try {
            const res = await conflict.listConflicts({ type: typeFilter.value }) // 0 = все типы
            conflicts.value = (res.conflicts ?? []).map((c) => toPlain(ConflictSchema, c))
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось загрузить конфликты', life: 5000 })
        } finally {
            loading.value = false
        }
    }

    async function loadAll() {
        try {
            const res = await conflict.listConflicts({})
            allConflicts.value = (res.conflicts ?? []).map((c) => toPlain(ConflictSchema, c))
        } catch { /* некритично: селекты родителей будут пустыми */ }
    }

    // Клиентский поиск + сортировка по дате начала (новые сверху)
    const visibleConflicts = computed(() => {
        const q = searchQuery.value.trim().toLowerCase()
        const list = q
            ? conflicts.value.filter((c) => (c.name ?? '').toLowerCase().includes(q))
            : [...conflicts.value]
        return list.sort((a, b) => (b.startDate ?? '').localeCompare(a.startDate ?? ''))
    })

    const parentName = (id?: string) =>
        id ? allConflicts.value.find((c) => c.id === id)?.name ?? id : '—'

    const hasChildren = (id?: string) =>
        !!id && allConflicts.value.some((c) => c.parentConflictId === id)

    // «1941 — 1945» или «2022 — н.в.»
    const periodLabel = (c: ConflictJson) => {
        const start = formatYear(c.startDate)
        const end = c.endDate ? formatYear(c.endDate) : 'н.в.'
        return `${start} — ${end}`
    }

    // Потомки (защита от циклов в иерархии)
    function descendantsOf(id: string): Set<string> {
        const result = new Set<string>()
        const walk = (pid: string) => {
            for (const c of allConflicts.value) {
                if (c.parentConflictId === pid && c.id && !result.has(c.id)) {
                    result.add(c.id)
                    walk(c.id)
                }
            }
        }
        walk(id)
        return result
    }

    watch(typeFilter, load)
    onMounted(() => { load(); loadAll() })

    // ---------------------------------------------------------------------------
    // Модалка создания/редактирования (МУТАЦИИ — админский сервис)
    // ---------------------------------------------------------------------------
    const dialogVisible = ref(false)
    const busy = ref(false)
    const editingId = ref<string | null>(null)

    const form = reactive({
        name: '',
        description: '',
        type: ConflictType.UNSPECIFIED as ConflictType,
        startDate: '', // YYYY-MM-DD, обязательна
        endDate: '',   // пусто = конфликт продолжается
        parentConflictId: '',
    })

    const parentOptions = computed(() => {
        const forbidden = editingId.value ? descendantsOf(editingId.value) : new Set<string>()
        return allConflicts.value
            .filter((c) => c.id !== editingId.value && !forbidden.has(c.id!))
            .map((c) => ({ label: `${c.name} (${periodLabel(c)})`, value: c.id }))
    })

    function openCreate() {
        editingId.value = null
        Object.assign(form, {
            name: '', description: '', type: ConflictType.UNSPECIFIED,
            startDate: '', endDate: '', parentConflictId: '',
        })
        dialogVisible.value = true
    }

    function openEdit(c: ConflictJson) {
        editingId.value = c.id ?? null
        Object.assign(form, {
            name: c.name ?? '',
            description: c.description ?? '',
            type: TYPE_JSON_TO_NUM[c.type ?? 'CONFLICT_TYPE_UNSPECIFIED'] ?? ConflictType.UNSPECIFIED,
            // TimestampJson → YYYY-MM-DD для input[type=date] (паттерн HeroForm.vue)
            startDate: c.startDate?.slice(0, 10) ?? '',
            endDate: c.endDate?.slice(0, 10) ?? '',
            parentConflictId: c.parentConflictId ?? '',
        })
        dialogVisible.value = true
    }

    async function save() {
        if (!form.name.trim()) {
            toast.add({ severity: 'warn', summary: 'Проверьте форму', detail: 'Название обязательно', life: 3000 })
            return
        }
        if (form.type === ConflictType.UNSPECIFIED) {
            toast.add({ severity: 'warn', summary: 'Проверьте форму', detail: 'Выберите тип конфликта', life: 3000 })
            return
        }
        if (!form.startDate) {
            toast.add({ severity: 'warn', summary: 'Проверьте форму', detail: 'Дата начала обязательна', life: 3000 })
            return
        }
        busy.value = true
        try {
            const payload = {
                name: form.name,
                description: form.description,
                type: form.type,
                startDate: form.startDate,
                endDate: form.endDate, // пустая строка валидна: текущий конфликт
                parentConflictId: form.parentConflictId,
            }
            if (editingId.value) {
                await conflictAdmin.updateConflict({
                    id: editingId.value,
                    ...payload,
                    fieldMask: ['name', 'description', 'type', 'start_date', 'end_date', 'parent_conflict_id'],
                })
                toast.add({ severity: 'success', summary: 'Сохранено', detail: form.name, life: 3000 })
            } else {
                await conflictAdmin.createConflict(payload)
                toast.add({ severity: 'success', summary: 'Создано', detail: form.name, life: 3000 })
            }
            dialogVisible.value = false
            await Promise.all([load(), loadAll()])
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось сохранить', life: 5000 })
        } finally {
            busy.value = false
        }
    }

    // ---------------------------------------------------------------------------
    // Удаление
    // ---------------------------------------------------------------------------
    function remove(c: ConflictJson) {
        confirm.require({
            message: hasChildren(c.id)
                ? `У конфликта «${c.name}» есть дочерние операции. Если бэкенд требует их предварительного удаления — сначала удалите их.`
                : `Удалить конфликт «${c.name}»?`,
            header: 'Удаление конфликта',
            icon: 'pi pi-exclamation-triangle',
            acceptLabel: 'Удалить',
            rejectLabel: 'Отмена',
            acceptClass: 'p-button-danger',
            accept: () => doDelete(c),
        })
    }

    async function doDelete(c: ConflictJson) {
        try {
            const res = await conflictAdmin.deleteConflict({ id: c.id ?? '' })
            if (res.success) {
                toast.add({ severity: 'success', summary: 'Удалено', detail: c.name, life: 3000 })
                await Promise.all([load(), loadAll()])
            }
        } catch (e: any) {
            // «При наличии связанных героев вернёт ошибку» — показываем текст бэкенда
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось удалить', life: 5000 })
        }
    }
</script>

<template>
    <Card>
        <template #title>Справочник конфликтов</template>

        <template #content>
            <div class="apanel__toolbar">
                <InputText v-model="searchQuery" placeholder="Поиск по названию…" />
                <Select v-model="typeFilter" :options="TYPE_OPTIONS" option-label="label" option-value="value" />
                <Button label="Добавить конфликт" icon="pi pi-plus" @click="openCreate" />
            </div>

            <!-- Создание / редактирование -->
            <Dialog v-model:visible="dialogVisible" modal
                :header="editingId ? 'Редактирование конфликта' : 'Новый конфликт'" style="width: 680px">
                <div class="flex flex-col gap-3">
                    <label>
                        <span class="afield__label">Название *</span>
                        <InputText v-model="form.name" class="w-full"
                            placeholder="Например: Великая Отечественная война" />
                    </label>

                    <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:0 1.2rem">
                        <label>
                            <span class="afield__label">Тип *</span>
                            <Select v-model="form.type" :options="TYPE_OPTIONS.slice(1)" option-label="label"
                                option-value="value" class="w-full" />
                        </label>

                        <label>
                            <span class="afield__label">Дата начала *</span>
                            <input v-model="form.startDate" type="date" class="afield__input" />
                        </label>

                        <label>
                            <span class="afield__label">Дата окончания</span>
                            <input v-model="form.endDate" type="date" class="afield__input" />
                            <small class="text-muted" style="font-size:.72rem">пусто = продолжается</small>
                        </label>
                    </div>

                    <label>
                        <span class="afield__label">Родительский конфликт (операция внутри войны)</span>
                        <Select v-model="form.parentConflictId" :options="parentOptions" option-label="label"
                            option-value="value" filter show-clear class="w-full" />
                    </label>

                    <label>
                        <span class="afield__label">Историческое описание</span>
                        <Textarea v-model="form.description" rows="4" class="w-full" />
                    </label>
                </div>

                <template #footer>
                    <Button label="Отмена" outlined severity="secondary" @click="dialogVisible = false" />
                    <Button :label="editingId ? 'Сохранить' : 'Создать'" :loading="busy" @click="save" />
                </template>
            </Dialog>
        </template>
    </Card>

    <DataTable :value="visibleConflicts" :loading="loading" striped-rows class="atable mt-4"
        table-style="min-width: 50rem">
        <Column field="name" header="Название">
            <template #body="{ data }">
                {{ data.name }}
                <i v-if="hasChildren(data.id)" class="pi pi-sitemap text-muted" title="Есть дочерние операции"
                    style="font-size:.7rem;margin-left:.35rem" />
            </template>
        </Column>

        <Column header="Тип">
            <template #body="{ data }">
                <Tag :value="typeLabel(data.type)"
                    :severity="TYPE_SEVERITY[data.type as ConflictTypeJson] ?? 'secondary'" />
            </template>
        </Column>

        <Column header="Период">
            <template #body="{ data }">
                {{ periodLabel(data) }}
            </template>
        </Column>

        <Column header="Родитель">
            <template #body="{ data }">
                {{ parentName(data.parentConflictId) }}
            </template>
        </Column>

        <Column header="" style="width: 150px">
            <template #body="{ data }">
                <Button icon="pi pi-pencil" text severity="secondary" title="Редактировать" @click="openEdit(data)" />
                <Button icon="pi pi-trash" text severity="danger" title="Удалить" @click="remove(data)" />
            </template>
        </Column>

        <template #empty>
            <div class="p-4 text-center">Конфликтов не найдено</div>
        </template>
    </DataTable>
</template>