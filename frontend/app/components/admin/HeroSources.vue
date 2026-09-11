<script setup lang="ts">
  import { useToast } from "primevue/usetoast";
  import type { HeroSourceJson } from "~/sdk/emh/v1/hero_pb";

  const props = defineProps<{ heroId: string; sources: HeroSourceJson[] }>();
  const emit = defineEmits<{ changed: [] }>();
  const { heroAdmin } = useApi();
  const toast = useToast();

  // source_type — строка в контракте, значения: vk / website / archive / book
  const typeOptions = [
    { label: "ВКонтакте", value: "vk" },
    { label: "Веб-сайт", value: "website" },
    { label: "Архив", value: "archive" },
    { label: "Книга", value: "book" },
  ];

  const typeLabel = (v?: string): string =>
    typeOptions.find((o) => o.value === v)?.label ?? v ?? "Источник";

  const attach = reactive({
    title: "",
    url: "",
    sourceType: "website",
    excerpt: "",
  });
  const busy = ref(false);

  const add = async () => {
    if (!attach.url) return;
    busy.value = true;
    try {
      await heroAdmin.addHeroSource({
        heroId: props.heroId,
        url: attach.url,
        title: attach.title,
        sourceType: attach.sourceType,
        excerpt: attach.excerpt,
      });
      attach.title = "";
      attach.url = "";
      attach.excerpt = "";
      toast.add({ severity: "success", summary: "Источник добавлен", life: 2500 });
      emit("changed");
    } catch (e: any) {
      toast.add({ severity: "error", summary: "Ошибка добавления", detail: e?.message ?? "", life: 5000 });
    } finally {
      busy.value = false;
    }
  };

  const remove = async (sourceId: string) => {
    try {
      await heroAdmin.removeHeroSource({ heroId: props.heroId, sourceId });
      toast.add({ severity: "info", summary: "Источник удалён", life: 2500 });
      emit("changed");
    } catch (e: any) {
      toast.add({ severity: "error", summary: "Ошибка удаления", detail: e?.message ?? "", life: 5000 });
    }
  };
</script>

<template>
  <Card>
    <template #title>
      <h2 class="apanel__title">Источники</h2>
    </template>

    <template #content>
      <DataTable v-if="sources.length" :value="sources" striped-rows class="atable" table-style="min-width: 50rem">
        <Column field="sourceType" header="Тип">
          <template #body="{ data }">
            {{ typeLabel(data.sourceType) }}
          </template>
        </Column>

        <Column field="title" header="Название" class="num">
          <template #body="{ data }">
            <a :href="data.url" target="_blank" rel="noopener" style="font-weight: 600">
              {{ data.title || data.url }}
            </a>
          </template>
        </Column>

        <Column field="excerpt" header="Фрагмент" class="hs__excerpt">
          <template #body="{ data }">
            {{ data.excerpt || "—" }}
          </template>
        </Column>

        <Column header="">
          <template #body="{ data }">
            <Button text size="small" severity="danger" label="Удалить" @click="remove(data.id!)" />
          </template>
        </Column>
      </DataTable>
      <p v-else style="color: var(--emh-muted)">Источников пока нет.</p>

      <div class="hs__field1">
        <label>
          <span class="afield__label">Название</span>
          <InputText v-model="attach.title" class="w-full" />
        </label>

        <label>
          <span class="afield__label">URL *</span>
          <InputText v-model="attach.url" placeholder="https://…" class="w-full" />
        </label>
      </div>

      <div class="hs__field2">
        <label>
          <span class="afield__label">Тип источника</span>
          <Select v-model="attach.sourceType" :options="typeOptions" optionLabel="label" optionValue="value"
            class="w-full" />
        </label>

        <label>
          <span class="afield__label">Фрагмент / цитата</span>
          <InputText v-model="attach.excerpt" class="w-full" />
        </label>

        <Button :loading="busy" :disabled="!attach.url" label="Добавить" @click="add" />
      </div>
    </template>
  </Card>
</template>

<style scoped>
  .hs__excerpt {
    color: var(--emh-muted);
    font-size: 0.85rem
  }

  .hs__field1 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0 1.2rem
  }

  .hs__field2 {
    display: grid;
    grid-template-columns: 1fr 2fr auto;
    gap: 0 0.9rem;
    align-items: end
  }
</style>