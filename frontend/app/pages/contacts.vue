<script setup lang="ts">
  import { useAnalytics } from '~/composables/useAnalytics';

  const site = useSiteConfig();
  const canonical = `${site.url}/contacts`;

  const contact = {
    email: 'info@neverforgotten.ru',
    codeberg: 'https://codeberg.org/Thr0TT1e/emh',
  };

  const mailtoHref = computed(() => {
    const subject = encodeURIComponent(`Вопрос по проекту «${site.name}»`);
    return `mailto:${contact.email}?subject=${subject}`;
  });

  useSeoMeta({
    title: `Контакты`,
    ogTitle: `Контакты — ${site.name}`,
    description:
      'Контакты автора проекта «Вечная память героям». Вопросы по дополнению книги памяти, исправлению сведений, передаче архивов и сотрудничеству.',
    ogDescription:
      'Контакты автора проекта «Вечная память героям». Вопросы по дополнению книги памяти, исправлению сведений, передаче архивов и сотрудничеству.',
    ogType: 'website',
    ogUrl: canonical,
    ogSiteName: site.name,
    ogLocale: 'ru_RU',
    robots: 'index, follow',
  });

  useHead({
    link: [
      { rel: 'canonical', href: canonical },
    ],
  });

  const { trackFormSubmit } = useAnalytics();

  const {
    form,
    errors,
    status,
    globalError,
    retryAfter,
    subjectOptions,
    submit,
    reset,
  } = useContactForm({
    onSuccess: (event) => {
      trackFormSubmit({
        pageUrl: event.pageUrl,
        subject: event.subject,
      });
    },
  });
</script>

<template>
  <main class="contacts">
    <!-- ================================================================== -->
    <!-- Шмуцтитул страницы контактов                                        -->
    <!-- ================================================================== -->
    <header class="contacts-hero">
      <div class="contacts-hero__inner reveal" v-reveal>
        <p class="contacts-hero__eyebrow">Связь с автором</p>

        <h1 class="contacts-hero__title">Контакты</h1>

        <p class="contacts-hero__lead">
          Проект «Вечная память героям» ведётся частным лицом. Если вы хотите
          дополнить книгу памяти, исправить сведения, передать фотографии или
          документы либо задать вопрос — воспользуйтесь электронной почтой.
        </p>
      </div>
    </header>

    <!-- ================================================================== -->
    <!-- Основные контактные карточки                                        -->
    <!-- ================================================================== -->
    <section class="contacts-main">
      <div class="contacts-main__inner">
        <div class="contacts-grid">

          <!-- 1. Основная почта -->
          <article class="contacts-card contacts-card--primary reveal" v-reveal>
            <span class="contacts-card__eyebrow">Основной способ связи</span>
            <h2 class="contacts-card__title">Электронная почта</h2>
            <p class="contacts-card__text">
              Пишите по вопросам наполнения реестра, уточнения данных о героях,
              передачи архивных материалов и сотрудничества. Телефон не используется.
            </p>

            <a class="contacts-button" :href="mailtoHref">
              <i class="pi pi-envelope" aria-hidden="true" />
              {{ contact.email }}
            </a>

            <p class="contacts-card__note">
              <i class="pi pi-clock" style="font-size: 0.85rem; margin-right: 0.3rem;" />
              Стараюсь отвечать в кратчайшие сроки.
            </p>
          </article>

          <!-- 2. Репозиторий (Codeberg) -->
          <article class="contacts-card reveal" v-reveal style="transition-delay: 60ms">
            <span class="contacts-card__eyebrow">Открытый исходный код</span>
            <h2 class="contacts-card__title">Репозиторий проекта</h2>
            <p class="contacts-card__text">
              Проект разрабатывается открыто. Исходный код, архитектура,
              протоколы и история изменений доступны на Codeberg.
            </p>

            <a class="contacts-button" :href="contact.codeberg" target="_blank" rel="noopener noreferrer">
              <i class="pi pi-code" aria-hidden="true" />
              codeberg.org/Thr0TT1e/emh
            </a>
          </article>

          <!-- 3. Что писать в обращении -->
          <article class="contacts-card reveal" v-reveal style="transition-delay: 120ms">
            <span class="contacts-card__eyebrow">Что указать в обращении</span>

            <h2 class="contacts-card__title">Так будет быстрее</h2>

            <ul class="contacts-list">
              <li>
                Фамилию, имя и отчество героя, если речь о записи в книге памяти;
              </li>
              <li>
                Конфликт, подразделение, звание или известные даты;
              </li>
              <li>
                Ссылку на страницу героя на сайте, если нужно исправить данные;
              </li>
              <li>
                Ссылки на источники, фотографии или документы.
              </li>
            </ul>
          </article>

          <!-- 4. Дисклеймер (Частный проект) -->
          <article class="contacts-card contacts-card--muted reveal" v-reveal style="transition-delay: 180ms">
            <span class="contacts-card__eyebrow">Важно</span>
            <h2 class="contacts-card__title">Частный мемориальный проект</h2>
            <p class="contacts-card__text">
              Сайт не является официальной организацией, государственным архивом
              или СМИ. Данные собираются из открытых источников и семейных
              материалов.
            </p>
            <p class="contacts-card__text" style="margin-top: 0.5rem;">
              Если вы представляете ведомство или фонд и хотите передать
              данные для реестра — обязательно укажите это в теме письма.
            </p>
          </article>

        </div>
      </div>
    </section>

    <!-- ================================================================== -->
    <!-- Форма обратной связи                                                -->
    <!-- ================================================================== -->
    <section class="contact-form-section">
      <div class="contact-form-section__inner">
        <div class="contact-form-head reveal" v-reveal>
          <p class="contacts-card__eyebrow">Обратная связь</p>
          <h2 class="contact-form-head__title">Написать сообщение</h2>
        </div>
        <p class="contact-form-head__text">
          Форма предназначена для сотрудничества, предложений по проекту и
          общих вопросов. Для дополнения книги памяти используйте
          <NuxtLink to="/submit">форму «Сообщить о герое»</NuxtLink>.
        </p>

        <!-- ClientOnly: убираем hydration mismatch от PrimeVue -->
        <ClientOnly>
          <article class="contact-form-card reveal is-revealed">
            <div v-if="status === 'success'" class="contact-form-success">
              <Message severity="success" :closable="false">
                Сообщение отправлено. Спасибо за обращение.
              </Message>
              <p class="contact-form-success__text">
                Если ваш вопрос требует ответа, он будет направлен на указанный
                вами email. Телефон не используется.
              </p>
              <Button label="Отправить ещё одно сообщение" icon="pi pi-refresh" outlined @click="reset" />
            </div>
            <form v-else novalidate @submit.prevent="submit">
              <!-- Honeypot: скрытое поле от простых ботов -->
              <div class="contact-form__hp" aria-hidden="true">
                <label>
                  Не заполняйте это поле
                  <input v-model="form.honeypot" type="text" tabindex="-1" autocomplete="off" maxlength="200" />
                </label>
              </div>
              <div class="contact-form__grid">
                <label class="contact-form__field">
                  <span class="contact-form__label">Имя *</span>
                  <InputText v-model="form.name" placeholder="Как к вам обращаться" class="w-full" autocomplete="name"
                    maxlength="200" :invalid="Boolean(errors.name)" />
                  <small v-if="errors.name" class="contact-form__error">
                    {{ errors.name }}
                  </small>
                </label>
                <label class="contact-form__field">
                  <span class="contact-form__label">Email *</span>
                  <InputText v-model="form.email" type="email" placeholder="you@example.ru" class="w-full"
                    autocomplete="email" maxlength="254" :invalid="Boolean(errors.email)" />
                  <small v-if="errors.email" class="contact-form__error">
                    {{ errors.email }}
                  </small>
                </label>
                <label class="contact-form__field contact-form__field--full">
                  <span class="contact-form__label">Тема обращения</span>
                  <Select v-model="form.subject" :options="subjectOptions" option-label="label" option-value="value"
                    placeholder="Выберите тему" class="w-full" />
                </label>
                <label class="contact-form__field contact-form__field--full">
                  <span class="contact-form__label">Сообщение *</span>
                  <Textarea v-model="form.message" rows="6" auto-resize
                    placeholder="Опишите вопрос, предложение или тему сотрудничества" class="w-full" maxlength="5000"
                    :invalid="Boolean(errors.message)" />
                  <div class="contact-form__meta">
                    <small v-if="errors.message" class="contact-form__error">
                      {{ errors.message }}
                    </small>
                    <small class="contact-form__counter">
                      {{ form.message.length }} / 5000
                    </small>
                  </div>
                </label>
              </div>
              <div class="contact-form__consent">
                <Checkbox inputId="contact-consent" v-model="form.consent" :binary="true"
                  :invalid="Boolean(errors.consent)" />
                <label for="contact-consent">
                  Согласен на обработку персональных данных: имя, email и текст
                  сообщения — для целей обратной связи. *
                </label>
              </div>
              <small v-if="errors.consent" class="contact-form__error">
                {{ errors.consent }}
              </small>
              <Message v-if="globalError" severity="error" :closable="false" class="contact-form__global-error">
                {{ globalError }}
              </Message>
              <div class="contact-form__actions">
                <Button type="submit" label="Отправить сообщение" icon="pi pi-send" :loading="status === 'submitting'"
                  :disabled="status === 'submitting' || retryAfter > 0" />
                <p class="contact-form__note">
                  Ответ придёт на email. Телефон не используется.
                </p>
              </div>
            </form>
          </article>

          <!-- Fallback на время загрузки (SSR) -->
          <template #fallback>
            <article class="contact-form-card reveal is-revealed">
              <div class="contact-form__grid">
                <div class="contact-form__field">
                  <span class="contact-form__label">Имя *</span>
                  <div class="contact-form__skeleton" />
                </div>
                <div class="contact-form__field">
                  <span class="contact-form__label">Email *</span>
                  <div class="contact-form__skeleton" />
                </div>
                <div class="contact-form__field contact-form__field--full">
                  <span class="contact-form__label">Сообщение *</span>
                  <div class="contact-form__skeleton contact-form__skeleton--tall" />
                </div>
              </div>
            </article>
          </template>
        </ClientOnly>
      </div>
    </section>
  </main>
</template>

<style scoped>
  .contacts-hero {
    border-bottom: 1px solid var(--emh-line);
    background: linear-gradient(180deg,
        rgba(255, 255, 255, 0) 0%,
        rgba(163, 22, 33, 0.045) 100%);
  }

  .contacts-hero__inner,
  .contacts-main__inner {
    max-width: 72rem;
    margin: 0 auto;
    padding-inline: 2rem;
  }

  .contacts-hero__inner {
    max-width: 58rem;
    padding-block: 3rem 3.5rem;
    text-align: center;
  }

  .contacts-hero__eyebrow,
  .contacts-card__eyebrow {
    margin: 0;
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: var(--emh-bronze);
  }

  .contacts-hero__title {
    margin: 0.4rem 0 1rem;
    font-size: clamp(2.4rem, 5vw, 4rem);
  }

  .contacts-hero__lead {
    margin: 0 auto;
    max-width: 46rem;
    color: var(--emh-muted);
    font-size: 1.05rem;
  }

  .contacts-main {
    padding: 3rem 0 5rem;
  }

  .contacts-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1.4rem;
  }

  .contacts-card {
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    min-height: 100%;
    padding: 1.6rem;
    border: 1px solid var(--emh-line);
    border-radius: 8px;
    background: var(--emh-surface);
    box-shadow: 0 0 0 rgba(26, 28, 32, 0);
    transition:
      border-color 0.25s ease,
      box-shadow 0.25s ease,
      transform 0.25s ease;
  }

  .contacts-card:hover {
    border-color: var(--emh-bronze);
    box-shadow: 0 14px 30px -18px rgba(26, 28, 32, 0.24);
    transform: translateY(-2px);
  }

  .contacts-card--primary {
    background:
      radial-gradient(28rem 16rem at 100% 0%,
        rgba(176, 141, 87, 0.12),
        transparent 55%),
      var(--emh-surface);
  }

  .contacts-card--muted {
    background: rgba(242, 243, 245, 0.65);
  }

  .contacts-card__title {
    margin: 0;
    font-size: 1.45rem;
  }

  .contacts-card__text {
    margin: 0;
    color: var(--emh-muted);
    line-height: 1.6;
  }

  .contacts-button {
    display: inline-flex;
    align-items: center;
    gap: 0.65rem;
    width: fit-content;
    margin-top: 0.4rem;
    padding: 0.75rem 1rem;
    border: 1px solid var(--emh-line);
    border-radius: 6px;
    background: var(--emh-bg);
    color: var(--emh-ink);
    font-weight: 600;
    word-break: break-all;
    transition: all 0.2s ease;
  }

  .contacts-button:hover {
    border-color: var(--emh-crimson);
    color: var(--emh-crimson);
  }

  .contacts-list {
    display: grid;
    gap: 0.65rem;
    margin: 0;
    padding-left: 1.1rem;
    color: var(--emh-muted);
    line-height: 1.5;
  }

  .contacts-card__note {
    margin: 0;
    margin-top: auto;
    padding-top: 0.5rem;
    color: var(--emh-muted);
    font-size: 0.92rem;
    display: flex;
    align-items: center;
  }

  /* Адаптив */
  @media (max-width: 900px) {
    .contacts-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 640px) {

    .contacts-hero__inner,
    .contacts-main__inner {
      padding-inline: 1rem;
    }

    .contacts-hero__inner {
      padding-block: 2.2rem 2.6rem;
    }

    .contacts-main {
      padding: 2rem 0 4rem;
    }

    .contacts-card {
      padding: 1.2rem;
    }
  }

  /* ========================================================================= */
  /* Форма обратной связи                                                       */
  /* ========================================================================= */
  .contact-form-section {
    padding: 0 0 5rem;
  }

  .contact-form-section__inner {
    max-width: 72rem;
    margin: 0 auto;
    padding-inline: 2rem;
  }

  .contact-form-head {
    max-width: 52rem;
    margin-bottom: 1.6rem;
  }

  .contact-form-head__title {
    margin: 0.4rem 0 0.8rem;
    font-size: clamp(1.8rem, 3.4vw, 2.5rem);
  }

  .contact-form-head__text {
    margin-bottom: 1rem;
    margin-left: 1rem;
    color: var(--emh-muted);
  }

  .contact-form-card {
    border: 1px solid var(--emh-line);
    border-radius: 8px;
    background: var(--emh-surface);
    padding: 1.8rem;
    box-shadow: 0 0 0 rgba(26, 28, 32, 0);
  }

  .contact-form__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1.2rem;
  }

  .contact-form__field {
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
  }

  .contact-form__field--full {
    grid-column: 1 / -1;
  }

  .contact-form__label {
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: var(--emh-bronze);
  }

  .contact-form__error {
    display: block;
    color: var(--emh-crimson);
    font-size: 0.85rem;
  }

  .contact-form__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .contact-form__counter {
    margin-left: auto;
    color: var(--emh-muted);
    font-size: 0.85rem;
    font-variant-numeric: tabular-nums;
  }

  .contact-form__consent {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    margin-top: 1.2rem;
    color: var(--emh-muted);
    font-size: 0.92rem;
  }

  .contact-form__consent label {
    cursor: pointer;
  }

  .contact-form__actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 1rem;
    margin-top: 1.4rem;
  }

  .contact-form__note {
    margin: 0;
    color: var(--emh-muted);
    font-size: 0.9rem;
  }

  .contact-form__global-error {
    margin-top: 1.2rem;
  }

  .contact-form-success {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .contact-form-success__text {
    margin: 0;
    color: var(--emh-muted);
  }

  /* Honeypot: визуально скрыто, но доступно для простых ботов */
  .contact-form__hp {
    position: absolute;
    width: 1px;
    height: 1px;
    margin: -1px;
    padding: 0;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
    border: 0;
  }

  @media (max-width: 760px) {
    .contact-form-section {
      padding-bottom: 4rem;
    }

    .contact-form-section__inner {
      padding-inline: 1rem;
    }

    .contact-form-card {
      padding: 1.2rem;
    }

    .contact-form__grid {
      grid-template-columns: 1fr;
    }

    .contact-form__actions {
      flex-direction: column;
      align-items: stretch;
    }

    .contact-form__note {
      text-align: center;
    }
  }

  .contact-form__skeleton {
    height: 2.5rem;
    border-radius: 6px;
    background: linear-gradient(90deg,
        var(--emh-line) 0%,
        rgba(217, 220, 225, 0.4) 50%,
        var(--emh-line) 100%);
    background-size: 200% 100%;
    animation: skeleton-pulse 1.4s ease-in-out infinite;
  }

  .contact-form__skeleton--tall {
    height: 8rem;
  }

  @keyframes skeleton-pulse {
    0% {
      background-position: 200% 0;
    }

    100% {
      background-position: -200% 0;
    }
  }
</style>