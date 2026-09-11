<script setup lang="ts">
  import { useAuth } from "~/composables/useAuth";
  import { PublicationStatus } from "~/sdk/emh/v1/enums_emh_pb";

  definePageMeta({ layout: "admin", middleware: "admin" });

  const { claims } = useAuth();
  const { hero, award, conflict, location, submissionAdmin } = useApi();

  const counters = ref({ heroes: 0, submissions: 0, awards: 0, conflicts: 0, locations: 0 });

  onMounted(async () => {
    try {
      const [h, s, aw, cf, lc] = await Promise.all([
        hero.listHeroes({ pagination: { pageSize: 1 } }),
        submissionAdmin.listSubmissions({ pagination: { pageSize: 1 }, status: PublicationStatus.DRAFT }),
        award.listAwards({}),
        conflict.listConflicts({}),
        location.listLocations({}),
      ]);
      counters.value = {
        heroes: Number(h.pagination?.totalCount ?? 0),
        submissions: Number(s.pagination?.totalCount ?? 0),
        awards: aw.awards?.length ?? 0,
        conflicts: cf.conflicts?.length ?? 0,
        locations: lc.locations?.length ?? 0,
      };
    } catch {
      /* счётчики не критичны */
    }
  });

  // Ключевые показатели
  const mainStats = computed(() => [
    {
      label: "Героев в базе",
      value: counters.value.heroes,
      icon: "pi pi-users",
      page: "/admin/heroes"
    },
    {
      label: "Заявок на модерации",
      value: counters.value.submissions,
      icon: "pi pi-inbox",
      page: "/admin/submissions"
    },
  ]);

  // Наполненность справочников
  const refStats = computed(() => [
    {
      label: "Наград",
      value: counters.value.awards,
      icon: "pi pi-trophy",
      page: "/admin/awards"
    },
    {
      label: "Конфликтов",
      value: counters.value.conflicts,
      icon: "pi pi-flag",
      page: "/admin/conflicts"
    },
    {
      label: "Локаций",
      value: counters.value.locations,
      icon: "pi pi-map-marker",
      page: "/admin/locations"
    },
  ]);

  useHead({ title: "Сводка — канцелярия" });
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- Приветствие -->
    <Card>
      <template #title>Добрый день, {{ claims?.username ?? "модератор" }}</template>
      <template #subtitle>Краткая сводка по проекту «Вечная память героям»</template>
    </Card>

    <!-- Ключевые показатели -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <NuxtLink :to="s.page" v-for="s in mainStats" :key="s.label">
        <Card>
          <template #content>
            <div class="flex items-center gap-4">
              <div class="stat-tile__icon"><i :class="s.icon" /></div>
              <div>
                <div class="stat-tile__value">{{ s.value }}</div>
                <div class="stat-tile__label">{{ s.label }}</div>
              </div>
            </div>
          </template>
        </Card>
      </NuxtLink>

    </div>

    <!-- Справочники -->
    <div>
      <div class="mb-2 text-xs font-semibold uppercase tracking-wider" style="color: var(--p-text-muted-color)">
        Справочники
      </div>

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <NuxtLink :to="s.page" v-for="s in refStats" :key="s.label">
          <Card>
            <template #content>
              <div class="flex items-center gap-4">
                <div class="stat-tile__icon"><i :class="s.icon" /></div>
                <div>
                  <div class="stat-tile__value">{{ s.value }}</div>
                  <div class="stat-tile__label">{{ s.label }}</div>
                </div>
              </div>
            </template>
          </Card>
        </NuxtLink>
      </div>
    </div>

    <!-- Быстрые действия -->
    <Card>
      <template #title>Быстрые действия</template>
      <template #content>
        <div class="flex flex-wrap gap-3">
          <Button as="router-link" to="/admin/heroes/new" icon="pi pi-plus" label="Новый герой" />
          <Button as="router-link" to="/admin/heroes" outlined icon="pi pi-users" label="Реестр героев" />
          <Button as="router-link" to="/admin/submissions" outlined icon="pi pi-inbox" label="Модерация заявок" />
          <Button as="router-link" to="/admin/keys" outlined icon="pi pi-key" label="API-ключи" />
        </div>
      </template>
    </Card>0
  </div>
</template>

<style scoped>
  .stat-tile__icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 44px;
    border-radius: 10px;
    font-size: 1.2rem;
    background: var(--p-primary-50, var(--p-surface-100));
    color: var(--p-primary-color);
    flex-shrink: 0;
  }

  .stat-tile__value {
    font-size: 1.6rem;
    font-weight: 700;
    line-height: 1.1;
  }

  .stat-tile__label {
    font-size: 0.85rem;
    color: var(--p-text-muted-color);
  }
</style>