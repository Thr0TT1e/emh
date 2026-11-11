# Инструкция по подготовке картинок наград для сида

## 1. Структура папок

```
backend/seeds/awards/
  descriptions.json
  images/                    # знаки наград (медали, ордена, знаки)
    medal-for-courage.webp
    order-st-george-1.webp
    ...
  ribbons/                   # ленты (планки) наград
    ribbon-medal-for-courage.webp
    ribbon-order-st-george-1.webp
    ...
```

## 2. Требования к изображениям

### 2.1. Знак награды (медаль/орден/знак)

| Параметр | Значение |
|----------|----------|
| Формат | WebP |
| Размер | **400×400 px** (строго квадрат) |
| Фон | Белый (`#FFFFFF`) |
| Композиция | Знак центрирован, contain-fit (умещается целиком) |
| Качество | q80–q90 |
| Макс. размер файла | 100 КБ |
| Именование | `<slug>.webp` (латиница, kebab-case) |

**Как подготовить:**
1. Исходник (обычно PNG с прозрачностью)
2. Canvas 400×400, белый фон
3. Масштабирование по длинной стороне (contain)
4. Центрирование
5. Экспорт в WebP q85

### 2.2. Лента (планка)

| Параметр | Значение |
|----------|----------|
| Формат | WebP |
| Размер | **144×48 px** (ratio 3:1) |
| Фон | Белый (`#FFFFFF`) |
| Композиция | Лента во всю ширину, центрирована вертикально |
| Качество | q80–q90 |
| Макс. размер файла | 20 КБ |
| Именование | `ribbon-<slug>.webp` |

**Как подготовить:**
1. Исходник ленты (обычно вектор SVG или PNG)
2. Canvas 144×48, белый фон
3. Лента занимает всю ширину
4. Экспорт в WebP q85

## 3. Именование

- Только латиница + цифры + `-` (kebab-case)
- `slug` = машиночитаемое имя награды (например, `medal-for-courage`, `order-st-george-1`)
- Ленты имеют префикс `ribbon-`

**Примеры:**
- `medal-for-courage.webp` + `ribbon-medal-for-courage.webp`
- `order-st-andrew.webp` + `ribbon-order-st-andrew.webp`
- `badge-merited-military-pilot.webp` + `ribbon-badge-merited-military-pilot.webp`

**Не использовать:**
- Кириллицу (`медаль-за-отвагу.webp` ❌)
- Пробелы (`medal for courage.webp` ❌)
- Подчёркивания (`medal_for_courage.webp` ❌)
- Заглавные буквы (`Medal-For-Courage.webp` ❌)

## 4. Инструменты

### 4.1. ImageMagick (пакетная конвертация)

```bash
# Знаки: PNG → 400×400 WebP с белым фоном
for f in *.png; do
  convert "$f" \
    -background white \
    -resize 400x400 \
    -gravity center \
    -extent 400x400 \
    -quality 85 \
    "${f%.png}.webp"
done

# Ленты: PNG → 144×48 WebP с белым фоном
for f in ribbons/*.png; do
  convert "$f" \
    -background white \
    -resize 144x48 \
    -gravity center \
    -extent 144x48 \
    -quality 85 \
    "${f%.png}.webp"
done
```

### 4.2. cwebp (официальный encoder)

```bash
cwebp -q 85 input.png -o output.webp
```

### 4.3. Проверка размеров

```bash
# Проверить все картинки
for f in images/*.webp ribbons/*.webp; do
  identify -format "%f: %wx%h\n" "$f"
done
```

## 5. Чеклист перед коммитом

- [ ] Все файлы в WebP (проверка: `file images/*.webp | grep -i webp`)
- [ ] Знаки ровно 400×400 px
- [ ] Ленты ровно 144×48 px
- [ ] Файлы ≤ 100 КБ (знаки) / ≤ 20 КБ (ленты)
- [ ] Белый фон
- [ ] Именование в kebab-case
- [ ] Соответствие 1:1 между `images/` и `ribbons/` (каждая награда имеет обе картинки)
- [ ] `descriptions.json` ссылается на существующие пути
- [ ] Нет внутренних дубликатов `(name, type, jurisdiction)` в `descriptions.json`
- [ ] `seed.json` собран (`just build-seed`) и валиден

## 6. Git LFS

```bash
# Один раз (при первом коммите):
git lfs install
git lfs track "backend/seeds/awards/images/*.webp"
git lfs track "backend/seeds/awards/ribbons/*.webp"
git lfs track "backend/seeds/awards/seed.json"
git add .gitattributes

# Коммит:
git add backend/seeds/awards/
git commit -m "seed: add initial awards seed (N awards)"
git push
```

## 7. Источники качественных изображений

- **Награды РФ/СССР:** официальный сайт Президента РФ (`kremlin.ru`) — раздел «Награды»
- **Ведомственные:** сайты МВД, Росгвардии, ФСБ — разделы «Награды» / «Символика»
- **Векторные SVG:** Wikimedia Commons (Public Domain / CC-BY-SA)
- **Ленты:** таблицы из приказа МО РФ №1500 (раздел «Описание лент»)
