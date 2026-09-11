<script setup lang="ts">
  import { useToast } from 'primevue/usetoast'
  import FaceBoxEditor from '~/components/admin/FaceBoxEditor.vue'
  import { useHeroPhotos } from '~/composables/useHeroPhotos'
  import type { FaceBoxJson, PhotoJson } from '~/sdk/emh/v1/hero_pb'
  import { UploadType } from '~/sdk/emh/v1/media_pb'

  const props = defineProps<{ heroId: string }>()
  const emit = defineEmits<{ changed: [] }>()
  const { media, heroAdmin } = useApi()
  const toast = useToast()
  const { photos, totalCount, loading, hasMore, loadMore, refresh } = useHeroPhotos({
    heroId: props.heroId,
    pageSize: 20,
    mode: 'admin',
  })

  const uploading = ref(false)
  const uploadProgress = ref({ done: 0, total: 0 })
  const makeMain = ref(false)
  const selectAllPhoto = ref(false)
  const selected = ref<string[]>([])
  const settingMain = ref<string | null>(null)

  const toggle = (id: string) => {
    selected.value = selected.value.includes(id)
      ? selected.value.filter((x) => x !== id)
      : [...selected.value, id]
  }

  // ---------------------------------------------------------------------------
  // Batch-загрузка
  // ---------------------------------------------------------------------------
  const onFiles = async (e: Event) => {
    const input = e.target as HTMLInputElement
    const files = Array.from(input.files ?? [])
    if (!files.length) return
    input.value = ''

    if (files.length > 50) {
      toast.add({
        severity: 'warn',
        summary: 'Слишком много файлов',
        detail: `Максимум 50 фото за один раз. Выбрано: ${files.length}`,
        life: 5000,
      })
      return
    }

    uploading.value = true
    uploadProgress.value = { done: 0, total: files.length }

    try {
      const batch = await media.batchGetUploadUrls({
        type: UploadType.HERO_PHOTO,
        files: files.map((f) => ({
          filename: f.name,
          contentType: f.type || 'image/jpeg',
        })),
      })

      if (!batch.urls || batch.urls.length !== files.length) {
        throw new Error('Бэкенд вернул некорректное количество URL')
      }

      const CONCURRENCY = 6
      const uploaded: { url: string; filename: string }[] = []
      const failed: string[] = []

      for (let i = 0; i < batch.urls.length; i += CONCURRENCY) {
        const chunk = batch.urls.slice(i, i + CONCURRENCY)
        const chunkFiles = files.slice(i, i + CONCURRENCY)

        const results = await Promise.allSettled(
          chunk.map(async (urlInfo, idx) => {
            const file = chunkFiles[idx]
            if (!file || !urlInfo) return null

            const put = await fetch(urlInfo.uploadUrl, {
              method: 'PUT',
              headers: { 'Content-Type': file.type || 'image/jpeg' },
              body: file,
            })
            if (!put.ok) throw new Error(`MinIO: ${put.status}`)

            return { url: urlInfo.publicUrl, filename: file.name }
          })
        )

        results.forEach((r, idx) => {
          const fname = chunkFiles[idx]?.name ?? '?'
          if (r.status === 'fulfilled' && r.value) {
            uploaded.push(r.value)
          } else {
            failed.push(fname)
            console.error(`Не удалось загрузить ${fname}:`, r.status === 'rejected' ? r.reason : 'null')
          }
        })

        uploadProgress.value.done = Math.min(i + CONCURRENCY, files.length)
      }

      if (uploaded.length === 0) {
        throw new Error('Ни один файл не удалось загрузить в хранилище')
      }

      await heroAdmin.batchAddHeroPhotos({
        heroId: props.heroId,
        photos: uploaded.map((u, i) => ({
          url: u.url,
          description: '',
          isMain: makeMain.value && i === 0,
          sortOrder: 0,
        })),
      })

      toast.add({
        severity: 'success',
        summary: `Загружено фото: ${uploaded.length}`,
        life: 2500,
      })
      if (failed.length > 0) {
        toast.add({
          severity: 'warn',
          summary: `Не загрузилось: ${failed.length}`,
          detail: failed.join(', '),
          life: 6000,
        })
      }

      // Обновляем галерею после загрузки
      await refresh()
      emit('changed')
    } catch (err: any) {
      toast.add({
        severity: 'error',
        summary: 'Ошибка загрузки',
        detail: err?.message ?? '',
        life: 5000,
      })
    } finally {
      uploading.value = false
      uploadProgress.value = { done: 0, total: 0 }
    }
  }

  // ---------------------------------------------------------------------------
  // Face box
  // ---------------------------------------------------------------------------
  const editorVisible = ref(false)
  const editorSrc = ref('')
  const editorBox = ref<FaceBoxJson | null>(null)
  const editingPhotoId = ref<string | null>(null)

  const openEditFaceBox = (photo: PhotoJson) => {
    editingPhotoId.value = photo.id ?? null
    editorSrc.value = photo.url ?? ''
    editorBox.value = photo.faceBox ?? null
    editorVisible.value = true
  }

  const onEditorSave = async (box: FaceBoxJson | null) => {
    if (!editingPhotoId.value) return

    try {
      await heroAdmin.updateHeroPhoto({
        photoId: editingPhotoId.value,
        heroId: props.heroId,
        faceBox: box as any,
        fieldMask: ['face_box'],
      })
      toast.add({
        severity: 'success',
        summary: 'Область сохранена',
        detail: 'Превью будет перегенерировано через пару секунд',
        life: 3000,
      })
      await refresh()
      emit('changed')
    } catch (err: any) {
      toast.add({
        severity: 'error',
        summary: 'Ошибка обновления',
        detail: err?.message ?? '',
        life: 5000,
      })
    }
    editingPhotoId.value = null
  }

  const onEditorCancel = () => {
    editingPhotoId.value = null
  }

  const removeSelected = async () => {
    if (!selected.value.length) return

    await heroAdmin.deleteHeroPhotos({
      heroId: props.heroId,
      photoIds: selected.value,
    })
    selected.value = []
    toast.add({ severity: 'info', summary: 'Фото удалены', life: 2500 })
    await refresh()
    emit('changed')
  }

  const move = async (index: number, dir: -1 | 1) => {
    const order = photos.value.map((p) => p.id)
    const j = index + dir
    if (j < 0 || j >= order.length) return

      ;[order[index], order[j]] = [order[j], order[index]]
    await heroAdmin.reorderHeroPhotos({
      heroId: props.heroId,
      photoIds: order as string[],
    })
    await refresh()
    emit('changed')
  }

  const setMain = async (photoId: string) => {
    if (settingMain.value) return
    settingMain.value = photoId

    try {
      await heroAdmin.setMainHeroPhoto({ heroId: props.heroId, photoId })
      toast.add({ severity: 'success', summary: 'Главное фото обновлено', life: 2500 })
      await refresh()
      emit('changed')
    } catch (err: any) {
      toast.add({
        severity: 'error',
        summary: 'Ошибка',
        detail: err?.message ?? 'Не удалось назначить главное фото',
        life: 5000,
      })
    } finally {
      settingMain.value = null
    }
  }

  function selectAllPhotos() {
    if (selectAllPhoto.value) {
      selected.value = photos.value.map((p) => p.id!) || []
    } else {
      selected.value = []
    }
  }
</script>

<template>
  <div class="apanel">
    <h2 class="apanel__title">Фотографии</h2>

    <div class="ph-toolbar">
      <label class="afield" style="margin: 0">
        <span class="afield__label">Загрузить (можно несколько)</span>
        <input type="file" accept="image/*" multiple class="afield__input" :disabled="uploading" @change="onFiles" />
      </label>
      <div class="ph-toolbar__make-main">
        <Checkbox v-model="makeMain" :binary="true" input-id="makeMain" />
        <label for="makeMain">первое из пакета — главным</label>
      </div>
      <div class="ph-toolbar__make-main">
        <Checkbox v-model="selectAllPhoto" @change="selectAllPhotos()" :binary="true" input-id="selectAllPhoto" />
        <label for="selectAllPhoto">выбрать все фото</label>
      </div>
      <Button v-if="selected.length" severity="danger" outlined size="small" :label="`Удалить (${selected.length})`"
        @click="removeSelected" />
      <span v-if="uploading" class="ph-busy"> Загружаем {{ uploadProgress.done }}/{{ uploadProgress.total }}… </span>
    </div>

    <div v-if="photos.length" class="ph-grid">
      <figure v-for="(p, i) in photos" :key="p.id" class="ph" :class="{ 'is-selected': selected.includes(p.id!) }">
        <div class="ph__top-actions">
          <button type="button" class="ph__icon-btn" :class="{ 'is-active': p.isMain }" :disabled="settingMain !== null"
            v-tooltip.bottom="p.isMain ? 'Это главное фото' : 'Сделать главным'" @click="setMain(p.id!)">
            <i :class="p.isMain ? 'pi pi-star-fill' : 'pi pi-star'" />
          </button>
          <Button icon="pi pi-arrow-up-right-and-arrow-down-left-from-center" size="small" severity="secondary"
            aria-label="Выделить лицо (face box)" v-tooltip.bottom="'Выделить лицо (face box)'"
            @click="openEditFaceBox(p)" />
        </div>
        <label class="ph__check">
          <Checkbox :model-value="selected.includes(p.id!)" :binary="true" @update:model-value="toggle(p.id!)" />
        </label>
        <Tag v-if="p.isMain" value="главное" severity="warn" class="ph-main-badge" />
        <Image :src="p.thumbnailUrl || p.url" :alt="p.description || 'Фотография героя'" preview loading="lazy" />
        <figcaption>
          <button :disabled="i === 0" title="Левее" @click="move(i, -1)">←</button>
          <button :disabled="i === photos.length - 1" title="Правее" @click="move(i, 1)">→</button>
        </figcaption>
      </figure>
    </div>
    <p v-else-if="!loading" class="ph-empty">Фотографий пока нет.</p>

    <!-- Кнопка "Загрузить ещё" -->
    <div v-if="hasMore && !loading" class="ph-load-more">
      <Button outlined label="Загрузить ещё" @click="loadMore" />
    </div>

    <!-- Индикатор загрузки -->
    <div v-if="loading && !photos.length" class="ph-loading">
      <ProgressSpinner style="width: 30px; height: 30px" />
      <span>Загружаем фотографии…</span>
    </div>

    <!-- Счётчик -->
    <p v-if="totalCount > photos.length" class="ph-counter">
      Показано {{ photos.length }} из {{ totalCount }}
    </p>

    <FaceBoxEditor v-model:visible="editorVisible" :src="editorSrc" :initial-box="editorBox"
      title="Выделите область с лицом героя" :allow-cancel="true" skip-label="Сбросить (обрезать по центру)"
      @save="onEditorSave" @cancel="onEditorCancel" />
  </div>
</template>

<style scoped>
  .ph-toolbar {
    display: flex;
    gap: 1.4rem;
    align-items: center;
    margin-bottom: 1.2rem;
    flex-wrap: wrap;
  }

  .ph-toolbar__make-main {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.88rem;
    color: var(--p-text-muted-color);
  }

  .ph-toolbar__make-main label {
    cursor: pointer;
    user-select: none;
  }

  .ph-busy {
    color: var(--emh-bronze);
    font-size: 0.85rem;
  }

  .ph-grid {
    display: flex;
    gap: 0.9rem;
    flex-wrap: wrap;
  }

  .ph {
    margin: 0;
    width: 150px;
    border: 1px solid var(--emh-line);
    background: #fff;
    padding: 4px;
    position: relative;
    transition: box-shadow 0.2s ease, transform 0.2s ease, border-color 0.2s ease;
  }

  .ph:hover {
    box-shadow: 0 10px 24px -16px rgba(26, 28, 32, 0.4);
    transform: translateY(-2px);
  }

  .ph.is-selected {
    border-color: var(--emh-crimson);
    box-shadow: 0 0 0 2px rgba(163, 22, 33, 0.2);
  }

  .ph img {
    width: 100%;
    height: 130px;
    object-fit: cover;
    display: block;
  }

  .ph__top-actions {
    position: absolute;
    top: 8px;
    left: 8px;
    z-index: 2;
    display: flex;
    gap: 4px;
  }

  .ph__icon-btn {
    background: rgba(255, 255, 255, 0.85);
    border: 1px solid var(--emh-line);
    border-radius: 4px;
    padding: 4px 6px;
    cursor: pointer;
    font-size: 0.95rem;
    line-height: 1;
    transition: background 0.2s, border-color 0.2s, color 0.2s;
  }

  .ph__icon-btn:hover:not(:disabled) {
    border-color: var(--emh-bronze);
    background: rgba(176, 141, 87, 0.15);
  }

  .ph__icon-btn:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .ph__icon-btn.is-active {
    color: var(--emh-bronze);
    border-color: var(--emh-bronze);
    background: rgba(176, 141, 87, 0.2);
  }

  .ph__check {
    position: absolute;
    top: 2px;
    right: 8px;
    z-index: 2;
    background: rgba(255, 255, 255, 0);
    padding: 2px 4px;
  }

  .ph figcaption {
    display: flex;
    justify-content: center;
    gap: 0.4rem;
    padding: 0.35rem 0 0.15rem;
  }

  .ph figcaption button {
    border: 1px solid var(--emh-line);
    background: #fff;
    cursor: pointer;
    padding: 0.1rem 0.5rem;
    transition: border-color 0.2s, color 0.2s;
  }

  .ph figcaption button:hover:not(:disabled) {
    border-color: var(--emh-crimson);
    color: var(--emh-crimson);
  }

  .ph figcaption button:disabled {
    opacity: 0.35;
    cursor: default;
  }

  .ph-empty {
    color: var(--emh-muted);
  }

  .ph-main-badge {
    position: absolute;
    bottom: 50px;
    right: 10px;
    z-index: 2;
  }

  .ph-load-more {
    display: flex;
    justify-content: center;
    margin-top: 1.5rem;
  }

  .ph-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    margin-top: 1.5rem;
    color: var(--emh-muted);
    font-size: 0.9rem;
  }

  .ph-counter {
    text-align: center;
    margin-top: 1rem;
    font-size: 0.85rem;
    color: var(--emh-muted);
  }
</style>