import type { FlexibleDateJson } from '~/sdk/emh/v1/common_pb';

// Тип для Timestamp, который может быть строкой или объектом
type TimestampLike = string | { seconds?: string | number; nanos?: number } | null | undefined;

// Хелпер для извлечения Date из TimestampLike
const toDate = (iso?: TimestampLike): Date | null => {
  if (!iso) return null;
  if (typeof iso === 'string') {
    const d = new Date(iso);
    return isNaN(d.getTime()) ? null : d;
  }
  if (typeof iso === 'object' && iso.seconds) {
    const d = new Date(Number(iso.seconds) * 1000);
    return isNaN(d.getTime()) ? null : d;
  }
  return null;
};

// ---------------------------------------------------------------------------
// Базовые форматтеры (для обратной совместимости)
// ---------------------------------------------------------------------------
export const formatYear = (iso?: TimestampLike): string => {
  const d = toDate(iso);
  return d ? d.getFullYear().toString() : '—';
};

export const formatDate = (iso?: TimestampLike): string => {
  const d = toDate(iso);
  return d
    ? d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })
    : '—';
};

// ---------------------------------------------------------------------------
// Гибкие даты (FlexibleDate)
// ---------------------------------------------------------------------------
export const precisionLabel = (p?: string): string => {
  const map: Record<string, string> = {
    DATE_PRECISION_UNSPECIFIED: 'Не указано',
    DATE_PRECISION_EXACT: 'Точная дата',
    DATE_PRECISION_MONTH: 'Месяц и год',
    DATE_PRECISION_YEAR: 'Только год',
    DATE_PRECISION_SEASON: 'Сезон и год',
    DATE_PRECISION_DAY_MONTH: 'День и месяц',
    DATE_PRECISION_RANGE: 'Диапазон дат',
    DATE_PRECISION_UNKNOWN: 'Дата неизвестна',
  };
  return map[p ?? 'DATE_PRECISION_UNSPECIFIED'] ?? 'Не указано';
};

/**
 * Форматирует FlexibleDate для детального отображения.
 * @param fd - Новый объект FlexibleDateJson.
 * @param fallbackIso - Старое поле (TimestampJson / string) для обратной совместимости.
 */
export const formatFlexibleDate = (fd?: FlexibleDateJson, fallbackIso?: TimestampLike): string => {
  if (fd) {
    if (fd.displayText?.trim()) return fd.displayText;

    const precision = fd.precision ?? 'DATE_PRECISION_UNSPECIFIED';
    const anchorIso = fd.anchorDate;
    const date = toDate(anchorIso);

    if (date) {
      switch (precision) {
        case 'DATE_PRECISION_EXACT':
          return date.toLocaleDateString('ru-RU');
        case 'DATE_PRECISION_MONTH':
          return date.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' });
        case 'DATE_PRECISION_YEAR':
          return date.getFullYear().toString();
        case 'DATE_PRECISION_SEASON': {
          const m = date.getMonth();
          const season =
            m >= 2 && m <= 4
              ? 'Весна'
              : m >= 5 && m <= 7
                ? 'Лето'
                : m >= 8 && m <= 10
                  ? 'Осень'
                  : 'Зима';
          return `${season} ${date.getFullYear()}`;
        }
      }
    }

    if (precision === 'DATE_PRECISION_UNKNOWN') return 'дата неизвестна';
    if (precision === 'DATE_PRECISION_DAY_MONTH' || precision === 'DATE_PRECISION_RANGE')
      return '—';
  }

  if (fallbackIso) return formatDate(fallbackIso);
  return '—';
};

/**
 * Форматирует FlexibleDate для списков (превью), где важен год или короткий текст.
 */
export const formatFlexibleYear = (fd?: FlexibleDateJson, fallbackIso?: TimestampLike): string => {
  if (fd) {
    if (fd.displayText?.trim()) return fd.displayText;

    const anchorIso = fd.anchorDate;
    const date = toDate(anchorIso);
    if (date) return date.getFullYear().toString();

    if (fd.precision === 'DATE_PRECISION_UNKNOWN') return '—';
  }

  if (fallbackIso) return formatYear(fallbackIso);
  return '—';
};

// ---------------------------------------------------------------------------
// Лейблы для Enum (остаются без изменений)
// ---------------------------------------------------------------------------
export const locationTypeLabel = (t?: string): string =>
  ({
    HERO_LOCATION_TYPE_BIRTH: 'Место рождения',
    HERO_LOCATION_TYPE_DEATH: 'Место гибели',
    HERO_LOCATION_TYPE_BURIAL: 'Место захоронения',
    HERO_LOCATION_TYPE_RESIDENCE: 'Место жительства',
  })[t ?? ''] ?? 'Место';

export const conflictTypeLabel = (t?: string): string =>
  ({
    CONFLICT_TYPE_GLOBAL: 'Мировая война',
    CONFLICT_TYPE_LOCAL: 'Локальный конфликт',
    CONFLICT_TYPE_PEACEKEEPING: 'Миротворческая операция',
    CONFLICT_TYPE_COUNTER_TERRORISM: 'КТО',
    CONFLICT_TYPE_SPECIAL_OPERATION: 'СВО',
  })[t ?? ''] ?? 'Конфликт';

export const sourceTypeLabel = (t?: string): string =>
  ({
    HERO_SOURCE_TYPE_VK: 'ВКонтакте',
    HERO_SOURCE_TYPE_WEBSITE: 'Веб-сайт',
    HERO_SOURCE_TYPE_ARCHIVE: 'Архив',
    HERO_SOURCE_TYPE_BOOK: 'Книга',
  })[t ?? ''] ?? 'Источник';

export const relationTypeLabel = (t?: string): string =>
  ({
    HERO_RELATION_TYPE_FATHER_SON: 'Отец и сын',
    HERO_RELATION_TYPE_COMRADE: 'Однополчанин',
    HERO_RELATION_TYPE_COMMANDER: 'Командир',
  })[t ?? ''] ?? 'Связь';
