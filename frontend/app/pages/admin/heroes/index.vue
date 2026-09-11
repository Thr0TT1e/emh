<script setup lang="ts">
  import { useConfirm } from "primevue/useconfirm";
  import { useToast } from "primevue/usetoast";
  import { formatFlexibleYear } from "~/lib/format";
  import { toPlain } from "~/lib/pb";
  import { PublicationStatus, type PublicationStatusJson } from "~/sdk/emh/v1/enums_emh_pb";
  import type { ListAdminHeroesResponse } from "~/sdk/emh/v1/hero_admin_pb";
  import { HeroSummarySchema, type HeroSummaryJson } from "~/sdk/emh/v1/hero_pb";

  definePageMeta({ layout: "admin", middleware: "admin" });

  const { heroAdmin } = useApi();
  const toast = useToast();
  const confirm = useConfirm();

  const search = ref("");
  const heroes = ref<HeroSummaryJson[]>([]);
  const cursorStack = ref<string[]>([""]);
  const page = ref(0);
  const hasMore = ref(false);
  const loading = ref(false);

  // ---------------------------------------------------------------------------
  // Фильтр по статусу публикации
  // ---------------------------------------------------------------------------
  type StatusFilter = "all" | "published" | "draft" | "archived";

  const statusFilter = ref<StatusFilter>("all");

  const STATUS_OPTIONS = [
    { label: "Показать всех", value: "all" },
    { label: "Опубликованных", value: "published" },
    { label: "В черновике", value: "draft" },
    { label: "Удалённых (архив)", value: "archived" },
  ];

  const FILTER_PARAMS: Record<StatusFilter, { status: PublicationStatus; includeArchived: boolean }> = {
    all: { status: PublicationStatus.UNSPECIFIED, includeArchived: true },
    published: { status: PublicationStatus.PUBLISHED, includeArchived: false },
    draft: { status: PublicationStatus.DRAFT, includeArchived: false },
    archived: { status: PublicationStatus.ARCHIVED, includeArchived: true },
  };

  // Мета для отображения статуса в таблице
  const STATUS_META: Record<PublicationStatusJson, { label: string; severity: "success" | "warn" | "secondary" }> = {
    PUBLICATION_STATUS_UNSPECIFIED: { label: "—", severity: "secondary" },
    PUBLICATION_STATUS_DRAFT: { label: "Черновик", severity: "warn" },
    PUBLICATION_STATUS_PUBLISHED: { label: "Опубликовано", severity: "success" },
    PUBLICATION_STATUS_ARCHIVED: { label: "Архив", severity: "secondary" },
  };

  // типобезопасное получение меты (дефолт для пустых значений)
  const statusMeta = (status?: PublicationStatusJson) =>
    STATUS_META[status ?? "PUBLICATION_STATUS_UNSPECIFIED"];

  // сравнение со строковым JSON-значением, а не с несуществующим
  // числовым ключом PublicationStatus.PUBLICATION_STATUS_ARCHIVED (undefined)
  const isArchived = (status?: PublicationStatusJson) =>
    status === "PUBLICATION_STATUS_ARCHIVED";

  // Обратный маппинг: статус записи → значение фильтра (быстрый фильтр кликом)
  const STATUS_TO_FILTER: Record<PublicationStatusJson, StatusFilter> = {
    PUBLICATION_STATUS_UNSPECIFIED: "all",
    PUBLICATION_STATUS_DRAFT: "draft",
    PUBLICATION_STATUS_PUBLISHED: "published",
    PUBLICATION_STATUS_ARCHIVED: "archived",
  };

  const applyStatusFilter = (status?: PublicationStatusJson) => {
    if (!status) return;
    statusFilter.value = STATUS_TO_FILTER[status];
  };

  // строгая типизация: res: ListAdminHeroesResponse,
  // h выводится как HeroSummary, toPlain возвращает HeroSummaryJson
  const mapHeroes = (res: ListAdminHeroesResponse) => ({
    heroes: res.heroes.map((h) => toPlain(HeroSummarySchema, h)),
    nextCursor: res.pagination?.nextCursor ?? "",
  });

  // Первичная загрузка
  const { data } = await useAsyncData("admin-heroes", async () => {
    const params = FILTER_PARAMS[statusFilter.value];
    const res = await heroAdmin.listHeroes({
      pagination: { pageSize: 15 },
      status: params.status,
      includeArchived: params.includeArchived,
    });
    return mapHeroes(res);
  });

  // Инициализация состояния из данных первичной загрузки
  heroes.value = data.value?.heroes ?? [];
  hasMore.value = (data.value?.nextCursor ?? "") !== "";
  if (hasMore.value) cursorStack.value[1] = data.value?.nextCursor ?? "";

  // Загрузка страницы для пагинации — мутирует состояние
  const load = async (p: number) => {
    loading.value = true;
    try {
      const params = FILTER_PARAMS[statusFilter.value];
      const res = await heroAdmin.listHeroes({
        pagination: { pageSize: 15, cursor: cursorStack.value[p] ?? "" },
        searchQuery: search.value,
        status: params.status,
        includeArchived: params.includeArchived,
      });

      const mapped = mapHeroes(res);

      heroes.value = mapped.heroes;
      hasMore.value = mapped.nextCursor !== "";

      if (hasMore.value) cursorStack.value[p + 1] = mapped.nextCursor;

      page.value = p;
    } finally {
      loading.value = false;
    }
  };

  // Debounce поиска
  let t: ReturnType<typeof setTimeout> | undefined;
  watch(search, () => {
    clearTimeout(t);
    t = setTimeout(() => {
      cursorStack.value = [""];
      load(0);
    }, 400);
  });

  // Сброс пагинации при смене фильтра
  watch(statusFilter, () => {
    cursorStack.value = [""];
    load(0);
  });

  const remove = (h: HeroSummaryJson) => {
    const archived = isArchived(h.status);
    const message = archived
      ? `Окончательно удалить запись «${h.lastName} ${h.firstName}»? Это действие необратимо.`
      : `Удалить запись «${h.lastName} ${h.firstName}»? Запись будет перемещена в архив.`;
    const header = archived ? "Окончательное удаление" : "Удаление героя";

    confirm.require({
      message,
      header,
      rejectLabel: "Отмена",
      acceptLabel: archived ? "Удалить навсегда" : "Удалить",
      acceptClass: "p-button-danger",
      accept: async () => {
        await heroAdmin.deleteHero({ id: h.id, hardDelete: archived });
        toast.add({
          severity: "success",
          summary: archived ? "Удалено окончательно" : "Удалено",
          detail: archived ? "Запись физически удалена" : "Запись перемещена в архив",
          life: 3000,
        });
        await load(page.value);
      },
    });
  };

  useHead({ title: "Герои — канцелярия" });
</script>

<template>
  <Card class="mb-4">
    <template #title>Герои</template>

    <template #subtitle>
      Реестр карточек: поиск, фильтрация, редактирование, архив
      <Tag v-if="statusFilter !== 'all'" :value="STATUS_OPTIONS.find((o) => o.value === statusFilter)?.label"
        severity="info" />
    </template>

    <template #content>
      <div class="admin_card_general">
        <div class="admin_card_filters">
          <label class="w-full">
            <span class="afield__label">Поиск по ФИО</span>
            <InputText type="text" v-model="search" fluid placeholder="Фамилия Имя Отчество" />
          </label>

          <label>
            <span class="afield__label">Статус</span>
            <Select v-model="statusFilter" :options="STATUS_OPTIONS" option-label="label" option-value="value"
              style="min-width: 14rem" />
          </label>
        </div>

        <Button as="router-link" to="/admin/heroes/new" class="admin_card_btn" label="+ Новый герой" />
      </div>
    </template>
  </Card>

  <DataTable :value="heroes" :loading="loading" data-key="id" class="atable" table-style="min-width: 55rem">
    <Column header="Фото" style="width: 56px">
      <template #body="{ data }">
        <img v-if="data.mainThumbnailUrl" :src="data.mainThumbnailUrl" alt=""
          class="h-13 w-10.5 border object-cover grayscale-60" style="border-color: var(--emh-bronze)" />
        <span v-else class="text-[#c4c8cf]">—</span>
      </template>
    </Column>

    <Column header="ФИО">
      <template #body="{ data }">
        <NuxtLink :to="`/admin/heroes/${data.id}`" class="font-semibold" style="color: var(--emh-ink)">
          {{ data.lastName }} {{ data.firstName }} {{ data.middleName }}
        </NuxtLink>
      </template>
    </Column>

    <Column header="Звание">
      <template #body="{ data }">{{ data.rank || "—" }}</template>
    </Column>

    <Column header="Годы">
      <template #body="{ data }">
        <span class="num">
          {{ formatFlexibleYear(data.birthDateInfo, data.birthDate) }} —
          {{ formatFlexibleYear(data.deathDateInfo, data.deathDate) }}
        </span>
      </template>
    </Column>

    <Column header="Статус" style="width: 140px">
      <template #body="{ data }">
        <button v-if="data.status" type="button" class="status-filter-btn"
          v-tooltip.bottom="`Показать только: ${statusMeta(data.status).label}`"
          @click="applyStatusFilter(data.status)">
          <Tag :value="statusMeta(data.status).label" :severity="statusMeta(data.status).severity" />
        </button>
        <span v-else class="text-[#c4c8cf]">—</span>
      </template>
    </Column>

    <Column header="Награды" style="width: 90px">
      <template #body="{ data }">
        <span v-if="data.awardNames?.length">{{ data.awardNames.length }}</span>
        <span v-else class="text-[#c4c8cf]">—</span>
      </template>
    </Column>

    <Column header="" style="width: 160px">
      <template #body="{ data }">
        <div class="flex justify-end gap-1 whitespace-nowrap text-right">
          <Button as="router-link" :to="`/admin/heroes/${data.id}`" text size="small" label="Изменить" />

          <Button text size="small" severity="danger" :label="isArchived(data.status) ? 'Удалить' : 'В архив'"
            @click="remove(data)" />
        </div>
      </template>
    </Column>

    <template #empty>
      <div class="p-6 text-center" style="color: var(--emh-muted)">Ничего не найдено</div>
    </template>
  </DataTable>

  <!-- Курсорная пагинация -->
  <div class="mt-3 flex gap-3">
    <Button outlined :disabled="page === 0" label="← Назад" @click="load(page - 1)" severity="secondary" />
    <Button outlined :disabled="!hasMore" label="Вперёд →" @click="load(page + 1)" severity="secondary" />
  </div>
</template>

<style scoped>
  .admin_card_general {
    display: grid;
    grid-template-columns: auto max-content;
    column-gap: 2.5rem;
  }

  .admin_card_filters {
    display: grid;
    grid-template-columns: 1fr 14rem;
    gap: 1.2rem;
    align-items: end;
    width: 100%;
  }

  .admin_card_btn {
    align-content: end;
  }

  /* Кнопка-обёртка тега статуса: наследуем вид тега, убираем нативные стили */
  .status-filter-btn {
    appearance: none;
    border: none;
    background: none;
    padding: 0;
    margin: 0;
    cursor: pointer;
    font: inherit;
    line-height: 0;
  }

  .status-filter-btn:hover :deep(.p-tag) {
    filter: brightness(1.08);
    box-shadow: 0 1px 6px -2px rgba(26, 28, 32, 0.4);
  }
</style>