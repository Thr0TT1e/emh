// Базовый SEO-хелпер проекта.
// Канонический домен: вежливые.рус (кириллический, ориентация на СНГ).

export const SITE_URL = 'https://вежливые.рус';
export const SITE_NAME = 'Вечная память героям';

/**
 * Делает абсолютный URL из пути.
 * Пример: absoluteUrl('/about') → https://вежливые.рус/about
 */
export function absoluteUrl(path: string): string {
  const normalized = path.startsWith('/') ? path : `/${path}`;

  return `${SITE_URL}${normalized}`;
}

/**
 * Превращает относительный URL изображения в абсолютный.
 * Если URL уже абсолютный — возвращаем как есть.
 */
export function toAbsoluteUrl(url?: string): string | undefined {
  if (!url) {
    return undefined;
  }

  if (/^https?:\/\//i.test(url)) {
    return url;
  }

  return absoluteUrl(url);
}

/**
 * Обрезает текст до SEO-длины.
 */
export function truncate(text?: string, max = 160): string {
  const value = text?.trim() ?? '';

  if (!value) {
    return '';
  }

  if (value.length <= max) {
    return value;
  }

  return `${value.slice(0, max - 1).trimEnd()}…`;
}

/**
 * Достаёт первый встреченный год из строки.
 */
export function extractYear(text?: string): string | undefined {
  return text?.match(/\d{4}/)?.[0];
}
