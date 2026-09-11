<script setup lang="ts">
  import { useConfirm } from "primevue/useconfirm";
  import { useToast } from "primevue/usetoast";
  import { formatDate } from "~/lib/format";

  definePageMeta({ layout: "admin", middleware: "admin" });

  const { apiKeyAdmin } = useApi();
  const toast = useToast();
  const confirm = useConfirm();

  const { data, refresh } = await useAsyncData("admin-keys", async () => {
    const res = await apiKeyAdmin.listApiKeys({ includeRevoked: true });
    return res.apiKeys;
  });

  const keys = computed(() => data.value ?? []);

  // Создание
  const showCreate = ref(false);
  const created = ref(""); // полный ключ, показывается один раз
  const form = reactive({ name: "", description: "", role: "admin" });
  const busy = ref(false);

  const create = async () => {
    busy.value = true;
    try {
      const res = await apiKeyAdmin.createApiKey({ name: form.name, description: form.description, role: form.role });
      created.value = res.fullKey;
      form.name = ""; form.description = "";
      await refresh();
    } finally { busy.value = false; }
  };

  const copy = async () => {
    await navigator.clipboard.writeText(created.value);
    toast.add({ severity: "info", summary: "Скопировано", life: 2000 });
  };

  const revoke = (k: any) => {
    confirm.require({
      message: `Отозвать ключ «${k.name}»? Действие необратимо, ключ перестанет работать немедленно.`,
      header: "Отзыв ключа",
      rejectLabel: "Отмена",
      acceptLabel: "Отозвать",
      acceptClass: "p-button-danger",
      accept: async () => {
        await apiKeyAdmin.revokeApiKey({ id: k.id });
        toast.add({ severity: "success", summary: "Ключ отозван", life: 3000 });
        await refresh();
      },
    });
  };

  const isActive = (k: any) => !k.revokedAt && (!k.expiresAt || new Date(k.expiresAt) > new Date());

  useHead({ title: "API-ключи — канцелярия" });
</script>

<template>
  <Card class="mb-4">
    <template #title>API-ключи</template>
    <template #subtitle>Доступ для скриптов и интеграций · отзыв без перезапуска</template>

    <template #content>
      <div class="admin-pagehead mb-3">
        <Button label="+ Новый ключ" @click="showCreate = true; created = ''" />
      </div>

      <Dialog v-model:visible="showCreate" modal header="Новый API-ключ" :style="{ width: '34rem' }">
        <template v-if="created">
          <p style="margin-top:0">Ключ создан. <b>Сохраните его сейчас</b> — повторно он не показывается.</p>

          <div
            style="display:flex;gap:.6rem;align-items:center;background:var(--emh-bg);border:1px solid var(--emh-line);padding:.7rem .8rem">
            <code class="mono" style="word-break:break-all;flex:1">{{ created }}</code>
            <Button size="small" outlined label="Копировать" @click="copy" />
          </div>

          <div style="margin-top:1.2rem;text-align:right"><Button label="Готово" @click="showCreate = false" /></div>
        </template>

        <template v-else>
          <label>
            <span class="afield__label">Имя *</span>
            <InputText v-model="form.name" class="w-full" placeholder="CI/CD Pipeline" />
          </label>

          <label>
            <span class="afield__label">Описание</span>
            <InputText v-model="form.description" class="w-full" placeholder="Для чего этот ключ" />
          </label>

          <label>
            <span class="afield__label">Роль</span>
            <InputText v-model="form.role" class="w-full" />
          </label>

          <div class="mt-3" style="text-align:right">
            <Button :loading="busy" :disabled="!form.name" label="Создать" @click="create" />
          </div>
        </template>
      </Dialog>
    </template>
  </Card>

  <DataTable :value="keys" :loading="busy" data-key="id" class="atable" table-style="min-width: 50rem">
    <Column header="Имя" style="width: 56px">
      <template #body="{ data }">
        <span v-if="data.name" class="font-semibold">{{ data.name }}</span>
        <span v-else class="text-[#c4c8cf]">—</span>
        <div v-if="data.description" style="font-size:.78rem;color:var(--emh-muted)">{{ data.description }}</div>
      </template>
    </Column>

    <Column header="Key ID">
      <template #body="{ data }">
        <span class="mono">{{ data.keyId.slice(0, 8) }}…</span>
      </template>
    </Column>

    <Column header="Роль">
      <template #body="{ data }">{{ data.role }}</template>
    </Column>

    <Column header="Создал">
      <template #body="{ data }">{{ data.createdBy || "—" }}</template>
    </Column>

    <Column header="Последнее использование">
      <template #body="{ data }">
        <span class="num">{{ data.lastUsedAt ? formatDate(data.lastUsedAt) : "не использовался" }}</span>
      </template>
    </Column>

    <Column header="Статус">
      <template #body="{ data }">
        <Tag v-if="isActive(data)" severity="success" value="активен" />
        <Tag v-else severity="danger" value="отозван" />
      </template>
    </Column>

    <Column header="" style="width: 150px">
      <template #body="{ data }">
        <Button v-if="isActive(data)" text size="small" severity="danger" label="Отозвать" @click="revoke(data)" />
      </template>
    </Column>

    <template #empty>
      <div class="p-6 text-center" style="color: var(--emh-muted)">Ничего не найдено</div>
    </template>
  </DataTable>
</template>