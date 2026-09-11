import { Code, ConnectError } from '@connectrpc/connect';

export function isConnectError(err: unknown): err is ConnectError {
  return err instanceof ConnectError;
}

/** 429 Too Many Requests — бэкенд вернул Rate-limit. */
export function isRateLimited(err: unknown): boolean {
  return isConnectError(err) && err.code === Code.ResourceExhausted;
}

/** 401 Unauthenticated — access истёк или невалиден. */
export function isUnauthenticated(err: unknown): boolean {
  return isConnectError(err) && err.code === Code.Unauthenticated;
}

/** 403 PermissionDenied — пользователь авторизован, но нет роли `admin`. */
export function isPermissionDenied(err: unknown): boolean {
  return isConnectError(err) && err.code === Code.PermissionDenied;
}

/** 404 NotFound */
export function isNotFound(err: unknown): boolean {
  return isConnectError(err) && err.code === Code.NotFound;
}

/** 409 AlreadyExists — заявка с таким содержанием уже существует. */
export function isAlreadyExists(err: unknown): boolean {
  return isConnectError(err) && err.code === Code.AlreadyExists;
}

/** Читаем заголовок Retry-After из ConnectError (если бэкенд его шлёт). */
export function retryAfterSeconds(err: unknown): number {
  if (!isConnectError(err)) return 60;
  const raw = err.rawMessage?.match(/retry[- ]?after[:\s]*(\d+)/i)?.[1];
  const n = raw ? Number(raw) : 0;
  return Number.isFinite(n) && n > 0 ? n : 60;
}
