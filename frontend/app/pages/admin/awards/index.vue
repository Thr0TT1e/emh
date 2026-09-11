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
    import { toPlain } from '~/lib/pb'
    import type { AwardJson } from '~/sdk/emh/v1/award_pb'
    import { AwardSchema } from '~/sdk/emh/v1/award_pb'
    import { UploadType } from '~/sdk/emh/v1/media_pb'

    definePageMeta({ layout: 'admin', middleware: 'admin' })
    useHead({ title: 'Награды — Вечная память героям' })

    const { award, awardAdmin } = useApi()
    const toast = useToast()
    const confirm = useConfirm()

    // ---------------------------------------------------------------------------
    // Список + серверный поиск (search_query есть в ListAwardsRequest)
    // ---------------------------------------------------------------------------
    const awards = ref<AwardJson[]>([])
    const loading = ref(false)
    const searchQuery = ref('')

    async function load() {
        loading.value = true
        try {
            const res = await award.listAwards({ searchQuery: searchQuery.value })
            // Сортируем по sort_order по возрастанию: меньший номер = старшая награда.
            awards.value = (res.awards ?? [])
                .map((a) => toPlain(AwardSchema, a))
                .sort((x, y) => (x.sortOrder ?? 0) - (y.sortOrder ?? 0))
        } catch (e: any) {
            toast.add({ severity: 'error', summary: 'Ошибка', detail: e?.message ?? 'Не удалось загрузить награды', life: 5000 })
        } finally {
            loading.value = false
        }
    }

    let searchTimer: ReturnType<typeof setTimeout> | undefined
    watch(searchQuery, () => {
        clearTimeout(searchTimer)
        searchTimer = setTimeout(load, 400)
    })

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
    })

    function openCreate() {
        editingId.value = null
        Object.assign(form, { name: '', description: '', imageUrl: '', sortOrder: 0 })
        dialogVisible.value = true
    }

    function openEdit(a: AwardJson) {
        editingId.value = a.id ?? null
        Object.assign(form, {
            name: a.name ?? '',
            description: a.description ?? '',
            imageUrl: a.imageUrl ?? '',
            sortOrder: a.sortOrder ?? 0,
        })
        dialogVisible.value = true
    }

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
        busy.value = true
        try {
            const payload = {
                name: form.name,
                description: form.description,
                imageUrl: form.imageUrl,
                sortOrder: form.sortOrder,
            }
            if (editingId.value) {
                await awardAdmin.updateAward({
                    id: editingId.value,
                    ...payload,
                    fieldMask: ['name', 'description', 'image_url', 'sort_order'],
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
            await load()
        } catch (e: any) {
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
                await load()
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
                <Button label="Добавить награду" icon="pi pi-plus" @click="openCreate" />
            </div>

            <!-- Создание / редактирование -->
            <Dialog v-model:visible="dialogVisible" modal
                :header="editingId ? 'Редактирование награды' : 'Новая награда'" style="width: 640px">
                <div class="flex flex-col gap-3">
                    <div class="award_pic">
                        <div>
                            <span class="afield__label">Изображение (лента / знак)</span>
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

                    <label>
                        <span class="afield__label">Описание / статут</span>
                        <Textarea v-model="form.description" rows="3" class="w-full" />
                    </label>
                </div>

                <template #footer>
                    <Button label="Отмена" outlined severity="secondary" @click="dialogVisible = false" />
                    <Button :label="editingId ? 'Сохранить' : 'Создать'" :loading="busy" @click="save" />
                </template>
            </Dialog>
        </template>
    </Card>

    <DataTable :value="awards" :loading="loading" striped-rows class="atable mt-4" table-style="min-width: 50rem">
        <Column header="Изображение" style="width: 90px">
            <template #body="{ data }">
                <img v-if="data.imageUrl" :src="data.imageUrl" :alt="data.name" class="award-thumb" />
                <span v-else class="award-thumb award-thumb--empty"><i class="pi pi-image" /></span>
            </template>
        </Column>

        <Column field="name" header="Название" />

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
                <Button icon="pi pi-pencil" text severity="secondary" title="Редактировать" @click="openEdit(data)" />
                <Button icon="pi pi-trash" text severity="danger" title="Удалить" @click="remove(data)" />
            </template>
        </Column>

        <template #empty>
            <div class="p-4 text-center">Наград не найдено</div>
        </template>
    </DataTable>
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

    .award-desc {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
</style>