<script setup lang="ts">
  import { useAuth } from "~/composables/useAuth";
  import { PublicationStatus } from "~/sdk/emh/v1/enums_emh_pb";
  import AccountGroupOutlineIcon from '~icons/mdi/account-group-outline?width=1.25em&height=1.25em';
  import AccountPlusOutlineIcon from '~icons/mdi/account-plus-outline?width=1.25em&height=1.25em';
  import InboxFullOutlineIcon from '~icons/mdi/inbox-full-outline?width=1.25em&height=1.25em';
  import KeyOutlineIcon from '~icons/mdi/key-outline?width=1.25em&height=1.25em';

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
      viewBox: '0 0 24 24',
      icon: `<path fill="currentColor" d="M12 5a3.5 3.5 0 0 0-3.5 3.5A3.5 3.5 0 0 0 12 12a3.5 3.5 0 0 0 3.5-3.5A3.5 3.5 0 0 0 12 5m0 2a1.5 1.5 0 0 1 1.5 1.5A1.5 1.5 0 0 1 12 10a1.5 1.5 0 0 1-1.5-1.5A1.5 1.5 0 0 1 12 7M5.5 8A2.5 2.5 0 0 0 3 10.5c0 .94.53 1.75 1.29 2.18c.36.2.77.32 1.21.32s.85-.12 1.21-.32c.37-.21.68-.51.91-.87A5.42 5.42 0 0 1 6.5 8.5v-.28c-.3-.14-.64-.22-1-.22m13 0c-.36 0-.7.08-1 .22v.28c0 1.2-.39 2.36-1.12 3.31c.12.19.25.34.4.49a2.48 2.48 0 0 0 1.72.7c.44 0 .85-.12 1.21-.32c.76-.43 1.29-1.24 1.29-2.18A2.5 2.5 0 0 0 18.5 8M12 14c-2.34 0-7 1.17-7 3.5V19h14v-1.5c0-2.33-4.66-3.5-7-3.5m-7.29.55C2.78 14.78 0 15.76 0 17.5V19h3v-1.93c0-1.01.69-1.85 1.71-2.52m14.58 0c1.02.67 1.71 1.51 1.71 2.52V19h3v-1.5c0-1.74-2.78-2.72-4.71-2.95M12 16c1.53 0 3.24.5 4.23 1H7.77c.99-.5 2.7-1 4.23-1"/>`,
      page: "/admin/heroes"
    },
    {
      label: "Заявок на модерации",
      value: counters.value.submissions,
      viewBox: '0 0 24 24',
      icon: `<path fill="currentColor" d="M19 3c1.1 0 2 .9 2 2v14c0 1.1-.9 2-2 2H5c-1.1 0-2-.9-2-2V5c0-1.1.9-2 2-2zM5 19h4.4a4.13 4.13 0 0 1-1.27-2H5zm14 0v-2h-3.13c-.22.78-.66 1.47-1.27 2zm0-4V5H5v10h5v1c0 2.67 4 2.67 4 0v-1zM7 7h10v2H7zm10 4v2H7v-2z"/>`,
      page: "/admin/submissions"
    },
  ]);

  // Наполненность справочников
  const refStats = computed(() => [
    {
      label: "Наград",
      value: counters.value.awards,
      viewBox: '0 0 24 24',
      icon: `<path fill="currentColor" d="M18 2c-.9 0-2 1-2 2H8c0-1-1.1-2-2-2H2v9c0 1 1 2 2 2h2.2c.4 2 1.7 3.7 4.8 4v2.08C8 19.54 8 22 8 22h8s0-2.46-3-2.92V17c3.1-.3 4.4-2 4.8-4H20c1 0 2-1 2-2V2zM6 11H4V4h2zm10 .5c0 1.93-.58 3.5-4 3.5c-3.41 0-4-1.57-4-3.5V6h8zm4-.5h-2V4h2z"/>`,
      page: "/admin/awards"
    },
    {
      label: "Конфликтов",
      value: counters.value.conflicts,
      viewBox: '0 0 24 24',
      icon: `<path fill="currentColor" d="M7 5h16v4h-1v1h-6a1 1 0 0 0-1 1v1a2 2 0 0 1-2 2H9.62c-.38 0-.73.22-.9.56l-2.45 4.89c-.17.34-.51.55-.89.55H2s-3 0 1-6c0 0 3-4-1-4V5h1l.5-1h3zm7 7v-1a1 1 0 0 0-1-1h-1s-1 1 0 2a2 2 0 0 1-2-2a1 1 0 0 0-1 1v1a1 1 0 0 0 1 1h3a1 1 0 0 0 1-1"/>`,
      page: "/admin/conflicts"
    },
    {
      label: "Локаций",
      value: counters.value.locations,
      viewBox: '0 0 32 32',
      icon: `<path fill="currentColor" d="m16 24l-6.09-8.6A8.14 8.14 0 0 1 16 2a8.08 8.08 0 0 1 8 8.13a8.2 8.2 0 0 1-1.8 5.13Zm0-20a6.07 6.07 0 0 0-6 6.13a6.2 6.2 0 0 0 1.49 4L16 20.52L20.63 14A6.24 6.24 0 0 0 22 10.13A6.07 6.07 0 0 0 16 4"/><circle cx="16" cy="9" r="2" fill="currentColor"/><path fill="currentColor" d="M28 12h-2v2h2v14H4V14h2v-2H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h24a2 2 0 0 0 2-2V14a2 2 0 0 0-2-2"/>`,
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
              <div class="stat-tile__icon">
                <svg xmlns="http://www.w3.org/2000/svg" width="1.2em" height="1.2em" :viewBox="s.viewBox"
                  v-html="s.icon" />
              </div>

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
                <div class="stat-tile__icon">
                  <svg xmlns="http://www.w3.org/2000/svg" width="1.2em" height="1.2em" :viewBox="s.viewBox"
                    v-html="s.icon" />
                </div>

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
          <Button as="router-link" to="/admin/heroes/new">
            <AccountPlusOutlineIcon />
            <span>Новый герой</span>
          </Button>
          <Button as="router-link" to="/admin/heroes" outlined>
            <AccountGroupOutlineIcon />
            <span>Реестр героев</span>
          </Button>
          <Button as="router-link" to="/admin/submissions" outlined>
            <InboxFullOutlineIcon />
            <span>Модерация заявок</span>
          </Button>
          <Button as="router-link" to="/admin/keys" outlined>
            <KeyOutlineIcon />
            <span>API-ключи</span>
          </Button>
        </div>
      </template>
    </Card>
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