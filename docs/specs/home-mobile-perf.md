# Спека: Мобильная производительность главной (/)

**Статус:** Implemented (v0.0.9, 2026-11-11)
**Дата:** 2026-11-11
**Связано:** ADR-001 (Image Delivery), ADR-006 (performance-контур)
**Базовый замер:** Lighthouse mobile ~45/100, TTI 15.2 с, TBT 431 мс (Спринт 10)

## 0. Диагноз (верифицирован по master + VPS Caddyfile)

| # | Причина | Статус |
|---|---------|--------|
| F1 | Нет HTTP-компрессии: `encode` отсутствует в Caddy (репо = VPS), `compressPublicAssets` нет в Nitro. ~0.7–1 МБ сырых JS/CSS/HTML при 1.6 Мбит/с ≈ 4–6 с | ✅ подтверждена |
| F2 | `.reveal { opacity: 0 }` + `v-reveal` на всех карточках реестра → контент невидим до гидрации + IO + 0.7 с transition; LCP не регистрируется | ✅ подтверждена |
| F3 | Логотип через PrimeVue `<Image>` без width/height, без preload/fetchpriority (заявление Спринта 9 в master отсутствует) | ✅ подтверждена |
| F4 | Нет `preconnect` к S3; нет `Cache-Control` на S3-изображениях | ✅ подтверждена |
| H1/H2 | Отсутствие lazy / большой pageSize | ❌ опровергнуты (lazy есть, pageSize=20) |

## 1. Не-цели (не трогать)

- ISR (`isr: false`) — отладка прода, не включать.
- Плавающие Lighthouse-проблемы (Метрика, BFCache, Unused JS, TBT-джиттер).
- `srcset`/AVIF/`<picture>` — запрещено ADR-001.
- `manualChunks` — ломает Nitro.
- pageSize=20 и кнопка «Показать ещё» — не менять.

## 2. Pre-flight (до любых правок Caddy)

```bash
# 1. Проверить, что текущий конфиг вообще валиден
podman exec emh-caddy caddy validate --config /etc/caddy/Caddyfile
```
⚠️ Блок `console.minio.neverforgotten.ru` содержит placeholder bcrypt-хеш.
Если validate падает — сначала сгенерировать реальный хеш
(`podman exec emh-caddy caddy hash-password --plaintext '<пароль>'`)
или временно убрать console-блок. **Без этого reload с новыми правками не пройдёт.**
Путь к конфигу внутри контейнера сверить с volume в `emh-caddy.container`.

## 3. P0 — Caddy (инфра, без пересборки образа фронта)

### T1. Компрессия — `infra/podman/Caddyfile`

В блок `xn--b1aaibdmu4k.xn--p1acf` первой строкой:
```caddyfile
encode zstd gzip
```
Caddy сам пропускает уже сжатые ответы (WebP, woff2, gzip от Connect) и не сжимает
несжимаемое. API-хост не трогаем: Connect уже отдаёт gzip (`connect.WithCompression`).

### T2. Cache-Control на S3 — блок `s3.neverforgotten.ru`, перед `reverse_proxy`

```caddyfile
# Write-once ключи: UUID (photos/heroes) и контент-хеш (awards)
@immutable_get {
    method GET
    path */awards/* */photos/* */heroes/*
}
header @immutable_get Cache-Control "public, max-age=31536000, immutable"

# Thumbnails перезаписываются воркером при FaceBox-редактировании (тот же ключ)
@thumb_get {
    method GET
    path */thumbnails/*
}
header @thumb_get Cache-Control "public, max-age=300, stale-while-revalidate=604800"
```
Матчеры `*/...` — bucket-agnostic (имя бакета в публичном URL не фиксируем в спеке).
Обоснование политик: `Reference - photo pipeline.md` §1–2 (ключ `thumbnails/<photoID>.webp`
перезаписывается; оригиналы и awards — write-once).

**Деплой T1+T2:** синхронизировать Caddyfile на VPS → `caddy validate` →
`podman exec emh-caddy caddy reload --config /etc/caddy/Caddyfile`.
Откат: предыдущая версия файла + reload.

## 4. P0 — Фронтенд (новый образ через CI)

### T3. Убрать reveal с реестра полностью (решение G2) — `app/pages/index.vue`

Было:
```html
<NuxtLink v-for="(h, i) in heroes" :key="h.id" :to="`/heroes/${h.id}`"
  class="hero-card reveal" v-reveal :style="{ transitionDelay: `${(i%12)*30}ms` }">
```
Стало:
```html
<NuxtLink v-for="(h, i) in heroes" :key="h.id" :to="`/heroes/${h.id}`" class="hero-card">
```
Плагин `reveal.ts` и класс `.reveal` остаются для других страниц.

### T4. Первые карточки — eager — `app/pages/index.vue`

```html
<img v-if="h.mainThumbnailUrl" :src="h.mainThumbnailUrl" :alt="fullName(h)"
  width="480" height="600" :loading="i < 4 ? 'eager' : 'lazy'"
  decoding="async" class="hero-card__photo-img" />
```
- `i < 4` — два ряда мобильной сетки; остальные lazy (не менять).
- `fetchpriority="high"` карточкам НЕ давать: на мобильных они ниже первого
  экрана, high-приоритет конкурировал бы с LCP-логотипом.
- Мёртвый `sizes` (без `srcset` не работает) — убрать.

### T5. Логотип → нативный `<img>` + preload — `app/pages/index.vue`

```html
<img src="/logo_v5_full_vert_wbr.svg" alt="Вечная память героям"
  width="300" height="H" fetchpriority="high" decoding="async" />
```
`H` — из viewBox SVG (сохранить пропорции; проверить `public/logo_v5_full_vert_wbr.svg`).
CSS: `.masthead__logo img { height: 100% }` → `height: auto`.

В `useHead` страницы:
```ts
link: [
  { rel: 'canonical', href: canonical },
  { rel: 'preload', as: 'image', href: '/logo_v5_full_vert_wbr.svg',
    type: 'image/svg+xml', fetchpriority: 'high' },
],
```

### T6. Preconnect — `nuxt.config.ts`, `app.head.link`

```ts
// Изображения <img> грузятся в no-cors режиме → preconnect БЕЗ crossorigin
{ rel: 'preconnect', href: 'https://s3.neverforgotten.ru' },
{ rel: 'dns-prefetch', href: 'https://s3.neverforgotten.ru' },
// Connect fetch — anonymous CORS → preconnect С crossorigin
{ rel: 'preconnect', href: 'https://api.neverforgotten.ru', crossorigin: '' },
```
Домены — по фактическому Caddyfile (api/s3 живут на neverforgotten.ru,
редирект зеркал покрывает только апекс). Метаданные снапшота про
`s3.вежливые.рус` — неточны, исправить при обновлении снапшота.

### T7. prefers-reduced-motion — `app/assets/css/main.css`

```css
@media (prefers-reduced-motion: reduce) {
  .reveal { opacity: 1; transform: none; transition: none; }
}
```
Глобальный a11y-fallback (чинит и первый экран contacts, и любые будущие reveal).

## 5. P2 — опционально (вне среза, не блокирует)

- `nitro.compressPublicAssets: { gzip: true, brotli: true }` — precompression
  статики на билде (экономия CPU VPS; Caddy encode уже покрывает).
- Уборка `sizes` на остальных страницах, где нет `srcset`.

## 6. Верификация (прод, выполняет владелец — решение G4)

Порядок даёт чистую атрибуцию эффекта:
1. **До всего:** Lighthouse mobile `/` (зафиксировать TTI/TBT/LCP/FCP/CLS).
   → Базовый замер (Спринт 10): ~45/100, TTI 15.2 с, TBT 431 мс.
2. **После T1+T2 (Caddy):** ✅ выполнено 2026-11-11:
   - `content-encoding: zstd` на HTML/JS/SVG; HTML 364→57 КиБ (−6.4×), entry JS 338→94 КиБ (−3.6×), SVG 12.8→3.8 КиБ (−3.4×)
   - S3 GET: thumbnails `public, max-age=300, stale-while-revalidate=604800`; photos/awards `public, max-age=31536000, immutable` (HEAD не тегирован — матчеры по `method GET`)
   - Lighthouse mobile: **59/100**, TTI 15.2→**5.7 с**, TBT 431→**280 мс**, FCP **2.8 с**, LCP **4.3 с**, CLS 0.313 (старый образ фронта + reveal ещё в проде)
   - ⚠️ Замер curl: `Cache-Control`/`content-encoding` проверять на **GET** (`curl -s -o /dev/null -D-`), HEAD не матчится `method GET` и отдаёт content-length: 13 от бэкенда.
   Команда (HEAD не годится: `method GET`-матчеры его не тегируют, а тело HEAD —
   заглушка бэкенда с content-length: 13; проверять GET-ом):
```bash
curl -s -o /dev/null -D- -H 'Accept-Encoding: gzip, zstd' https://xn--b1aaibdmu4k.xn--p1acf/ | grep -i content-encoding
curl -s -o /dev/null -D- -H 'Accept-Encoding: gzip, zstd' 'https://xn--b1aaibdmu4k.xn--p1acf/_nuxt/<entry>.js' | grep -i content-encoding
curl -s -o /dev/null -D- 'https://s3.neverforgotten.ru/<bucket>/thumbnails/<id>.webp' | grep -i cache-control
curl -s -o /dev/null -D- 'https://s3.neverforgotten.ru/<bucket>/photos/<id>.webp' | grep -i cache-control
# размеры: сравнить size_download с Accept-Encoding: identity vs zstd
```
   Повторный Lighthouse → эффект компрессии отдельно.
3. **После фронта (T3–T7):** ✅ выполнено 2026-11-11 (v0.0.9 на проде, два прогона mobile):
   - **Perf 56–61** (было ~45), **TTI 6.7–7.0 с** (было 15.2; цель <8 ✅), **TBT 450–630 мс** (lab-джиттер),
     **FCP 2.8 с**, **CLS 0.000** ✅ (было 0.313 на post-caddy прогоне — reveal скрывал контент),
     **LCP 6.7–7.0 с** — элемент зарегистрирован: нативный `<img>` логотип (до фикса LCP не фиксировался вообще).
     Цель LCP <2.5 ❌: фазы — TTFB ~0.7, Load Delay ~1.1, Load Time ~1.7–2.1 (декод SVG на throttled CPU),
     Render Delay 43–50% (~2.9–3.4 с, main thread занят гидрацией). render-blocking: 0 — сеть не узел.
     Дальнейшее сокращение — вне ADR-006 (JS-бандл = отклонённые «Unused JS/TBT-джиттер», не-цели §1).

## 7. Критерии приёмки

- [x] `content-encoding: zstd/gzip` на HTML, JS, CSS, SVG (замер 2026-11-11: HTML −6.4×, JS −3.6×, SVG −3.4×)
- [x] `Cache-Control` на S3 GET по таблице §3 (замер: thumbnails `max-age=300, SWR=604800`; photos/awards `immutable`; матчеры method GET, HEAD не тегирован — намеренно)
- [x] В HTML главной у карточек нет `reveal`/`v-reveal`/stagger (прод: 0 совпадений; prerender: 0)
- [x] Первые 4 карточки `loading="eager"`, остальные `lazy` (прод: eager=4/lazy=7)
- [x] Логотип — нативный `<img>` с width/height/fetchpriority + preload в `<head>` (прод подтверждён)
- [x] `preconnect` к s3/api в `<head>` всех страниц (глобальный head; проверено on /, /about, /contacts)
- [x] `prefers-reduced-motion` fallback в CSS (в собранном entry.css)
- [ ] Цели (ориентир, не гарантия): TTI < 8 с ✅ (6.7–7.0), LCP < 2.5 с ❌ (6.7–7.0, Render Delay CPU-декод SVG), Perf > 75 ❌ (56–61; +11–16 к базовому 45)
- [x] `pnpm build`, `oxlint` — зелёные; `caddy validate` — ок

## 8. Деплой и откат

- Caddy: файл на VPS → validate → reload (без пересборки образов). Откат — старый файл + reload.
- Фронт: коммит → **lightweight**-тег `v0.0.9` (annotated не триггерит CI — прецедент v0.0.8/run 42)
  → CI → quadlet `:latest` → `systemctl --user restart emh-frontend`.
- Главная prerender-ится (`'/': { prerender: true }`) — изменения HTML видны только после пересборки образа.

## 9. Риски

| Риск | Митигация |
|---|---|
| Placeholder bcrypt в console-блоке роняет validate/reload | Pre-flight §2 |
| Stale thumbnail до 5 мин после FaceBox-правки | Принято; админ видит кроп в кроппере клиентски |
| encode добавляет CPU на VPS | zstd дефолтный уровень; мониторить после деплоя |
