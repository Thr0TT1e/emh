# ADR-001: Стратегия доставки изображений (Image Delivery)

**Статус**: Принято
**Дата**: 2026-09-16
**Контекст**: Спринт 10

### Проблема

В Lighthouse для `/heroes/[id]` аудит **Image Delivery** показывает score 0. Требование — «добавить `<picture>` с WebP/AVIF и responsive sizes».

Однако анализ бэкенд-пайплайна (`thumbnail_worker.go`) показал, что стандартная рекомендация неприменима к текущей архитектуре.

### Решение

**Отказ от классического `<picture>`/`srcset` в пользу двух фиксированных rendition + fallback.**

| Аспект | Решение | Обоснование |
|---|---|---|
| **AVIF** | Не генерировать | В стеке только `deepteams/webp` (lossy). AVIF требует `libavif` (cgo), CPU воркера ×N. Выгода ~10–15% на карточках, которые и так весят десятки КБ. |
| **`srcset`** | Не строить | Ровно два rendition: `thumbnail_url` (480×600) и `url` (оригинал). Ширина оригинала неизвестна клиенту без метаданных. |
| **`<picture>`** | Только на асинхронном окне (1–5 с) | После обработки воркером оба URL — WebP. Конструкция вырождается в обычный `<img>`. |
| **imgproxy / Nuxt Image** | Отклонены | Чистый MinIO не ресайзит на лету. Nuxt Image с S3-провайдером без прокси = проблемы CORS и неконтролируемый CPU. |
| **Контракт `media.proto`** | Без изменений | `thumbnail_url` в `GetUploadUrlResponse` был бы ошибкой: на момент presign thumbnail ещё не существует. |

### Фронтенд-паттерн

```vue
<!-- Карточка: единственный случай, когда <picture> имеет смысл -->
<picture>
  <source v-if="isWebp(photo.url)" type="image/webp" :srcset="photo.url" />
  <img
    :src="photo.thumbnailUrl || photo.url"
    width="480" height="600"
    loading="lazy" decoding="async"
    :alt="altText"
  />
</picture>
```

После обработки воркером оба источника — WebP, и `<picture>` можно заменить на обычный `<img>`.

### Альтернативы

| Вариант | Плюсы | Минусы | Вердикт |
|---|---|---|---|
| (а) Расширить воркер матрицей `formats:[webp,avif]`, `widths:[480,800,1200]` | Единая точка контроля, метрики, guards | AVIF = cgo, CPU ×N | **Когда станет product-требованием** |
| (б) imgproxy перед MinIO | On-the-fly AVIF/srcset | Новая инфраструктура, ещё один CORS-домен | Оправдан при >3–4 rendition |
| (в) Nuxt Image с S3-провайдером | — | Чистый MinIO не ресайзит, CORS, CPU на SSR | **Отклонён** |

## Ссылки
- `backend/internal/usecase/thumbnail_worker.go` — пайплайн, guards, strict delete
- `backend/proto/emh/v1/media.proto` — контракт загрузки
- `backend/proto/emh/v1/hero.proto` — FaceBox, Photo, HeroSummary
