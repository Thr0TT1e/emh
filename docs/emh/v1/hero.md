# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/hero.proto](#emh-v1-hero-proto)
  - **Services**
    - [HeroService](#emh-v1-heroservice)
  - **Messages**
    - [HeroSummary](#emh-v1-herosummary)
    - [HeroDetail](#emh-v1-herodetail)
    - [HeroSource](#emh-v1-herosource)
    - [HeroRelation](#emh-v1-herorelation)
    - [FaceBox](#emh-v1-facebox)
    - [Photo](#emh-v1-photo)
    - [HeroAward](#emh-v1-heroaward)
    - [AwardDevice](#emh-v1-awarddevice)
    - [HeroConflict](#emh-v1-heroconflict)
    - [HeroLocation](#emh-v1-herolocation)
    - [GetHeroRequest](#emh-v1-getherorequest)
    - [GetHeroResponse](#emh-v1-getheroresponse)
    - [ListHeroesRequest](#emh-v1-listheroesrequest)
    - [ListHeroesResponse](#emh-v1-listheroesresponse)
    - [ListHeroPhotosRequest](#emh-v1-listherophotosrequest)
    - [ListHeroPhotosResponse](#emh-v1-listherophotosresponse)
    - [ListHeroSourcesRequest](#emh-v1-listherosourcesrequest)
    - [ListHeroSourcesResponse](#emh-v1-listherosourcesresponse)
    - [ListHeroRelationsRequest](#emh-v1-listherorelationsrequest)
    - [ListHeroRelationsResponse](#emh-v1-listherorelationsresponse)

<a name="emh-v1-hero-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/hero.proto

**Package:** `emh.v1`

<a name="emh-v1-heroservice"></a>

## HeroService

HeroService публичный сервис для чтения данных о героях.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [GetHero](#emh-v1-heroservice-gethero) | ➡️ Unary | — | GetHero возвращает полную карточку �... |
| [ListHeroes](#emh-v1-heroservice-listheroes) | ➡️ Unary | — | ListHeroes выполняет поиск и фильтрац... |
| [ListHeroPhotos](#emh-v1-heroservice-listherophotos) | ➡️ Unary | — | ListHeroPhotos возвращает все фотограф�... |
| [ListHeroSources](#emh-v1-heroservice-listherosources) | ➡️ Unary | — | ListHeroSources возвращает источники да... |
| [ListHeroRelations](#emh-v1-heroservice-listherorelations) | ➡️ Unary | — | ListHeroRelations возвращает связи героя... |

<a name="emh-v1-heroservice-gethero"></a>

### GetHero

```protobuf
rpc GetHero([GetHeroRequest](#emh-v1-getherorequest)) returns ([GetHeroResponse](#emh-v1-getheroresponse))
```

GetHero возвращает полную карточку героя по ID.

#### Request Example

```json
{
  "id": "string"
}
```

#### Response Example

```json
{
  "hero": {
    "audit": {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "createdBy": "string",
      "updatedAt": {
        "nanos": 0,
        "seconds": 0
      }
    },
    "awards": [
      {
        "awardDate": {
          "nanos": 0,
          "seconds": 0
        },
        "awardDateInfo": {
          "anchorDate": {
            "nanos": 0,
            "seconds": 0
          },
          "displayText": "string",
          "precision": "DatePrecision_VALUE"
        },
        "awardId": "string",
        "awardName": "string",
        "decreeNumber": "string",
        "devices": [
          {
            "count": 0,
            "type": "string"
          }
        ],
        "imageUrl": "string",
        "isJubilee": true,
        "jurisdiction": "AwardJurisdiction_VALUE",
        "ribbonImageUrl": "string",
        "type": "AwardType_VALUE",
        "wornWithoutBar": true
      }
    ],
    "causeOfDeath": "string",
    "conflicts": [
      {
        "conflictId": "string",
        "conflictName": "string",
        "rankAtConflict": "string",
        "specificLocation": "string"
      }
    ],
    "fullBio": "string",
    "locations": [
      {
        "location": {
          "historicalName": "string",
          "id": "string",
          "latitude": 0,
          "longitude": 0,
          "name": "string",
          "parentId": "string",
          "type": "LocationType_VALUE"
        },
        "locationId": "string",
        "type": "HeroLocationType_VALUE"
      }
    ],
    "memberships": [
      "string"
    ],
    "photos": [
      {
        "description": "string",
        "faceBox": {
          "height": 0,
          "width": 0,
          "x": 0,
          "y": 0
        },
        "id": "string",
        "isMain": true,
        "sortOrder": 0,
        "thumbnailUrl": "string",
        "url": "string"
      }
    ],
    "position": "string",
    "relations": [
      {
        "description": "string",
        "fromHeroId": "string",
        "id": "string",
        "relatedHeroName": "string",
        "relationType": "string",
        "toHeroId": "string"
      }
    ],
    "serviceStartDate": {
      "nanos": 0,
      "seconds": 0
    },
    "serviceStartDateInfo": {
      "anchorDate": {
        "nanos": 0,
        "seconds": 0
      },
      "displayText": "string",
      "precision": "DatePrecision_VALUE"
    },
    "sources": [
      {
        "excerpt": "string",
        "id": "string",
        "sourceType": "string",
        "title": "string",
        "url": "string"
      }
    ],
    "status": "PublicationStatus_VALUE",
    "summary": {
      "awardNames": [
        "string"
      ],
      "birthDate": {
        "nanos": 0,
        "seconds": 0
      },
      "birthDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "deathDate": {
        "nanos": 0,
        "seconds": 0
      },
      "deathDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "firstName": "string",
      "id": "string",
      "lastName": "string",
      "mainPhotoUrl": "string",
      "mainThumbnailUrl": "string",
      "middleName": "string",
      "nickname": "string",
      "rank": "string",
      "serviceBranch": "string",
      "shortBio": "string",
      "status": "PublicationStatus_VALUE",
      "unit": "string"
    }
  }
}
```

---

<a name="emh-v1-heroservice-listheroes"></a>

### ListHeroes

```protobuf
rpc ListHeroes([ListHeroesRequest](#emh-v1-listheroesrequest)) returns ([ListHeroesResponse](#emh-v1-listheroesresponse))
```

ListHeroes выполняет поиск и фильтрацию героев с пагинацией.

#### Request Example

```json
{
  "conflictId": "string",
  "dateFrom": {
    "nanos": 0,
    "seconds": 0
  },
  "dateTo": {
    "nanos": 0,
    "seconds": 0
  },
  "locationId": "string",
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "searchQuery": "string"
}
```

#### Response Example

```json
{
  "heroes": [
    {
      "awardNames": [
        "string"
      ],
      "birthDate": {
        "nanos": 0,
        "seconds": 0
      },
      "birthDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "deathDate": {
        "nanos": 0,
        "seconds": 0
      },
      "deathDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "firstName": "string",
      "id": "string",
      "lastName": "string",
      "mainPhotoUrl": "string",
      "mainThumbnailUrl": "string",
      "middleName": "string",
      "nickname": "string",
      "rank": "string",
      "serviceBranch": "string",
      "shortBio": "string",
      "status": "PublicationStatus_VALUE",
      "unit": "string"
    }
  ],
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  }
}
```

---

<a name="emh-v1-heroservice-listherophotos"></a>

### ListHeroPhotos

```protobuf
rpc ListHeroPhotos([ListHeroPhotosRequest](#emh-v1-listherophotosrequest)) returns ([ListHeroPhotosResponse](#emh-v1-listherophotosresponse))
```

ListHeroPhotos возвращает все фотографии героя для галереи.

#### Request Example

```json
{
  "heroId": "string",
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  }
}
```

#### Response Example

```json
{
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  },
  "photos": [
    {
      "description": "string",
      "faceBox": {
        "height": 0,
        "width": 0,
        "x": 0,
        "y": 0
      },
      "id": "string",
      "isMain": true,
      "sortOrder": 0,
      "thumbnailUrl": "string",
      "url": "string"
    }
  ]
}
```

---

<a name="emh-v1-heroservice-listherosources"></a>

### ListHeroSources

```protobuf
rpc ListHeroSources([ListHeroSourcesRequest](#emh-v1-listherosourcesrequest)) returns ([ListHeroSourcesResponse](#emh-v1-listherosourcesresponse))
```

ListHeroSources возвращает источники данных героя.

#### Request Example

```json
{
  "heroId": "string"
}
```

#### Response Example

```json
{
  "sources": [
    {
      "excerpt": "string",
      "id": "string",
      "sourceType": "string",
      "title": "string",
      "url": "string"
    }
  ]
}
```

---

<a name="emh-v1-heroservice-listherorelations"></a>

### ListHeroRelations

```protobuf
rpc ListHeroRelations([ListHeroRelationsRequest](#emh-v1-listherorelationsrequest)) returns ([ListHeroRelationsResponse](#emh-v1-listherorelationsresponse))
```

ListHeroRelations возвращает связи героя с другими героями.

#### Request Example

```json
{
  "heroId": "string"
}
```

#### Response Example

```json
{
  "relations": [
    {
      "description": "string",
      "fromHeroId": "string",
      "id": "string",
      "relatedHeroName": "string",
      "relationType": "string",
      "toHeroId": "string"
    }
  ]
}
```

---

<a name="emh-v1-herosummary"></a>

### HeroSummary

HeroSummary содержит краткую информацию о герое для отображения в списках и результатах поиска.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор героя (UUID). |
| last_name | string | optional | last_name - фамилия героя. |
| first_name | string | optional | first_name - имя героя. |
| middle_name | string | optional | middle_name - отчество героя (может быть пустым). |
| main_photo_url | string | optional | main_photo_url - URL главной фотографии для превью. |
| rank | string | optional | rank - воинское звание героя. |
| short_bio | string | optional | short_bio - краткая биография (до 300 символов) для карточек. |
| award_names | string | repeated | award_names - список названий наград для отображения бейджей. |
| birth_date | [Timestamp](#google-protobuf-timestamp) | optional | birth_date - дата рождения героя. DEPRECATED: используйте birth_date_info. Поле сохранено для обратной совместимости. |
| death_date | [Timestamp](#google-protobuf-timestamp) | optional | death_date - дата гибели героя. DEPRECATED: используйте death_date_info. Поле сохранено для обратной совместимости. |
| nickname | string | optional | nickname - позывной героя (например, «Тридцатый»). |
| unit | string | optional | unit - подразделение или воинская часть. |
| service_branch | string | optional | service_branch - ведомство или род войск (ФСБ, ВДВ и т.д.). |
| main_thumbnail_url | string | optional | main_thumbnail_url - URL превью главной фотографии (обрезано по face_box). Используется для карточек в списках. Может быть пустым, если thumbnail ещё не сгенерирован воркером. |
| birth_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | birth_date_info - гибкая дата рождения героя. |
| death_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | death_date_info - гибкая дата гибели героя. |
| status | PublicationStatus | optional | status - текущий статус публикации записи |

<details>
<summary>JSON Example</summary>

```json
{
  "awardNames": [
    "string"
  ],
  "birthDate": {
    "nanos": 0,
    "seconds": 0
  },
  "birthDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "deathDate": {
    "nanos": 0,
    "seconds": 0
  },
  "deathDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "firstName": "string",
  "id": "string",
  "lastName": "string",
  "mainPhotoUrl": "string",
  "mainThumbnailUrl": "string",
  "middleName": "string",
  "nickname": "string",
  "rank": "string",
  "serviceBranch": "string",
  "shortBio": "string",
  "status": "PublicationStatus_VALUE",
  "unit": "string"
}
```

</details>

<a name="emh-v1-herodetail"></a>

### HeroDetail

HeroDetail содержит полную информацию о герое для детальной страницы.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| summary | [HeroSummary](#emh-v1-herosummary) | optional | summary - базовая информация о герое. |
| full_bio | string | optional | full_bio - полный текст биографии или рассказа о подвиге. |
| photos | [Photo](#emh-v1-photo) | repeated | photos - галерея фотографий героя. |
| awards | [HeroAward](#emh-v1-heroaward) | repeated | awards - список наград героя с деталями. |
| conflicts | [HeroConflict](#emh-v1-heroconflict) | repeated | conflicts - список конфликтов, в которых участвовал герой. |
| locations | [HeroLocation](#emh-v1-herolocation) | repeated | locations - географические привязки (рождение, гибель, захоронение). |
| status | PublicationStatus | optional | status - текущий статус публикации записи. |
| audit | [AuditInfo](#emh-v1-auditinfo) | optional | audit - системные метаданные создания и обновления. |
| position | string | optional | position - должность героя. |
| cause_of_death | string | optional | cause_of_death - причина или обстоятельства гибели. |
| service_start_date | [Timestamp](#google-protobuf-timestamp) | optional | service_start_date - дата начала службы. |
| memberships | string | repeated | memberships - членство в организациях и объединениях. |
| sources | [HeroSource](#emh-v1-herosource) | repeated | sources - источники данных о герое. |
| relations | [HeroRelation](#emh-v1-herorelation) | repeated | relations - связи с другими героями. |
| service_start_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | service_start_date_info - гибкая дата начала службы. |

<details>
<summary>JSON Example</summary>

```json
{
  "audit": {
    "createdAt": {
      "nanos": 0,
      "seconds": 0
    },
    "createdBy": "string",
    "updatedAt": {
      "nanos": 0,
      "seconds": 0
    }
  },
  "awards": [
    {
      "awardDate": {
        "nanos": 0,
        "seconds": 0
      },
      "awardDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "awardId": "string",
      "awardName": "string",
      "decreeNumber": "string",
      "devices": [
        {
          "count": 0,
          "type": "string"
        }
      ],
      "imageUrl": "string",
      "isJubilee": true,
      "jurisdiction": "AwardJurisdiction_VALUE",
      "ribbonImageUrl": "string",
      "type": "AwardType_VALUE",
      "wornWithoutBar": true
    }
  ],
  "causeOfDeath": "string",
  "conflicts": [
    {
      "conflictId": "string",
      "conflictName": "string",
      "rankAtConflict": "string",
      "specificLocation": "string"
    }
  ],
  "fullBio": "string",
  "locations": [
    {
      "location": {
        "historicalName": "string",
        "id": "string",
        "latitude": 0,
        "longitude": 0,
        "name": "string",
        "parentId": "string",
        "type": "LocationType_VALUE"
      },
      "locationId": "string",
      "type": "HeroLocationType_VALUE"
    }
  ],
  "memberships": [
    "string"
  ],
  "photos": [
    {
      "description": "string",
      "faceBox": {
        "height": 0,
        "width": 0,
        "x": 0,
        "y": 0
      },
      "id": "string",
      "isMain": true,
      "sortOrder": 0,
      "thumbnailUrl": "string",
      "url": "string"
    }
  ],
  "position": "string",
  "relations": [
    {
      "description": "string",
      "fromHeroId": "string",
      "id": "string",
      "relatedHeroName": "string",
      "relationType": "string",
      "toHeroId": "string"
    }
  ],
  "serviceStartDate": {
    "nanos": 0,
    "seconds": 0
  },
  "serviceStartDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "sources": [
    {
      "excerpt": "string",
      "id": "string",
      "sourceType": "string",
      "title": "string",
      "url": "string"
    }
  ],
  "status": "PublicationStatus_VALUE",
  "summary": {
    "awardNames": [
      "string"
    ],
    "birthDate": {
      "nanos": 0,
      "seconds": 0
    },
    "birthDateInfo": {
      "anchorDate": {
        "nanos": 0,
        "seconds": 0
      },
      "displayText": "string",
      "precision": "DatePrecision_VALUE"
    },
    "deathDate": {
      "nanos": 0,
      "seconds": 0
    },
    "deathDateInfo": {
      "anchorDate": {
        "nanos": 0,
        "seconds": 0
      },
      "displayText": "string",
      "precision": "DatePrecision_VALUE"
    },
    "firstName": "string",
    "id": "string",
    "lastName": "string",
    "mainPhotoUrl": "string",
    "mainThumbnailUrl": "string",
    "middleName": "string",
    "nickname": "string",
    "rank": "string",
    "serviceBranch": "string",
    "shortBio": "string",
    "status": "PublicationStatus_VALUE",
    "unit": "string"
  }
}
```

</details>

<a name="emh-v1-herosource"></a>

### HeroSource

HeroSource представляет источник данных о герое (ссылка на статью, архив, книгу).

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор источника. |
| url | string | optional | url - ссылка на источник. |
| title | string | optional | title - заголовок или название источника. |
| source_type | string | optional | source_type - тип источника (vk, website, archive, book). |
| excerpt | string | optional | excerpt - сохранённый фрагмент текста из источника. |

<details>
<summary>JSON Example</summary>

```json
{
  "excerpt": "string",
  "id": "string",
  "sourceType": "string",
  "title": "string",
  "url": "string"
}
```

</details>

<a name="emh-v1-herorelation"></a>

### HeroRelation

HeroRelation представляет связь героя с другим героем.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор связи. |
| from_hero_id | string | optional | from_hero_id - UUID героя-источника связи. |
| to_hero_id | string | optional | to_hero_id - UUID связанного героя. |
| relation_type | string | optional | relation_type - тип связи (father_son, comrade, commander). |
| description | string | optional | description - описание связи. |
| related_hero_name | string | optional | related_hero_name - ФИО связанного героя (денормализация для удобства). |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "fromHeroId": "string",
  "id": "string",
  "relatedHeroName": "string",
  "relationType": "string",
  "toHeroId": "string"
}
```

</details>

<a name="emh-v1-facebox"></a>

### FaceBox

FaceBox область интереса на фотографии (обычно лицо героя).
Координаты задаются в пикселях от левого верхнего угла изображения.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| x | int32 | optional | x - координата левого верхнего угла (пиксели от левого края). |
| y | int32 | optional | y - координата левого верхнего угла (пиксели от верха). |
| width | int32 | optional | width - ширина области в пикселях. |
| height | int32 | optional | height - высота области в пикселях. |

<details>
<summary>JSON Example</summary>

```json
{
  "height": 0,
  "width": 0,
  "x": 0,
  "y": 0
}
```

</details>

<a name="emh-v1-photo"></a>

### Photo

Photo представляет отдельную фотографию в галерее героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор фотографии. |
| url | string | optional | url - полный URL оригинального изображения. |
| thumbnail_url | string | optional | thumbnail_url - URL уменьшенной версии для галереи. |
| description | string | optional | description - подпись или описание фотографии. |
| sort_order | int32 | optional | sort_order - порядок отображения в галерее. |
| is_main | bool | optional | is_main - флаг главной фотографии (используется в превью). |
| face_box | [FaceBox](#emh-v1-facebox) | optional | face_box - область интереса для crop thumbnail (опционально). |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "faceBox": {
    "height": 0,
    "width": 0,
    "x": 0,
    "y": 0
  },
  "id": "string",
  "isMain": true,
  "sortOrder": 0,
  "thumbnailUrl": "string",
  "url": "string"
}
```

</details>

<a name="emh-v1-heroaward"></a>

### HeroAward

HeroAward описывает конкретную награду героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| award_id | string | optional | award_id - UUID награды из справочника. |
| award_name | string | optional | award_name - название награды (денормализовано для удобства). |
| award_date | [Timestamp](#google-protobuf-timestamp) | optional | award_date - дата вручения или издания указа. |
| decree_number | string | optional | decree_number - номер приказа или указа о награждении. |
| award_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | award_date_info - гибкая дата награждения. Если заполнено, имеет приоритет над award_date. |
| devices | [AwardDevice](#emh-v1-awarddevice) | repeated | devices - знаки повторности (звёздочки, цифры, дубовые листья). Например, при повторном награждении одной и той же медалью. |
| ribbon_image_url | string | optional | ribbon_image_url - URL изображения ленты награды (денормализация из справочника). Пусто, если лента для награды ещё не загружена. |
| type | AwardType | optional | type - тип награды (денормализация из справочника). |
| worn_without_bar | bool | optional | worn_without_bar - награда носится без колодки (денормализация из справочника). |
| is_jubilee | bool | optional | is_jubilee - юбилейная награда (денормализация из справочника). |
| jurisdiction | AwardJurisdiction | optional | jurisdiction - государственная принадлежность награды (денормализация из справочника). |
| image_url | string | optional | image_url - URL изображения знака награды (денормализация из справочника). |

<details>
<summary>JSON Example</summary>

```json
{
  "awardDate": {
    "nanos": 0,
    "seconds": 0
  },
  "awardDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "awardId": "string",
  "awardName": "string",
  "decreeNumber": "string",
  "devices": [
    {
      "count": 0,
      "type": "string"
    }
  ],
  "imageUrl": "string",
  "isJubilee": true,
  "jurisdiction": "AwardJurisdiction_VALUE",
  "ribbonImageUrl": "string",
  "type": "AwardType_VALUE",
  "wornWithoutBar": true
}
```

</details>

<a name="emh-v1-awarddevice"></a>

### AwardDevice

AwardDevice знак повторности на планке.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | string | optional | type - тип знака (звёздочка, цифра, дубовые листья). |
| count | int32 | optional | count - количество (например, 2 звёздочки). |

<details>
<summary>JSON Example</summary>

```json
{
  "count": 0,
  "type": "string"
}
```

</details>

<a name="emh-v1-heroconflict"></a>

### HeroConflict

HeroConflict описывает участие героя в конкретном конфликте.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| conflict_id | string | optional | conflict_id - UUID конфликта из справочника. |
| conflict_name | string | optional | conflict_name - название конфликта (денормализовано). |
| specific_location | string | optional | specific_location - место боя или гибели в рамках данного конфликта. |
| rank_at_conflict | string | optional | rank_at_conflict - звание героя на момент участия в данном конфликте. |

<details>
<summary>JSON Example</summary>

```json
{
  "conflictId": "string",
  "conflictName": "string",
  "rankAtConflict": "string",
  "specificLocation": "string"
}
```

</details>

<a name="emh-v1-herolocation"></a>

### HeroLocation

HeroLocation связывает героя с географическим объектом.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| location_id | string | optional | location_id - UUID локации из справочника. |
| location | [Location](#emh-v1-location) | optional | location - полные данные локации. |
| type | HeroLocationType | optional | type - тип связи (место рождения, гибели и т.д.). |

<details>
<summary>JSON Example</summary>

```json
{
  "location": {
    "historicalName": "string",
    "id": "string",
    "latitude": 0,
    "longitude": 0,
    "name": "string",
    "parentId": "string",
    "type": "LocationType_VALUE"
  },
  "locationId": "string",
  "type": "HeroLocationType_VALUE"
}
```

</details>

<a name="emh-v1-getherorequest"></a>

### GetHeroRequest

GetHeroRequest запрос на получение полной карточки героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID героя. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-getheroresponse"></a>

### GetHeroResponse

GetHeroResponse ответ с полной информацией о герое.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero | [HeroDetail](#emh-v1-herodetail) | optional | hero - детальная карточка героя. |

<details>
<summary>JSON Example</summary>

```json
{
  "hero": {
    "audit": {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "createdBy": "string",
      "updatedAt": {
        "nanos": 0,
        "seconds": 0
      }
    },
    "awards": [
      {
        "awardDate": {
          "nanos": 0,
          "seconds": 0
        },
        "awardDateInfo": {
          "anchorDate": {
            "nanos": 0,
            "seconds": 0
          },
          "displayText": "string",
          "precision": "DatePrecision_VALUE"
        },
        "awardId": "string",
        "awardName": "string",
        "decreeNumber": "string",
        "devices": [
          {
            "count": 0,
            "type": "string"
          }
        ],
        "imageUrl": "string",
        "isJubilee": true,
        "jurisdiction": "AwardJurisdiction_VALUE",
        "ribbonImageUrl": "string",
        "type": "AwardType_VALUE",
        "wornWithoutBar": true
      }
    ],
    "causeOfDeath": "string",
    "conflicts": [
      {
        "conflictId": "string",
        "conflictName": "string",
        "rankAtConflict": "string",
        "specificLocation": "string"
      }
    ],
    "fullBio": "string",
    "locations": [
      {
        "location": {
          "historicalName": "string",
          "id": "string",
          "latitude": 0,
          "longitude": 0,
          "name": "string",
          "parentId": "string",
          "type": "LocationType_VALUE"
        },
        "locationId": "string",
        "type": "HeroLocationType_VALUE"
      }
    ],
    "memberships": [
      "string"
    ],
    "photos": [
      {
        "description": "string",
        "faceBox": {
          "height": 0,
          "width": 0,
          "x": 0,
          "y": 0
        },
        "id": "string",
        "isMain": true,
        "sortOrder": 0,
        "thumbnailUrl": "string",
        "url": "string"
      }
    ],
    "position": "string",
    "relations": [
      {
        "description": "string",
        "fromHeroId": "string",
        "id": "string",
        "relatedHeroName": "string",
        "relationType": "string",
        "toHeroId": "string"
      }
    ],
    "serviceStartDate": {
      "nanos": 0,
      "seconds": 0
    },
    "serviceStartDateInfo": {
      "anchorDate": {
        "nanos": 0,
        "seconds": 0
      },
      "displayText": "string",
      "precision": "DatePrecision_VALUE"
    },
    "sources": [
      {
        "excerpt": "string",
        "id": "string",
        "sourceType": "string",
        "title": "string",
        "url": "string"
      }
    ],
    "status": "PublicationStatus_VALUE",
    "summary": {
      "awardNames": [
        "string"
      ],
      "birthDate": {
        "nanos": 0,
        "seconds": 0
      },
      "birthDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "deathDate": {
        "nanos": 0,
        "seconds": 0
      },
      "deathDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "firstName": "string",
      "id": "string",
      "lastName": "string",
      "mainPhotoUrl": "string",
      "mainThumbnailUrl": "string",
      "middleName": "string",
      "nickname": "string",
      "rank": "string",
      "serviceBranch": "string",
      "shortBio": "string",
      "status": "PublicationStatus_VALUE",
      "unit": "string"
    }
  }
}
```

</details>

<a name="emh-v1-listheroesrequest"></a>

### ListHeroesRequest

ListHeroesRequest параметры поиска и фильтрации списка героев.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [PaginationRequest](#emh-v1-paginationrequest) | optional | pagination - параметры курсорной пагинации. |
| search_query | string | optional | search_query - поисковый запрос по ФИО. |
| conflict_id | string | optional | conflict_id - фильтр по ID конфликта. |
| location_id | string | optional | location_id - фильтр по ID локации (рождение/гибель). |
| date_from | [Timestamp](#google-protobuf-timestamp) | optional | date_from - нижняя граница даты гибели для фильтрации. |
| date_to | [Timestamp](#google-protobuf-timestamp) | optional | date_to - верхняя граница даты гибели для фильтрации. |

<details>
<summary>JSON Example</summary>

```json
{
  "conflictId": "string",
  "dateFrom": {
    "nanos": 0,
    "seconds": 0
  },
  "dateTo": {
    "nanos": 0,
    "seconds": 0
  },
  "locationId": "string",
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "searchQuery": "string"
}
```

</details>

<a name="emh-v1-listheroesresponse"></a>

### ListHeroesResponse

ListHeroesResponse результат поиска героев с пагинацией.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| heroes | [HeroSummary](#emh-v1-herosummary) | repeated | heroes - список кратких карточек героев. |
| pagination | [PaginationResponse](#emh-v1-paginationresponse) | optional | pagination - метаданные для перехода к следующей странице. |

<details>
<summary>JSON Example</summary>

```json
{
  "heroes": [
    {
      "awardNames": [
        "string"
      ],
      "birthDate": {
        "nanos": 0,
        "seconds": 0
      },
      "birthDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "deathDate": {
        "nanos": 0,
        "seconds": 0
      },
      "deathDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "firstName": "string",
      "id": "string",
      "lastName": "string",
      "mainPhotoUrl": "string",
      "mainThumbnailUrl": "string",
      "middleName": "string",
      "nickname": "string",
      "rank": "string",
      "serviceBranch": "string",
      "shortBio": "string",
      "status": "PublicationStatus_VALUE",
      "unit": "string"
    }
  ],
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  }
}
```

</details>

<a name="emh-v1-listherophotosrequest"></a>

### ListHeroPhotosRequest

ListHeroPhotosRequest запрос на получение всех фотографий героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| pagination | [PaginationRequest](#emh-v1-paginationrequest) | optional | pagination - параметры пагинации для больших галерей. |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string",
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  }
}
```

</details>

<a name="emh-v1-listherophotosresponse"></a>

### ListHeroPhotosResponse

ListHeroPhotosResponse список фотографий героя, упорядоченный по sort_order.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| photos | [Photo](#emh-v1-photo) | repeated | photos - фотографии героя. |
| pagination | [PaginationResponse](#emh-v1-paginationresponse) | optional | pagination - метаданные пагинации. |

<details>
<summary>JSON Example</summary>

```json
{
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  },
  "photos": [
    {
      "description": "string",
      "faceBox": {
        "height": 0,
        "width": 0,
        "x": 0,
        "y": 0
      },
      "id": "string",
      "isMain": true,
      "sortOrder": 0,
      "thumbnailUrl": "string",
      "url": "string"
    }
  ]
}
```

</details>

<a name="emh-v1-listherosourcesrequest"></a>

### ListHeroSourcesRequest

ListHeroSourcesRequest запрос на получение источников данных героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string"
}
```

</details>

<a name="emh-v1-listherosourcesresponse"></a>

### ListHeroSourcesResponse

ListHeroSourcesResponse список источников данных героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sources | [HeroSource](#emh-v1-herosource) | repeated | sources - источники данных героя. |

<details>
<summary>JSON Example</summary>

```json
{
  "sources": [
    {
      "excerpt": "string",
      "id": "string",
      "sourceType": "string",
      "title": "string",
      "url": "string"
    }
  ]
}
```

</details>

<a name="emh-v1-listherorelationsrequest"></a>

### ListHeroRelationsRequest

ListHeroRelationsRequest запрос на получение связей героя с другими героями.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string"
}
```

</details>

<a name="emh-v1-listherorelationsresponse"></a>

### ListHeroRelationsResponse

ListHeroRelationsResponse список связей героя с другими героями.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| relations | [HeroRelation](#emh-v1-herorelation) | repeated | relations - связи героя. |

<details>
<summary>JSON Example</summary>

```json
{
  "relations": [
    {
      "description": "string",
      "fromHeroId": "string",
      "id": "string",
      "relatedHeroName": "string",
      "relationType": "string",
      "toHeroId": "string"
    }
  ]
}
```

</details>

