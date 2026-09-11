import type { JsonWriteOptions, MessageShape } from '@bufbuild/protobuf';
import { toJson } from '@bufbuild/protobuf';
import type { GenMessage } from '@bufbuild/protobuf/codegenv2';

// Эмитить implicit-поля (пустые строки, пустие массивы) явно.
export const PLAIN_JSON = { alwaysEmitImplicit: true } as const;

// Преобразует JSON-представление enum (строку-имя или число) в числовое
// значение, которое ожидает protobuf-es при создании сообщений.
export function toEnum(
  enumObj: Record<string, number | string>,
  value: string | number | undefined | null,
  fallback = 0,
): number {
  if (value === undefined || value === null) return fallback;
  if (typeof value === 'number') return value;
  const resolved = enumObj[value];
  return typeof resolved === 'number' ? resolved : fallback;
}

// Извлекает JSON-тип из сгенерированной схемы:
// GenMessage<Shape, { jsonType: J }> -> J
type JsonOf<G> =
  G extends GenMessage<any, infer Ext> ? (Ext extends { jsonType: infer J } ? J : never) : never;

/**
 * Типизированная сериализация protobuf-сообщения в plain-JSON.
 *
 * toJson() возвращает широкий JsonValue, из-за чего TS не видит конкретные
 * поля. Здесь возвращаемый тип выводится из схемы (XxxJson), поэтому на
 * выходе сразу HeroDetailJson со всеми полями — без ручных `as`.
 *
 * Пример: toPlain(HeroDetailSchema, res.hero)  // -> HeroDetailJson
 */
export function toPlain<G extends GenMessage<any, any>>(
  schema: G,
  message: MessageShape<G>,
  options: JsonWriteOptions = PLAIN_JSON,
): JsonOf<G> {
  return toJson(schema, message, options) as JsonOf<G>;
}
