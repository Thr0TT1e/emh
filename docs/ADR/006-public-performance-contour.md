# ADR-006: Performance-контур публичной части — компрессия, LCP, кэш медиа

**Статус:** Accepted · **Дата:** 2026-11-11 · **Связано:** ADR-001, ADR-004

## Контекст

Lighthouse mobile `/`: ~45/100, TTI 15.2 с при TBT 431 мс — сетевой паттерн, не JS.
Верификация по master и VPS Caddyfile: компрессии нет нигде; `.reveal{opacity:0}`
скрывает карточки до гидрации; логотип (LCP-кандидат) рендерится JS-компонентом
без preload; S3 без preconnect и без Cache-Control. Гипотезы «нет lazy» и
«большой pageSize» опровергнуты кодом (lazy есть, HEROES_PAGE_SIZE=20).

## Решения

1. **Компрессия на edge:** `encode zstd gzip` в Caddy (канонический хост).
   Nitro `compressPublicAssets` — опциональный P2. API не трогаем (Connect gzip уже есть).
2. **Reveal не применяется к контенту первого экрана и к реестру главной вообще.**
   Анимация появления — прогрессивное улучшение ниже сгиба, не условие видимости.
   Глобальный `prefers-reduced-motion` fallback обязателен.
3. **LCP-кандидат (логотип) — нативный `<img>`** с width/height, `fetchpriority="high"`,
   `preload`. PrimeVue `<Image>` для LCP-элементов запрещён.
4. **Eager только для первых 4 карточек**, без fetchpriority high (ниже сгиба на
   мобильных). Остальные — lazy.
5. **Preconnect:** s3.neverforgotten.ru — без `crossorigin` (no-cors `<img>`);
   api.neverforgotten.ru — с `crossorigin=''` (anonymous CORS fetch). Технические
   поддомены живут на neverforgotten.ru (Caddyfile), не на каноническом IDN.
6. **Кэш S3 по мутабельности ключа:** awards/photos/heroes — `immutable, 1 год`
   (write-once); thumbnails — `max-age=300, stale-while-revalidate=604800`
   (перезапись того же ключа при FaceBox-редактировании, см. Reference - photo pipeline §1).

## Альтернативы (отклонены)

| Альтернатива | Почему нет |
|---|---|
| Только Nitro precompression | Требует пересборки образа; Caddy-фикс деплоится отдельно и покрывает всё, включая HTML |
| `immutable` на все S3-пути | Stale-кропы после FaceBox-правок (тот же ключ `thumbnails/<photoID>.webp`) |
| Оставить reveal с исключением первого экрана | Сложнее поддерживать границу «первого экрана» при адаптивной сетке; решение — убрать с реестра полностью |
| preconnect с crossorigin к S3 | CORS-пул сокетов не переиспользуется no-cors `<img>` — холостое соединение |
| imgproxy / Nuxt Image | Отклонено ранее (ADR-001, Reference §7) |

## Последствия

+ TTI: −4–6 с ожидаемо от компрессии; LCP разблокирован (reveal + логотип).
+ Повторные визиты: S3-кэш закрывает аудит efficient cache lifetimes.
− Thumbnails могут быть stale до 5 мин после правки кропа (принято).
− Компрессия добавляет CPU на VPS (мониторить).
