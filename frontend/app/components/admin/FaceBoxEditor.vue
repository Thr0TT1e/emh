<script setup lang="ts">
    import { Cropper } from "vue-advanced-cropper";
    import "vue-advanced-cropper/dist/style.css";
    import "vue-advanced-cropper/dist/theme.classic.css";
    import type { FaceBoxJson } from "~/sdk/emh/v1/hero_pb";
    import AccountCancelIcon from '~icons/mdi/account-cancel?width=1.5em&height=1.5em';
    import CheckIcon from '~icons/mdi/check?width=1.25em&height=1.25em';

    const props = withDefaults(
        defineProps<{
            visible: boolean;
            src: string;
            initialBox?: FaceBoxJson | null;
            title?: string;
            allowCancel?: boolean;
            skipLabel?: string;
        }>(),
        {
            initialBox: null,
            title: "Выделите область с лицом героя",
            allowCancel: false,
            skipLabel: "Пропустить (обрезать по центру)",
        },
    );

    const emit = defineEmits<{
        "update:visible": [value: boolean];
        save: [box: FaceBoxJson | null];
        cancel: [];
    }>();

    const cropper = ref<InstanceType<typeof Cropper>>();
    const cleared = ref(false);
    const previewUrl = ref("");
    const coords = ref({ x: 0, y: 0, width: 0, height: 0 });
    let rafId = 0;

    // Кропер готов: ставим пресет, если лицо уже было выделено ранее.
    const onReady = () => {
        const box = props.initialBox;
        if (box?.width && box?.height) {
            cropper.value?.setCoordinates({
                left: box.x ?? 0,
                top: box.y ?? 0,
                width: box.width,
                height: box.height,
            });
        }
        refreshMeta();
    };

    // Читаем координаты (natural) и строим превью thumbnail.
    const refreshMeta = () => {
        if (!cropper.value) return;
        const { coordinates, canvas } = cropper.value.getResult();
        if (coordinates) {
            coords.value = {
                x: Math.round(coordinates.left),
                y: Math.round(coordinates.top),
                width: Math.round(coordinates.width),
                height: Math.round(coordinates.height),
            };
        }
        if (canvas) previewUrl.value = canvas.toDataURL("image/webp", 0.8);
    };

    const onChange = () => {
        cleared.value = false;
        cancelAnimationFrame(rafId);
        rafId = requestAnimationFrame(refreshMeta);
    };

    // Логический сброс: отдадим null на бэкенд → центральный crop.
    const resetSelection = () => {
        cleared.value = true;
        previewUrl.value = "";
        coords.value = { x: 0, y: 0, width: 0, height: 0 };
    };

    const save = () => {
        emit("save", cleared.value ? null : { ...coords.value });
        emit("update:visible", false);
    };

    const skip = () => {
        emit("save", null);
        emit("update:visible", false);
    };

    const cancel = () => {
        emit("cancel");
        emit("update:visible", false);
    };

    onBeforeUnmount(() => cancelAnimationFrame(rafId));
</script>

<template>
    <Dialog :visible="visible" modal :header="title" :closable="allowCancel" :dismissable="allowCancel"
        :close-on-escape="allowCancel" :style="{ width: '860px' }">
        <div class="fbed">
            <p class="fbed__hint">
                Рамка фиксирует пропорции 4:5 — именно так бэкенд обрежет превью 480×600.
                Если герой на фото один и крупным планом, этот шаг можно пропустить.
            </p>

            <div class="fbed__grid">
                <div class="fbed__cropper">
                    <ClientOnly>
                        <Cropper :key="src" ref="cropper" :src="src" :stencil-props="{ aspectRatio: 4 / 5 }"
                            @ready="onReady" @change="onChange" />
                    </ClientOnly>
                </div>

                <div class="fbed__side">
                    <span class="afield__label">Превью 4:5</span>
                    <div class="fbed__preview">
                        <img v-if="previewUrl" :src="previewUrl" alt="Превью области" />
                        <div v-else class="fbed__preview-empty">центр</div>
                    </div>
                    <div class="fbed__coords num">
                        <span>x {{ coords.x }}</span>
                        <span>y {{ coords.y }}</span>
                        <span>w {{ coords.width }}</span>
                        <span>h {{ coords.height }}</span>
                    </div>
                    <Button outlined severity="secondary" size="small" @click="resetSelection">
                        <AccountCancelIcon />
                        <span>Сбросить выделение</span>
                    </Button>
                </div>
            </div>
        </div>

        <template #footer>
            <Button v-if="allowCancel" label="Отмена" outlined severity="secondary" @click="cancel" />
            <Button :label="skipLabel" outlined severity="secondary" @click="skip" />
            <Button @click="save">
                <CheckIcon />
                <span>Сохранить</span>
            </Button>
        </template>
    </Dialog>
</template>

<style scoped>
    .fbed__hint {
        margin: 0 0 1rem;
        font-size: 0.88rem;
        color: var(--emh-muted);
        line-height: 1.5;
    }

    .fbed__grid {
        display: grid;
        grid-template-columns: 1fr 200px;
        gap: 1.2rem;
        align-items: start;
    }

    .fbed__cropper {
        height: 440px;
        background: #1a1c20;
        overflow: hidden;
    }

    .fbed__side {
        display: flex;
        flex-direction: column;
        gap: 0.7rem;
    }

    .fbed__preview {
        width: 200px;
        height: 250px;
        /* 4:5 */
        border: 1px solid var(--emh-line);
        background: var(--emh-bg);
        overflow: hidden;
    }

    .fbed__preview img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        display: block;
    }

    .fbed__preview-empty {
        width: 100%;
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.7rem;
        letter-spacing: 0.14em;
        text-transform: uppercase;
        color: var(--emh-muted);
    }

    .fbed__coords {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 0.25rem 0.8rem;
        font-size: 0.8rem;
        color: var(--emh-muted);
    }

    @media (max-width: 720px) {
        .fbed__grid {
            grid-template-columns: 1fr;
        }
    }
</style>