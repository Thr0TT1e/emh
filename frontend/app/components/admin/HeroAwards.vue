<script setup lang="ts">
  import { fromJson } from "@bufbuild/protobuf";
  import { useToast } from "primevue/usetoast";
  import { formatFlexibleDate } from "~/lib/format";
  import { FlexibleDateSchema, type FlexibleDateJson } from "~/sdk/emh/v1/common_pb";
  import type { HeroAwardJson } from "~/sdk/emh/v1/hero_pb";

  const props = defineProps<{ heroId: string; awards: HeroAwardJson[] }>();
  const emit = defineEmits<{ changed: [] }>();
  const { award, heroAdmin } = useApi();
  const toast = useToast();

  const { data: catalog } = await useAsyncData("awards-catalog", async () => {
    const res = await award.listAwards({});
    return res.awards.map((a) => ({ id: a.id, name: a.name }));
  });

  const attach = reactive({
    awardId: "",
    awardDateInfo: undefined as FlexibleDateJson | undefined,
    decreeNumber: "",
  });
  const busy = ref(false);

  // Конвертация гибкой даты в Message перед отправкой.
  // Паттерн идентичен prepareDateForApi из HeroForm: пустую/неуказанную дату
  // отправляем как UNKNOWN, чтобы бэкенд корректно её очистил.
  const prepareAwardDateForApi = (fd?: FlexibleDateJson | null) => {
    const json = (!fd || fd.precision === 'DATE_PRECISION_UNSPECIFIED')
      ? { precision: 'DATE_PRECISION_UNKNOWN' as const, displayText: '' }
      : fd;
    return fromJson(FlexibleDateSchema, json);
  };

  const add = async () => {
    if (!attach.awardId) return;
    busy.value = true;
    try {
      await heroAdmin.addHeroAward({
        heroId: props.heroId,
        awardId: attach.awardId,
        // award_date_info имеет приоритет над award_date (контракт).
        // award_date оставляем пустым: паттерн-валидация не применяется к пустой строке.
        awardDate: "",
        awardDateInfo: prepareAwardDateForApi(attach.awardDateInfo),
        decreeNumber: attach.decreeNumber,
      });
      attach.awardId = "";
      attach.awardDateInfo = undefined;
      attach.decreeNumber = "";
      toast.add({ severity: "success", summary: "Награда добавлена", life: 2500 });
      emit("changed");
    } finally { busy.value = false; }
  };

  const remove = async (awardId: string) => {
    await heroAdmin.removeHeroAward({ heroId: props.heroId, awardId });
    toast.add({ severity: "info", summary: "Награда снята", life: 2500 });
    emit("changed");
  };
</script>

<template>
  <Card>
    <template #title>
      <h2 class="apanel__title">Награды</h2>
    </template>

<template #content>
      <DataTable v-if="awards.length" :value="awards" striped-rows class="atable" table-style="min-width: 50rem">
        <Column field="awardName" header="Награда" />

        <Column header="Дата" class="num">
          <template #body="{ data }">
            {{ formatFlexibleDate(data.awardDateInfo, data.awardDate) }}
          </template>
</Column>

<Column field="decreeNumber" header="Приказ">
  <template #body="{ data }">
            {{ data.decreeNumber || "—" }}
          </template>
</Column>

<Column header="" style="width: 150px">
  <template #body="{ data }">
            <Button text size="small" severity="danger" label="Снять" @click="remove(data.awardId!)" />
          </template>
</Column>
</DataTable>
<p v-else style="color:var(--emh-muted)">Наград пока нет.</p>

<!-- Награда, приказ, кнопка — в одну строку -->
<div class="ha__award mt-4">
  <label class="m-0">
    <span class="afield__label">Награда</span>
    <Select v-model="attach.awardId" :options="catalog ?? []" optionLabel="name" optionValue="id"
      placeholder="Выберите…" class="w-full" />
  </label>

  <label class="m-0">
    <span class="afield__label">Приказ</span>
    <InputText v-model="attach.decreeNumber" class="w-full" />
  </label>

  <Button :loading="busy" :disabled="!attach.awardId" label="Добавить" @click="add" />
</div>

<!-- Гибкая дата: многострочный блок (точность + дата + текст),
           поэтому вынесена на отдельную строку -->
<div class="ha__date mt-3">
  <AdminFlexibleDateInput v-model="attach.awardDateInfo" label="Дата награждения" />
</div>
</template>
</Card>
</template>

<style scoped>
  .ha__award {
    display: grid;
    grid-template-columns: 2fr 1fr auto;
    gap: 0 .9rem;
    align-items: end;
  }

  /* Компонент даты сам несёт mb-4; прижимаем его к форме сверху */
  .ha__date :deep(.fdi-container) {
    margin-bottom: 0;
  }
</style>
<script setup lang="ts">
import { fromJson } from "@bufbuild/protobuf";
import { useToast } from "primevue/usetoast";
import { formatFlexibleDate } from "~/lib/format";
import { FlexibleDateSchema, type FlexibleDateJson } from "~/sdk/emh/v1/common_pb";
import type { HeroAwardJson } from "~/sdk/emh/v1/hero_pb";

const props = defineProps<{ heroId: string; awards: HeroAwardJson[] }>();
const emit = defineEmits<{ changed: [] }>();
const { award, heroAdmin } = useApi();
const toast = useToast();

const { data: catalog } = await useAsyncData("awards-catalog", async () => {
  const res = await award.listAwards({});
  return res.awards.map((a) => ({ id: a.id, name: a.name }));
});

const attach = reactive({
  awardId: "",
  awardDateInfo: undefined as FlexibleDateJson | undefined,
  decreeNumber: "",
});
const busy = ref(false);

// Конвертация гибкой даты в Message перед отправкой.
// Паттерн идентичен prepareDateForApi из HeroForm: пустую/неуказанную дату
// отправляем как UNKNOWN, чтобы бэкенд корректно её очистил.
const prepareAwardDateForApi = (fd?: FlexibleDateJson | null) => {
  const json = (!fd || fd.precision === 'DATE_PRECISION_UNSPECIFIED')
    ? { precision: 'DATE_PRECISION_UNKNOWN' as const, displayText: '' }
    : fd;
  return fromJson(FlexibleDateSchema, json);
};

const add = async () => {
  if (!attach.awardId) return;
  busy.value = true;
  try {
    await heroAdmin.addHeroAward({
      heroId: props.heroId,
      awardId: attach.awardId,
      // award_date_info имеет приоритет над award_date (контракт).
      // award_date оставляем пустым: паттерн-валидация не применяется к пустой строке.
      awardDate: "",
      awardDateInfo: prepareAwardDateForApi(attach.awardDateInfo),
      decreeNumber: attach.decreeNumber,
    });
    attach.awardId = "";
    attach.awardDateInfo = undefined;
    attach.decreeNumber = "";
    toast.add({ severity: "success", summary: "Награда добавлена", life: 2500 });
    emit("changed");
  } finally { busy.value = false; }
};

const remove = async (awardId: string) => {
  await heroAdmin.removeHeroAward({ heroId: props.heroId, awardId });
  toast.add({ severity: "info", summary: "Награда снята", life: 2500 });
  emit("changed");
};
</script>

<template>
  <Card>
    <template #title>
      <h2 class="apanel__title">Награды</h2>
    </template>

    <template #content>
      <DataTable v-if="awards.length" :value="awards" striped-rows class="atable" table-style="min-width: 50rem">
        <Column field="awardName" header="Награда" />

        <Column header="Дата" class="num">
          <template #body="{ data }">
            {{ formatFlexibleDate(data.awardDateInfo, data.awardDate) }}
          </template>
        </Column>

        <Column field="decreeNumber" header="Приказ">
          <template #body="{ data }">
            {{ data.decreeNumber || "—" }}
          </template>
        </Column>

        <Column header="" style="width: 150px">
          <template #body="{ data }">
            <Button text size="small" severity="danger" label="Снять" @click="remove(data.awardId!)" />
          </template>
        </Column>
      </DataTable>
      <p v-else style="color:var(--emh-muted)">Наград пока нет.</p>

      <!-- Награда, приказ, кнопка — в одну строку -->
      <div class="ha__award mt-4">
        <label class="m-0">
          <span class="afield__label">Награда</span>
          <Select v-model="attach.awardId" :options="catalog ?? []" optionLabel="name" optionValue="id"
            placeholder="Выберите…" class="w-full" />
        </label>

        <label class="m-0">
          <span class="afield__label">Приказ</span>
          <InputText v-model="attach.decreeNumber" class="w-full" />
        </label>

        <Button :loading="busy" :disabled="!attach.awardId" label="Добавить" @click="add" />
      </div>

      <!-- Гибкая дата: многострочный блок (точность + дата + текст),
           поэтому вынесена на отдельную строку -->
      <div class="ha__date mt-3">
        <AdminFlexibleDateInput v-model="attach.awardDateInfo" label="Дата награждения" />
      </div>
    </template>
  </Card>
</template>

<style scoped>
  .ha__award {
    display: grid;
    grid-template-columns: 2fr 1fr auto;
    gap: 0 .9rem;
    align-items: end;
  }

  /* Компонент даты сам несёт mb-4; прижимаем его к форме сверху */
  .ha__date :deep(.fdi-container) {
    margin-bottom: 0;
  }
</style>