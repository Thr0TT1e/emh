// Защищает маршруты админ-панели.
// Работает и при SSR: useCookie читает cookie из заголовка запроса.
// Сессия считается активной, пока есть refresh-токен (7 дней).
// Короткоживущий access восстановится прозрачно через interceptor.
export default defineNuxtRouteMiddleware((to) => {
  const { isAuthenticated, isAdmin } = useAuth();

  if (!isAuthenticated.value) {
    // Сохраняем целевой URL, чтобы после логина вернуться.
    return navigateTo(`/admin/login?next=${encodeURIComponent(to.fullPath)}`);
  }

  // Ранний отказ: сессия есть, но роли `admin` в клеймах нет.
  // Без этого запросы к admin-сервисам упадут с 403 на бэкенде.
  if (!isAdmin.value) {
    return navigateTo(`/admin/login?next=${encodeURIComponent(to.fullPath)}&reason=no-role`);
  }
});
