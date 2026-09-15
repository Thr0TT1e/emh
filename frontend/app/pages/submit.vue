<script setup lang="ts">
  import { ConnectError } from '@connectrpc/connect';
  import { useToast } from "primevue/usetoast";
  import { useAnalytics } from '~/composables/useAnalytics';
  import { useDuplicateCheck } from '~/composables/useDuplicateCheck';
  import { isAlreadyExists, isRateLimited } from '~/lib/errors';
  import { formatFlexibleYear } from "~/lib/format";
  import { toPlain } from "~/lib/pb";
  import type { FlexibleDateJson } from "~/sdk/emh/v1/common_pb";
  import { HeroSummarySchema, type HeroSummaryJson } from "~/sdk/emh/v1/hero_pb";
  import { UploadType } from "~/sdk/emh/v1/media_pb";
  import SendAltIcon from '~icons/carbon/send-alt?width=1.25em&height=1.25em';
  import AttachmentIcon from '~icons/mdi/attachment?width=1.25em&height=1.25em';
  import FileOutlineIcon from '~icons/mdi/file-outline?width=1.25em&height=1.25em';

  const { hero, submission, media } = useApi();
  const { trackSubmissionCreate } = useAnalytics();
  const { checkLocal, recordSubmission } = useDuplicateCheck();
  const toast = useToast();
  const route = useRoute();

  // ---------------------------------------------------------------------------
  // Режим: новый герой / дополнение к существующему
  // ---------------------------------------------------------------------------
  type Mode = "new" | "supplement";
  const qHero = typeof route.query.hero === "string" ? route.query.hero : "";
  const qName = typeof route.query.name === "string" ? route.query.name : "";
  const mode = ref<Mode>(qHero ? "supplement" : "new");

  // ---------------------------------------------------------------------------
  // Отправитель
  // ---------------------------------------------------------------------------
  const submitterName = ref("");
  const submitterEmail = ref("");

  // ---------------------------------------------------------------------------
  // Цель заявки (режим supplement)
  // ---------------------------------------------------------------------------
  const targetHeroId = ref(qHero);
  const targetHeroName = ref(qName);

  // ---------------------------------------------------------------------------
  // Данные нового героя (режим new)
  // ---------------------------------------------------------------------------
  const newHero = reactive({
    lastName: "",
    firstName: "",
    middleName: "",
    birthDateInfo: undefined as FlexibleDateJson | undefined,
    deathDateInfo: undefined as FlexibleDateJson | undefined,
    rank: "",
    conflict: "",
  });

  // ---------------------------------------------------------------------------
  // Рассказ
  // ---------------------------------------------------------------------------
  const narrative = ref("");

  // ---------------------------------------------------------------------------
  // Поиск героя (для supplement без предзаполненного героя)
  // ---------------------------------------------------------------------------
  const heroSearch = ref("");
  const heroResults = ref<HeroSummaryJson[]>([]);
  const searching = ref(false);
  const showHeroPicker = ref(false);

  let searchTimer: ReturnType<typeof setTimeout> | undefined;

  watch(heroSearch, () => {
    clearTimeout(searchTimer);
    if (!heroSearch.value.trim()) {
      heroResults.value = [];
      showHeroPicker.value = false;
      return;
    }
    searchTimer = setTimeout(async () => {
      searching.value = true;
      try {
        const res = await hero.listHeroes({
          pagination: { pageSize: 10 },
          searchQuery: heroSearch.value,
        });
        heroResults.value = (res.heroes ?? []).map((h) => toPlain(HeroSummarySchema, h));
        showHeroPicker.value = true;
      } finally {
        searching.value = false;
      }
    }, 400);
  });

  const selectHero = (h: HeroSummaryJson) => {
    targetHeroId.value = h.id ?? "";
    targetHeroName.value = [h.lastName, h.firstName, h.middleName].filter(Boolean).join(" ");
    showHeroPicker.value = false;
    heroSearch.value = "";
  };

  const clearHero = () => {
    targetHeroId.value = "";
    targetHeroName.value = "";
  };

  // ---------------------------------------------------------------------------
  // Вложения (presigned URL → MinIO)
  // ---------------------------------------------------------------------------
  const attachments = ref<{ name: string; url: string }[]>([]);
  const uploading = ref(false);
  const fileInput = ref<HTMLInputElement | null>(null);

  const onFiles = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    const files = Array.from(input.files ?? []);
    if (!files.length) return;
    input.value = ""; // сбрасываем, чтобы повторный выбор того же файла сработал
    uploading.value = true;
    try {
      for (const file of files) {
        const { uploadUrl, publicUrl } = await media.getUploadUrl({
          type: UploadType.SUBMISSION_ATTACHMENT,
          filename: file.name,
          contentType: file.type || "application/octet-stream",
        });
        const put = await fetch(uploadUrl, {
          method: "PUT",
          headers: { "Content-Type": file.type || "application/octet-stream" },
          body: file,
        });
        if (!put.ok) throw new Error("MinIO: " + put.status);
        attachments.value.push({ name: file.name, url: publicUrl });
      }
    } catch (err: any) {
      toast.add({ severity: "error", summary: "Ошибка загрузки", detail: err?.message ?? "", life: 5000 });
    } finally {
      uploading.value = false;
    }
  };

  const removeAttachment = (i: number) => attachments.value.splice(i, 1);

  // ---------------------------------------------------------------------------
  // Отправка заявки
  // ---------------------------------------------------------------------------
  const busy = ref(false);
  const submittedId = ref("");

  const submit = async () => {
    if (!submitterName.value.trim()) {
      toast.add({ severity: "warn", summary: "Укажите ваше имя", life: 3000 });
      return;
    }
    if (submitterEmail.value && !/^\S+@\S+\.\S+$/.test(submitterEmail.value)) {
      toast.add({ severity: "warn", summary: "Проверьте корректность email", life: 3000 });
      return;
    }
    if (mode.value === "supplement" && !targetHeroId.value) {
      toast.add({ severity: "warn", summary: "Выберите героя, о котором хотите рассказать", life: 3000 });
      return;
    }
    if (mode.value === "new" && !newHero.lastName.trim() && !narrative.value.trim()) {
      toast.add({ severity: "warn", summary: "Заполните фамилию героя или расскажите о нём", life: 3000 });
      return;
    }

    busy.value = true;
    try {
      const payload: Record<string, unknown> = {
        mode: mode.value,
        narrative: narrative.value,
      };
      if (mode.value === "new") {
        payload.hero = {
          last_name: newHero.lastName,
          first_name: newHero.firstName,
          middle_name: newHero.middleName,
          birth_date_info: newHero.birthDateInfo,
          death_date_info: newHero.deathDateInfo,
          rank: newHero.rank,
          conflict: newHero.conflict,
        };
      }

      const payloadJson = JSON.stringify(payload);
      const submitTargetHeroId = mode.value === "supplement" ? targetHeroId.value : "";

      // Клиентская мягкая проверка дубликатов
      const localDup = checkLocal(submitTargetHeroId, payloadJson);
      if (localDup) {
        toast.add({
          severity: 'warn',
          summary: 'Возможный дубликат',
          detail: 'Вы недавно отправляли похожую заявку. Убедитесь, что данные не повторяются.',
          life: 6000,
        });
      }

      const res = await submission.createSubmission({
        submitterName: submitterName.value,
        submitterEmail: submitterEmail.value,
        targetHeroId: submitTargetHeroId,
        payloadJson,
        attachmentUrls: attachments.value.map((a) => a.url),
      });

      submittedId.value = res.submissionId;
      recordSubmission(submitTargetHeroId, payloadJson);

      trackSubmissionCreate({
        mode: mode.value,
        targetHeroId: mode.value === "supplement" ? targetHeroId.value : undefined,
        hasAttachments: attachments.value.length > 0,
      });

      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err: any) {
      console.log('[submit] catch block reached', {
        err,
        code: err?.code,
        message: err?.message,
        isConnectError: err instanceof ConnectError,
        isAlreadyExists: isAlreadyExists(err),
      });

      // ── 409 AlreadyExists: сервер отклонил дубликат активной заявки ──
      if (isAlreadyExists(err)) {
        toast.add({
          severity: 'warn',
          summary: 'Заявка уже существует',
          detail: 'Такая заявка уже находится на рассмотрении. Пожалуйста, дождитесь ответа модератора или отправьте заявку с дополнительными данными.',
          life: 8000,
        });
        return;
      }
      // ── 429 ResourceExhausted: превышен лимит запросов ──
      if (isRateLimited(err)) {
        toast.add({
          severity: 'warn',
          summary: 'Слишком много заявок',
          detail: 'Пожалуйста, подождите немного и попробуйте снова.',
          life: 6000,
        });
        return;
      }
      // ── Прочие ошибки ──
      toast.add({
        severity: "error",
        summary: "Не удалось отправить заявку",
        detail: err?.message ?? String(err),
        life: 5000,
      });
    } finally {
      busy.value = false;
    }
  };

  const resetForm = () => {
    submittedId.value = "";
    submitterName.value = "";
    submitterEmail.value = "";
    narrative.value = "";
    attachments.value = [];
    mode.value = "new";
    targetHeroId.value = "";
    targetHeroName.value = "";
    Object.assign(newHero, {
      lastName: "", firstName: "", middleName: "",
      birthDateInfo: undefined,
      deathDateInfo: undefined,
      rank: "", conflict: "",
    });
  };

  // ═══════════════════════════════════════════════════════════
  // SEO страницы «Сообщить о герое»
  // ═══════════════════════════════════════════════════════════
  const site = useSiteConfig();
  const canonical = `${site.url}/submit`;

  useSeoMeta({
    title: 'Сообщить о герое',
    ogTitle: 'Сообщить о герое',
    description:
      'Знаете о герое, которого нет в реестре, или хотите дополнить существующую запись? Отправьте нам — мы проверим по достоверным источникам.',
    ogDescription:
      'Знаете о герое, которого нет в реестре, или хотите дополнить существующую запись? Отправьте нам — мы проверим по достоверным источникам.',
    ogType: 'website',
    ogUrl: canonical,
    ogSiteName: site.name,
    ogLocale: 'ru_RU',
    robots: 'noindex, follow', // ← страница-форма, не индексируем
  });

  useHead({
    link: [
      { rel: 'canonical', href: canonical },
    ],
  });

  // OG-карточка для соцсетей — используем HomeCard (единообразие)
  defineOgImage('HomeCard', {
    totalHeroes: 0,   // На submit нет useHeroRegistry, не дублируем запрос
    totalConflicts: 0,
  });

  useSchemaOrg([
    defineWebPage({
      '@type': 'ContactPage',
      name: 'Сообщить о герое',
      description: 'Форма для добавления или дополнения записей в книге памяти.',
      url: canonical,
    }),
  ]);
</script>

<template>
  <main>
    <!-- Экран благодарности после отправки -->
    <section v-if="submittedId" class="submit-success">
      <div class="submit-success__inner">
        <span class="flame" aria-hidden="true"></span>
        <h1 class="submit-success__title">Спасибо! Ваша заявка принята</h1>
        <p class="submit-success__lead">
          Мы получили ваше сообщение и обязательно его рассмотрим. Каждая судьба важна.
          Если вы указали email — сообщим о результате модерации.
        </p>
        <div class="submit-success__actions">
          <Button as="router-link" to="/" outlined label="Вернуться к реестру" />
          <Button label="Отправить ещё одну заявку" @click="resetForm" />
        </div>
      </div>
    </section>

    <!-- Форма -->
    <section v-else class="submit">
      <header class="submit__head">
        <h1 class="submit__title">Сообщить о герое</h1>

        <p class="submit__lead">
          Знаете о герое, которого ещё нет в нашем реестре, или хотите дополнить существующую
          запись? Расскажите нам — мы проверим информацию по достоверным источникам и добавим её.
        </p>
      </header>

      <form class="submit__form" @submit.prevent="submit">
        <!-- Переключатель режима -->
        <div class="submit__mode">
          <button type="button" class="submit__mode-btn" :class="{ 'is-active': mode === 'new' }"
            @click="mode = 'new'">Добавить нового героя</button>

          <button type="button" class="submit__mode-btn" :class="{ 'is-active': mode === 'supplement' }"
            @click="mode = 'supplement'">Дополнить существующего</button>
        </div>

        <!-- Выбор героя (режим supplement) -->
        <div v-if="mode === 'supplement'" class="submit__block">
          <div v-if="targetHeroId" class="submit__selected-hero">
            <span class="submit__selected-label">Выбранный герой:</span>
            <strong>{{ targetHeroName || targetHeroId }}</strong>
            <button type="button" class="submit__clear-hero" @click="clearHero">изменить</button>
          </div>

          <div v-else>
            <label class="afield">
              <span class="afield__label">Найдите героя по ФИО *</span>
              <InputText v-model="heroSearch" placeholder="Иванов Иван Иванович" class="w-full" />
            </label>

            <div v-if="showHeroPicker && heroResults.length" class="submit__hero-results">
              <button v-for="h in heroResults" :key="h.id" type="button" class="submit__hero-result"
                @click="selectHero(h)">
                <span class="submit__hero-name">{{ h.lastName }} {{ h.firstName }} {{ h.middleName }}</span>
                <span class="submit__hero-dates">
                  {{ formatFlexibleYear(h.birthDateInfo, h.birthDate) }} —
                  {{ formatFlexibleYear(h.deathDateInfo, h.deathDate) }}
                </span>
              </button>
            </div>
            <p v-else-if="showHeroPicker && !heroResults.length && !searching" class="submit__no-results">
              Никого не нашли. Уточните написание или выберите «Добавить нового героя».
            </p>
          </div>
        </div>

        <!-- Данные нового героя (режим new) -->
        <div v-if="mode === 'new'" class="submit__block">
          <div class="submit__grid submit__grid--3">
            <label class="afield"><span class="afield__label">Фамилия *</span>
              <InputText v-model="newHero.lastName" class="w-full" />
            </label>

            <label class="afield"><span class="afield__label">Имя *</span>
              <InputText v-model="newHero.firstName" class="w-full" />
            </label>

            <label class="afield"><span class="afield__label">Отчество</span>
              <InputText v-model="newHero.middleName" class="w-full" />
            </label>
          </div>

          <div class="submit__grid submit__grid--4">
            <AdminFlexibleDateInput v-model="newHero.birthDateInfo" label="Дата рождения" />

            <AdminFlexibleDateInput v-model="newHero.deathDateInfo" label="Дата гибели" />

            <label class="afield"><span class="afield__label">Звание</span>
              <InputText v-model="newHero.rank" class="w-full" />
            </label>

            <label class="afield"><span class="afield__label">Конфликт</span>
              <InputText v-model="newHero.conflict" class="w-full" />
            </label>
          </div>
        </div>

        <!-- Рассказ -->
        <label class="afield">
          <span class="afield__label">
            {{ mode === 'new' ? 'Рассказ о герое' : 'Что вы хотите дополнить или исправить? *' }}
          </span>
          <Textarea v-model="narrative" :rows="6" class="w-full" :placeholder="mode === 'new'
            ? 'Расскажите о подвиге, боевом пути, обстоятельствах гибели…'
            : 'Опишите, какую информацию нужно добавить или исправить…'" />
        </label>

        <!-- Вложения -->
        <div class="afield">
          <span class="afield__label">Документы и фотографии</span>

          <input ref="fileInput" type="file" multiple class="submit__file-input" :disabled="uploading"
            @change="onFiles" />

          <Button outlined severity="secondary" :loading="uploading" @click="fileInput?.click()">
            <AttachmentIcon />
            <span v-if="uploading">Загружаем…</span>
            <span v-else>Прикрепить файлы</span>
          </Button>

          <ul v-if="attachments.length" class="submit__attachment-list">
            <li v-for="(a, i) in attachments" :key="a.url" class="submit__attachment">
              <FileOutlineIcon />
              <span class="submit__attachment-name">{{ a.name }}</span>

              <button type="button" class="submit__attachment-remove" aria-label="Удалить"
                @click="removeAttachment(i)">×</button>
            </li>
          </ul>
        </div>

        <!-- Контакты отправителя -->
        <div class="submit__grid submit__grid--2">
          <label class="afield"><span class="afield__label">Ваше имя *</span>
            <InputText v-model="submitterName" class="w-full" placeholder="Как к вам обращаться" />
          </label>

          <label class="afield"><span class="afield__label">Email (для обратной связи)</span>
            <InputText v-model="submitterEmail" type="email" class="w-full" placeholder="you@example.com" />
          </label>
        </div>

        <div class="submit__footer">
          <Button type="submit" :loading="busy">
            <SendAltIcon />
            <span>Отправить заявку</span>
          </Button>
          <p class="submit__note">
            Заявка попадёт на модерацию. Публикуются только проверенные данные из достоверных источников.
          </p>
        </div>
      </form>
    </section>
  </main>
</template>

<style scoped>

  /* ========================================================================= */
  /* Форма                                                                      */
  /* ========================================================================= */
  .submit {
    max-width: 46rem;
    margin: 0 auto;
    padding: 3.5rem 2rem 5rem;
  }

  .submit__head {
    margin-bottom: 2.5rem;
    text-align: center;
  }

  .submit__title {
    font-size: clamp(2rem, 5vw, 3rem);
    line-height: 1.1;
    margin: 0 0 1rem;
  }

  .submit__lead {
    max-width: 38rem;
    margin: 0 auto;
    color: var(--emh-muted);
    font-size: 1.05rem;
    line-height: 1.6;
  }

  .submit__form {
    display: flex;
    flex-direction: column;
    gap: 1.8rem;
  }

  /* Переключатель режима */
  .submit__mode {
    display: flex;
    gap: 0.4rem;
    border: 1px solid var(--emh-line);
    border-radius: 8px;
    padding: 4px;
    background: var(--emh-bg);
  }

  .submit__mode-btn {
    flex: 1;
    padding: 0.7rem 1rem;
    border: none;
    border-radius: 6px;
    background: transparent;
    font: inherit;
    font-size: 0.95rem;
    font-weight: 500;
    color: var(--emh-muted);
    cursor: pointer;
    transition: background 0.2s, color 0.2s;
  }

  .submit__mode-btn.is-active {
    background: #fff;
    color: var(--emh-ink);
    box-shadow: 0 1px 4px rgba(26, 28, 32, 0.12);
  }

  .submit__block {
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
  }

  .submit__grid {
    display: grid;
    gap: 0 1.2rem;
  }

  .submit__grid--2 {
    grid-template-columns: repeat(2, 1fr);
  }

  .submit__grid--3 {
    grid-template-columns: repeat(3, 1fr);
  }

  .submit__grid--4 {
    grid-template-columns: repeat(4, 1fr);
  }

  @media (max-width: 640px) {

    .submit__grid--2,
    .submit__grid--3,
    .submit__grid--4 {
      grid-template-columns: 1fr;
    }
  }

  /* Выбранный герой */
  .submit__selected-hero {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    padding: 0.9rem 1rem;
    border: 1px solid var(--emh-bronze);
    border-radius: 6px;
    background: rgba(176, 141, 87, 0.07);
  }

  .submit__selected-label {
    color: var(--emh-muted);
    font-size: 0.9rem;
  }

  .submit__clear-hero {
    margin-left: auto;
    border: none;
    background: none;
    color: var(--emh-crimson);
    cursor: pointer;
    font-size: 0.85rem;
    text-decoration: underline;
  }

  /* Результаты поиска героя */
  .submit__hero-results {
    display: flex;
    flex-direction: column;
    border: 1px solid var(--emh-line);
    border-radius: 6px;
    overflow: hidden;
    background: #fff;
  }

  .submit__hero-result {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 1rem;
    padding: 0.8rem 1rem;
    border: none;
    border-bottom: 1px solid var(--emh-line);
    background: none;
    font: inherit;
    text-align: left;
    cursor: pointer;
    transition: background 0.15s;
  }

  .submit__hero-result:last-child {
    border-bottom: none;
  }

  .submit__hero-result:hover {
    background: var(--emh-bg);
  }

  .submit__hero-name {
    font-weight: 600;
  }

  .submit__hero-dates {
    color: var(--emh-muted);
    font-size: 0.85rem;
  }

  .submit__no-results {
    color: var(--emh-muted);
    font-size: 0.9rem;
  }

  /* Вложения */
  .submit__file-input {
    display: none;
  }

  .submit__attachment-list {
    list-style: none;
    margin: 0.8rem 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .submit__attachment {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.5rem 0.8rem;
    border: 1px solid var(--emh-line);
    border-radius: 4px;
    background: var(--emh-bg);
    font-size: 0.9rem;
  }

  .submit__attachment i {
    color: var(--emh-bronze);
  }

  .submit__attachment-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .submit__attachment-remove {
    border: none;
    background: none;
    color: var(--emh-muted);
    cursor: pointer;
    font-size: 1.1rem;
    padding: 0;
    line-height: 1;
  }

  .submit__attachment-remove:hover {
    color: var(--emh-crimson);
  }

  /* Футер формы */
  .submit__footer {
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    align-items: flex-start;
  }

  .submit__note {
    margin: 0;
    color: var(--emh-muted);
    font-size: 0.85rem;
    line-height: 1.5;
  }

  /* ========================================================================= */
  /* Экран благодарности                                                        */
  /* ========================================================================= */
  .submit-success {
    max-width: 40rem;
    margin: 0 auto;
    padding: 6rem 2rem;
    text-align: center;
  }

  .submit-success__inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.2rem;
  }

  .submit-success__title {
    font-size: clamp(1.8rem, 4vw, 2.6rem);
    margin: 0;
  }

  .submit-success__lead {
    color: var(--emh-muted);
    font-size: 1.05rem;
    line-height: 1.6;
    margin: 0;
  }

  .submit-success__actions {
    display: flex;
    gap: 1rem;
    margin-top: 1rem;
    flex-wrap: wrap;
    justify-content: center;
  }
</style>