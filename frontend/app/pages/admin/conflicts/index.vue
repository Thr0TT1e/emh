<script setup lang="ts">
    import { fromJson } from '@bufbuild/protobuf'
    import { useConfirm } from 'primevue/useconfirm'
    import { useToast } from 'primevue/usetoast'
    import { conflictTypeLabel, formatFlexibleDate } from '~/lib/format'
    import { toPlain } from '~/lib/pb'
    import { FlexibleDateSchema, type FlexibleDateJson } from '~/sdk/emh/v1/common_pb'
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
    // Хелпер: конвертация гибкой даты в protobuf Message перед отправкой.
    // Паттерн идентичен prepareAwardDateForApi из HeroAwards: пустую/неуказанную
    // дату отправляем как UNKNOWN, чтобы бэкенд корректно её обработал.
    // ---------------------------------------------------------------------------
    const prepareDateForApi = (fd?: FlexibleDateJson | null, allowEmpty = false) => {
        if (!fd || fd.precision === 'DATE_PRECISION_UNSPECIFIED') {
            return allowEmpty ? undefined : fromJson(FlexibleDateSchema, { precision: 'DATE_PRECISION_UNKNOWN', displayText: '' });
        }

        return fromJson(FlexibleDateSchema, fd);
    };

    // ---------------------------------------------------------------------------
    // Список + фильтры (ЧТЕНИЕ — публичный сервис)
    // ---------------------------------------------------------------------------
    const conflicts = ref<ConflictJson[]>([])
    const allConflicts = ref<ConflictJson[]>([])
    const loading = ref(false)
    const typeFilter = ref<ConflictType>(ConflictType.UNSPECIFIED)
    const searchQuery = ref('')

    async function load() {
        loading.value = true
        try {
            const res = await conflict.listConflicts({ type: typeFilter.value })
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

    // «1941 — 1945» или «2022 — н.в.» — через гибкие даты
    const periodLabel = (c: ConflictJson) => {
        const start = formatFlexibleDate(c.startDateInfo, c.startDate)
        const end = (c.endDateInfo || c.endDate)
            ? formatFlexibleDate(c.endDateInfo, c.endDate)
            : 'н.в.'
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
        startDateInfo: undefined as FlexibleDateJson | undefined,
        endDateInfo: undefined as FlexibleDateJson | undefined,
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
            startDateInfo: undefined, endDateInfo: undefined, parentConflictId: '',
        })
        dialogVisible.value = true
    }

    function openEdit(c: ConflictJson) {
        editingId.value = c.id ?? null
        Object.assign(form, {
            name: c.name ?? '',
            description: c.description ?? '',
            type: TYPE_JSON_TO_NUM[c.type ?? 'CONFLICT_TYPE_UNSPECIFIED'] ?? ConflictType.UNSPECIFIED,
            startDateInfo: c.startDateInfo ?? undefined,
            endDateInfo: c.endDateInfo ?? undefined,
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
        if (!form.startDateInfo || form.startDateInfo.precision === 'DATE_PRECISION_UNSPECIFIED') {
            toast.add({ severity: 'warn', summary: 'Проверьте форму', detail: 'Дата начала обязательна', life: 3000 })
            return
        }
        busy.value = true
        try {
            const payload = {
                name: form.name,
                description: form.description,
                type: form.type,
                // startDate/endDate оставляем пустыми — приоритет у *_date_info (контракт)
                startDate: '',
                endDate: '',
                startDateInfo: prepareDateForApi(form.startDateInfo),
                endDateInfo: prepareDateForApi(form.endDateInfo, true),
                parentConflictId: form.parentConflictId,
            }
            if (editingId.value) {
                await conflictAdmin.updateConflict({
                    id: editingId.value,
                    ...payload,
                    fieldMask: [
                        'name',
                        'description',
                        'type',
                        'start_date',
                        'end_date',
                        'parent_conflict_id',
                        'start_date_info',
                        'end_date_info',
                    ],
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
                    <label>
                        <span class="afield__label">Тип *</span>
                        <Select v-model="form.type" :options="TYPE_OPTIONS.slice(1)" option-label="label"
                            option-value="value" class="w-full" />
                    </label>
                    <!-- Гибкая дата начала (обязательна) -->
                    <AdminFlexibleDateInput v-model="form.startDateInfo" label="Дата начала *" />
                    <!-- Гибкая дата окончания (пусто = продолжается) -->
                    <AdminFlexibleDateInput v-model="form.endDateInfo" label="Дата окончания (пусто = продолжается)" />
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