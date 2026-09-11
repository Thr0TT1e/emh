<script setup lang="ts">
  import { useToast } from "primevue/usetoast";
  import type { HeroConflictJson } from "~/sdk/emh/v1/hero_pb";

  const props = defineProps<{ heroId: string; conflicts: HeroConflictJson[] }>();
  const emit = defineEmits<{ changed: [] }>();
  const { conflict, heroAdmin } = useApi();
  const toast = useToast();

  const { data: catalog } = await useAsyncData("conflicts-catalog", async () => {
    const res = await conflict.listConflicts({});
    return res.conflicts.map((c) => ({ id: c.id, name: c.name }));
  });

  const attach = reactive({ conflictId: "", specificLocation: "", rankAtConflict: "" });
  const busy = ref(false);

  const add = async () => {
    if (!attach.conflictId) return;
    busy.value = true;
    try {
      await heroAdmin.addHeroConflict({
        heroId: props.heroId, conflictId: attach.conflictId,
        specificLocation: attach.specificLocation, rankAtConflict: attach.rankAtConflict,
      });
      attach.conflictId = ""; attach.specificLocation = ""; attach.rankAtConflict = "";
      toast.add({ severity: "success", summary: "Конфликт добавлен", life: 2500 });
      emit("changed");
    } finally { busy.value = false; }
  };

  const remove = async (conflictId: string) => {
    await heroAdmin.removeHeroConflict({ heroId: props.heroId, conflictId });
    toast.add({ severity: "info", summary: "Конфликт снят", life: 2500 });
    emit("changed");
  };
</script>

<template>
  <Card>
    <template #title>
      <h2 class="apanel__title">Боевой путь</h2>
    </template>

    <template #content>
      <DataTable v-if="conflicts.length" :value="conflicts" striped-rows class="atable" table-style="min-width: 50rem">
        <Column field="conflictName" header="Конфликт" />

        <Column field="awardDate" header="Место" class="num">
          <template #body="{ data }">
            {{ data.specificLocation || "—" }}
          </template>
        </Column>

        <Column field="decreeNumber" header="Звание">
          <template #body="{ data }">
            {{ data.rankAtConflict || "—" }}
          </template>
        </Column>

        <Column header="" style="width: 150px">
          <template #body="{ data }">
            <Button text size="small" severity="danger" label="Снять" @click="remove(data.conflictId!)" />
          </template>
        </Column>
      </DataTable>
      <p v-else style="color:var(--emh-muted)">Конфликтов пока нет.</p>

      <div class="hc__conlict mt-4">
        <label class="m-0">
          <span class="afield__label">Конфликт</span>
          <Select v-model="attach.conflictId" :options="catalog ?? []" optionLabel="name" optionValue="id"
            placeholder="Выберите…" class="w-full" />
        </label>

        <label class="m-0">
          <span class="afield__label">Место</span>
          <InputText v-model="attach.specificLocation" />
        </label>

        <label class="m-0">
          <span class="afield__label">Звание</span>
          <InputText v-model="attach.rankAtConflict" />
        </label>

        <Button :loading="busy" :disabled="!attach.conflictId" label="Добавить" @click="add" />
      </div>
    </template>
  </Card>
</template>

<style scoped>
  .hc__conlict {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr auto;
    gap: 0 .9rem;
    align-items: end
  }
</style>
