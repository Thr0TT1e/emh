import { toPlain } from '~/lib/pb';
import { ConflictSchema, type ConflictJson } from '~/sdk/emh/v1/conflict_pb';
import { HeroSummarySchema, type HeroSummaryJson } from '~/sdk/emh/v1/hero_pb';

export const HEROES_PAGE_SIZE = 20;

export interface HeroesPageData {
  heroes: HeroSummaryJson[];
  nextCursor: string;
  total: number;
}

/**
 * Справочник конфликтов — общий кэш:
 * счётчик «конфликтов» в шапке (layout) и фильтр на главной (index).
 */
export function useConflictsCatalog() {
  const { conflict } = useApi();
  return useAsyncData<ConflictJson[]>('conflicts-catalog', async () => {
    const res = await conflict.listConflicts({});
    return (res.conflicts ?? []).map((c) => toPlain(ConflictSchema, c));
  });
}

/**
 * Первая страница реестра героев + total_count.
 *
 * full = true  → pageSize 20, ключ «heroes:page»  (нужно на «/»);
 * full = false → pageSize 1,  ключ «heroes:count» (только счётчик в шапке).
 *
 * На «/» layout и index вызывают full=true — Nuxt дедуплицирует
 * по ключу, и listHeroes выполняется один раз.
 */
export function useHeroesPage(full: boolean) {
  const { hero } = useApi();
  return useAsyncData<HeroesPageData>(`heroes:${full ? 'page' : 'count'}`, async () => {
    const res = await hero.listHeroes({
      pagination: { pageSize: full ? HEROES_PAGE_SIZE : 1 },
    });
    return {
      heroes: (res.heroes ?? []).map((h) => toPlain(HeroSummarySchema, h)),
      nextCursor: res.pagination?.nextCursor ?? '',
      total: Number(res.pagination?.totalCount ?? 0n),
    };
  });
}
