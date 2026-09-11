<script setup lang="ts">
  import { fromJson } from "@bufbuild/protobuf";
  import { useToast } from "primevue/usetoast";
  import { useCachePurge } from '~/composables/useCachePurge';
  import { useExtractionLinks, type ExtractionLinkReport } from '~/composables/useExtractionLinks';
  import { toEnum } from "~/lib/pb";
  import { FlexibleDateSchema, type FlexibleDateJson } from "~/sdk/emh/v1/common_pb";
  import { PublicationStatus, type PublicationStatusJson } from "~/sdk/emh/v1/enums_emh_pb";
  import type { ExtractFromSubmissionResponseJson, ExtractHeroDataResponseJson } from "~/sdk/emh/v1/extraction_pb";
  import type { HeroDetailJson } from "~/sdk/emh/v1/hero_pb";

  const { purgeHero } = useCachePurge()
  const { applyExtractionLinks } = useExtractionLinks()

  // Общее извлечение: подходит и из панели (текст/URL), и из модерации заявки
  type ExtractionData = ExtractHeroDataResponseJson | ExtractFromSubmissionResponseJson;


  // Полный ответ извлечения — сохраняем для авто-привязки связей после CreateHero
  const extractionData = ref<ExtractionData | null>(null)

  // ---------------------------------------------------------------------------
  // Типы и Props
  // ---------------------------------------------------------------------------
  type Initial = {
    firstName: string; lastName: string; middleName: string;
    rank: string; nickname: string; unit: string; position: string;
    serviceBranch: string; causeOfDeath: string;
    // Старые поля (оставляем для обратной совместимости, если вдруг придут)
    birthDate: string | null; deathDate: string | null; serviceStartDate: string | null;
    // Новые гибкие даты
    birthDateInfo?: FlexibleDateJson | null;
    deathDateInfo?: FlexibleDateJson | null;
    serviceStartDateInfo?: FlexibleDateJson | null;

    shortBio: string; fullBio: string; status: PublicationStatusJson; memberships: string[];
  };

  const props = defineProps<{
    mode: "create" | "edit";
    heroId?: string;
    initial?: Initial;
    initialExtraction?: ExtractionData;
  }>();
  const emit = defineEmits<{ saved: [id: string, updated?: HeroDetailJson] }>();

  const { heroAdmin } = useApi();
  const toast = useToast();
  const router = useRouter();

  // ---------------------------------------------------------------------------
  // Reactive Form State
  // ---------------------------------------------------------------------------
  const form = reactive({
    lastName: props.initial?.lastName ?? "",
    firstName: props.initial?.firstName ?? "",
    middleName: props.initial?.middleName ?? "",
    nickname: props.initial?.nickname ?? "",
    unit: props.initial?.unit ?? "",
    position: props.initial?.position ?? "",
    serviceBranch: props.initial?.serviceBranch ?? "",
    causeOfDeath: props.initial?.causeOfDeath ?? "",
    rank: props.initial?.rank ?? "",

    // Гибкие даты (используем новые поля, если они есть)
    birthDateInfo: props.initial?.birthDateInfo ?? undefined,
    deathDateInfo: props.initial?.deathDateInfo ?? undefined,
    serviceStartDateInfo: props.initial?.serviceStartDateInfo ?? undefined,

    shortBio: props.initial?.shortBio ?? "",
    fullBio: props.initial?.fullBio ?? "",
    status: toEnum(PublicationStatus, props.initial?.status, PublicationStatus.DRAFT),
    memberships: [...(props.initial?.memberships ?? [])],
  });

  const statusOptions = [
    { label: "Черновик", value: PublicationStatus.DRAFT },
    { label: "Опубликовано", value: PublicationStatus.PUBLISHED },
    { label: "Архив", value: PublicationStatus.ARCHIVED },
  ];

  const busy = ref(false);
  const membershipInput = ref("");
  const addMembership = () => {
    const v = membershipInput.value.trim();
    if (v && !form.memberships.includes(v)) form.memberships.push(v);
    membershipInput.value = "";
  };
  const removeMembership = (i: number) => form.memberships.splice(i, 1);

  // Нормализация ФИО: первая буква заглавная, остальные строчные.
  // Учитывает дефисные фамилии: "ИВАНОВ-ПЕТРОВ" → "Иванов-Петров"
  const capitalizeName = (val: string): string => {
    if (!val) return val;
    return val
      .split('-')
      .map(part => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
      .join('-');
  };

  const onBlurName = (field: 'lastName' | 'firstName' | 'middleName') => {
    form[field] = capitalizeName(form[field]);
  };

  // ---------------------------------------------------------------------------
  // Dirty-checking: собираем fieldMask из реально изменённых полей.
  // ---------------------------------------------------------------------------
  const isFlexibleDateEmpty = (fd?: FlexibleDateJson | null) =>
    !fd || fd.precision === 'DATE_PRECISION_UNSPECIFIED' || fd.precision === 'DATE_PRECISION_UNKNOWN';

  const eqFlexibleDate = (a?: FlexibleDateJson | null, b?: FlexibleDateJson | null): boolean => {
    const aEmpty = isFlexibleDateEmpty(a);
    const bEmpty = isFlexibleDateEmpty(b);
    if (aEmpty && bEmpty) return true; // Оба пустые/неизвестные — считаем равными
    if (aEmpty || bEmpty) return false;

    return a!.precision === b!.precision &&
      a!.anchorDate === b!.anchorDate &&
      a!.displayText === b!.displayText;
  };

  const eqArray = (a: unknown, b: unknown): boolean => {
    if (Array.isArray(a) && Array.isArray(b)) {
      if (a.length !== b.length) return false;
      const sa = [...a].sort();
      const sb = [...b].sort();
      return sa.every((x, i) => x === sb[i]);
    }
    return a === b;
  };

  type FieldSpec = {
    key: keyof typeof form;
    mask: string;
    isObject?: boolean; // Для глубокого сравнения объектов
  };

  const FIELD_SPECS: FieldSpec[] = [
    { key: "firstName", mask: "first_name" },
    { key: "lastName", mask: "last_name" },
    { key: "middleName", mask: "middle_name" },
    { key: "nickname", mask: "nickname" },
    { key: "unit", mask: "unit" },
    { key: "position", mask: "position" },
    { key: "serviceBranch", mask: "service_branch" },
    { key: "causeOfDeath", mask: "cause_of_death" },
    { key: "rank", mask: "rank" },

    // Новые гибкие даты
    { key: "birthDateInfo", mask: "birth_date_info", isObject: true },
    { key: "deathDateInfo", mask: "death_date_info", isObject: true },
    { key: "serviceStartDateInfo", mask: "service_start_date_info", isObject: true },

    { key: "shortBio", mask: "short_bio" },
    { key: "fullBio", mask: "full_bio" },
    { key: "status", mask: "status" },
    { key: "memberships", mask: "memberships" },
  ];

  function buildFieldMask(): string[] {
    if (!props.initial) return FIELD_SPECS.map((f) => f.mask);

    const init = props.initial as Record<string, any>;
    const mask: string[] = [];

    for (const f of FIELD_SPECS) {
      const current = form[f.key];
      const original = init[f.key as keyof Initial];

      let isEqual: boolean;
      if (f.isObject) {
        isEqual = eqFlexibleDate(current as FlexibleDateJson | undefined, original as FlexibleDateJson | undefined);
      } else if (Array.isArray(current)) {
        isEqual = eqArray(current, original);
      } else {
        // Для строк: пустая строка ↔ null/undefined считаем одинаковым
        const aEmpty = current === "" || current === null || current === undefined;
        const bEmpty = original === "" || original === null || original === undefined;
        if (aEmpty && bEmpty) continue;
        isEqual = current === original;
      }

      if (!isEqual) mask.push(f.mask);
    }
    return mask;
  }

  // ---------------------------------------------------------------------------
  // Нормализация payload (Правило контракта п.8: очистка даты)
  // ---------------------------------------------------------------------------
  const prepareDateForApi = (fd?: FlexibleDateJson | null) => {
    // Если дата пустая или_UNSPECIFIED, отправляем UNKNOWN, чтобы бэкенд её очистил
    const json = (!fd || fd.precision === 'DATE_PRECISION_UNSPECIFIED')
      ? { precision: 'DATE_PRECISION_UNKNOWN' as const, displayText: '' }
      : fd;

    // fromJson преобразует JSON-представление (где anchorDate - строка)
    // в полноценный protobuf Message, который ожидает Connect-клиент.
    return fromJson(FlexibleDateSchema, json);
  };

  // ---------------------------------------------------------------------------
  // Парсинг дат из LLM-извлечения (форматы: YYYY | YYYY-MM | YYYY-MM-DD)
  // ---------------------------------------------------------------------------
  const parseExtractedDate = (dateStr?: string): FlexibleDateJson | null => {
    const s = dateStr?.trim();
    if (!s) return null;

    // YYYY-MM-DD → точная дата
    if (/^\d{4}-\d{2}-\d{2}$/.test(s)) {
      return { precision: 'DATE_PRECISION_EXACT', anchorDate: s, displayText: '' };
    }

    // YYYY-MM → месяц и год (якорь на первое число)
    if (/^\d{4}-\d{2}$/.test(s)) {
      return { precision: 'DATE_PRECISION_MONTH', anchorDate: `${s}-01`, displayText: '' };
    }

    // YYYY → только год (якорь на 1 января)
    if (/^\d{4}$/.test(s)) {
      return { precision: 'DATE_PRECISION_YEAR', anchorDate: `${s}-01-01`, displayText: '' };
    }

    // Нераспознанное — сохраняем как текстовое описание
    return { precision: 'DATE_PRECISION_UNKNOWN', displayText: s };
  };

  // ---------------------------------------------------------------------------
  // Save
  // ---------------------------------------------------------------------------
  async function save() {
    busy.value = true;

    // Нормализуем ФИО перед отправкой
    form.lastName = capitalizeName(form.lastName);
    form.firstName = capitalizeName(form.firstName);
    form.middleName = capitalizeName(form.middleName);

    const payload = {
      firstName: form.firstName,
      lastName: form.lastName,
      middleName: form.middleName,
      nickname: form.nickname,
      unit: form.unit,
      position: form.position,
      serviceBranch: form.serviceBranch,
      causeOfDeath: form.causeOfDeath,
      rank: form.rank,
      // Используем новую функцию для конвертации в MessageInit
      birthDateInfo: prepareDateForApi(form.birthDateInfo),
      deathDateInfo: prepareDateForApi(form.deathDateInfo),
      serviceStartDateInfo: prepareDateForApi(form.serviceStartDateInfo),
      shortBio: form.shortBio,
      fullBio: form.fullBio,
      status: form.status,
      memberships: form.memberships,
    };

    try {
      if (props.mode === "create") {
        const res = await heroAdmin.createHero(payload)
        toast.add({ severity: "success", summary: "Создано", detail: "Карточка героя создана", life: 3000 })

        // авто-привязка извлечённых связей
        if (extractionData.value) {
          const report = await applyExtractionLinks(res.id, extractionData.value)
          showLinksReport(report)
          extractionData.value = null
        }

        // Purge cache для списка и главной (новый герой)
        await purgeHero(res.id)

        emit("saved", res.id)
        router.push(`/admin/heroes/${res.id}`)
        return
      }

      const fieldMask = buildFieldMask()
      if (fieldMask.length === 0) {
        toast.add({ severity: "info", summary: "Нет изменений", detail: "Форма не менялась", life: 2500 })
        busy.value = false
        return
      }

      const res = await heroAdmin.updateHero({
        id: props.heroId!,
        ...payload,
        fieldMask,
      })

      const updatedHero = res.hero as HeroDetailJson | undefined
      toast.add({ severity: "success", summary: "Сохранено", detail: "Изменения сохранены", life: 3000 })

      // Purge cache для обновлённого героя
      await purgeHero(props.heroId!)

      emit("saved", props.heroId!, updatedHero)
    } catch (e: any) {
      toast.add({ severity: "error", summary: "Ошибка", detail: e?.message ?? "Не удалось сохранить", life: 5000 })
    } finally {
      busy.value = false
    }
  }

  const showLinksReport = (report: ExtractionLinkReport) => {
    if (report.applied.length) {
      toast.add({
        severity: 'success',
        summary: `Привязано связей: ${report.applied.length}`,
        detail: report.applied.join('\n'),
        life: 6000,
      });
    }
    if (report.skipped.length) {
      toast.add({
        severity: 'warn',
        summary: `Пропущено (нет в справочниках): ${report.skipped.length}`,
        detail: report.skipped.join('\n'),
        life: 8000,
      });
    }
    if (report.errors.length) {
      toast.add({
        severity: 'error',
        summary: `Ошибки привязки: ${report.errors.length}`,
        detail: report.errors.join('\n'),
        life: 8000,
      });
    }
  };

  // ---------------------------------------------------------------------------
  // Применение извлечённых данных из ExtractionPanel
  // ---------------------------------------------------------------------------
  const applyExtractedData = (data: ExtractionData) => {
    // Сохраняем для авто-привязки связей после создания героя
    extractionData.value = data;

    const hero = data.hero;
    if (!hero) return;

    // Заполняем основные поля
    if (hero.lastName) form.lastName = capitalizeName(hero.lastName);
    if (hero.firstName) form.firstName = capitalizeName(hero.firstName);
    if (hero.middleName) form.middleName = capitalizeName(hero.middleName);
    if (hero.nickname) form.nickname = hero.nickname;
    if (hero.rank) form.rank = hero.rank;
    if (hero.unit) form.unit = hero.unit;
    if (hero.position) form.position = hero.position;
    if (hero.serviceBranch) form.serviceBranch = hero.serviceBranch;
    if (hero.causeOfDeath) form.causeOfDeath = hero.causeOfDeath;
    if (hero.shortBio) form.shortBio = hero.shortBio;
    if (hero.fullBio) form.fullBio = hero.fullBio;
    if (hero.memberships?.length) form.memberships = [...hero.memberships];

    // Гибкие даты: парсим формат и выставляем корректную точность
    const birth = parseExtractedDate(hero.birthDate);
    if (birth) form.birthDateInfo = birth;

    const death = parseExtractedDate(hero.deathDate);
    if (death) form.deathDateInfo = death;

    const serviceStart = parseExtractedDate(hero.serviceStartDate);
    if (serviceStart) form.serviceStartDateInfo = serviceStart;

    toast.add({
      severity: 'success',
      summary: 'Данные применены',
      detail: 'Проверьте и дополните форму перед сохранением',
      life: 3000,
    });
  };

  // Применяем извлечение, переданное извне (например, из модерации заявок).
  // immediate: true — сработает и при монтировании, и если проп придёт позже.
  watch(
    () => props.initialExtraction,
    (data) => {
      if (props.mode === 'create' && data) {
        applyExtractedData(data);
      }
    },
    { immediate: true }
  );
</script>

<template>
  <!-- Панель извлечения данных (только в режиме создания) -->
  <AdminExtractionPanel v-if="mode === 'create'" class="mb-4" @apply="applyExtractedData" />

  <Card>
    <template #title>
      <h2 class="apanel__title">Основные сведения</h2>
    </template>

    <template #content>
      <form @submit.prevent="save">
        <!-- ФИО -->
        <div class="hf__col-3">
          <label>
            <span class="afield__label">Фамилия *</span>
            <InputText v-model="form.lastName" class="w-full" @blur="onBlurName('lastName')" required />
          </label>

          <label>
            <span class="afield__label">Имя *</span>
            <InputText v-model="form.firstName" class="w-full" @blur="onBlurName('firstName')" required />
          </label>

          <label>
            <span class="afield__label">Отчество</span>
            <InputText v-model="form.middleName" class="w-full" @blur="onBlurName('middleName')" />
          </label>
        </div>

        <!-- Гибкие даты (замена старых input[type=date]) -->
        <div class="hf__col-3">
          <AdminFlexibleDateInput v-model="form.birthDateInfo" label="Дата рождения" />
          <AdminFlexibleDateInput v-model="form.serviceStartDateInfo" label="Начало службы" />
          <AdminFlexibleDateInput v-model="form.deathDateInfo" label="Дата гибели" />
        </div>

        <!-- Звание, должность, позывной -->
        <div class="hf__col-3">
          <label>
            <span class="afield__label">Позывной</span>
            <InputText v-model="form.nickname" class="w-full" />
          </label>

          <label>
            <span class="afield__label">Звание</span>
            <InputText v-model="form.rank" class="w-full" />
          </label>

          <label>
            <span class="afield__label">Должность</span>
            <InputText v-model="form.position" class="w-full" />
          </label>
        </div>

        <!-- Ведомство и подразделение -->
        <div class="hf__col-2">
          <label>
            <span class="afield__label">Ведомство / род войск</span>
            <InputText v-model="form.serviceBranch" class="w-full" />
          </label>

          <label>
            <span class="afield__label">Подразделение</span>
            <InputText v-model="form.unit" class="w-full" />
          </label>
        </div>

        <label>
          <span class="afield__label">Причина гибели</span>
          <InputText v-model="form.causeOfDeath" class="w-full" />
        </label>

        <!-- Членство -->
        <div>
          <span class="afield__label">Членство в организациях</span>
          <div class="hf__row-flex">
            <InputText v-model="membershipInput" class="w-full" placeholder="Добавить и нажать Enter"
              @keyup.enter.prevent="addMembership" />
            <Button type="button" outlined label="+" @click="addMembership" />
          </div>
          <div v-if="form.memberships.length" class="mt-2 flex flex-wrap gap-2">
            <span v-for="(m, i) in form.memberships" :key="m" class="emh-chip">
              {{ m }}
              <button type="button" class="emh-chip__remove" @click="removeMembership(i)">×</button>
            </span>
          </div>
        </div>

        <!-- Биографии -->
        <label>
          <span class="afield__label">Краткая биография ({{ form.shortBio.length }}/500)</span>
          <Textarea v-model="form.shortBio" class="w-full" :maxlength="500" rows="2" />
        </label>
        <label>
          <span class="afield__label">Полный рассказ</span>
          <Textarea v-model="form.fullBio" class="w-full" rows="8" />
        </label>

        <!-- Статус -->
        <div class="hf__col-2 mb-3">
          <label>
            <span class="afield__label">Статус публикации</span>
            <Select v-model="form.status" :options="statusOptions" option-label="label" option-value="value"
              class="w-full" /></label>
        </div>

        <!-- Кнопки -->
        <div class="hf__row-flex">
          <Button type="submit" :loading="busy" :label="mode === 'create' ? 'Создать героя' : 'Сохранить'" />
          <Button as="router-link" to="/admin/heroes" outlined label="Отмена" />
        </div>
      </form>
    </template>
  </Card>
</template>

<style scoped>
  .hf__col-3 {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0 1.2rem;
    margin-bottom: 1rem;
  }

  .hf__col-2 {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0 1.2rem;
    margin-bottom: 1rem;
  }

  .hf__row-flex {
    display: flex;
    gap: .6rem;
  }

  .emh-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.8rem;
    padding: 0.25rem 0.6rem;
    border: 1px solid var(--emh-bronze);
    color: #8a6a3c;
    background: rgba(176, 141, 87, 0.08);
  }

  .emh-chip__remove {
    border: 0;
    background: none;
    cursor: pointer;
    color: inherit;
    padding: 0;
    line-height: 1;
    font-size: 1rem;
  }

  .emh-chip__remove:hover {
    color: var(--emh-crimson);
  }
</style>
