# Справка: пайплайн фото в EMH backend

## 0. TL;DR — ответы на 4 вопроса

| Вопрос | Ответ |
|---|---|
| **Как генерируются WebP/AVIF?** | **Вариант (а): бэкенд-воркер при загрузке.** `ThumbnailWorker` внутри процесса `emh-backend`. **AVIF не генерируется вообще** — в стеке нет AVIF-энкодера (только `deepteams/webp` lossy). |
| **Где происходит трансформация?** | В воркере, асинхронно после загрузки. MinIO остаётся «глупым» хранилищем. **Ни imgproxy, ни Nuxt Image remote provider не используются** — и не должны: чистый MinIO не умеет on-the-fly ресайз, а Nuxt Image с S3-провайдером без бэкенд-прокси дал бы нам ровно те проблемы CORS/отсутствия ресайза, которые вы предсказали. |
| **Меняется ли контракт `media.proto`?** | **Нет.** `thumbnail_url` в `GetUploadUrlResponse` **не нужен и был бы ошибкой**: на момент presign thumbnail ещё не существует (генерация асинхронная, 1–5 с, может упасть). Финальные URL клиент читает из `Photo`/`GetHero` **после** обработки. |
| **Как FaceBox-thumbnail вписывается в `<picture>`?** | `main_thumbnail_url` — это **уже WebP** (480×600, q80, crop по FaceBox). Полноразмерный оригинал после обработки — **тоже WebP** (q90). Т.е. после обработки оба renditionWebP, и `<source type="image/webp">` избыточен; `<picture>` нужен только на асинхронном окне (см. §3). |

---

## 1. Пайплайн (фактическая реализация)

```
Браузер                     Backend (emh-backend)              MinIO
   │                                  │                           │
   ├─ GetUploadUrl/Batch ────────────►│ MIME-whitelist +          │
   │◄─ presigned PUT (Content-Type    │ генерация ключа           │
   │    в подписи, expiry 15m) ───────┤                           │
   ├─ PUT файла напрямую ────────────────────────────────────────►│ heroes/2026/07/30/<uuid>.jpg
   │                                  │                           │
   ├─ AddHeroPhoto / BatchAdd ───────►│ INSERT photo (url=jpg)    │
   │                                  │ Enqueue(ThumbnailTask)───►│ (очередь in-memory,
   │                                  │   retry 3× + backpressure │    retry при переполнении)
   │                                  │                           │
   │                    ThumbnailWorker (асинхронно, 1–5 с):      │
   │                      1. KeyFromPublicURL + Download ◄────────┤
   │                      2. LimitReader 50MB + DecodeConfig      │
   │                         (guard: ≤10000px/сторона, ≤50MP)     │
   │                      3. Decode → crop (FaceBox || center 4:5)│
   │                      4. Fill 480×600 Lanczos                 │
   │                      5. Upload thumbnails/<photoID>.webp ───►│ (q80)
   │                      6. Если формат ≠ webp:                  │
   │                         Upload photos/<photoID>.webp ───────►│ (q90)
   │                         DELETE старый оригинал (strict,      │
   │                           3 retry, 404=успех)                │
   │                      7. UpdateAfterProcessing(photoID,       │
   │                         originalURL, thumbnailURL) → БД      │
   │                                  │                           │
   │◄─ GetHero / ListHeroes: url и thumbnail_url УЖЕ webp ────────┤
```

**Ключевые свойства:**
- **Триггеры Enqueue:** `AddHeroPhoto`, `BatchAddHeroPhotos`, `UpdateHeroPhoto` при `face_box` в `field_mask`, `RescanLoop` (каждые 5 мин, batch 100, фото без thumbnail), `StartupBackfill` при старте.
- **Strict delete:** если удаление старого оригинала после конвертации не удалось — задача **прерывается до обновления БД** (данные не теряются, восстановится rescan'ом).
- **Если клиент сразу загрузил WebP** — оригинал не переконвертируется и не удаляется, `url` не меняется. Конвертируются только JPEG/PNG.
- **Воркер покрывает только фото героев.** Изображения наград (`awards/…`) и вложения заявок (`submissions/…`, включая PDF) хранятся как загружены, без обработки.

---

## 2. Renditions (что реально лежит в бакете)

| Rendition | Ключ | Формат | Геометрия | Качество | Когда готов |
|---|---|---|---|---|---|
| Оригинал (до обработки) | `heroes/YYYY/MM/DD/<uuid>.{jpg,png,webp}` | как загружен | как загружен | — | сразу после PUT |
| Оригинал (после обработки) | `photos/<photoID>.webp` | **WebP lossy** | полный кадр | q90 (`original_quality`) | 1–5 с |
| Thumbnail | `thumbnails/<photoID>.webp` | **WebP lossy** | 480×600 (4:5), crop по FaceBox или center-crop | q80 (`quality`) | 1–5 с |

**srcset-вариантов нет** — ровно два rendition. Encoder: `webp.Encode{Lossless:false, Method:4}`.

---

## 3. Контракт `media.proto` и асинхронное окно

- `GetUploadUrlResponse{upload_url, public_url, expires_at}` — **`public_url` здесь транзиентный**: для JPEG/PNG этот ключ будет **удалён** после конвертации (strict delete) и начнёт отдавать 404. Клиент **не должен** кешировать его как канонический URL фото.
- Канонические URL живут только в БД (`photos.url`, `photos.thumbnail_url`) и приходят через `Photo`/`HeroSummary`/`HeroDetail`.
- Поэтому `thumbnail_url` в `GetUploadUrlResponse` добавлять **нельзя**: на момент ответа его не существует, а детерминированный ключ `thumbnails/<photoID>.webp` нельзя выдать до `AddHeroPhoto` (photoID ещё нет) и нельзя гарантировать (обработка могла упасть).
- **Поведение фронтенда в асинхронном окне (1–5 с)** уже зафиксировано в вашем снапшоте (техдолг #2): `thumbnailUrl || url` как fallback + polling до появления thumbnail. Это корректный контракт, менять ничего не нужно.
- В окне обработки `url` может быть JPEG/PNG, после — WebP. Единственный случай, когда `<picture>`/`<source type="image/webp">` имеет смысл:

```vue
<!-- Карточка: source актуален ТОЛЬКО в асинхронном окне -->
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

После обработки оба URL — WebP, и конструкция вырождается в обычный `<img src=thumbnailUrl>`. **Рекомендация: не строить srcset из двух rendition** (`480w` + оригинал) — ширина оригинала неизвестна клиенту без метаданных; проще выбирать rendition по контексту (thumbnail для сеток, `url` для详情页/lightbox).

---

## 4. FaceBox и `<picture>`

- `FaceBox{x,y,width,height}` задаёт область интереса (лицо) в пикселях оригинала; воркер делает `imaging.Crop` по нему, иначе center-crop 4:5, затем `Fill(480,600, Lanczos)`.
- `main_thumbnail_url` в `HeroSummary` и `thumbnail_url` в `Photo` — это один и тот же объект `thumbnails/<photoID>.webp`.
- **Нужен ли отдельный AVIF для thumbnail?** Сейчас нет: thumbnail уже WebP (поддержка >95% браузеров). AVIF-дубль удвоил бы хранение и CPU воркера ради ~10–15% экономии байт на карточках, которые и так весят десятки КБ.

---

## 5. Надёжность и безопасность (почему это не «просто воркер»)

| Механизм | Детали |
|---|---|
| Decompression bomb protection | `LimitReader` 50 МБ; `DecodeConfig` **до** полного decode: ≤10000 px/сторона, ≤50 МП |
| MIME-whitelist | `image/jpeg|png|webp` (+`application/pdf` для вложений); Content-Type входит в SigV4-подпись presigned URL — подменить тип нельзя |
| Очередь | retry 3× с backoff при переполнении; при исчерпании — восстановление через `RescanLoop` |
| Strict delete | удаление оригинала только после успешного upload WebP; 404 = идемпотентный успех; неудача → задача прерывается до UPDATE БД |
| S3-операции | `withRetry` + exponential backoff на download/upload; `countingReader` для метрик байт |
| Ключи | UUIDv4 + группировка по дате — непредсказуемые пути, нет enumerable-имён |

---

## 6. Конфиг и метрики

```yaml
thumbnails:
  width: 480            # геометрия thumbnail
  height: 600
  quality: 80           # WebP-качество thumbnail
  original_quality: 90  # WebP-качество конвертированного оригинала
  workers: 2            # параллельность воркера
  queue_size: 100
```

Метрики для мониторинга пайплайна: `emh_thumbnail_queue_size`, `emh_thumbnail_processed_total{status}`, `emh_thumbnail_processing_duration_seconds{step=download|decode|crop|upload_thumbnail|upload_original|db_update}`, `emh_thumbnail_skipped_total{reason=dimensions|megapixels|download_error}`, `emh_thumbnail_enqueue_retries_total`, `emh_thumbnail_enqueue_failures_total`, `emh_thumbnail_rescan_triggered_total`, `emh_thumbnail_startup_backfill_*`, `emh_thumbnail_original_delete_failures_total`.

---

## 7. Если понадобится AVIF/srcset — варианты эволюции

| Вариант | Плюсы | Минусы | Вердикт |
|---|---|---|---|
| **(а) Расширить ThumbnailWorker** матрицей renditions (`formats:[webp,avif]`, `widths:[480,800,1200]`) | Единая точка контроля, метрики, security-guards уже есть; ключи детерминированные | AVIF-энкодер в Go = тяжёлая cgo-зависимость (libavif); CPU воркера ×N | Рекомендую, **когда AVIF станет product-требованием**; сейчас преждевременно |
| **(б) imgproxy перед MinIO** | On-the-fly AVIF/srcset/WebP без изменения бэкенда; кеширование | Новая инфраструктура + ещё один CORS-домен через Caddy; дублирование логики guards | Оправдан при росте числа renditions >3–4 |
| **(в) Nuxt Image с S3-провайдером** | — | Чистый MinIO не ресайзит on-the-fly; CORS; неконтролируемый CPU на SSR | **Отклонён** (ваше подозрение подтверждено кодом) |

Текущий выбор (а в минимальной форме: WebP × 2 rendition) осознанный: один бинарник, ноль дополнительной инфраструктуры, детерминированное качество, все security-guards в одном месте.

### Вердикт по эволюции (2026-09-16)

Текущий выбор (вариант «а» в минимальной форме: WebP × 2 rendition) подтверждён как осознанный:

- один бинарник, ноль дополнительной инфраструктуры;
- детерминированное качество;
- все security-guards в одном месте.

**Фронтенд** не должен строить `srcset` из двух rendition. Вместо этого — выбирать
rendition по контексту: `thumbnail_url` для сеток, `url` для детальных страниц / lightbox.
`<picture>` нужен только на асинхронном окне (1–5 с после загрузки).

---

## 8. Первоисточники (для верификации)

- `backend/internal/usecase/thumbnail_worker.go` — пайплайн, guards, strict delete, rescan/backfill
- `backend/internal/usecase/photo_usecase.go` — триггеры Enqueue (Add/Batch/Update face_box)
- `backend/internal/usecase/media_usecase.go` — MIME-whitelist, генерация ключей `heroes/YYYY/MM/DD/<uuid>.ext`
- `backend/internal/storage/s3/media_storage.go` — presign с signed Content-Type, Upload/Download с retry
- `backend/proto/emh/v1/media.proto`, `hero.proto` (FaceBox, Photo, HeroSummary)
- Конфиг: секция `thumbnails`

Замечание: зеркало GitHub может отставать от Codeberg по Спринту 7, но фото-пайплайн в Спринте 7 не изменялся — справка соответствует и локальному коду. Если хочешь, следующим шагом оформлю это как `docs/architecture/photo-pipeline.md` в репозиторий + короткий ответ-письмо архитектору на 10 строк. 📸
