/**
 * Синхронизирует @nuxtjs/color-mode с PrimeVue theme.
 * Использовать в app.vue или layout.
 */
export function useThemeSync() {
  const colorMode = useColorMode();

  const syncTheme = (newMode: string) => {
    // Проверяем, что мы в браузере (не на сервере)
    if (!import.meta.client) return;

    if (newMode === 'dark') {
      document.documentElement.classList.add('dark-mode');
      document.documentElement.classList.remove('light-mode');
    } else {
      document.documentElement.classList.add('light-mode');
      document.documentElement.classList.remove('dark-mode');
    }
  };

  // Наблюдаем за изменением темы
  watch(() => colorMode.value, syncTheme, { immediate: true });
}
