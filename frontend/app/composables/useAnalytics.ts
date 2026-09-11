import type { AnalyticsEvent, METRIKA_TARGETS } from '~/lib/analytics';

declare global {
  interface Window {
    ym?: (id: number, method: string, ...args: unknown[]) => void;
  }
}

// ID счётчика Яндекс.Метрики
const METRIKA_ID = 111894071;

/**
 * Composable для отправки событий в Яндекс.Метрику.
 *
 * Особенности:
 * - В dev-режиме логирует события в консоль (без реальных запросов)
 * - Проверяет наличие window.ym перед отправкой (скрипт может быть ещё не загружен)
 * - Типизированный API: TS проверяет правильность параметров
 */
export function useAnalytics() {
  const isDev = process.dev;

  /**
   * Отправить событие в Метрику через reachGoal.
   * Если скрипт ещё не загружен или мы в dev — логируем в консоль.
   */
  function track(evt: AnalyticsEvent): void {
    if (isDev) {
      console.log('[analytics]', evt.event, evt);
      return;
    }

    if (typeof window === 'undefined') {
      // SSR — пропускаем, событие отправится на клиенте
      return;
    }

    if (!window.ym) {
      console.warn('[analytics] ym() is not available yet');
      return;
    }

    // Деструктурируем event из параметров — в Метрику идут только данные
    const { event, ...params } = evt;
    const target = event as keyof typeof METRIKA_TARGETS;

    try {
      window.ym(METRIKA_ID, 'reachGoal', target, params);
    } catch (err) {
      console.error('[analytics] ym() failed:', err);
    }
  }

  /**
   * Хелпер: трек просмотра карточки героя.
   */
  function trackHeroView(params: Omit<AnalyticsEvent & { event: 'hero_view' }, 'event'>): void {
    track({ event: 'hero_view', ...params });
  }

  /**
   * Хелпер: трек просмотра фото.
   */
  function trackPhotoView(params: Omit<AnalyticsEvent & { event: 'photo_view' }, 'event'>): void {
    track({ event: 'photo_view', ...params });
  }

  /**
   * Хелпер: трек отправки формы обратной связи.
   */
  function trackFormSubmit(params: Omit<AnalyticsEvent & { event: 'form_submit' }, 'event'>): void {
    track({ event: 'form_submit', ...params });
  }

  /**
   * Хелпер: трек создания заявки.
   */
  function trackSubmissionCreate(
    params: Omit<AnalyticsEvent & { event: 'submission_create' }, 'event'>,
  ): void {
    track({ event: 'submission_create', ...params });
  }

  /**
   * Хелпер: трек поискового запроса.
   */
  function trackSearch(params: Omit<AnalyticsEvent & { event: 'search' }, 'event'>): void {
    track({ event: 'search', ...params });
  }

  return {
    track,
    trackHeroView,
    trackPhotoView,
    trackFormSubmit,
    trackSubmissionCreate,
    trackSearch,
  };
}
