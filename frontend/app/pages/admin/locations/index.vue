<script setup lang="ts">
    /**
     * /admin/locations — CRUD справочника локаций.
     * Разделение по контракту: чтение — публичный LocationService (location),
     * мутации — LocationAdminService (locationAdmin).
     * Удаление локации с детьми требует сначала ReparentLocations.
     */
    import { useConfirm } from 'primevue/useconfirm'
    import { useToast } from 'primevue/usetoast'
    import LocationMap from '~/components/admin/LocationMap.vue'
    import { toPlain } from '~/lib/pb'
    import type { LocationTypeJson } from '~/sdk/emh/v1/enums_emh_pb'
    import { LocationType } from '~/sdk/emh/v1/enums_emh_pb'
    import type { LocationJson } from '~/sdk/emh/v1/location_pb'
    import { LocationSchema } from '~/sdk/emh/v1/location_pb'
    import EditIcon from '~icons/carbon/edit?width=1.25em&height=1.25em'
    import LocationIcon from '~icons/carbon/location?width=1.25em&height=1.25em'
    import TrashCanIcon from '~icons/carbon/trash-can?width=1.25em&height=1.25em'
    import TreeViewIcon from '~icons/carbon/tree-view?width=1.25em&height=1.25em'

    definePageMeta({ layout: 'admin', middleware: 'admin' })
    useHead({ title: 'Локации — Вечная память героям' })

    const { location, locationAdmin } = useApi()
    const toast = useToast()
    const confirm = useConfirm()

    // ---------------------------------------------------------------------------
    // Справочные мапы
    // ---------------------------------------------------------------------------
    const TYPE_OPTIONS = [
        { label: 'Все типы', value: LocationType.UNSPECIFIED },
        { label: 'Страна', value: LocationType.COUNTRY },
        { label: 'Регион', value: LocationType.REGION },
        { label: 'Город', value: LocationType.CITY },
        { label: 'Село / посёлок', value: LocationType.VILLAGE },
        { label: 'Кладбище / мемориал', value: LocationType.CEMETERY },
    ]

    const TYPE_META: Record<LocationTypeJson, { label: string; severity: 'info' | 'success' | 'warn' | 'secondary' }> = {
        LOCATION_TYPE_UNSPECIFIED: { label: '—', severity: 'secondary' },
        LOCATION_TYPE_COUNTRY: { label: 'Страна', severity: 'info' },
        LOCATION_TYPE_REGION: { label: 'Регион', severity: 'info' },
        LOCATION_TYPE_CITY: { label: 'Город', severity: 'success' },
        LOCATION_TYPE_VILLAGE: { label: 'Село / посёлок', severity: 'success' },
        LOCATION_TYPE_CEMETERY: { label: 'Кладбище / мемориал', severity: 'warn' },
    }

    // JSON-имя enum → число (toEnum полные имена не резолвит)
    const TYPE_JSON_TO_NUM: Record<LocationTypeJson, LocationType> = {
        LOCATION_TYPE_UNSPECIFIED: LocationType.UNSPECIFIED,
        LOCATION_TYPE_COUNTRY: LocationType.COUNTRY,
        LOCATION_TYPE_REGION: LocationType.REGION,
        LOCATION_TYPE_CITY: LocationType.CITY,
        LOCATION_TYPE_VILLAGE: LocationType.VILLAGE,
        LOCATION_TYPE_CEMETERY: LocationType.CEMETERY,
    }

    // В JSON double может прийти строкой "NaN"/"Infinity" — нормализуем.
    const num = (v: number | string | undefined, fallback = 0): number => {
        const n = Number(v ?? fallback)
        return Number.isFinite(n) ? n : fallback
    }

    // ---------------------------------------------------------------------------
    // Список + фильтры (ЧТЕНИЕ — публичный сервис)
    // ---------------------------------------------------------------------------
    const locations = ref<LocationJson[]>([])      // отфильтрованный список для таблицы
    const allLocations = ref<LocationJson[]>([])   // полный список: lookup родителей + селекты
    const loading = ref(false)
    const searchQuery = ref('')
    const typeFilter = ref<LocationType>(LocationType.UNSPECIFIED)

    async function load() {
        loading.value = true
        try {
            const res = await location.listLocations({
                searchQuery: searchQuery.value,
                type: typeFilter.value, // 0 = все типы
            })
            locations.value = (res.locations ?? []).map((l) => toPlain(LocationSchema, l))
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось загрузить локации', life: 5000 })
        } finally {
            loading.value = false
        }
    }

    async function loadAll() {
        try {
            const res = await location.listLocations({})
            allLocations.value = (res.locations ?? []).map((l) => toPlain(LocationSchema, l))
        } catch { /* некритично: селекты родителей просто будут пустыми */ }
    }

    const parentName = (id?: string) =>
        id ? allLocations.value.find((l) => l.id === id)?.name ?? id : '—'

    const fmtCoords = (l: LocationJson) => {
        const la = num(l.latitude)
        const lo = num(l.longitude)
        return la || lo ? `${la.toFixed(4)}, ${lo.toFixed(4)}` : '—'
    }

    // Потомки (для защиты от циклов в иерархии)
    function descendantsOf(id: string): Set<string> {
        const result = new Set<string>()
        const walk = (pid: string) => {
            for (const l of allLocations.value) {
                if (l.parentId === pid && l.id && !result.has(l.id)) {
                    result.add(l.id)
                    walk(l.id)
                }
            }
        }
        walk(id)
        return result
    }

    const hasChildren = (id?: string) => !!id && allLocations.value.some((l) => l.parentId === id)

    let searchTimer: ReturnType<typeof setTimeout> | undefined
    watch(searchQuery, () => {
        clearTimeout(searchTimer)
        searchTimer = setTimeout(load, 400)
    })
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
        historicalName: '',
        type: LocationType.UNSPECIFIED as LocationType,
        parentId: '',
        latitude: 0,
        longitude: 0,
    })

    const parentOptions = computed(() => {
        const forbidden = editingId.value ? descendantsOf(editingId.value) : new Set<string>()
        return allLocations.value
            .filter((l) => l.id !== editingId.value && !forbidden.has(l.id!))
            .map((l) => ({ label: `${l.name} — ${TYPE_META[l.type!]?.label ?? '—'}`, value: l.id }))
    })

    function openCreate() {
        editingId.value = null
        Object.assign(form, {
            name: '', historicalName: '', type: LocationType.UNSPECIFIED,
            parentId: '', latitude: 0, longitude: 0,
        })
        dialogVisible.value = true
    }

    function openEdit(loc: LocationJson) {
        editingId.value = loc.id ?? null
        Object.assign(form, {
            name: loc.name ?? '',
            historicalName: loc.historicalName ?? '',
            type: TYPE_JSON_TO_NUM[loc.type ?? 'LOCATION_TYPE_UNSPECIFIED'] ?? LocationType.UNSPECIFIED,
            parentId: loc.parentId ?? '',
            latitude: num(loc.latitude),
            longitude: num(loc.longitude),
        })
        dialogVisible.value = true
    }

    function onPickCoords(lat: number, lon: number) {
        form.latitude = lat
        form.longitude = lon
    }

    async function save() {
        if (!form.name.trim()) {
            toast.add({ severity: 'warn', summary: 'Проверьте форму', detail: 'Название обязательно', life: 3000 })
            return
        }
        if (form.type === LocationType.UNSPECIFIED) {
            toast.add({ severity: 'warn', summary: 'Проверьте форму', detail: 'Выберите тип локации', life: 3000 })
            return
        }
        busy.value = true
        try {
            const payload = {
                name: form.name,
                historicalName: form.historicalName,
                type: form.type,
                parentId: form.parentId,
                latitude: form.latitude,
                longitude: form.longitude,
            }
            if (editingId.value) {
                await locationAdmin.updateLocation({
                    id: editingId.value,
                    ...payload,
                    fieldMask: ['name', 'historical_name', 'type', 'parent_id', 'latitude', 'longitude'],
                })
                toast.add({ severity: 'success', summary: 'Сохранено', detail: form.name, life: 3000 })
            } else {
                await locationAdmin.createLocation(payload)
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
    // Удаление с проверкой детей + ReParent
    // ---------------------------------------------------------------------------
    const reparentVisible = ref(false)
    const reparentBusy = ref(false)
    const reparentSource = ref<LocationJson | null>(null)
    const reparentTarget = ref('')
    const reparentAndDelete = ref(false)

    const reparentTargets = computed(() => {
        if (!reparentSource.value?.id) return []
        const forbidden = descendantsOf(reparentSource.value.id)
        return allLocations.value
            .filter((l) => l.id !== reparentSource.value?.id && !forbidden.has(l.id!))
            .map((l) => ({ label: `${l.name} — ${TYPE_META[l.type!]?.label ?? '—'}`, value: l.id }))
    })

    function openReparent(loc: LocationJson, andDelete = false) {
        reparentSource.value = loc
        reparentTarget.value = ''
        reparentAndDelete.value = andDelete
        reparentVisible.value = true
    }

    async function doReparent() {
        if (!reparentSource.value?.id || !reparentTarget.value || reparentBusy.value) return
        reparentBusy.value = true
        try {
            const res = await locationAdmin.reparentLocations({
                oldParentId: reparentSource.value.id,
                newParentId: reparentTarget.value,
            })
            toast.add({
                severity: 'success',
                summary: 'Перепривязано',
                detail: `Перенесено дочерних локаций: ${res.affectedCount}`,
                life: 4000,
            })
            const source = reparentSource.value
            reparentVisible.value = false
            await Promise.all([load(), loadAll()])
            if (reparentAndDelete.value) doDelete(source) // теперь детей нет — можно удалять
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось перепривязать', life: 5000 })
        } finally {
            reparentBusy.value = false
        }
    }

    function remove(loc: LocationJson) {
        if (hasChildren(loc.id)) {
            // Бэкенд: «удаляет локацию (требует предварительной репривязки детей)»
            confirm.require({
                message: `У локации «${loc.name}» есть дочерние. Перед удалением их нужно перепривязать к другому родителю.`,
                header: 'Есть дочерние локации',
                icon: 'pi pi-sitemap',
                acceptLabel: 'Перепривязать и удалить',
                rejectLabel: 'Отмена',
                acceptClass: 'p-button-danger',
                accept: () => openReparent(loc, true),
            })

            return
        }
        confirm.require({
            message: `Удалить локацию «${loc.name}»?`,
            header: 'Удаление локации',
            icon: 'pi pi-exclamation-triangle',
            acceptLabel: 'Удалить',
            rejectLabel: 'Отмена',
            acceptClass: 'p-button-danger',
            accept: () => doDelete(loc),
        })
    }

    async function doDelete(loc: LocationJson) {
        try {
            const res = await locationAdmin.deleteLocation({ id: loc.id ?? '' })
            if (res.success) {
                toast.add({ severity: 'success', summary: 'Удалено', detail: loc.name, life: 3000 })
                await Promise.all([load(), loadAll()])
            }
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось удалить', life: 5000 })
        }
    }
</script>

<template>
    <Card>
        <template #title>Справочник локаций</template>

        <template #content>
            <div class="apanel__toolbar">
                <InputText v-model="searchQuery" placeholder="Поиск по названию…" />

                <Select v-model="typeFilter" :options="TYPE_OPTIONS" option-label="label" option-value="value" />

                <Button @click="openCreate">
                    <LocationIcon />
                    <span>Добавить локацию</span>
                </Button>
            </div>

            <!-- Создание / редактирование -->
            <Dialog v-model:visible="dialogVisible" modal
                :header="editingId ? 'Редактирование локации' : 'Новая локация'" style="width: 780px">
                <div class="flex flex-col gap-3">
                    <div style="display:grid;grid-template-columns:1fr 1fr;gap:0 1.2rem">
                        <label>
                            <span class="afield__label">Название *</span>
                            <InputText v-model="form.name" class="w-full" />
                        </label>

                        <label>
                            <span class="afield__label">Историческое название</span>
                            <InputText v-model="form.historicalName" class="w-full" />
                        </label>
                    </div>

                    <div style="display:grid;grid-template-columns:1fr 1.4fr 1.6fr;gap:0 1.2rem;align-items:end">
                        <label>
                            <span class="afield__label">Тип *</span>
                            <Select v-model="form.type" :options="TYPE_OPTIONS.slice(1)" option-label="label"
                                option-value="value" class="w-full" />
                        </label>

                        <label>
                            <span class="afield__label">Родитель (страна / регион)</span>
                            <Select v-model="form.parentId" :options="parentOptions" option-label="label"
                                option-value="value" filter show-clear class="w-full" />
                        </label>

                        <div>
                            <span class="afield__label">Широта / долгота</span>
                            <div style="display:flex;gap:.5rem">
                                <InputNumber v-model="form.latitude" :min="-89.999999" :max="89.999999"
                                    :max-fraction-digits="6" placeholder="Широта" />
                                <InputNumber v-model="form.longitude" :min="-179.999999" :max="179.999999"
                                    :max-fraction-digits="6" placeholder="Долгота" />
                            </div>
                        </div>
                    </div>

                    <LocationMap :lat="form.latitude" :lon="form.longitude" @pick="onPickCoords" />
                </div>

                <template #footer>
                    <Button label="Отмена" outlined severity="secondary" @click="dialogVisible = false" />
                    <Button :label="editingId ? 'Сохранить' : 'Создать'" :loading="busy" @click="save" />
                </template>
            </Dialog>

            <!-- Массовая перепривязка детей -->
            <Dialog v-model:visible="reparentVisible" modal header="Перепривязка дочерних локаций" style="width: 520px">
                <p style="margin:0 0 .8rem">
                    Дочерние локации «{{ reparentSource?.name }}» будут перенесены к новому родителю.
                </p>
                <label>
                    <span class="afield__label">Новый родитель *</span>
                    <Select v-model="reparentTarget" :options="reparentTargets" option-label="label"
                        option-value="value" filter class="w-full" />
                </label>
                <template #footer>
                    <Button label="Отмена" outlined severity="secondary" @click="reparentVisible = false" />
                    <Button :label="reparentAndDelete ? 'Перепривязать и удалить' : 'Перепривязать'"
                        :loading="reparentBusy" :disabled="!reparentTarget" @click="doReparent" />
                </template>
            </Dialog>
        </template>
    </Card>

    <DataTable :value="locations" :loading="loading" striped-rows class="atable mt-4">
        <Column field="name" header="Название">
            <template #body="{ data }">
                <div class="location__text-img">
                    {{ data.name }}
                    <TreeViewIcon v-if="hasChildren(data.id)" class="ml-1 text-muted location__icon" />
                </div>
            </template>
        </Column>

        <Column header="Историческое название">
            <template #body="{ data }">
                {{ data.historicalName || '—' }}
            </template>
        </Column>

        <Column header="Тип">
            <template #body="{ data }">
                <Tag :value="TYPE_META[data.type as LocationTypeJson]?.label ?? data.type"
                    :severity="TYPE_META[data.type as LocationTypeJson]?.severity ?? 'secondary'" />
            </template>
        </Column>

        <Column header="Родитель">
            <template #body="{ data }">
                {{ parentName(data.parentId) }}
            </template>
        </Column>

        <Column header="Координаты">
            <template #body="{ data }">
                <code>{{ fmtCoords(data) }}</code>
            </template>
        </Column>

        <Column header="" style="width: 180px">
            <template #body="{ data }">
                <Button text severity="secondary" title="Редактировать" @click="openEdit(data)">
                    <EditIcon />
                </Button>

                <Button v-if="hasChildren(data.id)" text severity="secondary" title="Перепривязать дочерние"
                    @click="openReparent(data)">
                    <TreeViewIcon />
                </Button>

                <Button text severity="danger" title="Удалить" @click="remove(data)">
                    <TrashCanIcon />
                </Button>
            </template>
        </Column>

        <template #empty>
            <div class="p-4 text-center">Локаций не найдено</div>
        </template>
    </DataTable>
</template>

<style scoped>
    .location__text-img {
        display: flex;
        align-items: center;
    }

    .location__icon {
        font-size: .8rem
    }
</style>