// ═══════════════════════════════════════════════════════════
// Плагин инициализации Яндекс.Метрики.
// Запускается на клиенте через @nuxt/scripts (lazy, onNuxtReady).
// Без Web Worker — чтобы работали Вебвизор и карта кликов.
// ═══════════════════════════════════════════════════════════

const METRIKA_ID = 111894071;

export default defineNuxtPlugin(() => {
  // В dev-режиме Метрику не подключаем — избегаем мусора в статистике
  if (process.dev) {
    console.info('[yandex-metrika] disabled in dev mode');
    return;
  }

  const { load } = useScript(
    {
      // Загружаем скрипт Метрики только после готовности Nuxt (не блокируем FCP)
      src: `https://mc.yandex.ru/metrika/tag.js?id=${METRIKA_ID}`,
      async: true,
    },
    {
      trigger: 'onNuxtReady',
      // После загрузки скрипта инициализируем счётчик
      use() {
        // Метрика создаёт глобальную функцию ym()
        const w = window as Window & {
          ym?: (id: number, method: string, ...args: unknown[]) => void;
        };

        // Стандартная обёртка Метрики: очередь вызовов до загрузки скрипта
        w.ym =
          w.ym ||
          function (...args: unknown[]) {
            ((w.ym as unknown as { a: unknown[][] }).a ||= []).push(args);
          };
        (w.ym as unknown as { l: number }).l = Date.now();

        // Инициализация с включённым Вебвизором и картой кликов
        w.ym(METRIKA_ID, 'init', {
          ssr: true,
          webvisor: true,
          trackHash: true,
          clickmap: true,
          ecommerce: 'dataLayer',
          referrer: document.referrer,
          url: location.href,
          accurateTrackBounce: true,
          trackLinks: true,
        });

        return w.ym;
      },
    },
  );

  // Вызываем load() явно — useScript вернёт Promise
  load().catch((err) => {
    console.error('[yandex-metrika] failed to load script:', err);
  });
});
