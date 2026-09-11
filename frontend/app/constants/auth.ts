// Ключи cookie для пары токенов.
// Access — короткоживущий (15 мин), refresh — долгоживущий (7 дней).
// Оба в cookie для SSR-совместимости: при гидратации серверный рендер
// сразу видит access и не делает лишний 401 → refresh цикл.
export const ACCESS_COOKIE = 'emh_access';
export const REFRESH_COOKIE = 'emh_refresh';

/** @deprecated используйте ACCESS_COOKIE */
export const AUTH_COOKIE = ACCESS_COOKIE;
