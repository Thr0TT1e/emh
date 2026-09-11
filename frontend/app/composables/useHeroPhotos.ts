import type { PhotoJson } from '~/sdk/emh/v1/hero_pb';

export interface UseHeroPhotosOptions {
  heroId: string;
  initialPhotos?: PhotoJson[]; // из GetHero (SSR)
  pageSize?: number;
  mode?: 'admin' | 'public';
}

export const useHeroPhotos = (options: UseHeroPhotosOptions) => {
  const { heroId, initialPhotos = [], pageSize = 20, mode = 'public' } = options;
  const { hero } = useApi();

  const photos = ref<PhotoJson[]>(initialPhotos);
  const totalCount = ref<number>(initialPhotos.length);
  const nextCursor = ref<string>('');
  const loading = ref(false);
  const hasMore = ref(false);

  const loadMore = async () => {
    if (loading.value) return;

    loading.value = true;
    try {
      const res = await hero.listHeroPhotos({
        heroId,
        pagination: {
          pageSize,
          cursor: nextCursor.value,
        },
      });

      // Добавляем новые фото (избегаем дублей)
      const existingIds = new Set(photos.value.map((p) => p.id));
      const newPhotos = res.photos.filter((p) => !existingIds.has(p.id));
      photos.value = [...photos.value, ...newPhotos];

      // Обновляем пагинацию
      nextCursor.value = res.pagination?.nextCursor ?? '';
      hasMore.value = !!res.pagination?.nextCursor;

      // total_count доступен только на первой странице
      if (res.pagination?.totalCount) {
        totalCount.value = Number(res.pagination.totalCount);
      }
    } catch (error) {
      console.error('[useHeroPhotos] Failed to load photos:', error);
    } finally {
      loading.value = false;
    }
  };

  const refresh = async () => {
    // Полный сброс и загрузка с нуля (для админки после upload/delete)
    photos.value = [];
    nextCursor.value = '';
    hasMore.value = true;
    await loadMore();
  };

  // Автоматически загружаем первую страницу, если initialPhotos пустой
  onMounted(async () => {
    if (initialPhotos.length === 0 || mode === 'admin') {
      hasMore.value = true;
      await loadMore();
    } else {
      // Для публичного режима с SSR: проверяем, есть ли ещё фото
      if (initialPhotos.length >= pageSize) {
        hasMore.value = true;
        nextCursor.value = initialPhotos[initialPhotos.length - 1]?.id ?? '';
      }
    }
  });

  return {
    photos,
    totalCount,
    loading,
    hasMore,
    loadMore,
    refresh,
  };
};
