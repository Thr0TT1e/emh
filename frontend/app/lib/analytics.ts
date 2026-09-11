// ═══════════════════════════════════════════════════════════
// Типы событий аналитики
// Единый контракт для всех страниц проекта
// ═══════════════════════════════════════════════════════════

/**
 * Событие просмотра карточки героя.
 * Триггер: страница /heroes/[id] успешно загрузила данные.
 */
export interface HeroViewEvent {
  event: 'hero_view';
  heroId: string;
  heroName: string;
  rank?: string;
  conflicts?: string;
}

/**
 * Событие просмотра фотографии в галерее.
 * Триггер: пользователь кликнул на фото в [id].vue.
 */
export interface PhotoViewEvent {
  event: 'photo_view';
  heroId: string;
  photoId: string;
  isMain: boolean;
}

/**
 * Событие отправки формы обратной связи.
 * Триггер: успешный ответ от ContactService.
 */
export interface FormSubmitEvent {
  event: 'form_submit';
  pageUrl: string;
  subject: string;
}

/**
 * Событие создания заявки на добавление/дополнение героя.
 * Триггер: успешный ответ от SubmissionService.
 */
export interface SubmissionCreateEvent {
  event: 'submission_create';
  mode: 'new' | 'supplement';
  targetHeroId?: string;
  hasAttachments: boolean;
}

/**
 * Событие поискового запроса.
 * Триггер: пользователь ввёл запрос (debounced 400ms).
 */
export interface SearchEvent {
  event: 'search';
  query: string;
  conflictId?: string;
  resultsCount: number;
}

/**
 * Union-тип всех событий. TypeScript проверяет,
 * что передаются правильные поля для каждого типа.
 */
export type AnalyticsEvent =
  | HeroViewEvent
  | PhotoViewEvent
  | FormSubmitEvent
  | SubmissionCreateEvent
  | SearchEvent;

/**
 * Маппинг события на target Метрики (reachGoal).
 * Target должен совпадать с целью, созданной в интерфейсе Метрики.
 */
export const METRIKA_TARGETS: Record<AnalyticsEvent['event'], string> = {
  hero_view: 'hero_view',
  photo_view: 'photo_view',
  form_submit: 'form_submit',
  submission_create: 'submission_create',
  search: 'search',
};
