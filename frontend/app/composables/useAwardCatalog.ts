// ---------------------------------------------------------------------------
// Полный каталог наград справочника для клиентской проверки дублей (/admin/awards).
// Мягкий уровень 1: не заменяет серверный AlreadyExists, а предупреждает до отправки.
// ---------------------------------------------------------------------------

import { toPlain } from '~/lib/pb';
import { AwardSchema, type AwardJson } from '~/sdk/emh/v1/award_pb';

/**
 * Курсорный обход всего справочника: pageSize=100 (hard max usecase'а) до
 * пустого nextCursor. ListAwards использует keyset-курсор (sort_order,name,id) —
 * обход без потерь/повторов; ограничение цикла — предохранитель, не логика.
 *
 * Ключ отличается от «awards-catalog» (HeroAwards.vue): там урезанные {id,name}
 * первой страницы; переиспользование ключа отдало бы неполный/чужой формат
 * в дубль-чек при SPA-навигации.
 */
export function useAwardCatalog() {
    const { award } = useApi();
    return useAsyncData<AwardJson[]>('awards-catalog-full', async () => {
        const all: AwardJson[] = [];
        let cursor = '';
        for (let page = 0; page < 100; page++) {
            const res = await award.listAwards({ pagination: { pageSize: 100, cursor } });
            all.push(...(res.awards ?? []).map((a) => toPlain(AwardSchema, a)));
            cursor = res.pagination?.nextCursor ?? '';
            if (!cursor) break;
        }
        return all;
    });
}
