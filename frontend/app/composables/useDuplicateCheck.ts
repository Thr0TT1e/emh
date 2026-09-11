// ---------------------------------------------------------------------------
// Клиентская мягкая проверка дубликатов заявок.
// НЕ заменяет серверную проверку (бэкенд использует SHA-256 хеш),
// а лишь предупреждает пользователя о возможном повторе перед отправкой.
// ---------------------------------------------------------------------------

const RECENT_SUBMISSIONS_KEY = 'emh_recent_submissions';
const MAX_RECENT = 10;

// Быстрая не-криптографическая хеш-функция (только для эвристики)
const simpleHash = (str: string): string => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash |= 0; // 32-bit int
  }
  return hash.toString(36);
};

export const useDuplicateCheck = () => {
  // Проверка: отправлялась ли недавно заявка с тем же целевым героем и похожим содержимым.
  // Возвращает найденную запись или null. НЕ блокирует отправку.
  const checkLocal = (targetHeroId: string, payloadJson: string) => {
    if (!import.meta.client) return null;
    try {
      const recent = JSON.parse(localStorage.getItem(RECENT_SUBMISSIONS_KEY) || '[]');
      const hash = simpleHash(payloadJson);
      return (
        recent.find(
          (s: { targetHeroId: string; hash: string; timestamp: number }) =>
            s.targetHeroId === targetHeroId && s.hash === hash,
        ) ?? null
      );
    } catch {
      return null;
    }
  };

  // Запись после успешной отправки.
  const recordSubmission = (targetHeroId: string, payloadJson: string) => {
    if (!import.meta.client) return;
    try {
      const recent = JSON.parse(localStorage.getItem(RECENT_SUBMISSIONS_KEY) || '[]');
      recent.unshift({
        targetHeroId,
        hash: simpleHash(payloadJson),
        timestamp: Date.now(),
      });
      localStorage.setItem(RECENT_SUBMISSIONS_KEY, JSON.stringify(recent.slice(0, MAX_RECENT)));
    } catch {
      // localStorage недоступен (приватный режим) — молча игнорируем
    }
  };

  return { checkLocal, recordSubmission };
};
