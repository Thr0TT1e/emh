<script setup lang="ts">
    import { useToast } from 'primevue/usetoast';
    import { useApi } from '~/composables/useApi';
    import { UploadType } from '~/sdk/emh/v1/media_pb';
    import FileOutlineIcon from '~icons/mdi/file-outline?width=1.25em&height=1.25em';

    // Props
    const attachments = defineModel<{ name: string; url: string }[]>('attachments', { default: () => [] })

    // API
    const { media } = useApi()
    const toast = useToast()

    // State
    const uploading = ref(false)
    const fileInput = ref<HTMLInputElement | null>(null)

    // Handle file selection
    const onFiles = async (e: Event) => {
        const input = e.target as HTMLInputElement
        const files = Array.from(input.files ?? [])
        if (!files.length) return

        input.value = '' // Reset input
        uploading.value = true

        try {
            for (const file of files) {
                const { uploadUrl, publicUrl } = await media.getUploadUrl({
                    type: UploadType.SUBMISSION_ATTACHMENT,
                    filename: file.name,
                    contentType: file.type || 'application/octet-stream',
                })

                const put = await fetch(uploadUrl, {
                    method: 'PUT',
                    headers: { 'Content-Type': file.type || 'application/octet-stream' },
                    body: file,
                })

                if (!put.ok) throw new Error('MinIO: ' + put.status)

                attachments.value.push({ name: file.name, url: publicUrl })
            }
        } catch (err: any) {
            toast.add({
                severity: 'error',
                summary: 'Ошибка загрузки',
                detail: err?.message ?? '',
                life: 5000,
            })
        } finally {
            uploading.value = false
        }
    }

    // Remove attachment
    const removeAttachment = (i: number) => {
        attachments.value.splice(i, 1)
    }
</script>

<template>
    <div class="attachment-upload">
        <div class="afield">
            <label class="afield__label">
                Прикрепить документы
                <span class="text-muted text-sm ml-2">(опционально)</span>
            </label>

            <div class="mt-2">
                <input ref="fileInput" type="file" multiple accept=".pdf,.doc,.docx,.jpg,.jpeg,.png,.txt" class="hidden"
                    @change="onFiles" />

                <Button label="Выбрать файлы" icon="i-mdi-upload" :loading="uploading" @click="fileInput?.click()" />

                <div v-if="uploading" class="mt-2 text-sm text-muted">
                    <ProgressSpinner style="width: 20px; height: 20px" class="inline-block mr-2" />
                    Загрузка...
                </div>
            </div>
        </div>

        <!-- Attachments list -->
        <div v-if="attachments.length > 0" class="mt-4 space-y-2">
            <div class="text-sm text-muted mb-2">Прикреплённые файлы ({{ attachments.length }}):</div>

            <div v-for="(att, i) in attachments" :key="i"
                class="attachment-item flex items-center justify-between p-3 bg-surface-50 dark:bg-surface-800 rounded">
                <div class="flex items-center min-w-0">
                    <FileOutlineIcon />
                    <span class="truncate">{{ att.name }}</span>
                </div>

                <Button icon="i-mdi-close" severity="danger" text size="small" @click="removeAttachment(i)" />
            </div>
        </div>
    </div>
</template>

<style scoped>
    .attachment-item {
        transition: background-color 0.15s ease;
    }
</style>