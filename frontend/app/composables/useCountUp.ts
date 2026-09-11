export const useCountUp = (target: MaybeRefOrGetter<number>, duration = 900) => {
  const current = ref(0);
  const targetRef = toRef(target);

  // Анимация возможна только в браузере. На сервере оставляем 0,
  // чтобы не было hydration-mismatch (клиент тоже стартует с 0).
  if (import.meta.client) {
    let frame: number | undefined;

    const animate = (to: number) => {
      if (frame) cancelAnimationFrame(frame);
      const from = current.value;
      const start = performance.now();

      const tick = (now: number) => {
        const p = Math.min((now - start) / duration, 1);
        const eased = 1 - Math.pow(1 - p, 3); // ease-out cubic
        current.value = Math.round(from + (to - from) * eased);
        if (p < 1) frame = requestAnimationFrame(tick);
      };
      frame = requestAnimationFrame(tick);
    };

    watch(targetRef, (to) => animate(to), { immediate: true });

    // Отменяем незавершённый кадр при размонтировании,
    // чтобы не было утечки и обновления current после уничтожения.
    onScopeDispose(() => {
      if (frame) cancelAnimationFrame(frame);
    });
  }

  return current;
};
