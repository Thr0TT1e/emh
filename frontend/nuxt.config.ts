// https://nuxt.com/docs/api/configuration/nuxt-config
import type { SitemapUrlInput } from '@nuxtjs/sitemap';
import tailwindcss from '@tailwindcss/vite';
// import { FileSystemHMRIconLoader } from 'unplugin-icons/loaders';
// import IconsResolver from 'unplugin-icons/resolver';
// import ViteComponents from 'unplugin-vue-components/vite';

export default defineNuxtConfig({
  compatibilityDate: '2026-07-01',
  devtools: { enabled: process.env.NODE_ENV === 'development' },

  modules: [
    '@primevue/nuxt-module',
    '@nuxt/hints',
    'nuxt-seo-utils',
    '@nuxtjs/sitemap',
    '@nuxtjs/robots',
    'nuxt-schema-org',
    'nuxt-og-image',
    '@nuxt/scripts',
    // Локальные шрифты
    '@nuxt/fonts',
    // HTML Validator
    '@nuxtjs/html-validator',
    // VueUse composables
    '@vueuse/nuxt',
    // Dark/Light mode
    '@nuxtjs/color-mode',
    [
      'unplugin-icons/nuxt',
      {
        // customCollections: {
        //   ...FileSystemHMRIconLoader('app/custom-a', 'custom'),
        // },
      },
    ],
  ],

  css: ['~/assets/css/main.css'],

  primevue: {
    importTheme: { from: '@/themes/mytheme.ts' },
    options: {
      ripple: true,
      // theme: { options: { darkModeSelector: 'system' } },
      // Синхронизация с @nuxtjs/color-mode
      // Теперь PrimeVue будет реагировать на класс .dark-mode
      theme: { options: { darkModeSelector: '.dark-mode' } },
    },
    autoImport: true,
  },

  vite: {
    plugins: [
      tailwindcss(),
      // ViteComponents({
      //   resolvers: [
      //     IconsResolver({
      //       prefix: '',
      //       strict: true,
      //       customCollections: ['custom'],
      //     }),
      //   ],
      //   dts: true,
      // }),
    ],
    // LightningCSS для ускорения сборки
    css: {
      transformer: 'lightningcss',
    },
    build: {
      target: 'es2024', // или 'esnext'
    },
    optimizeDeps: {
      include: [
        '@bufbuild/protobuf',
        '@connectrpc/connect-web',
        '@connectrpc/connect',
        '@unhead/schema-org/vue',
        '@bufbuild/protobuf/codegenv2',
        '@bufbuild/protobuf/wkt',
        'vue-advanced-cropper',
        'ol/Collection',
        'ol/Feature',
        'ol/Map',
        'ol/View',
        'ol/geom/Point',
        'ol/interaction/Translate',
        'ol/layer/Tile',
        'ol/layer/Vector',
        'ol/proj',
        'ol/source/OSM',
        'ol/source/Vector',
        'ol/style/Circle',
        'ol/style/Fill',
        'ol/style/Stroke',
        'ol/style/Style',
      ],
    },
  },

  // components: {
  //   dirs: [
  //     {
  //       path: '~/components',
  //       pathPrefix: false,
  //     },
  //   ],
  // },

  // Локальные шрифты (конфигурация)
  fonts: {
    provider: 'local',
    families: [
      {
        name: 'Playfair Display',
        weights: [500, 600, 700, 800],
        styles: ['normal', 'italic'],
        subsets: ['cyrillic', 'latin'],
      },
      {
        name: 'Golos Text',
        weights: [400, 500, 600, 700],
        subsets: ['cyrillic', 'latin'],
      },
    ],
  },

  // HTML Validator (только в dev и CI)
  htmlValidator: {
    // Валидируем только в dev и CI
    enabled: process.env.NODE_ENV !== 'production' || process.env.VALIDATE_HTML === 'true',
    failOnError: process.env.CI === 'true',

    // Ограничиваем валидацию только публичными страницами
    // Админка — внутренний инструмент, строгая HTML-валидность там не нужна
    ignore: [
      /^\/admin/, // Исключаем всю админку
    ],

    options: {
      rules: {
        'meta-refresh': 'off',

        // PrimeVue рендерит <BUTTON> в верхнем регистре при SSR
        'element-case': 'off',

        // PrimeVue Checkbox содержит <div> внутри <label> — это их внутренняя структура
        // NuxtLink + Button = <a><button> (распространённый паттерн)
        'element-permitted-content': 'off',

        // DataTable добавляет role="table/row/cell" явно для screen readers.
        // Это избыточно по HTML5 spec, но улучшает a11y — оставляем как есть.
        'no-redundant-role': 'off',

        // DataTable не добавляет scope к <th>, но это не критично
        // для таблиц с одной строкой заголовков (а у нас именно такие)
        'wcag/h63': 'off',

        // PrimeVue динамически создаёт ID на клиенте (pv_id_*)
        'no-missing-references': 'off',
      },
    },
  },

  // Color Mode (автоопределение темы)
  colorMode: {
    preference: 'system',
    fallback: 'light',
    classSuffix: '-mode', // Генерирует классы: light-mode, dark-mode
  },

  // ═══════════════════════════════════════════════════════════
  // Runtime config — только для API, больше не для SEO
  // ═══════════════════════════════════════════════════════════
  runtimeConfig: {
    // Только для сервера: внутренний адрес бэкенда при SSR (в контейнере).
    apiInternalBaseUrl: process.env.NUXT_API_INTERNAL_BASE_URL || '',
    public: {
      // Адрес бэкенда, доступный из браузера.
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL ?? 'https://api.neverforgotten.ru',
    },
  },

  // ═══════════════════════════════════════════════════════════
  // Глобальные SEO-константы (nuxt-seo-utils)
  // Читается всеми модулями: sitemap, schema-org, og-image, robots
  // Env: NUXT_SITE_URL, NUXT_SITE_NAME
  // ═══════════════════════════════════════════════════════════
  site: {
    url: 'https://вежливые.рус',
    name: 'Вечная память героям',
    description:
      'Книга памяти о героях, погибших в глобальных и локальных конфликтах: ВОВ, Афган, Чечня, Вьетнам, Сирия, Африка, Новороссия, спецоперации.',
    defaultLocale: 'ru',
    identity: {
      type: 'Organization',
      name: 'Thr0TT1e',
      logo: '/logo_v5_full_vert_wbr.svg',
      sameAs: ['https://codeberg.org/Thr0TT1e/emh'],
    },
  },

  // ═══════════════════════════════════════════════════════════
  // Глобальные мета-теги (nuxt-seo-utils)
  // Заменяет старый app.head.meta
  // ═══════════════════════════════════════════════════════════
  app: {
    head: {
      htmlAttrs: { lang: 'ru' },
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/logo_favicon.svg' },
        // OpenSearch для браузерного поиска
        {
          rel: 'search',
          type: 'application/opensearchdescription+xml',
          title: 'Поиск героев',
          href: '/opensearch.xml',
        },
      ],
    },
    // @ts-expect-error
    seoMeta: {
      // Глобальный title-template: подставляется %s из каждой страницы
      titleTemplate: '%s — Вечная память героям',
      // Fallback, если страница не задала description
      description: 'Книга памяти о героях, погибших во время глобальных и локальных конфликтов.',
      // Open Graph
      ogType: 'website',
      ogLocale: 'ru_RU',
      ogSiteName: 'Вечная память героям',
      // Robots по умолчанию (можно переопределить на странице)
      robots: 'index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1',
    },
  },

  // ═══════════════════════════════════════════════════════════
  // OG-карточки (nuxt-og-image)
  // ═══════════════════════════════════════════════════════════
  ogImage: {
    defaults: {
      width: 1200,
      height: 630,
    },
    // Защита от DoS-атак на генерацию OG-картинок
    security: {
      strict: !!process.env.NUXT_OG_IMAGE_SECRET,
      secret: process.env.NUXT_OG_IMAGE_SECRET,
    },
  },

  // ═══════════════════════════════════════════════════════════
  // Robots.txt (@nuxtjs/robots)
  // ═══════════════════════════════════════════════════════════
  robots: {
    groups: [
      // Поисковики и обычные пользователи — доступ разрешён
      {
        userAgent: '*',
        allow: '/',
        disallow: ['/admin', '/admin/'],
      },
      // AI-скрейперы и боты для обучения моделей — полный запрет
      {
        userAgent: [
          'GPTBot',
          'OAI-SearchBot',
          'ChatGPT-User',
          'Google-Extended',
          'Google-CloudVertexBot',
          'anthropic-ai',
          'ClaudeBot',
          'Claude-Web',
          'cohere-ai',
          'PerplexityBot',
          'Bytespider',
          'Amazonbot',
          'Applebot-Extended',
          'Meta-ExternalAgent',
          'FacebookBot',
          'Omgilibot',
          'Omgili',
          'ImagesiftBot',
          'Diffbot',
          'YouBot',
        ],
        disallow: ['/'],
      },
    ],
  },

  // ═══════════════════════════════════════════════════════════
  // Sitemap (@nuxtjs/sitemap)
  // zeroRuntime: генерация на этапе nuxt generate (SSG)
  // urls: динамические /heroes/* из API (если доступен)
  // ═══════════════════════════════════════════════════════════
  sitemap: {
    zeroRuntime: true,

    urls: async (): Promise<SitemapUrlInput[]> => {
      try {
        const apiBaseUrl =
          process.env.NUXT_API_INTERNAL_BASE_URL ||
          process.env.NUXT_PUBLIC_API_BASE_URL ||
          'https://api.neverforgotten.ru';

        const heroUrls: SitemapUrlInput[] = [];
        let cursor = '';
        const MAX_HEROES = 10_000;

        while (heroUrls.length < MAX_HEROES) {
          const res = await fetch(`${apiBaseUrl}/emh.v1.HeroService/ListHeroes`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              pagination: { pageSize: 100, cursor },
            }),
          });

          if (!res.ok) break;

          const data = await res.json();

          for (const hero of data.heroes ?? []) {
            const url: SitemapUrlInput = {
              loc: `/heroes/${hero.id}`,
              changefreq: 'monthly', // ← TypeScript теперь знает контекст
              priority: 0.7,
            };
            heroUrls.push(url);
          }

          cursor = data.pagination?.nextCursor ?? '';
          if (!cursor) break;
        }

        console.log(`[sitemap] ${heroUrls.length} hero URLs collected`);
        return heroUrls;
      } catch {
        console.warn('[sitemap] API unavailable, skipping dynamic hero URLs');
        return [];
      }
    },
  },

  // ═══════════════════════════════════════════════════════════
  // Schema.org (nuxt-schema-org)
  // identity берётся из site.identity автоматически
  // ═══════════════════════════════════════════════════════════
  schemaOrg: {
    minify: true,
  },

  // ═══════════════════════════════════════════════════════════
  // @nuxt/scripts — оптимизация загрузки скриптов
  // ═══════════════════════════════════════════════════════════
  scripts: {
    // Дефолтные пресеты: не нужны, настраиваем вручную
    defaultScriptOptions: {
      bundle: false, // НЕ бандлим внешние скрипты
    },
  },

  // ═══════════════════════════════════════════════════════════
  // Nitro / Prerender
  // ═══════════════════════════════════════════════════════════
  nitro: {
    storage: {
      // Явное указание драйвера ISR-кэша.
      // По умолчанию Nitro может использовать memory-драйвер (который сбрасывается при рестарте),
      // fs-драйвер гарантирует персистентность между запросами.
      'nitro:routes': {
        driver: 'fs',
        base: './.data/nitro-routes',
      },
    },
    prerender: {
      failOnError: false,
      ignore: ['/api/**'],
      crawlLinks: false,
    },
    routeRules: {
      // ВРЕМЕННО ОТКЛЮЧЕНО ДЛЯ ОТЛАДКИ ПРОДА
      // После стабилизации вернуть ISR с expiration.

      '/heroes': {
        isr: false,
        // isr: {
        //   expiration: 60,
        //   passQuery: true,
        //   allowQuery: [
        //     'search_query',
        //     'conflict_id',
        //     'location_id',
        //     'cursor',
        //     'date_from',
        //     'date_to',
        //   ],
        // },
      },
      '/heroes/': {
        isr: false,
        // isr: {
        //   expiration: 60,
        //   passQuery: true,
        //   allowQuery: [
        //     'search_query',
        //     'conflict_id',
        //     'location_id',
        //     'cursor',
        //     'date_from',
        //     'date_to',
        //   ],
        // },
      },
      '/heroes/**': {
        isr: false,
        // isr: { expiration: 60 },
      },
      '/conflicts': { isr: false },
      '/locations': { isr: false },

      // Статические страницы — можно оставить как есть
      '/': { prerender: true },
      '/contacts': { prerender: true },
      '/about': { prerender: true },

      // Админка не индексируется
      '/admin/': { robots: false },
    },
  },

  // Плавная прокрутка
  router: {
    options: {
      scrollBehaviorType: 'smooth',
    },
  },

  experimental: {
    typedPages: true,
  },
});
