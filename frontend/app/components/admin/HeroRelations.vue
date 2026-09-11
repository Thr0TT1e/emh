<script setup lang="ts">
  import { useToast } from "primevue/usetoast";
  import type { HeroRelationJson, HeroSummary } from "~/sdk/emh/v1/hero_pb";

  const props = defineProps<{ heroId: string; relations: HeroRelationJson[] }>();
  const emit = defineEmits<{ changed: [] }>();
  const { hero, heroAdmin } = useApi();
  const toast = useToast();

  // relation_type — строка в контракте, значения: father_son / comrade / commander
  const typeOptions = [
    { label: "Отец и сын", value: "father_son" },
    { label: "Однополчанин", value: "comrade" },
    { label: "Командир", value: "commander" },
  ];

  const typeLabel = (v?: string): string =>
    typeOptions.find((o) => o.value === v)?.label ?? v ?? "Связь";

  // ---------------------------------------------------------------------------
  // Поиск героя-цели через PrimeVue AutoComplete.
  // Серверный поиск: клиентской фильтрации нет, список = ответ ListHeroes.
  // debounce (delay) и порог (minLength) встроены в компонент.
  // ---------------------------------------------------------------------------
  type HeroOption = HeroSummary & { fullName: string };

  const fullNameOf = (h: HeroSummary) =>
    [h.lastName, h.firstName, h.middleName].filter(Boolean).join(" ");

  const attach = reactive({
    toHeroId: "",
    toHeroName: "",
    relationType: "comrade",
    description: "",
  });
  const busy = ref(false);

  const selectedHero = ref<HeroOption | null>(null);
  const results = ref<HeroOption[]>([]);

  const onComplete = async (event: { query: string }) => {
    const query = event.query.trim();
    if (query.length < 2) {
      results.value = [];
      return;
    }
    try {
      const res = await hero.listHeroes({
        pagination: { pageSize: 8 },
        searchQuery: query,
      });
      // Исключаем текущего героя — нельзя связать героя с самим собой
      results.value = res.heroes
        .filter((h) => h.id !== props.heroId)
        .map((h) => ({ ...h, fullName: fullNameOf(h) }));
    } catch {
      results.value = [];
    }
  };

  // Выбор из списка → заполняем attach.toHeroId (активирует кнопку «Добавить»)
  watch(selectedHero, (h) => {
    attach.toHeroId = h?.id ?? "";
    attach.toHeroName = h?.fullName ?? "";
  });

  const add = async () => {
    if (!attach.toHeroId) return;
    busy.value = true;
    try {
      await heroAdmin.addHeroRelation({
        fromHeroId: props.heroId,
        toHeroId: attach.toHeroId,
        relationType: attach.relationType,
        description: attach.description,
      });
      selectedHero.value = null;
      attach.toHeroId = "";
      attach.toHeroName = "";
      attach.description = "";
      toast.add({ severity: "success", summary: "Связь добавлена", life: 2500 });
      emit("changed");
    } catch (e: any) {
      toast.add({ severity: "error", summary: "Ошибка добавления", detail: e?.message ?? "", life: 5000 });
    } finally {
      busy.value = false;
    }
  };

  // Удаление — только по id связи (согласно RemoveHeroRelationRequest)
  const remove = async (relationId: string) => {
    try {
      await heroAdmin.removeHeroRelation({ id: relationId });
      toast.add({ severity: "info", summary: "Связь удалена", life: 2500 });
      emit("changed");
    } catch (e: any) {
      toast.add({ severity: "error", summary: "Ошибка удаления", detail: e?.message ?? "", life: 5000 });
    }
  };
</script>

<template>
  <Card>
    <template #title>
      <h2 class="apanel__title">Связи с героями</h2>
    </template>

    <template #content>
      <DataTable v-if="relations.length" :value="relations" striped-rows class="atable" table-style="min-width: 50rem">
        <Column field="relationType" header="Тип">
          <template #body="{ data }">{{ typeLabel(data.relationType) }}</template>
        </Column>

        <Column field="relatedHeroName" header="Герой">
          <template #body="{ data }">
            <NuxtLink :to="`/admin/heroes/${data.toHeroId}`">
              {{ data.relatedHeroName || "Герой" }}
            </NuxtLink>
          </template>
        </Column>

        <Column field="description" header="Описание" class="hr__desc">
          <template #body="{ data }">{{ data.description || "—" }}</template>
        </Column>

        <Column header="" style="width: 150px">
          <template #body="{ data }">
            <Button text size="small" severity="danger" label="Удалить" @click="remove(data.id!)" />
          </template>
        </Column>
      </DataTable>
      <p v-else style="color: var(--emh-muted)">Связей пока нет.</p>

      <!-- Поиск героя-цели: AutoComplete (серверный поиск, без клиентской фильтрации) -->
      <div class="mt-4">
        <span class="afield__label">Герой (поиск по ФИО)</span>
        <AutoComplete v-model="selectedHero" :suggestions="results" option-label="fullName" :min-length="2" :delay="300"
          :force-selection="true" placeholder="Начните вводить фамилию…" class="w-full hr-autocomplete"
          @complete="onComplete">
          <template #option="slotProps">
            <div class="flex flex-col gap-0.5 py-0.5">
              <span class="font-semibold">{{ slotProps.option.fullName }}</span>
              <span v-if="slotProps.option.rank || slotProps.option.unit" class="text-xs"
                style="color: var(--emh-muted)">
                {{ [slotProps.option.rank, slotProps.option.unit].filter(Boolean).join(" · ") }}
              </span>
            </div>
          </template>
        </AutoComplete>
      </div>

      <div class="hr__field mt-4">
        <label>
          <span class="afield__label">Тип связи</span>
          <Select v-model="attach.relationType" :options="typeOptions" optionLabel="label" optionValue="value"
            class="w-full" />
        </label>

        <label>
          <span class="afield__label">Описание</span>
          <InputText v-model="attach.description" class="w-full" />
        </label>
        <Button :loading="busy" :disabled="!attach.toHeroId" label="Добавить" @click="add" />
      </div>
    </template>
  </Card>
</template>

<style scoped>
  .hr__desc {
    color: var(--emh-muted);
    font-size: 0.85rem;
  }

  .hr__field {
    display: grid;
    grid-template-columns: 1fr 2fr auto;
    gap: 0 0.9rem;
    align-items: end;
  }

  /* AutoComplete на всю ширину контейнера */
  :deep(.hr-autocomplete.p-autocomplete) {
    width: 100%;
  }

  :deep(.hr-autocomplete .p-autocomplete-input) {
    width: 100%;
  }
</style>