import { HeroLocationType } from '~/sdk/emh/v1/enums_emh_pb';
import type {
  ExtractedAwardJson,
  ExtractedConflictJson,
  ExtractedLocationJson,
} from '~/sdk/emh/v1/extraction_pb';

// Общий вход: подходит и для ExtractHeroDataResponse, и для ExtractFromSubmissionResponse
export interface ExtractionLinksInput {
  conflicts?: ExtractedConflictJson[];
  awards?: ExtractedAwardJson[];
  locations?: ExtractedLocationJson[];
  sourceUrls?: string[];
}

export interface ExtractionLinkReport {
  applied: string[];
  skipped: string[];
  errors: string[];
}

export const useExtractionLinks = () => {
  const { heroAdmin, conflict, location, award } = useApi();

  // Кэши справочников на время одной сессии привязки (избегаем повторных загрузок)
  let conflictsCache: { id?: string; name?: string }[] | null = null;
  let awardsCache: { id?: string; name?: string }[] | null = null;

  const getConflicts = async () => {
    if (!conflictsCache) {
      const res = await conflict.listConflicts({});
      conflictsCache = res.conflicts;
    }
    return conflictsCache;
  };

  const getAwards = async () => {
    if (!awardsCache) {
      // Проверь сигнатуру в award_pb: если есть searchQuery — можно сузить
      const res = await award.listAwards({});
      awardsCache = res.awards;
    }
    return awardsCache;
  };

  // Регистронезависимый поиск по названию
  const findByName = (list: { id?: string; name?: string }[], name: string): string | null => {
    const found = list.find((item) => item.name?.toLowerCase() === name.toLowerCase());
    return found?.id ?? null;
  };

  // Маппинг строкового типа локации из LLM → HeroLocationType
  const mapHeroLocationType = (type?: string): HeroLocationType => {
    switch (type?.toLowerCase()) {
      case 'birth':
        return HeroLocationType.BIRTH;
      case 'death':
        return HeroLocationType.DEATH;
      case 'burial':
        return HeroLocationType.BURIAL;
      case 'residence':
        return HeroLocationType.RESIDENCE;
      default:
        return HeroLocationType.UNSPECIFIED;
    }
  };

  // Определение типа источника по URL (формат из format.ts)
  const guessSourceType = (url: string): string => {
    if (url.includes('vk.com')) return 'HERO_SOURCE_TYPE_VK';
    if (url.includes('pamyat-naroda.ru') || url.includes('archive')) {
      return 'HERO_SOURCE_TYPE_ARCHIVE';
    }
    return 'HERO_SOURCE_TYPE_WEBSITE';
  };

  // Нормализация даты награждения к формату, который требует бэкенд (^\d{4}-\d{2}-\d{2}$).
  // Пустая строка проходит валидацию (паттерн не применяется к пустым в буфе).
  const normalizeAwardDate = (date?: string): string => {
    const s = date?.trim();
    if (!s) return '';

    // YYYY-MM-DD — уже валидно
    if (/^\d{4}-\d{2}-\d{2}$/.test(s)) return s;
    // YYYY-MM → дополняем до первого числа
    if (/^\d{4}-\d{2}$/.test(s)) return `${s}-01`;
    // YYYY → дополняем до 1 января
    if (/^\d{4}$/.test(s)) return `${s}-01-01`;
    // Нераспознанное — отдаём пустым, чтобы не сломать валидацию
    return '';
  };

  const applyExtractionLinks = async (
    heroId: string,
    data: ExtractionLinksInput,
  ): Promise<ExtractionLinkReport> => {
    const report: ExtractionLinkReport = { applied: [], skipped: [], errors: [] };

    // ── Конфликты ──────────────────────────────────────────────
    for (const c of data.conflicts ?? []) {
      if (!c.name) continue;
      try {
        const conflictId = findByName(await getConflicts(), c.name);
        if (conflictId) {
          await heroAdmin.addHeroConflict({
            heroId,
            conflictId,
            specificLocation: c.specificLocation ?? '',
            rankAtConflict: c.rankAtConflict ?? '',
          });
          report.applied.push(`Конфликт: ${c.name}`);
        } else {
          report.skipped.push(`Конфликт не в справочнике: ${c.name}`);
        }
      } catch (err: any) {
        report.errors.push(`Конфликт «${c.name}»: ${err?.message ?? err}`);
      }
    }

    // ── Награды ────────────────────────────────────────────────
    for (const a of data.awards ?? []) {
      if (!a.name) continue;
      try {
        const awardId = findByName(await getAwards(), a.name);
        if (awardId) {
          await heroAdmin.addHeroAward({
            heroId,
            awardId,
            awardDate: normalizeAwardDate(a.awardDate),
            decreeNumber: a.decreeNumber ?? '',
          });
          report.applied.push(`Награда: ${a.name}`);
        } else {
          report.skipped.push(`Награда не в справочнике: ${a.name}`);
        }
      } catch (err: any) {
        report.errors.push(`Награда «${a.name}»: ${err?.message ?? err}`);
      }
    }

    // ── Локации ────────────────────────────────────────────────
    for (const l of data.locations ?? []) {
      if (!l.name) continue;
      const locType = mapHeroLocationType(l.heroLocationType);
      // AddHeroLocationRequest.type имеет not_in = 0 — UNSPECIFIED недопустим
      if (locType === HeroLocationType.UNSPECIFIED) {
        report.skipped.push(`Локация «${l.name}»: не определён тип связи`);
        continue;
      }
      try {
        const res = await location.listLocations({ searchQuery: l.name });
        const found = res.locations.find(
          (loc) => loc.name?.toLowerCase() === l.name!.toLowerCase(),
        );
        if (found?.id) {
          await heroAdmin.addHeroLocation({ heroId, locationId: found.id, type: locType });
          report.applied.push(`Локация: ${l.name}`);
        } else {
          report.skipped.push(`Локация не в справочнике: ${l.name}`);
        }
      } catch (err: any) {
        report.errors.push(`Локация «${l.name}»: ${err?.message ?? err}`);
      }
    }

    // ── Источники (извлекаются из source_urls) ────────────────
    for (const url of data.sourceUrls ?? []) {
      try {
        await heroAdmin.addHeroSource({
          heroId,
          url,
          title: '',
          sourceType: guessSourceType(url),
          excerpt: '',
        });
        report.applied.push(`Источник: ${url}`);
      } catch (err: any) {
        report.errors.push(`Источник ${url}: ${err?.message ?? err}`);
      }
    }

    return report;
  };

  return { applyExtractionLinks };
};
