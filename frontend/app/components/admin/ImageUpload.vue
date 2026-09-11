<script setup lang="ts">
    import { useToast } from 'primevue/usetoast';
    import type { UploadType } from '~/sdk/emh/v1/media_pb';

    const props = defineProps<{
        modelValue: string
        type: UploadType
    }>()
    const emit = defineEmits<{ 'update:modelValue': [url: string] }>()

    const { media } = useApi()
    const toast = useToast()
    const uploading = ref(false)

    async function onFile(e: Event) {
        const input = e.target as HTMLInputElement
        const file = input.files?.[0]
        if (!file) return

        // КРИТИЧНО: сбрасываем value СРАЗУ, синхронно, до любых реактивных изменений.
        // Файл уже получен (объект File валиден), а input type="file" не должен
        // удерживать непустое значение к моменту ре-рендеров Vue — иначе InvalidStateError.
        input.value = ''

        uploading.value = true
        try {
            const { uploadUrl, publicUrl } = await media.getUploadUrl({
                type: props.type,
                filename: file.name,
                contentType: file.type || 'image/jpeg',
            })
            const put = await fetch(uploadUrl, {
                method: 'PUT',
                headers: { 'Content-Type': file.type || 'image/jpeg' },
                body: file,
            })
            if (!put.ok) throw new Error('MinIO: ' + put.status)
            emit('update:modelValue', publicUrl)
        } catch (err: any) {
            toast.add({ severity: 'error', summary: 'Ошибка загрузки', detail: err?.message ?? '', life: 5000 })
        } finally {
            uploading.value = false
        }
    }
</script>

<template>
    <div class="imgup">
        <div v-if="modelValue" class="imgup__preview">
            <Image :src="modelValue" alt="Изображение" />
        </div>
        <div v-else class="imgup__empty">нет изображения</div>

        <div class="imgup__actions">
            <label class="imgup__btn">
                {{ modelValue ? 'Заменить' : 'Загрузить' }}
                <input type="file" accept="image/*" :disabled="uploading" @change="onFile" />
            </label>
            <span v-if="uploading" class="imgup__busy">загружаем…</span>
            <button v-if="modelValue && !uploading" type="button" class="imgup__btn imgup__btn--danger"
                @click="emit('update:modelValue', '')">
                Удалить
            </button>
        </div>
    </div>
</template>

<style scoped>
    .imgup {
        width: 140px;
    }

    .imgup__preview {
        width: 140px;
        height: 140px;
        border: 1px solid var(--p-surface-200);
        background: #fff;
    }

    .imgup__preview img {
        width: 100%;
        height: 100%;
        object-fit: contain;
    }

    .imgup__empty {
        width: 140px;
        height: 140px;
        display: flex;
        align-items: center;
        justify-content: center;
        border: 1px dashed var(--p-surface-300);
        color: var(--p-text-muted-color);
        font-size: .8rem;
    }

    .imgup__actions {
        display: flex;
        align-items: center;
        gap: .7rem;
        margin-top: .5rem;
    }

    .imgup__btn {
        position: relative;
        overflow: hidden;
        cursor: pointer;
        font-size: .82rem;
        color: var(--p-primary-color);
        border-bottom: 1px dashed currentColor;
    }

    .imgup__btn input {
        position: absolute;
        inset: 0;
        opacity: 0;
        cursor: pointer;
    }

    .imgup__btn--danger {
        color: var(--emh-carmine, #bd2f3c);
        border: 0;
        background: none;
    }

    .imgup__busy {
        font-size: .78rem;
        color: var(--p-primary-color);
    }
</style>