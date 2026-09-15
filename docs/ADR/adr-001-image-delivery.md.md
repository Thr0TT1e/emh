# ADR-001: Image Delivery Strategy

## Статус
**Принято** (2026-09-15)

## Контекст
Lighthouse показывает **score 0** для Image Delivery на странице `/heroes/[id]`. Требуется добавить `<picture>` с WebP/AVIF и responsive sizes.

Бэкенд уже генерирует WebP через `ThumbnailWorker`:
- **Thumbnail**: `thumbnails/<photoID>.webp` (480×600, q80, crop по FaceBox)
- **Original**: `photos/<photoID>.webp` (q90, полный кадр)

AVIF **не генерируется** — в стеке нет AVIF-энкодера (только `deepteams/webp` lossy).

## Решение

### 1. Не строить `<picture>` с `<source type="image/webp">`

После обработки воркером **оба URL уже WebP**:
- `photo.url` → `photos/<photoID>.webp`
- `photo.thumbnailUrl` → `thumbnails/<photoID>.webp`

Конструкция `<picture>` с `<source type="image/webp">` **избыточна** и добавляет лишний HTML.

**Исключение**: асинхронное окно (1-5 сек) после загрузки, когда `url` ещё JPEG/PNG, а `thumbnailUrl` ещё не готов. В этом окне нужен fallback.

### 2. Fallback стратегия в асинхронном окне

Текущая реализация в Спринте 9 корректна:
```vue
<img
  :src="photo.thumbnailUrl || photo.url"
  width="480"
  height="600"
  loading="lazy"
  decoding="async"
  :alt="altText"
/>
```

### 3. Responsive sizes для карточек

Для карточек в сетке (главная, `/heroes`):
```vue
<img
  :src="photo.thumbnailUrl || photo.url"
  width="480"
  height="600"
  sizes="(max-width: 768px) 50vw, (max-width: 1024px) 33vw, 25vw"
  loading="lazy"
  decoding="async"
  :alt="altText"
/>
```

Для детальной страницы (`/heroes/[id]`):
```vue
<img
  :src="photo.url"
  sizes="(max-width: 768px) 100vw, 50vw"
  loading="lazy"
  decoding="async"
  :alt="altText"
/>
```

### 4. Не добавлять AVIF

**Причины:**
- AVIF-энкодер в Go = тяжёлая cgo-зависимость (`libavif`)
- CPU воркера ×N для генерации AVIF
- Thumbnail уже WebP (поддержка >95% браузеров)
- Экономия ~10-15% байт на карточках, которые весят десятки КБ

**Когда добавлять AVIF:**
- Когда AVIF станет product-требованием
- Когда поддержка браузеров >98%
- Когда появится нативный Go AVIF-энкодер без cgo

### 5. Не использовать imgproxy или Nuxt Image

**Причины:**
- Чистый MinIO не умеет on-the-fly ресайз
- imgproxy = новая инфраструктура + ещё один CORS-домен через Caddy
- Nuxt Image с S3-провайдером без бэкенд-прокси = проблемы CORS/отсутствия ресайза
- Дублирование логики guards (decompression bomb, MIME-whitelist)

**Текущий выбор** (воркер в `emh-backend`):
- Один бинарник
- Ноль дополнительной инфраструктуры
- Детерминированное качество
- Все security-guards в одном месте

## Последствия

### Положительные
- ✅ Простота: не нужно менять контракт `media.proto`
- ✅ Производительность: WebP уже генерируется воркером
- ✅ Безопасность: все guards в одном месте
- ✅ Надёжность: strict delete + retry + rescan

### Отрицательные
- ❌ Нет AVIF (но это ок, WebP достаточно)
- ❌ Нет srcset с несколькими ширинами (но это ок, thumbnail фиксированной ширины)

## Альтернативы (отклонены)

### A. Добавить AVIF в ThumbnailWorker
**Вердикт**: отклонено (преждевременно)
- CPU воркера ×2
- cgo-зависимость
- Экономия ~10-15% байт

### B. imgproxy перед MinIO
**Вердикт**: отклонено (overengineering)
- Новая инфраструктура
- Ещё один CORS-домен
- Дублирование логики

### C. Nuxt Image с S3-провайдером
**Вердикт**: отклонено (не работает)
- Чистый MinIO не ресайзит on-the-fly
- CORS проблемы
- Неконтролируемый CPU на SSR

## Ссылки
- `backend/internal/usecase/thumbnail_worker.go` — пайплайн, guards, strict delete
- `backend/proto/emh/v1/media.proto` — контракт загрузки
- `backend/proto/emh/v1/hero.proto` — FaceBox, Photo, HeroSummary
