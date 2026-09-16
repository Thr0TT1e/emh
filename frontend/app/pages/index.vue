<script setup lang="ts">
  import { useAnalytics } from '~/composables/useAnalytics';
  import { formatFlexibleYear } from "~/lib/format";
  import { toPlain } from "~/lib/pb";
  import { HeroSummarySchema, type HeroSummaryJson } from "~/sdk/emh/v1/hero_pb";
  import AccountTieHatOutlineIcon from '~icons/mdi/account-tie-hat-outline?width=1.25em&height=1.25em';
  import ChevronDownIcon from '~icons/mdi/chevron-down?width=1.25em&height=1.25em';

  const { trackSearch } = useAnalytics();
  const { hero } = useApi();
  const route = useRoute();

  const search = ref("");
  const conflictId = ref("");

  const { data: conflicts } = await useConflictsCatalog();
  const { data: heroesData } = await useHeroesPage(route.path === "/");

  const heroes = ref<HeroSummaryJson[]>(heroesData.value?.heroes ?? []);
  const cursor = ref(heroesData.value?.nextCursor ?? "");
  const total = ref(heroesData.value?.total ?? 0);
  const loading = ref(false);

  const fetchPage = async (append: boolean) => {
    loading.value = true;

    try {
      const res = await hero.listHeroes({
        pagination: { pageSize: 20, cursor: append ? cursor.value : "" },
        searchQuery: search.value,
        conflictId: conflictId.value,
      });

      const page = res.heroes.map((h) => toPlain(HeroSummarySchema, h));
      heroes.value = append ? [...heroes.value, ...page] : page;
      cursor.value = res.pagination?.nextCursor ?? "";

      if (!append) total.value = Number(res.pagination?.totalCount ?? 0n);
    } finally {
      loading.value = false;
    }
  };

  let searchTimer: ReturnType<typeof setTimeout> | undefined;

  watch(search, (newQuery) => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      const query = newQuery.trim();
      if (query || conflictId.value) {
        trackSearch({
          query,
          conflictId: conflictId.value || undefined,
          resultsCount: 0,
        });
      }
      fetchPage(false);
    }, 400);
  });

  // Аналогично для фильтра по конфликту
  watch(conflictId, (newConflictId) => {
    trackSearch({
      query: search.value.trim(),
      conflictId: newConflictId || undefined,
      resultsCount: 0,
    });
    fetchPage(false);
  });

  const hasMore = computed(() => cursor.value !== "");

  function fullName(hero: HeroSummaryJson) {
    const nickname = hero?.nickname !== '' ? `«${hero?.nickname}»` : '';
    return hero ? [hero.lastName, hero.firstName, hero.middleName, nickname].filter(Boolean).join(" ") : "";
  }

  // ═══════════════════════════════════════════════════════════
  // SEO главной страницы (nuxt-seo-utils + nuxt-og-image)
  // ═══════════════════════════════════════════════════════════

  const site = useSiteConfig();
  const canonical = `${site.url}/`;

  useSeoMeta({
    title: 'Книга памяти',
    ogTitle: 'Книга памяти',
    description:
      'Книга памяти о героях, погибших в глобальных и локальных конфликтах: ВОВ, Афган, Чечня, Вьетнам, Сирия, Африка, Новороссия, спецоперации.',
    ogDescription:
      'Книга памяти о героях, погибших в глобальных и локальных конфликтах: ВОВ, Афган, Чечня, Вьетнам, Сирия, Африка, Новороссия, спецоперации.',
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

  // OG-карточка для соцсетей (nuxt-og-image)
  defineOgImage('HomeCard', {
    totalHeroes: total.value,
    totalConflicts: conflicts.value?.length ?? 0,
  });
</script>

<template>
  <main>
    <!-- ================================================================== -->
    <!-- Masthead (шмуцтитул)                                                 -->
    <!-- ================================================================== -->
    <header class="masthead">
      <div class="masthead__inner">
        <div class="masthead__logo">
          <Image src="/logo_v5_full_vert_wbr.svg" alt="Вечная память героям" />
        </div>

        <p class="masthead__slogan">
          География разная. Подвиг один.
          <span class="masthead__slogan-accent">Память вечна.</span>
        </p>

        <p class="masthead__lead">
          Мы помним каждого, кто не вернулся с поля боя, независимо от географии
          и эпохи: ВОВ, Афган, Чечня, Вьетнам, Сирия, Африка, Новороссия,
          спецоперации…<br>
          Каждая запись — судьба, собранная по крупицам из
          открытых источников.
        </p>
      </div>
    </header>

    <!-- ================================================================== -->
    <!-- Поиск и фильтр (PrimeVue)                                            -->
    <!-- ================================================================== -->
    <section class="search">
      <div class="search__inner">
        <div class="search__field">
          <label for="search-bio" class="search__label">Поиск по фамилии, имени, отчеству</label>
          <InputText id="search-bio" v-model="search" placeholder="Иванов Иван Иванович" size="large" class="w-full"
            autocomplete="off" />
        </div>

        <div class="search__field">
          <span class="search__label">Конфликт</span>
          <Select v-model="conflictId" :options="conflicts ?? []" option-label="name" option-value="id" show-clear
            placeholder="Все конфликты" class="w-full" />
        </div>
      </div>
    </section>

    <!-- ================================================================== -->
    <!-- Реестр имён — резиновый грид с карточками                            -->
    <!-- ================================================================== -->
    <section class="registry">
      <div class="registry__inner">
        <div class="registry__head">
          <h1 class="registry__title">Реестр имён</h1>
          <p class="registry__count">Найдено: {{ total }}</p>
        </div>

        <div v-if="heroes.length" class="heroes-grid">
          <NuxtLink v-for="(h, i) in heroes" :key="h.id" :to="`/heroes/${h.id}`" class="hero-card reveal" v-reveal
            :style="{ transitionDelay: `${(i % 12) * 30}ms` }">
            <div class="hero-card__photo">
              <img v-if="h.mainThumbnailUrl" :src="h.mainThumbnailUrl" :alt="fullName(h)" width="480" height="600"
                sizes="(max-width: 768px) 50vw, (max-width: 1024px) 33vw, 25vw" loading="lazy" decoding="async"
                class="hero-card__photo-img" />
              <span v-else class="hero-card__photo-empty">
                <AccountTieHatOutlineIcon />
              </span>
            </div>

            <div class="hero-card__body">
              <div class="hero-card__name">
                {{ h.lastName }} {{ h.firstName }}
                <span v-if="h.middleName">{{ h.middleName }}</span>
              </div>

              <div v-if="h.nickname" class="hero-card__callsign">
                «{{ h.nickname }}»
              </div>

              <div v-if="h.rank" class="hero-card__rank">{{ h.rank }}</div>

              <div class="hero-card__dates">
                <span class="num">{{ formatFlexibleYear(h.birthDateInfo, h.birthDate) }}</span>
                <span class="hero-card__dates-sep">—</span>
                <span class="num">{{ formatFlexibleYear(h.deathDateInfo, h.deathDate) }}</span>
              </div>
            </div>
          </NuxtLink>
        </div>

        <div v-if="loading" class="registry__loading">
          <ProgressSpinner style="width: 40px; height: 40px" />
          <span>Загружаем записи…</span>
        </div>

        <p v-else-if="!heroes.length" class="registry__status">
          Никого не нашли. Измените запрос или уточните написание.
        </p>

        <div v-if="hasMore && !loading" class="registry__more-wrap">
          <Button outlined @click="fetchPage(true)">
            <ChevronDownIcon />
            <span>Показать ещё</span>
          </Button>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>

  /* ========================================================================= */
  /* Masthead                                                                   */
  /* ========================================================================= */
  .masthead {
    border-bottom: 1px solid var(--emh-line);
    background: linear-gradient(180deg,
        rgba(255, 255, 255, 0) 0%,
        rgba(189, 47, 60, 0.04) 100%);
  }

  .masthead__inner {
    max-width: 72rem;
    margin: 0 auto;
    padding: 2.5rem 2rem 3.5rem;
    text-align: center;
  }

  .masthead__logo {
    width: 300px;
    /* width: 120px; */
    /* height: 150px; */
    margin: 0 auto 2rem;
  }

  .masthead__logo img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  .masthead__title {
    font-size: clamp(2.8rem, 7.5vw, 5.2rem);
    line-height: 1.02;
    margin: 0 0 1.2rem;
    letter-spacing: -0.01em;
  }

  .masthead__title em {
    font-style: italic;
    color: var(--emh-crimson);
  }

  .masthead__slogan {
    font-size: clamp(1.15rem, 2.2vw, 1.6rem);
    font-weight: 500;
    letter-spacing: 0.02em;
    margin: 0 auto 1.8rem;
    max-width: 48rem;
    color: var(--emh-ink);
    line-height: 1.4;
  }

  .masthead__slogan-accent {
    color: var(--emh-crimson);
    font-weight: 700;
  }

  .masthead__lead {
    max-width: 44rem;
    margin: 0 auto;
    font-size: 1.05rem;
    color: var(--emh-muted);
  }

  .masthead__cta {
    margin-top: 1.8rem;
  }

  .masthead__stats {
    display: flex;
    justify-content: center;
    gap: 4rem;
    margin: 3rem 0 0;
  }

  .masthead__stat dd {
    font-family: var(--font-display);
    font-size: 2.6rem;
    font-weight: 700;
    color: var(--emh-ink);
    margin: 0;
  }

  .masthead__stat dt {
    font-size: 0.78rem;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: var(--emh-bronze);
    margin: 0.2rem 0 0;
  }

  /* ========================================================================= */
  /* Поиск                                                                      */
  /* ========================================================================= */
  .search {
    border-bottom: 1px solid var(--emh-line);
    background: rgba(255, 255, 255, 0.55);
  }

  .search__inner {
    /* Резиновость: широкая секция на больших экранах */
    max-width: 96rem;
    margin: 0 auto;
    padding: 1.8rem 2rem;
    display: flex;
    gap: 1.5rem;
    flex-wrap: wrap;
    align-items: end;
  }

  .search__field {
    flex: 1 1 320px;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .search__label {
    font-size: 0.72rem;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    color: var(--emh-bronze);
    font-weight: 600;
  }

  /* ========================================================================= */
  /* Реестр — резиновый контейнер                                                */
  /* ========================================================================= */
  .registry__inner {
    max-width: 96rem;
    /* 1536px — практически вся ширина 2K-монитора */
    margin: 0 auto;
    padding: 3rem 2rem 5rem;
  }

  .registry__head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 1.5rem;
    border-bottom: 1px solid var(--emh-line);
    padding-bottom: 0.8rem;
  }

  .registry__title {
    font-size: 1.7rem;
    margin: 0;
  }

  .registry__count {
    margin: 0;
    color: var(--emh-muted);
    font-size: 0.92rem;
    font-variant-numeric: tabular-nums;
  }

  /* ========================================================================= */
  /* Грид карточек — адаптивный                                                  */
  /* ========================================================================= */
  .heroes-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 1.4rem;
  }

  @media (min-width: 640px) {
    .heroes-grid {
      grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
      gap: 1.6rem;
    }
  }

  @media (min-width: 1280px) {
    .heroes-grid {
      grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
      gap: 1.8rem;
    }
  }

  /* ========================================================================= */
  /* Карточка героя                                                              */
  /* ========================================================================= */
  .hero-card {
    display: flex;
    flex-direction: column;
    text-decoration: none;
    color: inherit;
    background: var(--p-surface-0);
    border: 1px solid var(--emh-line, var(--p-surface-200));
    border-radius: 6px;
    overflow: hidden;
    transition:
      transform 0.3s ease,
      box-shadow 0.3s ease,
      border-color 0.3s ease;
  }

  .hero-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 14px 32px -14px rgba(26, 28, 32, 0.25);
    border-color: var(--emh-bronze);
  }

  .hero-card__photo {
    aspect-ratio: 4 / 5;
    overflow: hidden;
    background: var(--p-surface-100);
    position: relative;
    border-bottom: 1px solid var(--emh-line, var(--p-surface-200));
  }

  .hero-card__photo-img {
    filter: grayscale(0.7);
    transition: filter 0.4s ease, transform 0.4s ease;
  }

  .hero-card:hover .hero-card__photo-img {
    filter: grayscale(0);
    transform: scale(1.03);
  }

  .hero-card__photo-empty {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--p-text-muted-color);
    font-size: 3.2rem;
    opacity: 0.5;
  }

  .hero-card__body {
    padding: 0.9rem 1rem 1.1rem;
    display: flex;
    flex-direction: column;
    flex: 1;
  }

  .hero-card__name {
    font-weight: 600;
    font-size: 0.98rem;
    line-height: 1.3;
    color: var(--emh-ink, var(--p-text-color));
    margin-bottom: 0.25rem;
  }

  .hero-card__callsign {
    font-size: 0.82rem;
    color: var(--emh-bronze);
    font-style: italic;
    margin-bottom: 0.3rem;
  }

  .hero-card__rank {
    font-size: 0.78rem;
    color: var(--p-text-muted-color);
    letter-spacing: 0.02em;
    text-transform: lowercase;
    margin-bottom: 0.5rem;
  }

  .hero-card__dates {
    margin-top: auto;
    font-size: 0.88rem;
    color: var(--p-text-muted-color);
    padding-top: 0.6rem;
    border-top: 1px dashed var(--emh-line, var(--p-surface-200));
  }

  .hero-card__dates-sep {
    margin: 0 0.35rem;
    color: var(--emh-bronze);
  }

  /* ========================================================================= */
  /* Состояния реестра                                                           */
  /* ========================================================================= */
  .registry__loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 3rem 0;
    color: var(--emh-muted);
  }

  .registry__status {
    padding: 3rem 0;
    text-align: center;
    color: var(--emh-muted);
  }

  .registry__more-wrap {
    text-align: center;
    margin-top: 2.5rem;
  }

  /* ========================================================================= */
  /* Мобильные правки                                                            */
  /* ========================================================================= */
  @media (max-width: 640px) {
    .masthead__stats {
      flex-direction: column;
      gap: 1.4rem;
    }

    .search__inner {
      padding: 1.2rem 1rem;
    }

    .registry__inner {
      padding: 2rem 1rem 4rem;
    }
  }
</style>