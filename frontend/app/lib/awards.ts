import type { AwardJurisdictionJson, AwardTypeJson } from '~/sdk/emh/v1/enums_emh_pb';

/**
 * Награда, пригодная для сортировки по старшинству и отрисовки планки.
 * Структурно подходит и `AwardJson` (справочник), и `HeroAwardJson` (привязка к герою).
 */
export interface SortableAward {
  type?: AwardTypeJson;
  jurisdiction?: AwardJurisdictionJson;
  isJubilee?: boolean;
  wornWithoutBar?: boolean;
  sortOrder?: number;
  ribbonImageUrl?: string;
  imageUrl?: string;
  name?: string;
  awardName?: string;
}

/**
 * Порог переключения (ADR-005 §3): при большем числе наград с лентой
 * вместо текстового списка рендерится орденская планка.
 */
export const RIBBON_BAR_THRESHOLD = 3;

// Старшинство по приказу МО РФ №1500: сначала принадлежность, затем тип награды.
const JURISDICTION_RANK: Record<string, number> = {
  AWARD_JURISDICTION_RUSSIAN_FEDERATION: 0,
  AWARD_JURISDICTION_USSR: 1,
  AWARD_JURISDICTION_DEPARTMENTAL: 2,
};

const TYPE_RANK: Record<string, number> = {
  AWARD_TYPE_ORDER: 0,
  AWARD_TYPE_MEDAL: 1,
  AWARD_TYPE_BADGE: 2,
};

// Не указанная принадлежность или тип всегда младше известных значений.
const UNKNOWN_RANK = 99;

const rankOf = (ranks: Record<string, number>, value?: string): number =>
  value === undefined ? UNKNOWN_RANK : (ranks[value] ?? UNKNOWN_RANK);

const nameOf = (a: SortableAward): string => a.awardName ?? a.name ?? '';

/**
 * Сравнивает награды по старшинству: jurisdiction → type → isJubilee → sortOrder → name.
 *
 * Enum'ы приходят из JSON строками, поэтому ранги берутся из таблиц соответствия:
 * арифметика над строковыми значениями дала бы NaN.
 *
 * `sortOrder` отсутствует у `HeroAward` — для привязок к герою ось пропускается,
 * и порядок внутри равных рангов сохраняет серверный `ORDER BY a.sort_order ASC`
 * (Array.prototype.sort стабилен).
 */
export const compareAwards = (a: SortableAward, b: SortableAward): number => {
  const byJurisdiction =
    rankOf(JURISDICTION_RANK, a.jurisdiction) - rankOf(JURISDICTION_RANK, b.jurisdiction);
  if (byJurisdiction !== 0) return byJurisdiction;

  const byType = rankOf(TYPE_RANK, a.type) - rankOf(TYPE_RANK, b.type);
  if (byType !== 0) return byType;

  // Боевые награды старше юбилейных.
  if (a.isJubilee !== b.isJubilee) return a.isJubilee ? 1 : -1;

  const bySortOrder = (a.sortOrder ?? 0) - (b.sortOrder ?? 0);
  if (bySortOrder !== 0) return bySortOrder;

  return nameOf(a).localeCompare(nameOf(b), 'ru');
};

/**
 * Награда входит в орденскую планку: лента загружена и награда носится на колодке.
 * Награды без ленты и `worn_without_bar` (звёзды орденов, знаки) в планку не попадают.
 */
export const isRibbonBearing = (a: SortableAward): boolean =>
  Boolean(a.ribbonImageUrl) && !a.wornWithoutBar;

/** Возвращает отсортированные по старшинству награды, входящие в планку. */
export const ribbonAwardsOf = <T extends SortableAward>(awards: readonly T[]): T[] =>
  awards.filter(isRibbonBearing).sort(compareAwards);

/** Возвращает все награды, отсортированные по старшинству (для текстового списка). */
export const sortedAwards = <T extends SortableAward>(awards: readonly T[]): T[] =>
  [...awards].sort(compareAwards);

export const awardTypeLabel = (t?: string): string =>
  ({
    AWARD_TYPE_ORDER: 'Орден',
    AWARD_TYPE_MEDAL: 'Медаль',
    AWARD_TYPE_BADGE: 'Знак отличия',
  })[t ?? ''] ?? 'Не указан';

export const awardJurisdictionLabel = (j?: string): string =>
  ({
    AWARD_JURISDICTION_RUSSIAN_FEDERATION: 'Российская Федерация',
    AWARD_JURISDICTION_USSR: 'СССР',
    AWARD_JURISDICTION_DEPARTMENTAL: 'Ведомственная',
  })[j ?? ''] ?? 'Не указана';
