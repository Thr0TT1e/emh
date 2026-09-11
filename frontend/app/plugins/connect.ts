import { Code, ConnectError, type Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

import { createApi } from '~/lib/api';

// Методы, которые НЕ retry-им при 401: иначе бесконечный цикл
// (Login → 401 → Refresh → 401 → Refresh → …).
const NO_RETRY_METHODS = ['/Login', '/Refresh', '/Logout'];

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig();
  const baseUrl = import.meta.server
    ? config.apiInternalBaseUrl || config.public.apiBaseUrl
    : config.public.apiBaseUrl;

  // useAuth вызываем внутри defineNuxtPlugin, а не на верхнем уровне модуля,
  // чтобы useCookie/useRuntimeConfig работали корректно и на сервере, и на клиенте.
  const { accessToken, refreshSession, clearSession } = useAuth();
  const router = useRouter();

  const authInterceptor: Interceptor = (next) => async (req) => {
    // 1. Ставим текущий access (если есть).
    if (accessToken.value) {
      req.header.set('Authorization', `Bearer ${accessToken.value}`);
    }

    try {
      return await next(req);
    } catch (err) {
      if (!(err instanceof ConnectError)) throw err;

      // 429 Rate-limit: НЕ retry-им, пробрасываем дальше — пусть UI решает.
      if (err.code === Code.ResourceExhausted) throw err;

      // 403 PermissionDenied: пользователь авторизован, но нет роли `admin`.
      // НЕ делаем редирект на логин (сессия валидна) — пробрасываем,
      // чтобы UI показал «Нет прав доступа».
      if (err.code === Code.PermissionDenied) throw err;

      // 401 Unauthenticated: пробуем refresh + повтор запроса.
      if (err.code === Code.Unauthenticated) {
        const method = req.url;
        if (NO_RETRY_METHODS.some((m) => method.endsWith(m))) {
          // Это сам Login/Refresh/Logout — не повторяем, отдаём ошибкой.
          throw err;
        }

        const newToken = await refreshSession();
        if (!newToken) {
          // Сессия не восстановима — чистим и отправляем на логин.
          clearSession();
          if (import.meta.client && router.currentRoute.value.path.startsWith('/admin')) {
            router.push('/admin/login?reason=expired');
          }
          throw err;
        }

        // Повторяем ИСХОДНЫЙ запрос с новым access.
        req.header.set('Authorization', `Bearer ${newToken}`);
        return next(req);
      }

      // Все остальные коды (NotFound, InvalidArgument, Internal, …) —
      // пробрасываем как есть, UI обрабатывает их локально.
      throw err;
    }
  };

  const transport = createConnectTransport({
    baseUrl,
    useBinaryFormat: false,
    interceptors: [authInterceptor],
  });

  return { provide: { api: createApi(transport) } };
});
