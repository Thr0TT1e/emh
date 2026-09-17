# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/hero_admin.proto](#emh-v1-hero-admin-proto)
  - **Services**
    - [HeroAdminService](#emh-v1-heroadminservice)
  - **Messages**
    - [CreateHeroRequest](#emh-v1-createherorequest)
    - [CreateHeroResponse](#emh-v1-createheroresponse)
    - [UpdateHeroRequest](#emh-v1-updateherorequest)
    - [UpdateHeroResponse](#emh-v1-updateheroresponse)
    - [DeleteHeroRequest](#emh-v1-deleteherorequest)
    - [DeleteHeroResponse](#emh-v1-deleteheroresponse)
    - [ListAdminHeroesRequest](#emh-v1-listadminheroesrequest)
    - [ListAdminHeroesResponse](#emh-v1-listadminheroesresponse)
    - [AddHeroAwardRequest](#emh-v1-addheroawardrequest)
    - [RemoveHeroAwardRequest](#emh-v1-removeheroawardrequest)
    - [AddHeroConflictRequest](#emh-v1-addheroconflictrequest)
    - [RemoveHeroConflictRequest](#emh-v1-removeheroconflictrequest)
    - [AddHeroLocationRequest](#emh-v1-addherolocationrequest)
    - [RemoveHeroLocationRequest](#emh-v1-removeherolocationrequest)
    - [AddHeroPhotoRequest](#emh-v1-addherophotorequest)
    - [ReorderHeroPhotosRequest](#emh-v1-reorderherophotosrequest)
    - [NewPhoto](#emh-v1-newphoto)
    - [AddedPhoto](#emh-v1-addedphoto)
    - [BatchAddHeroPhotosRequest](#emh-v1-batchaddherophotosrequest)
    - [BatchAddHeroPhotosResponse](#emh-v1-batchaddherophotosresponse)
    - [DeleteHeroPhotosRequest](#emh-v1-deleteherophotosrequest)
    - [DeleteHeroPhotosResponse](#emh-v1-deleteherophotosresponse)
    - [AddHeroSourceRequest](#emh-v1-addherosourcerequest)
    - [RemoveHeroSourceRequest](#emh-v1-removeherosourcerequest)
    - [AddHeroRelationRequest](#emh-v1-addherorelationrequest)
    - [RemoveHeroRelationRequest](#emh-v1-removeherorelationrequest)
    - [SetMainHeroPhotoRequest](#emh-v1-setmainherophotorequest)
    - [UpdateHeroPhotoRequest](#emh-v1-updateherophotorequest)
    - [UpdateHeroPhotoResponse](#emh-v1-updateherophotoresponse)

<a name="emh-v1-hero-admin-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/hero_admin.proto

**Package:** `emh.v1`

<a name="emh-v1-heroadminservice"></a>

## HeroAdminService

HeroAdminService защищенный сервис для полного управления карточками героев.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [CreateHero](#emh-v1-heroadminservice-createhero) | ➡️ Unary | — | CreateHero создает новую карточку гер... |
| [UpdateHero](#emh-v1-heroadminservice-updatehero) | ➡️ Unary | — | UpdateHero обновляет основные поля ге... |
| [DeleteHero](#emh-v1-heroadminservice-deletehero) | ➡️ Unary | — | DeleteHero удаляет или архивирует кар... |
| [AddHeroAward](#emh-v1-heroadminservice-addheroaward) | ➡️ Unary | — | AddHeroAward привязывает награду к гер... |
| [RemoveHeroAward](#emh-v1-heroadminservice-removeheroaward) | ➡️ Unary | — | RemoveHeroAward удаляет связь героя с на... |
| [AddHeroConflict](#emh-v1-heroadminservice-addheroconflict) | ➡️ Unary | — | AddHeroConflict привязывает конфликт к �... |
| [RemoveHeroConflict](#emh-v1-heroadminservice-removeheroconflict) | ➡️ Unary | — | RemoveHeroConflict удаляет связь героя с �... |
| [AddHeroLocation](#emh-v1-heroadminservice-addherolocation) | ➡️ Unary | — | AddHeroLocation привязывает локацию к г�... |
| [RemoveHeroLocation](#emh-v1-heroadminservice-removeherolocation) | ➡️ Unary | — | RemoveHeroLocation удаляет связь героя с �... |
| [AddHeroPhoto](#emh-v1-heroadminservice-addherophoto) | ➡️ Unary | — | AddHeroPhoto добавляет фото в галерею �... |
| [ReorderHeroPhotos](#emh-v1-heroadminservice-reorderherophotos) | ➡️ Unary | — | ReorderHeroPhotos устанавливает новый по... |
| [BatchAddHeroPhotos](#emh-v1-heroadminservice-batchaddherophotos) | ➡️ Unary | — | BatchAddHeroPhotos добавляет несколько ф�... |
| [DeleteHeroPhotos](#emh-v1-heroadminservice-deleteherophotos) | ➡️ Unary | — | DeleteHeroPhotos удаляет отмеченные фот�... |
| [AddHeroSource](#emh-v1-heroadminservice-addherosource) | ➡️ Unary | — | AddHeroSource добавляет источник данны... |
| [RemoveHeroSource](#emh-v1-heroadminservice-removeherosource) | ➡️ Unary | — | RemoveHeroSource удаляет источник данны�... |
| [AddHeroRelation](#emh-v1-heroadminservice-addherorelation) | ➡️ Unary | — | AddHeroRelation создает связь между дву�... |
| [RemoveHeroRelation](#emh-v1-heroadminservice-removeherorelation) | ➡️ Unary | — | RemoveHeroRelation удаляет связь между ге... |
| [SetMainHeroPhoto](#emh-v1-heroadminservice-setmainherophoto) | ➡️ Unary | — | SetMainHeroPhoto назначает указанную фо�... |
| [UpdateHeroPhoto](#emh-v1-heroadminservice-updateherophoto) | ➡️ Unary | — | UpdateHeroPhoto обновляет метаданные фо... |
| [ListHeroes](#emh-v1-heroadminservice-listheroes) | ➡️ Unary | — | ListHeroes возвращает список героев с... |

<a name="emh-v1-heroadminservice-createhero"></a>

### CreateHero

```protobuf
rpc CreateHero([CreateHeroRequest](#emh-v1-createherorequest)) returns ([CreateHeroResponse](#emh-v1-createheroresponse))
```

CreateHero создает новую карточку героя со статусом DRAFT.

#### Request Example

```json
{
  "birthDate": "string",
  "birthDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "causeOfDeath": "string",
  "deathDate": "string",
  "deathDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "firstName": "string",
  "fullBio": "string",
  "lastName": "string",
  "memberships": [
    "string"
  ],
  "middleName": "string",
  "nickname": "string",
  "position": "string",
  "rank": "string",
  "serviceBranch": "string",
  "serviceStartDate": "string",
  "serviceStartDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "shortBio": "string",
  "status": "PublicationStatus_VALUE",
  "unit": "string"
}
```

#### Response Example

```json
{
  "id": "string"
}
```

---

<a name="emh-v1-heroadminservice-updatehero"></a>

### UpdateHero

```protobuf
rpc UpdateHero([UpdateHeroRequest](#emh-v1-updateherorequest)) returns ([UpdateHeroResponse](#emh-v1-updateheroresponse))
```

UpdateHero обновляет основные поля героя с поддержкой partial update.

#### Request Example

```json
{
  "birthDate": "string",
  "birthDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "causeOfDeath": "string",
  "deathDate": "string",
  "deathDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "fieldMask": [
    "string"
  ],
  "firstName": "string",
  "fullBio": "string",
  "id": "string",
  "lastName": "string",
  "memberships": [
    "string"
  ],
  "middleName": "string",
  "nickname": "string",
  "position": "string",
  "rank": "string",
  "serviceBranch": "string",
  "serviceStartDate": "string",
  "serviceStartDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "shortBio": "string",
  "status": "PublicationStatus_VALUE",
  "unit": "string"
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

<a name="emh-v1-heroadminservice-deletehero"></a>

### DeleteHero

```protobuf
rpc DeleteHero([DeleteHeroRequest](#emh-v1-deleteherorequest)) returns ([DeleteHeroResponse](#emh-v1-deleteheroresponse))
```

DeleteHero удаляет или архивирует карточку героя.

#### Request Example

```json
{
  "hardDelete": true,
  "id": "string"
}
```

#### Response Example

```json
{
  "success": true
}
```

---

<a name="emh-v1-heroadminservice-addheroaward"></a>

### AddHeroAward

```protobuf
rpc AddHeroAward([AddHeroAwardRequest](#emh-v1-addheroawardrequest)) returns ([AddHeroAwardRequest](#emh-v1-addheroawardrequest))
```

AddHeroAward привязывает награду к герою. Возвращает подтверждение.

#### Request Example

```json
{
  "awardDate": "string",
  "awardDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "awardId": "string",
  "decreeNumber": "string",
  "devices": [
    {
      "count": 0,
      "type": "string"
    }
  ],
  "heroId": "string"
}
```

#### Response Example

```json
{
  "awardDate": "string",
  "awardDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "awardId": "string",
  "decreeNumber": "string",
  "devices": [
    {
      "count": 0,
      "type": "string"
    }
  ],
  "heroId": "string"
}
```

---

<a name="emh-v1-heroadminservice-removeheroaward"></a>

### RemoveHeroAward

```protobuf
rpc RemoveHeroAward([RemoveHeroAwardRequest](#emh-v1-removeheroawardrequest)) returns ([RemoveHeroAwardRequest](#emh-v1-removeheroawardrequest))
```

RemoveHeroAward удаляет связь героя с наградой.

#### Request Example

```json
{
  "awardId": "string",
  "heroId": "string"
}
```

#### Response Example

```json
{
  "awardId": "string",
  "heroId": "string"
}
```

---

<a name="emh-v1-heroadminservice-addheroconflict"></a>

### AddHeroConflict

```protobuf
rpc AddHeroConflict([AddHeroConflictRequest](#emh-v1-addheroconflictrequest)) returns ([AddHeroConflictRequest](#emh-v1-addheroconflictrequest))
```

AddHeroConflict привязывает конфликт к герою. Возвращает подтверждение.

#### Request Example

```json
{
  "conflictId": "string",
  "heroId": "string",
  "rankAtConflict": "string",
  "specificLocation": "string"
}
```

#### Response Example

```json
{
  "conflictId": "string",
  "heroId": "string",
  "rankAtConflict": "string",
  "specificLocation": "string"
}
```

---

<a name="emh-v1-heroadminservice-removeheroconflict"></a>

### RemoveHeroConflict

```protobuf
rpc RemoveHeroConflict([RemoveHeroConflictRequest](#emh-v1-removeheroconflictrequest)) returns ([RemoveHeroConflictRequest](#emh-v1-removeheroconflictrequest))
```

RemoveHeroConflict удаляет связь героя с конфликтом.

#### Request Example

```json
{
  "conflictId": "string",
  "heroId": "string"
}
```

#### Response Example

```json
{
  "conflictId": "string",
  "heroId": "string"
}
```

---

<a name="emh-v1-heroadminservice-addherolocation"></a>

### AddHeroLocation

```protobuf
rpc AddHeroLocation([AddHeroLocationRequest](#emh-v1-addherolocationrequest)) returns ([AddHeroLocationRequest](#emh-v1-addherolocationrequest))
```

AddHeroLocation привязывает локацию к герою с указанным типом связи.

#### Request Example

```json
{
  "heroId": "string",
  "locationId": "string",
  "type": "HeroLocationType_VALUE"
}
```

#### Response Example

```json
{
  "heroId": "string",
  "locationId": "string",
  "type": "HeroLocationType_VALUE"
}
```

---

<a name="emh-v1-heroadminservice-removeherolocation"></a>

### RemoveHeroLocation

```protobuf
rpc RemoveHeroLocation([RemoveHeroLocationRequest](#emh-v1-removeherolocationrequest)) returns ([RemoveHeroLocationRequest](#emh-v1-removeherolocationrequest))
```

RemoveHeroLocation удаляет связь героя с локацией указанного типа.

#### Request Example

```json
{
  "heroId": "string",
  "locationId": "string",
  "type": "HeroLocationType_VALUE"
}
```

#### Response Example

```json
{
  "heroId": "string",
  "locationId": "string",
  "type": "HeroLocationType_VALUE"
}
```

---

<a name="emh-v1-heroadminservice-addherophoto"></a>

### AddHeroPhoto

```protobuf
rpc AddHeroPhoto([AddHeroPhotoRequest](#emh-v1-addherophotorequest)) returns ([AddHeroPhotoRequest](#emh-v1-addherophotorequest))
```

AddHeroPhoto добавляет фото в галерею героя.

#### Request Example

```json
{
  "description": "string",
  "faceBox": {
    "height": 0,
    "width": 0,
    "x": 0,
    "y": 0
  },
  "heroId": "string",
  "isMain": true,
  "sortOrder": 0,
  "url": "string"
}
```

#### Response Example

```json
{
  "description": "string",
  "faceBox": {
    "height": 0,
    "width": 0,
    "x": 0,
    "y": 0
  },
  "heroId": "string",
  "isMain": true,
  "sortOrder": 0,
  "url": "string"
}
```

---

<a name="emh-v1-heroadminservice-reorderherophotos"></a>

### ReorderHeroPhotos

```protobuf
rpc ReorderHeroPhotos([ReorderHeroPhotosRequest](#emh-v1-reorderherophotosrequest)) returns ([ReorderHeroPhotosRequest](#emh-v1-reorderherophotosrequest))
```

ReorderHeroPhotos устанавливает новый порядок отображения фотографий.

#### Request Example

```json
{
  "heroId": "string",
  "photoIds": [
    "string"
  ]
}
```

#### Response Example

```json
{
  "heroId": "string",
  "photoIds": [
    "string"
  ]
}
```

---

<a name="emh-v1-heroadminservice-batchaddherophotos"></a>

### BatchAddHeroPhotos

```protobuf
rpc BatchAddHeroPhotos([BatchAddHeroPhotosRequest](#emh-v1-batchaddherophotosrequest)) returns ([BatchAddHeroPhotosResponse](#emh-v1-batchaddherophotosresponse))
```

BatchAddHeroPhotos добавляет несколько фотографий за один запрос.

#### Request Example

```json
{
  "heroId": "string",
  "photos": [
    {
      "description": "string",
      "faceBox": {
        "height": 0,
        "width": 0,
        "x": 0,
        "y": 0
      },
      "isMain": true,
      "sortOrder": 0,
      "url": "string"
    }
  ]
}
```

#### Response Example

```json
{
  "added": [
    {
      "id": "string",
      "isMain": true,
      "url": "string"
    }
  ],
  "addedCount": 0
}
```

---

<a name="emh-v1-heroadminservice-deleteherophotos"></a>

### DeleteHeroPhotos

```protobuf
rpc DeleteHeroPhotos([DeleteHeroPhotosRequest](#emh-v1-deleteherophotosrequest)) returns ([DeleteHeroPhotosResponse](#emh-v1-deleteherophotosresponse))
```

DeleteHeroPhotos удаляет отмеченные фотографии пакетно.

#### Request Example

```json
{
  "heroId": "string",
  "photoIds": [
    "string"
  ]
}
```

#### Response Example

```json
{
  "deletedCount": 0,
  "success": true
}
```

---

<a name="emh-v1-heroadminservice-addherosource"></a>

### AddHeroSource

```protobuf
rpc AddHeroSource([AddHeroSourceRequest](#emh-v1-addherosourcerequest)) returns ([AddHeroSourceRequest](#emh-v1-addherosourcerequest))
```

AddHeroSource добавляет источник данных к герою.

#### Request Example

```json
{
  "excerpt": "string",
  "heroId": "string",
  "sourceType": "string",
  "title": "string",
  "url": "string"
}
```

#### Response Example

```json
{
  "excerpt": "string",
  "heroId": "string",
  "sourceType": "string",
  "title": "string",
  "url": "string"
}
```

---

<a name="emh-v1-heroadminservice-removeherosource"></a>

### RemoveHeroSource

```protobuf
rpc RemoveHeroSource([RemoveHeroSourceRequest](#emh-v1-removeherosourcerequest)) returns ([RemoveHeroSourceRequest](#emh-v1-removeherosourcerequest))
```

RemoveHeroSource удаляет источник данных героя.

#### Request Example

```json
{
  "heroId": "string",
  "sourceId": "string"
}
```

#### Response Example

```json
{
  "heroId": "string",
  "sourceId": "string"
}
```

---

<a name="emh-v1-heroadminservice-addherorelation"></a>

### AddHeroRelation

```protobuf
rpc AddHeroRelation([AddHeroRelationRequest](#emh-v1-addherorelationrequest)) returns ([AddHeroRelationRequest](#emh-v1-addherorelationrequest))
```

AddHeroRelation создает связь между двумя героями.

#### Request Example

```json
{
  "description": "string",
  "fromHeroId": "string",
  "relationType": "string",
  "toHeroId": "string"
}
```

#### Response Example

```json
{
  "description": "string",
  "fromHeroId": "string",
  "relationType": "string",
  "toHeroId": "string"
}
```

---

<a name="emh-v1-heroadminservice-removeherorelation"></a>

### RemoveHeroRelation

```protobuf
rpc RemoveHeroRelation([RemoveHeroRelationRequest](#emh-v1-removeherorelationrequest)) returns ([RemoveHeroRelationRequest](#emh-v1-removeherorelationrequest))
```

RemoveHeroRelation удаляет связь между героями.

#### Request Example

```json
{
  "id": "string"
}
```

#### Response Example

```json
{
  "id": "string"
}
```

---

<a name="emh-v1-heroadminservice-setmainherophoto"></a>

### SetMainHeroPhoto

```protobuf
rpc SetMainHeroPhoto([SetMainHeroPhotoRequest](#emh-v1-setmainherophotorequest)) returns ([SetMainHeroPhotoRequest](#emh-v1-setmainherophotorequest))
```

SetMainHeroPhoto назначает указанную фотографию главной для героя.
Триггер ensure_single_main_photo автоматически сбросит флаг у остальных.

#### Request Example

```json
{
  "heroId": "string",
  "photoId": "string"
}
```

#### Response Example

```json
{
  "heroId": "string",
  "photoId": "string"
}
```

---

<a name="emh-v1-heroadminservice-updateherophoto"></a>

### UpdateHeroPhoto

```protobuf
rpc UpdateHeroPhoto([UpdateHeroPhotoRequest](#emh-v1-updateherophotorequest)) returns ([UpdateHeroPhotoResponse](#emh-v1-updateherophotoresponse))
```

UpdateHeroPhoto обновляет метаданные фотографии (описание, face_box).

#### Request Example

```json
{
  "description": "string",
  "faceBox": {
    "height": 0,
    "width": 0,
    "x": 0,
    "y": 0
  },
  "fieldMask": [
    "string"
  ],
  "heroId": "string",
  "photoId": "string"
}
```

#### Response Example

```json
{
  "photo": {
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
}
```

---

<a name="emh-v1-heroadminservice-listheroes"></a>

### ListHeroes

```protobuf
rpc ListHeroes([ListAdminHeroesRequest](#emh-v1-listadminheroesrequest)) returns ([ListAdminHeroesResponse](#emh-v1-listadminheroesresponse))
```

ListHeroes возвращает список героев с поддержкой фильтрации по статусам для админки.

#### Request Example

```json
{
  "includeArchived": true,
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "searchQuery": "string",
  "status": "PublicationStatus_VALUE"
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

<a name="emh-v1-createherorequest"></a>

### CreateHeroRequest

CreateHeroRequest данные для создания карточки героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| first_name | string | optional | first_name - имя героя (обязательно). |
| last_name | string | optional | last_name - фамилия героя (обязательно). |
| middle_name | string | optional | middle_name - отчество героя (опционально). |
| short_bio | string | optional | short_bio - краткая биография для карточек в списках (до 500 символов). |
| full_bio | string | optional | full_bio - полная биография или рассказ о подвиге. |
| rank | string | optional | rank - воинское звание на момент гибели или последнее звание. |
| birth_date | string | optional | birth_date - дата рождения в формате YYYY-MM-DD. |
| death_date | string | optional | death_date - дата гибели в формате YYYY-MM-DD. |
| status | PublicationStatus | optional | status - начальный статус публикации. По умолчанию DRAFT. |
| nickname | string | optional | nickname - позывной или прозвище героя. |
| unit | string | optional | unit - подразделение или воинская часть. |
| position | string | optional | position - должность героя. |
| service_branch | string | optional | service_branch - ведомство или род войск. |
| cause_of_death | string | optional | cause_of_death - причина и обстоятельства гибели. |
| service_start_date | string | optional | service_start_date - дата начала службы в формате YYYY-MM-DD. |
| memberships | string | repeated | memberships - членство в организациях и объединениях. |
| birth_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | birth_date_info - гибкая дата рождения. Если заполнено, имеет приоритет над birth_date. |
| death_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | death_date_info - гибкая дата гибели. Если заполнено, имеет приоритет над death_date. |
| service_start_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | service_start_date_info - гибкая дата начала службы. Если заполнено, имеет приоритет над service_start_date. |

<details>
<summary>JSON Example</summary>

```json
{
  "birthDate": "string",
  "birthDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "causeOfDeath": "string",
  "deathDate": "string",
  "deathDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "firstName": "string",
  "fullBio": "string",
  "lastName": "string",
  "memberships": [
    "string"
  ],
  "middleName": "string",
  "nickname": "string",
  "position": "string",
  "rank": "string",
  "serviceBranch": "string",
  "serviceStartDate": "string",
  "serviceStartDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "shortBio": "string",
  "status": "PublicationStatus_VALUE",
  "unit": "string"
}
```

</details>

<a name="emh-v1-createheroresponse"></a>

### CreateHeroResponse

CreateHeroResponse результат создания героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID созданной записи героя. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-updateherorequest"></a>

### UpdateHeroRequest

UpdateHeroRequest запрос на частичное обновление данных героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID обновляемого героя (обязательно). |
| first_name | string | optional | first_name - новое имя. Игнорируется если пустое. |
| last_name | string | optional | last_name - новая фамилия. Игнорируется если пустое. |
| middle_name | string | optional | middle_name - новое отчество. |
| short_bio | string | optional | short_bio - новая краткая биография. |
| full_bio | string | optional | full_bio - новая полная биография. |
| rank | string | optional | rank - новое звание. |
| birth_date | string | optional | birth_date - новая дата рождения (строка). |
| death_date | string | optional | death_date - новая дата гибели (строка). |
| status | PublicationStatus | optional | status - новый статус публикации. |
| field_mask | string | repeated | field_mask - список полей для обновления. |
| nickname | string | optional | nickname - новый позывной или прозвище. |
| unit | string | optional | unit - новое подразделение или воинская часть. |
| position | string | optional | position - новая должность. |
| service_branch | string | optional | service_branch - новое ведомство или род войск. |
| cause_of_death | string | optional | cause_of_death - новая причина гибели. |
| service_start_date | string | optional | service_start_date - новая дата начала службы. |
| memberships | string | repeated | memberships - новый список членства в организациях. |
| birth_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | birth_date_info - новая гибкая дата рождения. Применяется, если указано в field_mask. |
| death_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | death_date_info - новая гибкая дата гибели. Применяется, если указано в field_mask. |
| service_start_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | service_start_date_info - новая гибкая дата начала службы. Применяется, если указано в field_mask. |

<details>
<summary>JSON Example</summary>

```json
{
  "birthDate": "string",
  "birthDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "causeOfDeath": "string",
  "deathDate": "string",
  "deathDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "fieldMask": [
    "string"
  ],
  "firstName": "string",
  "fullBio": "string",
  "id": "string",
  "lastName": "string",
  "memberships": [
    "string"
  ],
  "middleName": "string",
  "nickname": "string",
  "position": "string",
  "rank": "string",
  "serviceBranch": "string",
  "serviceStartDate": "string",
  "serviceStartDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "shortBio": "string",
  "status": "PublicationStatus_VALUE",
  "unit": "string"
}
```

</details>

<a name="emh-v1-updateheroresponse"></a>

### UpdateHeroResponse

UpdateHeroResponse обновленная полная карточка героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero | [HeroDetail](#emh-v1-herodetail) | optional | hero - актуальное состояние записи героя после обновления. |

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

<a name="emh-v1-deleteherorequest"></a>

### DeleteHeroRequest

DeleteHeroRequest запрос на удаление или архивацию героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID героя. |
| hard_delete | bool | optional | hard_delete - если true, запись удаляется физически. |

<details>
<summary>JSON Example</summary>

```json
{
  "hardDelete": true,
  "id": "string"
}
```

</details>

<a name="emh-v1-deleteheroresponse"></a>

### DeleteHeroResponse

DeleteHeroResponse результат операции удаления.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | bool | optional | success - true если операция выполнена успешно. |

<details>
<summary>JSON Example</summary>

```json
{
  "success": true
}
```

</details>

<a name="emh-v1-listadminheroesrequest"></a>

### ListAdminHeroesRequest

ListAdminHeroesRequest запрос на получение списка героев в админке.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [PaginationRequest](#emh-v1-paginationrequest) | optional | pagination - параметры пагинации. |
| search_query | string | optional | search_query - поисковый запрос по ФИО. |
| status | PublicationStatus | optional | status - фильтр по статусу. 0 (Unspecified) означает "не фильтровать строго" (см. include_archived). |
| include_archived | bool | optional | include_archived - если true, включает архивные записи в общий список (работает при status=0). |

<details>
<summary>JSON Example</summary>

```json
{
  "includeArchived": true,
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "searchQuery": "string",
  "status": "PublicationStatus_VALUE"
}
```

</details>

<a name="emh-v1-listadminheroesresponse"></a>

### ListAdminHeroesResponse

ListAdminHeroesResponse ответ со списком героев для админки.

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

<a name="emh-v1-addheroawardrequest"></a>

### AddHeroAwardRequest

AddHeroAwardRequest запрос на привязку награды к герою.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| award_id | string | optional | award_id - UUID награды из справочника. |
| award_date | string | optional | award_date - дата награждения (строка YYYY-MM-DD). |
| decree_number | string | optional | decree_number - номер приказа или указа о награждении. |
| award_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | award_date_info - гибкая дата награждения. Если заполнено, имеет приоритет над award_date. |
| devices | [AwardDevice](#emh-v1-awarddevice) | repeated | devices - знаки повторности награды (звёздочки, цифры, дубовые листья). |

<details>
<summary>JSON Example</summary>

```json
{
  "awardDate": "string",
  "awardDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "awardId": "string",
  "decreeNumber": "string",
  "devices": [
    {
      "count": 0,
      "type": "string"
    }
  ],
  "heroId": "string"
}
```

</details>

<a name="emh-v1-removeheroawardrequest"></a>

### RemoveHeroAwardRequest

RemoveHeroAwardRequest запрос на отвязку награды от героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| award_id | string | optional | award_id - UUID награды для удаления. |

<details>
<summary>JSON Example</summary>

```json
{
  "awardId": "string",
  "heroId": "string"
}
```

</details>

<a name="emh-v1-addheroconflictrequest"></a>

### AddHeroConflictRequest

AddHeroConflictRequest запрос на привязку конфликта к герою.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| conflict_id | string | optional | conflict_id - UUID конфликта из справочника. |
| specific_location | string | optional | specific_location - место боя или гибели в рамках данного конфликта. |
| rank_at_conflict | string | optional | rank_at_conflict - звание героя на момент участия в данном конфликте. |

<details>
<summary>JSON Example</summary>

```json
{
  "conflictId": "string",
  "heroId": "string",
  "rankAtConflict": "string",
  "specificLocation": "string"
}
```

</details>

<a name="emh-v1-removeheroconflictrequest"></a>

### RemoveHeroConflictRequest

RemoveHeroConflictRequest запрос на отвязку конфликта от героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| conflict_id | string | optional | conflict_id - UUID конфликта для удаления. |

<details>
<summary>JSON Example</summary>

```json
{
  "conflictId": "string",
  "heroId": "string"
}
```

</details>

<a name="emh-v1-addherolocationrequest"></a>

### AddHeroLocationRequest

AddHeroLocationRequest запрос на привязку локации к герою.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| location_id | string | optional | location_id - UUID локации из справочника. |
| type | HeroLocationType | optional | type - тип связи (место рождения, гибели, захоронения, проживания). |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string",
  "locationId": "string",
  "type": "HeroLocationType_VALUE"
}
```

</details>

<a name="emh-v1-removeherolocationrequest"></a>

### RemoveHeroLocationRequest

RemoveHeroLocationRequest запрос на отвязку локации от героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| location_id | string | optional | location_id - UUID локации для удаления. |
| type | HeroLocationType | optional | type - тип удаляемой связи (часть составного первичного ключа). |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string",
  "locationId": "string",
  "type": "HeroLocationType_VALUE"
}
```

</details>

<a name="emh-v1-addherophotorequest"></a>

### AddHeroPhotoRequest

AddHeroPhotoRequest запрос на добавление фотографии в галерею героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| url | string | optional | url - URL загруженного изображения в S3/MinIO. |
| description | string | optional | description - подпись к фотографии. |
| is_main | bool | optional | is_main - если true, фото становится главным. |
| sort_order | int32 | optional | sort_order - позиция в галерее. |
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
  "heroId": "string",
  "isMain": true,
  "sortOrder": 0,
  "url": "string"
}
```

</details>

<a name="emh-v1-reorderherophotosrequest"></a>

### ReorderHeroPhotosRequest

ReorderHeroPhotosRequest запрос на изменение порядка фотографий.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| photo_ids | string | repeated | photo_ids - упорядоченный список ID фотографий. |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string",
  "photoIds": [
    "string"
  ]
}
```

</details>

<a name="emh-v1-newphoto"></a>

### NewPhoto

NewPhoto элемент пакета загружаемых фотографий.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | string | optional | url - URL загруженного изображения в S3/MinIO. |
| description | string | optional | description - подпись к фотографии. |
| is_main | bool | optional | is_main - если true, фото становится главным. При нескольких true в пакете главным становится первое. |
| sort_order | int32 | optional | sort_order - позиция в галерее. Если 0, фото добавляется в конец. |
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
  "isMain": true,
  "sortOrder": 0,
  "url": "string"
}
```

</details>

<a name="emh-v1-addedphoto"></a>

### AddedPhoto

AddedPhoto результат добавления одной фотографии с присвоенным ID.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID созданной фотографии. |
| url | string | optional | url - URL изображения. |
| is_main | bool | optional | is_main - стало ли фото главным. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string",
  "isMain": true,
  "url": "string"
}
```

</details>

<a name="emh-v1-batchaddherophotosrequest"></a>

### BatchAddHeroPhotosRequest

BatchAddHeroPhotosRequest запрос на пакетное добавление фотографий.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| photos | [NewPhoto](#emh-v1-newphoto) | repeated | photos - список добавляемых фотографий (не более 50 за один запрос). |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string",
  "photos": [
    {
      "description": "string",
      "faceBox": {
        "height": 0,
        "width": 0,
        "x": 0,
        "y": 0
      },
      "isMain": true,
      "sortOrder": 0,
      "url": "string"
    }
  ]
}
```

</details>

<a name="emh-v1-batchaddherophotosresponse"></a>

### BatchAddHeroPhotosResponse

BatchAddHeroPhotosResponse результат пакетного добавления фотографий.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| added | [AddedPhoto](#emh-v1-addedphoto) | repeated | added - успешно добавленные фотографии с присвоенными ID. |
| added_count | int32 | optional | added_count - количество добавленных фотографий. |

<details>
<summary>JSON Example</summary>

```json
{
  "added": [
    {
      "id": "string",
      "isMain": true,
      "url": "string"
    }
  ],
  "addedCount": 0
}
```

</details>

<a name="emh-v1-deleteherophotosrequest"></a>

### DeleteHeroPhotosRequest

DeleteHeroPhotosRequest запрос на пакетное удаление отмеченных фотографий.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя (для проверки принадлежности фотографий). |
| photo_ids | string | repeated | photo_ids - список UUID фотографий для удаления (отмеченные галочками). |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string",
  "photoIds": [
    "string"
  ]
}
```

</details>

<a name="emh-v1-deleteherophotosresponse"></a>

### DeleteHeroPhotosResponse

DeleteHeroPhotosResponse результат пакетного удаления фотографий.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deleted_count | int32 | optional | deleted_count - количество фактически удалённых фотографий. |
| success | bool | optional | success - true если операция выполнена успешно. |

<details>
<summary>JSON Example</summary>

```json
{
  "deletedCount": 0,
  "success": true
}
```

</details>

<a name="emh-v1-addherosourcerequest"></a>

### AddHeroSourceRequest

AddHeroSourceRequest запрос на добавление источника данных к герою.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| url | string | optional | url - URL источника (статья, архив, публикация). |
| title | string | optional | title - заголовок источника. |
| source_type | string | optional | source_type - тип источника (vk, website, archive, book). |
| excerpt | string | optional | excerpt - сохранённый фрагмент текста источника. |

<details>
<summary>JSON Example</summary>

```json
{
  "excerpt": "string",
  "heroId": "string",
  "sourceType": "string",
  "title": "string",
  "url": "string"
}
```

</details>

<a name="emh-v1-removeherosourcerequest"></a>

### RemoveHeroSourceRequest

RemoveHeroSourceRequest запрос на удаление источника данных героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| source_id | string | optional | source_id - UUID удаляемого источника. |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string",
  "sourceId": "string"
}
```

</details>

<a name="emh-v1-addherorelationrequest"></a>

### AddHeroRelationRequest

AddHeroRelationRequest запрос на создание связи между двумя героями.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from_hero_id | string | optional | from_hero_id - UUID героя-источника связи. |
| to_hero_id | string | optional | to_hero_id - UUID героя-цели связи. |
| relation_type | string | optional | relation_type - тип связи (father_son, comrade, commander). |
| description | string | optional | description - описание связи. |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "fromHeroId": "string",
  "relationType": "string",
  "toHeroId": "string"
}
```

</details>

<a name="emh-v1-removeherorelationrequest"></a>

### RemoveHeroRelationRequest

RemoveHeroRelationRequest запрос на удаление связи между героями.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID удаляемой связи. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-setmainherophotorequest"></a>

### SetMainHeroPhotoRequest

SetMainHeroPhotoRequest запрос на явное назначение главной фотографии героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero_id | string | optional | hero_id - UUID героя. |
| photo_id | string | optional | photo_id - UUID фотографии, которую нужно сделать главной. |

<details>
<summary>JSON Example</summary>

```json
{
  "heroId": "string",
  "photoId": "string"
}
```

</details>

<a name="emh-v1-updateherophotorequest"></a>

### UpdateHeroPhotoRequest

UpdateHeroPhotoRequest запрос на обновление метаданных фотографии.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| photo_id | string | optional | photo_id - UUID фотографии. |
| hero_id | string | optional | hero_id - UUID героя (для проверки принадлежности). |
| description | string | optional | description - новая подпись к фотографии. |
| face_box | [FaceBox](#emh-v1-facebox) | optional | face_box - новая область интереса для crop thumbnail. |
| field_mask | string | repeated | field_mask - список полей для обновления (description, face_box). |

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
  "fieldMask": [
    "string"
  ],
  "heroId": "string",
  "photoId": "string"
}
```

</details>

<a name="emh-v1-updateherophotoresponse"></a>

### UpdateHeroPhotoResponse

UpdateHeroPhotoResponse результат обновления фотографии.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| photo | [Photo](#emh-v1-photo) | optional | photo - обновлённая фотография. |

<details>
<summary>JSON Example</summary>

```json
{
  "photo": {
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
}
```

</details>

