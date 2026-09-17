<script setup lang="ts">
  import { defineBreadcrumb, defineOrganization } from 'nuxt-schema-org/schema';
  import { useAnalytics } from '~/composables/useAnalytics';
  import { RIBBON_BAR_THRESHOLD, ribbonAwardsOf, sortedAwards } from '~/lib/awards';
  import { formatFlexibleDate, locationTypeLabel, relationTypeLabel, sourceTypeLabel } from "~/lib/format";
  import { toPlain } from "~/lib/pb";
  import { HeroDetailSchema } from "~/sdk/emh/v1/hero_pb";

  // Ленивая загрузка карты (не блокирует основной поток)
  const HeroLocationMap = defineAsyncComponent(() =>
    import('~/components/heroes/HeroLocationMap.vue')
  )

  const route = useRoute()
  const { hero } = useApi()
  const site = useSiteConfig()
  const id = route.params.id as string
  const { trackHeroView, trackPhotoView } = useAnalytics()

  // Рефы для автоподгрузки
  const galleryGridRef = ref<HTMLElement | null>(null);
  const sentinelRef = ref<HTMLElement | null>(null);

  // Автоподгрузка при приближении к концу галереи
  useIntersectionObserver(
    sentinelRef,
    (entries: IntersectionObserverEntry[]) => {
      const entry = entries[0];
      if (entry?.isIntersecting && hasMore.value && !loading.value) {
        loadMore();
      }
    },
    { rootMargin: '400px' }
  );

  const { data, error } = await useAsyncData(`hero-${id}`, async () => {
    const res = await hero.getHero({ id })
    if (!res.hero) {
      throw new Error('GetHero вернул пустой ответ: поле hero отсутствует')
    }

    return toPlain(HeroDetailSchema, res.hero)
  })

  // Используем composable для пагинации галереи
  const { photos, totalCount, loading, hasMore, loadMore } = useHeroPhotos({
    heroId: id,
    initialPhotos: data.value?.photos ?? [],
    pageSize: 20,
    mode: 'public',
  })

  // Трек просмотра карточки — один раз при первой загрузке данных
  watch(
    () => data.value,
    (newData) => {
      if (!newData) return;

      trackHeroView({
        heroId: id,
        heroName: fullName.value,
        rank: newData.summary?.rank || undefined,
        conflicts: conflictsLine.value || undefined,
      });
    },
    { once: true }, // ← только при первом появлении данных
  );

  const activeIndex = ref(0)
  const responsiveOptions = ref([
    { breakpoint: '1500px', numVisible: 5 },
    { breakpoint: '1024px', numVisible: 3 },
    { breakpoint: '768px', numVisible: 2 },
    { breakpoint: '560px', numVisible: 1 },
  ])
  const displayCustom = ref(false)

  const imageClick = (index: number): void => {
    activeIndex.value = index
    displayCustom.value = true  // ← UI реагирует мгновенно

    // Аналитику — в фон, когда браузер свободен
    const photo = photos.value[index]
    if (photo && import.meta.client) {
      const idle =
        window.requestIdleCallback ??
        ((cb: IdleRequestCallback) => window.setTimeout(cb, 1));

      idle(() => {
        trackPhotoView({
          heroId: id,
          photoId: photo.id || `index-${index}`,
          isMain: Boolean(photo.isMain),
        });
      });
    }
  }

  const preloadCache = new Set<string>()

  const preloadFullImage = (url: string): void => {
    if (!url || preloadCache.has(url) || !import.meta.client) return
    preloadCache.add(url)
    const img = new Image()
    img.src = url
  }

  const mainPhoto = computed(() => {
    const p = photos.value.find((x) => x.isMain) ?? photos.value[0]
    return p?.thumbnailUrl || data.value?.summary?.mainPhotoUrl || ''
  })

  const fullName = computed(() => {
    const s = data.value?.summary;
    const nickname = s?.nickname !== '' ? `«${s?.nickname}»` : '';

    return s ? [s.lastName, s.firstName, s.middleName, nickname].filter(Boolean).join(" ") : "";
  });

  const serviceYears = computed(() => {
    const start = data.value?.serviceStartDateInfo
      ? formatFlexibleDate(data.value.serviceStartDateInfo, data.value.serviceStartDate)
      : null;
    const end = data.value?.summary?.deathDateInfo
      ? formatFlexibleDate(data.value.summary.deathDateInfo, data.value.summary.deathDate)
      : null;

    return start && end ? `${start} — ${end}` : start ?? end ?? " ";
  });

  const conflictsLine = computed(() =>
    data.value?.conflicts?.map((c) => c.conflictName).join(", ") || ""
  );

  /* ====================================================================== */
  /* Орденская планка (ADR-005)                                              */
  /* ====================================================================== */

  // Награды, которые попадают в планку: лента загружена и награда носится на колодке.
  const ribbonAwards = computed(() => ribbonAwardsOf(data.value?.awards ?? []));

  // Порог переключения: планка вместо текстового списка только когда лент больше трёх.
  const showRibbonBar = computed(() => ribbonAwards.value.length > RIBBON_BAR_THRESHOLD);

  // Текстовый fallback показывает все награды героя, включая носимые без колодки
  // и награды без загруженной ленты.
  const listAwards = computed(() => sortedAwards(data.value?.awards ?? []));

  /* ====================================================================== */
  /* SEO карточки героя                                                      */
  /* ====================================================================== */

  const canonical = `${site.url}/heroes/${id}`;

  const seoDescription = computed(() => {
    const shortBio = data.value?.summary?.shortBio?.trim();
    if (shortBio) {
      return shortBio.length > 160 ? `${shortBio.slice(0, 159).trimEnd()}…` : shortBio;
    }

    const fullBio = data.value?.fullBio?.trim();
    if (fullBio) {
      return fullBio.length > 160 ? `${fullBio.slice(0, 159).trimEnd()}…` : fullBio;
    }

    return `${fullName.value}. Книга памяти «${site.name}».`;
  });

  const birthYear = computed(() => {
    const text = formatFlexibleDate(
      data.value?.summary?.birthDateInfo,
      data.value?.summary?.birthDate,
    );
    return text?.match(/\d{4}/)?.[0];
  });

  const deathYear = computed(() => {
    const text = formatFlexibleDate(
      data.value?.summary?.deathDateInfo,
      data.value?.summary?.deathDate,
    );
    return text?.match(/\d{4}/)?.[0];
  });

  const ogPhoto = computed(() => {
    const photos = data.value?.photos ?? [];
    const photo = photos.find((x) => x.isMain) ?? photos[0];
    const url = photo?.url || data.value?.summary?.mainPhotoUrl || '';
    if (!url) return undefined;
    if (/^https?:\/\//i.test(url)) return url;
    return `${site.url}${url.startsWith('/') ? '' : '/'}${url}`;
  });

  // SEO мета-теги (useSeoMeta — стандартный Nuxt API, автоимпортируется)
  useSeoMeta({
    title: () => fullName.value || 'Герой',
    ogTitle: () => fullName.value || 'Герой',
    description: () => seoDescription.value,
    ogDescription: () => seoDescription.value,
    ogType: 'profile',
    ogUrl: canonical,
    ogSiteName: site.name,
    ogLocale: 'ru_RU',
    ogImage: () => ogPhoto.value,
    robots: 'index, follow',
  });

  useHead({
    link: [
      { rel: 'canonical', href: canonical },
    ],
  });

  useSchemaOrg([
    defineOrganization({
      name: fullName.value,
      givenName: data.value?.summary?.firstName,
      familyName: data.value?.summary?.lastName,
      additionalName: data.value?.summary?.middleName || undefined,
      alternateName: data.value?.summary?.nickname
        ? `«${data.value.summary.nickname}»`
        : undefined,
      image: ogPhoto.value,
      description: seoDescription.value,
      url: canonical,
      birthDate: birthYear.value,
      deathDate: deathYear.value,
      award: data.value?.awards?.map((a) => a.awardName).filter(Boolean) || undefined,
      knowsAbout: data.value?.conflicts?.map((c) => c.conflictName).filter(Boolean) || undefined,
    }),
    defineBreadcrumb({
      itemListElement: [
        { name: 'Книга памяти', item: `${site.url}/` },
        { name: fullName.value || 'Герой', item: canonical },
      ],
    }),
  ]);

  // ═══════════════════════════════════════════════════════════
  // OG-карточка для соцсетей (nuxt-og-image)
  // ═══════════════════════════════════════════════════════════
  defineOgImage('HeroCard', {
    fullName: fullName.value,
    birthYear: birthYear.value,
    deathYear: deathYear.value,
    rank: data.value?.summary?.rank || '',
    conflicts: conflictsLine.value,
    photo: ogPhoto.value,
    nickname: data.value?.summary?.nickname || '',
  });
</script>

<template>
  <main v-if="data" class="hero">
    <!-- Шмуцтитул карточки -->
    <header class="hero__opening">
      <figure class="hero__portrait">
        <img v-if="mainPhoto" :src="mainPhoto" :alt="fullName" width="480" height="600"
          sizes="(max-width: 768px) 100vw, 340px" loading="eager" fetchpriority="high" decoding="async"
          class="hero__portrait-img" />
        <div v-else class="hero__portrait-empty">портрет не сохранился</div>
      </figure>

      <div class="hero__intro">
        <p class="hero__kicker">
          <span class="flame" aria-hidden="true"></span>
          <template v-if="data.summary?.rank">{{ data.summary?.rank }}</template>
          <template v-if="conflictsLine"> · {{ conflictsLine }}</template>
        </p>

        <h1 class="hero__name">{{ fullName }}</h1>

        <p class="hero__dates">
          {{ formatFlexibleDate(data.summary?.birthDateInfo, data.summary?.birthDate) }} —
          {{ formatFlexibleDate(data.summary?.deathDateInfo, data.summary?.deathDate) }}
        </p>

        <HeroesAwardRibbonBar v-if="showRibbonBar" :awards="ribbonAwards" />

        <ul v-else-if="listAwards.length" class="hero__awards">
          <li v-for="a in listAwards" :key="a.awardId" class="hero__award">
            <img
              v-if="a.imageUrl"
              class="hero__award-sign"
              :src="a.imageUrl"
              alt=""
              width="48"
              height="48"
              loading="lazy"
              decoding="async"
            />
            <span class="hero__award-text">
              <span class="hero__award-name">{{ a.awardName }}</span>
              <span v-if="a.decreeNumber" class="hero__award-decree">{{ a.decreeNumber }}</span>
            </span>
          </li>
        </ul>

        <p v-if="data.summary?.nickname" class="hero__nickname">
          позывной <strong>«{{ data.summary?.nickname }}»</strong>
        </p>

        <p class="hero__hint">
          Знаете больше об этом герое?
          <NuxtLink :to="{ path: '/submit', query: { hero: id, name: fullName } }">Предложите дополнение</NuxtLink>.
        </p>
      </div>
    </header>

    <!-- Биография -->
    <section v-if="data.fullBio" class="hero__section">
      <h2 class="hero__section-title">Жизнь и подвиг</h2>
      <p class="hero__bio">{{ data.fullBio }}</p>
    </section>

    <!-- Учётная карточка (военное досье) -->
    <section
      v-if="data.summary?.unit || data.position || data.summary?.serviceBranch || data.causeOfDeath || serviceYears || data.memberships?.length"
      class="hero__section">
      <h2 class="hero__section-title">Учётная карточка</h2>

      <dl class="dossier">
        <div v-if="data.summary?.serviceBranch" class="dossier__row">
          <dt>Ведомство / род войск</dt>
          <dd>{{ data.summary?.serviceBranch }}</dd>
        </div>

        <div v-if="data.summary?.unit" class="dossier__row">
          <dt>Подразделение</dt>
          <dd>{{ data.summary?.unit }}</dd>
        </div>

        <div v-if="data.position" class="dossier__row">
          <dt>Должность</dt>
          <dd>{{ data.position }}</dd>
        </div>

        <div v-if="data.summary?.rank" class="dossier__row">
          <dt>Звание</dt>
          <dd>{{ data.summary?.rank }}</dd>
        </div>

        <div v-if="serviceYears" class="dossier__row">
          <dt>Годы службы</dt>
          <dd>{{ serviceYears }}</dd>
        </div>

        <div v-if="data.causeOfDeath" class="dossier__row">
          <dt>Причина гибели</dt>
          <dd>{{ data.causeOfDeath }}</dd>
        </div>
      </dl>

      <ul v-if="data.memberships?.length" class="hero__memberships">
        <li v-for="m in data.memberships ?? []" :key="m">{{ m }}</li>
      </ul>
    </section>

    <!-- Источники -->
    <section v-if="data.sources?.length" class="hero__section">
      <h2 class="hero__section-title">Источники</h2>
      <ul class="sources">
        <li v-for="s in data.sources ?? []" :key="s.id" class="sources__item">
          <span class="sources__type">{{ sourceTypeLabel(s.sourceType) }}</span>
          <div class="sources__body">
            <a :href="s.url" target="_blank" rel="noopener" class="sources__title">{{ s.title || s.url }}</a>
            <p v-if="s.excerpt" class="sources__excerpt">{{ s.excerpt }}</p>
          </div>
        </li>
      </ul>
    </section>

    <!-- Связи с другими героями -->
    <section v-if="data.relations?.length" class="hero__section">
      <h2 class="hero__section-title">Связи</h2>
      <ul class="relations">
        <li v-for="r in data.relations ?? []" :key="r.id" class="relations__item">
          <span class="relations__type">{{ relationTypeLabel(r.relationType) }}</span>
          <NuxtLink :to="`/heroes/${r.toHeroId}`" class="relations__name">{{ r.relatedHeroName || "Герой" }}</NuxtLink>
          <span v-if="r.description" class="relations__desc">— {{ r.description }}</span>
        </li>
      </ul>
    </section>

    <!-- Галерея -->
    <section v-if="photos.length" class="hero__section">
      <h2 class="hero__section-title">Фотографии</h2>

      <Galleria v-model:activeIndex="activeIndex" v-model:visible="displayCustom" :value="photos"
        :responsiveOptions="responsiveOptions" :numVisible="7" containerStyle="max-width: 850px" :circular="true"
        :fullScreen="true" :showItemNavigators="true" :showThumbnails="false">
        <template #item="slotProps">
          <img :src="slotProps.item.url" :alt="slotProps.item?.description || fullName" width="1200" height="1500"
            sizes="(max-width: 768px) 100vw, 850px" loading="lazy" decoding="async"
            style="width: 100%; display: block" />
        </template>

        <template #thumbnail="slotProps">
          <img :src="slotProps.item.thumbnailUrl || slotProps.item.url" :alt="slotProps.item?.description || fullName"
            width="480" height="600" loading="lazy" decoding="async" style="display: block" />
          <figcaption v-if="slotProps.item?.description">{{ slotProps.item.description }}</figcaption>
        </template>
      </Galleria>

      <div ref="galleryGridRef" class="grid grid-cols-12 max-md:grid-cols-2 gap-4" style="max-width: 100%">
        <div v-for="(image, index) of photos" :key="image.id || index" class="col-span-2">
          <img class="galleria__thumbnail" :src="image.thumbnailUrl || image.url"
            :alt="image.description || `Фотография ${fullName}`" width="480" height="600" loading="lazy"
            decoding="async" style="cursor: pointer" @pointerenter="preloadFullImage(image.url!)"
            @click="imageClick(index)" />
        </div>
      </div>

      <!-- Сентинел для автоподгрузки -->
      <div v-if="hasMore" ref="sentinelRef" class="hero__gallery-sentinel" aria-hidden="true" />

      <!-- Кнопка "Загрузить ещё" -->
      <div v-if="hasMore && !loading" class="hero__gallery-load-more">
        <Button outlined :label="`Загрузить ещё (${totalCount - photos.length})`" @click="loadMore" />
      </div>

      <!-- Индикатор загрузки -->
      <div v-if="loading" class="hero__gallery-loading">
        <ProgressSpinner style="width: 30px; height: 30px" />
      </div>
    </section>

    <!-- Боевой путь -->
    <section v-if="data.conflicts?.length" class="hero__section">
      <h2 class="hero__section-title">Боевой путь</h2>
      <ul class="hero__conflicts">
        <li v-for="c in data.conflicts ?? []" :key="c.conflictId" class="hero__conflict">
          <strong class="hero__conflict-name">{{ c.conflictName }}</strong>
          <span v-if="c.specificLocation" class="hero__conflict-place">{{ c.specificLocation }}</span>
          <span v-if="c.rankAtConflict" class="hero__conflict-rank">{{ c.rankAtConflict }}</span>
        </li>
      </ul>
    </section>

    <!-- Места -->
    <section v-if="data.locations?.length" class="hero__section mb-4">
      <h2 class="hero__section-title">Места, связанные с героем</h2>
      <ul class="hero__locations">
        <li v-for="l in data.locations ?? []" :key="`${l.locationId}-${l.type}`" class="hero__location">
          <span class="hero__location-type">{{ locationTypeLabel(l.type) }}</span>
          <span class="hero__location-name">
            {{ l.location?.name }}
            <em v-if="l.location?.historicalName">({{ l.location.historicalName }})</em>
          </span>
        </li>
      </ul>
    </section>

    <!-- Карта локаций (лениво, только клиент) -->
    <ClientOnly>
      <HeroLocationMap v-if="data?.locations?.length" :locations="data.locations" />

      <template #fallback>
        <div class="hero-map-skeleton"
          style="height: 320px; border-radius: 8px; background: var(--emh-surface, #f5f5f5)" />
      </template>
    </ClientOnly>
  </main>

  <main v-else-if="error" class="hero hero--error">
    <p>Не удалось загрузить карточку героя. {{ error.message }}</p>
    <NuxtLink to="/">Вернуться к реестру</NuxtLink>
  </main>
</template>

<style scoped>
  .galleria__thumbnail {
    aspect-ratio: 1 / 1;
    object-fit: cover;
  }

  .hero {
    max-width: 80rem;
    margin: 0 auto;
    padding: 3.5rem 2rem 5rem;
  }

  .hero--error {
    padding-top: 5rem;
  }

  /* Шмуцтитул */
  .hero__opening {
    display: grid;
    grid-template-columns: minmax(240px, 340px) 1fr;
    gap: 3rem;
    align-items: start;
    margin-bottom: 3.5rem;
    content-visibility: visible;
  }

  .hero__portrait {
    margin: 0;
    aspect-ratio: 4 / 5;
    border: 1px solid var(--emh-bronze);
    padding: 6px;
    background: #fff;
    box-shadow: 0 24px 50px -30px rgba(26, 28, 32, 0.4);
  }

  .hero__portrait img {
    width: 100%;
    display: block;
    object-fit: cover;
    aspect-ratio: 3 / 4;
  }

  .hero__portrait-empty {
    aspect-ratio: 3 / 4;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 1rem;
    font-size: 0.8rem;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--emh-muted);
    background: var(--emh-bg);
  }

  .hero__kicker {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
    font-size: 0.8rem;
    letter-spacing: 0.2em;
    text-transform: uppercase;
    color: var(--emh-bronze);
    font-weight: 600;
    margin: 0 0 1.1rem;
  }

  .hero__name {
    font-size: clamp(2.2rem, 5.5vw, 3.6rem);
    line-height: 1.05;
    margin: 0 0 0.6rem;
  }

  .hero__dates {
    font-family: var(--font-display);
    font-size: 1.25rem;
    color: var(--emh-crimson);
    font-weight: 600;
    letter-spacing: 0.04em;
    margin: 0 0 1.6rem;
  }

  .hero__awards {
    list-style: none;
    margin: 0 0 1.8rem;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
  }

  .hero__award {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    border-left: 3px solid var(--emh-bronze);
    padding: 0.35rem 0 0.35rem 0.9rem;
    background: rgba(176, 141, 87, 0.07);
  }

  /* Знак награды рядом с названием; декоративен, alt="" */
  .hero__award-sign {
    flex: none;
    width: 48px;
    height: 48px;
    object-fit: contain;
  }

  .hero__award-name {
    font-weight: 600;
  }

  /* Текст имени и приказа остаётся на общей базовой линии внутри flex-строки */
  .hero__award-text {
    display: flex;
    align-items: baseline;
    gap: 0.8rem;
  }

  .hero__award-decree {
    font-size: 0.8rem;
    color: var(--emh-muted);
  }

  .hero__hint {
    font-size: 0.92rem;
    color: var(--emh-muted);
  }

  :deep(.p-galleria-mask) {
    isolation: isolate;
    contain: layout paint;
  }

  /* Секции */
  .hero__section {
    margin: auto;
    margin-top: 3rem;
    content-visibility: auto;
    contain-intrinsic-size: auto 400px;
  }

  .hero__section-title {
    font-size: 1.5rem;
    margin: 0 0 1.2rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--emh-line);
  }

  .hero__bio {
    font-size: 1.06rem;
    line-height: 1.8;
    white-space: pre-line;
    text-align: justify;
  }

  .hero__bio::first-letter {
    font-family: var(--font-display);
    font-size: 3.4em;
    font-weight: 700;
    float: left;
    line-height: 0.82;
    padding: 0.04em 0.12em 0 0;
    color: var(--emh-crimson);
  }

  .hero__nickname {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 1.15rem;
    color: var(--emh-muted);
    margin: 0 0 1.4rem;
  }

  .hero__nickname strong {
    color: var(--emh-crimson);
    font-style: normal;
  }

  /* Досье */
  .dossier {
    margin: 0;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0 2.5rem;
  }

  .dossier__row {
    display: flex;
    flex-direction: column;
    padding: .65rem 0;
    border-bottom: 1px solid var(--emh-line);
  }

  .dossier__row dt {
    font-size: .68rem;
    letter-spacing: .16em;
    text-transform: uppercase;
    color: var(--emh-bronze);
    font-weight: 600;
  }

  .dossier__row dd {
    margin: .25rem 0 0;
    font-size: 1rem;
    font-variant-numeric: tabular-nums;
  }

  .hero__memberships {
    list-style: none;
    margin: 1.2rem 0 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: .45rem;
  }

  .hero__memberships li {
    font-size: .78rem;
    padding: .28rem .7rem;
    border: 1px solid var(--emh-bronze);
    color: #7a5c2e;
    background: rgba(176, 141, 87, .08);
  }

  /* Источники */
  .sources {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: .8rem;
  }

  .sources__item {
    display: flex;
    gap: 1rem;
    align-items: flex-start;
    border-left: 3px solid var(--emh-bronze);
    padding: .55rem 0 .55rem 1rem;
    background: rgba(176, 141, 87, .05);
  }

  .sources__type {
    flex-shrink: 0;
  }

  .sources__body {
    min-width: 0;
  }

  .sources__title {
    font-weight: 600;
  }

  .sources__excerpt {
    margin: .3rem 0 0;
    font-size: .88rem;
    color: var(--emh-muted);
    font-style: italic;
  }

  /* Связи */
  .relations {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: .6rem;
  }

  .relations__item {
    display: flex;
    align-items: center;
    gap: .9rem;
    flex-wrap: wrap;
    border-left: 3px solid var(--emh-crimson);
    padding: .5rem 0 .5rem 1rem;
    background: rgba(163, 22, 33, .04);
  }

  .relations__type {
    flex-shrink: 0;
  }

  .relations__name {
    font-weight: 600;
  }

  .relations__desc {
    color: var(--emh-muted);
    font-size: .9rem;
  }

  /* Galleria — вписываем в EMH */
  .hero-galleria__main {
    width: 100%;
    max-height: 520px;
    object-fit: cover;
    display: block;
    border: 1px solid var(--emh-line);
    background: #fff;
  }

  .hero-galleria__thumb {
    width: 84px;
    height: 84px;
    object-fit: cover;
    display: block;
  }

  :deep(.hero-galleria .p-galleria-thumbnail-container) {
    gap: 0.6rem;
  }

  :deep(.hero-galleria .p-galleria-caption) {
    color: var(--emh-muted);
    font-size: 0.88rem;
    padding: 0.6rem 0.2rem 0;
    background: transparent;
  }

  /* Боевой путь и места */
  .hero__conflicts,
  .hero__locations {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.7rem;
  }

  .hero__conflict,
  .hero__location {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem 1.2rem;
    border-left: 3px solid var(--emh-crimson);
    padding: 0.5rem 0 0.5rem 1rem;
    background: rgba(163, 22, 33, 0.04);
  }

  .hero__conflict-name {
    font-weight: 700;
  }

  .hero__conflict-place {
    color: var(--emh-muted);
  }

  .hero__conflict-rank {
    font-size: 0.82rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--emh-bronze);
  }

  .hero__location-type {
    flex-shrink: 0;
  }

  .hero__location-name {
    font-weight: 600;
  }

  .hero__location-name em {
    font-weight: 400;
    color: var(--emh-muted);
  }

  @media (max-width: 640px) {
    .dossier {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 720px) {
    .hero__opening {
      grid-template-columns: 1fr;
    }

    .hero__portrait {
      max-width: 400px;
      display: grid;
      justify-self: center;
    }
  }

  .hero__gallery-load-more {
    display: flex;
    justify-content: center;
    margin-top: 1.5rem;
  }

  .hero__gallery-loading {
    display: flex;
    justify-content: center;
    margin-top: 1.5rem;
  }

  .hero__gallery-sentinel {
    height: 1px;
    width: 100%;
  }
</style>
