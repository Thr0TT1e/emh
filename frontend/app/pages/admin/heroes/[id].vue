<script setup lang="ts">
  import { useCachePurge } from '~/composables/useCachePurge';
  import { formatFlexibleYear } from "~/lib/format";
  import { toPlain } from "~/lib/pb";
  import type { PublicationStatusJson } from "~/sdk/emh/v1/enums_emh_pb";
  import type { HeroDetailJson } from "~/sdk/emh/v1/hero_pb";
  import { HeroDetailSchema } from "~/sdk/emh/v1/hero_pb";
  import RestartIcon from '~icons/carbon/restart?width=1em&height=1em';
  import ExternalLinkIcon from '~icons/mdi/external-link?width=1.25em&height=1.25em';

  definePageMeta({ layout: "admin", middleware: "admin" });

  const { purgeHero } = useCachePurge()
  const route = useRoute();
  const { hero } = useApi();
  const id = route.params.id as string;

  const { data, error, refresh } = await useAsyncData(`admin-hero-${id}`, async () => {
    const res = await hero.getHero({ id });

    if (!res.hero) throw new Error("GetHero вернул пустой ответ: поле hero отсутствует");

    return toPlain(HeroDetailSchema, res.hero);
  });

  // ---------------------------------------------------------------------------
  // После UpdateHero бэкенд возвращает полный HeroDetail — используем его
  // напрямую, без повторного GetHero (бэкенд-фик из снапшота, техдолг #14).
  // ---------------------------------------------------------------------------
  const onSaved = (savedId: string, updatedHero?: HeroDetailJson) => {
    if (updatedHero) data.value = updatedHero;
  };

  const onChanged = async () => {
    await refresh() // для M:N-панелей, которые мутируют связи отдельно

    // Invalidate cache после изменения связей
    if (id) {
      await purgeHero(id)
    }
  }

  // Мета статуса публикации для Tag
  const STATUS_META: Record<PublicationStatusJson, { label: string; severity: "success" | "warn" | "secondary" }> = {
    PUBLICATION_STATUS_UNSPECIFIED: { label: "—", severity: "secondary" },
    PUBLICATION_STATUS_DRAFT: { label: "Черновик", severity: "warn" },
    PUBLICATION_STATUS_PUBLISHED: { label: "Опубликовано", severity: "success" },
    PUBLICATION_STATUS_ARCHIVED: { label: "Архив", severity: "secondary" },
  };

  const statusMeta = computed(() => {
    const s = data.value?.status;
    return (s && STATUS_META[s]) || STATUS_META.PUBLICATION_STATUS_UNSPECIFIED;
  });

  const initials = computed(() => {
    const l = data.value?.summary?.lastName?.[0] ?? "";
    const f = data.value?.summary?.firstName?.[0] ?? "";
    return (l + f).toUpperCase() || "Г";
  });

  const fullBio = computed(() => {
    const s = data.value?.summary;
    const nickname = s?.nickname !== '' ? `«${s?.nickname}»` : '';

    return `${s?.lastName} ${s?.firstName} ${s?.middleName} «${nickname}»`
  });

  useHead({
    title: () => (data.value?.summary?.lastName ? `${data.value.summary.lastName} — правка` : "Герой"),
  });
</script>

<template>
  <Card v-if="error" class="mb-4">
    <template #content>
      <Message severity="error" :closable="false">
        {{ error.message || "Не удалось загрузить карточку героя" }}
      </Message>

      <Button class="mt-3" @click="refresh()">
        <RestartIcon />
        <span>Повторить</span>
      </Button>
    </template>
  </Card>

  <template v-else-if="data?.summary">
    <Card class="mb-4">
      <template #title>
        {{ data.summary.lastName }} {{ data.summary.firstName }}
        <span v-if="data.summary.middleName"> {{ data.summary.middleName }}</span>
      </template>

      <template #subtitle>
        {{ data.summary.rank || "звание не указано" }}
        <span v-if="data.position"> · {{ data.position }}</span>
      </template>

      <template #content>
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div class="flex items-center gap-4">
            <Image v-if="data.summary.mainThumbnailUrl" :src="data.summary.mainThumbnailUrl" :alt="fullBio"
              class="h-20 w-16 border object-cover grayscale-60" style="border-color: var(--emh-bronze)" />

            <Avatar v-else :label="initials" shape="circle" size="large" />

            <div class="flex flex-col gap-1">
              <Tag :value="statusMeta.label" :severity="statusMeta.severity" class="w-fit" />

              <span
                v-if="data.summary.birthDateInfo || data.summary.deathDateInfo || data.summary.birthDate || data.summary.deathDate"
                class="num text-sm" style="color: var(--emh-muted)">
                {{ formatFlexibleYear(data.summary.birthDateInfo, data.summary.birthDate) }} —
                {{ formatFlexibleYear(data.summary.deathDateInfo, data.summary.deathDate) }}
              </span>

              <span v-if="data.summary.unit" class="text-sm" style="color: var(--emh-muted)">
                {{ data.summary.unit }}
              </span>
            </div>
          </div>

          <Button as="router-link" :to="`/heroes/${id}`" outlined>
            <ExternalLinkIcon />
            <span>Открыть на сайте</span>
          </Button>
        </div>
      </template>
    </Card>

    <div class="hero-personal_container">
      <AdminHeroForm mode="edit" :hero-id="id" :initial="{
        firstName: data.summary.firstName!,
        lastName: data.summary.lastName!,
        middleName: data.summary.middleName!,
        rank: data.summary.rank!,
        nickname: data.summary.nickname!,
        unit: data.summary.unit!,
        position: data.position!,
        serviceBranch: data.summary.serviceBranch!,
        causeOfDeath: data.causeOfDeath!,

        // Старые поля (можно оставить для безопасности, но форма теперь смотрит в *Info)
        birthDate: data.summary.birthDate ?? null,
        deathDate: data.summary.deathDate ?? null,
        serviceStartDate: data.serviceStartDate ?? null,

        // НОВЫЕ ГИБКИЕ ДАТЫ
        birthDateInfo: data.summary.birthDateInfo ?? null,
        deathDateInfo: data.summary.deathDateInfo ?? null,
        serviceStartDateInfo: data.serviceStartDateInfo ?? null,

        shortBio: data.summary.shortBio!,
        fullBio: data.fullBio!,
        status: data.status!,
        memberships: data.memberships!,
      }" @saved="onSaved" />

      <AdminHeroPhotos :hero-id="id" @changed="onChanged" />
      <AdminHeroAwards :hero-id="id" :awards="data.awards ?? []" @changed="onChanged" />
      <AdminHeroConflicts :hero-id="id" :conflicts="data.conflicts ?? []" @changed="onChanged" />
      <AdminHeroLocations :hero-id="id" :locations="data.locations ?? []" @changed="onChanged" />
      <AdminHeroSources :hero-id="id" :sources="data.sources ?? []" @changed="onChanged" />
      <AdminHeroRelations :hero-id="id" :relations="data.relations ?? []" @changed="onChanged" />
    </div>
  </template>

  <div v-else class="flex justify-center py-16">
    <ProgressSpinner style="width: 40px; height: 40px" />
  </div>
</template>

<style scoped>
  .hero-personal_container {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(750px, 1fr));
    align-items: start;
    grid-gap: .8rem;
  }

</style>