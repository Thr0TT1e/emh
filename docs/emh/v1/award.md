# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/award.proto](#emh-v1-award-proto)
  - **Services**
    - [AwardService](#emh-v1-awardservice)
  - **Messages**
    - [Award](#emh-v1-award)
    - [GetAwardRequest](#emh-v1-getawardrequest)
    - [GetAwardResponse](#emh-v1-getawardresponse)
    - [ListAwardsRequest](#emh-v1-listawardsrequest)
    - [ListAwardsResponse](#emh-v1-listawardsresponse)

<a name="emh-v1-award-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/award.proto

**Package:** `emh.v1`

<a name="emh-v1-awardservice"></a>

## AwardService

AwardService публичный сервис для чтения справочника наград.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [GetAward](#emh-v1-awardservice-getaward) | ➡️ Unary | — | GetAward возвращает награду по ID. |
| [ListAwards](#emh-v1-awardservice-listawards) | ➡️ Unary | — | ListAwards возвращает список всех наг... |

<a name="emh-v1-awardservice-getaward"></a>

### GetAward

```protobuf
rpc GetAward([GetAwardRequest](#emh-v1-getawardrequest)) returns ([GetAwardResponse](#emh-v1-getawardresponse))
```

GetAward возвращает награду по ID.

#### Request Example

```json
{
  "id": "string"
}
```

#### Response Example

```json
{
  "award": {
    "description": "string",
    "id": "string",
    "imageUrl": "string",
    "isJubilee": true,
    "jurisdiction": "AwardJurisdiction_VALUE",
    "name": "string",
    "ribbonImageUrl": "string",
    "sortOrder": 0,
    "type": "AwardType_VALUE",
    "wornWithoutBar": true
  }
}
```

---

<a name="emh-v1-awardservice-listawards"></a>

### ListAwards

```protobuf
rpc ListAwards([ListAwardsRequest](#emh-v1-listawardsrequest)) returns ([ListAwardsResponse](#emh-v1-listawardsresponse))
```

ListAwards возвращает список всех наград (справочник обычно небольшой).

#### Request Example

```json
{
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
  "awards": [
    {
      "description": "string",
      "id": "string",
      "imageUrl": "string",
      "isJubilee": true,
      "jurisdiction": "AwardJurisdiction_VALUE",
      "name": "string",
      "ribbonImageUrl": "string",
      "sortOrder": 0,
      "type": "AwardType_VALUE",
      "wornWithoutBar": true
    }
  ],
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  }
}
```

---

<a name="emh-v1-award"></a>

### Award

Award представляет государственную или ведомственную награду.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор награды (UUID). |
| name | string | optional | name - официальное название награды. |
| description | string | optional | description - краткое описание или статут награды. |
| image_url | string | optional | image_url - URL изображения ленты или знака награды. |
| sort_order | int32 | optional | sort_order - приоритет отображения (старшие награды выше). |
| ribbon_image_url | string | optional | ribbon_image_url - URL изображения ленты награды для орденской планки. |
| type | AwardType | optional | type - тип награды (орден, медаль, знак). Влияет на старшинство согласно приказу МО РФ №1500. |
| worn_without_bar | bool | optional | worn_without_bar - награда носится без колодки (звёзды орденов, знаки). Такие награды не включаются в общий блок планок. |
| is_jubilee | bool | optional | is_jubilee - юбилейная награда (влияет на старшинство: боевые выше юбилейных). |
| jurisdiction | AwardJurisdiction | optional | jurisdiction - государственная принадлежность награды (РФ, СССР, ведомственная). Первичная ось старшинства в орденской планке. |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "id": "string",
  "imageUrl": "string",
  "isJubilee": true,
  "jurisdiction": "AwardJurisdiction_VALUE",
  "name": "string",
  "ribbonImageUrl": "string",
  "sortOrder": 0,
  "type": "AwardType_VALUE",
  "wornWithoutBar": true
}
```

</details>

<a name="emh-v1-getawardrequest"></a>

### GetAwardRequest

GetAwardRequest запрос на получение награды по ID.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID награды. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-getawardresponse"></a>

### GetAwardResponse

GetAwardResponse ответ с данными награды.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| award | [Award](#emh-v1-award) | optional | award - объект награды. |

<details>
<summary>JSON Example</summary>

```json
{
  "award": {
    "description": "string",
    "id": "string",
    "imageUrl": "string",
    "isJubilee": true,
    "jurisdiction": "AwardJurisdiction_VALUE",
    "name": "string",
    "ribbonImageUrl": "string",
    "sortOrder": 0,
    "type": "AwardType_VALUE",
    "wornWithoutBar": true
  }
}
```

</details>

<a name="emh-v1-listawardsrequest"></a>

### ListAwardsRequest

ListAwardsRequest параметры выборки списка наград.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| search_query | string | optional | search_query - поиск по названию награды. |
| pagination | [PaginationRequest](#emh-v1-paginationrequest) | optional | pagination - параметры курсорной пагинации. |

<details>
<summary>JSON Example</summary>

```json
{
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "searchQuery": "string"
}
```

</details>

<a name="emh-v1-listawardsresponse"></a>

### ListAwardsResponse

ListAwardsResponse список доступных наград.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| awards | [Award](#emh-v1-award) | repeated | awards - массив объектов наград. |
| pagination | [PaginationResponse](#emh-v1-paginationresponse) | optional | pagination - метаданные для перехода к следующей странице. |

<details>
<summary>JSON Example</summary>

```json
{
  "awards": [
    {
      "description": "string",
      "id": "string",
      "imageUrl": "string",
      "isJubilee": true,
      "jurisdiction": "AwardJurisdiction_VALUE",
      "name": "string",
      "ribbonImageUrl": "string",
      "sortOrder": 0,
      "type": "AwardType_VALUE",
      "wornWithoutBar": true
    }
  ],
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  }
}
```

</details>

