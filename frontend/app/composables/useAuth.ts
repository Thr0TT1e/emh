import { ACCESS_COOKIE, REFRESH_COOKIE } from '~/constants/auth';
import { isUnauthenticated } from '~/lib/errors';

const COOKIE_OPTS = {
  sameSite: 'lax' as const,
  // secure: true  // включите в prod при HTTPS; для localhost dev оставляем off
  path: '/',
};

// ---------------------------------------------------------------------------
// Мьютекс на refresh. Если 5 параллельных запросов получили 401,
// refresh выполнится ОДИН раз, остальные дождутся его результата
// и повторят свои запросы с новым access.
// ---------------------------------------------------------------------------
let refreshInflight: Promise<string | null> | null = null;

export const useAuth = () => {
  // Access: короткоживущий (15 мин = 900 сек).
  // При истечении interceptor автоматически сделает refresh.
  const accessToken = useCookie<string | null>(ACCESS_COOKIE, {
    ...COOKIE_OPTS,
    maxAge: 60 * 15,
  });

  // Refresh: долгоживущий (7 дней). Главный признак сессии.
  const refreshToken = useCookie<string | null>(REFRESH_COOKIE, {
    ...COOKIE_OPTS,
    maxAge: 60 * 60 * 24 * 7,
  });

  // Сессия активна, пока жив refresh. Access может быть пустым
  // (только что истёк) — interceptor его восстановит прозрачно.
  const isAuthenticated = computed(() => !!refreshToken.value);

  // -----------------------------------------------------------------------
  // Вход: получаем пару токенов и сохраняем оба.
  // -----------------------------------------------------------------------
  const login = async (username: string, password: string) => {
    const { auth } = useApi();
    const res = await auth.login({ username, password });
    accessToken.value = res.accessToken;
    refreshToken.value = res.refreshToken;
    return res;
  };

  // -----------------------------------------------------------------------
  // Выход: отзыв refresh на бэкенде + локальная очистка.
  // Даже если RPC упадёт (сеть), локально чистим обязательно — иначе
  // при следующем заходе будет «призрачная» сессия.
  // -----------------------------------------------------------------------
  const logout = async () => {
    const { auth } = useApi();
    const rt = refreshToken.value;
    if (rt) {
      try {
        await auth.logout({ refreshToken: rt });
      } catch {
        /* не блокируем выход */
      }
    }
    accessToken.value = null;
    refreshToken.value = null;
  };

  // -----------------------------------------------------------------------
  // Refresh: обмен refresh на новую пару (ротация).
  // Мьютекс гарантирует один refresh одновременно.
  // -----------------------------------------------------------------------
  const refreshSession = async (): Promise<string | null> => {
    if (refreshInflight) return refreshInflight;
    refreshInflight = (async () => {
      const { auth } = useApi();
      const rt = refreshToken.value;
      if (!rt) return null;
      try {
        const res = await auth.refresh({ refreshToken: rt });
        // Ротация: старый refresh отозван, сохраняем новую пару.
        accessToken.value = res.accessToken;
        refreshToken.value = res.refreshToken;
        return res.accessToken;
      } catch (err) {
        // Refresh не удался (истёк / отозван / сеть) — сессия мертва.
        accessToken.value = null;
        refreshToken.value = null;
        if (import.meta.client && !isUnauthenticated(err)) {
          // На 401 редирект сделает interceptor / layout, здесь не дублируем.
          console.warn('[auth] refresh failed:', err);
        }
        return null;
      }
    })();
    try {
      return await refreshInflight;
    } finally {
      refreshInflight = null;
    }
  };

  // Очистка без RPC (для interceptor'а, когда refresh уже провалился).
  const clearSession = () => {
    accessToken.value = null;
    refreshToken.value = null;
  };

  // Claims из JWT access-токена. Автоматически пересчитываются
  // при refresh, т.к. accessToken — реактивный.
  const claims = computed<{ username?: string; role?: string } | null>(() => {
    const t = accessToken.value;
    if (!t) return null;
    try {
      const b64 = t.split('.')[1]!.replace(/-/g, '+').replace(/_/g, '/');
      return JSON.parse(atob(b64));
    } catch {
      return null;
    }
  });

  // Проверка роли `admin` из JWT-клейм (бэкенд требует её для всех
  // admin-сервисов, см. справку M1). Используем для раннего отказа
  // и отображения «Нет прав доступа» без лишнего запроса.
  const isAdmin = computed(() => claims.value?.role === 'admin');

  return {
    accessToken,
    refreshToken,
    isAuthenticated,
    claims,
    isAdmin,
    login,
    logout,
    refreshSession,
    clearSession,
  };
};
