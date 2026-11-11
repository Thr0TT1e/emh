<script setup lang="ts">
    /**
     * /admin/awards — CRUD справочника наград.
     * Разделение по контракту: чтение — публичный AwardService (award),
     * мутации — AwardAdminService (awardAdmin).
     * Изображения — presigned URL через MediaService (UploadType.AWARD_IMAGE).
     */
    import { useConfirm } from 'primevue/useconfirm'
    import { useToast } from 'primevue/usetoast'
    import ImageUpload from '~/components/admin/ImageUpload.vue'
    import { awardJurisdictionLabel, awardTypeLabel } from '~/lib/awards'
    import { isAlreadyExists } from '~/lib/errors'
    import { toPlain } from '~/lib/pb'
    import type { AwardJson } from '~/sdk/emh/v1/award_pb'
    import { AwardSchema } from '~/sdk/emh/v1/award_pb'
    import type { AwardJurisdictionJson, AwardTypeJson } from '~/sdk/emh/v1/enums_emh_pb'
    import { AwardJurisdiction, AwardType } from '~/sdk/emh/v1/enums_emh_pb'
    import { UploadType } from '~/sdk/emh/v1/media_pb'
    import EditIcon from '~icons/carbon/edit?width=1.25em&height=1.25em'
    import ImageIcon from '~icons/carbon/image?width=1.25em&height=1.25em'
    import TrashCanIcon from '~icons/carbon/trash-can?width=1.25em&height=1.25em'
    import WarningAltIcon from '~icons/carbon/warning-alt?width=1.25em&height=1.25em'
    import MedalOutlineIcon from '~icons/mdi/medal-outline?width=1.25em&height=1.25em'

    definePageMeta({ layout: 'admin', middleware: 'admin' })
    useHead({ title: 'Награды — Вечная память героям' })

    const { award, awardAdmin } = useApi()
    const toast = useToast()
    const confirm = useConfirm()

    // ---------------------------------------------------------------------------
    // Справочные мапы (тип и принадлежность — enum'ы, в форме держим numeric)
    // ---------------------------------------------------------------------------
    const TYPE_OPTIONS = [
        { label: 'Орден', value: AwardType.ORDER },
        { label: 'Медаль', value: AwardType.MEDAL },
        { label: 'Знак отличия', value: AwardType.BADGE },
    ]

    const JURISDICTION_OPTIONS = [
        { label: 'Не указана', value: AwardJurisdiction.UNSPECIFIED },
        { label: 'Российская Федерация', value: AwardJurisdiction.RUSSIAN_FEDERATION },
        { label: 'СССР', value: AwardJurisdiction.USSR },
        { label: 'Ведомственная', value: AwardJurisdiction.DEPARTMENTAL },
    ]

    const TYPE_JSON_TO_NUM: Record<AwardTypeJson, AwardType> = {
        AWARD_TYPE_UNSPECIFIED: AwardType.UNSPECIFIED,
        AWARD_TYPE_ORDER: AwardType.ORDER,
        AWARD_TYPE_MEDAL: AwardType.MEDAL,
        AWARD_TYPE_BADGE: AwardType.BADGE,
    }

    const JURISDICTION_JSON_TO_NUM: Record<AwardJurisdictionJson, AwardJurisdiction> = {
        AWARD_JURISDICTION_UNSPECIFIED: AwardJurisdiction.UNSPECIFIED,
        AWARD_JURISDICTION_RUSSIAN_FEDERATION: AwardJurisdiction.RUSSIAN_FEDERATION,
        AWARD_JURISDICTION_USSR: AwardJurisdiction.USSR,
        AWARD_JURISDICTION_DEPARTMENTAL: AwardJurisdiction.DEPARTMENTAL,
    }

    // ---------------------------------------------------------------------------
    // Список + серверный поиск (search_query есть в ListAwardsRequest)
    // Курсорная пагинация pageSize=50: keyset (sort_order,name,id) — обход
    // полный; без явного pagination usecase обрезал до 20 (сид ~400 = невидимы).
    // ---------------------------------------------------------------------------
    const PAGE_SIZE = 50

    const awards = ref<AwardJson[]>([])
    const loading = ref(false)
    const searchQuery = ref('')
    const page = ref(0)
    const cursorStack = ref<string[]>([''])
    const hasMore = ref(false)

    async function load(p: number = page.value) {
        loading.value = true
        try {
            const res = await award.listAwards({
                searchQuery: searchQuery.value,
                pagination: { pageSize: PAGE_SIZE, cursor: cursorStack.value[p] ?? '' },
            })
            // Сортируем по sort_order по возрастанию: меньший номер = старшая награда.
            awards.value = (res.awards ?? [])
                .map((a) => toPlain(AwardSchema, a))
                .sort((x, y) => (x.sortOrder ?? 0) - (y.sortOrder ?? 0))
            const next = res.pagination?.nextCursor ?? ''
            hasMore.value = next !== ''
            if (next) cursorStack.value[p + 1] = next
            page.value = p
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось загрузить награды', life: 5000 })
        } finally {
            loading.value = false
        }
    }

    let searchTimer: ReturnType<typeof setTimeout> | undefined
    watch(searchQuery, () => {
        clearTimeout(searchTimer)
        searchTimer = setTimeout(() => {
            cursorStack.value = ['']
            load(0)
        }, 400)
    })

    // ---------------------------------------------------------------------------
    // Каталог для клиентской проверки дублей (уровень 1, ADR-005).
    // Полный справочник курсорным циклом (useAwardCatalog); сравнение — строгое
    // равенство тройки (name, type, jurisdiction): зеркалит уникальный индекс
    // idx_awards_seed_key, не блокирует то, что сервер принял бы.
    // duplicateExisting определён после формы (зависит от form/editingId).
    // ---------------------------------------------------------------------------
    const { data: catalog, refresh: refreshCatalog } = await useAwardCatalog()

    onMounted(load)

    // ---------------------------------------------------------------------------
    // Модалка создания/редактирования (МУТАЦИИ — админский сервис)
    // ---------------------------------------------------------------------------
    const dialogVisible = ref(false)
    const busy = ref(false)
    const editingId = ref<string | null>(null)

    const form = reactive({
        name: '',
        description: '',
        imageUrl: '',
        sortOrder: 0,
        ribbonImageUrl: '',
        type: AwardType.ORDER as AwardType,
        jurisdiction: AwardJurisdiction.UNSPECIFIED as AwardJurisdiction,
        wornWithoutBar: false,
        isJubilee: false,
    })

    // Подписи numeric-значений формы для текста дубль-предупреждения
    // (awardTypeLabel/awardJurisdictionLabel принимают JSON-строки proto).
    const formTypeLabel = computed(() =>
        TYPE_OPTIONS.find((o) => o.value === form.type)?.label ?? 'Не указан')
    const formJurisdictionLabel = computed(() =>
        JURISDICTION_OPTIONS.find((o) => o.value === form.jurisdiction)?.label ?? 'Не указана')

    // Полный набор изменяемых полей для partial update (snake_case, Q24а).
    const UPDATE_FIELD_MASK = [
        'name', 'description', 'image_url', 'sort_order',
        'ribbon_image_url', 'type', 'jurisdiction', 'worn_without_bar', 'is_jubilee',
    ]

    function openCreate() {
        editingId.value = null
        Object.assign(form, {
            name: '', description: '', imageUrl: '', sortOrder: 0,
            ribbonImageUrl: '', type: AwardType.ORDER,
            jurisdiction: AwardJurisdiction.UNSPECIFIED,
            wornWithoutBar: false, isJubilee: false,
        })
        dialogVisible.value = true
    }

    function openEdit(a: AwardJson) {
        editingId.value = a.id ?? null
        Object.assign(form, {
            name: a.name ?? '',
            description: a.description ?? '',
            imageUrl: a.imageUrl ?? '',
            sortOrder: a.sortOrder ?? 0,
            ribbonImageUrl: a.ribbonImageUrl ?? '',
            type: TYPE_JSON_TO_NUM[a.type ?? 'AWARD_TYPE_UNSPECIFIED'] ?? AwardType.UNSPECIFIED,
            jurisdiction: JURISDICTION_JSON_TO_NUM[a.jurisdiction ?? 'AWARD_JURISDICTION_UNSPECIFIED'] ?? AwardJurisdiction.UNSPECIFIED,
            wornWithoutBar: a.wornWithoutBar ?? false,
            isJubilee: a.isJubilee ?? false,
        })
        dialogVisible.value = true
    }

    // Уровень 1: совпадение ключа (name, type, jurisdiction) с другой записью
    // каталога. В режиме правки собственная запись исключена (rename/edit без
    // смены ключа — не дубль). Строгое равенство строк — зеркало индекса.
    const duplicateExisting = computed<AwardJson | null>(() => {
        if (!form.name) return null
        return (catalog.value ?? []).find((a) =>
            a.id !== (editingId.value ?? '') &&
            a.name === form.name &&
            TYPE_JSON_TO_NUM[a.type ?? 'AWARD_TYPE_UNSPECIFIED'] === form.type &&
            JURISDICTION_JSON_TO_NUM[a.jurisdiction ?? 'AWARD_JURISDICTION_UNSPECIFIED'] === form.jurisdiction,
        ) ?? null
    })

    async function save() {
        if (!form.name.trim()) {
            toast.add({
                severity: 'warn',
                summary: 'Проверьте форму',
                detail: 'Название обязательно',
                life: 3000
            })
            return
        }
        if (form.type === AwardType.UNSPECIFIED) {
            toast.add({
                severity: 'warn',
                summary: 'Проверьте форму',
                detail: 'Тип награды обязателен',
                life: 3000
            })
            return
        }
        // Уровень 1 (страховка): кнопка уже disabled при дубле, но save()
        // может быть вызван иначе — не отправляем заведомый конфликт ключа.
        if (duplicateExisting.value) return
        busy.value = true
        try {
            const payload = {
                name: form.name,
                description: form.description,
                imageUrl: form.imageUrl,
                sortOrder: form.sortOrder,
                ribbonImageUrl: form.ribbonImageUrl,
                type: form.type,
                jurisdiction: form.jurisdiction,
                wornWithoutBar: form.wornWithoutBar,
                isJubilee: form.isJubilee,
            }
            if (editingId.value) {
                await awardAdmin.updateAward({
                    id: editingId.value,
                    ...payload,
                    fieldMask: UPDATE_FIELD_MASK,
                })
                toast.add({
                    severity: 'success',
                    summary: 'Сохранено',
                    detail: form.name,
                    life: 3000
                })
            } else {
                await awardAdmin.createAward(payload)
                toast.add({
                    severity: 'success',
                    summary: 'Создано',
                    detail: form.name,
                    life: 3000
                })
            }
            dialogVisible.value = false
            await Promise.all([load(), refreshCatalog()])
        } catch (e: any) {
            // Уровень 2: серверный конфликт idx_awards_seed_key (гонка двух
            // админов, пропуски уровня 1). Сообщение человекочитаемое — с бэкенда.
            if (isAlreadyExists(e)) {
                toast.add({
                    severity: 'error',
                    summary: 'Уже существует',
                    detail: e?.message ?? 'Награда с таким названием, типом и юрисдикцией уже существует',
                    life: 5000,
                })
                // Каталог устарел (дубль создан конкурентом) — обновляем,
                // индикатор уровня 1 подхватит совпадение.
                await refreshCatalog()
                return
            }
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось сохранить', life: 5000 })
        } finally {
            busy.value = false
        }
    }

    // ---------------------------------------------------------------------------
    // Удаление
    // ---------------------------------------------------------------------------
    function remove(a: AwardJson) {
        confirm.require({
            message: `Удалить награду «${a.name}»? Если она привязана к героям, бэкенд вернёт ошибку.`,
            header: 'Удаление награды',
            icon: 'pi pi-exclamation-triangle',
            acceptLabel: 'Удалить',
            rejectLabel: 'Отмена',
            acceptClass: 'p-button-danger',
            accept: () => doDelete(a),
        })
    }

    async function doDelete(a: AwardJson) {
        try {
            const res = await awardAdmin.deleteAward({ id: a.id ?? '' })
            if (res.success) {
                toast.add({
                    severity: 'success',
                    summary: 'Удалено',
                    detail: a.name,
                    life: 3000
                })
                // Ключ удалённой записи освобождается — каталог дубль-чека обновляем.
                await Promise.all([load(), refreshCatalog()])
            }
        } catch (e: any) {
            // «При наличии связей с героями вернёт ошибку» — показываем текст бэкенда
            toast.add({
                severity: 'error',
                summary: 'Ошибка',
                detail: e?.message ?? 'Не удалось удалить',
                life: 5000
            })
        }
    }
</script>

<template>
    <Card>
        <template #title>Справочник наград</template>

        <template #content>
            <div class="award_general">
                <InputText v-model="searchQuery" placeholder="Поиск по названию…" />
                <Button @click="openCreate">
                    <MedalOutlineIcon />
                    <span>Добавить награду</span>
                </Button>
            </div>

            <!-- Создание / редактирование -->
            <Dialog v-model:visible="dialogVisible" modal
                :header="editingId ? 'Редактирование награды' : 'Новая награда'" style="width: 640px">
                <div class="flex flex-col gap-3">
                    <div class="award_pic">
                        <div>
                            <span class="afield__label">Знак награды</span>
                            <ImageUpload v-model="form.imageUrl" :type="UploadType.AWARD_IMAGE" />
                        </div>

                        <div class="flex flex-col gap-3">
                            <label>
                                <span class="afield__label">Название *</span>
                                <InputText v-model="form.name" class="w-full"
                                    placeholder="Герой Российской Федерации" />
                            </label>

                            <label>
                                <span class="afield__label">Приоритет (меньше = старше)</span>
                                <InputNumber v-model="form.sortOrder" :min="0" />
                            </label>
                        </div>
                    </div>

                    <div class="award_pic">
                        <div>
                            <span class="afield__label">Лента (для планки)</span>
                            <ImageUpload v-model="form.ribbonImageUrl" :type="UploadType.AWARD_RIBBON" />
                        </div>

                        <div class="flex flex-col gap-3">
                            <label>
                                <span class="afield__label">Тип награды *</span>
                                <Select v-model="form.type" :options="TYPE_OPTIONS" option-label="label"
                                    option-value="value" placeholder="Выберите тип…" class="w-full" />
                            </label>

                            <label>
                                <span class="afield__label">Принадлежность</span>
                                <Select v-model="form.jurisdiction" :options="JURISDICTION_OPTIONS" option-label="label"
                                    option-value="value" placeholder="Выберите…" class="w-full" />
                            </label>
                        </div>
                    </div>

                    <div class="award_flags">
                        <div class="award_flag">
                            <Checkbox v-model="form.wornWithoutBar" :binary="true" input-id="awardWornWithoutBar" />
                            <label for="awardWornWithoutBar">Носится без колодки (не входит в планку)</label>
                        </div>
                        <div class="award_flag">
                            <Checkbox v-model="form.isJubilee" :binary="true" input-id="awardIsJubilee" />
                            <label for="awardIsJubilee">Юбилейная награда</label>
                        </div>
                    </div>

                    <!-- Уровень 1: совпадение ключа (name, type, jurisdiction) -->
                    <Message v-if="duplicateExisting" severity="warn" :closable="false">
                        <div class="award_dup">
                            <WarningAltIcon />
                            <span>В справочнике уже есть «{{ form.name }}» (тип: {{ formTypeLabel }}, принадлежность: {{ formJurisdictionLabel }}).</span>
                        </div>
                    </Message>

                    <label>
                        <span class="afield__label">Описание / статут</span>
                        <Textarea v-model="form.description" rows="3" class="w-full" />
                    </label>
                </div>

                <template #footer>
                    <Button label="Отмена" outlined severity="secondary" @click="dialogVisible = false" />
                    <Button :label="editingId ? 'Сохранить' : 'Создать'" :loading="busy" :disabled="!!duplicateExisting"
                        :title="duplicateExisting ? 'Такая награда уже есть в справочнике' : undefined" @click="save" />
                </template>
            </Dialog>
        </template>
    </Card>

    <DataTable :value="awards" :loading="loading" striped-rows data-key="id" class="atable mt-4" table-style="min-width: 50rem">
        <Column header="Изображение" style="width: 90px">
            <template #body="{ data }">
                <img v-if="data.imageUrl" :src="data.imageUrl" :alt="data.name" class="award-thumb" />
                <span v-else class="award-thumb award-thumb--empty">
                    <ImageIcon />
                </span>
            </template>
        </Column>

        <Column field="name" header="Название" />

        <Column header="Лента" style="width: 70px">
            <template #body="{ data }">
                <img v-if="data.ribbonImageUrl" :src="data.ribbonImageUrl" :alt="`Лента: ${data.name}`"
                    class="award-ribbon" width="36" height="12" />
                <span v-else class="award-ribbon award-ribbon--empty" aria-hidden="true"></span>
            </template>
        </Column>

        <Column header="Тип" style="width: 180px">
            <template #body="{ data }">
                <div class="award-tags">
                    <Tag :value="awardTypeLabel(data.type)" severity="secondary" />
                    <Tag v-if="data.jurisdiction && data.jurisdiction !== 'AWARD_JURISDICTION_UNSPECIFIED'"
                        :value="awardJurisdictionLabel(data.jurisdiction)" severity="info" />
                </div>
            </template>
        </Column>

        <Column header="Описание">
            <template #body="{ data }">
                <span class="award-desc">{{ data.description || '—' }}</span>
            </template>
        </Column>

        <Column header="Приоритет" style="width: 110px">
            <template #body="{ data }"><code>{{ data.sortOrder ?? 0 }}</code></template>
        </Column>

        <Column header="" style="width: 150px">
            <template #body="{ data }">
                <Button text severity="secondary" title="Редактировать" @click="openEdit(data)">
                    <EditIcon />
                </Button>

                <Button text severity="danger" title="Удалить" @click="remove(data)">
                    <TrashCanIcon />
                </Button>
            </template>
        </Column>

        <template #empty>
            <div class="p-4 text-center">Наград не найдено</div>
        </template>
    </DataTable>

    <!-- Курсорная пагинация (паттерн admin/heroes): stack курсоров, Назад/Вперёд -->
    <div class="mt-3 flex gap-3">
        <Button outlined :disabled="page === 0 || loading" label="← Назад" severity="secondary"
            @click="load(page - 1)" />
        <Button outlined :disabled="!hasMore || loading" label="Вперёд →" severity="secondary"
            @click="load(page + 1)" />
    </div>
</template>

<style scoped>
    .award_general {
        display: grid;
        grid-template-columns: auto max-content;
        column-gap: 2.5rem;
    }

    .award_pic {
        display: grid;
        grid-template-columns: 140px 1fr;
        gap: 1.2rem
    }

    .award_flags {
        display: flex;
        flex-direction: column;
        gap: 0.6rem;
    }

    .award_flag {
        display: flex;
        align-items: center;
        gap: 0.6rem;
    }

    .award_dup {
        display: flex;
        align-items: flex-start;
        gap: 0.5rem;
    }

    .award-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 0.3rem;
    }

    .award-thumb {
        width: 56px;
        height: 56px;
        object-fit: contain;
        border: 1px solid var(--p-surface-200);
        background: #fff;
    }

    .award-thumb--empty {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 56px;
        height: 56px;
        color: var(--p-text-muted-color);
        border: 1px dashed var(--p-surface-300);
        background: #fff;
    }

    .award-ribbon {
        display: block;
        width: 36px;
        height: 12px;
        object-fit: cover;
        border: 1px solid var(--p-surface-200);
    }

    .award-ribbon--empty {
        border: 1px dashed var(--p-surface-300);
        background: #fff;
    }

    .award-desc {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
</style>