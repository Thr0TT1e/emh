<script setup lang="ts">
  import { enumFromJson } from "@bufbuild/protobuf";
  import { useToast } from "primevue/usetoast";
  import { toPlain } from "~/lib/pb";
  import type { HeroLocationTypeJson, LocationTypeJson } from "~/sdk/emh/v1/enums_emh_pb";
  import { HeroLocationType, HeroLocationTypeSchema } from "~/sdk/emh/v1/enums_emh_pb";
  import type { HeroLocationJson } from "~/sdk/emh/v1/hero_pb";
  import type { LocationJson } from "~/sdk/emh/v1/location_pb";
  import { LocationSchema } from "~/sdk/emh/v1/location_pb";

  const props = defineProps<{ heroId: string; locations: HeroLocationJson[] }>();
  const emit = defineEmits<{ changed: [] }>();
  const { location, heroAdmin } = useApi();
  const toast = useToast();

  // Тип СВЯЗИ героя с локацией (HeroLocationType) — для таблицы и attach.type.
  const typeOptions = [
    { label: "Место рождения", value: HeroLocationType.BIRTH },
    { label: "Место гибели", value: HeroLocationType.DEATH },
    { label: "Место захоронения", value: HeroLocationType.BURIAL },
    { label: "Место жительства", value: HeroLocationType.RESIDENCE },
  ];

  const locationTypeLabel = (v?: HeroLocationTypeJson): string => {
    if (v === undefined) return "Место";
    const num = enumFromJson(HeroLocationTypeSchema, v);
    return typeOptions.find((o) => o.value === num)?.label ?? v;
  };

  // Тип САМОЙ локации (LocationType) — только для подсказки в результатах поиска.
  const GEO_TYPE_LABELS: Record<LocationTypeJson, string> = {
    LOCATION_TYPE_UNSPECIFIED: "Место",
    LOCATION_TYPE_COUNTRY: "Страна",
    LOCATION_TYPE_REGION: "Регион",
    LOCATION_TYPE_CITY: "Город",
    LOCATION_TYPE_VILLAGE: "Село / посёлок",
    LOCATION_TYPE_CEMETERY: "Кладбище / мемориал",
  };
  const geoTypeLabel = (t?: LocationTypeJson): string =>
    t === undefined ? "Место" : GEO_TYPE_LABELS[t];

  // ---------------------------------------------------------------------------
  // Поиск локации через PrimeVue AutoComplete.
  // Серверный поиск: клиентской фильтрации нет, список = ответ ListLocations.
  // debounce (delay) и порог (minLength) встроены в компонент.
  // ---------------------------------------------------------------------------
  const selectedLocation = ref<LocationJson | null>(null);
  const results = ref<LocationJson[]>([]);

  const onComplete = async (event: { query: string }) => {
    const query = event.query.trim();
    if (query.length < 2) {
      results.value = [];
      return;
    }
    try {
      const res = await location.listLocations({ searchQuery: query });
      results.value = (res.locations ?? []).map((l) => toPlain(LocationSchema, l));
    } catch {
      results.value = [];
    }
  };

  // attach.type — числовой enum для отправки в protobuf-es.
  const attach = reactive({
    locationId: "",
    locationName: "",
    type: HeroLocationType.BIRTH,
  });
  const busy = ref(false);

  // Выбор из списка → заполняем attach (активирует кнопку «Добавить»).
  watch(selectedLocation, (l) => {
    attach.locationId = l?.id ?? "";
    attach.locationName = l?.name ?? "";
  });

  const add = async () => {
    if (!attach.locationId) return;
    busy.value = true;
    try {
      await heroAdmin.addHeroLocation({
        heroId: props.heroId,
        locationId: attach.locationId,
        type: attach.type,
      });
      selectedLocation.value = null;
      attach.locationId = "";
      attach.locationName = "";
      toast.add({ severity: "success", summary: "Локация добавлена", life: 2500 });
      emit("changed");
    } catch (e: any) {
      toast.add({ severity: "error", summary: "Ошибка добавления", detail: e?.message ?? "", life: 5000 });
    } finally {
      busy.value = false;
    }
  };

  // l.type приходит из toJson как строка (HeroLocationTypeJson) —
  // преобразуем в числовой enum через enumFromJson.
  const remove = async (l: HeroLocationJson) => {
    if (!l.locationId || l.type === undefined) return;
    try {
      await heroAdmin.removeHeroLocation({
        heroId: props.heroId,
        locationId: l.locationId,
        type: enumFromJson(HeroLocationTypeSchema, l.type),
      });
      toast.add({ severity: "info", summary: "Локация удалена", life: 2500 });
      emit("changed");
    } catch (e: any) {
      toast.add({ severity: "error", summary: "Ошибка удаления", detail: e?.message ?? "", life: 5000 });
    }
  };
</script>

<template>
  <Card>
    <template #title>
      <h2 class="apanel__title">Места</h2>
    </template>

    <template #content>
      <DataTable v-if="locations.length" :value="locations" striped-rows class="atable" table-style="min-width: 50rem">
        <Column field="type" header="Тип">
          <template #body="{ data }">
            {{ locationTypeLabel(data.type) }}
          </template>
        </Column>

        <Column field="location.name" header="Место">
          <template #body="{ data }">
            {{ data.location?.name }}
            <em v-if="data.location?.historicalName" style="color: var(--emh-muted)">
              ({{ data.location.historicalName }})
            </em>
          </template>
        </Column>

        <Column header="" style="width: 150px">
          <template #body="{ data }">
            <Button text size="small" severity="danger" label="Снять" @click="remove(data)" />
          </template>
        </Column>
      </DataTable>
      <p v-else style="color: var(--emh-muted)">Мест пока нет.</p>

      <!-- Поиск локации: AutoComplete (серверный поиск, без клиентской фильтрации) -->
      <div class="mt-4">
        <span class="afield__label">Место (поиск)</span>
        <AutoComplete v-model="selectedLocation" :suggestions="results" option-label="name" :min-length="2" :delay="300"
          :force-selection="true" :placeholder="attach.locationName || 'Начните вводить…'"
          class="w-full hl-autocomplete" @complete="onComplete">
          <template #option="slotProps">
            <div class="flex flex-col gap-0.5 py-0.5">
              <span class="font-semibold">{{ slotProps.option.name }}</span>
              <span v-if="slotProps.option.historicalName || slotProps.option.type" class="text-xs"
                style="color: var(--emh-muted)">
                {{ slotProps.option.historicalName }}
                {{ slotProps.option.historicalName && slotProps.option.type ? "·" : "" }}
                {{ geoTypeLabel(slotProps.option.type) }}
              </span>
            </div>
          </template>
        </AutoComplete>
      </div>

      <div class="hl__container mt-4">
        <label>
          <span class="afield__label">Тип связи</span>
          <Select v-model="attach.type" :options="typeOptions" optionLabel="label" optionValue="value" class="w-full" />
        </label>
        <Button :loading="busy" :disabled="!attach.locationId" label="Добавить" @click="add" />
      </div>
    </template>
  </Card>
</template>

<style scoped>
  .hl__container {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0 0.9rem;
    align-items: end;
  }

  /* AutoComplete на всю ширину контейнера */
  :deep(.hl-autocomplete.p-autocomplete) {
    width: 100%;
  }

  :deep(.hl-autocomplete .p-autocomplete-input) {
    width: 100%;
  }
</style>