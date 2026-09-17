
# Глоссарий проекта «Вечная память героям»

## Награды (Award System)

| Термин | Определение | Поле в proto |
|--------|-------------|--------------|
| **Award (Справочник)** | Запись в справочнике наград: название, тип, изображения, старшинство | `award.proto:Award` |
| **HeroAward (Привязка)** | Связь героя с наградой: дата, номер указа, знаки повторности | `hero.proto:HeroAward` |
| **Ribbon (Лента)** | Узкая полоса ткани на колодке, визуально представляет награду | `ribbon_image_url` |
| **Bar (Колодка)** | Металлическое основание для лент | - |
| **Order of Precedence** | Порядок старшинства наград (ГОСТ, приказ МО РФ №1500) | `type` + `is_jubilee` + `sort_order` |
| **Worn without bar** | Награда без колодки (звёзды, знаки) | `worn_without_bar` |
| **Device (Знак повторности)** | Доп. элемент на планке за повторное награждение | `HeroAward.devices[]` |
| **AwardType** | Классификация: `ORDER`, `MEDAL`, `BADGE` | `type` |

## География (Locations)

| Термин | Определение | Поле в proto |
|--------|-------------|--------------|
| **Location** | Географический объект: город, населённый пункт, регион | `location.proto:Location` |
| **HeroLocation** | Привязка героя к локации с типом связи | `hero.proto:HeroLocation` |
| **HeroLocationType** | Тип связи: BIRTH, DEATH, BURIAL, RESIDENCE | `HeroLocation.type` |

## Конфликты (Conflicts)

| Термин | Определение | Поле в proto |
|--------|-------------|--------------|
| **Conflict** | Военный конфликт или операция | `conflict.proto:Conflict` |
| **HeroConflict** | Привязка героя к конфликту | `hero.proto:HeroConflict` |
| **ConflictType** | Категория конфликта: WAR, OPERATION, BATTLE | `Conflict.type` |

## Источники и связи (Sources & Relations)

| Термин | Определение | Поле в proto |
|--------|-------------|--------------|
| **HeroSource** | Источник данных о герое: статья, архив, книга | `hero.proto:HeroSource` |
| **HeroRelation** | Связь между двумя героями | `hero.proto:HeroRelation` |

## Медиа (Media)

| Термин | Определение | Поле в proto |
|--------|-------------|--------------|
| **Photo** | Фотография в галерее героя | `hero.proto:Photo` |
| **FaceBox** | Область интереса на фото (обычно лицо) | `hero.proto:FaceBox` |
| **UploadType** | Тип загружаемого контента | `media.proto:UploadType` |
| **Rendition** | Вариант изображения (thumbnail, full) | - |

## Общие термины

| Термин | Определение |
|--------|-------------|
| **FlexibleDate** | Гибкая дата с точностью (год, месяц, день, текст) |
| **PublicationStatus** | Статус публикации: DRAFT, PUBLISHED, ARCHIVED |
| **AuditInfo** | Системные метаданные создания/обновления |
| **PaginationRequest/Response** | Курсорная пагинация |
